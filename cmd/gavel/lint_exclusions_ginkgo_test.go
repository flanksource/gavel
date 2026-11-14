package main

import (
	"context"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/lint"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Shared rule exclusions", func() {
	DescribeTable("binds shared options before and after linter subcommands", func(args []string) {
		root := &cobra.Command{Use: "gavel"}
		var captured LintOptions
		run := func(opts LintOptions) (any, error) { captured = opts; return nil, nil }
		parent := clicky.AddNamedCommand("lint", root, LintOptions{}, run)
		clicky.AddNamedCommand("uir", parent, LintOptions{}, run)
		root.SetArgs(args)
		Expect(root.Execute()).To(Succeed())
		Expect(captured.Exclude).To(Equal([]string{"uir:small-*", "other", "third"}))
		Expect(captured.Ignore).To(Equal([]string{"generated/**"}))
	},
		Entry("parent", []string{"lint", "--exclude", "uir:small-*,other", "--exclude", "third", "--ignore", "generated/**"}),
		Entry("before child", []string{"lint", "--exclude", "uir:small-*,other", "--exclude", "third", "--ignore", "generated/**", "uir"}),
		Entry("after child", []string{"lint", "uir", "--exclude", "uir:small-*,other", "--exclude", "third", "--ignore", "generated/**"}),
		Entry("around child", []string{"lint", "--exclude", "uir:small-*,other", "uir", "--exclude", "third", "--ignore", "generated/**"}),
	)
	It("binds exclusions for programmatic command requests", func() {
		parent := &cobra.Command{Use: "gavel"}
		var captured LintOptions
		cmd := clicky.AddNamedCommand("lint", parent, LintOptions{}, func(opts LintOptions) (any, error) { captured = opts; return nil, nil })
		_, err := clicky.GetDataFunc(cmd)(map[string]string{"exclude": "uir:small-*"}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(captured.Exclude).To(Equal([]string{"uir:small-*"}))
	})
	It("validates programmatic prompt exclusions before calling a provider", func() {
		_, err := lintprompt.Run(context.Background(), lintprompt.Options{Exclude: []string{"missing:rule"}})
		Expect(err).To(MatchError(ContainSubstring("missing:rule")))
	})
	It("retains exclusions when a UI rerun constructs fresh options", func() {
		root := GinkgoT().TempDir()
		_, err := executeLintRerun(LintOptions{WorkDir: root, Exclude: []string{"missing:rule"}}, testui.RerunRequest{})
		Expect(err).To(MatchError(ContainSubstring("missing:rule")))
	})
	It("filters prompt report findings while keeping prompt execution errors", func() {
		exclusions, err := lint.ParseRuleExclusions([]string{"prompt:style-*"})
		Expect(err).NotTo(HaveOccurred())
		results := []*linters.LinterResult{{Linter: "prompt", Success: true, Violations: []models.Violation{{Source: "prompt", Rule: &models.Rule{Method: "style-spacing"}, Severity: models.SeverityError}}}, {Linter: "prompt", Error: "provider failed"}}
		exclusions.FilterResults(results)
		Expect(results[0].Violations).To(BeEmpty())
		Expect(linters.HasFailures(results)).To(BeTrue())
	})
})
