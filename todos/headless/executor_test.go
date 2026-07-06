package headless

import (
	"context"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func fakeStream(events ...captainai.Event) streamFunc {
	return func(_ context.Context, _ captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
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

// TestHeadlessPromptOverrideReplacesBody pins the override contract: the edited
// prompt replaces the auto-built body, but the structured-output envelope rides
// on the native SchemaJSON field so an override cannot break it, and the schema
// text is never injected into the prompt body.
func TestHeadlessPromptOverrideReplacesBody(t *testing.T) {
	var gotPrompt string
	var gotSchema string
	capture := func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
		gotPrompt = req.Prompt.User
		gotSchema = string(req.Prompt.SchemaJSON)
		ch := make(chan captainai.Event, 1)
		ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
		close(ch)
		return ch, nil
	}
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", PromptOverride: "EDITED PROMPT BODY", Stream: capture})
	if _, err := e.Execute(newTestCtx(), &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "auto section"}}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(gotPrompt, "EDITED PROMPT BODY") {
		t.Errorf("dispatched prompt missing the override body: %q", gotPrompt)
	}
	if strings.Contains(gotPrompt, "auto section") {
		t.Error("override should replace the auto-built todo sections")
	}
	if !strings.Contains(gotSchema, `"endStatus"`) {
		t.Errorf("override must not drop the native envelope schema: %q", gotSchema)
	}
	if strings.Contains(gotPrompt, "conforms to this JSON Schema") {
		t.Error("schema instruction text must not be injected into the prompt body")
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

// captureReq is a stream that records the dispatched request and the tool-
// permission callback (which now lives on the provider Config, threaded through
// the seam) and immediately returns a successful result.
func captureReq(into *captainai.Request, canUseTool *captainai.PermissionFunc) streamFunc {
	return func(_ context.Context, req captainai.Request, fn captainai.PermissionFunc) (<-chan captainai.Event, error) {
		*into = req
		*canUseTool = fn
		ch := make(chan captainai.Event, 1)
		ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
		close(ch)
		return ch, nil
	}
}

func TestHeadlessNoApprovalCallbackByDefault(t *testing.T) {
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: captureReq(&captured, &canUseTool)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if canUseTool != nil {
		t.Fatal("CanUseTool must be nil when approvals are disabled (CLI runs have no resolver)")
	}
	if !contains(captured.Permissions.Tools.Allow, "Bash") {
		t.Errorf("Bash should stay allow-listed when approvals are off: %v", captured.Permissions.Tools.Allow)
	}
}

// TestHeadlessApprovalRoutesToRegistry verifies the approval callback the executor
// passes to captain blocks on the shared registry and maps the dashboard's
// decision (allow + updated input) back onto the captain decision shape.
func TestHeadlessApprovalRoutesToRegistry(t *testing.T) {
	const sessionID = "headless-approval-sess"
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Approvals: true, Stream: captureReq(&captured, &canUseTool)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if canUseTool == nil {
		t.Fatal("expected CanUseTool to be set when Approvals is enabled")
	}
	if contains(captured.Permissions.Tools.Allow, "Bash") {
		t.Errorf("Bash must be removed from the allowlist so it routes through approval: %v", captured.Permissions.Tools.Allow)
	}

	type outcome struct {
		decision captainai.PermissionDecision
		err      error
	}
	done := make(chan outcome, 1)
	go func() {
		d, err := canUseTool(context.Background(), captainai.PermissionRequest{
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

func TestHeadlessBuildsPermissionsFromToolModes(t *testing.T) {
	capture := func(out *captainai.Request) streamFunc {
		return func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
			*out = req
			ch := make(chan captainai.Event, 1)
			ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
			close(ch)
			return ch, nil
		}
	}

	t.Run("sdk backend maps modes to allow/deny and honours permission mode", func(t *testing.T) {
		var req captainai.Request
		e := NewExecutor(Config{
			WorkDir:        t.TempDir(),
			Agent:          "claude",
			Backend:        string(captainai.BackendClaudeAgent),
			ToolModes:      map[string]string{"Read": "enabled", "Write": "disabled", "Bash": "ask"},
			PermissionMode: "acceptEdits",
			Stream:         capture(&req),
		})
		if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if got := strings.Join(req.Permissions.Tools.Allow, ","); got != "Read" {
			t.Errorf("allow = %q, want Read", got)
		}
		if got := strings.Join(req.Permissions.Tools.Deny, ","); got != "Write" {
			t.Errorf("deny = %q, want Write", got)
		}
		if req.Permissions.Mode != api.PermissionAcceptEdits {
			t.Errorf("mode = %q, want acceptEdits", req.Permissions.Mode)
		}
		if len(req.Permissions.Presets) != 0 {
			t.Errorf("presets = %v, want none (explicit modes replace the edit preset)", req.Permissions.Presets)
		}
	})

	// The plan posture comes from the plan template's frontmatter now, not a
	// Plan flag: the rendered request carries permissions.mode: plan.
	t.Run("cmux plan run forces plan mode via the template", func(t *testing.T) {
		var req captainai.Request
		e := NewExecutor(Config{
			WorkDir: t.TempDir(),
			Agent:   "claude",
			Backend: string(captainai.BackendClaudeCmux),
			Mode:    types.ModePlan,
			Stream:  capture(&req),
		})
		// A plan run demands an envelope; serve one via the capture stream.
		e.config.Stream = func(_ context.Context, r captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
			req = r
			ch := make(chan captainai.Event, 2)
			ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"planned","endStatus":"completed","plan":{"status":"new","path":"/tmp/plan.md"}}`}
			ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
			close(ch)
			return ch, nil
		}
		if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if req.Permissions.Mode != api.PermissionPlan {
			t.Errorf("mode = %q, want plan", req.Permissions.Mode)
		}
	})

	t.Run("explicit user permission mode beats the template", func(t *testing.T) {
		var req captainai.Request
		e := NewExecutor(Config{
			WorkDir:        t.TempDir(),
			Agent:          "claude",
			Backend:        string(captainai.BackendClaudeAgent),
			Mode:           types.ModePlan,
			PermissionMode: "default",
			Stream: func(_ context.Context, r captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
				req = r
				ch := make(chan captainai.Event, 2)
				ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"planned","endStatus":"completed","plan":{"status":"new","path":"/tmp/plan.md"}}`}
				ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
				close(ch)
				return ch, nil
			},
		})
		if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if req.Permissions.Mode != api.PermissionDefault {
			t.Errorf("mode = %q, want default (explicit user override)", req.Permissions.Mode)
		}
	})
}

func TestHeadlessModelDefaults(t *testing.T) {
	claudeP, err := (&Executor{config: Config{Agent: "claude"}}).newStreamer(nil, "", "claude", "")
	if err != nil {
		t.Fatalf("claude streamer: %v", err)
	}
	if claudeP.GetBackend() != captainai.BackendClaudeAgent {
		t.Errorf("claude backend = %v, want claude-agent", claudeP.GetBackend())
	}
	codexP, err := (&Executor{config: Config{Agent: "codex"}}).newStreamer(nil, "", "codex", "")
	if err != nil {
		t.Fatalf("codex streamer: %v", err)
	}
	if codexP.GetBackend() != captainai.BackendCodexCLI {
		t.Errorf("codex backend = %v, want codex-cli", codexP.GetBackend())
	}
}

// fakeVerifier votes a fixed verdict every iteration.
type fakeVerifier struct{ ok bool }

func (f fakeVerifier) Name() string { return "fake" }
func (f fakeVerifier) Verify(*agent.RunContext, *captainai.LoopIteration) (agent.Verdict, error) {
	return agent.Verdict{OK: f.ok, Feedback: "fix it"}, nil
}

// TestHeadlessDoDVerdict pins the definition-of-done capture: the loop stops
// "condition-met" only when the verifier passes (→ DoD.Passed), otherwise the
// iteration budget runs out (→ not passed); with no verifiers there is no DoD.
func TestHeadlessDoDVerdict(t *testing.T) {
	events := []captainai.Event{
		{Kind: captainai.EventSystem, SessionID: "s"},
		{Kind: captainai.EventResult, Success: true},
	}

	t.Run("verifier passes → DoD passed", func(t *testing.T) {
		e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", MaxIterations: 3,
			Verifiers: []agent.Plugin{fakeVerifier{ok: true}}, Stream: fakeStream(events...)})
		result, err := e.Execute(newTestCtx(), &types.TODO{})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.DoD == nil || !result.DoD.Ran || !result.DoD.Passed {
			t.Fatalf("DoD = %+v, want ran+passed", result.DoD)
		}
	})

	t.Run("verifier keeps failing → DoD not passed", func(t *testing.T) {
		e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", MaxIterations: 2,
			Verifiers: []agent.Plugin{fakeVerifier{ok: false}}, Stream: fakeStream(events...)})
		result, err := e.Execute(newTestCtx(), &types.TODO{})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.DoD == nil || !result.DoD.Ran || result.DoD.Passed {
			t.Fatalf("DoD = %+v, want ran+not-passed", result.DoD)
		}
	})

	t.Run("no verifiers → no DoD", func(t *testing.T) {
		e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(events...)})
		result, err := e.Execute(newTestCtx(), &types.TODO{})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.DoD != nil {
			t.Fatalf("DoD = %+v, want nil (no definition of done)", result.DoD)
		}
	})
}
