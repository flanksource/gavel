package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR link references", func() {
	DescribeTable("derive the provider, repository key and number from a hosted PR URL",
		func(raw string, want PRRef) {
			ref, err := PRRefFromURL(raw)
			Expect(err).NotTo(HaveOccurred())
			Expect(ref).To(Equal(want))
		},
		Entry("GitHub keys the repository case-insensitively", "https://github.com/Acme/Widgets/pull/12",
			PRRef{Provider: PRProviderGitHub, Repo: "acme/widgets", Number: 12, URL: "https://github.com/Acme/Widgets/pull/12"}),
		Entry("Azure with a project", "https://dev.azure.com/Org/Product/_git/Service/pullrequest/28",
			PRRef{Provider: PRProviderAzure, Repo: "org/product/service", Number: 28, URL: "https://dev.azure.com/Org/Product/_git/Service/pullrequest/28"}),
		Entry("Azure without a project", "https://dev.azure.com/org/_git/service/pullrequest/3",
			PRRef{Provider: PRProviderAzure, Repo: "org/service", Number: 3, URL: "https://dev.azure.com/org/_git/service/pullrequest/3"}),
		Entry("Azure legacy host drops DefaultCollection", "https://org.visualstudio.com/DefaultCollection/product/_git/service/pullrequest/27",
			PRRef{Provider: PRProviderAzure, Repo: "org/product/service", Number: 27, URL: "https://org.visualstudio.com/DefaultCollection/product/_git/service/pullrequest/27"}),
	)

	It("refuses a URL that is not a hosted PR", func() {
		_, err := PRRefFromURL("https://github.com/acme/widgets/issues/12")
		Expect(err).To(MatchError(ContainSubstring("unsupported PR URL")))
	})

	It("keys two spellings of one PR identically", func() {
		upper, err := PRRefFromURL("https://github.com/ACME/widgets/pull/12")
		Expect(err).NotTo(HaveOccurred())
		lower, err := PRRefFromURL("https://github.com/acme/widgets/pull/12")
		Expect(err).NotTo(HaveOccurred())
		Expect(upper.SamePR(lower)).To(BeTrue())
		Expect(upper.SamePR(PRRef{Provider: PRProviderGitHub, Repo: "acme/widgets", Number: 13})).To(BeFalse())
	})

	DescribeTable("parses the PR relation, defaulting to related",
		func(raw string, want PRRelation, fails bool) {
			got, err := ParsePRRelation(raw)
			if fails {
				Expect(err).To(HaveOccurred())
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("empty", "", PRRelationRelated, false),
		Entry("implements", " Implements ", PRRelationImplements, false),
		Entry("related", "related", PRRelationRelated, false),
		Entry("unknown", "fixes", PRRelation(""), true),
	)

	It("accepts only the known link sources", func() {
		Expect(PRLinkSourceTriage.Validate()).To(Succeed())
		Expect(PRLinkSourceManual.Validate()).To(Succeed())
		Expect(PRLinkSource("bot").Validate()).To(MatchError(ContainSubstring("unknown PR link source")))
	})
})
