package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/verify"
)

var _ = Describe("commit batch CLI options", func() {
	It("maps -i --batch into commit options", func() {
		got := buildCommitOptions(CommitOptions{Interactive: true, Batch: true}, "/repo", verify.GavelConfig{}, nil)
		Expect(got.Interactive).To(BeTrue())
		Expect(got.Batch).To(BeTrue())
	})

	It("registers b as the batch shorthand", func() {
		command, _, err := rootCmd.Find([]string{"commit"})
		Expect(err).NotTo(HaveOccurred())
		Expect(command.Flags().Lookup("batch")).NotTo(BeNil())
		Expect(command.Flags().Lookup("batch").Shorthand).To(Equal("b"))
	})
})
