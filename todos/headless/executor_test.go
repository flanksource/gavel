package headless

import (
	"context"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func fakeStream(events ...captainai.Event) streamFunc {
	return func(_ context.Context, _ captainai.Request) (<-chan captainai.Event, error) {
		ch := make(chan captainai.Event, len(events))
		for _, ev := range events {
			ch <- ev
		}
		close(ch)
		return ch, nil
	}
}

func newTestCtx() *todopkg.ExecutorContext {
	return todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
}

func TestHeadlessCompletesOnResult(t *testing.T) {
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
		captainai.Event{Kind: captainai.EventSystem, SessionID: "sess-1"},
		captainai.Event{Kind: captainai.EventText, Text: "working on it"},
		captainai.Event{Kind: captainai.EventToolUse, Tool: "Edit", Input: map[string]any{"file_path": "/repo/x.go"}},
		captainai.Event{Kind: captainai.EventResult, Success: true, CostUSD: 0.12, Usage: &captainai.Usage{InputTokens: 100, OutputTokens: 50}},
	)})
	todo := &types.TODO{}
	result, err := e.Execute(newTestCtx(), todo)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.TokensUsed != 150 {
		t.Errorf("tokens = %d, want 150", result.TokensUsed)
	}
	if result.CostUSD != 0.12 {
		t.Errorf("cost = %v, want 0.12", result.CostUSD)
	}
	if todo.LLM == nil || todo.LLM.SessionId != "sess-1" {
		t.Errorf("session id not recorded on todo: %+v", todo.LLM)
	}
}

func TestHeadlessFailsWhenResultUnsuccessful(t *testing.T) {
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
		captainai.Event{Kind: captainai.EventError, Error: "boom"},
		captainai.Event{Kind: captainai.EventResult, Success: false, Error: "boom"},
	)})
	_, err := e.Execute(newTestCtx(), &types.TODO{})
	if err == nil {
		t.Fatal("expected an error when the result reports failure")
	}
}

func TestHeadlessErrorsWithoutResult(t *testing.T) {
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
		captainai.Event{Kind: captainai.EventText, Text: "hi"},
	)})
	_, err := e.Execute(newTestCtx(), &types.TODO{})
	if err == nil {
		t.Fatal("expected an error when the stream ends without a result event")
	}
}

func TestHeadlessModelDefaults(t *testing.T) {
	claudeP, err := (&Executor{config: Config{Agent: "claude", Model: "claude"}}).newStreamer()
	if err != nil {
		t.Fatalf("claude streamer: %v", err)
	}
	if claudeP.GetBackend() != captainai.BackendClaudeCLI {
		t.Errorf("claude backend = %v, want claude-cli", claudeP.GetBackend())
	}
	codexP, err := (&Executor{config: Config{Agent: "codex", Model: "codex"}}).newStreamer()
	if err != nil {
		t.Fatalf("codex streamer: %v", err)
	}
	if codexP.GetBackend() != captainai.BackendCodexCLI {
		t.Errorf("codex backend = %v, want codex-cli", codexP.GetBackend())
	}
}
