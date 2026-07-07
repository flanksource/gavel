package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func seedReviewTodo(t *testing.T, workDir string, status types.Status) *types.TODO {
	t.Helper()
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Reviewable",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	if err := todos.UpdateTODOState(created, todos.StateUpdate{Status: &status}); err != nil {
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
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve?provider=todos", strings.NewReader(string(body))))
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
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve?provider=todos", strings.NewReader(string(body))))
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
	s.handleTodoPlanApprove(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/approve?provider=todos", strings.NewReader(string(body))))
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
	if err := todos.UpdateTODOState(created, todos.StateUpdate{PlanPath: &planPath, PlanStatus: &planNew}); err != nil {
		t.Fatalf("seed plan: %v", err)
	}

	body, _ := json.Marshal(todoRejectPayload{Ref: todos.TODOReference(created)})
	rec := httptest.NewRecorder()
	s.handleTodoPlanReject(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/reject?provider=todos", strings.NewReader(string(body))))
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
	s.handleTodoPlanReject(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/reject?provider=todos", strings.NewReader(string(body))))
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
	if err := todos.UpdateTODOState(created, todos.StateUpdate{SessionID: &sid, RunMode: &mode}); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	oldStart := startTodoAnswer
	var gotReq todoRunRequest
	var gotFeedback string
	startTodoAnswer = func(req todoRunRequest, feedback string) error {
		gotReq = req
		gotFeedback = feedback
		return nil
	}
	t.Cleanup(func() { startTodoAnswer = oldStart })

	body, _ := json.Marshal(todoRevisePayload{
		Ref:      todos.TODOReference(created),
		Feedback: "use a bounded queue, not unbounded",
		Options:  &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Effort: "medium"}}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise?provider=todos", strings.NewReader(string(body))))
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

	// A revise without a recorded session is a 409 (nothing to resume).
	noSession := seedReviewTodo(t, workDir, types.StatusReview)
	body2, _ := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(noSession), Feedback: "x"})
	rec = httptest.NewRecorder()
	s.handleTodoPlanRevise(rec, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise?provider=todos", strings.NewReader(string(body2))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("no-session revise status = %d, want 409", rec.Code)
	}
}

func TestTodoAPIAnswer(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created := seedReviewTodo(t, workDir, types.StatusAsk)
	sid := "sess-answer-1"
	mode := types.ModePlan
	if err := todos.UpdateTODOState(created, todos.StateUpdate{SessionID: &sid, RunMode: &mode}); err != nil {
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
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer?provider=todos", strings.NewReader(string(body))))
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
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("answer status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}

	// Ask status but no recorded session: also a 409 (nothing to resume).
	ask := types.StatusAsk
	if err := todos.UpdateTODOState(created, todos.StateUpdate{Status: &ask}); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	s.handleTodoAnswer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/answer?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusConflict {
		t.Fatalf("no-session answer status = %d, want 409; body = %q", rec.Code, rec.Body.String())
	}
}
