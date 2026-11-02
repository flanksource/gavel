package lint

import (
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/oxlint"
	"github.com/flanksource/gavel/linters/ruff"
	"github.com/flanksource/gavel/linters/uir"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Native rule exclusions", func() {
	It("disables golangci rules even when native arguments use enable-only", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		_, args, err := linters.PrepareCommand(golangci.NewGolangciLint(root), linters.RunOptions{WorkDir: root, NativeRuleExclusions: []string{"errcheck"}, Config: &models.LinterConfig{Args: []string{"--enable-only", "errcheck,gocritic"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(ContainElement("--enable-only=gocritic"))
		Expect(args).NotTo(ContainElement("errcheck,gocritic"))
	})
	DescribeTable("translates only supported selectors and retains configured arguments", func(factory func(string) linters.Linter, patterns, expected []string) {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		adapter := factory(root)
		provider, ok := adapter.(linters.RuleExcluder)
		Expect(ok).To(BeTrue())
		args, err := provider.RuleExclusionArgs(patterns)
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(Equal(expected))
		_, command, err := linters.PrepareCommand(adapter, linters.RunOptions{WorkDir: root, NativeRuleExclusions: patterns, ExtraArgs: []string{"--configured-value"}, Config: &models.LinterConfig{Args: []string{"--configured-rule"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(command).To(ContainElements(append([]string{"--configured-value", "--configured-rule"}, expected...)))
	},
		Entry("UIR exact IDs", func(root string) linters.Linter { return uir.New(root) }, []string{"function-lines", "function-*"}, []string{"--exclude", "function-lines"}),
		Entry("golangci exact IDs", func(root string) linters.Linter { return golangci.NewGolangciLint(root) }, []string{"errcheck", "*"}, []string{"--disable", "errcheck"}),
		Entry("Ruff exact codes and prefixes", func(root string) linters.Linter { return ruff.NewRuff(root) }, []string{"F401", "E5*", "*401", "E501"}, []string{"--ignore", "F401,E5,E501"}),
		Entry("ESLint JSON rule overrides", func(root string) linters.Linter { return eslint.NewESLint(root) }, []string{"no-debugger", "@scope/plugin/rule", "*"}, []string{"--rule", `{"@scope/plugin/rule":"off","no-debugger":"off"}`}),
		Entry("Oxlint rule IDs without category shortcuts", func(root string) linters.Linter { return oxlint.NewOxlint(root) }, []string{"eslint/no-debugger", "all", "correctness", "react/*"}, []string{"--allow", "eslint/no-debugger"}),
	)
})
