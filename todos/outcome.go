package todos

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
)

// SetMode selects the executor's run mode (run or plan); it drives the
// envelope→status mapping in applyOutcome. Empty keeps the default (run).
func (e *TODOExecutor) SetMode(m types.RunMode) {
	if m != "" {
		e.mode = m
	}
}

// Mode returns the executor's run mode (ModeRun when unset).
func (e *TODOExecutor) Mode() types.RunMode {
	if e.mode == "" {
		return types.ModeRun
	}
	return e.mode
}

// applyOutcome maps a successful agent run's result envelope onto the todo:
// the status transition, plan bookkeeping, summary/questions persistence, LLM
// metrics, and the attempt record. It replaces the old always-completed
// frontmatter update.
//
//	plan:  plan new/updated → review (plan file validated), unchanged → pending,
//	       ask → ask
//	run:   ask → ask, verified/unverified from an executed definition of done,
//	       otherwise pending. Closing is a separate human workflow transition.
func (e *TODOExecutor) applyOutcome(ctx context.Context, todo *types.TODO, result *ExecutionResult) error {
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	if strings.TrimSpace(result.Runtime.ResolvedModel) != "" {
		todo.LLM.Model = result.Runtime.ResolvedModel
	}
	todo.LLM.TokensUsed = result.TokensUsed
	todo.LLM.CostIncurred = result.CostUSD

	todo.Attempts++
	now := time.Now()
	mode := e.Mode()
	update := StateUpdate{Attempts: &todo.Attempts, LastRun: &now, RunMode: &mode}
	if result.Summary != "" {
		update.LastRunSummary = &result.Summary
	}
	// Questions persist only for an ask outcome; any other outcome clears stale ones.
	questions := result.Questions
	if result.EndStatus != types.EndAsk {
		questions = nil
	}
	update.Questions = &questions

	switch {
	case mode == types.ModePlan:
		if err := e.applyPlanOutcome(todo, result, &update); err != nil {
			return err
		}
	case result.EndStatus == types.EndAsk:
		todo.Status = types.StatusAsk
	case result.DoD != nil && result.DoD.Ran:
		// The run iterated against its definition-of-done fixture: verified when
		// every check passed within the budget, unverified when it ran out with a
		// check still red.
		if result.DoD.Passed {
			todo.Status = types.StatusVerified
		} else {
			todo.Status = types.StatusUnverified
		}
	default:
		todo.Status = types.StatusPending
	}
	update.Status = &todo.Status

	if err := e.saveAttempt(ctx, todo, result); err != nil {
		return fmt.Errorf("persist TODO attempt: %w", err)
	}
	e.updateProviderState(ctx, todo, update)
	return nil
}

// applyPlanOutcome handles the plan-mode envelope: validate the agent's native
// plan file and pick the review/pending/ask transition.
func (e *TODOExecutor) applyPlanOutcome(todo *types.TODO, result *ExecutionResult, update *StateUpdate) error {
	if result.EndStatus == types.EndAsk {
		todo.Status = types.StatusAsk
		return nil
	}
	if result.Plan == nil {
		return fmt.Errorf("plan run finished without a plan definition")
	}
	switch result.Plan.Status {
	case types.PlanUnchanged:
		// The existing plan (and its recorded path) still stands: ready to execute.
		todo.Status = types.StatusPending
		update.PlanStatus = &result.Plan.Status
	case types.PlanNew, types.PlanUpdated:
		path := result.Plan.Path
		if err := ValidatePlanFile(path); err != nil {
			// The agent misreported its plan file; captain's session plan
			// resolver is the canonical fallback before failing the run.
			resolvedPath, resolvedContent := ResolveSessionPlan(todo)
			switch {
			case resolvedPath != "" && ValidatePlanFile(resolvedPath) == nil:
				path = resolvedPath
			case strings.TrimSpace(result.Plan.Content) != "":
				path = ""
			case strings.TrimSpace(resolvedContent) != "":
				path = ""
			default:
				return fmt.Errorf("plan run reported an invalid plan file and no inline plan content: %w", err)
			}
		}
		todo.Status = types.StatusReview
		update.PlanStatus = &result.Plan.Status
		update.PlanPath = &path
	default:
		return fmt.Errorf("plan run reported unknown plan.status %q", result.Plan.Status)
	}
	return nil
}

// Resume resumes the todo's prior agent session with a user message — the
// answer to the agent's questions — and applies the same envelope-driven
// transitions as a fresh run. The executor must support session resumption.
func (e *TODOExecutor) Resume(ctx *ExecutorContext, todosInGroup []*types.TODO, message string) ([]*ExecutionResult, error) {
	fb, ok := e.executor.(FeedbackExecutor)
	if !ok {
		return nil, fmt.Errorf("executor %s cannot resume a session", e.executor.Name())
	}
	if lifecycle, ok := e.activeProvider().(RunLifecycleProvider); ok {
		for _, todo := range todosInGroup {
			if err := lifecycle.PrepareRun(ctx, todo, RunPreparation{
				Mode: e.Mode(), ExecutorName: e.executor.Name(), Resume: true,
				// A resumed turn dispatches the answer as its user prompt; the rest
				// of the spec is the session's, already persisted by the run that
				// opened it.
				Spec: api.Spec{Prompt: api.Prompt{User: message}},
			}); err != nil {
				return nil, fmt.Errorf("prepare resumed native TODO run: %w", err)
			}
		}
	}
	// A resumed turn records its run start the same way a fresh one does: from
	// the executor, once it has resolved the runtime it is actually using.
	ctx.SetSessionIDHook(e.sessionIDPersister(ctx, todosInGroup))
	ctx.SetRunStartHook(e.runStartPersister(ctx, todosInGroup))
	now := time.Now()
	for _, todo := range todosInGroup {
		todo.Status = types.StatusInProgress
		todo.LastRun = &now
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, LastRun: &now})
	}

	groupResult, err := fb.SendFeedback(ctx, todosInGroup, message)
	if err != nil {
		var results []*ExecutionResult
		for _, todo := range todosInGroup {
			todo.Status = types.StatusFailed
			todo.Attempts++
			perTodo := groupResult
			if groupResult != nil {
				perTodo = e.splitResult(groupResult, len(todosInGroup))
				if saveErr := e.saveAttempt(ctx, todo, perTodo); saveErr != nil {
					fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
				}
			}
			e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
			results = append(results, perTodo)
		}
		return results, err
	}

	var results []*ExecutionResult
	for _, todo := range todosInGroup {
		perTodo := e.splitResult(groupResult, len(todosInGroup))
		if applyErr := e.applyOutcome(ctx, todo, perTodo); applyErr != nil {
			return results, applyErr
		}
		results = append(results, perTodo)
	}
	return results, nil
}
