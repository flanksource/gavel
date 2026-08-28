package headless

import (
	"context"
	"encoding/json"
	"os"
	osExec "os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type Config struct {
	WorkDir        string
	Agent          string
	Model          string
	Backend        string
	Effort         string
	Fallbacks      api.ModelList
	MaxTurns       int
	MaxBudgetUsd   float64
	Tools          []string
	Timeout        time.Duration
	Mode           types.RunMode
	ExistingPlan   string
	PromptOverride string
	Verifiers      []agent.Verify
	MaxIterations  int
	Resume         bool
	SessionID      string
	ToolModes      map[string]string
	PermissionMode string
	Approvals      bool
	Stream         streamFunc
}

func newTestExecutor(config Config) *Executor {
	model, err := (api.Model{Name: config.Model, Fallbacks: config.Fallbacks}).Expand()
	if err != nil {
		panic(err)
	}
	if model.Name == "" && config.Agent == "codex" {
		model.Name = "codex"
	}
	if config.Backend != "" {
		model.Backend = captainai.Backend(config.Backend)
		model.Mode = model.Backend.Mode()
	} else if model.Mode == "" {
		model.Mode = api.ModeAgent
	}
	if config.Effort != "" {
		model.Effort = api.Effort(config.Effort)
	}
	permissions := api.Permissions{Mode: testPermissionMode(config.PermissionMode)}
	if len(config.Tools) > 0 {
		permissions.Tools = api.ToolsFromLists(config.Tools, nil)
	}
	for tool, mode := range config.ToolModes {
		if permissions.Tools == nil {
			permissions.Tools = api.Tools{}
		}
		switch mode {
		case "enabled":
			permissions.Tools[tool] = api.ToolPolicyAllow
		case "disabled":
			permissions.Tools[tool] = api.ToolPolicyDeny
		case "ask":
			permissions.Tools[tool] = api.ToolPolicyAsk
		}
	}
	budget := api.Budget{Cost: config.MaxBudgetUsd, MaxTurns: config.MaxTurns}
	if config.Timeout > 0 {
		budget.Timeout = config.Timeout.String()
	}
	runConfig := todopkg.AgentRunConfig{
		Spec: api.Spec{
			Model:       model,
			Prompt:      api.Prompt{User: config.PromptOverride},
			Budget:      budget,
			Permissions: permissions,
			SessionID:   config.SessionID,
		},
		WorkDir:       config.WorkDir,
		Mode:          config.Mode,
		ExistingPlan:  config.ExistingPlan,
		Verifiers:     config.Verifiers,
		MaxIterations: config.MaxIterations,
		Resume:        config.Resume,
		Approvals:     config.Approvals,
	}
	if config.Stream != nil {
		return NewExecutor(runConfig, withStream(config.Stream))
	}
	return NewExecutor(runConfig)
}

