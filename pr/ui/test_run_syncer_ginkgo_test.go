package ui

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("TestRunSyncer", Ordered, func() {
	var (
		ctx   = context.Background()
		store *cache.Store
		db    *gorm.DB
	)

	BeforeAll(func() {
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_prui_test_run_syncer"}).DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db = opened.Gorm()
		store, err = cache.Open()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(store.Close()).To(Succeed()) })
		GinkgoT().Setenv(database.EnvDSN, "")
	})

	workspace := func() string {
		dir := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(dir, snapshots.Dir), 0o755)).To(Succeed())
		return dir
	}
	saveRun := func(dir string, started time.Time) {
		snap := testui.Snapshot{
			Metadata: &testui.SnapshotMetadata{Started: started},
			Tests:    []parsers.Test{{Name: "parses", Passed: true}},
		}
		_, err := snapshots.SavePerRun(dir, &snap, started, "")
		Expect(err).NotTo(HaveOccurred())
	}
	cursorSyncedAt := func(dir string) []time.Time {
		var stamps []time.Time
		Expect(db.Raw(`SELECT synced_at FROM test_run_cursors WHERE workspace_dir = ?`, dir).Scan(&stamps).Error).To(Succeed())
		return stamps
	}
	cachedRuns := func(dir string) int64 {
		var n int64
		Expect(db.Raw(`SELECT count(*) FROM test_run_caches WHERE workspace_dir = ?`, dir).Scan(&n).Error).To(Succeed())
		return n
	}

	It("writes no cursor for a workspace whose ticks find nothing", func() {
		dir := workspace()
		syncer := NewTestRunSyncer(nil)

		Expect(syncer.syncWorkspace(ctx, store, dir)).To(BeFalse())
		Expect(syncer.syncWorkspace(ctx, store, dir)).To(BeFalse())

		Expect(cursorSyncedAt(dir)).To(BeEmpty())
	})

	It("leaves the cursor alone on an idle tick and advances it with the next new run", func() {
		dir := workspace()
		first := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
		saveRun(dir, first)
		syncer := NewTestRunSyncer(nil)

		Expect(syncer.syncWorkspace(ctx, store, dir)).To(BeTrue())
		stamped := cursorSyncedAt(dir)
		Expect(stamped).To(HaveLen(1))
		Expect(cachedRuns(dir)).To(Equal(int64(1)))

		Expect(syncer.syncWorkspace(ctx, store, dir)).To(BeFalse(), "an idle tick")
		Expect(cursorSyncedAt(dir)).To(Equal(stamped), "an idle tick writes no cursor")

		saveRun(dir, first.Add(time.Hour))
		Expect(syncer.syncWorkspace(ctx, store, dir)).To(BeTrue())
		Expect(cachedRuns(dir)).To(Equal(int64(2)))
		Expect(cursorSyncedAt(dir)[0]).To(BeTemporally(">", stamped[0]))
	})

	It("resumes from the stored cursor in a new syncer", func() {
		dir := workspace()
		started := time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)
		saveRun(dir, started)
		Expect(NewTestRunSyncer(nil).syncWorkspace(ctx, store, dir)).To(BeTrue())
		stamped := cursorSyncedAt(dir)

		Expect(NewTestRunSyncer(nil).syncWorkspace(ctx, store, dir)).To(BeFalse(), "the stored watermark already covers the run")
		Expect(cursorSyncedAt(dir)).To(Equal(stamped))
	})
})
