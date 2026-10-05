package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// worktreeRepo is a main checkout with a linked worktree on an unmerged
// branch, a detached worktree, a prunable worktree, a fast-forward merged
// branch and an unmerged branch with no worktree.
type worktreeRepo struct {
	root, featureWT, detachedWT, prunableWT string
	mainHead, featureHead, soloHead, base   string
}

func runGit(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commitIn(dir, file, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644)).To(Succeed())
	runGit(dir, "add", file)
	runGit(dir, "commit", "-q", "-m", message)
	return runGit(dir, "rev-parse", "HEAD")
}

// committedAt is git's own committer date for rev, in UTC.
func committedAt(dir, rev string) time.Time {
	GinkgoHelper()
	at, err := time.Parse(time.RFC3339, runGit(dir, "log", "-1", "--format=%cI", rev))
	Expect(err).NotTo(HaveOccurred())
	return at.UTC()
}

func newWorktreeRepo() worktreeRepo {
	GinkgoHelper()
	tmp, err := filepath.EvalSymlinks(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	r := worktreeRepo{
		root:       filepath.Join(tmp, "repo"),
		featureWT:  filepath.Join(tmp, "wt-feature"),
		detachedWT: filepath.Join(tmp, "wt-detached"),
		prunableWT: filepath.Join(tmp, "wt-gone"),
	}
	Expect(os.MkdirAll(r.root, 0o755)).To(Succeed())
	runGit(r.root, "init", "-q", "-b", "main")
	runGit(r.root, "config", "user.email", "test@example.com")
	runGit(r.root, "config", "user.name", "Test User")
	r.base = commitIn(r.root, "README.md", "one\ntwo\n", "chore: initial")

	runGit(r.root, "worktree", "add", "-q", "-b", "feature", r.featureWT, r.base)
	commitIn(r.featureWT, "a.txt", "a1\na2\na3\n", "feat: a")
	r.featureHead = commitIn(r.featureWT, "b.txt", "b1\n", "feat: b")

	runGit(r.root, "branch", "solo", r.base)
	runGit(r.root, "checkout", "-q", "solo")
	r.soloHead = commitIn(r.root, "README.md", "one\nTWO\n", "fix: shout")
	runGit(r.root, "checkout", "-q", "main")

	runGit(r.root, "checkout", "-q", "-b", "merged")
	merged := commitIn(r.root, "merged.txt", "m\n", "feat: merged")
	runGit(r.root, "checkout", "-q", "main")
	runGit(r.root, "merge", "-q", "--ff-only", "merged")
	Expect(runGit(r.root, "rev-parse", "HEAD")).To(Equal(merged))
	r.mainHead = merged

	runGit(r.root, "worktree", "add", "-q", "--detach", r.detachedWT, r.base)
	runGit(r.root, "worktree", "add", "-q", "--detach", r.prunableWT, r.base)
	Expect(os.RemoveAll(r.prunableWT)).To(Succeed())
	return r
}

var _ = Describe("ListWorktrees", func() {
	It("lists the primary checkout first, then linked, detached and prunable worktrees", func() {
		r := newWorktreeRepo()

		worktrees, err := gavelgit.ListWorktrees(r.root)

		Expect(err).NotTo(HaveOccurred())
		Expect(worktrees).To(HaveLen(4))
		Expect(worktrees[0]).To(Equal(gavelgit.Worktree{Path: r.root, Branch: "main", Head: r.mainHead, Primary: true}))
		Expect(worktrees[1:]).To(ConsistOf(
			gavelgit.Worktree{Path: r.featureWT, Branch: "feature", Head: r.featureHead},
			gavelgit.Worktree{Path: r.detachedWT, Head: r.base, Detached: true},
			gavelgit.Worktree{Path: r.prunableWT, Head: r.base, Detached: true, Prunable: true},
		))
	})

	It("lists the same worktrees when asked from inside a linked worktree", func() {
		r := newWorktreeRepo()

		worktrees, err := gavelgit.ListWorktrees(r.featureWT)

		Expect(err).NotTo(HaveOccurred())
		Expect(worktrees[0].Path).To(Equal(r.root))
		Expect(worktrees[0].Primary).To(BeTrue())
	})

	It("fails for a directory that is not a git repository", func() {
		_, err := gavelgit.ListWorktrees(GinkgoT().TempDir())
		Expect(err).To(MatchError(ContainSubstring("worktree list")))
	})
})

var _ = Describe("UnmergedBranches", func() {
	It("returns branches with commits not in base, with ahead/behind, worktree, diff footprint and last commit time", func() {
		r := newWorktreeRepo()

		branches, err := gavelgit.UnmergedBranches(r.root, "main")

		Expect(err).NotTo(HaveOccurred())
		Expect(branches).To(Equal([]gavelgit.BranchInfo{
			{
				Name: "feature", Head: r.featureHead, Ahead: 2, Behind: 1, Worktree: r.featureWT,
				Diff:         gavelgit.DiffStat{Commits: 2, Files: 2, Adds: 4, Dels: 0},
				LastCommitAt: committedAt(r.root, r.featureHead),
			},
			{
				Name: "solo", Head: r.soloHead, Ahead: 1, Behind: 1,
				Diff:         gavelgit.DiffStat{Commits: 1, Files: 1, Adds: 1, Dels: 1},
				LastCommitAt: committedAt(r.root, r.soloHead),
			},
		}))
	})

	It("fails when the base branch does not exist locally", func() {
		r := newWorktreeRepo()
		_, err := gavelgit.UnmergedBranches(r.root, "trunk")
		Expect(err).To(MatchError(ContainSubstring(`base branch "trunk"`)))
	})
})

var _ = Describe("CommitTime", func() {
	It("returns the committer date of a revision in UTC", func() {
		r := newWorktreeRepo()

		at, err := gavelgit.CommitTime(r.root, r.featureHead)

		Expect(err).NotTo(HaveOccurred())
		Expect(at).To(Equal(committedAt(r.root, r.featureHead)))
		Expect(at.Location()).To(Equal(time.UTC))
	})

	It("fails for a revision git cannot resolve", func() {
		r := newWorktreeRepo()
		_, err := gavelgit.CommitTime(r.root, "no-such-rev")
		Expect(err).To(MatchError(ContainSubstring("no-such-rev")))
	})
})
