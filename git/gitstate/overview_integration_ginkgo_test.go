package gitstate_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/git/gitstate"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// overviewRow is one git_worktree_overview row as the comparisons read it.
type overviewRow struct {
	Path              string
	StatusFingerprint *string
}

// overviewRows reads the overview and, independently, what rebuilding it from
// git_worktrees would hold; the two must always agree on the fingerprint.
func (r *trackedRepo) overviewRows() (actual, rebuilt []overviewRow) {
	GinkgoHelper()
	Expect(r.db.Raw(`SELECT path, status_fingerprint FROM git_worktree_overview ORDER BY path`).Scan(&actual).Error).To(Succeed())
	Expect(r.db.Raw(`
		SELECT path, CASE WHEN status_error IS NULL THEN status_fingerprint END AS status_fingerprint
		FROM git_worktrees ORDER BY path`).Scan(&rebuilt).Error).To(Succeed())
	return actual, rebuilt
}

func (r *trackedRepo) loggedVersion(path string) string {
	GinkgoHelper()
	var version string
	Expect(r.db.Raw(`SELECT xmin::text FROM git_worktrees WHERE path = ?`, path).Scan(&version).Error).To(Succeed())
	return version
}

func (r *trackedRepo) overviewScannedAt(path string) time.Time {
	GinkgoHelper()
	var scanned time.Time
	Expect(r.db.Raw(`SELECT status_scanned_at FROM git_worktree_overview WHERE path = ?`, path).Scan(&scanned).Error).To(Succeed())
	return scanned
}

var _ = Describe("Worktree overview", func() {
	It("confirms an unchanged status scan on the overview without rewriting the logged worktree row", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		version := r.loggedVersion(r.feature)
		scanned := r.overviewScannedAt(r.feature)

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(r.loggedVersion(r.feature)).To(Equal(version), "an unchanged scan leaves git_worktrees alone")
		Expect(r.overviewScannedAt(r.feature)).To(BeTemporally(">", scanned))
		state, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		Expect(*state.Worktrees[1].StatusScannedAt).To(BeTemporally("~", r.overviewScannedAt(r.feature), time.Millisecond))
	})

	It("keeps the overview equal to its rebuild through a status change, a failed scan and its recovery", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		actual, rebuilt := r.overviewRows()
		Expect(actual).To(HaveLen(2))
		Expect(actual).To(Equal(rebuilt), "after the first scan")

		write(r.feature, "b.txt", "b\n")
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		actual, rebuilt = r.overviewRows()
		Expect(actual).To(Equal(rebuilt), "after a status change")

		var repoID string
		Expect(r.db.Raw(`SELECT id::text FROM git_repos WHERE root_dir = ?`, r.root).Scan(&repoID).Error).To(Succeed())
		store := r.tracker.Store()
		Expect(store.RecordStatusError(ctx, uuid.MustParse(repoID), r.feature, errors.New("status failed"))).To(Succeed())
		actual, rebuilt = r.overviewRows()
		Expect(actual).To(Equal(rebuilt), "after a failed scan")
		Expect(actual[1].StatusFingerprint).To(BeNil(), "a failed scan forces the next one to be full")

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		actual, rebuilt = r.overviewRows()
		Expect(actual).To(Equal(rebuilt), "after the recovering scan")
		Expect(actual[1].StatusFingerprint).NotTo(BeNil())
	})

	It("rebuilds the overview when crash recovery leaves it empty", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		before, _ := r.overviewRows()
		store := r.tracker.Store()
		Expect(r.db.Exec(`TRUNCATE git_worktree_overview`).Error).To(Succeed())

		rebuilt, err := store.EnsureOverview(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(rebuilt).To(BeTrue())
		after, _ := r.overviewRows()
		Expect(after).To(Equal(before))
		rebuilt, err = store.EnsureOverview(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(rebuilt).To(BeFalse(), "a populated overview is left as is")
	})

	It("refills a missing overview row on the worktree's next scan", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		Expect(r.db.Exec(`TRUNCATE git_worktree_overview`).Error).To(Succeed())

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		actual, rebuilt := r.overviewRows()
		Expect(actual).To(HaveLen(2))
		Expect(actual).To(Equal(rebuilt))
	})
})

