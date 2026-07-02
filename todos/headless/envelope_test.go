package headless

import (
	"context"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/gavel/todos/types"
)

// scriptedStream serves one event set per call and records every request, so
// tests can observe the final-result resume turn.
type scriptedStream struct {
	calls    []captainai.Request
	episodes [][]captainai.Event
}

func (s *scriptedStream) fn(ctx context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
	idx := len(s.calls)
	s.calls = append(s.calls, req)
	var events []captainai.Event
	if idx < len(s.episodes) {
		events = s.episodes[idx]
	}
	ch := make(chan captainai.Event, len(events))
	for _, ev := range events {
		select {
		case <-ctx.Done():
		default:
			ch <- ev
		}
	}
	close(ch)
	return ch, nil
}

const runEnvelopeJSON = `{"summary":"implemented the fix","endStatus":"completed"}`
const askEnvelopeJSON = `{"summary":"blocked","endStatus":"ask","questions":[{"text":"which db?"}]}`
const planEnvelopeJSON = `{"summary":"planned","endStatus":"completed","plan":{"status":"new","path":"/home/u/.claude/plans/p.md"}}`

func TestEnvelopeParsedFromFinalText(t *testing.T) {
	s := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: "working…"},
		{Kind: captainai.EventText, Text: runEnvelopeJSON},
		{Kind: captainai.EventResult, Success: true},
	}}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: s.fn})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Summary != "implemented the fix" || result.EndStatus != types.EndCompleted {
		t.Errorf("envelope not captured: %+v", result)
	}
	if len(s.calls) != 1 {
		t.Errorf("expected no final-result resume when the text carries the envelope; calls = %d", len(s.calls))
	}
}

func TestEnvelopeAskCapturesQuestions(t *testing.T) {
	s := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: askEnvelopeJSON},
		{Kind: captainai.EventResult, Success: true},
	}}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: s.fn})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.EndStatus != types.EndAsk || len(result.Questions) != 1 || result.Questions[0].Text != "which db?" {
		t.Errorf("ask envelope not captured: %+v", result)
	}
}

func TestEnvelopeMissingTriggersFinalResultResume(t *testing.T) {
	s := &scriptedStream{episodes: [][]captainai.Event{
		{
			{Kind: captainai.EventSystem, SessionID: "sess-42"},
			{Kind: captainai.EventText, Text: "done, but I forgot the JSON"},
			{Kind: captainai.EventResult, Success: true},
		},
		{
			{Kind: captainai.EventText, Text: runEnvelopeJSON},
			{Kind: captainai.EventResult, Success: true},
		},
	}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: s.fn})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(s.calls) != 2 {
		t.Fatalf("expected a final-result resume call, got %d call(s)", len(s.calls))
	}
	resume := s.calls[1]
	if resume.SessionID != "sess-42" {
		t.Errorf("resume SessionID = %q, want sess-42", resume.SessionID)
	}
	if !strings.Contains(resume.Prompt.User, "ONLY the final result JSON") {
		t.Errorf("resume prompt is not the final-result turn: %q", resume.Prompt.User)
	}
	if result.EndStatus != types.EndCompleted || result.Summary == "" {
		t.Errorf("envelope from resume not applied: %+v", result)
	}
}

func TestPlanRunWithoutEnvelopeFails(t *testing.T) {
	s := &scriptedStream{episodes: [][]captainai.Event{
		{
			{Kind: captainai.EventSystem, SessionID: "sess-9"},
			{Kind: captainai.EventText, Text: "narrative only"},
			{Kind: captainai.EventResult, Success: true},
		},
		{
			{Kind: captainai.EventText, Text: "still no json"},
			{Kind: captainai.EventResult, Success: true},
		},
	}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Mode: types.ModePlan, Stream: s.fn})
	_, err := e.Execute(newTestCtx(), &types.TODO{})
	if err == nil || !strings.Contains(err.Error(), "no result envelope") {
		t.Fatalf("plan run without an envelope must fail loudly, got %v", err)
	}
}

func TestPlanEnvelopeCapturesPlanPath(t *testing.T) {
	s := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: planEnvelopeJSON},
		{Kind: captainai.EventResult, Success: true},
	}}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Mode: types.ModePlan, Stream: s.fn})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.Plan == nil || result.Plan.Status != types.PlanNew || result.Plan.Path != "/home/u/.claude/plans/p.md" {
		t.Errorf("plan not captured: %+v", result.Plan)
	}
}

func TestRunModeDegradesWithoutEnvelope(t *testing.T) {
	// No session id, no envelope anywhere: run mode falls back to the transport
	// result (visible via empty EndStatus) instead of destroying the work.
	s := &scriptedStream{episodes: [][]captainai.Event{{
		{Kind: captainai.EventText, Text: "did the thing"},
		{Kind: captainai.EventResult, Success: true},
	}}}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: s.fn})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Success || result.EndStatus != "" {
		t.Errorf("expected transport success with empty EndStatus, got %+v", result)
	}
}

func TestTimeoutStillCapturesEnvelopeViaResume(t *testing.T) {
	hang := 0
	s := &scriptedStream{episodes: [][]captainai.Event{
		{
			{Kind: captainai.EventSystem, SessionID: "sess-t"},
			// no result event; the stream stays open past the run timeout
		},
		{
			{Kind: captainai.EventText, Text: runEnvelopeJSON},
			{Kind: captainai.EventResult, Success: true},
		},
	}}
	slow := func(ctx context.Context, req captainai.Request, p captainai.PermissionFunc) (<-chan captainai.Event, error) {
		ch, err := s.fn(ctx, req, p)
		if err != nil {
			return nil, err
		}
		if hang == 0 {
			hang++
			out := make(chan captainai.Event)
			go func() {
				defer close(out)
				for ev := range ch {
					out <- ev
				}
				<-ctx.Done() // hold the stream open until the run times out
			}()
			return out, nil
		}
		return ch, nil
	}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Timeout: 150 * time.Millisecond, Stream: slow})
	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute after timeout resume: %v", err)
	}
	if result.EndStatus != types.EndCompleted {
		t.Errorf("timed-out run should still capture the envelope via resume, got %+v", result)
	}
	if len(s.calls) != 2 {
		t.Errorf("expected 2 stream calls (run + resume), got %d", len(s.calls))
	}
	if !strings.Contains(s.calls[1].Prompt.User, "time limit") {
		t.Errorf("timed-out resume prompt should mention the time limit: %q", s.calls[1].Prompt.User)
	}
}
