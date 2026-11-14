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

var workerEventKinds = []string{"progress", "result", workerEventPrompt, workerEventPromptCancel}

// workerAnswer is the parent's reply to one `prompt` event.
type workerAnswer struct {
	Version  int            `json:"version"`
	PromptID string         `json:"prompt_id"`
	Answer   *prompt.Answer `json:"answer,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// workerPrompter is the Prompter a worker's form steps ask through.
type workerPrompter struct {
	send      func(workerEvent) error
	mu        sync.Mutex
	next      int
	pending   map[string]chan workerAnswer
	abandoned map[string]bool
	closed    error
}

func newWorkerPrompter(send func(workerEvent) error) *workerPrompter {
	return &workerPrompter{send: send, pending: map[string]chan workerAnswer{}, abandoned: map[string]bool{}}
}

// startWorkerPrompter relays the worker's prompts over send and reads the
// parent's answers from answers.
func startWorkerPrompter(send func(workerEvent) error, answers io.Reader) (*workerPrompter, error) {
	if answers == nil {
		return nil, fmt.Errorf("fixture worker relays prompts but has no answer descriptor")
	}
	relay := newWorkerPrompter(send)
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
		if err := w.deliver(answer); err != nil {
			w.close(err)
			return
		}
	}
}

func (w *workerPrompter) deliver(answer workerAnswer) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if answer.Version != 1 {
		return fmt.Errorf("fixture parent answer version %d", answer.Version)
	}
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
// its worker relays and writes the answer back.
type promptRelay struct {
	prompter Prompter
	answers  io.Writer
	ctx      context.Context
	stop     context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.Mutex
	cancels  map[string]context.CancelFunc
	errs     []error
}

func newPromptRelay(ctx context.Context, prompter Prompter, answers io.Writer) *promptRelay {
	ctx, stop := context.WithCancel(ctx)
	return &promptRelay{prompter: prompter, answers: answers, ctx: ctx, stop: stop, cancels: map[string]context.CancelFunc{}}
}

// handle routes a prompt or prompt_cancel event; any other kind is not its.
func (r *promptRelay) handle(event workerEvent) error {
	switch event.Kind {
	case workerEventPrompt:
		if r.prompter == nil {
			return fmt.Errorf("fixture worker asked prompt %s, but the run has no prompter", event.PromptID)
		}
		if event.Prompt == nil || event.PromptID == "" {
			return fmt.Errorf("fixture worker sent a prompt event without a prompt or id")
		}
		r.ask(event.PromptID, *event.Prompt)
	case workerEventPromptCancel:
		r.mu.Lock()
		if cancel, ok := r.cancels[event.PromptID]; ok {
			cancel()
		}
		r.mu.Unlock()
	}
	return nil
}

func (r *promptRelay) ask(id string, p prompt.Prompt) {
	ctx, cancel := context.WithCancel(r.ctx)
	r.mu.Lock()
	r.cancels[id] = cancel
	r.mu.Unlock()
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		defer cancel()
		reply := workerAnswer{Version: 1, PromptID: id}
		if answer, err := r.prompter.Ask(ctx, p); err != nil {
			reply.Error = err.Error()
		} else {
			reply.Answer = &answer
		}
		r.write(reply)
	}()
}

func (r *promptRelay) write(reply workerAnswer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cancels, reply.PromptID)
	if r.ctx.Err() != nil {
		// The worker has exited: nobody reads the answer.
		return
	}
	if err := json.NewEncoder(r.answers).Encode(reply); err != nil {
		r.errs = append(r.errs, fmt.Errorf("answer fixture worker prompt %s: %w", reply.PromptID, err))
	}
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
