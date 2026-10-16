// Tests that a PR client resolved with a token source acts with that source alone, as a hosted
// organization's GitHub App installation does.
package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("a PR client with a token source", func() {
	It("acts with the token source", func() {
		source := func() (string, error) { return "installation-token", nil }
		client, number, err := ResolveURL(Options{TokenSource: source}, "https://github.com/acme/app/pull/7")
		Expect(err).NotTo(HaveOccurred())
		Expect(number).To(Equal(7))
		github, ok := client.(gitHubClient)
		Expect(ok).To(BeTrue())
		Expect(github.opts.Repo).To(Equal("acme/app"))
		Expect(github.opts.TokenSource).NotTo(BeNil())
		token, err := github.opts.TokenSource()
		Expect(err).NotTo(HaveOccurred())
		Expect(token).To(Equal("installation-token"))
	})
})
