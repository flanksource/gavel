package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO Captain session hierarchy", Ordered, func() {
	var provider *Provider

	BeforeAll(func(ctx SpecContext) {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}

		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_todo_session_hierarchy",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

		workDir := GinkgoT().TempDir()
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "session-hierarchy", RootPath: workDir,
			Repositories: []string{"acme/session-hierarchy"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates the TODO UUID as the Captain aggregate root", func(ctx SpecContext) {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Own every operation session", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())

		todoID := uuid.MustParse(created.ID)
		root, err := provider.Captain().GetSession(ctx, todoID)
		Expect(err).NotTo(HaveOccurred())
		Expect(root.ID).To(Equal(todoID))
		Expect(root.ParentSessionID).To(BeNil())
		Expect(root.RootSessionID).To(BeNil())
		Expect(root.AgentType).To(Equal("todo"))
		Expect(root.Metadata).To(HaveKeyWithValue("links", HaveKeyWithValue("todo", created.ID)))
	})

	DescribeTable("tags each operation session and links it beneath the TODO root",
		func(ctx SpecContext, mode types.RunMode, tag string) {
			created, err := provider.Create(ctx, todos.CreateRequest{
				Title: "Track the " + tag + " operation", Status: types.StatusPending,
			})
			Expect(err).NotTo(HaveOccurred())
			preparation, err := provider.PrepareRun(ctx, created, todos.RunPreparation{
				Mode: mode, ExecutorName: "codex",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(preparation.SessionID).NotTo(BeEmpty())

			issue, err := provider.Repository().GetIssue(ctx, uuid.MustParse(created.ID))
			Expect(err).NotTo(HaveOccurred())
			Expect(issue.ActivePromptRunID).NotTo(BeNil())
			run, err := provider.Captain().GetPromptRun(ctx, *issue.ActivePromptRunID)
			Expect(err).NotTo(HaveOccurred())
			operation, err := provider.Captain().GetSession(ctx, run.SessionID)
			Expect(err).NotTo(HaveOccurred())
			Expect(operation.ParentSessionID).To(Equal(&issue.ID))
			Expect(operation.RootSessionID).To(Equal(&issue.ID))
			Expect(operation.AgentType).To(Equal(tag))
			Expect(operation.Metadata).To(HaveKeyWithValue("tags", ConsistOf("todo", tag)))
			Expect(operation.Metadata).To(HaveKeyWithValue("links", HaveKeyWithValue("todo", created.ID)))
		},
		Entry("plan", types.ModePlan, "plan"),
		Entry("run", types.ModeRun, "run"),
		Entry("verify", types.ModeVerify, "verify"),
	)

	It("records adding verification as a linked operation session", func(ctx SpecContext) {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Add durable verification", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		body := todos.UpsertVerificationFixture(created.MarkdownBody, "```yaml test\npaths: [./todos/runtime]\n```")
		Expect(provider.Edit(ctx, created, todos.EditRequest{Body: &body})).To(Succeed())

		todoID := uuid.MustParse(created.ID)
		thread, err := provider.Captain().ListThreadSessionOverviews(ctx, todoID)
		Expect(err).NotTo(HaveOccurred())
		Expect(thread).To(HaveLen(2))
		Expect(thread[0].ID).To(Equal(todoID))
		Expect(thread[1].ParentSessionID).To(Equal(&todoID))
		Expect(thread[1].RootSessionID).To(Equal(&todoID))
		Expect(thread[1].AgentType).NotTo(BeNil())
		Expect(*thread[1].AgentType).To(Equal(todoSessionAddVerification))
		var metadata map[string]any
		Expect(json.Unmarshal(thread[1].Metadata, &metadata)).To(Succeed())
		Expect(metadata).To(HaveKeyWithValue("tags", ConsistOf(todoSessionType, todoSessionAddVerification)))
		Expect(metadata).To(HaveKeyWithValue("links", HaveKeyWithValue("todo", created.ID)))
	})

	It("places the provider transcript below its operation session", func(ctx SpecContext) {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Nest the provider transcript", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		preparation, err := provider.PrepareRun(ctx, created, todos.RunPreparation{
			Mode: types.ModeRun, ExecutorName: "codex",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(preparation.SessionID).NotTo(BeEmpty())
		Expect(provider.RecordRunStart(ctx, created, todos.RunStartMetadata{
			SessionID: "todo-hierarchy-provider", Provider: "openai",
			Backend: "codex-agent", Mode: "run",
		})).To(Succeed())

		issue, err := provider.Repository().GetIssue(ctx, uuid.MustParse(created.ID))
		Expect(err).NotTo(HaveOccurred())
		run, err := provider.Captain().GetPromptRun(ctx, *issue.ActivePromptRunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(run.ExecutionSessionID).NotTo(BeNil())
		execution, err := provider.Captain().GetSession(ctx, *run.ExecutionSessionID)
		Expect(err).NotTo(HaveOccurred())
		Expect(execution.ParentSessionID).To(Equal(&run.SessionID))
		Expect(execution.RootSessionID).To(Equal(&issue.ID))

		thread, err := provider.Captain().ListThreadSessionOverviews(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(thread).To(HaveLen(3))
		Expect(thread[0].ID).To(Equal(issue.ID))
	})
})
