package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// EventLifecycleOutcome is the history event kind OnOutcome records: which
// step ran, which status its outcome chose, and the run's comment.
const EventLifecycleOutcome = "lifecycle_outcome"

// OutcomeErrorKey is the lifecycle_outcome payload key naming why a finished
// run's outcome could not be persisted. Its event's status is always failed.
const OutcomeErrorKey = "error"

// OnOutcome applies a finished step to the todo. It is the single place a run
// writes a status: the attempt is persisted (which, for a plan step, is where
// the plan revision lands), the outcome's status is written unless the step
// kept it, and one lifecycle_outcome event records the run — its comment is the
// envelope's, and its payload lists the writes the run's tools made.
//
// The tools wrote while the run was going, so nothing is applied here. A step
// that keeps the status still records the one its tools left the todo in.
func (h *Host) OnOutcome(ctx context.Context, todo *types.TODO, step lifecycle.Step, outcome *StepOutcome, status string) error {
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
	if status != lifecycle.OutcomeKeep && !validStatuses[status] {
		return fmt.Errorf("lifecycle outcome for step %s: %q is not a todo status", step.Name, status)
	}
	execution := outcome.Execution
	persistCtx, cancel := todos.PersistenceContext(ctx)
	defer cancel()

	recorded := status
	if status == lifecycle.OutcomeKeep && outcome.StartStatus != "" && todo.Status != outcome.StartStatus {
		recorded = string(todo.Status)
	}
	update := h.stateUpdate(todo, step, execution, status)
	if err := h.Provider.SaveAttempt(persistCtx, todo, execution); err != nil {
		return h.unpersistedOutcome(persistCtx, events, todo, step, outcome, fmt.Errorf("persist TODO attempt: %w", err))
	}
	if err := h.Provider.UpdateState(persistCtx, todo, update); err != nil {
		return h.unpersistedOutcome(persistCtx, events, todo, step, outcome, fmt.Errorf("update TODO state: %w", err))
	}
	if err := events.AppendEvent(persistCtx, todo, outcomeEvent(step, outcome, recorded, nil)); err != nil {
		return fmt.Errorf("record lifecycle outcome: %w", err)
	}
	return nil
}

// unpersistedOutcome records a finished run whose outcome could not be
// persisted as a failed outcome, carrying the reason, and returns that reason.
// The run may well have succeeded in Captain — the failure is gavel's, after
// it — so this event is the only record of why the attempt did not land; the
// todo projects failed from it, and the session detail shows its error.
//
// The todo is reloaded first: the write that failed may have left it at a
// version the store has moved past, or with a status it never reached.
func (h *Host) unpersistedOutcome(ctx context.Context, events todos.EventProvider, todo *types.TODO, step lifecycle.Step, outcome *StepOutcome, persistErr error) error {
	current, err := h.Provider.Get(ctx, todo.ID)
	if err != nil {
		return errors.Join(persistErr, fmt.Errorf("reload TODO to record its unpersisted outcome: %w", err))
	}
	*todo = *current
	if err := events.AppendEvent(ctx, todo, outcomeEvent(step, outcome, string(types.StatusFailed), persistErr)); err != nil {
		return errors.Join(persistErr, fmt.Errorf("record unpersisted lifecycle outcome: %w", err))
	}
	return persistErr
}

// stateUpdate is everything the run changes about the todo besides the
// attempt itself: the LLM metrics, the summary the dashboard shows, the plan
// bookkeeping, and — unless the step kept it — the status.
func (h *Host) stateUpdate(todo *types.TODO, step lifecycle.Step, execution *todos.ExecutionResult, status string) todos.StateUpdate {
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
	if execution.Plan != nil && status != lifecycle.OutcomeKeep {
		planStatus := execution.Plan.Status
		update.PlanStatus = &planStatus
		if status == string(types.StatusReview) {
			path := execution.Plan.Path
			update.PlanPath = &path
		}
	}
	if status != lifecycle.OutcomeKeep {
		next := types.Status(status)
		todo.Status = next
		update.Status = &next
	}
	return update
}

// outcomeEvent is the run's single history entry: the step and the status it
// chose, the run's comment, the definition-of-done line when verification ran,
// and any error. persistErr, when set, is why the outcome was not persisted.
func outcomeEvent(step lifecycle.Step, outcome *StepOutcome, status string, persistErr error) todos.Event {
	execution := outcome.Execution
	body := fmt.Sprintf("**Lifecycle:** `%s` → `%s`", step.Name, status)
	if summary := strings.TrimSpace(execution.Summary); summary != "" {
		body += "\n\n" + summary
	}
	if line := verifyLine(outcome.Result.Verify); line != "" && classOf(step) != types.ModeVerify {
		body += "\n\n" + line
	}
	if execution.ErrorMessage != "" && outcome.Result.Run.State != RunSucceeded {
		body += "\n\n```text\n" + strings.TrimSpace(execution.ErrorMessage) + "\n```"
	}
	if persistErr != nil {
		body += "\n\n**Outcome not persisted:**\n\n```text\n" + strings.TrimSpace(persistErr.Error()) + "\n```"
	}
	payload := outcomePayload(step, outcome, status)
	if persistErr != nil {
		payload[OutcomeErrorKey] = persistErr.Error()
	}
	return todos.Event{Kind: EventLifecycleOutcome, Body: body, Payload: payload}
}

func outcomePayload(step lifecycle.Step, outcome *StepOutcome, status string) map[string]any {
	execution := outcome.Execution
	envelope := map[string]any{"success": false, "event": ""}
	if execution.Envelope != nil {
		envelope = map[string]any{"success": execution.Envelope.Success, "event": execution.Envelope.Event}
	}
	writes := execution.ToolWrites
	if writes == nil {
		writes = []todos.ToolWrite{}
	}
	payload := map[string]any{
		"step":   step.Name,
		"status": status,
		"run": map[string]any{
			"state": outcome.Result.Run.State, "error": outcome.Result.Run.Error,
			"iterations": outcome.Result.Run.Iterations, "costUsd": outcome.Result.Run.CostUSD,
		},
		"envelope": envelope,
		"verify": map[string]any{
			"ran":    outcome.Result.Verify != nil && outcome.Result.Verify.Ran,
			"passed": outcome.Result.Verify != nil && outcome.Result.Verify.Passed,
		},
		"tools": writes,
	}
	if outcome.Admission.PromptRunID.String() != "00000000-0000-0000-0000-000000000000" {
		payload["promptRunId"] = outcome.Admission.PromptRunID.String()
	}
	if outcome.Source != "" {
		payload["source"] = outcome.Source
	}
	if execution.Plan != nil {
		payload["plan"] = map[string]any{"status": string(execution.Plan.Status), "path": execution.Plan.Path}
	}
	return payload
}

// verifyLine is the definition-of-done verdict a run's history entry carries
// when verification ran: passed or failed, and how much of the checklist held.
func verifyLine(report *api.VerifyReport) string {
	if report == nil || !report.Ran {
		return ""
	}
	verdict := "failed"
	if report.Passed {
		verdict = "passed"
	}
	line := "**Verification:** " + verdict
	if len(report.Checklist) == 0 {
		return line
	}
	held := 0
	for _, item := range report.Checklist {
		if item.Passed != nil && *item.Passed {
			held++
		}
	}
	return fmt.Sprintf("%s, checklist %d/%d", line, held, len(report.Checklist))
}