func testPermissionMode(mode string) api.PermissionMode {
	switch mode {
	case "plan":
		return api.PermissionPlan
	case "acceptEdits":
		return api.PermissionAcceptEdits
	case "auto":
		return api.PermissionAuto
	case "bypassPermissions":
		return api.PermissionBypass
	case "default", "dontAsk":
		return api.PermissionDefault
	default:
		return ""
	}
}

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
	e := newTestExecutor(Config{
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
	// Three reports, each adding what the one before it could not know: the
	// resolved runtime before dispatch, the spec once setup has transformed it,
	// and the provider's session id once the stream names one.
	if len(starts) != 3 {
		t.Fatalf("run starts = %d, want pre-dispatch, post-setup and session-bound metadata: %#v", len(starts), starts)
	}
	if starts[0].SessionID != "" || starts[0].Driver != "agent" || starts[0].Agent != "claude" || starts[0].ResolvedModel != "claude-agent-sonnet" {
		t.Fatalf("pre-dispatch run metadata = %+v", starts[0])
	}
	if starts[0].Spec != nil {
		t.Fatalf("pre-dispatch report carried a spec before setup ran: %+v", starts[0].Spec)
	}
	if starts[1].Spec == nil {
		t.Fatalf("post-setup run metadata carried no spec: %+v", starts[1])
	}
	if starts[2].SessionID != "sess-1" || starts[2].Mode != "run" || starts[2].ResolvedModel != "claude-agent-sonnet" || starts[2].Effort != "high" {
		t.Fatalf("session-bound run metadata = %+v", starts[2])
	}
	if starts[2].Spec != nil {
		t.Fatalf("session-bound report must leave the persisted spec alone: %+v", starts[2].Spec)
	}
	var messages []string
	for _, entry := range log.GetLogs() {
		messages = append(messages, entry.Message)
	}
	if got := strings.Join(messages, "\n"); !strings.Contains(got,
		"Resolved TODO runtime: driver=agent agent=claude provider=anthropic backend=claude-agent model=claude-agent-sonnet effort=high model_source=run-option") {
		t.Fatalf("resolved runtime log missing selection details:\n%s", got)
	}
}

func TestHeadlessLogsPromptDefaultModelSource(t *testing.T) {
	log := logger.NewBufferedLogger(20)
	e := newTestExecutor(Config{
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
		"driver=agent agent=claude provider=unknown backend=default model=claude effort=default model_source=todos.run-prompt") {
		t.Fatalf("prompt-default runtime log missing selection source:\n%s", got)
	}
}

func TestHeadlessParksOnAskUserQuestionWithoutRequestingEnvelope(t *testing.T) {
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
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
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", PromptOverride: "EDITED PROMPT BODY", Stream: capture})
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
	e := newTestExecutor(Config{
		WorkDir: t.TempDir(), Agent: "claude", Effort: "high",
		ExistingPlan: "# Approved plan\n\n1. Keep the database authoritative.",
		Stream:       capture,
	})
	todo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{Title: "Cut over runtime storage"},
		Implementation:  "Remove provider fallback.",
	}
	ctx := newTestCtx()
	preparedSpec, err := e.RenderRunSpec(ctx, todo)
	if err != nil {
		t.Fatalf("RenderRunSpec: %v", err)
	}
	prepared := preparedSpec.Prompt.User
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

func TestHeadlessPreservesCanonicalSpecAndRunsNativeWorkflow(t *testing.T) {
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "ready"), []byte("yes"), 0644); err != nil {
		t.Fatalf("write verifier fixture: %v", err)
	}
	spec := api.Spec{
		Model:  api.Model{Name: "claude-sonnet-5", Backend: captainai.BackendClaudeAgent, Effort: api.EffortHigh},
		Prompt: api.Prompt{User: "Implement the reviewed change.", System: "Keep the patch narrow."},
		Budget: api.Budget{Cost: 3, MaxTurns: 9, Timeout: "2m"},
		Memory: api.Memory{Skills: []string{"gavel-todos"}},
		Permissions: api.Permissions{
			Mode:    api.PermissionAcceptEdits,
			Tools:   api.Tools{"Bash": api.ToolPolicyAsk, "Read": api.ToolPolicyAllow},
			MCP:     api.MCP{Servers: []string{"postgres"}},
			Plugins: api.ResourcePolicies{"review": api.ResourceEnabled},
			Skills:  api.ResourcePolicies{"gavel-todos": api.ResourceEnabled},
		},
		Setup:     &shell.Setup{Cwd: "."},
		Workflow:  &api.Workflow{Verify: &api.Verify{Commands: []string{"test -f ready"}, Scope: api.VerifyScopeChanged, MaxIterations: 2}},
		SessionID: "session-123",
		CLIArgs:   map[string]any{"fullAuto": true},
	}
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	executor := NewExecutor(todopkg.AgentRunConfig{
		Spec:      spec,
		WorkDir:   workDir,
		Mode:      types.ModeRun,
		Approvals: true,
	}, withStream(captureReq(&captured, &canUseTool)))

	result, err := executor.Execute(newTestCtx(), &types.TODO{})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result.DoD == nil || !result.DoD.Passed {
		t.Fatalf("native workflow verification did not gate the run: %+v", result.DoD)
	}
	if canUseTool == nil {
		t.Fatal("dashboard approval callback was not constructed")
	}
	if captured.Cwd() != workDir {
		t.Fatalf("prepared cwd = %q, want %q", captured.Cwd(), workDir)
	}
	if captured.Prompt.System != spec.Prompt.System {
		t.Fatalf("system prompt = %q, want %q", captured.Prompt.System, spec.Prompt.System)
	}
	if !reflect.DeepEqual(captured.Memory, spec.Memory) ||
		!reflect.DeepEqual(captured.Permissions, spec.Permissions) ||
		!reflect.DeepEqual(captured.Workflow, spec.Workflow) ||
		!reflect.DeepEqual(captured.CLIArgs, spec.CLIArgs) {
		t.Fatalf("canonical Spec fields changed during dispatch:\n got: %#v\nwant: %#v", captured, spec)
	}
}

