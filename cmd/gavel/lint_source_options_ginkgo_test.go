package main

import (
	"os"
	"path/filepath"

	"github.com/flanksource/clicky"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Lint source options", func() {
	DescribeTable("binds generated policy and repeated tags around linter subcommands", func(args []string) {
		root := &cobra.Command{Use: "gavel"}
		var captured LintOptions
		add := func(name string, parent *cobra.Command) *cobra.Command {
			var command *cobra.Command
			command = clicky.AddNamedCommand(name, parent, LintOptions{}, func(options LintOptions) (any, error) {
				var err error
				captured, err = splitLintPassThroughArgs(command, options)
				return nil, err
			})
			bindLintSourceFlags(command)
			return command
		}
		add("uir", add("lint", root))
		root.SetArgs(args)
		Expect(root.Execute()).To(Succeed())
		Expect(captured.Generated).To(HaveValue(BeFalse()))
		Expect(captured.Tags).To(Equal([]string{"integration", "postgres", "feature"}))
	},
		Entry("parent", []string{"lint", "--generated=false", "--tags=integration,postgres", "--tags=feature"}),
		Entry("before child", []string{"lint", "--generated=false", "--tags=integration,postgres", "--tags=feature", "uir"}),
		Entry("after child", []string{"lint", "uir", "--generated=false", "--tags=integration,postgres", "--tags=feature"}),
		Entry("around child", []string{"lint", "--tags=integration,postgres", "uir", "--generated=false", "--tags=feature"}),
	)

	DescribeTable("preserves unset and explicit inclusion", func(args []string, explicit bool) {
		command := &cobra.Command{Use: "lint"}
		bindLintSourceFlags(command)
		Expect(command.ParseFlags(args)).To(Succeed())
		options, err := splitLintPassThroughArgs(command, LintOptions{})
		Expect(err).NotTo(HaveOccurred())
		if explicit {
			Expect(options.Generated).To(HaveValue(BeTrue()))
		} else {
			Expect(options.Generated).To(BeNil())
		}
	}, Entry("omitted", []string{}, false), Entry("bare", []string{"--generated"}, true), Entry("true", []string{"--generated=true"}, true))

	It("retains generated policy and tags through UI reruns", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("PATH", root+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("UIR_TEST_REPORT", `[{"linter":"uir","success":true,"violations":[{"file":"sample_gen.go","severity":"error"}]}]`)
		GinkgoT().Setenv("UIR_TEST_EXIT", "1")
		script, err := os.ReadFile("../../linters/uir/testdata/uir.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "uir"), script, 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "sample_gen.go"), []byte("package sample\n"), 0o600)).To(Succeed())
		results, err := executeLintRerun(LintOptions{WorkDir: root, Generated: new(false), Tags: []string{"feature"}}, testui.RerunRequest{LintLinters: []string{"uir"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(ConsistOf(And(HaveField("Success", true), HaveField("Violations", BeEmpty()), HaveField("Args", ContainElement("--tags=feature")))))
	})
})
