package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestProvider(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "PR providers") }

var _ = Describe("provider resolution", func() {
	DescribeTable("resolves GitHub and hosted Azure repository references", func(ref, kind, repo string) {
		client, err := Resolve(Options{Repo: ref})
		Expect(err).NotTo(HaveOccurred())
		Expect(client.Kind()).To(Equal(kind))
		if kind == "github" {
			Expect(client.(gitHubClient).opts.Repo).To(Equal(repo))
		}
	}, Entry("GitHub shorthand", "acme/service", "github", "acme/service"),
		Entry("GitHub SSH", "git@github.com:acme/service.git", "github", "acme/service"),
		Entry("modern Azure HTTPS", "https://dev.azure.com/acme/product/_git/service", "azuredevops", ""),
		Entry("Azure SSH", "git@ssh.dev.azure.com:v3/acme/product/service", "azuredevops", ""))
	It("rejects unsupported repository hosts", func() {
		_, err := Resolve(Options{Repo: "https://git.example/acme/service"})
		Expect(err).To(HaveOccurred())
	})
})