// A todo's relative CWD is joined onto the run's WorkDir exactly once — by
// groupWorkDir, the executor's single notion of where a group's work happens.
// Callers pass the un-joined base (the dashboard passes the source dir, the CLI
// the discovery root); pre-joining it would land the run one level too deep.
func TestHeadlessJoinsRelativeTodoCWDOnce(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	executor := NewExecutor(todopkg.AgentRunConfig{
		Spec:    api.Spec{Model: api.Model{Name: "claude"}},
		WorkDir: base,
		Mode:    types.ModeRun,
	}, withStream(captureReq(&captured, &canUseTool)))

	todo := &types.TODO{}
	todo.CWD = "sub"
	if _, err := executor.Execute(newTestCtx(), todo); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if want := filepath.Join(base, "sub"); captured.Cwd() != want {
		t.Fatalf("prepared cwd = %q, want %q", captured.Cwd(), want)
	}
}

func TestHeadlessPreparesAndCleansCanonicalWorktree(t *testing.T) {
	repo := initHeadlessGitRepo(t)
	if err := os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("local change\n"), 0644); err != nil {
		t.Fatalf("write dirty file: %v", err)
	}
	var preparedDir string
	stream := func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
		preparedDir = req.Cwd()
		if preparedDir == repo {
			t.Fatalf("request stayed in source repository instead of prepared worktree: %q", preparedDir)
		}
		if _, err := os.Stat(filepath.Join(preparedDir, "untracked.txt")); err != nil {
			t.Fatalf("prepared worktree omitted dirty file: %v", err)
		}
		// Relative setup paths anchor at the group's working directory, not at
		// the process working directory.
		if want := filepath.Join(repo, ".runtime"); req.Setup.BaseDir != want {
			t.Fatalf("setup base dir = %q, want %q", req.Setup.BaseDir, want)
		}
		// The checkout has been performed, so the spec no longer asks for it:
		// dispatching this same spec again must not create a second worktree.
		if req.Setup.Checkout != nil {
			t.Fatalf("prepared spec still requests a checkout: %+v", req.Setup.Checkout)
		}
		return fakeStream(captainai.Event{Kind: captainai.EventResult, Success: true})(context.Background(), req, nil)
	}
	executor := NewExecutor(todopkg.AgentRunConfig{
		Spec: api.Spec{
			Model: api.Model{Name: "claude"},
			Setup: &shell.Setup{
				Cwd:     ".",
				BaseDir: ".runtime",
				Checkout: &shell.Checkout{
					Mode: shell.CheckoutLocal,
					Path: ".",
					Worktree: &shell.Worktree{
						Mode:        shell.WorktreeNew,
						Uncommitted: shell.CloneClone,
					},
				},
			},
		},
		WorkDir: repo,
		Mode:    types.ModeRun,
	}, withStream(stream))

	if _, err := executor.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if preparedDir == "" {
		t.Fatal("stream did not observe the prepared worktree")
	}
	if _, err := os.Stat(preparedDir); !os.IsNotExist(err) {
		t.Fatalf("prepared worktree was not cleaned up: %q, stat error %v", preparedDir, err)
	}
}

