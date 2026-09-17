package native_test

import (
	"context"
	"os"
	"path/filepath"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("atomic plan review", func() {
	It("rejects and deselects one linked plan as an idempotent transaction", func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres execution integration tests")
		}

		ctx := context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_native_plan_review",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")

		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

		repository, err := native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey: "github.com/acme/plan-review", RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
		captain, err := captaindb.Use(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		coordinator, err := native.NewLaunchCoordinator(captain, repository)
		Expect(err).NotTo(HaveOccurred())

		issue, plan, revision := seedSelectedReviewPlan(ctx, repository, captain, workspace.ID, "Reject atomically")
		originalVersion := issue.Version
		review := captaindb.SetPlanReviewStateInput{
			PlanID: plan.ID, State: captaindb.PlanApprovalRejected,
			Actor: "reviewer", Comment: "superseded",
		}
		attachment := native.PlanSelectionAttachment{
			IssueID: issue.ID, Ordinal: 0, ExpectedIssueVersion: originalVersion, Actor: "reviewer",
		}

		rejected, err := coordinator.ReviewPlan(ctx, review, attachment)
		Expect(err).NotTo(HaveOccurred())
		Expect(rejected.Plan.ApprovalState).To(Equal(captaindb.PlanApprovalRejected))
		Expect(rejected.Plan.ApprovedRevisionID).To(BeNil())
		Expect(rejected.Issue.SelectedPlanID).To(BeNil())
		Expect(rejected.Issue.Status).To(Equal(native.StatusOpen))
		Expect(rejected.Issue.Version).To(Equal(originalVersion + 1))

		links, err := repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(1))
		Expect(links[0].IssueID).To(Equal(issue.ID))
		Expect(links[0].PlanID).To(Equal(plan.ID))
		Expect(links[0].Ordinal).To(Equal(0))
		revisions, err := captain.ListPlanRevisions(ctx, plan.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(revisions).To(HaveLen(1))
		Expect(revisions[0].ID).To(Equal(revision.ID))

		events, err := repository.ListEvents(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(events[len(events)-1].Kind).To(Equal("plan_rejected"))
		eventCount := len(events)

		replayed, err := coordinator.ReviewPlan(ctx, review, attachment)
		Expect(err).NotTo(HaveOccurred())
		Expect(replayed.Issue.Version).To(Equal(rejected.Issue.Version))
		events, err = repository.ListEvents(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).To(HaveLen(eventCount))
	})

	It("rolls back the Captain decision when the issue version is stale", func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres execution integration tests")
		}

		ctx := context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_native_plan_review_conflict",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")

		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		repository, err := native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey: "github.com/acme/plan-review-conflict", RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
		captain, err := captaindb.Use(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		coordinator, err := native.NewLaunchCoordinator(captain, repository)
		Expect(err).NotTo(HaveOccurred())

		issue, plan, _ := seedSelectedReviewPlan(ctx, repository, captain, workspace.ID, "Reject with stale version")
		_, err = coordinator.ReviewPlan(ctx, captaindb.SetPlanReviewStateInput{
			PlanID: plan.ID, State: captaindb.PlanApprovalRejected, Actor: "reviewer",
		}, native.PlanSelectionAttachment{
			IssueID: issue.ID, Ordinal: 0, ExpectedIssueVersion: issue.Version - 1, Actor: "reviewer",
		})
		Expect(err).To(MatchError(ContainSubstring(native.ErrVersionConflict.Error())))

		storedPlan, err := captain.GetPlan(ctx, plan.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(storedPlan.ApprovalState).To(Equal(captaindb.PlanApprovalPending))
		storedIssue, err := repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(storedIssue.SelectedPlanID).To(Equal(&plan.ID))
	})
})

func seedSelectedReviewPlan(
	ctx context.Context,
	repository *native.Repository,
	captain *captaindb.DB,
	workspaceID uuid.UUID,
	title string,
) (*native.Issue, *captaindb.Plan, *captaindb.PlanRevision) {
	GinkgoHelper()
	issue, err := repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspaceID, Title: title})
	Expect(err).NotTo(HaveOccurred())
	session, err := captain.CreateOrGetSession(ctx, captaindb.CreateSessionInput{
		ID: uuid.New(), Source: "gavel-test", Provider: "test", CWD: "/workspace",
	})
	Expect(err).NotTo(HaveOccurred())
	plan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: session.ID, Variant: "primary", Title: title,
	})
	Expect(err).NotTo(HaveOccurred())
	revision, err := captain.AppendPlanRevision(ctx, captaindb.AppendPlanRevisionInput{
		PlanID: plan.ID, PlanMarkdown: "# Plan\n\n1. Replace it.", CreatedBy: "planner",
	})
	Expect(err).NotTo(HaveOccurred())
	integration, err := native.NewExecutionIntegration(repository)
	Expect(err).NotTo(HaveOccurred())
	issue, err = integration.SelectPlan(ctx, native.PlanAttachment{
		IssueID: issue.ID, PlanID: plan.ID, Ordinal: 0,
		ExpectedIssueVersion: issue.Version, Actor: "planner",
	})
	Expect(err).NotTo(HaveOccurred())
	return issue, plan, revision
}
