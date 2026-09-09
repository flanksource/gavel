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
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
)

// Run states the outcome predicates read from `run.state`.
const (
	RunSucceeded = "succeeded"
	RunFailed    = "failed"
	RunCancelled = "cancelled"
	RunWaiting   = "waiting"
)

// collect folds what came back from captain into the two records a finished
// step needs: the facts the outcome predicates read, and the execution result
// the provider persists as the attempt.
func (h *Host) collect(exec *todos.ExecutorContext, todo *types.TODO, step Step, prepared *preparedStep, d dispatched, start time.Time) *StepOutcome {
	execution := d.execution
	execution.Duration = time.Since(start)
	out := d.out
	if out.Response != nil && out.Response.Workspace != nil {
		// Flushed here rather than from the hooks that produced them: a hook firing
		// mid-turn cannot know the transcript session's id, because that row only
		// exists once the monitor has ingested the provider's log.
		exec.RecordNotices(execution.Runtime.SessionID, out.Response.Workspace.Notices)
		if commits := out.Response.Workspace.Commits; len(commits) > 0 {
			execution.CommitSHA = commits[len(commits)-1].SHA
		}
	}
	if out.CostUSD > 0 {
		execution.CostUSD = out.CostUSD
	}
	if tokens := out.Usage.TotalTokens(); tokens > 0 {
		execution.TokensUsed = tokens
	}
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
		h.collectVerify(execution, &facts, out.Report, out.Passed)
	default:
		h.collectEnvelope(execution, &facts, prepared, out)
	}
	// The runtime records a prompt run as failed whenever the run reported an
	// error — a provider result that was not a success, an error event, an
	// envelope that says failed — even when an envelope was still decoded. The
	// facts the outcomes read must say the same, or a run captain stores as
	// failed could land the todo in pending.
	if facts.Run.State == RunSucceeded && execution.ErrorMessage != "" {
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
	}
	execution.Success = facts.Run.State == RunSucceeded && execution.EndStatus != types.EndFailed
	return &StepOutcome{Step: step, Result: facts, Execution: execution}
}

// stopReason is why captain's generate loop ended — condition-met,
// max-iterations, max-cost or error — or empty for a verify-only run.
func stopReason(loop *captainai.LoopResult) string {
	if loop == nil {
		return ""
	}
	return loop.StopReason
}

// collectVerify is a verify-only step's result: the report is the verdict.
func (h *Host) collectVerify(execution *todos.ExecutionResult, facts *StepResult, report *api.VerifyReport, passed bool) {
	if report == nil {
		execution.ErrorMessage = "verification produced no report"
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
		return
	}
	facts.Run.State = RunSucceeded
	execution.Summary = report.Reason
	execution.EndStatus = types.EndCompleted
	if !passed {
		execution.Summary = firstNonEmpty(report.Reason, "definition of done failed")
	}
	facts.Envelope = Envelope{Summary: execution.Summary, EndStatus: string(execution.EndStatus)}
}

// collectEnvelope decodes the agent's structured result into the envelope its
// prompt promised and lifts it onto both records.
func (h *Host) collectEnvelope(execution *todos.ExecutionResult, facts *StepResult, prepared *preparedStep, out promptrun.Result) {
	env, err := decodeEnvelope(prepared, out.Response)
	if err != nil {
		execution.ErrorMessage = err.Error()
		facts.Run.State, facts.Run.Error = RunFailed, execution.ErrorMessage
		return
	}
	execution.Summary = env.Summary
	execution.EndStatus = env.EndStatus
	execution.Questions = env.Questions
	execution.Plan = env.Plan
	execution.Triage = env.Triage
	facts.Envelope = Envelope{Summary: env.Summary, EndStatus: string(env.EndStatus), Extra: structuredFields(out.StructuredData)}
	facts.Questions = questionVars(env.Questions)
	if env.Plan != nil {
		facts.Plan = &PlanFacts{Status: string(env.Plan.Status), Path: env.Plan.Path, Content: env.Plan.Content}
	}
	switch env.EndStatus {
	case types.EndAsk:
		facts.Run.State = RunWaiting
	case types.EndFailed:
		facts.Run.State = RunSucceeded
		execution.ErrorMessage = firstNonEmpty(execution.ErrorMessage, env.Summary)
	default:
		facts.Run.State = RunSucceeded
	}
}

