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
	"github.com/flanksource/gavel/todos/drivers"
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

func TestTodoAPIPlanApproveAndRun(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)

	oldStart := startTodoRun
	var got todoRunRequest
	startTodoRun = func(req todoRunRequest) error {
		got = req
		return nil
	}
	t.Cleanup(func() { startTodoRun = oldStart })

	body, _ := json.Marshal(todoApprovePayload{
		Ref: todos.TODOReference(created),
		Run: true,
		Options: &todoRunPayload{
			Driver: "claude-headless",
			Spec:   api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}},
			// Even a stale plan runMode is forced back to run for the chained run.
			RunMode: "plan",
		},
	})
	rec := httptest.NewRecorder()
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.RunMode != types.ModeRun {
		t.Errorf("chained run mode = %q, want run", got.Options.RunMode)
	}
	if len(got.Todos) != 1 {
		t.Errorf("chained run todos = %d, want 1", len(got.Todos))
	}
	var resp todoApproveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Run == nil || resp.Run.Status != "started" {
		t.Errorf("run response = %+v, want started", resp.Run)
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

func TestTodoAPIPlanRevise(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusReview)
	sid := "sess-revise-1"
	mode := types.ModePlan
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	oldStart := startTodoAnswer
	oldStartRun := startTodoRun
	var gotReq todoRunRequest
	var gotFeedback string
	var gotFreshReq todoRunRequest
	startTodoAnswer = func(req todoRunRequest, feedback string) error {
		gotReq = req
		gotFeedback = feedback
		return nil
	}
	startTodoRun = func(req todoRunRequest) error {
		gotFreshReq = req
		return nil
	}
	t.Cleanup(func() {
		startTodoAnswer = oldStart
		startTodoRun = oldStartRun
	})

	body, _ := json.Marshal(todoRevisePayload{
		Ref:      todos.TODOReference(created),
		Feedback: "use a bounded queue, not unbounded",
		Options:  &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if gotFeedback != "use a bounded queue, not unbounded" {
		t.Errorf("feedback = %q", gotFeedback)
	}
	if gotReq.Options.RunMode != types.ModePlan {
		t.Errorf("revise mode = %q, want plan", gotReq.Options.RunMode)
	}
	if !gotReq.Options.Resume {
		t.Error("resume flag not set")
	}
	var resp todoAnswerResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Status != "revising" || resp.SessionID != sid {
		t.Errorf("response = %+v", resp)
	}

	// A persisted plan without an agent session starts a fresh plan run.
	noSession := seedReviewTodo(t, workDir, types.StatusReview)
	body2, _ := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(noSession), Feedback: "x"})
	rec = httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body2))))
	if rec.Code != http.StatusOK {
		t.Fatalf("no-session revise status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if gotFreshReq.Options.Resume {
		t.Error("fresh plan revision unexpectedly requested resume")
	}
	if gotFreshReq.Options.RunMode != types.ModePlan {
		t.Errorf("fresh plan revision mode = %q, want plan", gotFreshReq.Options.RunMode)
	}
	if gotFreshReq.Todos[0].Prompt != "Revise the existing plan using this reviewer feedback:\n\nx" {
		t.Errorf("fresh plan revision prompt = %q", gotFreshReq.Todos[0].Prompt)
	}
}

