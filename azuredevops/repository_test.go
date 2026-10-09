package azuredevops

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/url"
)

var _ = Describe("Azure repository references", func() {
	DescribeTable("normalizes hosted remotes",
		func(raw string, project string) {
			repo, err := ParseRepository(raw)
			Expect(err).NotTo(HaveOccurred())
			Expect(repo).To(Equal(Repository{Organization: "acme", Project: project, Name: "service"}))
			Expect(repo.URL()).To(Equal("https://dev.azure.com/acme/" + url.PathEscape(project) + "/_git/service"))
		},
		Entry("HTTPS", "https://dev.azure.com/acme/product/_git/service", "product"),
		Entry("HTTPS username", "https://acme@dev.azure.com/acme/product/_git/service", "product"),
		Entry("encoded project", "https://dev.azure.com/acme/Product%20One/_git/service", "Product One"),
		Entry("SSH scp", "git@ssh.dev.azure.com:v3/acme/product/service", "product"),
		Entry("SSH URL", "ssh://git@ssh.dev.azure.com/v3/acme/product/service", "product"),
		Entry("legacy HTTPS", "https://acme.visualstudio.com/product/_git/service", "product"),
		Entry("legacy collection", "https://acme.visualstudio.com/DefaultCollection/product/_git/service", "product"),
		Entry("legacy SSH", "acme@vs-ssh.visualstudio.com:v3/acme/product/service", "product"),
	)
	DescribeTable("rejects malformed or unhosted targets", func(raw string) {
		_, err := ParseRepository(raw)
		Expect(err).To(HaveOccurred())
	},
		Entry("unrecognized host", "https://example.com/acme/product/_git/service"),
		Entry("deceptive host", "https://dev.azure.com.example.com/acme/product/_git/service"),
		Entry("missing repository", "https://dev.azure.com/acme/product/_git/"),
		Entry("missing SSH project", "git@ssh.dev.azure.com:v3/acme/service"),
		Entry("plain shorthand", "acme/product/service"),
	)
	It("parses a PR URL without losing its repository", func() {
		repo, number, err := ParsePRURL("https://dev.azure.com/acme/product/_git/service/pullrequest/27?_a=files")
		Expect(err).NotTo(HaveOccurred())
		Expect(repo).To(Equal(Repository{Organization: "acme", Project: "product", Name: "service"}))
		Expect(number).To(Equal(27))
	})
	DescribeTable("rejects invalid PR identifiers", func(raw string) {
		_, _, err := ParsePRURL(raw)
		Expect(err).To(HaveOccurred())
	},
		Entry("zero", "https://dev.azure.com/acme/product/_git/service/pullrequest/0"),
		Entry("negative", "https://dev.azure.com/acme/product/_git/service/pullrequest/-1"),
		Entry("extra path", "https://dev.azure.com/acme/product/_git/service/pullrequest/27/other"),
	)
})