type envelope struct {
	types.ResultEnvelope
	Plan   *types.PlanResult
	Triage *types.TriageEnvelope
}

// decodeEnvelope resolves the response contract in precedence order: native
// terminal outcome, structured data, response text.
func decodeEnvelope(prepared *preparedStep, response *api.Response) (*envelope, error) {
	if response == nil {
		return nil, fmt.Errorf("no result envelope: agent response is missing")
	}
	if response.TerminalOutcome != nil {
		return envelopeFromTerminalOutcome(prepared, response.TerminalOutcome)
	}
	if response.StructuredData != nil {
		text, err := structuredDataText(response.StructuredData)
		if err != nil {
			return nil, fmt.Errorf("result envelope structured data: %w", err)
		}
		env, err := parseEnvelope(prepared.definition.Envelope, text)
		if err != nil {
			return nil, fmt.Errorf("result envelope structured data: %w", err)
		}
		return env, nil
	}
	if strings.TrimSpace(response.Text) != "" {
		env, err := parseEnvelope(prepared.definition.Envelope, response.Text)
		if err != nil {
			return nil, fmt.Errorf("result envelope response text: %w", err)
		}
		return env, nil
	}
	return nil, fmt.Errorf("no result envelope: terminal outcome, structured data, and response text are empty")
}

