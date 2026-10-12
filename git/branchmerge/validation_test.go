package branchmerge_test

import (
	"github.com/flanksource/gavel/git/branchmerge"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// optionLike is a value git would parse as an option that runs a program, were
// it to reach a command line unchecked.
const optionLike = "--upload-pack=touch pwned"

var _ = Describe("option-like refs", func() {
	var f *mergeFixture

	BeforeEach(func() { f = newMergeFixture() })

	It("refuses a merge of an option-like branch before running git", func() {
		before := git(f.repo, "rev-parse", "HEAD")

		_, err := branchmerge.Merge(branchmerge.Options{Repo: f.repo, Branch: optionLike, Mode: branchmerge.Squash, Message: "m"})

		Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
		Expect(err).To(MatchError(ContainSubstring(`must not begin with "-"`)))
		Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(before))
	})

	It("refuses an option-like commit to cherry-pick", func() {
		_, err := branchmerge.Merge(branchmerge.Options{
			Repo: f.repo, Branch: featureBranch, Commits: []string{f.commits[0], optionLike}, Mode: branchmerge.Incremental,
		})

		Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
		Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(f.base))
	})

	It("refuses a cleanup with an option-like landed commit before touching the worktree or branch", func() {
		_, err := branchmerge.Cleanup(branchmerge.CleanupOptions{
			Repo: f.repo, Worktree: f.worktree, Branch: featureBranch, LandedSHA: optionLike, Mode: branchmerge.Incremental,
		})

		Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
		Expect(f.worktree).To(BeADirectory())
		Expect(branchExists(f.repo, featureBranch)).To(BeTrue())
	})

	It("refuses an option-like range bound", func() {
		_, err := branchmerge.RangeCommits(f.repo, "--all", featureBranch)

		Expect(err).To(MatchError(branchmerge.ErrInvalidOptions))
	})
})
