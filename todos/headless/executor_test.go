package headless

import (
	"context"
	"encoding/json"
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
			ch <- withRunEnvelope(ev)
		}
		close(ch)
		return ch, nil
	}
}

func withRunEnvelope(event captainai.Event) captainai.Event {
	if event.Kind == captainai.EventResult && event.Success && len(event.StructuredData) == 0 {
		event.StructuredData = json.RawMessage(runEnvelopeJSON)
	}
	return event
}

func newTestCtx() *todopkg.ExecutorContext {
	return todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
}

func TestHeadlessCompletesOnResult(t *testing.T) {
	log := logger.NewBufferedLogger(20)
	e := NewExecutor(Config{
		WorkDir: t.TempDir(), Agent: "claude", Model: "claude-agent-sonnet",
		Backend: string(captainai.BackendClaudeAgent), Effort: "high", Stream: fakeStream(
			captainai.Event{Kind: captainai.EventSystem, SessionID: "sess-1"},
			captainai.Event{Kind: captainai.EventText, Text: "working on it"},
			captainai.Event{Kind: captainai.EventToolUse, Tool: "Edit", Input: map[string]any{"file_path": "/repo/x.go"}},
			captainai.Event{Kind: captainai.EventResult, Success: true, CostUSD: 0.12, Usage: &captainai.Usage{InputTokens: 100, OutputTokens: 50}},
		)})
	todo := &types.TODO{}
	ctx := todopkg.NewExecutorContext(context.Background(), log, nil)
	var starts []todopkg.RunStartMetadata
	ctx.SetRunStartHook(func(meta todopkg.RunStartMetadata) {
		starts = append(starts, meta)
	})
	result, err := e.Execute(ctx, todo)
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
	if len(starts) != 2 {
		t.Fatalf("run starts = %d, want pre-dispatch and session-bound metadata: %#v", len(starts), starts)
	}
	if starts[0].SessionID != "" || starts[0].Driver != "cli" || starts[0].Agent != "claude" || starts[0].ResolvedModel != "claude-agent-sonnet" {
		t.Fatalf("pre-dispatch run metadata = %+v", starts[0])
	}
	if starts[1].SessionID != "sess-1" || starts[1].Mode != "run" || starts[1].ResolvedModel != "claude-agent-sonnet" || starts[1].Effort != "high" {
		t.Fatalf("session-bound run metadata = %+v", starts[1])
	}
	var messages []string
	for _, entry := range log.GetLogs() {
		messages = append(messages, entry.Message)
	}
	if got := strings.Join(messages, "\n"); !strings.Contains(got,
		"Resolved TODO runtime: driver=cli agent=claude provider=anthropic backend=claude-agent model=claude-agent-sonnet effort=high model_source=run-option") {
		t.Fatalf("resolved runtime log missing selection details:\n%s", got)
	}
}

