package native_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestNativePlanCreation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Native TODO plan creation")
}

var _ = Describe("atomic issue and plan creation", func() {
	It("rolls back the issue and Captain provenance when the revision is invalid", func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres execution integration tests")
		}

		ctx := context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_native_create_plan",
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

		repository, err := native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey:  "github.com/acme/atomic-plan",
			RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
		captain, err := captaindb.Use(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())
		coordinator, err := native.NewLaunchCoordinator(captain, repository)
		Expect(err).NotTo(HaveOccurred())

		_, err = coordinator.CreateIssueWithPlan(ctx, native.CreateIssuePlanInput{
			Issue: native.CreateIssueInput{
				WorkspaceID: workspace.ID,
				Title:       "Must roll back",
				Priority:    native.PriorityMedium,
				Status:      native.StatusOpen,
				Actor:       "gavel",
			},
			Session: captaindb.CreateSessionInput{
				Source:   "gavel",
				Provider: "human",
				CWD:      workspace.RootPath,
				Title:    "Must roll back",
			},
			Plan: captaindb.CreatePlanInput{
				Title:       "Must roll back",
				Variant:     "primary",
				SpecProfile: "gavel.todo.plan",
			},
			Revision: captaindb.AppendPlanRevisionInput{CreatedBy: "human"},
			Actor:    "gavel",
		})
		Expect(err).To(HaveOccurred())

		var issueCount, sessionCount, planCount int64
		Expect(opened.Gorm().Table("todo_issues").Count(&issueCount).Error).NotTo(HaveOccurred())
		Expect(opened.Gorm().Table("captain_sessions").Count(&sessionCount).Error).NotTo(HaveOccurred())
		Expect(opened.Gorm().Table("captain_plans").Count(&planCount).Error).NotTo(HaveOccurred())
		Expect(issueCount).To(BeZero())
		Expect(sessionCount).To(BeZero())
		Expect(planCount).To(BeZero())
	})
})
