package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/flanksource/clicky/prompt"
)

// Prompt relay across the worker protocol: a form step in an isolated worker
// asks through workerPrompter, which emits a `prompt` event to the parent and
// blocks on the matching answer the parent writes to the worker's answer
// descriptor. The parent's promptRelay asks its own Prompter and writes that
// answer back; a worker that gives up on a prompt (its step timed out) emits
// `prompt_cancel`, which ends the parent's ask.

const (
	workerEventPrompt       = "prompt"
	workerEventPromptCancel = "prompt_cancel"
)

var workerEventKinds = []string{"progress", "result", workerEventPrompt, workerEventPromptCancel, workerEventLookupResult}

// workerAnswer is a message on the answer descriptor: the parent's reply to
// one `prompt` event, or (LookupID set) a lookup the worker resolves.
type workerAnswer struct {
	Version  int            `json:"version"`
	PromptID string         `json:"prompt_id,omitempty"`
	Answer   *prompt.Answer `json:"answer,omitempty"`
	Error    string         `json:"error,omitempty"`
	LookupID string         `json:"lookup_id,omitempty"`
	Lookup   *workerLookup  `json:"lookup,omitempty"`
}

// workerPrompter is the Prompter a worker's form steps ask through, and what
// resolves the lookups the parent forwards to them.
type workerPrompter struct {
	ctx       context.Context
	send      func(workerEvent) error
	mu        sync.Mutex
	next      int
	pending   map[string]chan workerAnswer
	abandoned map[string]bool
	closed    error
	lookups   map[string]context.CancelFunc
	lookupWG  sync.WaitGroup
}

func newWorkerPrompter(ctx context.Context, send func(workerEvent) error) *workerPrompter {
	return &workerPrompter{ctx: ctx, send: send, pending: map[string]chan workerAnswer{}, abandoned: map[string]bool{},
		lookups: map[string]context.CancelFunc{}}
}

// startWorkerPrompter relays the worker's prompts over send and reads the
// parent's answers and lookups from answers.
func startWorkerPrompter(ctx context.Context, send func(workerEvent) error, answers io.Reader) (*workerPrompter, error) {
	if answers == nil {
		return nil, fmt.Errorf("fixture worker relays prompts but has no answer descriptor")
	}
	relay := newWorkerPrompter(ctx, send)
	go relay.read(answers)
	return relay, nil
}

func (w *workerPrompter) Ask(ctx context.Context, p prompt.Prompt) (prompt.Answer, error) {
	id, reply, err := w.register()
	if err != nil {
		return prompt.Answer{}, err
	}
	if err := w.send(workerEvent{Version: 1, Kind: workerEventPrompt, PromptID: id, Prompt: &p}); err != nil {
		w.abandon(id)
		return prompt.Answer{}, fmt.Errorf("relay prompt to the fixture parent: %w", err)
	}
	select {
	case answer := <-reply:
		if answer.Error != "" {
			return prompt.Answer{}, errors.New(answer.Error)
		}
		if answer.Answer == nil {
			return prompt.Answer{}, fmt.Errorf("fixture parent answered prompt %s with neither an answer nor an error", id)
		}
		return *answer.Answer, nil
	case <-ctx.Done():
		w.abandon(id)
		cancelErr := w.send(workerEvent{Version: 1, Kind: workerEventPromptCancel, PromptID: id})
		return prompt.Answer{}, errors.Join(context.Cause(ctx), cancelErr)
	}
}

func (w *workerPrompter) register() (string, chan workerAnswer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed != nil {
		return "", nil, w.closed
	}
	w.next++
	id := strconv.Itoa(w.next)
	reply := make(chan workerAnswer, 1)
	w.pending[id] = reply
	return id, reply, nil
}

func (w *workerPrompter) abandon(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, ok := w.pending[id]; ok {
		delete(w.pending, id)
		w.abandoned[id] = true
	}
}

// read delivers the parent's answers until the answer descriptor closes, then
// fails every prompt still pending.
func (w *workerPrompter) read(answers io.Reader) {
	decoder := json.NewDecoder(answers)
	for {
		var answer workerAnswer
		if err := decoder.Decode(&answer); err != nil {
			w.close(fmt.Errorf("fixture parent answer channel closed: %w", err))
			return
		}
		if err := w.receive(answer); err != nil {
			w.close(err)
			return
		}
	}
}

