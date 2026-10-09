package gitstate_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// trackedRepo is a main checkout with a linked worktree on an unmerged
// branch, tracked against a migrated database.
type trackedRepo struct {
	root, feature string
	base          string
	db            *gorm.DB
	tracker       *gitstate.Tracker
}

func git(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func write(dir, file, content string) {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644)).To(Succeed())
}

func commit(dir, file, content, message string) string {
	GinkgoHelper()
	write(dir, file, content)
	git(dir, "add", file)
	git(dir, "commit", "-q", "-m", message)
	return git(dir, "rev-parse", "HEAD")
}

func newTrackedRepo(ctx context.Context, opts gitstate.Options) *trackedRepo {
	GinkgoHelper()
	handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_gitstate"})
	GinkgoT().Setenv(database.EnvDSN, handle.DSN())
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	opened, err := database.Open(ctx, database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	tmp, err := filepath.EvalSymlinks(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	r := &trackedRepo{root: filepath.Join(tmp, "repo"), feature: filepath.Join(tmp, "wt-feature"), db: opened.Gorm()}
	Expect(os.MkdirAll(r.root, 0o755)).To(Succeed())
	git(r.root, "init", "-q", "-b", "main")
	git(r.root, "config", "user.email", "test@example.com")
	git(r.root, "config", "user.name", "Test User")
	// No origin: the PR base resolves to the default, main.
	r.base = commit(r.root, "README.md", "one\ntwo\n", "chore: initial")
	git(r.root, "worktree", "add", "-q", "-b", "feature", r.feature, r.base)
	commit(r.feature, "a.txt", "a1\na2\n", "feat: a")

	runCtx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	opts.Store, opts.Context = gitstate.NewStore(r.db), runCtx
	r.tracker = gitstate.NewTracker(opts)
	return r
}

func (r *trackedRepo) count(query string, args ...any) int64 {
	GinkgoHelper()
	var n int64
	Expect(r.db.Raw(query, args...).Scan(&n).Error).To(Succeed())
	return n
}

func (r *trackedRepo) generation() int64 {
	GinkgoHelper()
	return r.count(`SELECT generation FROM git_repos WHERE root_dir = ?`, r.root)
}

var _ = Describe("Tracker", func() {
	It("scans a repository on first read: base, worktrees with uncommitted changes, and unmerged branches", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		write(r.feature, "a.txt", "a1\na2\na3\n")

		state, err := r.tracker.State(ctx, r.root)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.Base).To(Equal("main"))
		Expect(state.CurrentBranch).To(Equal("main"))
		Expect(state.BaseCheckedOut).To(BeTrue())
		Expect(state.ComputedAt).To(BeTemporally("~", time.Now(), time.Minute))
		Expect(state.Worktrees).To(HaveLen(2))
		Expect(state.Worktrees[0].Path).To(Equal(r.root))
		Expect(state.Worktrees[0].Primary).To(BeTrue())
		Expect(state.Worktrees[1].Path).To(Equal(r.feature))
		Expect(state.Worktrees[1].Ahead).To(Equal(1))
		Expect(state.Worktrees[1].Changes).To(Equal(gitstate.Changes{Unstaged: 1, Adds: 1}))
		Expect(state.Branches).To(HaveLen(1))
		Expect(state.Branches[0].Name).To(Equal("feature"))
		Expect(state.Branches[0].Worktree).To(Equal(r.feature))
		Expect(state.Branches[0].Diff).To(Equal(gavelgit.DiffStat{Commits: 1, Files: 1, Adds: 2}))
	})

	It("computes no comparison and changes nothing when refs and worktrees are unchanged", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		generation := r.generation()
		ranges := r.count(`SELECT count(*) FROM git_range_stats`)

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(r.generation()).To(Equal(generation))
		Expect(r.count(`SELECT count(*) FROM git_range_stats`)).To(Equal(ranges))
	})

	It("computes exactly one new comparison when a branch head moves", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		ranges := r.count(`SELECT count(*) FROM git_range_stats`)

		head := commit(r.feature, "b.txt", "b1\n", "feat: b")
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())

		Expect(r.count(`SELECT count(*) FROM git_range_stats`)).To(Equal(ranges + 1))
		state, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Branches[0].Head).To(Equal(head))
		Expect(state.Branches[0].Diff).To(Equal(gavelgit.DiffStat{Commits: 2, Files: 2, Adds: 3}))
	})

	It("replaces a worktree's recorded files when its status changes, and only then", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		files, err := r.tracker.Files(ctx, r.root, r.feature)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeEmpty())

		write(r.feature, "new.txt", "n1\nn2\n")
		write(r.feature, "a.txt", "a1\n")
		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		generation := r.generation()

		files, err = r.tracker.Files(ctx, r.root, r.feature)
		Expect(err).NotTo(HaveOccurred())
		paths := map[string]string{}
		for _, file := range files {
			paths[file.Path] = string(file.State)
		}
		Expect(paths).To(Equal(map[string]string{"a.txt": "unstaged", "new.txt": "untracked"}))

		Expect(r.tracker.Rescan(ctx, r.root)).To(Succeed())
		Expect(r.generation()).To(Equal(generation), "an unchanged status writes nothing a reader sees")
	})

	It("rescans the refs of a repository on its own when a commit lands under .git", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{Debounce: 20 * time.Millisecond})
		Expect(r.tracker.Start()).To(Succeed())
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())

		head := commit(r.feature, "c.txt", "c1\n", "feat: c")

		Eventually(func(g Gomega) {
			state, err := r.tracker.State(ctx, r.root)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(state.Branches).To(HaveLen(1))
			g.Expect(state.Branches[0].Head).To(Equal(head))
		}).WithTimeout(10 * time.Second).WithPolling(100 * time.Millisecond).Should(Succeed())
	})

	It("rescans a worktree on its own while it is focused", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{Hot: 200 * time.Millisecond})
		Expect(r.tracker.Start()).To(Succeed())
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())

		r.tracker.Focus(r.feature, time.Minute)
		write(r.feature, "focus.txt", "f\n")

		Eventually(func(g Gomega) {
			state, err := r.tracker.State(ctx, r.root)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(state.Worktrees[1].Changes.Untracked).To(Equal(1))
		}).WithTimeout(10 * time.Second).WithPolling(100 * time.Millisecond).Should(Succeed())
	})

	It("dates the state by its ref scan and each worktree by its own status scan", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		staleStatus := time.Now().Add(-time.Hour)
		Expect(r.db.Exec(`UPDATE git_worktrees SET status_scanned_at = ? WHERE path = ?`, staleStatus, r.feature).Error).To(Succeed())

		state, err := r.tracker.State(ctx, r.root)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.ComputedAt).To(BeTemporally("~", time.Now(), time.Minute), "a stale worktree does not age the refs")
		Expect(state.Worktrees[0].StatusScannedAt).NotTo(BeNil())
		Expect(*state.Worktrees[0].StatusScannedAt).To(BeTemporally("~", time.Now(), time.Minute))
		Expect(state.Worktrees[1].StatusScannedAt).NotTo(BeNil())
		Expect(*state.Worktrees[1].StatusScannedAt).To(BeTemporally("~", staleStatus, time.Second))
	})

	It("tracks a repository whose first scan fails and reports the failure as its state's error", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		git(r.root, "branch", "-m", "main", "trunk")

		_, err := r.tracker.Track(ctx, r.root)
		Expect(err).NotTo(HaveOccurred(), "the repo is registered even though its refs scan failed")

		_, err = r.tracker.State(ctx, r.root)
		Expect(err).To(MatchError(ContainSubstring(`base branch "main" not found`)))
	})

	It("remembers a directory that is not a git work tree instead of running git on every read", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		plain := GinkgoT().TempDir()
		_, first := r.tracker.Track(ctx, plain)
		Expect(first).To(HaveOccurred())

		GinkgoT().Setenv("PATH", GinkgoT().TempDir())
		_, second := r.tracker.Track(ctx, plain)

		Expect(second).To(MatchError(first.Error()), "the second Track answers from memory, without git on PATH")
	})

	It("finds a worktree created after the last scan", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		_, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		late := filepath.Join(filepath.Dir(r.root), "wt-late")
		git(r.root, "worktree", "add", "-q", "-b", "late", late, r.base)

		wt, found, err := r.tracker.WorktreeOf(ctx, r.root, late)

		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(wt.Branch).To(Equal("late"))
	})
})

