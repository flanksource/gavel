package runtime

import (
	"context"
	"os"
	"path/filepath"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("setting a TODO plan from human-authored markdown", Ordered, func() {
	var (
		ctx      context.Context
		provider *Provider
	)

	issueOf := func(todo *types.TODO) *native.Issue {
		GinkgoHelper()
		id, err := uuid.Parse(todo.ID)
		Expect(err).NotTo(HaveOccurred())
		issue, err := provider.repository.GetIssue(ctx, id)
		Expect(err).NotTo(HaveOccurred())
		return issue
	}

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}

		ctx = context.Background()
		dataDir := filepath.Join(GinkgoT().TempDir(), "postgres")
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  dataDir,
			Database: "gavel_todo_plan_revision",
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
			RepoKey:  "github.com/acme/plan-revision",
			RootPath: workDir,
		})
		Expect(err).NotTo(HaveOccurred())
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "plan-revision", RootPath: workDir, Repositories: []string{"acme/plan-revision"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates and selects the first plan for a TODO that never had one", func() {
		const markdown = "# Inline hint styles\n\n1. Render hint chips in the sidebar row."
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Show hint styles inline", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(issueOf(created).SelectedPlanID).To(BeNil())

		saved, err := provider.SavePlanRevision(ctx, created, markdown, "moshe")
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Status).To(Equal(types.StatusReview))

		issue := issueOf(saved)
		Expect(issue.SelectedPlanID).NotTo(BeNil())
		plan, err := provider.captain.GetPlan(ctx, *issue.SelectedPlanID)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.ApprovalState).To(Equal(captaindb.PlanApprovalPending))
		Expect(plan.LatestRevision).NotTo(BeNil())
		Expect(plan.LatestRevision.Revision).To(Equal(1))
		Expect(plan.LatestRevision.PlanMarkdown).To(Equal(markdown))

		links, err := provider.repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(1))
		Expect(links[0].Ordinal).To(Equal(0))

		// The plan hangs off the same deterministic session `todos create --plan`
		// would have used, so repeated edits never fan out session rows.
		session, err := provider.captain.GetSession(ctx, suppliedPlanSessionID(issue.ID))
		Expect(err).NotTo(HaveOccurred())
		Expect(session.ID).To(Equal(plan.SourceSessionID))
		Expect(session.ParentSessionID).To(Equal(&issue.ID))

		content, err := provider.PlanMarkdown(ctx, saved, types.ModePlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal(markdown))
	})

	It("replaces the content of a plan it already selected", func() {
		const first = "# First\n\n1. Start here."
		const second = "# Second\n\n1. Actually start here."
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Replace an existing plan", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())

		saved, err := provider.SavePlanRevision(ctx, created, first, "moshe")
		Expect(err).NotTo(HaveOccurred())
		originalPlanID := *issueOf(saved).SelectedPlanID

		saved, err = provider.SavePlanRevision(ctx, saved, second, "moshe")
		Expect(err).NotTo(HaveOccurred())

		issue := issueOf(saved)
		Expect(*issue.SelectedPlanID).To(Equal(originalPlanID), "replacing content must reuse the plan, not fork one")
		plan, err := provider.captain.GetPlan(ctx, originalPlanID)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.LatestRevision.Revision).To(Equal(2))

		content, err := provider.PlanMarkdown(ctx, saved, types.ModePlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal(second))

		links, err := provider.repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(1))
	})

	It("returns an approved plan to review when its content is replaced", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Re-review a replaced plan", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())
		saved, err := provider.SavePlanRevision(ctx, created, "# Approved\n\n1. Ship it.", "moshe")
		Expect(err).NotTo(HaveOccurred())
		saved, err = provider.ApprovePlan(ctx, saved, "moshe", "looks right")
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Status).To(Equal(types.StatusPending))

		saved, err = provider.SavePlanRevision(ctx, saved, "# Approved\n\n1. Ship something else.", "moshe")
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Status).To(Equal(types.StatusReview))
		_, err = provider.PlanMarkdown(ctx, saved, types.ModeRun)
		Expect(err).To(MatchError(ContainSubstring("approve an immutable revision")))
	})

	It("takes the next free ordinal when a link outlived its selection", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Recover from an orphaned plan link", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())
		saved, err := provider.SavePlanRevision(ctx, created, "# Orphaned\n\n1. Superseded.", "moshe")
		Expect(err).NotTo(HaveOccurred())

		issue := issueOf(saved)
		cleared, err := provider.repository.SelectPlan(ctx, issue.ID, nil, issue.Version, "test")
		Expect(err).NotTo(HaveOccurred())
		Expect(cleared.SelectedPlanID).To(BeNil())
		Expect(provider.reloadTODO(ctx, saved, saved.CWD)).To(Succeed())

		saved, err = provider.SavePlanRevision(ctx, saved, "# Replacement\n\n1. Start over.", "moshe")
		Expect(err).NotTo(HaveOccurred())

		links, err := provider.repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(2))
		Expect(links[1].Ordinal).To(Equal(1))
		content, err := provider.PlanMarkdown(ctx, saved, types.ModePlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(content).To(Equal("# Replacement\n\n1. Start over."))
	})

	It("still refuses to approve a TODO that has no plan, and names the remedy", func() {
		created, err := provider.Create(ctx, todos.CreateRequest{
			Title: "Approve nothing", Status: types.StatusDraft,
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = provider.ApprovePlan(ctx, created, "moshe", "")
		Expect(err).To(MatchError(ContainSubstring("has no selected Captain plan")))
		Expect(err).To(MatchError(ContainSubstring("gavel todos edit --plan")))
	})
})
