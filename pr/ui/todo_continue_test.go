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
	"github.com/flanksource/gavel/todos/lifecycle"
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
			Mode: string(types.ModePlan), Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Mode: "agent", Model: priorCodexModel, Effort: "high",
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

// resolvedRun folds a captured run request through the real lifecycle host, so
// a test asserts against the spec captain would be handed rather than against
// the request layer alone.
func resolvedRun(t *testing.T, req todoRunRequest) *lifecycle.Resolution {
	t.Helper()
	prepared, err := run.Resolve(t.Context(), req)
	if err != nil {
		t.Fatalf("resolve run: %v", err)
	}
	return prepared.Resolution
}

// continuationSpec resolves a continuation for a todo in dir through the seam
// the dashboard dispatches through, and returns the options it produced
// alongside the folded spec.
func continuationSpec(t *testing.T, dir string, c run.Continuation) (run.Options, api.Spec) {
	t.Helper()
	c.Dir, c.Provider = dir, uiTestProviderFor(dir)
	if c.Override.Host == "" {
		c.Override.Host = lifecycle.HostDashboard
	}
	if c.Todo == nil {
		created, err := c.Provider.Create(t.Context(), todos.CreateRequest{Title: "Continued", Status: types.StatusPending})
		if err != nil {
			t.Fatalf("seed todo: %v", err)
		}
		c.Todo = created
	}
	opts, err := run.Continue(c)
	if err != nil {
		t.Fatalf("run.Continue: %v", err)
	}
	resolution := resolvedRun(t, todoRunRequest{Provider: c.Provider, Todo: c.Todo, Dir: dir, Options: opts})
	return opts, resolution.Spec
}

// A continuation that stays in the same behaviour class continues the run it
// came from: the spec that was dispatched is the base, concretised by the
// runtime that turn actually resolved. What must not travel is the turn itself
// — the previous user prompt (the renderer would re-send it as the request) and
// the workspace a consumed checkout left behind.
func TestContinueRunInheritsTheDispatchedSpecWithinAClass(t *testing.T) {
	opts, spec := continuationSpec(t, t.TempDir(), run.Continuation{
		Prior: codexPlanRun(t, "sess-prior-plan"), Step: "plan",
	})
	if opts.Step != "plan" {
		t.Errorf("step = %q, want plan", opts.Step)
	}
	if spec.Name != priorCodexModel || spec.Mode != api.ModeAgent || string(spec.Effort) != "high" {
		t.Errorf("runtime = %s/%s/%s, want the prior run's resolved %s/agent/high", spec.Name, spec.Mode, spec.Effort, priorCodexModel)
	}
	if spec.Budget.Cost != 4.5 || spec.Budget.MaxTurns != 12 {
		t.Errorf("budget = %+v, want the continued run's 4.5/12", spec.Budget)
	}
	if strings.Contains(spec.Prompt.User, "previous turn") {
		t.Errorf("prompt.user = %q, want the previous turn's instructions dropped", spec.Prompt.User)
	}
	if spec.Setup != nil && spec.Setup.Cwd == "/previous/worktree" {
		t.Error("continuation pinned to the workspace the previous run's setup produced")
	}
}

// Without an explicit resume the continuation is a fresh conversation, so the
// session the persisted spec was dispatched with must not come back with it.
func TestContinueRunWithoutResumeCarriesNoPriorSession(t *testing.T) {
	const priorSession = "sess-prior-plan"
	opts, spec := continuationSpec(t, t.TempDir(), run.Continuation{
		Prior: codexPlanRun(t, priorSession), Step: "plan",
	})
	if opts.Resume {
		t.Error("resume set on a continuation that did not ask for it")
	}
	if spec.SessionID == priorSession {
		t.Errorf("session id = %q, want the prior session dropped", spec.SessionID)
	}
}

