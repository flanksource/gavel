package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo plan review", func() {
	It("starts a fresh plan run with feedback when the persisted plan has no agent session", func() {
		workDir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: workDir}}
		created, err := uiTestProviderFor(workDir).Create(GinkgoT().Context(), todos.CreateRequest{
			Title:  "Reviewable persisted plan",
			Status: types.StatusReview,
		})
		Expect(err).NotTo(HaveOccurred())
		planStatus := types.PlanNew
		Expect(uiTestProviderFor(workDir).UpdateState(GinkgoT().Context(), created, todos.StateUpdate{PlanStatus: &planStatus})).To(Succeed())

		var request todoRunRequest
		previousStartRun := startTodoRun
		previousStartAnswer := startTodoAnswer
		startTodoRun = func(got todoRunRequest) error {
			request = got
			return nil
		}
		startTodoAnswer = func(todoRunRequest, string) error {
			Fail("a plan without an agent session cannot use the resume path")
			return nil
		}
		DeferCleanup(func() {
			startTodoRun = previousStartRun
			startTodoAnswer = previousStartAnswer
		})

		feedback := "Keep the migration reversible."
		body, err := json.Marshal(todoRevisePayload{Ref: todos.TODOReference(created), Feedback: feedback})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server.handleTodoPlanRevise(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/plan/revise", strings.NewReader(string(body))))

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(request.Options.Resume).To(BeFalse())
		Expect(request.Options.RunMode).To(Equal(types.ModePlan))
		Expect(request.Todos).To(HaveLen(1))
		Expect(request.Todos[0].Prompt).To(ContainSubstring(feedback))
	})
})
