package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Lint prompt command", func() {
	It("binds Captain prompt inputs and lint reporting without flag collisions", func() {
		parent := &cobra.Command{Use: "lint"}
		cmd := registerLintPromptCommand(parent)
		Expect(cmd.ParseFlags([]string{"--preset", "review", "--var", "change=-", "--vars", `{"language":"go"}`, "--dry-run", "--summary", "--no-presets", "-p", "@review.prompt"})).To(Succeed())
		Expect(cmd.Flags().GetString("prompt")).To(Equal("@review.prompt"))
		Expect(cmd.Flags().Lookup("required-presets")).To(BeNil())
		Expect(cmd.Args(cmd, []string{"review.prompt", "pkg", "cmd"})).To(Succeed())
		Expect(cmd.Args(cmd, nil)).To(Succeed())
	})

	It("binds change selection and ai-fix flags", func() {
		cmd := registerLintPromptCommand(&cobra.Command{Use: "lint"})
		Expect(cmd.ParseFlags([]string{"--changed", "--since", "main", "--changed-files", "a.go", "--ai-fix", "--ai-fix-max-iterations", "2"})).To(Succeed())
		Expect(cmd.Flags().GetBool("ai-fix")).To(BeTrue())
		Expect(cmd.Flags().GetString("since")).To(Equal("main"))
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
