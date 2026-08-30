package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// persistedSpec projects a spec the way todos/runtime writes `rendered_spec`, so
// a test's prior run carries exactly what a real one would.
func persistedSpec(t *testing.T, spec api.Spec) map[string]any {
	t.Helper()
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal persisted spec: %v", err)
	}
	var rendered map[string]any
	if err := json.Unmarshal(data, &rendered); err != nil {
		t.Fatalf("decode persisted spec: %v", err)
	}
	return rendered
}

const priorCodexModel = "gpt-5.6-sol"

// codexPlanRun is a finished plan turn as the native runtime persisted it: the
// spec it was dispatched with, plus the runtime it actually resolved.
func codexPlanRun(t *testing.T, sessionID string) *captaindb.PromptRun {
	t.Helper()
	return &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(types.ModePlan), Driver: string(drivers.Agent),
			Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Backend: "codex-agent", Model: priorCodexModel, Effort: "high",
			},
		},
		RenderedSpec: persistedSpec(t, api.Spec{
			SessionID: sessionID,
			// The family alias the run was requested with; Runtime.Resolved carries
			// the concrete model it became.
			Model:  api.Model{Name: "codex", Mode: api.ModeAgent},
			Budget: api.Budget{Cost: 4.5, MaxTurns: 12},
			// The turn that ran, and the tree it ran in — neither is configuration.
			Prompt: api.Prompt{User: "the previous turn's instructions"},
			Setup:  &shell.Setup{Cwd: "/previous/worktree"},
			Permissions: api.Permissions{
				Mode:  api.PermissionPlan,
				Tools: api.ToolsFromLists([]string{"Read", "Glob", "Grep"}, nil),
			},
		}),
	}
}

// A continuation that stays in the same mode continues the run it came from: the
// spec that was dispatched is the base, concretised by the runtime that turn
// actually resolved. What must not travel is the turn itself — the previous
// user prompt (todoprompt.Render would re-send it as the request) and the
// workspace a consumed checkout left behind.
func TestContinueRunInheritsTheDispatchedSpecWithinAMode(t *testing.T) {
	workDir := t.TempDir()
	opts, err := continueRun(continuation{
		Dir:   workDir,
		Todos: []*types.TODO{{}},
		Prior: codexPlanRun(t, "sess-prior-plan"),
		Mode:  types.ModePlan,
	})
	if err != nil {
		t.Fatalf("continueRun: %v", err)
	}

	if opts.Driver != string(drivers.Agent) {
		t.Errorf("driver = %q, want %q (the prior run's backend)", opts.Driver, drivers.Agent)
	}
	if opts.Spec.Name != priorCodexModel {
		t.Errorf("model = %q, want the prior run's resolved %q", opts.Spec.Name, priorCodexModel)
	}
	if string(opts.Spec.Backend) != "codex-agent" {
		t.Errorf("backend = %q, want codex-agent", opts.Spec.Backend)
	}
	if opts.Spec.Mode != api.ModeAgent {
		t.Errorf("authored backend = %q, want agent", opts.Spec.Mode)
	}
	if string(opts.Spec.Effort) != "high" {
		t.Errorf("effort = %q, want high", opts.Spec.Effort)
	}
	if opts.Spec.Budget.Cost != 4.5 || opts.Spec.Budget.MaxTurns != 12 {
		t.Errorf("budget = %+v, want the continued run's 4.5/12", opts.Spec.Budget)
	}

	if strings.Contains(opts.Spec.Prompt.User, "previous turn") {
		t.Errorf("prompt.user = %q, want the previous turn's instructions dropped", opts.Spec.Prompt.User)
	}
	if opts.Spec.Setup != nil && opts.Spec.Setup.Cwd == "/previous/worktree" {
		t.Error("continuation pinned to the workspace the previous run's setup produced")
	}
}

// Without an explicit resume the continuation is a fresh conversation, so the
// session the persisted spec was dispatched with must not come back with it.
func TestContinueRunWithoutResumeCarriesNoPriorSession(t *testing.T) {
	workDir := t.TempDir()
	const priorSession = "sess-prior-plan"
	opts, err := continueRun(continuation{
		Dir:   workDir,
		Todos: []*types.TODO{{}},
		Prior: codexPlanRun(t, priorSession),
		Mode:  types.ModePlan,
	})
	if err != nil {
		t.Fatalf("continueRun: %v", err)
	}
	if opts.Resume {
		t.Error("resume set on a continuation that did not ask for it")
	}
	if opts.Spec.SessionID == priorSession {
		t.Errorf("session id = %q, want the prior session dropped", opts.Spec.SessionID)
	}
}

