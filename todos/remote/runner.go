// Serve's side of a remote run: Runner starts a sandbox job for a step's prompt run and files the
// run in Captain's store live from the messages the job sends, which a Relay delivers.
package remote

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/google/uuid"
)

// Jobs starts sandbox jobs.
type Jobs interface {
	// Start starts the sandbox job that runs job as run id.
	Start(ctx context.Context, id uuid.UUID, job Job) (Started, error)
}

// Started is a sandbox job that has started.
type Started interface {
	// Wait blocks until the job exits and returns its exit code.
	Wait(ctx context.Context) (int, error)
	Cancel(ctx context.Context) error
}

// Runner runs a step's prompt run in a sandbox job. It is a lifecycle.RemoteRunner.
type Runner struct {
	Jobs  Jobs
	Relay *Relay
}

// Run admits run, tells the host it was admitted, starts its sandbox job, and files what the job
// reports until it reports a result, exits without one, or ctx ends, which cancels the job.
func (r *Runner) Run(ctx context.Context, run lifecycle.RemoteRun) (promptrun.Result, error) {
	id := uuid.New()
	messages, stopListening := r.Relay.Listen(id)
	defer stopListening()

	in := run.Input
	var recorder *promptrun.Recorder
	var broker *approval.Broker
	if in.Record != nil {
		admitted, err := promptrun.Admit(ctx, in)
		if err != nil {
			return promptrun.Result{}, err
		}
		model := in.Config.Model
		if model.Name == "" {
			model = in.Resolved.Spec.Model
		}
		recorder = promptrun.Track(ctx, in.Record, admitted, api.RuntimeOf(model.Provider, model.Mode), model, in.Resolved.Spec, in.Config.SessionID)
		if broker, err = approvals(ctx, in, admitted); err != nil {
			return promptrun.Result{}, err
		}
	}
	finish := func(result promptrun.Result, runErr error, stopped bool) (promptrun.Result, error) {
		if recorder == nil {
			return result, runErr
		}
		var classify func(promptrun.Result, error, bool) (promptrun.Outcome, error)
		if in.Record != nil {
			classify = in.Record.Outcome
		}
		return recorder.Complete(classify, result, runErr, stopped)
	}

	if run.Admitted != nil {
		if err := run.Admitted(); err != nil {
			return finish(promptrun.Result{}, err, false)
		}
	}
	job, err := r.Jobs.Start(ctx, id, JobFor(run))
	if err != nil {
		return finish(promptrun.Result{}, fmt.Errorf("start the sandbox job of run %s: %w", id, err), false)
	}
	exited := make(chan int, 1)
	go func() {
		code, err := job.Wait(context.WithoutCancel(ctx))
		if err != nil {
			code = -1
		}
		exited <- code
	}()
	stop := func(cause error) (promptrun.Result, error) {
		if err := job.Cancel(context.WithoutCancel(ctx)); err != nil {
			cause = errors.Join(cause, fmt.Errorf("cancel the sandbox job of run %s: %w", id, err))
		}
		return finish(promptrun.Result{}, cause, true)
	}

	for {
		select {
		case message := <-messages:
			switch message.Kind {
			case KindStarted:
				if recorder != nil && message.Spec != nil {
					if err := recorder.Start(message.Spec, message.Workspace); err != nil {
						return stop(err)
					}
				}
			case KindEvent:
				event := message.DecodedEvent()
				if recorder != nil {
					recorder.Observe(message.Iteration, event)
				}
				if in.OnEvent != nil {
					in.OnEvent(message.Iteration, event)
				}
			case KindProgress:
				if message.Progress == nil {
					continue
				}
				if recorder != nil {
					recorder.Progress(*message.Progress)
				}
				if in.Verify.Progress != nil {
					in.Verify.Progress(*message.Progress)
				}
			case KindTool, KindApproval:
				go func(message Message) { message.Answer(answer(ctx, in, broker, message)) }(message)
			case KindResult:
				var runErr error
				if message.Error != "" {
					runErr = errors.New(message.Error)
				}
				return finish(message.Result.Unwrap(), runErr, false)
			}
		case code := <-exited:
			return finish(promptrun.Result{}, fmt.Errorf("the sandbox job of run %s exited with code %d before it reported a result", id, code), false)
		case <-ctx.Done():
			return stop(fmt.Errorf("run %s stopped: %w", id, context.Cause(ctx)))
		}
	}
}

