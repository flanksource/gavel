package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("test outline options", func() {
	DescribeTable("normalizes supported framework names",
		func(input string, expected parsers.Framework) {
			framework, err := normalizeOutlineFramework(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(framework).To(Equal(expected))
		},
		Entry("Go", "go", parsers.GoTest),
		Entry("Go alias", "gotest", parsers.GoTest),
		Entry("Ginkgo", "ginkgo", parsers.Ginkgo),
		Entry("Jest", "jest", parsers.Jest),
		Entry("Vitest", "vitest", parsers.Vitest),
		Entry("Playwright", "playwright", parsers.Playwright),
		Entry("fixture singular", "fixture", parsers.Fixture),
		Entry("fixture plural", "fixtures", parsers.Fixture),
	)

	It("documents every outline source and the fixture override", func() {
		opts := testOutlineOptions{FixtureFiles: []string{"checks/*.fixture.md"}}
		Expect(opts.Help().String()).To(And(
			ContainSubstring("Go test"),
			ContainSubstring("Ginkgo spec"),
			ContainSubstring("Jest test"),
			ContainSubstring("Vitest test"),
			ContainSubstring("Playwright test"),
			ContainSubstring("Markdown fixture"),
		))
		Expect(opts.FixtureFiles).To(Equal([]string{"checks/*.fixture.md"}))
	})

	It("rejects unknown framework names", func() {
		_, err := normalizeOutlineFramework("unknown")
		Expect(err).To(MatchError(ContainSubstring("known")))
	})
})
