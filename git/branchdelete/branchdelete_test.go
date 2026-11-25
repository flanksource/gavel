package branchdelete_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/branchdelete"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBranchDelete(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Branch delete") }

func git(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1")
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commit(dir, name, contents string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600)).To(Succeed())
	git(dir, "add", name)
	git(dir, "commit", "-qm", "add "+name)
	return git(dir, "rev-parse", "HEAD")
}

func branchExists(repo, branch string) bool {
	return exec.Command("git", "-C", repo, "rev-parse", "--verify", "refs/heads/"+branch).Run() == nil
}

var _ = Describe("Branch delete", func() {
	var repo, worktree string
	var opts branchdelete.Options
	ctx := context.Background()

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		git(repo, "init", "-qb", "main")
		git(repo, "config", "user.email", "test@example.com")
		git(repo, "config", "user.name", "Test")
		commit(repo, "base.txt", "base\n")
		worktree = filepath.Join(GinkgoT().TempDir(), "topic")
		git(repo, "worktree", "add", "-qb", "topic", worktree)
		opts = branchdelete.Options{Repo: repo, Branch: "topic", Base: "main", Worktree: gavelgit.Worktree{Path: worktree, Branch: "topic"}}
	})

	It("deletes a clean, merged branch and its worktree without any acknowledgement", func() {
		plan, err := branchdelete.Inspect(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{plan.Uncommitted, plan.Unmerged, plan.Worktree}).To(Equal([]any{[]string{}, []branchdelete.Commit{}, worktree}))
		opts.ExpectedHead = plan.Head
		result, err := branchdelete.Delete(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchdelete.Result{WorktreeRemoved: true, BranchDeleted: true}))
		Expect(worktree).NotTo(BeADirectory())
		Expect(branchExists(repo, "topic")).To(BeFalse())
	})

	It("plans uncommitted files and unmerged commits, and refuses to delete them until each is acknowledged", func() {
		sha := commit(worktree, "work.txt", "work\n")
		Expect(os.WriteFile(filepath.Join(worktree, "scratch.txt"), []byte("draft\n"), 0o600)).To(Succeed())
		plan, err := branchdelete.Inspect(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Uncommitted).To(Equal([]string{"scratch.txt"}))
		Expect(plan.Unmerged).To(Equal([]branchdelete.Commit{{SHA: sha, Subject: "add work.txt"}}))

		opts.ExpectedHead = plan.Head
		_, err = branchdelete.Delete(ctx, opts)
		Expect(err).To(MatchError(branchdelete.ErrNeedsConfirmation))
		opts.DiscardUncommitted = true
		_, err = branchdelete.Delete(ctx, opts)
		Expect(err).To(MatchError(ContainSubstring("commits not in main")))
		Expect(worktree).To(BeADirectory())
		Expect(branchExists(repo, "topic")).To(BeTrue())

		opts.DeleteUnmerged = true
		result, err := branchdelete.Delete(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchdelete.Result{WorktreeRemoved: true, BranchDeleted: true}))
		Expect(branchExists(repo, "topic")).To(BeFalse())
	})

	It("treats a commit whose change already landed in the base as merged", func() {
		commit(worktree, "work.txt", "work\n")
		git(repo, "cherry-pick", "topic")
		plan, err := branchdelete.Inspect(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Unmerged).To(BeEmpty())
	})

	It("deletes a branch without a worktree", func() {
		git(repo, "branch", "parked")
		opts = branchdelete.Options{Repo: repo, Branch: "parked", Base: "main"}
		plan, err := branchdelete.Inspect(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		opts.ExpectedHead = plan.Head
		result, err := branchdelete.Delete(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchdelete.Result{BranchDeleted: true}))
	})

	It("refuses when the branch moved since it was planned", func() {
		plan, err := branchdelete.Inspect(ctx, opts)
		Expect(err).NotTo(HaveOccurred())
		commit(worktree, "late.txt", "late\n")
		opts.ExpectedHead, opts.DeleteUnmerged = plan.Head, true
		_, err = branchdelete.Delete(ctx, opts)
		Expect(err).To(MatchError(branchdelete.ErrStale))
		Expect(branchExists(repo, "topic")).To(BeTrue())
	})

	DescribeTable("refuses branches that cannot be deleted", func(mutate func(*branchdelete.Options), message string) {
		mutate(&opts)
		_, err := branchdelete.Inspect(ctx, opts)
		Expect(err).To(MatchError(branchdelete.ErrRefused))
		Expect(err).To(MatchError(ContainSubstring(message)))
	},
		Entry("the base branch", func(o *branchdelete.Options) { o.Branch, o.Worktree = "main", gavelgit.Worktree{} }, "is the base branch"),
		Entry("the main checkout's branch", func(o *branchdelete.Options) { o.Worktree.Primary = true }, "main checkout"),
		Entry("an invalid name", func(o *branchdelete.Options) { o.Branch = "--force" }, "branch cannot be deleted"),
	)
})
