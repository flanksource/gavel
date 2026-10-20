package lifecycle

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
)

// Run states the outcome predicates read from `run.state`.
const (
	RunSucceeded = "succeeded"
	RunFailed    = "failed"
	RunCancelled = "cancelled"
	RunWaiting   = "waiting"
)

// errAskThroughTool is why a run whose agent asked natively fails: every todo
// prompt denies AskUserQuestion, so a native question means the run escaped its
// own permissions rather than that it is waiting on an answer.
const errAskThroughTool = "todo prompts ask through todo_ask; the agent asked natively instead"

// collect folds what came back from captain into the two records a finished
// step needs: the facts the outcome predicates read, and the execution result
// the provider persists as the attempt.
func (h *Host) collect(step Step, prepared *preparedStep, d dispatched, start time.Time) *StepOutcome {
	execution := d.execution
	execution.Duration = time.Since(start)
	out := d.out
	if out.Response != nil {
		execution.Workspace = api.NewWorkspaceRecord(out.Response.Workspace)
	}
	execution.CostUSD = out.CostUSD
	execution.TokensUsed = out.Usage.TotalTokens()
	if out.Loop != nil {
		execution.NumTurns = len(out.Loop.Iterations)
	}
	if declaresVerify(prepared.request.Workflow) {
		// Ran follows the report: a report captain produced for a verifier that
		// never executed says so itself, and a DoD that claimed to have run on the
		// strength of the report's existence read as a verdict nobody reached.
		ran := out.Report != nil && out.Report.Ran
		execution.DoD = &todos.DoDOutcome{Ran: ran, Passed: ran && out.Passed && len(out.Verdicts) > 0, Report: out.Report}
	}
	if d.tools != nil {
		execution.ToolWrites = d.tools.Writes()
	}
	facts := StepResult{
		Run:    RunFacts{Iterations: execution.NumTurns, CostUSD: execution.CostUSD, StopReason: stopReason(out.Loop)},
		Verify: out.Report,
	}
	switch {
	case d.cancelled:
		execution.Cancelled = true
		execution.ErrorMessage = todos.ErrExecutionCancelled.Error()
		execution.Summary = todos.ErrExecutionCancelled.Error()
		facts.Run.State, facts.Run.Error = RunCancelled, execution.ErrorMessage
	case d.err != nil && !d.timedOut:
		execution.ErrorMessage = d.err.Error()
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
	case d.timedOut:
		execution.ErrorMessage = fmt.Sprintf("%s run did not complete within %s", execution.ExecutorName, prepared.timeout)
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
	case prepared.class == types.ModeVerify:
		collectVerify(execution, &facts, out.Report, out.Passed)
	default:
		collectEnvelope(execution, &facts, step, prepared, out, d.tools)
	}
	settleRunFacts(execution, &facts)
	return &StepOutcome{Step: step, Result: facts, Execution: execution}
}

// settleRunFacts reconciles the run's fate with its execution record once both
// are collected.
//
// The runtime records a prompt run as failed whenever the run reported an
// error — a provider result that was not a success, an error event — even when
// an envelope was still decoded. The facts the outcomes read must say the
// same, or a run captain stores as failed could land the todo in pending. The
// attempt succeeded only when the run did and the agent claimed success.
func settleRunFacts(execution *todos.ExecutionResult, facts *StepResult) {
	if facts.Run.State == RunSucceeded && execution.ErrorMessage != "" {
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
	}
	execution.Success = facts.Run.State == RunSucceeded && execution.Envelope != nil && execution.Envelope.Success
}

// stopReason is why captain's generate loop ended — condition-met,
// max-iterations, max-cost or error — or empty for a verify-only run.
func stopReason(loop *captainai.LoopResult) string {
	if loop == nil {
		return ""
	}
	return loop.StopReason
}

// collectVerify is a verify-only step's result: the report is the verdict, and
// it is shaped into the same envelope an agent step returns, so its history
// entry reads like every other run's.
func collectVerify(execution *todos.ExecutionResult, facts *StepResult, report *api.VerifyReport, passed bool) {
	if report == nil {
		execution.ErrorMessage = "verification produced no report"
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
		return
	}
	facts.Run.State = RunSucceeded
	env := types.Envelope{Success: passed, Event: VerifyEvents[0], Comment: verifyComment(report, passed)}
	if !passed {
		env.Event = VerifyEvents[1]
	}
	setEnvelope(execution, facts, env, nil)
}

