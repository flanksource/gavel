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

		mustImport := func(records ...taskhistory.Record) taskhistory.ImportResult {
			GinkgoHelper()
			result, err := store.Import(GinkgoT().Context(), records)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Rejected).To(BeEmpty())
			return result
		}

		mustImport(fresh, expired)
		fresh.Run.Name = "Updated"
		mustImport(fresh)
		Expect(store.Prune(GinkgoT().Context(), now)).To(Succeed())
		runs, err := store.ListRuns(GinkgoT().Context(), now)

		Expect(err).NotTo(HaveOccurred())
		Expect(runs).To(Equal([]taskhistory.RunSummary{{Run: fresh.Run, ArchivedAt: fresh.ArchivedAt}}))
		snapshots, err := store.Snapshot(GinkgoT().Context(), fresh.Run.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshots).To(Equal(fresh.Snapshots))
		missing, err := store.Snapshot(GinkgoT().Context(), expired.Run.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(missing).To(BeNil(), "a pruned run has no archived snapshot")

		// Archived runs are immutable, and the periodic spool sweep re-offers
		// every one of them. Re-importing identical content must not touch the
		// row: an unchanged upsert that still writes turns a 40-row table into
		// millions of dead tuples.
		rowVersion := func() string {
			var xmin string
			Expect(db.Gorm().Raw(
				"SELECT xmin::text FROM task_run_history WHERE id = ?", fresh.Run.ID,
			).Scan(&xmin).Error).To(Succeed())
			return xmin
		}
		before := rowVersion()
		mustImport(fresh)
		Expect(rowVersion()).To(Equal(before), "re-importing an unchanged record must not rewrite the row")

		fresh.ArchivedAt = now.Add(-30 * time.Minute)
		mustImport(fresh)
		Expect(rowVersion()).NotTo(Equal(before), "a genuinely changed record must still be written")

		// Snapshots carry raw subprocess stdout. A NUL byte in it is legal JSON
		// but unrepresentable in jsonb, and used to fail the whole batch with
		// SQLSTATE 22P05.
		record := func(id, stdout string) taskhistory.Record {
			return taskhistory.Record{
				Run: clickytask.RunMeta{ID: id, Name: id, StartedAt: now.Add(-time.Hour).Format(time.RFC3339Nano)},
				Snapshots: []clickytask.TaskSnapshot{{
					ID: id, GroupID: id, Type: "group",
					Status: string(clickytask.StatusSuccess), Stdout: stdout,
				}},
				ArchivedAt: now.Add(-time.Hour),
			}
		}
		noisy := record("noisy-db-run", "before\x00after")
		Expect(mustImport(noisy, record("sibling-db-run", "clean")).Imported).To(Equal(2))

		stored, err := store.Snapshot(GinkgoT().Context(), noisy.Run.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(HaveLen(1))
		Expect(stored[0].Stdout).To(Equal("beforeafter"), "the NUL is dropped and the rest of the output kept")

		// A record the database will never accept is skipped, not allowed to
		// abort the batch — that is what used to stall every later sweep.
		malformed := record("malformed-db-run", "")
		malformed.Snapshots = nil
		result, err := store.Import(GinkgoT().Context(), []taskhistory.Record{malformed, record("after-reject-db-run", "kept")})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Imported).To(Equal(1))
		Expect(result.Rejected).To(HaveLen(1))
		Expect(result.Rejected[0].RunID).To(Equal("malformed-db-run"))
		survivor, err := store.Snapshot(GinkgoT().Context(), "after-reject-db-run")
		Expect(err).NotTo(HaveOccurred())
		Expect(survivor).NotTo(BeEmpty(), "a record queued behind a rejected one must still land")
	})
})
