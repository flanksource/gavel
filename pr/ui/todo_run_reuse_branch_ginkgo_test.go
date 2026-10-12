package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run reuseBranch", func() {
	const runBranch = "shell/0123abcd"
	var (
		server   *Server
		provider *uiTestTODOProvider
		todo     *types.TODO
	)
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		provider = uiTestProviderFor(dir)
		var err error
		todo, err = provider.Create(context.Background(), todos.CreateRequest{Title: "Address review", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		server = &Server{ghOpts: github.Options{WorkDir: dir}}
		original := run.Start
		DeferCleanup(func() { run.Start = original })
		run.Start = func(run.Request) (run.StartResult, error) {
			Fail("a preview must not start a run")
			return run.StartResult{}, nil
		}
	})

	preview := func(reuse bool) *httptest.ResponseRecorder {
		GinkgoHelper()
		body, err := json.Marshal(map[string]any{"ref": todo.ID, "step": "run", "reuseBranch": reuse})
		Expect(err).NotTo(HaveOccurred())
		recorder := httptest.NewRecorder()
		server.handleTodoRunPreview(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		return recorder
	}

	It("maps the payload's reuseBranch onto the run options", func() {
		options, err := buildTodoRunOptions(todoRunPayload{ReuseBranch: true}, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(options.ReuseBranch).To(BeTrue())
	})

	It("previews the run on the previous run's branch", func() {
		provider.runWorktree = &native.RunWorktree{PromptRunID: uuid.New(), Worktree: api.WorktreeState{
			Branch: runBranch, Setup: "1111111aaaa", Head: "2222222bbbb", Removed: true,
		}}

		recorder := preview(true)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var response todoRunPreviewResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Spec.Setup.Checkout.Worktree.Mode).To(Equal(shell.WorktreeBranch))
		Expect(response.Spec.Setup.Checkout.Worktree.Branch).To(Equal(runBranch))
		Expect(response.Spec.Prompt.AppendSystem).To(ContainSubstring("This run continues on branch `" + runBranch + "`"))
		Expect(response.Prompt).NotTo(ContainSubstring(runBranch))
	})

	It("answers 400 when the todo has no branch to reuse", func() {
		recorder := preview(true)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("has no reusable branch"))
	})
})
