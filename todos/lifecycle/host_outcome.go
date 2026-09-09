package lifecycle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// EventLifecycleOutcome is the history event kind OnOutcome records: which
// step ran and which status its outcome chose.
const EventLifecycleOutcome = "lifecycle_outcome"

// OnOutcome applies a finished step to the todo. It is the single place a run
// writes a status: the attempt is persisted (which, for a plan step, is where
// the plan revision lands), the outcome's status is written unless the step
// kept it, and one lifecycle_outcome event records the decision.
//
// A triage verdict is applied first — the agent ran read-only, so this is where
// its edits take effect — and its own status assignment, if any, survives
// because the triage step's outcome is `keep`.
func (h *Host) OnOutcome(ctx context.Context, todo *types.TODO, step Step, outcome *StepOutcome, status string) error {
	if h.Provider == nil {
		return fmt.Errorf("lifecycle outcome: host has no provider")
	}
	if outcome == nil || outcome.Execution == nil {
		return fmt.Errorf("lifecycle outcome for step %s: no execution result", step.Name)
	}
	events, ok := h.Provider.(todos.EventProvider)
	if !ok {
		return fmt.Errorf("lifecycle outcome: provider %T cannot record events", h.Provider)
	}
	if status != OutcomeKeep && !validStatuses[status] {
		return fmt.Errorf("lifecycle outcome for step %s: %q is not a todo status", step.Name, status)
	}
	execution := outcome.Execution
	persistCtx, cancel := todos.PersistenceContext(ctx)
	defer cancel()

	// recorded is the status the event names: the outcome's, or — when the step
	// kept it and the triage verdict assigned one — the status actually written.
	// An event that said `keep` while the todo moved to completed would record a
	// transition that never happened.
	recorded := status
	if execution.Triage != nil {
		before := todo.Status
		if err := todos.ApplyTriage(persistCtx, h.Provider, todo, execution.Triage, todos.TriageOptions{WorkDir: h.WorkDir}); err != nil {
			return err
		}
		if status == OutcomeKeep && todo.Status != before {
			recorded = string(todo.Status)
		}
	}
	update := h.stateUpdate(todo, step, execution, status)
	if err := h.Provider.SaveAttempt(persistCtx, todo, execution); err != nil {
		return fmt.Errorf("persist TODO attempt: %w", err)
	}
	if err := h.Provider.UpdateState(persistCtx, todo, update); err != nil {
		return fmt.Errorf("update TODO state: %w", err)
	}
	if err := events.AppendEvent(persistCtx, todo, outcomeEvent(step, outcome, recorded)); err != nil {
		return fmt.Errorf("record lifecycle outcome: %w", err)
	}
	return nil
}

// stateUpdate is everything the run changes about the todo besides the
// attempt itself: the LLM metrics, the summary and questions the dashboard
// shows, the plan bookkeeping, and — unless the step kept it — the status.
func (h *Host) stateUpdate(todo *types.TODO, step Step, execution *todos.ExecutionResult, status string) todos.StateUpdate {
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	if model := strings.TrimSpace(execution.Runtime.ResolvedModel); model != "" {
		todo.LLM.Model = model
	}
	todo.LLM.TokensUsed = execution.TokensUsed
	todo.LLM.CostIncurred = execution.CostUSD

	// The update carries its own copy of the attempt count: SaveAttempt runs
	// before UpdateState and a provider that reloads the todo there would
	// otherwise overwrite the value a pointer into the todo still refers to.
	attempts := todo.Attempts + 1
	todo.Attempts = attempts
	now := time.Now()
	class := classOf(step)
	update := todos.StateUpdate{Attempts: &attempts, LastRun: &now, RunMode: &class}
	if execution.Summary != "" {
		update.LastRunSummary = &execution.Summary
	}
	// Questions persist only for an ask outcome; any other outcome clears stale ones.
	questions := execution.Questions
	if execution.EndStatus != types.EndAsk {
		questions = nil
	}
	update.Questions = &questions
	if execution.Plan != nil && status != OutcomeKeep {
		planStatus := execution.Plan.Status
		update.PlanStatus = &planStatus
		if status == string(types.StatusReview) {
			path := execution.Plan.Path
			update.PlanPath = &path
		}
	}
	if status != OutcomeKeep {
		next := types.Status(status)
		todo.Status = next
		update.Status = &next
	}
	return update
}

func outcomeEvent(step Step, outcome *StepOutcome, status string) todos.Event {
	execution := outcome.Execution
	body := fmt.Sprintf("**Lifecycle:** `%s` → `%s`", step.Name, status)
	if summary := strings.TrimSpace(execution.Summary); summary != "" {
		body += "\n\n" + summary
	}
	if execution.ErrorMessage != "" && outcome.Result.Run.State != RunSucceeded {
		body += "\n\n```text\n" + strings.TrimSpace(execution.ErrorMessage) + "\n```"
	}
	payload := map[string]any{
		"step":   step.Name,
		"status": status,
		"run": map[string]any{
			"state": outcome.Result.Run.State, "error": outcome.Result.Run.Error,
			"iterations": outcome.Result.Run.Iterations, "costUsd": outcome.Result.Run.CostUSD,
		},
		"envelope": map[string]any{"endStatus": string(execution.EndStatus)},
		"verify": map[string]any{
			"ran":    outcome.Result.Verify != nil && outcome.Result.Verify.Ran,
			"passed": outcome.Result.Verify != nil && outcome.Result.Verify.Passed,
		},
	}
	if outcome.Admission.PromptRunID.String() != "00000000-0000-0000-0000-000000000000" {
		payload["promptRunId"] = outcome.Admission.PromptRunID.String()
	}
	if execution.Plan != nil {
		payload["plan"] = map[string]any{"status": string(execution.Plan.Status), "path": execution.Plan.Path}
	}
	return todos.Event{Kind: EventLifecycleOutcome, Body: body, Payload: payload}
}
