package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// RunOutcome is how Captain files a finished step run: the terminal state and
// phase gavel's envelope maps to, and the summary, structured output and error
// the run leaves on its row. An ask envelope parks the run waiting on its
// answer; a failed verdict or a failed envelope fails it; a stop cancels it.
func RunOutcome(result *todos.ExecutionResult, verifyStep bool) promptrun.Outcome {
	if result == nil {
		return promptrun.Outcome{
			State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseGenerate, Error: "agent run returned no result",
		}
	}
	out := promptrun.Outcome{
		Phase: captaindb.PromptRunPhaseFinished, Text: strings.TrimSpace(result.Summary),
		JSON: result.OutputJSON, Error: strings.TrimSpace(result.ErrorMessage),
	}
	failedVerification := result.DoD != nil && result.DoD.Ran && !result.DoD.Passed
	switch {
	case result.Cancelled:
		out.State = captaindb.PromptRunStateCancelled
	case result.EndStatus == types.EndAsk:
		out.State, out.Phase = captaindb.PromptRunStateWaiting, captaindb.PromptRunPhaseGenerate
	case failedVerification || !result.Success || result.EndStatus == types.EndFailed || out.Error != "":
		out.State, out.Phase = captaindb.PromptRunStateFailed, captaindb.PromptRunPhaseGenerate
		if failedVerification || verifyStep {
			out.Phase = captaindb.PromptRunPhaseVerify
		}
		if out.Error == "" {
			out.Error = firstNonEmpty(out.Text, "agent run failed")
		}
	default:
		out.State = captaindb.PromptRunStateSucceeded
	}
	return out
}

// admit clears the run with the runtime before anything is dispatched — live
// runs, resumes, concurrency — and hands the Recording it returns to promptrun,
// which admits the run and files its lifecycle. The identity is bound to the
// execution now, so the approval broker built before dispatch addresses the run
// Captain is about to admit; the dashboard hears of it only once the admission
// is committed.
func (h *Host) admit(exec *todos.ExecutorContext, todo *types.TODO, step Step, prepared *preparedStep, input *stepInput, opts RunOptions) (todos.RunPreparationResult, error) {
	lifecycleProvider, ok := h.Provider.(todos.RunLifecycleProvider)
	if !ok {
		return todos.RunPreparationResult{}, nil
	}
	spec := prepared.request
	runtime := api.RuntimeOf(spec.Provider, spec.Mode)
	persistCtx, cancel := todos.PersistenceContext(exec)
	defer cancel()
	admission, err := lifecycleProvider.PrepareRun(persistCtx, todo, todos.RunPreparation{
		Mode: prepared.class, Prompt: step.Name, ExecutorName: h.executorName(prepared),
		Resume: opts.Resume, Concurrent: opts.Concurrent,
		Requested: captaindb.PromptRunRuntimeSelection{
			Provider: runtime.Provider, Mode: string(spec.Mode), Model: spec.Name, Effort: string(spec.Effort),
		},
		PromptMarkdown: spec.Prompt.User,
	})
	if err != nil {
		return todos.RunPreparationResult{}, fmt.Errorf("prepare native TODO run: %w", err)
	}
	if admission.Record == nil {
		return todos.RunPreparationResult{}, fmt.Errorf("prepare native TODO run %s: provider %T returned no recording", admission.PromptRunID, h.Provider)
	}
	exec.BindRun(admission.RunPreparationResult)
	exec.SetSessionIDHook(func(sessionID string) {
		setSessionID(todo, sessionID)
		h.updateState(exec, todo, todos.StateUpdate{SessionID: &sessionID})
	})
	exec.SetRunStartHook(h.runStartCommenter(exec, todo, prepared.class))
	input.record(admission.Record, func() error {
		persistCtx, cancel := todos.PersistenceContext(exec)
		defer cancel()
		if err := lifecycleProvider.RunAdmitted(persistCtx, todo, admission.RunPreparationResult); err != nil {
			return fmt.Errorf("drive admitted TODO run %s: %w", admission.PromptRunID, err)
		}
		exec.RecordRunPrepared(admission.RunPreparationResult)
		exec.RecordRunStart(input.meta)
		return nil
	})
	return admission.RunPreparationResult, nil
}

// recordedOutcome classifies the run Captain finished the way gavel reads it —
// its envelope, its definition of done, a stop — so Captain files the terminal
// state gavel's outcomes act on. It collects a copy of the execution record: the
// one the host keeps is collected once the run has returned.
func (h *Host) recordedOutcome(runCtx context.Context, step Step, prepared *preparedStep, input *stepInput, start time.Time) func(promptrun.Result, error, bool) (promptrun.Outcome, error) {
	return func(out promptrun.Result, runErr error, stopped bool) (promptrun.Outcome, error) {
		base, err := promptrun.DefaultOutcome(out, runErr, stopped)
		if err != nil || stopped || runErr != nil {
			return base, err
		}
		d := input.dispatched(runCtx, out, runErr)
		execution := *d.execution
		d.execution = &execution
		collected := h.collect(step, prepared, d, start)
		mapped := RunOutcome(collected.Execution, prepared.class == types.ModeVerify)
		if base.State != captaindb.PromptRunStateSucceeded {
			mapped.State, mapped.Phase = base.State, base.Phase
			if mapped.Error == "" {
				mapped.Error = base.Error
			}
		}
		return mapped, nil
	}
}

// dispatched is what came back from captain, with the two facts only the
// dispatching context can tell: whether the run was cancelled by its caller,
// and whether it hit its deadline before reporting a result.
type dispatched struct {
	out       promptrun.Result
	err       error
	cancelled bool
	timedOut  bool
	execution *todos.ExecutionResult
}

func (in *stepInput) dispatched(runCtx context.Context, out promptrun.Result, err error) dispatched {
	return dispatched{
		out: out, err: err, execution: in.execution,
		cancelled: errors.Is(context.Cause(runCtx), todos.ErrExecutionCancelled),
		timedOut:  errors.Is(runCtx.Err(), context.DeadlineExceeded) && !in.sawResult,
	}
}
