package commit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("commit target parsing", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "feature")
		runGit(workDir, "remote", "add", "origin", "https://github.com/acme/service.git")
		runGit(workDir, "remote", "add", "zim", "https://github.com/acme-fork/service.git")
	})

	DescribeTable("resolves the remote, base and optional topic",
		func(value string, expected commitTarget) {
			Expect(parseCommitTarget(workDir, value)).To(Equal(expected))
		},
		Entry("defaults to origin", "main", commitTarget{remote: "origin", base: "main", repo: "https://github.com/acme/service.git"}),
		Entry("selects a remote", "zim/main", commitTarget{remote: "zim", base: "main", repo: "https://github.com/acme-fork/service.git"}),
		Entry("selects a topic", "main:new-topic", commitTarget{remote: "origin", base: "main", head: "new-topic", repo: "https://github.com/acme/service.git"}),
		Entry("selects both", "zim/main:feat/new-topic", commitTarget{remote: "zim", base: "main", head: "feat/new-topic", repo: "https://github.com/acme-fork/service.git"}),
		Entry("supports a nested base", "zim/release/stable:fix/patch", commitTarget{remote: "zim", base: "release/stable", head: "fix/patch", repo: "https://github.com/acme-fork/service.git"}),
	)

	DescribeTable("rejects malformed targets",
		func(value string) {
			_, err := parseCommitTarget(workDir, value)
			Expect(err).To(HaveOccurred())
		},
		Entry("empty base", ""),
		Entry("empty remote", "/main"),
		Entry("empty remote base", "zim/"),
		Entry("empty topic", "main:"),
		Entry("extra colon", "main:topic:other"),
		Entry("invalid base", "main..next"),
		Entry("invalid topic", "main:bad topic"),
		Entry("option as base", "--all"),
		Entry("unknown remote", "missing/main"),
		Entry("same topic and base", "main:main"),
	)
})