func TestHeadlessLogsPromptDefaultModelSource(t *testing.T) {
	log := logger.NewBufferedLogger(20)
	e := NewExecutor(Config{
		WorkDir: t.TempDir(), Agent: "claude",
		Stream: fakeStream(captainai.Event{Kind: captainai.EventResult, Success: true}),
	})

	if _, err := e.Execute(todopkg.NewExecutorContext(t.Context(), log, nil), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var messages []string
	for _, entry := range log.GetLogs() {
		messages = append(messages, entry.Message)
	}
	if got := strings.Join(messages, "\n"); !strings.Contains(got,
		"driver=cli agent=claude provider=unknown backend=default model=claude effort=default model_source=todos.run-prompt") {
		t.Fatalf("prompt-default runtime log missing selection source:\n%s", got)
	}
}

func TestHeadlessParksOnAskUserQuestionWithoutRequestingEnvelope(t *testing.T) {
	e := NewExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
		captainai.Event{Kind: captainai.EventSystem, SessionID: "sess-ask"},
		captainai.Event{Kind: captainai.EventToolUse, Tool: "AskUserQuestion", Input: map[string]any{
			"questions": []any{map[string]any{
				"question": "Which database?", "header": "Storage",
				"options": []any{map[string]any{"label": "Postgres"}, map[string]any{"label": "SQLite"}},
			}},
		}},
		captainai.Event{Kind: captainai.EventResult, Success: true},
	)})

	result, err := e.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.EndStatus != types.EndAsk || len(result.Questions) != 1 {
		t.Fatalf("result = %+v, want one ask question", result)
	}
	if got := result.Questions[0]; got.Text != "Which database?" || got.Context != "Storage" || len(got.Options) != 2 {
		t.Fatalf("question = %+v", got)
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
		ch <- withRunEnvelope(captainai.Event{Kind: captainai.EventResult, Success: true})
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

func TestHeadlessPreparedPromptMatchesDispatchWithApprovedPlan(t *testing.T) {
	var dispatched string
	capture := func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
		dispatched = req.Prompt.User
		ch := make(chan captainai.Event, 1)
		ch <- withRunEnvelope(captainai.Event{Kind: captainai.EventResult, Success: true})
		close(ch)
		return ch, nil
	}
	e := NewExecutor(Config{
		WorkDir: t.TempDir(), Agent: "claude", Effort: "high",
		ExistingPlan: "# Approved plan\n\n1. Keep the database authoritative.",
		Stream:       capture,
	})
	todo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{Title: "Cut over runtime storage"},
		Implementation:  "Remove provider fallback.",
	}
	ctx := newTestCtx()
	prepared, err := e.RenderRunPrompt(ctx, todo)
	if err != nil {
		t.Fatalf("RenderRunPrompt: %v", err)
	}
	if _, err := e.Execute(ctx, todo); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if prepared != dispatched {
		t.Fatalf("prepared Captain prompt differs from dispatch:\nprepared: %q\ndispatched: %q", prepared, dispatched)
	}
	if !strings.Contains(prepared, "Approved plan") {
		t.Fatalf("prepared prompt omitted approved plan: %q", prepared)
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
		ch <- withRunEnvelope(captainai.Event{Kind: captainai.EventResult, Success: true})
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

// TestHeadlessForwardsBudgetCap asserts the run's --max-budget cost ceiling
// reaches the dispatched request's Budget.Cost (alongside MaxTurns). Regression
// guard for the drop bug where drivers.Config.MaxBudgetUsd was never plumbed
// into headless.Config, so the USD cap silently never reached captain.
func TestHeadlessForwardsBudgetCap(t *testing.T) {
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	e := NewExecutor(Config{
		WorkDir:      t.TempDir(),
		Agent:        "claude",
		MaxBudgetUsd: 7.5,
		MaxTurns:     12,
		Stream:       captureReq(&captured, &canUseTool),
	})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if captured.Budget.Cost != 7.5 {
		t.Errorf("Budget.Cost = %v, want 7.5 (the --max-budget cap must reach the request)", captured.Budget.Cost)
	}
	if captured.Budget.MaxTurns != 12 {
		t.Errorf("Budget.MaxTurns = %d, want 12", captured.Budget.MaxTurns)
	}
}

// TestHeadlessForwardsModelFallbacks asserts the run's fallback chain reaches the
// dispatched request so captain's Candidates() tries them after the primary.
// Regression guard for the drop bug where only the model NAME flowed end to end
// (Config.Model → req.Name) and Model.Fallbacks never entered the path.
func TestHeadlessForwardsModelFallbacks(t *testing.T) {
	t.Run("explicit Fallbacks field folds into the request", func(t *testing.T) {
		var captured captainai.Request
		var canUseTool captainai.PermissionFunc
		e := NewExecutor(Config{
			WorkDir:   t.TempDir(),
			Agent:     "claude",
			Model:     "claude-opus-4-6",
			Fallbacks: api.ModelList{{Name: "claude-sonnet-4-6"}},
			Stream:    captureReq(&captured, &canUseTool),
		})
		if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if captured.Name != "claude-opus-4-6" {
			t.Errorf("primary model = %q, want claude-opus-4-6", captured.Name)
		}
		if got := captured.Model.Fallbacks; len(got) != 1 || got[0].Name != "claude-sonnet-4-6" {
			t.Fatalf("Model.Fallbacks = %+v, want [claude-sonnet-4-6]", got)
		}
		cands := captured.Candidates()
		if len(cands) != 2 || cands[0].Name != "claude-opus-4-6" || cands[1].Name != "claude-sonnet-4-6" {
			t.Fatalf("Candidates() = %+v, want [claude-opus-4-6 claude-sonnet-4-6]", cands)
		}
	})

	t.Run("compact model tail expands into fallbacks", func(t *testing.T) {
		var captured captainai.Request
		var canUseTool captainai.PermissionFunc
		e := NewExecutor(Config{
			WorkDir: t.TempDir(),
			Agent:   "claude",
			Model:   "claude-opus-4-6, claude-sonnet-4-6",
			Stream:  captureReq(&captured, &canUseTool),
		})
		if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if captured.Name != "claude-opus-4-6" {
			t.Errorf("primary model = %q, want claude-opus-4-6 (compact tail moved to fallbacks)", captured.Name)
		}
		cands := captured.Candidates()
		if len(cands) != 2 || cands[1].Name != "claude-sonnet-4-6" {
			t.Fatalf("Candidates() = %+v, want a claude-sonnet-4-6 fallback", cands)
		}
	})
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
			ch <- withRunEnvelope(captainai.Event{Kind: captainai.EventResult, Success: true})
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
	if codexP.GetBackend() != captainai.BackendCodexAgent {
		t.Errorf("codex backend = %v, want codex-agent", codexP.GetBackend())
	}
	codexAgentP, err := (&Executor{config: Config{Agent: "codex"}}).newStreamer(nil, "", "gpt-5.5", string(captainai.BackendCodexAgent))
	if err != nil {
		t.Fatalf("codex-agent streamer: %v", err)
	}
	if codexAgentP.GetBackend() != captainai.BackendCodexAgent {
		t.Errorf("codex explicit backend = %v, want codex-agent", codexAgentP.GetBackend())
	}
	if _, err := (&Executor{config: Config{Agent: "codex"}}).newStreamer(nil, "", "gpt-5.5", string(captainai.BackendCodexCLI)); err == nil {
		t.Fatal("codex-cli backend should be rejected")
	}
}

// fakeVerifier votes a fixed verdict every iteration.
type fakeVerifier struct{ ok bool }

func (f fakeVerifier) Name() string { return "fake" }
func (f fakeVerifier) Verify(hc *agent.HookContext) (agent.VerifyResult, error) {
	if f.ok {
		return agent.VerifyResult{Valid: true}, nil
	}
	retry := *hc.Request
	retry.Prompt = api.Prompt{User: "fix it"}
	return agent.VerifyResult{Valid: false, Retry: &retry}, nil
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
			Verifiers: []agent.Verify{fakeVerifier{ok: true}}, Stream: fakeStream(events...)})
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
			Verifiers: []agent.Verify{fakeVerifier{ok: false}}, Stream: fakeStream(events...)})
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
