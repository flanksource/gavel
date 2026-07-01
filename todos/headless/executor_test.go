package headless

import (
	"context"
	"testing"
	"time"

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

func TestHeadlessUsesPromptOverrideVerbatim(t *testing.T) {
	var gotPrompt string
	capture := func(_ context.Context, req captainai.Request) (<-chan captainai.Event, error) {
		gotPrompt = req.Prompt
		ch := make(chan captainai.Event, 1)
		ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
		close(ch)
		return ch, nil
	}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", PromptOverride: "EDITED PROMPT BODY", Stream: capture})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if gotPrompt != "EDITED PROMPT BODY" {
		t.Errorf("dispatched prompt = %q, want the override used verbatim", gotPrompt)
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

// captureReq is a stream that records the dispatched request and immediately
// returns a successful result, so a test can inspect req.CanUseTool.
func captureReq(into *captainai.Request) streamFunc {
	return func(_ context.Context, req captainai.Request) (<-chan captainai.Event, error) {
		*into = req
		ch := make(chan captainai.Event, 1)
		ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
		close(ch)
		return ch, nil
	}
}

func TestHeadlessNoApprovalCallbackByDefault(t *testing.T) {
	var captured captainai.Request
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: captureReq(&captured)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured.CanUseTool != nil {
		t.Fatal("CanUseTool must be nil when approvals are disabled (CLI runs have no resolver)")
	}
	if !contains(captured.AllowedTools, "Bash") {
		t.Errorf("Bash should stay allow-listed when approvals are off: %v", captured.AllowedTools)
	}
}

// TestHeadlessApprovalRoutesToRegistry verifies the approval callback the executor
// passes to captain blocks on the shared registry and maps the dashboard's
// decision (allow + updated input) back onto the captain decision shape.
func TestHeadlessApprovalRoutesToRegistry(t *testing.T) {
	const sessionID = "headless-approval-sess"
	var captured captainai.Request
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Approvals: true, Stream: captureReq(&captured)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured.CanUseTool == nil {
		t.Fatal("expected CanUseTool to be set when Approvals is enabled")
	}
	if contains(captured.AllowedTools, "Bash") {
		t.Errorf("Bash must be removed from the allowlist so it routes through approval: %v", captured.AllowedTools)
	}

	type outcome struct {
		decision captainai.PermissionDecision
		err      error
	}
	done := make(chan outcome, 1)
	go func() {
		d, err := captured.CanUseTool(context.Background(), captainai.PermissionRequest{
			SessionID: sessionID,
			Tool:      "Bash",
			Input:     map[string]any{"command": "ls"},
		})
		done <- outcome{d, err}
	}()

	waitForPendingApproval(t, sessionID)
	if err := todopkg.GlobalApprovals().Resolve(sessionID, todopkg.ApprovalDecision{
		Allow:        true,
		UpdatedInput: map[string]any{"command": "ls -la"},
	}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	got := <-done
	if got.err != nil {
		t.Fatalf("callback returned error: %v", got.err)
	}
	if !got.decision.Allow {
		t.Error("expected the allow decision to propagate")
	}
	if got.decision.UpdatedInput["command"] != "ls -la" {
		t.Errorf("updated input not propagated: %v", got.decision.UpdatedInput)
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func waitForPendingApproval(t *testing.T, sessionID string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, ok := todopkg.GlobalApprovals().Pending(sessionID); ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("approval for session %s never became pending", sessionID)
}

func TestHeadlessModelDefaults(t *testing.T) {
	claudeP, err := (&Executor{config: Config{Agent: "claude", Model: "claude"}}).newStreamer()
	if err != nil {
		t.Fatalf("claude streamer: %v", err)
	}
	if claudeP.GetBackend() != captainai.BackendClaudeAgent {
		t.Errorf("claude backend = %v, want claude-agent", claudeP.GetBackend())
	}
	codexP, err := (&Executor{config: Config{Agent: "codex", Model: "codex"}}).newStreamer()
	if err != nil {
		t.Fatalf("codex streamer: %v", err)
	}
	if codexP.GetBackend() != captainai.BackendCodexCLI {
		t.Errorf("codex backend = %v, want codex-cli", codexP.GetBackend())
	}
}
