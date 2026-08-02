package main

import (
	"time"

	"github.com/flanksource/gavel/testrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var _ = Describe("test framework command flags", func() {
	DescribeTable("mirrors common test flags",
		func(name string) {
			command, _, err := testCmd.Find([]string{name})
			Expect(err).NotTo(HaveOccurred())

			testCmd.LocalNonPersistentFlags().VisitAll(func(parentFlag *pflag.Flag) {
				if parentFlag.Name == "framework" {
					return
				}
				childFlag := command.Flags().Lookup(parentFlag.Name)
				Expect(childFlag).NotTo(BeNil(), "%s missing --%s", name, parentFlag.Name)
				Expect(childFlag.DefValue).To(Equal(parentFlag.DefValue), "%s --%s default", name, parentFlag.Name)
				Expect(childFlag.NoOptDefVal).To(Equal(parentFlag.NoOptDefVal), "%s --%s no-value behavior", name, parentFlag.Name)
			})

			failed := command.Flags().Lookup("failed")
			Expect(failed).NotTo(BeNil())
			Expect(failed.NoOptDefVal).To(Equal(failedAutoSentinel))
		},
		Entry("go", "go"),
		Entry("ginkgo", "ginkgo"),
	)

	It("applies values and explicit skip-hooks state from the invoked command", func() {
		GinkgoT().Setenv("CI", "")
		command := &cobra.Command{Use: "go"}
		command.Flags().Bool("skip-hooks", false, "")
		Expect(command.Flags().Parse([]string{"--skip-hooks=false"})).To(Succeed())

		flags := testCommandFlags{
			AutoStop:    time.Minute,
			IdleTimeout: 2 * time.Minute,
			Timeout:     3 * time.Minute,
			LintTimeout: 4 * time.Minute,
			TestTimeout: 5 * time.Minute,
			Detach:      true,
		}
		opts, detach, err := finalizeTestCommandOptions(command, testrunner.RunOptions{}, flags)
		Expect(err).NotTo(HaveOccurred())
		Expect(detach).To(BeTrue())
		Expect(opts.AutoStop).To(Equal(time.Minute))
		Expect(opts.IdleTimeout).To(Equal(2 * time.Minute))
		Expect(opts.Timeout).To(Equal(3 * time.Minute))
		Expect(opts.LintTimeout).To(Equal(4 * time.Minute))
		Expect(opts.TestTimeout).To(Equal(5 * time.Minute))
		Expect(opts.SkipHooks).To(BeFalse())
	})

	It("uses the local skip-hooks default when the invoked command did not set it", func() {
		GinkgoT().Setenv("CI", "")
		command := &cobra.Command{Use: "go"}
		command.Flags().Bool("skip-hooks", false, "")

		opts, _, err := finalizeTestCommandOptions(command, testrunner.RunOptions{}, testCommandFlags{})
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.SkipHooks).To(BeTrue())
	})

	It("persists tags in snapshot metadata", func() {
		Expect(snapshotArgs(testrunner.RunOptions{Tags: []string{"integration", "postgres"}})).To(
			HaveKeyWithValue("tags", []string{"integration", "postgres"}),
		)
	})

	It("serializes project action tags as repeated flags", func() {
		args, err := (projectActionCommandProvider{}).Args("test", map[string]any{
			"tags": []any{"integration", "postgres"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(ConsistOf("--tags=integration", "--tags=postgres"))
	})
})
