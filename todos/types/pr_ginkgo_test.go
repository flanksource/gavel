package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("explicit PR URL validation", func() {
	DescribeTable("rejects malformed or unrelated URLs", func(ref string) {
		_, err := PRFromURL(ref)
		Expect(err).To(HaveOccurred())
	},
		Entry("empty owner", "https://github.com//service/pull/27"),
		Entry("wrong Azure path", "https://dev.azure.com/org/project/_git/service/extra/pullrequest/27"),
		Entry("nested Azure host", "https://extra.org.visualstudio.com/project/_git/service/pullrequest/27"),
		Entry("credentials", "https://user:secret@github.com/acme/service/pull/27"),
		Entry("number mismatch", "https://github.com/acme/service/pull/0"),
	)
	DescribeTable("accepts hosted PR links", func(ref string) {
		pr, err := PRFromURL(ref)
		Expect(err).NotTo(HaveOccurred())
		Expect(pr.Validate()).To(Succeed())
		Expect(pr.Number).To(Equal(27))
	},
		Entry("GitHub", "https://github.com/acme/service/pull/27"),
		Entry("Azure", "https://dev.azure.com/org/project/_git/service/pullrequest/27"),
		Entry("Azure collection", "https://org.visualstudio.com/DefaultCollection/project/_git/service/pullrequest/27"),
	)
})
