package main

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel lint --help", func() {
	It("renders the lint long description and examples", func() {
		lintCmd, _, err := rootCmd.Find([]string{"lint"})
		Expect(err).NotTo(HaveOccurred())
		var out bytes.Buffer
		lintCmd.SetOut(&out)
		lintCmd.SetErr(&out)
		Expect(lintCmd.Help()).To(Succeed())
		Expect(out.String()).To(And(
			ContainSubstring("Automatically detects which linters are available"),
			ContainSubstring("gavel lint --fix"),
		))
	})
})