// What gets persisted as the run's rendered spec must describe the tree the
// agent worked in, not the checkout it was asked to make — replaying a spec that
// still asked would clone a second worktree. The setup plugin performs that
// transform during PreRun, so the run re-reports its metadata afterwards with
// the rewritten spec attached.
func TestHeadlessReportsPostSetupSpec(t *testing.T) {
	repo := initHeadlessGitRepo(t)
	var dispatchedCwd string
	stream := func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
		dispatchedCwd = req.Cwd()
		return fakeStream(captainai.Event{Kind: captainai.EventResult, Success: true})(context.Background(), req, nil)
	}
	executor := NewExecutor(todopkg.AgentRunConfig{
		Spec: api.Spec{
			Model: api.Model{Name: "claude"},
			Setup: &shell.Setup{
				Cwd: ".",
				Checkout: &shell.Checkout{
					Mode:     shell.CheckoutLocal,
					Path:     ".",
					Worktree: &shell.Worktree{Mode: shell.WorktreeNew},
				},
			},
		},
		WorkDir: repo,
		Mode:    types.ModeRun,
	}, withStream(stream))

	ctx := newTestCtx()
	var reported []todopkg.RunStartMetadata
	ctx.SetRunStartHook(func(meta todopkg.RunStartMetadata) { reported = append(reported, meta) })

	if _, err := executor.Execute(ctx, &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var prepared *api.Spec
	for _, meta := range reported {
		if meta.Spec != nil {
			prepared = meta.Spec
		}
	}
	if prepared == nil {
		t.Fatalf("no run-start metadata carried the dispatched spec (%d reports)", len(reported))
	}
	if prepared.Setup == nil || prepared.Setup.Checkout != nil {
		t.Fatalf("reported spec still requests a checkout: %+v", prepared.Setup)
	}
	if dispatchedCwd == repo || dispatchedCwd == "" {
		t.Fatalf("run did not move into a prepared worktree: %q", dispatchedCwd)
	}
	if prepared.Setup.Cwd != dispatchedCwd {
		t.Fatalf("reported cwd = %q, want the prepared worktree %q", prepared.Setup.Cwd, dispatchedCwd)
	}
}

func initHeadlessGitRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runHeadlessGit(t, repo, "init", "-q", "-b", "main")
	runHeadlessGit(t, repo, "config", "user.email", "test@example.com")
	runHeadlessGit(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("seed\n"), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	runHeadlessGit(t, repo, "add", "seed.txt")
	runHeadlessGit(t, repo, "commit", "-q", "-m", "seed")
	return repo
}

func runHeadlessGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := osExec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func TestHeadlessFailsWhenResultUnsuccessful(t *testing.T) {
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
		captainai.Event{Kind: captainai.EventError, Error: "boom"},
		captainai.Event{Kind: captainai.EventResult, Success: false, Error: "boom"},
	)})
	_, err := e.Execute(newTestCtx(), &types.TODO{})
	if err == nil {
		t.Fatal("expected an error when the result reports failure")
	}
}

