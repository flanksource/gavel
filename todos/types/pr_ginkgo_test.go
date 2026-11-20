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
		Entry("GitLab merge request page past its number", "https://gitlab.example.com/acme/service/-/merge_requests/27/diffs"),
		Entry("GitLab new merge request form", "https://gitlab.example.com/acme/service/-/merge_requests/new"),
		Entry("GitLab merge request without a project", "https://gitlab.example.com/-/merge_requests/27"),
		Entry("GitLab over plain HTTP", "http://gitlab.example.com/acme/service/-/merge_requests/27"),
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
		Entry("GitLab", "https://gitlab.com/acme/service/-/merge_requests/27"),
		Entry("self-managed GitLab with nested groups", "https://gitlab.example.com:8443/acme/platform/service/-/merge_requests/27"),
	)
})
