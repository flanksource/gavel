package ui

import (
	"context"
	"net/http"

	captaindb "github.com/flanksource/captain/pkg/database"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GET /api/sessions", func() {
	var (
		f       *projectGitFixture
		tracked trackedGitServer
	)

	BeforeEach(func(ctx SpecContext) {
		f = newProjectGitFixture()
		tracked = newTrackedGitServer()
		provider, err := todoruntime.New(ctx, tracked.db, todoruntime.WorkspaceOptions{
			Name: "acme", RootPath: f.repo, Repositories: []string{"acme/acme"},
		})
		Expect(err).NotTo(HaveOccurred())
		previous := openTodoProvider
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		DeferCleanup(func() { openTodoProvider = previous })

		_, err = provider.Captain().CreateOrGetSession(ctx, captaindb.CreateSessionInput{
			ID: uuid.New(), Source: "claude", Provider: "anthropic", CWD: f.worktree, Title: "Add b",
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("answers 503 naming the missing database when the process has no git state tracker", func() {
		recorder := serveProjectGit(&Server{ctx: gavelctx.New(context.Background())}, http.MethodGet, "/api/sessions", nil)

		Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
		Expect(decodeProjectGit[map[string]string](recorder)).To(Equal(map[string]string{"error": gavelctx.ErrNoGitTracker.Error()}))
	})

	It("joins a session to its worktree's stored git state without running git", func() {
		first := serveProjectGit(tracked.server, http.MethodGet, "/api/sessions", nil)
		Expect(first.Code).To(Equal(http.StatusOK), first.Body.String())

		withoutGit()
		recorder := serveProjectGit(tracked.server, http.MethodGet, "/api/sessions", nil)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(recorder.Body.String()).To(Equal(first.Body.String()))
		response := decodeProjectGit[sessionsResponse](recorder)
		Expect(response.Errors).To(BeEmpty())
		Expect(response.Sessions).To(HaveLen(1))
		row := response.Sessions[0]
		Expect(row.Project).To(Equal("acme"))
		Expect(row.Worktree).To(Equal(&sessionWorktree{Path: f.worktree, Branch: pgFeature, Base: "main"}))
		Expect(row.Git).To(Equal(&sessionGit{
			Changes: gitstate.Changes{Unstaged: 1, Adds: 1}, Ahead: 2, WorktreeLive: true, BaseCheckedOut: true,
		}))
	})
})