func verifyComment(report *api.VerifyReport, passed bool) string {
	reason := strings.TrimSpace(report.Reason)
	if reason == "" {
		reason = "Definition of done passed."
		if !passed {
			reason = "Definition of done failed."
		}
	}
	var lines []string
	for _, item := range report.Checklist {
		mark := "[ ]"
		if item.Passed != nil && *item.Passed {
			mark = "[x]"
		}
		lines = append(lines, fmt.Sprintf("- %s %s", mark, strings.TrimSpace(item.Item)))
	}
	if len(lines) == 0 {
		return reason
	}
	return reason + "\n\n" + strings.Join(lines, "\n")
}

func setEnvelope(execution *todos.ExecutionResult, facts *StepResult, env types.Envelope, extra map[string]any) {
	execution.Envelope = &env
	execution.Summary = strings.TrimSpace(env.Comment)
	facts.Envelope = Envelope{Success: env.Success, Event: env.Event, Comment: env.Comment, Extra: extra}
}

// collectEnvelope decodes the agent's structured result and lifts it onto both
// records, with the plan a plan-mode run produced.
func collectEnvelope(execution *todos.ExecutionResult, facts *StepResult, step Step, prepared *preparedStep, out promptrun.Result, run *tools.Run) {
	// Captured before the decode, because the decode is what loses it: an agent
	// that answered well in the wrong shape would otherwise leave nothing but
	// "failed" for a reader to go on.
	execution.ResponseText = responseText(out.Response)
	result, err := decodeEnvelope(step, prepared, out.Response)
	if err != nil {
		execution.ErrorMessage = err.Error()
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
		return
	}
	if out.Response.TerminalOutcome == nil {
		output, err := captainai.ParseStructured(execution.ResponseText, func(value *map[string]any) error {
			if (*value)["event"] != result.envelope.Event {
				return fmt.Errorf("output event does not match the validated envelope")
			}
			return nil
		})
		if err != nil {
			execution.ErrorMessage = fmt.Sprintf("decode emitted structured output: %v", err)
			facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
			return
		}
		execution.OutputJSON = *output
	}
	facts.Run.State = RunSucceeded
	setEnvelope(execution, facts, result.envelope, structuredFields(out.StructuredData))
	if prepared.class == types.ModePlan {
		execution.Plan = planResult(result, run, prepared.existingPlan)
	}
}

// decoded is a run's envelope with the native plan outcome it came from, when
// the agent ended by presenting a plan rather than with structured output.
type decoded struct {
	envelope types.Envelope
	plan     *api.TerminalPlan
}

// planResult is the plan a plan-mode run produced: the native plan outcome's,
// else the content todo_plan recorded. Its status is the envelope's event when
// the step speaks the plan vocabulary, and otherwise what comparing the content
// with the existing plan shows. A run that produced no content and no plan
// status produced no plan.
func planResult(d decoded, run *tools.Run, existing string) *types.PlanResult {
	plan := &types.PlanResult{}
	if d.plan != nil {
		plan.Path, plan.Content = d.plan.Path, d.plan.Content
	} else if run != nil {
		plan.Content = run.Plan()
	}
	switch {
	case types.KnownPlanStatus(d.envelope.Event):
		plan.Status = types.PlanStatus(d.envelope.Event)
	case strings.TrimSpace(plan.Content) != "":
		plan.Status = nativePlanStatus(existing, plan.Content)
	default:
		return nil
	}
	return plan
}

// responseText is whatever the agent actually returned, in the same precedence
// order decodeEnvelope reads it: the structured payload when there is one, else
// the plain text. It is best-effort — a payload that cannot even be re-encoded
// contributes nothing rather than failing a run that already produced a result.
func responseText(response *api.Response) string {
	if response == nil {
		return ""
	}
	if response.StructuredData != nil {
		if text, err := structuredDataText(response.StructuredData); err == nil {
			return text
		}
	}
	return strings.TrimSpace(response.Text)
}

