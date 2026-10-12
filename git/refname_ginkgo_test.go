package git_test

import (
	gavelgit "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ValidateRevision", func() {
	DescribeTable("accepts the revision syntax callers pass to git",
		func(rev string) {
			Expect(gavelgit.ValidateRevision(rev)).To(Succeed())
		},
		Entry("HEAD", "HEAD"),
		Entry("an ancestor of HEAD", "HEAD~2"),
		Entry("a short sha", "abc1234"),
		Entry("a full sha", "0123456789abcdef0123456789abcdef01234567"),
		Entry("a remote-tracking branch", "origin/main"),
		Entry("a fully qualified branch", "refs/heads/feature/x"),
		Entry("a peeled commit", "abc1234^{commit}"),
	)

	DescribeTable("rejects a value git would not read as a revision",
		func(rev, reason string) {
			Expect(gavelgit.ValidateRevision(rev)).To(MatchError(ContainSubstring(reason)))
		},
		Entry("empty", "", "is empty"),
		Entry("an option", "--upload-pack=touch /tmp/pwned", `must not begin with "-"`),
		Entry("a short option", "-n", `must not begin with "-"`),
		Entry("embedded whitespace", "main extra", "contains whitespace or control character"),
		Entry("a newline", "main\n--all", "contains whitespace or control character"),
		Entry("a NUL", "main\x00", "contains whitespace or control character"),
	)
})

const badBranchChars = `must not contain a space, a control character, or any of ~^:?*[\`

var _ = Describe("ValidateBranchName", func() {
	DescribeTable("accepts names git check-ref-format --branch accepts",
		func(name string) {
			Expect(gavelgit.ValidateBranchName(name)).To(Succeed())
		},
		Entry("a plain name", "main"),
		Entry("a nested name", "feat/todo-hierarchy-and-presets"),
		Entry("a generated topic with a suffix", "fix/login-redirect-m1x2y3"),
		Entry("dots inside a component", "release/v1.2.3"),
	)

	DescribeTable("rejects names git check-ref-format --branch rejects",
		func(name, reason string) {
			Expect(gavelgit.ValidateBranchName(name)).To(MatchError(ContainSubstring(reason)))
		},
		Entry("empty", "", "is empty"),
		Entry("an option", "-D", `must not begin with "-"`),
		Entry("HEAD", "HEAD", `"HEAD" is reserved`),
		Entry("a lone @", "@", `"@" is reserved`),
		Entry("a double dot", "a..b", `must not contain ".."`),
		Entry("a reflog selector", "a@{1}", `must not contain "@{"`),
		Entry("a refspec separator", "main:refs/heads/x", badBranchChars),
		Entry("a revision suffix", "main~1", badBranchChars),
		Entry("a peel suffix", "main^0", badBranchChars),
		Entry("a glob", "feat/*", badBranchChars),
		Entry("a character class", "feat/[ab]", badBranchChars),
		Entry("a question mark", "feat?", badBranchChars),
		Entry("a backslash", `a\b`, badBranchChars),
		Entry("a space", "a b", badBranchChars),
		Entry("a control character", "a\x7fb", badBranchChars),
		Entry("a leading slash", "/main", `must not begin or end with "/"`),
		Entry("a trailing slash", "main/", `must not begin or end with "/"`),
		Entry("an empty component", "a//b", "empty path component"),
		Entry("a hidden component", "a/.b", `component ".b" must not begin with "."`),
		Entry("a .lock component", "a.lock/b", `component "a.lock" must not end with ".lock"`),
		Entry("a trailing dot", "main.", `must not end with "."`),
	)
})
