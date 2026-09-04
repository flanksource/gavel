package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

func seedReviewTodo(t *testing.T, workDir string, status types.Status) *types.TODO {
	t.Helper()
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Reviewable",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{Status: &status}); err != nil {
		t.Fatalf("seed status: %v", err)
	}
	return created
}

// stubbedRunSession is the session the run stub admits a todo with no recorded
// session under, as the real dispatcher mints one.
const stubbedRunSession = "11111111-1111-4111-8111-111111111111"

// stubRunStart records the request the handler under test dispatches, without
// standing up an agent. The stub answers with the todo's recorded session — a
// resume continues it — or a minted one, as the real dispatcher does.
func stubRunStart(t *testing.T) (*todoRunRequest, *bool) {
	t.Helper()
	var got todoRunRequest
	called := false
	old := run.Start
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got, called = req, true
		session := run.PriorSessionID(req.Todo)
		if session == "" {
			session = stubbedRunSession
		}
		return todoRunStartResult{Status: "started", SessionID: session}, nil
	}
	t.Cleanup(func() { run.Start = old })
	return &got, &called
}

func TestTodoAPIPlanApprove(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)

	body, _ := json.Marshal(todoApprovePayload{Ref: todos.TODOReference(created)})
	rec := httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Todo.Status != types.StatusPending {
		t.Errorf("status = %s, want pending", resp.Todo.Status)
	}

	// A second approve hits a todo no longer in review: 409, not a silent write.
	rec = httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("second approve status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
}

// Approving with --run chains the implement step: the dialog's options are the
// request layer, but the step is the action's own.
func TestTodoAPIPlanApproveAndRun(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	got, _ := stubRunStart(t)

	body, _ := json.Marshal(todoApprovePayload{
		Ref: todos.TODOReference(created),
		Run: true,
		Options: &todoRunPayload{
			Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}},
		},
	})
	rec := httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Step != "run" {
		t.Errorf("chained step = %q, want run", got.Options.Step)
	}
	if got.Todo == nil || got.Todo.ID != created.ID {
		t.Errorf("chained run todo = %+v, want the approved todo", got.Todo)
	}
	if got.Options.Request.Name != "claude" || got.Options.Request.Mode != api.ModeAgent {
		t.Errorf("request layer = %+v, want the dialog's model", got.Options.Request.Model)
	}
	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Run == nil || resp.Run.Status != "started" || resp.Run.Step != "run" {
		t.Errorf("run response = %+v, want a started run step", resp.Run)
	}

	// Options naming another step contradict the action and are refused.
	second := seedReviewTodo(t, workDir, types.StatusReview)
	body, _ = json.Marshal(todoApprovePayload{
		Ref: todos.TODOReference(second), Run: true, Options: &todoRunPayload{Step: "plan"},
	})
	rec = httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "options.step") {
		t.Fatalf("approve with options.step=plan: status = %d, body = %q, want 400 naming options.step", rec.Code, rec.Body.String())
	}
}

func TestTodoAPIPlanReject(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	// Record a plan pointer so we can assert reject clears it.
	planPath := "/plans/plan-x.md"
	planNew := types.PlanNew
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{PlanPath: &planPath, PlanStatus: &planNew}); err != nil {
		t.Fatalf("seed plan: %v", err)
	}

	body, _ := json.Marshal(todoRejectPayload{Ref: todos.TODOReference(created)})
	rec := httptest.NewRecorder()
	s.handleTodoPlanReject(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/reject", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Todo.Status != types.StatusPending {
		t.Errorf("status = %s, want pending", resp.Todo.Status)
	}
	if resp.Todo.PlanPath != "" {
		t.Errorf("plan pointer not cleared: %q", resp.Todo.PlanPath)
	}

	// A second reject hits a todo no longer in review: 409.
	rec = httptest.NewRecorder()
	s.handleTodoPlanReject(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/reject", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("second reject status = %d, want 409", rec.Code)
	}
}

