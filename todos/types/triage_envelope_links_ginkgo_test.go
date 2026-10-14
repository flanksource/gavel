package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("triage duplicate and PR links", func() {
	const prURL = "https://github.com/acme/widgets/pull/12"

	DescribeTable("validates the non-closing link fields",
		func(mutate func(*TriageEnvelope), wantErr string) {
			env := validTriage()
			mutate(&env)
			err := env.Validate()
			if wantErr == "" {
				Expect(err).NotTo(HaveOccurred())
				return
			}
			Expect(err).To(MatchError(ContainSubstring(wantErr)))
		},
		Entry("duplicates and PRs on a ready verdict are valid", func(e *TriageEnvelope) {
			e.Duplicates = []string{"ff0011"}
			e.PRs = []string{prURL}
		}, ""),
		Entry("a blank duplicate is rejected", func(e *TriageEnvelope) {
			e.Duplicates = []string{" "}
		}, "duplicates contains a blank id"),
		Entry("a duplicate that is also the duplicate-of survivor is rejected", func(e *TriageEnvelope) {
			e.Verdict = VerdictDuplicateOf
			e.DuplicateOf = "ff0011"
			e.Comment = "same bug"
			e.Duplicates = []string{"FF0011"}
		}, `duplicates names "FF0011", which duplicateOf already closes this TODO into`),
		Entry("a duplicate that is also folded in is rejected", func(e *TriageEnvelope) {
			e.Verdict = VerdictMergeInto
			e.Merges = []string{"cc3344"}
			e.Body = "combined"
			e.Duplicates = []string{"cc3344"}
		}, `duplicates names "cc3344", which merges already folds into this TODO`),
		Entry("a PR that is not a hosted PR URL is rejected", func(e *TriageEnvelope) {
			e.PRs = []string{"https://github.com/acme/widgets/issues/12"}
		}, "triage prs"),
		Entry("a blank PR is rejected", func(e *TriageEnvelope) {
			e.PRs = []string{""}
		}, "triage prs contains a blank URL"),
	)
})