// approvals is the broker that asks a person to approve the run's tool uses, as a local recorded
// run's does; nil for a run that asks for none.
func approvals(ctx context.Context, in promptrun.Input, run *database.PromptRun) (*approval.Broker, error) {
	if in.Approvals == nil {
		return nil, nil
	}
	window, err := in.Resolved.Spec.Permissions.ParseApprovalTimeout()
	if err != nil {
		return nil, fmt.Errorf("approvals of run %s: %w", run.ID, err)
	}
	if window == 0 {
		window = approval.ProviderTimeout
	}
	deadline, _ := ctx.Deadline()
	broker := &approval.Broker{
		DB: in.Record.DB, SessionID: run.SessionID, PromptRunID: run.ID,
		RequestedBy: in.Approvals.RequestedBy, Timeout: window, Deadline: deadline,
		Notify: func(_ context.Context, event api.Event) error {
			if in.OnEvent != nil {
				in.OnEvent(0, event)
			}
			return nil
		},
	}
	if err := broker.Validate(); err != nil {
		return nil, fmt.Errorf("approvals of run %s: %w", run.ID, err)
	}
	return broker, nil
}

// answer runs one of the run's tools, or asks a person to approve a tool use, for its job.
func answer(ctx context.Context, in promptrun.Input, broker *approval.Broker, message Message) Reply {
	switch message.Kind {
	case KindTool:
		for _, tool := range in.Config.Tools {
			if tool.Name != message.Tool || tool.Handler == nil {
				continue
			}
			output, err := tool.Handler(ctx, message.Input)
			if err != nil {
				return Reply{Error: err.Error()}
			}
			data, err := json.Marshal(output)
			if err != nil {
				return Reply{Error: fmt.Sprintf("encode the output of tool %s: %v", message.Tool, err)}
			}
			return Reply{Output: data}
		}
		return Reply{Error: fmt.Sprintf("the run has no tool %q", message.Tool)}
	default:
		if broker == nil || message.Approval == nil {
			return Reply{Error: "the run asks for no approvals"}
		}
		decision, err := broker.OnApproval(ctx, *message.Approval)
		if err != nil {
			return Reply{Error: err.Error()}
		}
		return Reply{Decision: &decision}
	}
}

// ErrNoListener is a message for a run no host in this process is waiting for: serve restarted,
// or the run already ended.
var ErrNoListener = errors.New("no host is waiting for this run")

// Relay delivers the messages a sandbox job sends to the Runner waiting for its run.
type Relay struct {
	mu   sync.Mutex
	runs map[uuid.UUID]chan Message
}

// Listen starts taking run id's messages; call the returned func to stop.
func (r *Relay) Listen(id uuid.UUID) (<-chan Message, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.runs == nil {
		r.runs = map[uuid.UUID]chan Message{}
	}
	messages := make(chan Message)
	r.runs[id] = messages
	return messages, func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		delete(r.runs, id)
	}
}

// Call hands a call to run id's Runner and waits for its reply, or until ctx ends.
func (r *Relay) Call(ctx context.Context, id uuid.UUID, message Message) (Reply, error) {
	message.reply = make(chan Reply, 1)
	if err := r.Deliver(ctx, id, message); err != nil {
		return Reply{}, err
	}
	select {
	case reply := <-message.reply:
		return reply, nil
	case <-ctx.Done():
		return Reply{}, ctx.Err()
	}
}

// Deliver hands message to run id's Runner, waiting until it takes it or ctx ends.
func (r *Relay) Deliver(ctx context.Context, id uuid.UUID, message Message) error {
	r.mu.Lock()
	messages, ok := r.runs[id]
	r.mu.Unlock()
	if !ok {
		return fmt.Errorf("%w: run %s", ErrNoListener, id)
	}
	select {
	case messages <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
