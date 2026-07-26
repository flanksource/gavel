package runtime

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("configured project workspace initialization", Ordered, func() {
	var (
		ctx context.Context
		db  *gorm.DB
	)

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}

		ctx = context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
			Database: "gavel_todo_workspace_init",
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
		db = opened.Gorm()
	})

	It("creates and reuses the native workspace for a configured project", func() {
		root := filepath.Join(GinkgoT().TempDir(), "config-db")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		options := WorkspaceOptions{
			Name:         "config-db",
			RootPath:     root,
			Repositories: []string{"flanksource/config-db"},
		}

		first, err := New(ctx, db, options)
		Expect(err).NotTo(HaveOccurred())
		second, err := New(ctx, db, options)
		Expect(err).NotTo(HaveOccurred())

		Expect(second.Workspace().ID).To(Equal(first.Workspace().ID))
		Expect(second.Workspace().RepoKey).To(Equal("github.com/flanksource/config-db"))
		Expect(second.Workspace().RootPath).To(Equal(root))
		Expect(second.Workspace().DisplayName).To(Equal("config-db"))
	})

	It("reconciles a repository-matched workspace to the configured project", func() {
		repository, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		oldRoot := filepath.Join(GinkgoT().TempDir(), "old")
		newRoot := filepath.Join(GinkgoT().TempDir(), "new")
		Expect(os.MkdirAll(oldRoot, 0o755)).To(Succeed())
		Expect(os.MkdirAll(newRoot, 0o755)).To(Succeed())
		existing, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey:     "github.com/acme/reconciled",
			RootPath:    oldRoot,
			DisplayName: "Old name",
		})
		Expect(err).NotTo(HaveOccurred())

		provider, err := New(ctx, db, WorkspaceOptions{
			Name: "Reconciled", RootPath: newRoot, Repositories: []string{"acme/reconciled"},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(provider.Workspace().ID).To(Equal(existing.ID))
		Expect(provider.Workspace().RootPath).To(Equal(newRoot))
		Expect(provider.Workspace().DisplayName).To(Equal("Reconciled"))
	})

	It("converges concurrent initialization on one workspace", func() {
		root := filepath.Join(GinkgoT().TempDir(), "concurrent")
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		options := WorkspaceOptions{
			Name: "concurrent", RootPath: root, Repositories: []string{"acme/concurrent"},
		}
		type result struct {
			provider *Provider
			err      error
		}
		results := make(chan result, 8)
		var group sync.WaitGroup
		for range 8 {
			group.Add(1)
			go func() {
				defer group.Done()
				provider, err := New(ctx, db, options)
				results <- result{provider: provider, err: err}
			}()
		}
		group.Wait()
		close(results)

		var workspaceID string
		for result := range results {
			Expect(result.err).NotTo(HaveOccurred())
			Expect(result.provider).NotTo(BeNil())
			if workspaceID == "" {
				workspaceID = result.provider.Workspace().ID.String()
			}
			Expect(result.provider.Workspace().ID.String()).To(Equal(workspaceID))
		}

		var count int64
		Expect(db.Table("todo_workspaces").Where("repo_key = ?", "github.com/acme/concurrent").Count(&count).Error).To(Succeed())
		Expect(count).To(Equal(int64(1)))
	})
})