func (w *workerPrompter) receive(answer workerAnswer) error {
	if answer.Version != 1 {
		return fmt.Errorf("fixture parent answer version %d", answer.Version)
	}
	if answer.LookupID == "" {
		return w.deliver(answer)
	}
	if answer.Lookup == nil {
		return fmt.Errorf("fixture parent sent lookup %s without a request", answer.LookupID)
	}
	return w.lookup(answer.LookupID, *answer.Lookup)
}

func (w *workerPrompter) deliver(answer workerAnswer) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if reply, ok := w.pending[answer.PromptID]; ok {
		delete(w.pending, answer.PromptID)
		reply <- answer
		return nil
	}
	if w.abandoned[answer.PromptID] {
		delete(w.abandoned, answer.PromptID)
		return nil
	}
	return fmt.Errorf("fixture parent answered prompt %q, which the worker never asked", answer.PromptID)
}

func (w *workerPrompter) close(err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = err
	for id, reply := range w.pending {
		delete(w.pending, id)
		reply <- workerAnswer{PromptID: id, Error: err.Error()}
	}
}

// promptRelay is the parent's side: it asks the run's Prompter for each prompt
// its worker relays and writes the answer back, and forwards the lookups of
// the prompt's fields to the worker while it is open.
type promptRelay struct {
	prompter         Prompter
	answers          io.Writer
	ctx              context.Context
	stop             context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.Mutex
	cancels          map[string]context.CancelFunc
	errs             []error
	nextLookup       int
	lookups          map[string]chan workerEvent
	abandonedLookups map[string]bool
}

func newPromptRelay(ctx context.Context, prompter Prompter, answers io.Writer) *promptRelay {
	ctx, stop := context.WithCancel(ctx)
	return &promptRelay{prompter: prompter, answers: answers, ctx: ctx, stop: stop, cancels: map[string]context.CancelFunc{},
		lookups: map[string]chan workerEvent{}, abandonedLookups: map[string]bool{}}
}

// handle routes a prompt, prompt_cancel or lookup_result event; any other
// kind is not its.
func (r *promptRelay) handle(event workerEvent) error {
	switch event.Kind {
	case workerEventPrompt:
		if r.prompter == nil {
			return fmt.Errorf("fixture worker asked prompt %s, but the run has no prompter", event.PromptID)
		}
		if event.Prompt == nil || event.PromptID == "" {
			return fmt.Errorf("fixture worker sent a prompt event without a prompt or id")
		}
		release, err := r.forward(*event.Prompt)
		if err != nil {
			return err
		}
		r.ask(event.PromptID, *event.Prompt, release)
	case workerEventPromptCancel:
		r.mu.Lock()
		if cancel, ok := r.cancels[event.PromptID]; ok {
			cancel()
		}
		r.mu.Unlock()
	case workerEventLookupResult:
		return r.deliverLookup(event)
	}
	return nil
}

// ask asks p and writes its answer back, then releases its forwarded lookups.
func (r *promptRelay) ask(id string, p prompt.Prompt, release func()) {
	ctx, cancel := context.WithCancel(r.ctx)
	r.mu.Lock()
	r.cancels[id] = cancel
	r.mu.Unlock()
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer cancel()
		defer release()
		reply := workerAnswer{Version: 1, PromptID: id}
		if answer, err := r.prompter.Ask(ctx, p); err != nil {
			reply.Error = err.Error()
		} else {
			reply.Answer = &answer
		}
		r.mu.Lock()
		delete(r.cancels, id)
		r.mu.Unlock()
		if err := r.send(reply); err != nil && r.ctx.Err() == nil {
			r.mu.Lock()
			r.errs = append(r.errs, fmt.Errorf("answer fixture worker prompt %s: %w", id, err))
			r.mu.Unlock()
		}
	}()
}

// send writes message to the worker, refusing once the worker has exited.
func (r *promptRelay) send(message workerAnswer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ctx.Err() != nil {
		return fmt.Errorf("the fixture worker has exited")
	}
	return json.NewEncoder(r.answers).Encode(message)
}

// close ends every ask still open once the worker has exited and reports the
// answers that could not be written.
func (r *promptRelay) close() error {
	r.stop()
	r.wg.Wait()
	r.mu.Lock()
	defer r.mu.Unlock()
	return errors.Join(r.errs...)
}