// Approving a plan changes class, and a class change is not a continuation of
// configuration: the plan's read-only posture and its investigation budget
// belong to planning. Only the runtime selection — which agent actually ran —
// carries, so a codex plan can never be implemented by claude.
func TestContinueRunAcrossAClassChangeInheritsOnlyTheRuntime(t *testing.T) {
	_, spec := continuationSpec(t, t.TempDir(), run.Continuation{
		Prior: codexPlanRun(t, "sess-prior-plan"), Step: "run",
	})
	if spec.Name != priorCodexModel || spec.Mode != api.ModeAgent {
		t.Errorf("model/mode = %q/%q, want the plan run's resolved codex runtime", spec.Name, spec.Mode)
	}
	if spec.Permissions.Mode == api.PermissionPlan {
		t.Error("implement run inherited the plan run's read-only permission mode")
	}
	if spec.Budget.Cost == 4.5 {
		t.Error("implement run inherited the plan run's investigation budget instead of todos.run's")
	}
}

// A continuation runs one step by definition — approve implements, revise
// plans — so dialog options naming a different step are a contradiction, not a
// choice to honour silently.
func TestContinueRunRejectsOptionsNamingAnotherStep(t *testing.T) {
	dir := t.TempDir()
	provider := uiTestProviderFor(dir)
	_, err := run.Continue(run.Continuation{
		Dir: dir, Provider: provider, Todo: &types.TODO{}, Step: "run",
		Override: run.Options{Step: "plan", Host: lifecycle.HostDashboard},
	})
	if err == nil || !strings.Contains(err.Error(), "options.step") {
		t.Fatalf("run.Continue accepted a contradicting step: %v", err)
	}
}

// H1: revise resumes the plan session, so it must resume on the runtime that
// session belongs to. The revise dialog sends no run options.
func TestTodoAPIPlanReviseInheritsPlanRunRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	const sid = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3836"
	session := sid
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &session}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	uiTestProviderFor(workDir).activeRun = codexPlanRun(t, sid)
	gotReq, called := stubRunStart(t)

	body, _ := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(created), Feedback: "bound the queue"})
	rec := httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if !*called {
		t.Fatal("revise never dispatched a resume")
	}
	spec := resolvedRun(t, *gotReq).Spec
	if providerKey(spec.Model) != "openai" || spec.Name != priorCodexModel {
		t.Errorf("runtime = %s/%s, want openai/%s — a codex plan must not be revised by claude", providerKey(spec.Model), spec.Name, priorCodexModel)
	}
	if spec.SessionID != sid {
		t.Errorf("session id = %q, want the plan session %q", spec.SessionID, sid)
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
	session := sid
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &session}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	uiTestProviderFor(workDir).activeRun = codexPlanRun(t, sid)

	oldStart := run.Start
	var got todoRunRequest
	// The stub answers with the session the resolved run will use, as the real
	// dispatcher does. Returning an unrelated literal would make the
	// report-matches-dispatch assertion below compare a constant against a
	// derived value.
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got = req
		return todoRunStartResult{Status: "started", SessionID: resolvedRun(t, req).Spec.SessionID}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoApprovePayload{Ref: todos.TODOReference(created), Run: true})
	rec := httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Step != "run" || got.Options.Resume {
		t.Errorf("chained run = step %q resume %v, want a fresh run step", got.Options.Step, got.Options.Resume)
	}
	if got.Prepared == nil || got.Prepared.Step.Name != "run" {
		t.Errorf("chained run carried no pre-flight resolution; Start would fold a second time")
	}
	spec := resolvedRun(t, got).Spec
	if providerKey(spec.Model) != "openai" || spec.Name != priorCodexModel {
		t.Errorf("runtime = %s/%s, want openai/%s — a codex plan must not be implemented by claude", providerKey(spec.Model), spec.Name, priorCodexModel)
	}

	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Run == nil {
		t.Fatal("approve did not report the chained run")
	}
	if resp.Run.Step != "run" {
		t.Errorf("reported step = %q, want run", resp.Run.Step)
	}
	if resp.Run.SessionID != spec.SessionID {
		t.Errorf("reported session %q != dispatched session %q", resp.Run.SessionID, spec.SessionID)
	}
}