func (r *trackedRepo) repoID() uuid.UUID {
	GinkgoHelper()
	var id string
	Expect(r.db.Raw(`SELECT id::text FROM git_repos WHERE root_dir = ?`, r.root).Scan(&id).Error).To(Succeed())
	return uuid.MustParse(id)
}

func (r *trackedRepo) loggedRepoVersion() string {
	GinkgoHelper()
	var version string
	Expect(r.db.Raw(`SELECT xmin::text FROM git_repos WHERE root_dir = ?`, r.root).Scan(&version).Error).To(Succeed())
	return version
}

// listen subscribes to gitstate.ChangeChannel on a dedicated connection;
// drain returns the payloads delivered since the last drain.
func (r *trackedRepo) listen(ctx context.Context) (drain func() []string) {
	GinkgoHelper()
	conn, err := pgx.Connect(ctx, r.dsn)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(conn.Close(context.Background())).To(Succeed()) })
	_, err = conn.Exec(ctx, "LISTEN "+gitstate.ChangeChannel)
	Expect(err).NotTo(HaveOccurred())
	return func() []string {
		GinkgoHelper()
		var payloads []string
		for {
			waitCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			notification, err := conn.WaitForNotification(waitCtx)
			cancel()
			if errors.Is(err, context.DeadlineExceeded) {
				return payloads
			}
			Expect(err).NotTo(HaveOccurred())
			payloads = append(payloads, notification.Payload)
		}
	}
}

var _ = Describe("Repo overview", func() {
	It("confirms unchanged ref and status scans without a logged write or an announcement", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		drain := r.listen(ctx)
		repoVersion, generation := r.loggedRepoVersion(), r.generation()

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(r.loggedRepoVersion()).To(Equal(repoVersion), "an unchanged ref scan leaves git_repos alone")
		Expect(r.generation()).To(Equal(generation))
		Expect(drain()).To(BeEmpty())
	})

	It("announces one real change exactly once and advances the generation", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		drain := r.listen(ctx)
		generation := r.generation()

		write(r.feature, "c.txt", "c\n")
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(drain()).To(Equal([]string{r.repoID().String()}))
		Expect(r.generation()).To(BeNumerically(">", generation))
	})

	It("rebuilds both overviews after crash recovery with a generation that never goes backwards", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		generation := r.generation()
		drain := r.listen(ctx)
		Expect(r.db.Exec(`TRUNCATE git_worktree_overview, git_repo_overview`).Error).To(Succeed())

		rebuilt, err := r.tracker.Store().EnsureOverview(ctx)

		Expect(err).NotTo(HaveOccurred())
		Expect(rebuilt).To(BeTrue())
		Expect(r.generation()).To(BeNumerically(">", generation))
		Expect(drain()).To(Equal([]string{r.repoID().String()}), "listeners reload every repo")
		actual, expected := r.overviewRows()
		Expect(actual).To(Equal(expected))
	})

	It("recreates a missing repo overview row on the repo's next scan and announces it", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		drain := r.listen(ctx)
		Expect(r.db.Exec(`TRUNCATE git_repo_overview`).Error).To(Succeed())

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(r.count(`SELECT count(*) FROM git_repo_overview`)).To(Equal(int64(1)))
		Expect(drain()).To(ContainElement(r.repoID().String()))
		repoVersion := r.loggedRepoVersion()
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		Expect(r.loggedRepoVersion()).To(Equal(repoVersion), "the recreated row confirms the next unchanged scan")
	})
})

var _ = Describe("Hot worktrees", func() {
	It("heats only the deepest worktree containing an agent's directory", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		nested := filepath.Join(r.root, ".claude", "worktrees", "nested")
		git(r.root, "worktree", "add", "-q", "-b", "nested", nested, r.base)
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		sub := filepath.Join(nested, "pkg")
		Expect(os.MkdirAll(sub, 0o755)).To(Succeed())

		r.tracker.SetActive([]string{sub, r.feature})

		Expect(r.tracker.HotWorktrees()).To(Equal([]string{nested, r.feature}))
	})
})
