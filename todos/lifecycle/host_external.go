package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// ExternalTurn is a turn a step run's own provider session took without gavel
// dispatching it — the user answered the run's questions from Captain's session
// page, which resumed the same session. Gavel learns of it only afterwards, from
// what Captain recorded.
type ExternalTurn struct {
	// Source names who ran the turn; it lands in the lifecycle_outcome payload.
	Source string
	// PromptRunID is the gavel prompt run the turn continued, and settles.
	PromptRunID uuid.UUID
	// State is how the turn ended: RunSucceeded, RunFailed or RunCancelled.
	State   string
	Summary string
	Error   string
	// Envelope is the structured result the turn returned, nil when it returned
	// only text.
	Envelope map[string]any
}

// SettleExternalTurn classifies an external turn with the step's own outcomes
// and applies it through OnOutcome, so the todo lands exactly where a turn gavel
// dispatched would have left it. A turn the outcomes cannot classify is still
// recorded — under the status the todo already had — and the classification
// error is returned, as a dispatched run's is.
func (h *Host) SettleExternalTurn(ctx context.Context, todo *types.TODO, stepName string, turn ExternalTurn) error {
	if h.Def == nil {
		return fmt.Errorf("lifecycle host: no lifecycle loaded")
	}
	step, ok := h.Def.Definition().Step(stepName)
	if !ok {
		return fmt.Errorf("step %q is not part of lifecycle %s", stepName, h.Def.Definition().Name)
	}
	definition, err := h.promptFor(step)
	if err != nil {
		return &ConfigurationError{Err: err}
	}
	lc, err := h.Context(ctx, todo)
	if err != nil {
		return err
	}
	outcome, err := h.externalOutcome(step, definition, turn)
	if err != nil {
		return err
	}
	status, classifyErr := h.Def.Outcome(step, lc, outcome.Result)
	if classifyErr != nil {
		status = OutcomeKeep
	}
	outcome.Status = status
	return errors.Join(classifyErr, h.OnOutcome(ctx, todo, step, outcome, status))
}

// externalOutcome builds the facts and execution record for an external turn
// through the same collection a dispatched step goes through.
func (h *Host) externalOutcome(step Step, definition todoprompt.Definition, turn ExternalTurn) (*StepOutcome, error) {
	source := strings.TrimSpace(turn.Source)
	if source == "" {
		return nil, fmt.Errorf("external turn for step %s: source is required", step.Name)
	}
	if turn.PromptRunID == uuid.Nil {
		return nil, fmt.Errorf("external turn for step %s: prompt run is required", step.Name)
	}
	execution := &todos.ExecutionResult{ExecutorName: source}
	facts := StepResult{}
	switch turn.State {
	case RunSucceeded:
		if turn.Envelope == nil {
			execution.Summary, execution.EndStatus = strings.TrimSpace(turn.Summary), types.EndCompleted
			facts.Run.State = RunSucceeded
			facts.Envelope = Envelope{Summary: execution.Summary, EndStatus: string(types.EndCompleted)}
			break
		}
		prepared := &preparedStep{definition: definition, class: definition.Class}
		result := promptrun.Result{Response: &api.Response{StructuredData: turn.Envelope}, StructuredData: turn.Envelope}
		h.collectEnvelope(execution, &facts, prepared, result)
	case RunFailed, RunCancelled:
		reason := firstNonEmpty(strings.TrimSpace(turn.Error), fmt.Sprintf("the %s turn ended %s", source, turn.State))
		execution.ErrorMessage, execution.Cancelled = reason, turn.State == RunCancelled
		if execution.Cancelled {
			execution.Summary = reason
		}
		facts.Run.State, facts.Run.Error = turn.State, reason
	default:
		return nil, fmt.Errorf("external turn for step %s: state %q is not a finished turn", step.Name, turn.State)
	}
	settleRunFacts(execution, &facts)
	return &StepOutcome{
		Step: step, Result: facts, Execution: execution, Source: source,
		Admission: todos.RunPreparationResult{PromptRunID: turn.PromptRunID},
	}, nil
}
