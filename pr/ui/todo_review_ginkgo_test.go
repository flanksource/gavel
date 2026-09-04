package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// specTodo seeds one todo in the given status behind the package's test
// provider for workDir.
func specTodo(workDir string, status types.Status) *types.TODO {
	GinkgoHelper()
	created, err := uiTestProviderFor(workDir).Create(GinkgoT().Context(), todos.CreateRequest{
		Title: "Reviewable " + string(status), Status: status,
	})
	Expect(err).NotTo(HaveOccurred())
	return created
}

// specAskTodo seeds an ask todo whose asking turn, a run of the given phase,
// left a waiting codex prompt run behind.
func specAskTodo(workDir, sessionID string, phase types.Phase) *types.TODO {
	GinkgoHelper()
	created := specTodo(workDir, types.StatusAsk)
	provider := uiTestProviderFor(workDir)
	Expect(provider.UpdateState(GinkgoT().Context(), created, todos.StateUpdate{SessionID: &sessionID})).To(Succeed())
	seedActivePhase(created, phase)
	provider.activeRun = &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{Resolved: captaindb.PromptRunRuntimeSelection{
			Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol", Effort: "high",
		}},
	}
	provider.comments = nil
	return created
}

// stubSpecRunStart replaces the dispatcher for one spec, recording whether a
// run was started and answering with the outcome given.
func stubSpecRunStart(err error) *bool {
	GinkgoHelper()
	called := false
	previous := run.Start
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		called = true
		if err != nil {
			return todoRunStartResult{}, err
		}
		return todoRunStartResult{Status: "started", SessionID: run.PriorSessionID(req.Todo)}, nil
	}
	DeferCleanup(func() { run.Start = previous })
	return &called
}

// stubApprovalStore replaces the workspace's approval store for one spec.
func stubApprovalStore(err error) {
	GinkgoHelper()
	previous := todoApprovalStore
	todoApprovalStore = func(context.Context, string) (*captaindb.DB, error) { return nil, err }
	DeferCleanup(func() { todoApprovalStore = previous })
}

func postJSON(handler http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	GinkgoHelper()
	raw, err := json.Marshal(body)
	Expect(err).NotTo(HaveOccurred())
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw))))
	return rec
}

var _ = Describe("todo plan review", func() {
	var (
		workDir string
		server  *Server
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		server = &Server{ghOpts: github.Options{WorkDir: workDir}}
	})

	It("starts a fresh plan run with feedback when the persisted plan has no agent session", func() {
		created := specTodo(workDir, types.StatusReview)
		planStatus := types.PlanNew
		Expect(uiTestProviderFor(workDir).UpdateState(GinkgoT().Context(), created, todos.StateUpdate{PlanStatus: &planStatus})).To(Succeed())

		var request todoRunRequest
		previousStartRun := run.Start
		run.Start = func(got todoRunRequest) (todoRunStartResult, error) {
			request = got
			return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
		}
		DeferCleanup(func() { run.Start = previousStartRun })

		feedback := "Keep the migration reversible."
		rec := postJSON(server.handleTodoPlanRevise, "/api/todos/plan/revise",
			todoRevisePayload{Ref: todos.TODOReference(created), Feedback: feedback})

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		// A plan without an agent session cannot use the resume path: the feedback
		// travels in the todo's own prompt instead of as a resumed turn.
		Expect(request.Options.Resume).To(BeFalse())
		Expect(request.Options.Message).To(BeEmpty())
		Expect(request.Options.Step).To(Equal("plan"))
		Expect(request.Todo).NotTo(BeNil())
		Expect(request.Todo.Prompt).To(ContainSubstring(feedback))
		Expect(request.Prepared).NotTo(BeNil(), "the dispatch must carry the fold that was pre-flighted")
	})

	It("refuses contradicting run options before the plan is approved", func() {
		created := specTodo(workDir, types.StatusReview)
		called := stubSpecRunStart(nil)

		rec := postJSON(server.handleTodoPlanApprove, "/api/todos/plan/approve", todoApprovePayload{
			Ref: todos.TODOReference(created), Run: true, Options: &todoRunPayload{Step: "plan"},
		})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("options.step"))
		Expect(created.Status).To(Equal(types.StatusReview), "a refused request must change nothing")
		Expect(*called).To(BeFalse())
	})

	It("reports the approved plan together with a chained run that could not start", func() {
		created := specTodo(workDir, types.StatusReview)
		stubSpecRunStart(errors.New("agent binary missing"))

		rec := postJSON(server.handleTodoPlanApprove, "/api/todos/plan/approve", todoApprovePayload{
			Ref: todos.TODOReference(created), Run: true,
			Options: &todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent}}},
		})

		Expect(rec.Code).To(Equal(http.StatusInternalServerError), rec.Body.String())
		var resp todoApproveResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Error).To(ContainSubstring("agent binary missing"))
		Expect(resp.Run).To(BeNil())
		Expect(resp.Todo.Status).To(Equal(types.StatusPending), "the approval committed and the client must see it")
		Expect(created.Status).To(Equal(types.StatusPending))
		Expect(resp.Todo.Lifecycle).NotTo(BeNil(), "the approve response must carry the lifecycle, not just the status")
		Expect(resp.Todo.Lifecycle.Steps).NotTo(BeEmpty())
	})

	It("reports the requested revision together with a plan run that could not start", func() {
		created := specTodo(workDir, types.StatusReview)
		session := "sess-revise-fails"
		Expect(uiTestProviderFor(workDir).UpdateState(GinkgoT().Context(), created, todos.StateUpdate{SessionID: &session})).To(Succeed())
		stubSpecRunStart(errors.New("session log unreadable"))

		rec := postJSON(server.handleTodoPlanRevise, "/api/todos/plan/revise",
			todoRevisePayload{Ref: todos.TODOReference(created), Feedback: "split the migration"})

		Expect(rec.Code).To(Equal(http.StatusInternalServerError), rec.Body.String())
		var resp todoAnswerResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Status).To(Equal("failed"))
		Expect(resp.Error).To(ContainSubstring("session log unreadable"))
		Expect(resp.Todo.Ref).To(Equal(todos.TODOReference(created)))
		Expect(resp.Todo.Lifecycle).NotTo(BeNil(), "the revise-failure response must carry the lifecycle")
		Expect(resp.Todo.Lifecycle.Steps).NotTo(BeEmpty())
	})

	It("rejects a body carrying a key the endpoint does not read", func() {
		created := specTodo(workDir, types.StatusReview)

		rec := postJSON(server.handleTodoPlanApprove, "/api/todos/plan/approve",
			map[string]any{"ref": todos.TODOReference(created), "runMode": "plan"})

		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(SatisfyAll(ContainSubstring("invalid request body"), ContainSubstring("runMode")))
		Expect(created.Status).To(Equal(types.StatusReview))
	})
})

