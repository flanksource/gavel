package native_test

import (
	"context"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("durable PR links", func() {
	It("persists explicit links, preserves unlinked records, and supports unlinking", func() {
		ctx := context.Background()
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_pr_links"}).DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db := opened.Gorm()
		repo, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/acme/widgets"})
		Expect(err).NotTo(HaveOccurred())
		legacy, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Address feedback on acme/widgets#12", Body: "https://github.com/acme/widgets/pull/12"})
		Expect(err).NotTo(HaveOccurred())
		link := &types.PR{Number: 12, URL: "https://github.com/acme/widgets/pull/12"}
		linked, err := repo.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Fix parser validation", PR: link})
		Expect(err).NotTo(HaveOccurred())
		Expect(linked.PR).To(Equal(link))
		reloaded, err := repo.GetIssue(ctx, legacy.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.PR).To(BeNil())
		var unlinked int64
		Expect(db.Table("todo_issues").Where("id = ? AND pr IS NULL", legacy.ID).Count(&unlinked).Error).To(Succeed())
		Expect(unlinked).To(Equal(int64(1)))
		reloaded, err = repo.UpdateIssue(ctx, legacy.ID, legacy.Version, native.IssuePatch{PR: &link})
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.PR).To(Equal(link))
		var removed *types.PR
		reloaded, err = repo.UpdateIssue(ctx, legacy.ID, reloaded.Version, native.IssuePatch{PR: &removed})
		Expect(err).NotTo(HaveOccurred())
		Expect(reloaded.PR).To(BeNil())
	})
})
