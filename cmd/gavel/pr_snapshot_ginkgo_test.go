package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR reference parsing", func() {
	DescribeTable("accepts every reference form gavel prints",
		func(ref, wantRepo string, wantNumber int) {
			repo, number, err := parsePRRef(ref)
			Expect(err).NotTo(HaveOccurred())
			Expect(repo).To(Equal(wantRepo))
			Expect(number).To(Equal(wantNumber))
		},
		Entry("bare number", "12", "", 12),
		Entry("hash prefixed", "#12", "", 12),
		Entry("repo qualified", "acme/widgets#12", "acme/widgets", 12),
		Entry("pull request URL", "https://github.com/acme/widgets/pull/12", "acme/widgets", 12),
	)

	DescribeTable("rejects references that are not a PR",
		func(ref string) {
			_, _, err := parsePRRef(ref)
			Expect(err).To(HaveOccurred())
		},
		Entry("empty", ""),
		Entry("not a number", "main"),
		Entry("zero", "0"),
		Entry("negative", "-3"),
		Entry("non-PR URL", "https://github.com/acme/widgets/issues/12"),
	)

	It("rejects --pr alongside the other narrowing flags", func() {
		_, err := resolvePRFailed("12", "", ".gavel/last.json", "", prNarrowTests)
		Expect(err).To(MatchError(ContainSubstring("--pr and --failed")))

		_, err = resolvePRFailed("12", "", "", ".gavel/base.json", prNarrowLint)
		Expect(err).To(MatchError(ContainSubstring("--pr and --baseline")))
	})

	It("registers --pr on both test and lint", func() {
		for _, name := range []string{"test", "lint"} {
			command, _, err := rootCmd.Find([]string{name})
			Expect(err).NotTo(HaveOccurred())
			Expect(command.Flags().Lookup("pr")).NotTo(BeNil(), "gavel %s is missing --pr", name)
		}
	})
})
