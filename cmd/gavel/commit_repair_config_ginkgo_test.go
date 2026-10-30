package main

import (
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Commit repair configuration", func() {
	DescribeTable("lets an explicit CLI choice override the configured trigger", func(flag string, want bool) {
		configured := !want
		actual, err := commitRepairConfig(CommitOptions{AIFix: flag, AIFixMaxIterations: 2}, verify.CommitLintConfig{AIFix: &configured})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.AIFix).To(HaveValue(Equal(want)))
		Expect(actual.AIFixMaxIterations).To(HaveValue(Equal(2)))
	}, Entry("enabled", "true", true), Entry("disabled", "false", false))
	It("preserves the default interactive trigger when unset", func() {
		actual, err := commitRepairConfig(CommitOptions{}, verify.CommitLintConfig{})
		Expect(err).NotTo(HaveOccurred())
		Expect(actual.AIFix).To(BeNil())
	})
	It("accepts a bare --ai-fix flag without starting a commit", func() {
		var actual CommitOptions
		root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
		registerCommitCommand(root, func(options CommitOptions) (any, error) { actual = options; return nil, nil })
		root.SetArgs([]string{"commit", "--ai-fix", "--ai-fix-max-iterations=2"})
		Expect(root.Execute()).To(Succeed())
		Expect(actual.AIFix).To(Equal("true"))
		Expect(actual.AIFixMaxIterations).To(Equal(2))
	})
	It("blocks the implicit -y repair when explicitly disabled", func() {
		disabled := false
		result := &commitpkg.Result{Lint: &commitpkg.LintGateResult{Violations: 1, Repair: verify.CommitLintConfig{AIFix: &disabled}}}
		Expect(handleCommitLintFindings(GinkgoT().TempDir(), result, true)).To(Equal(lintFindingsBlocked))
	})
	DescribeTable("rejects invalid repair configuration", func(options CommitOptions) {
		_, err := commitRepairConfig(options, verify.CommitLintConfig{})
		Expect(err).To(HaveOccurred())
	}, Entry("invalid switch", CommitOptions{AIFix: "sometimes"}), Entry("negative iterations", CommitOptions{AIFixMaxIterations: -1}))
})
