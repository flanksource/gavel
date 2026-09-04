package run

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// Prepared is a run resolved but not dispatched: the host that owns the todo's
// lifecycle, the step chosen and why, and the exact request captain would be
// given. The dashboard's preview, `--dry-run`, the pre-flight validation every
// entrypoint runs, and Start itself all read it — one fold, so a preview can
// never describe a different run from the one that follows it.
type Prepared struct {
	Host *lifecycle.Host
	Step lifecycle.Step
	// Reason is why this step was chosen: named by the caller, or picked by the
	// lifecycle's own predicates.
	Reason     string
	Resolution *lifecycle.Resolution
	// SessionID is the agent session the run will use.
	SessionID string
}

// Resolve chooses the step and folds the run, without dispatching or persisting
// anything. It is a var so tests can substitute a stub without standing up an
// agent.
var Resolve = defaultResolve

func defaultResolve(ctx context.Context, req Request) (*Prepared, error) {
	if req.Todo == nil {
		return nil, errors.New("todo run: no todo")
	}
	host, err := lifecycle.NewHost(req.Provider, req.Dir, req.Options.Host)
	if err != nil {
		return nil, err
	}
	step, reason, err := host.StepFor(ctx, req.Todo, req.Options.Step)
	if err != nil {
		return nil, err
	}
	resolution, err := host.Resolve(ctx, req.Todo, step, runOptions(req, nil))
	if err != nil {
		return nil, err
	}
	sessionID := SessionIDFor(resolution.Spec, req.Todo, req.Options.Resume)
	resolution.UseSession(sessionID)
	return &Prepared{Host: host, Step: step, Reason: reason, Resolution: resolution, SessionID: sessionID}, nil
}

// runOptions is the caller's decisions in the host's own vocabulary. exec is nil
// for a resolution, and the run's context for a dispatch.
func runOptions(req Request, exec *todos.ExecutorContext) lifecycle.RunOptions {
	return lifecycle.RunOptions{
		Exec:       exec,
		Request:    req.Options.Request,
		Prior:      req.Options.Prior,
		Resume:     req.Options.Resume,
		Message:    req.Options.Message,
		Concurrent: req.Options.Concurrent,
		Broker:     req.Broker,
	}
}

// Start admits a run and returns as soon as the agent session is prepared; the
// run itself continues in the background. It is a var so tests can substitute a
// stub without standing up an agent.
//
// Start reports todos.ErrRunOwnedElsewhere rather than resolving it: whether to
// run alongside a live run is the caller's question to ask, and its callers are
// HTTP handlers that must not block on a dialog. Terminal callers ask with
// ConfirmConcurrent and retry with Options.Concurrent set.
var Start = defaultStart

func defaultStart(req Request) (StartResult, error) {
	if req.Registry == nil {
		return StartResult{}, errors.New("todo run registry is required")
	}
	prepared := req.Prepared
	if prepared == nil {
		resolved, err := Resolve(context.Background(), req)
		if err != nil {
			return StartResult{}, err
		}
		prepared = resolved
	}
	// The run outlives the call, so it gets its own context rather than the
	// caller's — a request finishing must not cancel the agent it started.
	ctx, timeoutCancel := context.WithTimeout(context.Background(), prepared.Resolution.Timeout)
	runCtx, stop := context.WithCancelCause(ctx)
	handle, err := req.Registry.Register(RegisterOptions{
		IssueIDs:   IssueIDs([]*types.TODO{req.Todo}),
		Stoppable:  IsStoppable(prepared.Resolution.Spec),
		Concurrent: req.Options.Concurrent,
		Cancel:     stop,
	})
	if err != nil {
		timeoutCancel()
		return StartResult{}, err
	}
	type startOutcome struct {
		result StartResult
		err    error
	}
	started := make(chan startOutcome, 1)
	done := make(chan error, 1)
	var notifyOnce sync.Once
	notify := func(result StartResult, err error) {
		notifyOnce.Do(func() {
			result.Done, result.Stop = done, stop
			started <- startOutcome{result: result, err: err}
		})
	}
	go func() {
		defer timeoutCancel()
		defer handle.Release()

		execCtx := todos.NewExecutorContext(runCtx, logger.StandardLogger(), nil)
		execCtx.SetRunPreparedHook(func(preparation todos.RunPreparationResult) {
			// The run is stoppable by the identity Captain admitted for it, which
			// only exists from here on.
			handle.BindPromptRun(preparation.PromptRunID)
			notify(StartResult{Status: "started", SessionID: preparation.SessionID}, nil)
		})
		err := runStep(execCtx, req, prepared)
		// A run that ended before Captain admitted it never notified: whatever
		// ended it is Start's error, and ending unadmitted is one even when the
		// step itself succeeded. A run that ended after admission reports its own
		// error through Done, the only channel its caller is still reading.
		startErr := err
		if startErr == nil {
			startErr = errors.New("todo run completed before Captain admission")
		}
		notify(StartResult{}, startErr)
		done <- err
		close(done)
	}()
	outcome := <-started
	return outcome.result, outcome.err
}

// runStep dispatches the prepared step and applies its outcome, returning the
// run's own error: nil when the step ran and its outcome was persisted.
func runStep(execCtx *todos.ExecutorContext, req Request, prepared *Prepared) error {
	outcome, err := prepared.Host.Dispatch(execCtx, req.Todo, prepared.Resolution, runOptions(req, execCtx))
	if outcome == nil {
		if err == nil {
			err = fmt.Errorf("lifecycle step %s produced no outcome", prepared.Step.Name)
		}
		logger.Warnf("todo run %s failed: %v", Label(req.Todo), err)
		return err
	}
	// The outcome is applied even when the step could not be classified: the
	// attempt happened, and losing the transcript of a run that failed to map
	// onto a status is exactly the run whose record is worth most.
	status := outcome.Status
	if err != nil && status == "" {
		status = lifecycle.OutcomeKeep
	}
	if applyErr := prepared.Host.OnOutcome(execCtx, req.Todo, prepared.Step, outcome, status); applyErr != nil {
		err = errors.Join(err, applyErr)
	}
	if err != nil && !outcome.Execution.Cancelled {
		logger.Warnf("todo run %s failed: %v", Label(req.Todo), err)
	}
	return err
}
