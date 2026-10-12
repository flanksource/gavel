package branchmerge_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/git/branchmerge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Merge", func() {
	var f *mergeFixture

	BeforeEach(func() { f = newMergeFixture() })

	Describe("incremental", func() {
		It("cherry-picks each commit onto the checked-out base branch", func() {
			result, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Base: "main", Commits: f.commits, Mode: branchmerge.Incremental,
			})

			Expect(err).NotTo(HaveOccurred())
			head := git(f.repo, "rev-parse", "HEAD")
			Expect(result).To(Equal(branchmerge.Result{TargetBranch: "main", LandedSHA: head, Commits: 2}))
			Expect(git(f.repo, "log", "--format=%s", f.base+"..HEAD")).To(Equal("feat: add b\nfeat: add a"))
		})

		It("aborts a conflicting cherry-pick and leaves the checkout where it was", func() {
			before := commitFile(f.repo, "a.txt", "main's a\n", "chore: conflicting a")

			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Commits: f.commits, Mode: branchmerge.Incremental,
			})

			var conflict *branchmerge.ConflictError
			Expect(errors.As(err, &conflict)).To(BeTrue(), "expected *ConflictError, got %v", err)
			Expect(conflict.Paths).To(Equal([]string{"a.txt"}))
			Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(before))
			Expect(git(f.repo, "status", "--porcelain")).To(BeEmpty())
			Expect(filepath.Join(f.repo, ".git", "CHERRY_PICK_HEAD")).NotTo(BeAnExistingFile())
		})

		It("refuses a staged index in the target checkout", func() {
			Expect(os.WriteFile(filepath.Join(f.repo, "staged.txt"), []byte("x\n"), 0o644)).To(Succeed())
			git(f.repo, "add", "staged.txt")

			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Commits: f.commits, Mode: branchmerge.Incremental,
			})

			Expect(err).To(MatchError(branchmerge.ErrCheckoutNotReady))
			Expect(err).To(MatchError(ContainSubstring("staged.txt")))
			Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(f.base))
		})

		It("refuses a target checkout that is not on the base branch, naming both", func() {
			git(f.repo, "checkout", "-q", "-b", "other")

			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Base: "main", Commits: f.commits, Mode: branchmerge.Incremental,
			})

			Expect(err).To(MatchError(branchmerge.ErrCheckoutNotReady))
			Expect(err).To(MatchError(And(ContainSubstring(`"other"`), ContainSubstring(`"main"`))))
		})

		It("refuses a target checkout that has the source branch checked out", func() {
			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.worktree, Branch: featureBranch, Commits: f.commits, Mode: branchmerge.Incremental,
			})
			Expect(err).To(MatchError(branchmerge.ErrCheckoutNotReady))
		})

		It("requires the commits to pick", func() {
			_, err := branchmerge.Merge(branchmerge.Options{Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Incremental})
			Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
		})
	})

	Describe("squash", func() {
		It("lands the branch as one commit with the given message", func() {
			commitFile(f.repo, "main.txt", "main moved on\n", "chore: main moves")
			before := git(f.repo, "rev-parse", "HEAD")

			result, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Base: "main", Mode: branchmerge.Squash, Message: "feat: a and b\n\nBoth files.",
			})

			Expect(err).NotTo(HaveOccurred())
			head := git(f.repo, "rev-parse", "HEAD")
			Expect(result).To(Equal(branchmerge.Result{TargetBranch: "main", LandedSHA: head, Commits: 2}))
			Expect(git(f.repo, "rev-parse", "HEAD~1")).To(Equal(before))
			Expect(git(f.repo, "log", "-1", "--format=%B")).To(Equal("feat: a and b\n\nBoth files."))
			Expect(git(f.repo, "show", "--name-only", "--format=", "HEAD")).To(Equal("a.txt\nb.txt"))
			Expect(git(f.repo, "status", "--porcelain")).To(BeEmpty())
		})

		It("aborts a conflicting squash and restores HEAD and the index", func() {
			before := commitFile(f.repo, "a.txt", "main's a\n", "chore: conflicting a")

			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Squash, Message: "feat: squash",
			})

			var conflict *branchmerge.ConflictError
			Expect(errors.As(err, &conflict)).To(BeTrue(), "expected *ConflictError, got %v", err)
			Expect(conflict.Paths).To(Equal([]string{"a.txt"}))
			Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(before))
			Expect(git(f.repo, "status", "--porcelain")).To(BeEmpty())
			Expect(filepath.Join(f.repo, ".git", "SQUASH_MSG")).NotTo(BeAnExistingFile())
		})

		It("requires a message and no explicit commits", func() {
			_, err := branchmerge.Merge(branchmerge.Options{Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Squash})
			Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
			_, err = branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Squash, Message: "m", Commits: f.commits,
			})
			Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
		})

		It("refuses a branch with nothing to squash", func() {
			git(f.repo, "merge", "-q", "--ff-only", featureBranch)

			_, err := branchmerge.Merge(branchmerge.Options{
				Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Squash, Message: "feat: squash",
			})

			Expect(err).To(MatchError(branchmerge.ErrNothingToMerge))
		})
	})

	It("rejects an unknown mode", func() {
		_, err := branchmerge.Merge(branchmerge.Options{Repo: f.repo, Branch: featureBranch, Mode: "rebase"})
		Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
	})
})

