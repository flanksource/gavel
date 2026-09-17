package database

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const rejectedPlanMigration = "141_deselect_rejected_plans.sql"

var _ = Describe("rejected plan selection migration", func() {
	It("is embedded after the current schema reconciliation", func() {
		data, err := fs.ReadFile(schemaFS, "schema/"+rejectedPlanMigration)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("-- dependsOn: 140_todo_prompt_run_step_open.sql"))
		Expect(string(data)).To(ContainSubstring("approval_state = 'rejected'"))
		Expect(string(data)).To(ContainSubstring("SET selected_plan_id = NULL"))
		Expect(string(data)).To(ContainSubstring("status = 'open'"))
	})

	It("deselects only rejected plans without rewriting issue history", func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
		}

		ctx := context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_rejected_plan_migration",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(EnvDSN, dsn)
		GinkgoT().Setenv(EnvDisable, "")
		GinkgoT().Setenv(LegacyEnvDSN, "")
		GinkgoT().Setenv(LegacyEnvDisable, "")

		opened, err := Open(ctx, WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		captain, err := captaindb.Use(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())

		workspaceID := uuid.New()
		Expect(opened.Gorm().Exec(`
			INSERT INTO todo_workspaces (id, repo_key, display_name)
			VALUES (?, 'github.com/acme/rejected-plan-migration', 'migration')`, workspaceID).Error).NotTo(HaveOccurred())
		rejectedIssue, rejectedPlan := seedMigrationPlan(ctx, opened.Gorm(), captain, workspaceID, "Rejected", captaindb.PlanApprovalRejected)
		approvedIssue, approvedPlan := seedMigrationPlan(ctx, opened.Gorm(), captain, workspaceID, "Approved", captaindb.PlanApprovalApproved)
		pendingIssue, pendingPlan := seedMigrationPlan(ctx, opened.Gorm(), captain, workspaceID, "Pending", captaindb.PlanApprovalPending)

		before := migrationIssueSnapshots(opened.Gorm(), rejectedIssue, approvedIssue, pendingIssue)
		Expect(opened.Gorm().Exec(`
			DELETE FROM schema_migration_scripts
			WHERE scope = 'gavel' AND path = ?`, rejectedPlanMigration).Error).NotTo(HaveOccurred())
		upgraded, err := Open(ctx, WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		Expect(upgraded.Close()).To(Succeed())

		after := migrationIssueSnapshots(opened.Gorm(), rejectedIssue, approvedIssue, pendingIssue)
		Expect(after[rejectedIssue].SelectedPlanID).To(BeNil())
		Expect(after[rejectedIssue].Status).To(Equal("open"))
		Expect(after[approvedIssue].SelectedPlanID).To(Equal(&approvedPlan))
		Expect(after[approvedIssue].Status).To(Equal("draft"))
		Expect(after[pendingIssue].SelectedPlanID).To(Equal(&pendingPlan))
		Expect(after[pendingIssue].Status).To(Equal("draft"))
		Expect(after[rejectedIssue].Version).To(Equal(before[rejectedIssue].Version))
		Expect(after[approvedIssue].Version).To(Equal(before[approvedIssue].Version))
		Expect(after[pendingIssue].Version).To(Equal(before[pendingIssue].Version))

		var links int64
		Expect(opened.Gorm().Table("todo_issue_plans").Where("issue_id = ? AND plan_id = ?", rejectedIssue, rejectedPlan).Count(&links).Error).NotTo(HaveOccurred())
		Expect(links).To(Equal(int64(1)))
		storedRejected, err := captain.GetPlan(ctx, rejectedPlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(storedRejected.ApprovalState).To(Equal(captaindb.PlanApprovalRejected))
		Expect(storedRejected.LatestRevision).NotTo(BeNil())

		Expect(opened.Gorm().Exec(`
			DELETE FROM schema_migration_scripts
			WHERE scope = 'gavel' AND path = ?`, rejectedPlanMigration).Error).NotTo(HaveOccurred())
		reapplied, err := Open(ctx, WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		Expect(reapplied.Close()).To(Succeed())
		Expect(migrationIssueSnapshots(opened.Gorm(), rejectedIssue)[rejectedIssue]).To(Equal(after[rejectedIssue]))
	})
})

type migrationIssueSnapshot struct {
	SelectedPlanID *uuid.UUID
	Status         string
	Version        int64
}

func seedMigrationPlan(
	ctx context.Context,
	db *gorm.DB,
	captain *captaindb.DB,
	workspaceID uuid.UUID,
	title string,
	state captaindb.PlanApprovalState,
) (uuid.UUID, uuid.UUID) {
	GinkgoHelper()
	issueID, sessionID := uuid.New(), uuid.New()
	Expect(db.Exec(`INSERT INTO todo_issues (id, workspace_id, title, status) VALUES (?, ?, ?, 'draft')`, issueID, workspaceID, title).Error).NotTo(HaveOccurred())
	Expect(db.Exec(`
		INSERT INTO captain_sessions (id, source, provider, host_id)
		VALUES (?, 'gavel-test', 'test', 'local')`, sessionID).Error).NotTo(HaveOccurred())
	plan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{SourceSessionID: sessionID, Variant: "primary", Title: title})
	Expect(err).NotTo(HaveOccurred())
	revision, err := captain.AppendPlanRevision(ctx, captaindb.AppendPlanRevisionInput{
		PlanID: plan.ID, PlanMarkdown: "# " + title, CreatedBy: "migration-test",
	})
	Expect(err).NotTo(HaveOccurred())
	switch state {
	case captaindb.PlanApprovalApproved:
		_, err = captain.ApprovePlanRevision(ctx, captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: "migration-test",
		})
	case captaindb.PlanApprovalRejected:
		_, err = captain.SetPlanReviewState(ctx, captaindb.SetPlanReviewStateInput{
			PlanID: plan.ID, State: state, Actor: "migration-test",
		})
	}
	Expect(err).NotTo(HaveOccurred())
	Expect(db.Exec(`INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal) VALUES (?, ?, 0)`, issueID, plan.ID).Error).NotTo(HaveOccurred())
	Expect(db.Exec(`UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`, plan.ID, issueID).Error).NotTo(HaveOccurred())
	return issueID, plan.ID
}

func migrationIssueSnapshots(db *gorm.DB, ids ...uuid.UUID) map[uuid.UUID]migrationIssueSnapshot {
	GinkgoHelper()
	out := make(map[uuid.UUID]migrationIssueSnapshot, len(ids))
	for _, id := range ids {
		var snapshot migrationIssueSnapshot
		Expect(db.Raw(`SELECT selected_plan_id, status, version FROM todo_issues WHERE id = ?`, id).Scan(&snapshot).Error).NotTo(HaveOccurred())
		out[id] = snapshot
	}
	return out
}
