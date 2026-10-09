package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure PR status targets", func() {
	DescribeTable("accepts hosted PR URLs and repository URLs", func(args []string, repo string) {
		target, err := parseStatusTarget(args)
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal(statusTarget{Repo: repo, PR: 27}))
	}, Entry("modern", []string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, "https://dev.azure.com/acme/product/_git/service"),
		Entry("legacy", []string{"https://acme.visualstudio.com/product/_git/service/pullrequest/27"}, "https://dev.azure.com/acme/product/_git/service"),
		Entry("repo and number", []string{"https://dev.azure.com/acme/product/_git/service", "27"}, "https://dev.azure.com/acme/product/_git/service"))
	It("rejects unsupported Azure flags before requesting credentials", func() {
		_, err := runPRStatus(PRStatusOptions{Args: []string{"https://dev.azure.com/acme/product/_git/service/pullrequest/27"}, AIFix: true})
		Expect(err).To(MatchError(ContainSubstring("--ai-fix is not supported")))
	})
})
