package gitstate_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const feature = "feature/a"

func gitIn(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commit(dir, name, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
	gitIn(dir, "add", name)
	gitIn(dir, "commit", "-q", "-m", message)
	return gitIn(dir, "rev-parse", "HEAD")
}

func committedAt(dir, rev string) time.Time {
	GinkgoHelper()
	at, err := time.Parse(time.RFC3339, gitIn(dir, "log", "-1", "--format=%cI", rev))
	Expect(err).NotTo(HaveOccurred())
	return at.UTC()
}

var _ = Describe("Compute", func() {
	// The repository is "acme": a main checkout on main, a linked worktree on
	// feature/a with two commits (a.txt +2, b.txt +1) and one uncommitted line in
	// b.txt, and a worktree-less branch solo that rewrites README.md's only line.
	var repo, worktree, featureHead, soloHead string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
		GinkgoT().Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
		root, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		repo, worktree = filepath.Join(root, "acme"), filepath.Join(root, "acme-feature")
		gitIn(root, "init", "-q", "-b", "main", repo)
		gitIn(repo, "config", "user.email", "test@example.com")
		gitIn(repo, "config", "user.name", "test")
		base := commit(repo, "README.md", "hello\n", "chore: initial")
		gitIn(repo, "worktree", "add", "-q", "-b", feature, worktree, base)
		commit(worktree, "a.txt", "a1\na2\n", "feat: add a")
		featureHead = commit(worktree, "b.txt", "b\n", "feat: add b")
		Expect(os.WriteFile(filepath.Join(worktree, "b.txt"), []byte("b\nmore\n"), 0o644)).To(Succeed())
		gitIn(repo, "checkout", "-q", "-b", "solo")
		soloHead = commit(repo, "README.md", "HELLO\n", "fix: shout hello")
		gitIn(repo, "checkout", "-q", "main")
	})

	It("reports the base, the primary checkout, each worktree's changes, ages and the unmerged branches", func() {
		touched := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
		Expect(os.Chtimes(filepath.Join(worktree, "b.txt"), touched, touched)).To(Succeed())
		before := time.Now().UTC()

		state, err := gitstate.Compute(context.Background(), repo)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.ComputedAt).To(BeTemporally(">=", before.Truncate(time.Second)))
		main := gitIn(repo, "rev-parse", "HEAD")
		Expect(state).To(Equal(gitstate.State{
			Base: "main", CurrentBranch: "main", BaseCheckedOut: true, ComputedAt: state.ComputedAt,
			Worktrees: []gitstate.Worktree{
				{Worktree: gavelgit.Worktree{Path: repo, Branch: "main", Head: main, Primary: true}, LastCommitAt: committedAt(repo, main)},
				{
					Worktree: gavelgit.Worktree{Path: worktree, Branch: feature, Head: featureHead},
					Changes:  gitstate.Changes{Unstaged: 1, Adds: 1}, Ahead: 2,
					LastCommitAt: committedAt(repo, featureHead), TouchedAt: &touched,
				},
			},
			Branches: []gavelgit.BranchInfo{
				{Name: feature, Head: featureHead, Ahead: 2, Worktree: worktree, Diff: gavelgit.DiffStat{Commits: 2, Files: 2, Adds: 3}, LastCommitAt: committedAt(repo, featureHead)},
				{Name: "solo", Head: soloHead, Ahead: 1, Diff: gavelgit.DiffStat{Commits: 1, Files: 1, Adds: 1, Dels: 1}, LastCommitAt: committedAt(repo, soloHead)},
			},
		}))
		// feature/a: a.txt +2, b.txt +1; solo: README +1 -1; worktree: b.txt +1.
		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{Name: "acme", Base: "main", Adds: 5, Dels: 1, Worktrees: 1, Branches: 2}))
	})
})
