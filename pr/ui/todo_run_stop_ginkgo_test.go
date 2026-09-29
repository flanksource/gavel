package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/flanksource/captain/pkg/ai/approval"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Stopping a run nobody drives any more goes through the provider's reclaim,
// which settles the run through captain. Captain's cancel writes the terminal
// state first and then withdraws the run's pending approvals, so the broker
// parked on one wakes to a finished run and leaves it finished. Gavel writes no
// captain row of its own on the way.
var _ = Describe("stopping an abandoned todo run", func() {
	type toolOutcome struct {
		decision api.ApprovalDecision
		err      error
	}

	var (
		provider *todoruntime.Provider
		server   *Server
		workDir  string
		todo     *types.TODO
		promptR  uuid.UUID
	)

	BeforeEach(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_run_stop"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

		workDir = GinkgoT().TempDir()
		provider, err = todoruntime.New(ctx, opened.Gorm(), todoruntime.WorkspaceOptions{
			Name: "run-stop", RootPath: workDir, Repositories: []string{"acme/run-stop"},
		})
		Expect(err).NotTo(HaveOccurred())
		previous := openTodoProvider
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		DeferCleanup(func() { openTodoProvider = previous })
		server = &Server{ghOpts: github.Options{WorkDir: workDir}}

		todo, err = provider.Create(ctx, todos.CreateRequest{Title: "Stop the abandoned run", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{Mode: types.ModePlan, Prompt: "plan", ExecutorName: "claude"})
		Expect(err).NotTo(HaveOccurred())
		Expect(admission.Record).NotTo(BeNil(), "PrepareRun must hand back the Recording captain admits the run with")
		// Admission alone, as promptrun.Run performs it before dispatch: the run
		// stays pending with no provider to drive it.
		admitted, err := promptrun.Admit(ctx, promptrun.Input{Record: admission.Record})
		Expect(err).NotTo(HaveOccurred())
		promptR = admitted.ID
		Expect(promptR).To(Equal(admission.PromptRunID))
		// The dispatcher that admitted the run is gone: its claim is released, so
		// nothing in any process is driving the run to a terminal state.
		owner, err := native.LocalOwner()
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.Repository().ReleasePromptRunOwner(ctx, promptR, owner.Token)).To(Succeed())
	})

	It("reclaims the run and withdraws the approval its broker is parked on", func(ctx SpecContext) {
		run, err := provider.Captain().GetPromptRun(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		raised := make(chan string, 1)
		broker := &approval.Broker{
			DB: provider.Captain(), SessionID: run.SessionID, PromptRunID: promptR,
			RequestedBy: "gavel-dashboard", Timeout: time.Hour, Poll: 10 * time.Millisecond,
			Notify: func(_ context.Context, event api.Event) error {
				if event.Reason == "" {
					raised <- event.ApprovalID
				}
				return nil
			},
		}
		outcomes := make(chan toolOutcome, 1)
		go func() {
			defer GinkgoRecover()
			decision, err := broker.OnApproval(ctx, api.ApprovalRequest{
				Tool: "Bash", Input: map[string]any{"command": "ls"}, ToolUseID: "toolu_run_stop",
				Kind: api.ApprovalKindTool,
			})
			outcomes <- toolOutcome{decision: decision, err: err}
		}()
		var approvalID string
		Eventually(raised, 5*time.Second).Should(Receive(&approvalID))

		body, err := json.Marshal(todoRunStopRequest{Ref: todo.ID, PromptRunID: promptR})
		Expect(err).NotTo(HaveOccurred())
		rec := httptest.NewRecorder()
		server.handleTodoRunStop(rec, httptest.NewRequest(http.MethodPost,
			"/api/todos/session/stop?dir="+url.QueryEscape(workDir), bytes.NewReader(body)))

		Expect(rec.Code).To(Equal(http.StatusAccepted), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring(`"status":"reclaimed"`))
		var got toolOutcome
		Eventually(outcomes, 5*time.Second).Should(Receive(&got))
		Expect(got.decision.Allow).To(BeFalse())
		Expect(got.err).To(HaveOccurred(), "a withdrawn approval must end the wait unanswered")
		request, err := provider.Captain().GetTurnRequest(ctx, uuid.MustParse(approvalID))
		Expect(err).NotTo(HaveOccurred())
		Expect(request.State).To(Equal(captaindb.TurnRequestStateCancelled))
		stopped, err := provider.Captain().GetPromptRun(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		Expect(stopped.State).To(Equal(captaindb.PromptRunStateCancelled),
			"the broker's release must leave the stopped run stopped")
	})
})
