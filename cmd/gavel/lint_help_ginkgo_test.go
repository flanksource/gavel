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
			ContainSubstring("--linters jscpd,eslint"),
			ContainSubstring("gavel lint --fix"),
			ContainSubstring("[path...] [-- linter args...]"),
		))
	})
	It("gives each public lint command its own synopsis and example", func() {
		parent, _, err := rootCmd.Find([]string{"lint"})
		Expect(err).NotTo(HaveOccurred())
		for _, cmd := range parent.Commands() {
			if !cmd.IsAvailableCommand() {
				continue
			}
			Expect(cmd.Short).NotTo(BeEmpty(), cmd.CommandPath())
			Expect(cmd.Long).NotTo(Equal(parent.Long), cmd.CommandPath())
			Expect(cmd.Example).To(ContainSubstring(cmd.CommandPath()+" "), cmd.CommandPath())
			Expect(cmd.UsageString()).To(ContainSubstring("Use gavel --help to show arguments available to all commands."))
			if cmd.Name() == "json" || cmd.Name() == "yaml" {
				Expect(cmd.Use).NotTo(ContainSubstring("[--"))
				Expect(cmd.Example).To(ContainSubstring("--schema"))
			}
			if cmd.Name() == "github" {
				Expect(cmd.Use).NotTo(ContainSubstring("[path"))
				Expect(cmd.Example).To(ContainSubstring("-- --repo"))
			}
		}
	})
})