var _ = Describe("Cleanup", func() {
	var f *mergeFixture

	BeforeEach(func() { f = newMergeFixture() })

	It("removes the worktree and deletes the branch once every picked commit landed", func() {
		landed, err := branchmerge.Merge(branchmerge.Options{
			Repo: f.repo, Branch: featureBranch, Commits: f.commits, Mode: branchmerge.Incremental,
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Worktree: f.worktree, Branch: featureBranch, LandedSHA: landed.LandedSHA, Mode: branchmerge.Incremental,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchmerge.CleanupResult{WorktreeRemoved: true, BranchDeleted: true}))
		Expect(f.worktree).NotTo(BeAnExistingFile())
		Expect(branchExists(f.repo, featureBranch)).To(BeFalse())
	})

	It("deletes a squashed branch even though main moved on before the squash", func() {
		commitFile(f.repo, "main.txt", "main moved on\n", "chore: main moves")
		landed, err := branchmerge.Merge(branchmerge.Options{
			Repo: f.repo, Branch: featureBranch, Mode: branchmerge.Squash, Message: "feat: squash",
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Worktree: f.worktree, Branch: featureBranch, LandedSHA: landed.LandedSHA, Mode: branchmerge.Squash,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchmerge.CleanupResult{WorktreeRemoved: true, BranchDeleted: true}))
		Expect(branchExists(f.repo, featureBranch)).To(BeFalse())
	})

	It("keeps a squash branch whose tree is not in the landed commit", func() {
		_, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Branch: featureBranch, LandedSHA: f.base, Mode: branchmerge.Squash,
		})

		Expect(err).To(MatchError(ContainSubstring("kept branch " + featureBranch)))
		Expect(branchExists(f.repo, featureBranch)).To(BeTrue())
	})

	It("keeps an incremental branch with commits not in the landed commit", func() {
		_, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Branch: featureBranch, LandedSHA: f.base, Mode: branchmerge.Incremental,
		})

		Expect(err).To(MatchError(ContainSubstring("kept branch " + featureBranch)))
		Expect(branchExists(f.repo, featureBranch)).To(BeTrue())
	})

	It("ignores commits at or before Limit when checking an incremental landing", func() {
		landed, err := branchmerge.Merge(branchmerge.Options{
			Repo: f.repo, Branch: featureBranch, Commits: f.commits[1:], Mode: branchmerge.Incremental,
		})
		Expect(err).NotTo(HaveOccurred())

		result, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Worktree: f.worktree, Branch: featureBranch, LandedSHA: landed.LandedSHA,
			Limit: f.commits[0], Mode: branchmerge.Incremental,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(branchmerge.CleanupResult{WorktreeRemoved: true, BranchDeleted: true}))
	})
})

var _ = Describe("RangeCommits", func() {
	It("lists from..to oldest first", func() {
		f := newMergeFixture()
		Expect(branchmerge.RangeCommits(f.repo, "main", featureBranch)).To(Equal(f.commits))
	})
})