// decodeEnvelope resolves the response contract in precedence order: native
// terminal outcome, structured data, response text.
func decodeEnvelope(step Step, prepared *preparedStep, response *api.Response) (decoded, error) {
	if response == nil {
		return decoded{}, fmt.Errorf("no result envelope: agent response is missing")
	}
	if response.TerminalOutcome != nil {
		return envelopeFromTerminalOutcome(step, prepared, response.TerminalOutcome)
	}
	if response.StructuredData != nil {
		text, err := structuredDataText(response.StructuredData)
		if err != nil {
			return decoded{}, fmt.Errorf("result envelope structured data: %w", err)
		}
		env, err := ParseEnvelope(step.Events, text)
		if err != nil {
			return decoded{}, fmt.Errorf("result envelope structured data: %w", err)
		}
		return decoded{envelope: *env}, nil
	}
	if strings.TrimSpace(response.Text) != "" {
		env, err := ParseEnvelope(step.Events, response.Text)
		if err != nil {
			return decoded{}, fmt.Errorf("result envelope response text: %w", err)
		}
		return decoded{envelope: *env}, nil
	}
	return decoded{}, fmt.Errorf("no result envelope: terminal outcome, structured data, and response text are empty")
}

// ParseEnvelope decodes a run's final result and holds its event to the step's
// declared vocabulary.
func ParseEnvelope(events []string, text string) (*types.Envelope, error) {
	return captainai.ParseStructured(text, func(env *types.Envelope) error { return env.Validate(events) })
}

// envelopeFromTerminalOutcome maps a native outcome onto the envelope. A
// native plan is a plan run's result; a native question is a failure, because
// every todo prompt asks through todo_ask.
func envelopeFromTerminalOutcome(step Step, prepared *preparedStep, outcome *api.TerminalOutcome) (decoded, error) {
	if err := outcome.Validate(); err != nil {
		return decoded{}, fmt.Errorf("invalid terminal outcome: %w", err)
	}
	switch outcome.Kind {
	case api.TerminalOutcomePlan:
		if prepared.class != types.ModePlan {
			return decoded{}, fmt.Errorf("native plan outcome is invalid in %s mode", prepared.class)
		}
		status := nativePlanStatus(prepared.existingPlan, outcome.Plan.Content)
		env := types.Envelope{Success: true, Event: string(status), Comment: nativePlanSummary(status)}
		if err := env.Validate(step.Events); err != nil {
			return decoded{}, fmt.Errorf("native plan outcome: %w", err)
		}
		return decoded{envelope: env, plan: outcome.Plan}, nil
	case api.TerminalOutcomeQuestions:
		return decoded{}, fmt.Errorf("%s", errAskThroughTool)
	default:
		return decoded{}, fmt.Errorf("unsupported terminal outcome kind %q", outcome.Kind)
	}
}

func nativePlanStatus(existing, content string) types.PlanStatus {
	existing = normalizePlanContent(existing)
	if existing == "" {
		return types.PlanNew
	}
	if existing == normalizePlanContent(content) {
		return types.PlanUnchanged
	}
	return types.PlanUpdated
}

func nativePlanSummary(status types.PlanStatus) string {
	switch status {
	case types.PlanNew:
		return "The agent created a plan."
	case types.PlanUpdated:
		return "The agent updated the plan."
	default:
		return "The existing plan is unchanged."
	}
}

func normalizePlanContent(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	return strings.TrimSpace(content)
}

func structuredDataText(value any) (string, error) {
	var data []byte
	switch typed := value.(type) {
	case json.RawMessage:
		data = typed
	case []byte:
		data = typed
	case string:
		data = []byte(typed)
	default:
		var err error
		data, err = json.Marshal(value)
		if err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(string(data)) == "" || string(data) == "null" {
		return "", fmt.Errorf("value is empty")
	}
	return string(data), nil
}

// structuredFields is the raw structured output as a map, so a custom prompt's
// outcomes can read fields the envelope does not name.
func structuredFields(value any) map[string]any {
	if value == nil {
		return nil
	}
	text, err := structuredDataText(value)
	if err != nil {
		return nil
	}
	var fields map[string]any
	if json.Unmarshal([]byte(text), &fields) != nil {
		return nil
	}
	return fields
}