func envelopeFromTerminalOutcome(prepared *preparedStep, outcome *api.TerminalOutcome) (*envelope, error) {
	if err := outcome.Validate(); err != nil {
		return nil, fmt.Errorf("invalid terminal outcome: %w", err)
	}
	switch outcome.Kind {
	case api.TerminalOutcomePlan:
		if prepared.class != types.ModePlan {
			return nil, fmt.Errorf("native plan outcome is invalid in %s mode", prepared.class)
		}
		status := nativePlanStatus(prepared.existingPlan, outcome.Plan.Content)
		return &envelope{
			ResultEnvelope: types.ResultEnvelope{Summary: nativePlanSummary(status), EndStatus: types.EndCompleted},
			Plan:           &types.PlanResult{Status: status, Path: outcome.Plan.Path, Content: outcome.Plan.Content},
		}, nil
	case api.TerminalOutcomeQuestions:
		questions := make([]types.AgentQuestion, len(outcome.Questions))
		for i, question := range outcome.Questions {
			questions[i] = types.AgentQuestion{
				Text: question.Text, Context: question.Context, Options: append([]string(nil), question.Options...),
			}
		}
		return &envelope{ResultEnvelope: types.ResultEnvelope{
			Summary: "The agent is waiting for answers.", EndStatus: types.EndAsk, Questions: questions,
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported terminal outcome kind %q", outcome.Kind)
	}
}

// parseEnvelope decodes the agent's final result into the shape its prompt
// promised. It switches on the envelope kind rather than the class: triage is
// plan-class but returns something entirely unlike a plan.
func parseEnvelope(kind todoprompt.EnvelopeKind, text string) (*envelope, error) {
	switch kind {
	case todoprompt.EnvelopePlan:
		parsed, err := captainai.ParseStructured(text, (*types.PlanEnvelope).Validate)
		if err != nil {
			return nil, err
		}
		return &envelope{ResultEnvelope: parsed.ResultEnvelope, Plan: &types.PlanResult{
			Status: parsed.PlanStatus, Path: parsed.PlanPath, Content: parsed.PlanContent,
		}}, nil
	case todoprompt.EnvelopeTriage:
		parsed, err := captainai.ParseStructured(text, (*types.TriageEnvelope).Validate)
		if err != nil {
			return nil, err
		}
		return &envelope{ResultEnvelope: parsed.ResultEnvelope, Triage: parsed}, nil
	default:
		parsed, err := captainai.ParseStructured(text, (*types.ResultEnvelope).Validate)
		if err != nil {
			return nil, err
		}
		return &envelope{ResultEnvelope: *parsed}, nil
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
	case types.PlanUnchanged:
		return "The existing plan is unchanged."
	default:
		return "The agent completed planning."
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
// outcomes can read fields the built-in envelopes do not name.
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

func questionVars(questions []types.AgentQuestion) []any {
	out := make([]any, 0, len(questions))
	for _, question := range questions {
		out = append(out, map[string]any{
			"text": question.Text, "context": question.Context, "options": question.Options,
		})
	}
	return out
}

// declaresVerify reports whether a workflow gives captain's verify registry
// anything to run. A bare `verify: {}` declares nothing.
func declaresVerify(workflow *api.Workflow) bool {
	if workflow == nil || workflow.Verify == nil {
		return false
	}
	v := workflow.Verify
	return strings.TrimSpace(v.Fixture) != "" || len(v.Commands) > 0 || len(v.Prompts) > 0
}

// handleEvent narrates the run into the transcript and the notification sink,
// and accounts the result event's usage.
func (h *Host) handleEvent(exec *todos.ExecutorContext, ev captainai.Event, execution *todos.ExecutionResult, todo *types.TODO, sawResult *bool, meta todos.RunStartMetadata) {
	transcript := exec.GetTranscript()
	switch ev.Kind {
	case captainai.EventText:
		if ev.Text == "" {
			return
		}
		transcript.AddExecutorMessage(truncate(ev.Text, 200), todos.EntryText, nil)
		exec.Notify(todos.Notification{Type: todos.NotifyProgress, Message: truncate(ev.Text, 100)})
	case captainai.EventThinking:
		transcript.AddExecutorMessage(ev.Text, todos.EntryThinking, nil)
		exec.Notify(todos.Notification{Type: todos.NotifyThinking, Message: truncate(ev.Text, 100)})
	case captainai.EventToolUse:
		action := toolSummary(ev)
		transcript.AddExecutorMessage(action, todos.EntryAction, map[string]any{"tool": ev.Tool})
		exec.Notify(todos.Notification{Type: todos.NotifyAction, Message: action})
	case captainai.EventPermission:
		action := toolSummary(ev)
		transcript.AddExecutorMessage("awaiting approval: "+action, todos.EntryAction, map[string]any{"tool": ev.Tool})
		exec.Notify(todos.Notification{Type: todos.NotifyApproval, Message: action})
	case captainai.EventSystem:
		if ev.SessionID != "" {
			setSessionID(todo, ev.SessionID)
			exec.RecordSessionID(ev.SessionID)
			meta.SessionID = ev.SessionID
			execution.Runtime.SessionID = ev.SessionID
			exec.RecordRunStart(meta)
		}
		if ev.Text != "" {
			transcript.AddExecutorMessage(ev.Text, todos.EntryAction, map[string]any{"role": "system"})
			exec.Notify(todos.Notification{Type: todos.NotifyAction, Message: ev.Text})
		}
	case captainai.EventResult:
		*sawResult = true
		if ev.Usage != nil {
			execution.TokensUsed += ev.Usage.TotalTokens()
		}
		execution.CostUSD += ev.CostUSD
		if !ev.Success {
			// A result that is not a success is an error even when the provider
			// attached no text: collect turns the message into a failed run, and a
			// blank one would have let the run read as succeeded.
			execution.ErrorMessage = firstNonEmpty(ev.Error, "agent reported an unsuccessful result")
		}
	case captainai.EventError:
		execution.ErrorMessage = ev.Error
		exec.Notify(todos.Notification{Type: todos.NotifyError, Message: ev.Error})
	}
}

func toolSummary(ev captainai.Event) string {
	for _, key := range []string{"command", "file_path", "path", "pattern", "query"} {
		if v, ok := ev.Input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return ev.Tool + ": " + truncate(s, 120)
			}
		}
	}
	return ev.Tool
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
