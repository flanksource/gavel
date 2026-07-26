package taskhistory_test

import (
	"os"
	"path/filepath"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/taskhistory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task history database store", func() {
	It("imports spool records idempotently and prunes expired runs", func() {
		dsn := os.Getenv("GAVEL_TASK_HISTORY_TEST_DSN")
		if dsn == "" && os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_TASK_HISTORY_TEST_DSN or GAVEL_DB_EMBEDDED_TEST=1 to run task history database tests")
		}
		if dsn == "" {
			var stop func() error
			var err error
			dsn, stop, err = commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
				DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
				Database: "gavel_task_history",
			})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() { Expect(stop()).To(Succeed()) })
		}
		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		db, err := database.Open(GinkgoT().Context(), database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(db.Close()).To(Succeed()) })
		store, err := taskhistory.NewStore(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		now := time.Date(2026, time.July, 21, 8, 0, 0, 0, time.UTC)
		fresh := taskhistory.Record{
			Run:        clickytask.RunMeta{ID: "fresh-db-run", Name: "Fresh", StartedAt: now.Add(-time.Hour).Format(time.RFC3339Nano)},
			Snapshots:  []clickytask.TaskSnapshot{{ID: "fresh-db-run", GroupID: "fresh-db-run", Type: "group", Status: string(clickytask.StatusSuccess)}},
			ArchivedAt: now.Add(-time.Hour),
		}
		expired := taskhistory.Record{
			Run:        clickytask.RunMeta{ID: "expired-db-run", Name: "Expired", StartedAt: now.Add(-31 * 24 * time.Hour).Format(time.RFC3339Nano)},
			Snapshots:  []clickytask.TaskSnapshot{{ID: "expired-db-run", GroupID: "expired-db-run", Type: "group", Status: string(clickytask.StatusSuccess)}},
			ArchivedAt: now.Add(-31 * 24 * time.Hour),
		}

		Expect(store.Import(GinkgoT().Context(), []taskhistory.Record{fresh, expired})).To(Succeed())
		fresh.Run.Name = "Updated"
		Expect(store.Import(GinkgoT().Context(), []taskhistory.Record{fresh})).To(Succeed())
		Expect(store.Prune(GinkgoT().Context(), now)).To(Succeed())
		records, err := store.List(GinkgoT().Context(), now)

		Expect(err).NotTo(HaveOccurred())
		Expect(records).To(Equal([]taskhistory.Record{fresh}))
	})
})
