package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/testrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

// timeoutCommand mirrors the duration flags a real test subcommand carries, so
// finalizeTestCommandOptions can tell an explicit --timeout from a default one.
func timeoutCommand() *cobra.Command {
	command := &cobra.Command{Use: "go"}
	command.Flags().Bool("skip-hooks", false, "")
	command.Flags().Duration("timeout", 10*time.Minute, "")
	command.Flags().Duration("lint-timeout", 5*time.Minute, "")
	command.Flags().Duration("test-timeout", 5*time.Minute, "")
	return command
}

// writeGavelConfig points --cwd at a repo carrying the given .gavel.yaml.
func writeGavelConfig(yaml string) {
	dir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(yaml), 0o600)).To(Succeed())
	previous := workingDir
	workingDir = dir
	DeferCleanup(func() { workingDir = previous })
}

// defaultFlags carries what cobra would hand over when nothing is passed: the
// flag defaults, which must not mask a configured value.
func defaultFlags() testCommandFlags {
	return testCommandFlags{
		Timeout:     10 * time.Minute,
		LintTimeout: 5 * time.Minute,
		TestTimeout: 5 * time.Minute,
	}
}

var _ = Describe("test timeouts from .gavel.yaml", func() {
	It("applies configured timeouts when no flag was passed", func() {
		writeGavelConfig("test:\n  timeout: 45m\n  testTimeout: 20m\n  lintTimeout: 8m\n")

		opts, _, err := finalizeTestCommandOptions(timeoutCommand(), testrunner.RunOptions{}, defaultFlags())
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Timeout).To(Equal(45 * time.Minute))
		Expect(opts.TestTimeout).To(Equal(20 * time.Minute))
		Expect(opts.LintTimeout).To(Equal(8 * time.Minute))
	})

	It("lets an explicit flag override the configured value", func() {
		writeGavelConfig("test:\n  testTimeout: 20m\n")

		command := timeoutCommand()
		Expect(command.Flags().Parse([]string{"--test-timeout=90s"})).To(Succeed())

		flags := defaultFlags()
		flags.TestTimeout = 90 * time.Second
		opts, _, err := finalizeTestCommandOptions(command, testrunner.RunOptions{}, flags)
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.TestTimeout).To(Equal(90 * time.Second))
	})

	It("leaves flag defaults alone when the config declares no timeouts", func() {
		writeGavelConfig("test:\n  outlineSummary: {}\n")

		opts, _, err := finalizeTestCommandOptions(timeoutCommand(), testrunner.RunOptions{}, defaultFlags())
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Timeout).To(Equal(10 * time.Minute))
		Expect(opts.TestTimeout).To(Equal(5 * time.Minute))
		Expect(opts.LintTimeout).To(Equal(5 * time.Minute))
	})

	It("reports a malformed duration instead of silently using the default", func() {
		writeGavelConfig("test:\n  testTimeout: twenty-minutes\n")

		_, _, err := finalizeTestCommandOptions(timeoutCommand(), testrunner.RunOptions{}, defaultFlags())
		Expect(err).To(MatchError(ContainSubstring("test.testTimeout")))
		Expect(err).To(MatchError(ContainSubstring("twenty-minutes")))
	})
})
