package headless

import (
	"encoding/json"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type envelope struct {
	types.ResultEnvelope
	Plan *types.PlanResult
}

func (e *Executor) captureEnvelope(result *todopkg.ExecutionResult, response *captainai.Response) error {
	env, err := e.envelopeFromResponse(response)
	if err != nil {
		return err
	}
	applyEnvelope(result, env)
	return nil
}

func (e *Executor) envelopeFromResponse(response *captainai.Response) (*envelope, error) {
	if response == nil {
		return nil, fmt.Errorf("no result envelope: agent response is missing")
	}
	if response.TerminalOutcome != nil {
		return e.envelopeFromTerminalOutcome(response.TerminalOutcome)
	}
	if response.StructuredData != nil {
		text, err := structuredDataText(response.StructuredData)
		if err != nil {
			return nil, fmt.Errorf("result envelope structured data: %w", err)
		}
		env, err := e.parseEnvelope(text)
		if err != nil {
			return nil, fmt.Errorf("result envelope structured data: %w", err)
		}
		return env, nil
	}
	if strings.TrimSpace(response.Text) != "" {
		env, err := e.parseEnvelope(response.Text)
		if err != nil {
			return nil, fmt.Errorf("result envelope response text: %w", err)
		}
		return env, nil
	}
	return nil, fmt.Errorf("no result envelope: terminal outcome, structured data, and response text are empty")
}

func (e *Executor) envelopeFromTerminalOutcome(outcome *captainai.TerminalOutcome) (*envelope, error) {
	if err := outcome.Validate(); err != nil {
		return nil, fmt.Errorf("invalid terminal outcome: %w", err)
	}
	switch outcome.Kind {
	case captainai.TerminalOutcomePlan:
		if e.config.Mode != types.ModePlan {
			return nil, fmt.Errorf("native plan outcome is invalid in %s mode", e.config.Mode)
		}
		status := e.nativePlanStatus(outcome.Plan.Content)
		return &envelope{
			ResultEnvelope: types.ResultEnvelope{Summary: nativePlanSummary(status), EndStatus: types.EndCompleted},
			Plan: &types.PlanResult{
				Status:  status,
				Path:    outcome.Plan.Path,
				Content: outcome.Plan.Content,
			},
		}, nil
	case captainai.TerminalOutcomeQuestions:
		questions := make([]types.AgentQuestion, len(outcome.Questions))
		for i, question := range outcome.Questions {
			questions[i] = types.AgentQuestion{
				Text:    question.Text,
				Context: question.Context,
				Options: append([]string(nil), question.Options...),
			}
		}
		return &envelope{ResultEnvelope: types.ResultEnvelope{
			Summary:   "The agent is waiting for answers.",
			EndStatus: types.EndAsk,
			Questions: questions,
		}}, nil
	default:
		return nil, fmt.Errorf("unsupported terminal outcome kind %q", outcome.Kind)
	}
}

func (e *Executor) nativePlanStatus(content string) types.PlanStatus {
	existing := normalizePlanContent(e.config.ExistingPlan)
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

func (e *Executor) parseEnvelope(text string) (*envelope, error) {
	if e.config.Mode == types.ModePlan {
		parsed, err := captainai.ParseStructured(text, (*types.PlanEnvelope).Validate)
		if err != nil {
			return nil, err
		}
		return &envelope{ResultEnvelope: parsed.ResultEnvelope, Plan: &parsed.Plan}, nil
	}
	parsed, err := captainai.ParseStructured(text, (*types.ResultEnvelope).Validate)
	if err != nil {
		return nil, err
	}
	return &envelope{ResultEnvelope: *parsed}, nil
}

func applyEnvelope(result *todopkg.ExecutionResult, env *envelope) {
	result.Summary = env.Summary
	result.EndStatus = env.EndStatus
	result.Questions = env.Questions
	result.Plan = env.Plan
}
