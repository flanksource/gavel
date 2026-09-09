package testrunner

import (
	"github.com/flanksource/gavel/internal/changegraph"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("diffOptionsFromRunOptions", func() {
	It("selects nothing when no change selector is set", func() {
		Expect(RunOptions{}.hasChangeSelector()).To(BeFalse())
	})

	// An explicit file list is the caller's statement of what changed. Git's
	// view of the working tree is not consulted, because it can only add files
	// the caller did not name.
	It("takes an explicit file list as the whole set", func() {
		opts := RunOptions{ChangedFiles: []string{"pkg/a.go"}}

		Expect(opts.hasChangeSelector()).To(BeTrue())
		Expect(diffOptionsFromRunOptions(opts)).To(Equal(changegraph.DiffOptions{Files: []string{"pkg/a.go"}}))
	})

	It("unions an explicit file list with --since's working-tree views", func() {
		diff := diffOptionsFromRunOptions(RunOptions{ChangedFiles: []string{"pkg/a.go"}, Since: "main"})

		Expect(diff).To(Equal(changegraph.DiffOptions{
			Files: []string{"pkg/a.go"}, Since: "main",
			IncludeStaged: true, IncludeUnstaged: true, IncludeUntracked: true,
		}))
	})

	It("keeps --changed on the configured base ref", func() {
		GinkgoT().Setenv("GAVEL_CHANGED_BASE", "origin/develop")

		diff := diffOptionsFromRunOptions(RunOptions{Changed: true})

		Expect(diff.Since).To(Equal("origin/develop"))
		Expect(diff.IncludeUnstaged).To(BeTrue())
		Expect(diff.Files).To(BeEmpty())
	})
})