// Revising resumes the plan session with the feedback as its next turn; a
// persisted plan without an agent session plans afresh with the feedback in the
// todo's own prompt.
func TestTodoAPIPlanRevise(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	sid := "sess-revise-1"
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	got, _ := stubRunStart(t)

	const feedback = "use a bounded queue, not unbounded"
	body, _ := json.Marshal(todoRevisePayload{
		Ref:      todos.TODOReference(created),
		Feedback: feedback,
		Options:  &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Step != "plan" || !got.Options.Resume || got.Options.Message != feedback {
		t.Errorf("revise dispatched %+v, want the plan step resumed with the feedback", got.Options)
	}
	var resp todoAnswerResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "revising" || resp.SessionID != sid {
		t.Errorf("response = %+v", resp)
	}

	noSession := seedReviewTodo(t, workDir, types.StatusReview)
	body2, _ := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(noSession), Feedback: "x"})
	rec = httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body2))))
	if rec.Code != http.StatusOK {
		t.Fatalf("no-session revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Resume || got.Options.Message != "" {
		t.Errorf("fresh plan revision = %+v, want no resume and no message", got.Options)
	}
	if got.Options.Step != "plan" {
		t.Errorf("fresh plan revision step = %q, want plan", got.Options.Step)
	}
	if got.Todo.Prompt != "Revise the existing plan using this reviewer feedback:\n\nx" {
		t.Errorf("fresh plan revision prompt = %q", got.Todo.Prompt)
	}
}

// seedActivePhase records the phase the todo's current attempt runs, the way the
// native runtime's phase index marks it: the latest run of that phase is the
// todo's active pointer, parked at waiting by the ask outcome.
func seedActivePhase(todo *types.TODO, phase types.Phase) {
	todo.PhaseRuns = types.PhaseRuns{phase: {
		Phase: phase, State: string(captaindb.PromptRunStateWaiting), Active: true,
	}}
}

// Answering resumes the step that asked, with the answer as the next turn.
func TestTodoAPIAnswer(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusAsk)
	sid := "sess-answer-1"
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	seedActivePhase(created, types.PlanPhase)
	got, _ := stubRunStart(t)

	body, _ := json.Marshal(todoAnswerPayload{
		Ref:     todos.TODOReference(created),
		Answer:  "use postgres",
		Options: &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Message != "use postgres" {
		t.Errorf("answer = %q", got.Options.Message)
	}
	if got.Options.Step != "plan" {
		t.Errorf("resumed step = %q, want the plan step that asked", got.Options.Step)
	}
	if !got.Options.Resume {
		t.Error("resume flag not set")
	}
	var resp todoAnswerResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.SessionID != sid || resp.Status != "resumed" {
		t.Errorf("response = %+v", resp)
	}
}

func TestTodoAPIAnswerRejectsWrongState(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusPending)

	body, _ := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), Answer: "hello"})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("answer status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}

	// Ask status but no recorded session: also a 409 (nothing to resume).
	ask := types.StatusAsk
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{Status: &ask}); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("no-session answer status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
}

// seedAskTodoWithRun seeds an ask todo whose asking turn — a run of the given
// phase — left the given prompt run behind: the shape every dashboard answer
// starts from.
func seedAskTodoWithRun(t *testing.T, workDir, sessionID string, phase types.Phase, run *captaindb.PromptRun) *types.TODO {
	t.Helper()
	created := seedReviewTodo(t, workDir, types.StatusAsk)
	provider := uiTestProviderFor(workDir)
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sessionID}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	seedActivePhase(created, phase)
	provider.activeRun = run
	provider.comments = nil
	return created
}

// A codex rollout answered by the dashboard must resume on codex. The answer box
// sends no run options, so without inheriting the asking run's recorded runtime
// the payload defaults would hand the session to a cmux/claude driver.
func TestTodoAPIAnswerInheritsAskingRunRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	const sid = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3836"
	const codexModel = "gpt-5.6-sol"
	created := seedAskTodoWithRun(t, workDir, sid, types.RunPhase, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Mode: "run",
			Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Mode: "agent", Model: codexModel, Effort: "high",
			},
		},
	})
	gotReq, called := stubRunStart(t)

	body, _ := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), Answer: "use postgres"})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if !*called {
		t.Fatal("resume was never dispatched")
	}
	if gotReq.Options.Step != "run" {
		t.Errorf("resumed step = %q, want the run step that asked", gotReq.Options.Step)
	}
	spec := resolvedRun(t, *gotReq).Spec
	if providerKey(spec.Model) != "openai" || spec.Name != codexModel || spec.Mode != api.ModeAgent || string(spec.Effort) != "high" {
		t.Errorf("runtime = %s/%s/%s/%s, want openai/%s/agent/high — a codex session must not resume under claude",
			providerKey(spec.Model), spec.Name, spec.Mode, spec.Effort, codexModel)
	}
	if spec.SessionID != sid {
		t.Errorf("session id = %q, want the todo's recorded session %q", spec.SessionID, sid)
	}
}

