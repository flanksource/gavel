package main

import (
	captaincli "github.com/flanksource/captain/pkg/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Lint prompt command", func() {
	It("binds change selection, reporting, ai-fix and fixer runtime flags before --", func() {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags([]string{"--changed", "--since", "main", "--changed-files", "a.go", "--ai-fix", "--ai-fix-max-iterations", "2", "--model", "agent:opus", "--dry-run", "--summary"})).To(Succeed())
		Expect(cmd.Flags().GetBool("ai-fix")).To(BeTrue())
		Expect(cmd.Flags().GetString("since")).To(Equal("main"))
		Expect(cmd.Flags().GetString("model")).To(Equal("agent:opus"))
		Expect(cmd.Args(cmd, []string{"review.prompt", "pkg", "cmd"})).To(Succeed())
		Expect(cmd.Args(cmd, nil)).To(Succeed())
	})

	DescribeTable("rejects prompt-only flags before --", func(flag string) {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags([]string{flag, "value"})).To(MatchError(ContainSubstring("unknown")))
	},
		Entry("inline prompt", "-p"),
		Entry("template variable", "-V"),
		Entry("preset", "--preset"),
	)

	DescribeTable("splits gavel flags from Captain prompt arguments at --", func(args []string, fixerModel, promptModel, file string, paths []string) {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags(args)).To(Succeed())
		positional, through, err := splitAtDash(cmd, cmd.Flags().Args())
		Expect(err).NotTo(HaveOccurred())
		prompt, promptArgs, err := captaincli.ParsePromptArgs(through)
		Expect(err).NotTo(HaveOccurred())
		gotFile, gotPaths := splitPromptArgs(append(positional, promptArgs...), prompt.Prompt != "")
		Expect(cmd.Flags().GetString("model")).To(Equal(fixerModel))
		Expect(prompt.Model).To(Equal(promptModel))
		Expect(gotFile).To(Equal(file))
		Expect(gotPaths).To(Equal(paths))
	},
		Entry("fixer and reviewer models with prompt and path after --",
			[]string{"--ai-fix", "--model", "agent:opus", "--", "--model", "agent:sonnet", "proposal-review", "report.mdx"},
			"agent:opus", "agent:sonnet", "proposal-review", []string{"report.mdx"}),
		Entry("prompt and paths before --",
			[]string{"review.prompt", "pkg", "--", "--model", "agent:sonnet"},
			"", "agent:sonnet", "review.prompt", []string{"pkg"}),
		Entry("inline prompt after -- takes every positional as a path",
			[]string{"pkg", "--", "-p", "Review the change", "cmd"},
			"", "", "", []string{"pkg", "cmd"}),
		Entry("no --", []string{"review.prompt", "pkg"}, "", "", "review.prompt", []string{"pkg"}),
	)

	It("fails on an unknown prompt flag after --", func() {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags([]string{"--", "--not-a-flag"})).To(Succeed())
		err := cmd.RunE(cmd, cmd.Flags().Args())
		Expect(err).To(MatchError(ContainSubstring("prompt arguments after --")))
	})

	It("binds --ui and --addr before --", func() {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags([]string{"--ui", "--addr", "localhost"})).To(Succeed())
		Expect(cmd.Flags().GetBool("ui")).To(BeTrue())
		Expect(cmd.Flags().GetString("addr")).To(Equal("localhost"))
	})

	DescribeTable("splits the prompt reference from the paths under review", func(args []string, inline bool, file string, paths []string) {
		gotFile, gotPaths := splitPromptArgs(args, inline)
		Expect(gotFile).To(Equal(file))
		Expect(gotPaths).To(Equal(paths))
	},
		Entry("prompt only", []string{"review.prompt"}, false, "review.prompt", []string{}),
		Entry("prompt and paths", []string{"review.prompt", "pkg", "cmd/x.go"}, false, "review.prompt", []string{"pkg", "cmd/x.go"}),
		Entry("inline prompt takes every arg as a path", []string{"pkg", "cmd/x.go"}, true, "", []string{"pkg", "cmd/x.go"}),
		Entry("no args", nil, false, "", nil),
	)
})