func TestTodoAPIAnswer(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusAsk)
	sid := "sess-answer-1"
	mode := types.ModePlan
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	oldStart := startTodoAnswer
	var gotReq todoRunRequest
	var gotAnswer string
	startTodoAnswer = func(req todoRunRequest, answer string) error {
		gotReq = req
		gotAnswer = answer
		return nil
	}
	t.Cleanup(func() { startTodoAnswer = oldStart })

	body, _ := json.Marshal(todoAnswerPayload{
		Ref:     todos.TODOReference(created),
		Answer:  "use postgres",
		Options: &todoRunPayload{Driver: "claude-headless", Spec: api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if gotAnswer != "use postgres" {
		t.Errorf("answer = %q", gotAnswer)
	}
	if gotReq.Options.RunMode != types.ModePlan {
		t.Errorf("resume mode = %q, want the todo's recorded plan mode", gotReq.Options.RunMode)
	}
	if !gotReq.Options.Resume {
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

// seedAskTodoWithRun seeds an ask todo whose asking turn left the given prompt
// run behind — the shape every dashboard answer starts from.
func seedAskTodoWithRun(t *testing.T, workDir, sessionID string, mode types.RunMode, run *captaindb.PromptRun) *types.TODO {
	t.Helper()
	created := seedReviewTodo(t, workDir, types.StatusAsk)
	provider := uiTestProviderFor(workDir)
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sessionID, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	provider.activeRun = run
	provider.comments = nil
	return created
}

func stubTodoAnswer(t *testing.T) (*todoRunRequest, *bool) {
	t.Helper()
	var got todoRunRequest
	called := false
	old := startTodoAnswer
	startTodoAnswer = func(req todoRunRequest, _ string) error {
		got, called = req, true
		return nil
	}
	t.Cleanup(func() { startTodoAnswer = old })
	return &got, &called
}

// A codex rollout answered by the dashboard must resume on codex. The answer box
// sends no run options, so without inheriting the asking run's recorded runtime
// the payload defaults would hand the session to a cmux/claude driver.
func TestTodoAPIAnswerInheritsAskingRunRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	const sid = "019fa17d-622a-7ef3-b8ad-d8b1d7cd3836"
	const codexModel = "gpt-5.6-sol"
	created := seedAskTodoWithRun(t, workDir, sid, types.ModeRun, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Mode:   "run",
			Driver: "headless-codex",
			Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Backend: "codex-agent", Model: codexModel, Effort: "high",
			},
		},
	})
	gotReq, called := stubTodoAnswer(t)

	body, _ := json.Marshal(todoAnswerPayload{Ref: todos.TODOReference(created), Answer: "use postgres"})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if !*called {
		t.Fatal("resume was never dispatched")
	}
	if gotReq.Options.Driver != string(drivers.Cli) {
		t.Errorf("driver = %q, want %q (the asking run's headless mechanism)", gotReq.Options.Driver, drivers.Cli)
	}
	if gotReq.Options.Agent != "codex" {
		t.Errorf("agent = %q, want codex — a codex session must not resume under claude", gotReq.Options.Agent)
	}
	if string(gotReq.Options.Backend) != "codex-agent" {
		t.Errorf("backend = %q, want codex-agent", gotReq.Options.Backend)
	}
	if gotReq.Options.Name != codexModel {
		t.Errorf("model = %q, want %q", gotReq.Options.Name, codexModel)
	}
	if string(gotReq.Options.Effort) != "high" {
		t.Errorf("effort = %q, want high", gotReq.Options.Effort)
	}
	if gotReq.Options.SessionID != sid {
		t.Errorf("session id = %q, want the todo's recorded session %q", gotReq.Options.SessionID, sid)
	}
}

// Explicit dashboard options still win over the asking run's runtime.
func TestTodoAPIAnswerOptionsOverrideInheritedRuntime(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedAskTodoWithRun(t, workDir, "sess-answer-override", types.ModeRun, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Driver:   "headless-codex",
			Resolved: captaindb.PromptRunRuntimeSelection{Backend: "codex-agent", Model: "gpt-5.6-sol", Effort: "high"},
		},
	})
	gotReq, _ := stubTodoAnswer(t)

	// Sent as wire JSON, not a marshaled todoRunPayload: api.Spec's promoted
	// MarshalJSON emits only the spec fields, so marshaling the struct would drop
	// the very driver override under test.
	body := `{"ref":"` + todos.TODOReference(created) + `","answer":"use postgres",` +
		`"options":{"driver":"cmux","model":"claude-sonnet-5","backend":"claude-cmux","effort":"medium"}}`
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if gotReq.Options.Driver != string(drivers.Cmux) {
		t.Errorf("driver = %q, want the explicitly requested cmux", gotReq.Options.Driver)
	}
	if gotReq.Options.Name != "claude-sonnet-5" || string(gotReq.Options.Effort) != "medium" {
		t.Errorf("model/effort = %q/%q, want the explicitly requested claude-sonnet-5/medium", gotReq.Options.Name, gotReq.Options.Effort)
	}
}

// A running prompt run means a live agent still owns the session; answering it
// would spawn a second agent against that session. Reject it, and record nothing.
func TestTodoAPIAnswerRejectsLiveRun(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedAskTodoWithRun(t, workDir, "sess-answer-live", types.ModeRun, &captaindb.PromptRun{
		State:   captaindb.PromptRunStateRunning,
		Runtime: captaindb.PromptRunRuntime{Driver: "headless-codex"},
	})
	_, called := stubTodoAnswer(t)

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
	created := seedAskTodoWithRun(t, workDir, "sess-answer-preflight", types.ModeRun, &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Driver:   "headless-codex",
			Resolved: captaindb.PromptRunRuntimeSelection{Backend: "codex-agent", Model: "gpt-5.6-sol", Effort: "high"},
		},
	})
	// An unreadable plan fails executor construction: the recorded plan is an
	// input to every run and plan mode, so the resume cannot be built.
	planPath := t.TempDir()
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{PlanPath: &planPath}); err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	_, called := stubTodoAnswer(t)

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
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusInProgress)
	sid := "sess-zombie-ask"
	mode := types.ModeRun
	if err := uiTestProviderFor(workDir).UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sid, RunMode: &mode}); err != nil {
		t.Fatal(err)
	}
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

	oldStart := startTodoAnswer
	var gotAnswer string
	startTodoAnswer = func(_ todoRunRequest, answer string) error { gotAnswer = answer; return nil }
	t.Cleanup(func() { startTodoAnswer = oldStart })

	body, _ := json.Marshal(todoAnswerPayload{
		Ref:     todos.TODOReference(created),
		Answers: map[string]any{"Which database?": "Postgres"},
	})
	rec := httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("answer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if gotAnswer != `{"answers":{"Which database?":"Postgres"}}` {
		t.Fatalf("resume feedback = %q", gotAnswer)
	}
}
