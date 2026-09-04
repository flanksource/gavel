package runtime

import (
	"context"
	"os"
	"path/filepath"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("creating TODOs with durable plans", Ordered, func() {
	var (
		ctx      context.Context
		provider *Provider
	)

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}

		ctx = context.Background()
		dataDir := filepath.Join(GinkgoT().TempDir(), "postgres")
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  dataDir,
			Database: "gavel_todo_create_plan",
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
		repository, err := native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey:  "github.com/acme/create-plan",
			RootPath: workDir,
		})
		Expect(err).NotTo(HaveOccurred())
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "create-plan", RootPath: workDir, Repositories: []string{"acme/create-plan"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates a selected plan awaiting review and a persisted verification fixture", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title:        "Review the parser plan",
			Body:         "Parser failures lose context.",
			Verification: "```yaml test\npaths: [./pkg/parser]\nframework: [go]\n```",
			Priority:     types.PriorityHigh,
			Status:       types.StatusPending,
			Plan: &todos.CreatePlanRequest{
				Markdown: "# Parser plan\n\n1. Preserve parse context.",
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Status).To(Equal(types.StatusReview))
		Expect(created.PlanStatus).To(Equal(types.PlanNew))
		Expect(created.VerificationMarkdown).To(ContainSubstring("./pkg/parser"))

		plan, err := provider.PlanMarkdown(ctx, created, types.ModePlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan).To(Equal("# Parser plan\n\n1. Preserve parse context."))
		_, err = provider.PlanMarkdown(ctx, created, types.ModeRun)
		Expect(err).To(MatchError(ContainSubstring("approve an immutable revision")))
	})

	It("stores embedded verification only in the dedicated field", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Move embedded verification",
			Body: `Parser failures lose context.

# Verification

body fixture`,
			Verification: "explicit fixture",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.MarkdownBody).To(Equal("Parser failures lose context."))
		Expect(created.VerificationMarkdown).To(Equal("explicit fixture\n\nbody fixture"))

		issueID, err := uuid.Parse(created.ID)
		Expect(err).NotTo(HaveOccurred())
		stored, err := provider.repository.GetIssue(ctx, issueID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.Body).To(Equal("Parser failures lose context."))
		Expect(stored.Verification).To(Equal("explicit fixture\n\nbody fixture"))
	})

	It("routes a planned draft through review before implementation", func() {
		const planMarkdown = "# Draft plan\n\n1. Implement the parser fix."
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title:  "Plan a draft parser fix",
			Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())

		preparation, err := provider.PrepareRun(ctx, created, todos.RunPreparation{
			Mode: types.ModePlan, ExecutorName: "claude",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(preparation.SessionID).NotTo(BeEmpty())
		Expect(provider.SaveAttempt(ctx, created, &todos.ExecutionResult{
			Success: true, ExecutorName: "claude", EndStatus: types.EndCompleted,
			Plan: &types.PlanResult{Status: types.PlanNew, Content: planMarkdown},
		})).To(Succeed())
		Expect(created.Status).To(Equal(types.StatusReview))

		_, err = provider.PlanMarkdown(ctx, created, types.ModeRun)
		Expect(err).To(MatchError(ContainSubstring("approve an immutable revision")))

		created, err = provider.ApprovePlan(ctx, created, "reviewer", "ready to implement")
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Status).To(Equal(types.StatusPending))
		runnablePlan, err := provider.PlanMarkdown(ctx, created, types.ModeRun)
		Expect(err).NotTo(HaveOccurred())
		Expect(runnablePlan).To(Equal(planMarkdown))
	})

	It("creates a reviewed and approved plan ready for run", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title:    "Run the approved parser plan",
			Priority: types.PriorityMedium,
			Status:   types.StatusPending,
			Plan: &todos.CreatePlanRequest{
				Markdown: "# Approved plan\n\n1. Implement the parser fix.",
				Approved: true,
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(created.Status).To(Equal(types.StatusPending))
		Expect(created.PlanStatus).To(Equal(types.PlanNew))

		plan, err := provider.PlanMarkdown(ctx, created, types.ModeRun)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan).To(Equal("# Approved plan\n\n1. Implement the parser fix."))
	})
})
