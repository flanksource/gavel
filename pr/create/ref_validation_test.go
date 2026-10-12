package prcreate

import (
	"context"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// optionLike is a value git would parse as an option that runs a program, were
// it to reach a command line unchecked.
const optionLike = "--upload-pack=touch pwned"

var _ = Describe("ref validation", func() {
	var f *repoFixture
	BeforeEach(func() { f = newRepoFixture() })

	It("refuses an option-like base ref before running git", func() {
		_, _, err := preflight(f.repo, Input{SHAs: []string{f.topicSHA}, Base: optionLike})
		Expect(err).To(MatchError(ContainSubstring(`prcreate: base ref: invalid revision "--upload-pack=touch pwned": must not begin with "-"`)))
	})

	It("refuses an option-like commit before running git", func() {
		_, _, err := preflight(f.repo, Input{SHAs: []string{optionLike}, Base: testBase})
		Expect(err).To(MatchError(ContainSubstring(`prcreate: commit: invalid revision "--upload-pack=touch pwned"`)))
	})

	It("refuses a generated branch name git would reject, pushing nothing", func() {
		var prs []model.CreatePRInput
		deps := recordingDeps(commitpkg.PRContent{Title: "feat: x", Branch: "feat/../x"}, &prs, nil)

		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Deps: deps})

		Expect(err).To(MatchError(ContainSubstring(`generated PR branch: invalid branch name`)))
		Expect(prs).To(BeEmpty())
		Expect(git(f.bare, "branch", "--list")).To(Equal("* main"))
		Expect(f.scratchEntries()).To(BeEmpty())
	})
})