// Explicit dashboard options still win over the asking run's runtime.
func TestTodoAPIAnswerOptionsOverrideInheritedRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedAskTodoWithRun(t, workDir, "sess-answer-override", types.RunPhase, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Resolved: captaindb.PromptRunRuntimeSelection{Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol", Effort: "high"},
		},
	})
	gotReq, _ := stubRunStart(t)

	raw, err := json.Marshal(todoAnswerPayload{
		Ref:    todos.TODOReference(created),
		Answer: "use postgres",
		Options: &todoRunPayload{
			Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeCmux, Effort: "medium"}},
		},
	})
	if err != nil {
		t.Fatalf("marshal answer payload: %v", err)
	}
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(raw))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	spec := resolvedRun(t, *gotReq).Spec
	if spec.Mode != api.ModeCmux || spec.Name != "claude-sonnet-5" || string(spec.Effort) != "medium" {
		t.Errorf("runtime = %s/%s/%s, want the explicitly requested cmux/claude-sonnet-5/medium", spec.Mode, spec.Name, spec.Effort)
	}
}

// A running prompt run means a live agent still owns the session; answering it
// would spawn a second agent against that session. Reject it, and record nothing.
func TestTodoAPIAnswerRejectsLiveRun(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedAskTodoWithRun(t, workDir, "sess-answer-live", types.RunPhase, &captaindb.PromptRun{
		State:   captaindb.PromptRunStateRunning,
		Runtime: captaindb.PromptRunRuntime{Driver: "agent"},
	})
	_, called := stubRunStart(t)

	body, _ := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), Answer: "use postgres"})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("answer status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
	if *called {
		t.Error("a second agent was dispatched against a session a live run still owns")
	}
	if comments := uiTestProviderFor(workDir).comments; len(comments) != 0 {
		t.Errorf("rejected answer still recorded comments: %q", comments)
	}
}

// A resume that cannot start must fail as a 4xx the answer box renders — not a
// 200 that records the answer and leaves the todo parked forever.
func TestTodoAPIAnswerPreflightFailureLeavesNoComment(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedAskTodoWithRun(t, workDir, "sess-answer-preflight", types.RunPhase, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Resolved: captaindb.PromptRunRuntimeSelection{Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol", Effort: "high"},
		},
	})
	// An unreadable plan fails resolution: the recorded plan is part of the
	// subject every step is evaluated against, so the resume cannot be folded.
	planPath := t.TempDir()
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{PlanPath: &planPath}); err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	_, called := stubRunStart(t)

	body, _ := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), Answer: "use postgres"})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("answer status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("unmarshal error body: %v", err)
	}
	if strings.TrimSpace(errResp.Error) == "" {
		t.Errorf("no error reported to the answer box; body = %q", rec.Body.String())
	}
	if *called {
		t.Error("resume dispatched despite a failed pre-flight")
	}
	if comments := uiTestProviderFor(workDir).comments; len(comments) != 0 {
		t.Errorf("failed answer still recorded comments: %q", comments)
	}
	if created.Status != types.StatusAsk {
		t.Errorf("status = %s, want the todo left in ask so the answer can be retried", created.Status)
	}
}

func TestTodoAPIAnswerResumesStoppedAskSessionFromInProgressTodo(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// This test replaces TestMain's home, so it must supply the model itself:
	// resolving a run without one is now an error rather than a silent default.
	writeHomeModel(t, home)
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusInProgress)
	sid := "sess-zombie-ask"
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid}); err != nil {
		t.Fatal(err)
	}
	seedActivePhase(created, types.RunPhase)
	logPath, err := cmuxprov.SessionLogPath(workDir, sid)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"type":"assistant","sessionId":"sess-zombie-ask","message":{"stop_reason":"tool_use","content":[{"type":"tool_use","id":"ask-1","name":"AskUserQuestion","input":{"questions":[{"question":"Which database?"}]}}]}}` + "\n"
	if err := os.WriteFile(logPath, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := stubRunStart(t)

	body, _ := json.Marshal(todoAnswerPayload{
		Ref:     todos.TODOReference(created),
		Answers: map[string]any{"Which database?": "Postgres"},
	})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Message != `{"answers":{"Which database?":"Postgres"}}` {
		t.Fatalf("resume feedback = %q", got.Options.Message)
	}
}
