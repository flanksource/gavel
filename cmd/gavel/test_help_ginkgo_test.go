package main

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel test --help", func() {
	DescribeTable("renders the focus, file path and passthrough guidance",
		func(args []string) {
			cmd, _, err := rootCmd.Find(args)
			Expect(err).NotTo(HaveOccurred())
			var out bytes.Buffer
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			Expect(cmd.Help()).To(Succeed())
			Expect(out.String()).To(And(
				ContainSubstring("--focus PATTERN runs only the tests whose name matches"),
				ContainSubstring("A test file runs just that"),
				ContainSubstring("--ginkgo.focus"),
				ContainSubstring("--testNamePattern"),
				ContainSubstring("--recursive=false tests only the named directories"),
			))
		},
		Entry("gavel test", []string{"test"}),
		Entry("gavel test ginkgo", []string{"test", "ginkgo"}),
	)
})
