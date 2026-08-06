package headless

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/gavel/todos/types"
)

type scriptedStream struct {
	calls    []captainai.Request
	episodes [][]captainai.Event
}

func (s *scriptedStream) fn(ctx context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
	index := len(s.calls)
	s.calls = append(s.calls, req)
	var events []captainai.Event
	if index < len(s.episodes) {
		events = s.episodes[index]
	}
	stream := make(chan captainai.Event, len(events))
	for _, event := range events {
		select {
		case <-ctx.Done():
		default:
			stream <- event
		}
	}
	close(stream)
	return stream, nil
}

const runEnvelopeJSON = `{"summary":"implemented the fix","endStatus":"completed"}`
const planEnvelopeJSON = `{"summary":"planned","endStatus":"completed","planStatus":"new","planPath":"/home/u/.claude/plans/p.md"}`

func TestEnvelopePrefersStructuredResult(t *testing.T) {
	stream := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: `{"summary":"interim","endStatus":"failed"}`},
		{Kind: captainai.EventResult, Success: true, StructuredData: json.RawMessage(runEnvelopeJSON)},
	}}}
	executor := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "codex", Stream: stream.fn})

	result, err := executor.Execute(newTestCtx(), &types.TODO{})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Summary != "implemented the fix" || result.EndStatus != types.EndCompleted {
		t.Fatalf("structured envelope not captured: %+v", result)
	}
}

func TestEnvelopeFallsBackToResponseText(t *testing.T) {
	stream := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: runEnvelopeJSON},
		{Kind: captainai.EventResult, Success: true},
	}}}
	executor := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: stream.fn})

	result, err := executor.Execute(newTestCtx(), &types.TODO{})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Summary != "implemented the fix" || result.EndStatus != types.EndCompleted {
		t.Fatalf("text envelope not captured: %+v", result)
	}
}

func TestEnvelopeNativeQuestions(t *testing.T) {
	stream := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventToolUse, Tool: "AskUserQuestion", Input: map[string]any{"questions": []any{
			map[string]any{
				"question": "Which database?",
				"header":   "Storage",
				"options":  []any{map[string]any{"label": "PostgreSQL"}, "SQLite"},
			},
		}}},
		{Kind: captainai.EventResult, Success: true},
	}}}
	executor := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: stream.fn})

	result, err := executor.Execute(newTestCtx(), &types.TODO{})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.EndStatus != types.EndAsk || len(result.Questions) != 1 {
		t.Fatalf("native questions not captured: %+v", result)
	}
	question := result.Questions[0]
	if question.Text != "Which database?" || question.Context != "Storage" || strings.Join(question.Options, ",") != "PostgreSQL,SQLite" {
		t.Fatalf("question = %+v", question)
	}
}

func TestEnvelopeNativePlanDerivesStatus(t *testing.T) {
	tests := []struct {
		name         string
		existingPlan string
		content      string
		wantStatus   types.PlanStatus
	}{
		{name: "new", content: "# Plan\n\n1. Inspect", wantStatus: types.PlanNew},
		{name: "updated", existingPlan: "# Old plan", content: "# New plan", wantStatus: types.PlanUpdated},
		{name: "unchanged", existingPlan: "# Plan\r\n\r\n1. Inspect\n", content: "# Plan\n\n1. Inspect", wantStatus: types.PlanUnchanged},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := newTestExecutor(Config{Mode: types.ModePlan, ExistingPlan: test.existingPlan})
			response := &captainai.Response{TerminalOutcome: &captainai.TerminalOutcome{
				Kind: captainai.TerminalOutcomePlan,
				Plan: &captainai.TerminalPlan{Content: test.content, Path: "/repo/.claude/plans/example.md"},
			}}

			env, err := executor.envelopeFromResponse(response)

			if err != nil {
				t.Fatalf("envelopeFromResponse: %v", err)
			}
			if env.Plan == nil || env.Plan.Status != test.wantStatus || env.Plan.Content != test.content {
				t.Fatalf("plan = %+v, want status %s", env.Plan, test.wantStatus)
			}
		})
	}
}

func TestEnvelopeNativeOutcomePrecedesStructuredData(t *testing.T) {
	executor := newTestExecutor(Config{Mode: types.ModePlan})
	response := &captainai.Response{
		TerminalOutcome: &captainai.TerminalOutcome{
			Kind: captainai.TerminalOutcomePlan,
			Plan: &captainai.TerminalPlan{Content: "# Native plan", Path: "/repo/native.md"},
		},
		StructuredData: json.RawMessage(planEnvelopeJSON),
	}

	env, err := executor.envelopeFromResponse(response)

	if err != nil {
		t.Fatalf("envelopeFromResponse: %v", err)
	}
	if env.Plan == nil || env.Plan.Content != "# Native plan" || env.Plan.Path != "/repo/native.md" {
		t.Fatalf("envelope = %+v, want native plan", env)
	}
}

func TestEnvelopeStructuredPlanFallback(t *testing.T) {
	executor := newTestExecutor(Config{Mode: types.ModePlan})
	response := &captainai.Response{StructuredData: json.RawMessage(planEnvelopeJSON)}

	env, err := executor.envelopeFromResponse(response)

	if err != nil {
		t.Fatalf("envelopeFromResponse: %v", err)
	}
	if env.Plan == nil || env.Plan.Status != types.PlanNew || env.Plan.Path != "/home/u/.claude/plans/p.md" {
		t.Fatalf("structured plan not captured: %+v", env.Plan)
	}
}

func TestEnvelopeInvalidStructuredDataDoesNotFallBackToText(t *testing.T) {
	executor := newTestExecutor(Config{})
	response := &captainai.Response{
		StructuredData: json.RawMessage(`{"summary":"missing status"}`),
		Text:           runEnvelopeJSON,
	}

	_, err := executor.envelopeFromResponse(response)

	if err == nil || !strings.Contains(err.Error(), "structured data") {
		t.Fatalf("invalid structured data must fail without text fallback, got %v", err)
	}
}

func TestRunWithoutEnvelopeFailsWithoutResume(t *testing.T) {
	stream := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventSystem, SessionID: "sess-42"},
		{Kind: captainai.EventText, Text: "done, but no result envelope"},
		{Kind: captainai.EventResult, Success: true},
	}}}
	executor := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: stream.fn})

	_, err := executor.Execute(newTestCtx(), &types.TODO{})

	if err == nil || !strings.Contains(err.Error(), "no result envelope") {
		t.Fatalf("missing envelope must fail loudly, got %v", err)
	}
	if len(stream.calls) != 1 {
		t.Fatalf("missing envelope triggered %d stream calls, want exactly 1", len(stream.calls))
	}
}
