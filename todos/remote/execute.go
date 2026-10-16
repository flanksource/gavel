// The sandbox side of a remote run: Execute runs a job's prompt run in the sandbox's checkout,
// with gavel's hooks rebuilt, and sends serve each message as it happens.
package remote

import (
	"context"
	"fmt"
	"sync"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
)

// Finished is how a run ended: its result, and its own error.
type Finished struct {
	Result promptrun.Result
	Err    error
}

// Execute runs job in dir, the sandbox's checkout, sends its start, events and verification
// progress, and returns how it ended, for the caller to report once it has uploaded what the run
// left behind. agent is the provider the run uses; nil builds it from the job's model, as a local
// run does. Execute returns an error only when a message couldn't be sent, and stops the run when
// its start couldn't.
func Execute(ctx context.Context, job Job, dir string, agent captainai.Provider, send func(Message) error) (Finished, error) {
	reporter := &reporter{send: send}
	in := job.Input(dir)
	in.Provider = agent
	hooks, err := job.HooksIn(dir).Build()
	if err != nil {
		return Finished{}, err
	}
	in.Hooks = append(hooks, &startedHook{reporter: reporter})
	if agent != nil {
		// A supplied provider skips Captain's setup hook, so the run is set up here, as a local
		// run with a supplied provider is.
		in.Hooks = append(in.Hooks, &capsetup.Plugin{BaseDir: dir})
	}
	in.OnEvent = func(iteration int, event api.Event) {
		reporter.report(EventMessage(iteration, event), "an event")
	}
	in.Verify.Progress = func(report api.VerifyReport) {
		reporter.report(ProgressMessage(report), "verification progress")
	}
	result, err := promptrun.Run(ctx, in)
	if reporter.startErr != nil {
		return Finished{}, reporter.startErr
	}
	reporter.mu.Lock()
	defer reporter.mu.Unlock()
	return Finished{Result: result, Err: err}, reporter.err
}

// reporter sends messages one at a time and keeps the first failure.
type reporter struct {
	send     func(Message) error
	mu       sync.Mutex
	err      error
	startErr error
}

func (r *reporter) report(message Message, what string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return
	}
	if err := r.send(message); err != nil {
		r.err = fmt.Errorf("report %s of the run to serve: %w", what, err)
	}
}

// startedHook reports the run's start. It is the last of the caller's hooks, so the spec it
// reports carries what gavel's hooks set on it.
type startedHook struct {
	reporter *reporter
}

func (h *startedHook) Name() string { return "gavel-remote-started" }

func (h *startedHook) PreRun(hc *agent.HookContext) error {
	var spec api.Spec
	if hc.Request != nil {
		spec = *hc.Request
	}
	h.reporter.mu.Lock()
	defer h.reporter.mu.Unlock()
	if err := h.reporter.send(StartedMessage(spec)); err != nil {
		h.reporter.startErr = fmt.Errorf("report the run's start to serve: %w", err)
		h.reporter.err = h.reporter.startErr
		return h.reporter.startErr
	}
	return nil
}