func TestHeadlessErrorsWithoutResult(t *testing.T) {
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(
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
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: captureReq(&captured, &canUseTool)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if canUseTool != nil {
		t.Fatal("CanUseTool must be nil when approvals are disabled (CLI runs have no resolver)")
	}
	if !contains(captured.Permissions.Tools.AllowList(), "Bash") {
		t.Errorf("Bash should stay allow-listed when approvals are off: %v", captured.Permissions.Tools.AllowList())
	}
}

// TestHeadlessApprovalRoutesToRegistry verifies the approval callback the executor
// passes to captain blocks on the shared registry and maps the dashboard's
// decision (allow + updated input) back onto the captain decision shape.
func TestHeadlessApprovalRoutesToRegistry(t *testing.T) {
	const sessionID = "headless-approval-sess"
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Approvals: true, Stream: captureReq(&captured, &canUseTool)})
	if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if canUseTool == nil {
		t.Fatal("expected CanUseTool to be set when Approvals is enabled")
	}
	if contains(captured.Permissions.Tools.AllowList(), "Bash") {
		t.Errorf("Bash must be removed from the allowlist so it routes through approval: %v", captured.Permissions.Tools.AllowList())
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

func TestHeadlessApprovalUsesCanonicalToolPolicies(t *testing.T) {
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	executor := NewExecutor(todopkg.AgentRunConfig{
		Spec: api.Spec{
			Model: api.Model{Name: "claude"},
			// auto, not allow: this is what a legacy `modes: {Read: on}` resolves to,
			// and buildCanUseTool must treat it as "run it" rather than falling
			// through to the approval registry.
			Permissions: api.Permissions{Tools: api.Tools{
				"Read":  api.ToolPolicyAuto,
				"Write": api.ToolPolicyDeny,
			}},
		},
		WorkDir:   t.TempDir(),
		Approvals: true,
	}, withStream(captureReq(&captured, &canUseTool)))
	if _, err := executor.Execute(newTestCtx(), &types.TODO{}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if canUseTool == nil {
		t.Fatal("expected approval callback")
	}

	allowed, err := canUseTool(t.Context(), captainai.PermissionRequest{Tool: "Read"})
	if err != nil || !allowed.Allow {
		t.Fatalf("Read decision = %+v, %v; want immediate allow", allowed, err)
	}
	denied, err := canUseTool(t.Context(), captainai.PermissionRequest{Tool: "Write"})
	if err != nil || denied.Allow || !strings.Contains(denied.Message, "disabled") {
		t.Fatalf("Write decision = %+v, %v; want immediate deny", denied, err)
	}
}

// TestHeadlessForwardsBudgetCap asserts the run's --max-budget cost ceiling
// reaches the dispatched request's Budget.Cost (alongside MaxTurns). Regression
// guard for the drop bug where the USD cap silently never reached Captain.
func TestHeadlessForwardsBudgetCap(t *testing.T) {
	var captured captainai.Request
	var canUseTool captainai.PermissionFunc
	e := newTestExecutor(Config{
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
		e := newTestExecutor(Config{
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
		e := newTestExecutor(Config{
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

	t.Run("agent backend maps modes to allow/deny and honours permission mode", func(t *testing.T) {
		var req captainai.Request
		e := newTestExecutor(Config{
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
		if got := strings.Join(req.Permissions.Tools.AllowList(), ","); got != "Read" {
			t.Errorf("allow = %q, want Read", got)
		}
		if got := strings.Join(req.Permissions.Tools.DenyList(), ","); got != "Write" {
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
		e := newTestExecutor(Config{
			WorkDir: t.TempDir(),
			Agent:   "claude",
			Backend: string(captainai.BackendClaudeCmux),
			Mode:    types.ModePlan,
			Stream:  capture(&req),
		})
		// A plan run demands an envelope; serve one via the capture stream.
		e.stream = func(_ context.Context, r captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
			req = r
			ch := make(chan captainai.Event, 2)
			ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"planned","endStatus":"completed","planStatus":"new","planPath":"/tmp/plan.md"}`}
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
		e := newTestExecutor(Config{
			WorkDir:        t.TempDir(),
			Agent:          "claude",
			Backend:        string(captainai.BackendClaudeAgent),
			Mode:           types.ModePlan,
			PermissionMode: "default",
			Stream: func(_ context.Context, r captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
				req = r
				ch := make(chan captainai.Event, 2)
				ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"planned","endStatus":"completed","planStatus":"new","planPath":"/tmp/plan.md"}`}
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

// TestHeadlessPlanRunIsReadOnly pins the plan posture as read-only. A plan run
// is nominally investigation, and its template says so — `permissions.mode:
// plan` plus an explicit read-only allowlist. Go used to read the template's
// silence on presets and tools as "no posture stated" and stamp the edit preset
// and the full edit toolset on top of it, so the one mode that must not touch
// the tree was handed Write, Edit and Bash anyway.
//
// The posture now travels whole: a request that states any part of it owns all
// of it.
func TestHeadlessPlanRunIsReadOnly(t *testing.T) {
	planStream := func(out *captainai.Request) streamFunc {
		return func(_ context.Context, req captainai.Request, _ captainai.PermissionFunc) (<-chan captainai.Event, error) {
			*out = req
			ch := make(chan captainai.Event, 2)
			ch <- captainai.Event{Kind: captainai.EventText, Text: `{"summary":"planned","endStatus":"completed","planStatus":"new","planPath":"/plan.md"}`}
			ch <- captainai.Event{Kind: captainai.EventResult, Success: true}
			close(ch)
			return ch, nil
		}
	}

	for _, backend := range []captainai.Backend{captainai.BackendClaudeAgent, captainai.BackendClaudeCmux} {
		t.Run(string(backend), func(t *testing.T) {
			var req captainai.Request
			e := newTestExecutor(Config{
				WorkDir: t.TempDir(),
				Agent:   "claude",
				Backend: string(backend),
				Mode:    types.ModePlan,
				Stream:  planStream(&req),
			})
			if _, err := e.Execute(newTestCtx(), &types.TODO{}); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if req.Permissions.Mode != api.PermissionPlan {
				t.Fatalf("mode = %q, want plan", req.Permissions.Mode)
			}
			if req.Permissions.HasPreset(api.PresetEdit) {
				t.Errorf("presets = %v, want no edit preset on a plan run", req.Permissions.Presets)
			}
			for _, tool := range []string{"Write", "Edit", "Bash"} {
				if contains(req.Permissions.Tools.AllowList(), tool) {
					t.Errorf("%s is allow-listed on a plan run: %v", tool, req.Permissions.Tools.AllowList())
				}
			}
			if !contains(req.Permissions.Tools.AllowList(), "Read") {
				t.Errorf("allow = %v, want the template's read-only tools", req.Permissions.Tools.AllowList())
			}
		})
	}
}

// TestPermissionDefaults pins the two halves of the posture rule separately:
// capability is all-or-nothing, the mode is always filled in.
//
// The mode carve-out is not symmetry for its own sake. An unset mode is not
// neutral downstream — claudeagent maps "" to bypassPermissions whenever no
// approval broker is attached — so a request that states only a toolset must
// still leave here with a mode, or narrowing the tools would silently widen the
// mode.
func TestPermissionDefaults(t *testing.T) {
	readOnly := api.ToolsFromLists([]string{"Read"}, nil)

	for _, tc := range []struct {
		name     string
		in       api.Permissions
		cmux     bool
		wantMode api.PermissionMode
		wantEdit bool // the edit preset + full toolset were stamped
	}{
		{"states nothing", api.Permissions{}, false, api.PermissionAcceptEdits, true},
		{"states nothing under cmux", api.Permissions{}, true, api.PermissionDefault, false},
		{"states only a mode", api.Permissions{Mode: api.PermissionPlan}, false, api.PermissionPlan, false},
		{"states only tools", api.Permissions{Tools: readOnly}, false, api.PermissionAcceptEdits, false},
		{"states only a preset", api.Permissions{Presets: []api.Preset{api.PresetBare}}, false, api.PermissionAcceptEdits, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := permissionDefaults(tc.in, tc.cmux)
			if got.Mode != tc.wantMode {
				t.Errorf("mode = %q, want %q", got.Mode, tc.wantMode)
			}
			if got.HasPreset(api.PresetEdit) != tc.wantEdit {
				t.Errorf("presets = %v, want edit preset stamped = %v", got.Presets, tc.wantEdit)
			}
			if tc.wantEdit && !contains(got.Tools.AllowList(), "Write") {
				t.Errorf("allow = %v, want the default edit toolset", got.Tools.AllowList())
			}
			if !tc.wantEdit && contains(got.Tools.AllowList(), "Write") {
				t.Errorf("allow = %v, want the stated toolset left alone", got.Tools.AllowList())
			}
		})
	}
}

func TestHeadlessModelDefaults(t *testing.T) {
	claudeP, err := newTestExecutor(Config{Agent: "claude"}).newStreamer(captainai.Request{Model: api.Model{Name: "claude"}}, nil, "")
	if err != nil {
		t.Fatalf("claude streamer: %v", err)
	}
	if claudeP.GetBackend() != captainai.BackendClaudeAgent {
		t.Errorf("claude backend = %v, want claude-agent", claudeP.GetBackend())
	}
	codexP, err := newTestExecutor(Config{Agent: "codex"}).newStreamer(captainai.Request{Model: api.Model{Name: "codex"}}, nil, "")
	if err != nil {
		t.Fatalf("codex streamer: %v", err)
	}
	if codexP.GetBackend() != captainai.BackendCodexAgent {
		t.Errorf("codex backend = %v, want codex-agent", codexP.GetBackend())
	}
	codexAgentP, err := newTestExecutor(Config{Agent: "codex"}).newStreamer(captainai.Request{Model: api.Model{Name: "gpt-5.5", Backend: captainai.BackendCodexAgent}}, nil, "")
	if err != nil {
		t.Fatalf("codex-agent streamer: %v", err)
	}
	if codexAgentP.GetBackend() != captainai.BackendCodexAgent {
		t.Errorf("codex explicit backend = %v, want codex-agent", codexAgentP.GetBackend())
	}
	if _, err := newTestExecutor(Config{Agent: "codex"}).newStreamer(captainai.Request{Model: api.Model{Name: "gpt-5.5", Backend: captainai.BackendCodexCLI}}, nil, ""); err == nil {
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
		e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", MaxIterations: 3,
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
		e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", MaxIterations: 2,
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
		e := newTestExecutor(Config{WorkDir: t.TempDir(), Agent: "claude", Stream: fakeStream(events...)})
		result, err := e.Execute(newTestCtx(), &types.TODO{})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if result.DoD != nil {
			t.Fatalf("DoD = %+v, want nil (no definition of done)", result.DoD)
		}
	})
}
