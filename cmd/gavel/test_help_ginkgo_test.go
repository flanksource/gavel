package main

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
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
	)
	It("describes every visible test option", func() {
		testCmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			if !flag.Hidden {
				Expect(flag.Usage).NotTo(BeEmpty(), "--%s", flag.Name)
			}
		})
	})
	It("gives each public test command its own usage and examples", func() {
		for _, cmd := range testCmd.Commands() {
			if !cmd.IsAvailableCommand() {
				continue
			}
			Expect(cmd.Long).NotTo(Equal(testCmd.Long), cmd.CommandPath())
			Expect(cmd.Example).To(ContainSubstring(cmd.CommandPath()+" "), cmd.CommandPath())
			Expect(cmd.UsageString()).To(ContainSubstring("Use gavel --help to show arguments available to all commands."))
			if cmd.Name() == "ansi" {
				Expect(cmd.UseLine()).To(ContainSubstring("[flags] [out.json] -- <command> [args...]"))
			} else {
				Expect(cmd.UseLine()).To(ContainSubstring("[path...]"), cmd.CommandPath())
			}
		}
	})
})
