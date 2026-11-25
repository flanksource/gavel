package main

import (
	"bytes"
	"strings"

	"github.com/flanksource/clicky/api"
	flagmeta "github.com/flanksource/clicky/flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("grouped command help", func() {
	It("renders root flags only at the root and retains a nested override", func() {
		var output bytes.Buffer
		root := &cobra.Command{Use: "gavel"}
		root.SetOut(&output)
		root.PersistentFlags().String("format", "", "Global output format")
		flagmeta.SetCategory(root.PersistentFlags().Lookup("format"), flagmeta.CategoryFormatting)
		root.PersistentFlags().String("cwd", "", "Working directory")
		root.PersistentFlags().Bool("interactive", false, "Global interactive mode")
		setGavelUsage(root)

		group := &cobra.Command{Use: "group", Short: "Command group"}
		group.PersistentFlags().String("site", "", "Group site")
		root.AddCommand(group)
		leaf := &cobra.Command{Use: "leaf [flags] [path...]", Run: func(*cobra.Command, []string) {}}
		leaf.Flags().Bool("interactive", false, "Leaf interactive mode")
		group.AddCommand(leaf)
		root.InitDefaultHelpFlag()
		leaf.InitDefaultHelpFlag()

		Expect(root.Help()).To(Succeed())
		Expect(output.String()).To(And(
			ContainSubstring("Command group"), ContainSubstring("--help"),
			ContainSubstring("Formatting:"), ContainSubstring("--format"),
			ContainSubstring("--cwd"), ContainSubstring("Global interactive mode"),
			Not(ContainSubstring("--site")), Not(ContainSubstring(globalHelpHint)),
		))

		output.Reset()
		Expect(leaf.Help()).To(Succeed())
		Expect(output.String()).To(And(
			ContainSubstring("gavel group leaf [flags] [path...]"), ContainSubstring("--help"),
			ContainSubstring("Group site"), ContainSubstring("Leaf interactive mode"),
			Not(ContainSubstring("--format")), Not(ContainSubstring("--cwd")),
			Not(ContainSubstring("Global interactive mode")),
			HaveSuffix(globalHelpHint+"\n"),
		))
	})
	It("lists root flags and commands in root usage", func() {
		Expect(rootCmd.UsageString()).To(And(
			ContainSubstring("Available Commands:"), ContainSubstring("--cwd"),
			ContainSubstring("--format"), ContainSubstring("--no-progress"),
			Not(ContainSubstring("--work-dir")),
		))
	})
	It("keeps group-specific flags visible to nested commands", func() {
		command, _, err := rootCmd.Find(strings.Fields("external-issues jira list"))
		Expect(err).NotTo(HaveOccurred())
		Expect(command.UsageString()).To(And(
			ContainSubstring("--site"), ContainSubstring("--profile"),
			Not(ContainSubstring("--cwd")), Not(ContainSubstring("--no-progress")),
			ContainSubstring("Use gavel --help to show arguments available to all commands."),
		))
	})
	DescribeTable("preserves grouping inside custom help guides", func(path string, render func(*cobra.Command) api.Text) {
		command, _, err := rootCmd.Find(strings.Fields(path))
		Expect(err).NotTo(HaveOccurred())
		output := render(command).String()
		Expect(output).To(And(
			ContainSubstring("\nCommand flags\n"), Not(ContainSubstring("\nFormatting\n")),
			Not(ContainSubstring("\nLogging\n")), Not(ContainSubstring("\nGlobal flags\n")),
			ContainSubstring("Use gavel --help to show arguments available to all commands."),
		))
	},
		Entry("config", "config", configHelp),
		Entry("fixtures", "fixtures", fixturesHelp),
		Entry("TODO create", "todos create", todosCreateHelp),
	)
	DescribeTable("keeps command-specific and shared model flags in their sections", func(path, heading string, names []string) {
		command, _, err := rootCmd.Find(strings.Fields(path))
		Expect(err).NotTo(HaveOccurred())
		Expect(command.CommandPath()).To(Equal("gavel " + path))
		usage := command.UsageString()
		Expect(usage).To(And(Not(ContainSubstring("--no-progress")), Not(ContainSubstring("--cwd"))))
		_, section, found := strings.Cut(usage, "\n"+heading+":\n")
		Expect(found).To(BeTrue(), "%s: missing %s", path, heading)
		section, _, _ = strings.Cut(section, "\n\n")
		for _, name := range names {
			Expect(section).To(ContainSubstring(name), "%s: %s must contain %s", path, heading, name)
		}
	},
		Entry("commit selection", "commit", "Model selection", []string{"--model", "--group-model", "--effort"}),
		Entry("commit inputs", "commit", "Command flags", []string{"--message", "--stage"}),
		Entry("TODO selection", "todos run", "Model selection", []string{"--model", "--effort"}),
		Entry("TODO runtime", "todos run", "Model runtime", []string{"--max-budget", "--max-turns", "--preset"}),
		Entry("merge selection", "todos merge", "Model selection", []string{"--model", "--effort"}),
		Entry("PR selection", "pr create", "Model selection", []string{"--model", "--no-cache"}),
		Entry("manual AI flags", "git analyze", "Model selection", []string{"--ai-model", "--ai-no-cache"}),
		Entry("lint permissions", "lint", "Model permissions", []string{"--access", "--sandbox", "--permission-mode"}),
		Entry("local override", "commit", "Command flags", []string{"--interactive", "Open an interactive tree picker"}),
	)
})
