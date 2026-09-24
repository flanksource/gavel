package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/session"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The Session tab reads a run's transcript from Captain's own session handler,
// mounted on the dashboard's mux and backed by the pool Gavel shares with
// Captain's CLI registry. Gavel's own endpoint shrank to the attempt list that
// names those sessions.
var _ = Describe("todo sessions served through Captain", Ordered, func() {
	const (
		providerSessionID = "0199f0aa-0000-7000-8000-00000000c0de"
		transcriptText    = "shadow activities by age"
	)
	var (
		handler http.Handler
		workDir string
		todo    *types.TODO
		promptR uuid.UUID
		execID  uuid.UUID
	)

	BeforeAll(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_captain_sessions"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		// The same wiring internal/database.Shared performs for `gavel serve`:
		// Captain's CLI registry reads through Gavel's pool. It is process-wide
		// and refuses a second pool, so this is the only spec that configures it;
		// TestMain clears the DSN environment so nothing else in the run can have
		// configured it first.
		Expect(captaincli.ConfigureNativeDatabase(opened.Gorm())).To(Succeed())

		workDir = GinkgoT().TempDir()
		provider, err := todoruntime.New(ctx, opened.Gorm(), todoruntime.WorkspaceOptions{
			Name: "captain-sessions", RootPath: workDir, Repositories: []string{"acme/captain-sessions"},
		})
		Expect(err).NotTo(HaveOccurred())
		previous := openTodoProvider
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		DeferCleanup(func() { openTodoProvider = previous })

		todo, err = provider.Create(ctx, todos.CreateRequest{Title: "Plan the shadow report", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{Mode: types.ModePlan, Prompt: "plan", ExecutorName: "claude"})
		Expect(err).NotTo(HaveOccurred())
		promptR = admission.PromptRunID
		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{SessionID: providerSessionID, Provider: "claude", Mode: "plan"})).To(Succeed())
		run, err := provider.Captain().GetPromptRun(ctx, promptR)
		Expect(err).NotTo(HaveOccurred())
		Expect(run.ExecutionSessionID).NotTo(BeNil())
		execID = *run.ExecutionSessionID
		Expect(provider.Captain().PutChatMessage(ctx, captaindb.PutChatMessageInput{
			SessionID: execID, ProviderMessageID: "assistant-1", Role: "assistant",
			Parts: json.RawMessage(`[{"type":"text","text":"` + transcriptText + `"}]`),
		})).To(Succeed())

		handler = (&Server{ghOpts: github.Options{WorkDir: workDir}}).Handler()
	})

	get := func(path string) *httptest.ResponseRecorder {
		GinkgoHelper()
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		return recorder
	}

	It("lists the todo's attempts with the session ids the Session tab loads", func() {
		recorder := get("/api/todos/session/detail?ref=" + url.QueryEscape(todo.ID) + "&dir=" + url.QueryEscape(workDir))

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var body map[string]json.RawMessage
		Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(Succeed())
		Expect(body).To(HaveLen(1), "the response is the attempt list and nothing else: %s", recorder.Body.String())
		var attempts []struct {
			PromptRunID        string `json:"promptRunId"`
			Ordinal            int    `json:"ordinal"`
			Step               string `json:"step"`
			ProviderSessionID  string `json:"providerSessionId"`
			ExecutionSessionID string `json:"executionSessionId"`
		}
		Expect(json.Unmarshal(body["attempts"], &attempts)).To(Succeed())
		Expect(attempts).To(HaveLen(1))
		Expect(attempts[0].PromptRunID).To(Equal(promptR.String()))
		Expect(attempts[0].Ordinal).To(Equal(1))
		Expect(attempts[0].Step).To(Equal("plan"))
		Expect(attempts[0].ProviderSessionID).To(Equal(providerSessionID))
		Expect(attempts[0].ExecutionSessionID).To(Equal(execID.String()))
	})

	It("serves the attempt's unified session from the mounted Captain handler", func() {
		recorder := get("/api/captain/sessions/" + execID.String())

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var detail session.Session
		Expect(json.Unmarshal(recorder.Body.Bytes(), &detail)).To(Succeed())
		texts := []string{}
		for _, message := range detail.Messages {
			for _, part := range message.Parts {
				texts = append(texts, part.Text)
			}
		}
		Expect(texts).To(ContainElement(transcriptText))
	})

	It("answers an unknown session with Captain's 404", func() {
		recorder := get("/api/captain/sessions/" + uuid.NewString())

		Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
	})

	It("no longer serves the retired session stream route", func() {
		recorder := get("/api/todos/session/stream?sessionId=" + providerSessionID)

		Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
	})
})