var _ = Describe("Store", func() {
	It("skips a scan whose lock another session holds", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		store := gitstate.NewStore(r.db)
		key := "git_refs:" + uuid.NewString()
		inner := make(chan bool, 1)

		held, err := store.Locked(ctx, key, false, func(*gitstate.Store) error {
			ran, err := gitstate.NewStore(r.db).Locked(ctx, key, false, func(*gitstate.Store) error { return nil })
			inner <- ran
			return err
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(held).To(BeTrue())
		Expect(<-inner).To(BeFalse())
	})

	It("stores a range's file list once", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		state, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		id, err := r.tracker.Track(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		key := gitstate.RangeKey{Base: r.base, Head: state.Branches[0].Head}
		store := r.tracker.Store()

		Expect(store.SaveRangeFiles(ctx, id, key, []gavelgit.CommitFile{{Path: "a.txt", Status: "added", Adds: 2}})).To(Succeed())
		Expect(store.SaveRangeFiles(ctx, id, key, []gavelgit.CommitFile{{Path: "other.txt"}})).To(Succeed())

		files, stored, err := store.RangeFiles(ctx, id, key)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(BeTrue())
		Expect(files).To(Equal([]gavelgit.CommitFile{{Path: "a.txt", Status: "added", Adds: 2}}))
	})
})

var _ = Describe("porcelain paths", func() {
	It("lists each record's path and skips a rename's source", func() {
		raw := []byte(" M a.txt\x00R  new.txt\x00old.txt\x00?? dir/u.txt\x00")
		Expect(gitstate.PorcelainPaths(raw)).To(Equal([]string{"a.txt", "new.txt", "dir/u.txt"}))
	})
})