var _ = Describe("todo answer", func() {
	var (
		workDir string
		server  *Server
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		server = &Server{ghOpts: github.Options{WorkDir: workDir}}
	})

	answer := func(todo *types.TODO) *httptest.ResponseRecorder {
		GinkgoHelper()
		return postJSON(server.handleTodoAnswer, "/api/todos/answer",
			todoAnswerPayload{Ref: todos.TODOReference(todo), Answer: "use postgres"})
	}

	It("fails instead of dispatching when the approval store cannot be read", func() {
		created := specAskTodo(workDir, "sess-store-down", types.RunPhase)
		stubApprovalStore(errors.New("connection refused"))
		called := stubSpecRunStart(nil)

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusInternalServerError), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("connection refused"))
		Expect(*called).To(BeFalse(), "a liveness check that could not run must not let a second agent through")
		Expect(uiTestProviderFor(workDir).comments).To(BeEmpty())
	})

	It("resumes a workspace whose runtime keeps no Captain database", func() {
		created := specAskTodo(workDir, "sess-no-captain", types.RunPhase)
		stubApprovalStore(fmt.Errorf("%w: the TODO runtime for %s cannot record tool approvals", errNoCaptainDatabase, workDir))
		called := stubSpecRunStart(nil)

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(*called).To(BeTrue())
	})

	It("resumes the one phase the index marks, whatever class the last run reported", func() {
		created := specAskTodo(workDir, "sess-phase", types.TriagePhase)
		var got todoRunRequest
		previous := run.Start
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			got = req
			return todoRunStartResult{Status: "started", SessionID: "sess-phase"}, nil
		}
		DeferCleanup(func() { run.Start = previous })

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(got.Options.Step).To(Equal("triage"))
		Expect(got.Prepared).NotTo(BeNil())
		var resp todoAnswerResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Todo.Lifecycle).NotTo(BeNil(), "a resumed answer must carry the lifecycle")
		Expect(resp.Todo.Lifecycle.Steps).NotTo(BeEmpty())
	})

	It("refuses to guess between two resumable phases", func() {
		created := specAskTodo(workDir, "sess-two-phases", types.PlanPhase)
		created.PhaseRuns[types.RunPhase] = types.PhaseRun{Phase: types.RunPhase, State: "failed", Active: true}
		called := stubSpecRunStart(nil)

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("plan, run"))
		Expect(*called).To(BeFalse())
		Expect(uiTestProviderFor(workDir).comments).To(BeEmpty())
	})

	It("refuses a todo whose index marks no phase to resume", func() {
		created := specAskTodo(workDir, "sess-no-phase", types.PlanPhase)
		created.PhaseRuns = nil
		called := stubSpecRunStart(nil)

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("no active or waiting step run"))
		Expect(*called).To(BeFalse())
	})

	It("reports the recorded answer together with a resume that could not start", func() {
		created := specAskTodo(workDir, "sess-start-fails", types.RunPhase)
		stubSpecRunStart(errors.New("registry refused the run"))

		rec := answer(created)

		Expect(rec.Code).To(Equal(http.StatusInternalServerError), rec.Body.String())
		var resp todoAnswerResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).To(Succeed())
		Expect(resp.Status).To(Equal("failed"))
		Expect(resp.Error).To(ContainSubstring("registry refused the run"))
		Expect(resp.Todo.Status).To(Equal(types.StatusAsk), "the todo still asks; the answer is on its timeline")
		Expect(uiTestProviderFor(workDir).comments).To(HaveLen(1))
		Expect(resp.Todo.Lifecycle).NotTo(BeNil(), "the failed-answer response must carry the lifecycle")
		Expect(resp.Todo.Lifecycle.Steps).NotTo(BeEmpty())
	})
})