// Approving a plan changes mode, and a mode change is not a continuation of
// configuration: the plan's read-only posture and its investigation budget
// belong to planning. Only the runtime selection — which agent actually ran —
// carries, so a codex plan can never be implemented by claude.
func TestContinueRunAcrossAModeChangeInheritsOnlyTheRuntime(t *testing.T) {
	workDir := t.TempDir()
	opts, err := continueRun(continuation{
		Dir:   workDir,
		Todos: []*types.TODO{{}},
		Prior: codexPlanRun(t, "sess-prior-plan"),
		Mode:  types.ModeRun,
	})
	if err != nil {
		t.Fatalf("continueRun: %v", err)
	}
	if opts.Spec.Name != priorCodexModel || string(opts.Spec.Backend) != "codex-agent" {
		t.Errorf("model/backend = %q/%q, want the plan run's resolved codex runtime",
			opts.Spec.Name, opts.Spec.Backend)
	}
	if opts.Spec.Permissions.Mode == api.PermissionPlan {
		t.Error("implement run inherited the plan run's read-only permission mode")
	}
	if opts.Spec.Budget.Cost == 4.5 {
		t.Error("implement run inherited the plan run's investigation budget instead of todos.run's")
	}
}

// H1: revise resumes the plan session, so it must resume on the runtime that
// session belongs to. The revise dialog sends no run options.
func TestTodoAPIPlanReviseInheritsPlanRunRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	const sid = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3836"
	mode := types.ModePlan
	session := sid
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &session, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	uiTestProviderFor(workDir).activeRun = codexPlanRun(t, sid)
	gotReq, called := stubTodoAnswer(t)

	body, _ := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(created), Feedback: "bound the queue"})
	rec := httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if !*called {
		t.Fatal("revise never dispatched a resume")
	}
	if gotReq.Options.Driver != string(drivers.Agent) {
		t.Errorf("driver = %q, want %q (the plan run's backend)", gotReq.Options.Driver, drivers.Agent)
	}
	if gotReq.Options.Spec.Backend.Family() != "codex" {
		t.Errorf("family = %q, want codex — a codex plan must not be revised by claude", gotReq.Options.Spec.Backend.Family())
	}
	if gotReq.Options.Spec.Name != priorCodexModel {
		t.Errorf("model = %q, want %q", gotReq.Options.Spec.Name, priorCodexModel)
	}
	if gotReq.Options.Spec.SessionID != sid {
		t.Errorf("session id = %q, want the plan session %q", gotReq.Options.Spec.SessionID, sid)
	}
}

// M3: the approved plan's implement run inherits the runtime that produced the
// plan, and reports the session id it will actually run under so the dashboard
// can attach to it.
func TestTodoAPIPlanApproveAndRunInheritsPlanRunRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	const sid = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3837"
	mode := types.ModePlan
	session := sid
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &session, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	uiTestProviderFor(workDir).activeRun = codexPlanRun(t, sid)

	oldStart := run.Start
	var got todoRunRequest
	// The stub answers with the session it was handed, as the real dispatcher
	// does. Returning an unrelated literal would make the report-matches-dispatch
	// assertion below compare a constant against a derived value.
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got = req
		return todoRunStartResult{Status: "started", SessionID: req.Options.Spec.SessionID}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoApprovePayload{Ref: todos.TODOReference(created), Run: true})
	rec := httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.RunMode != types.ModeRun {
		t.Errorf("chained run mode = %q, want run", got.Options.RunMode)
	}
	if got.Options.Driver != string(drivers.Agent) {
		t.Errorf("driver = %q, want %q (the plan run's backend)", got.Options.Driver, drivers.Agent)
	}
	if got.Options.Spec.Backend.Family() != "codex" {
		t.Errorf("family = %q, want codex — a codex plan must not be implemented by claude", got.Options.Spec.Backend.Family())
	}
	if got.Options.Spec.Name != priorCodexModel {
		t.Errorf("model = %q, want %q", got.Options.Spec.Name, priorCodexModel)
	}

	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Run == nil {
		t.Fatal("approve did not report the chained run")
	}
	if resp.Run.SessionID == "" {
		t.Error("approve reported no session id, so the dashboard cannot attach to the run it started")
	}
	if resp.Run.SessionID != got.Options.Spec.SessionID {
		t.Errorf("reported session %q != dispatched session %q", resp.Run.SessionID, got.Options.Spec.SessionID)
	}
}
