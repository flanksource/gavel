package main

import (
	"strings"

	"github.com/flanksource/clicky/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("grouped command help", func() {
	DescribeTable("preserves grouping inside custom help guides", func(path string, render func(*cobra.Command) api.Text) {
		command, _, err := rootCmd.Find(strings.Fields(path))
		Expect(err).NotTo(HaveOccurred())
		output := render(command).String()
		for _, heading := range []string{"Command flags", "Formatting", "Logging", "Global flags"} {
			Expect(output).To(ContainSubstring(heading), "%s: missing %s", path, heading)
		}
	},
		Entry("config", "config", configHelp),
		Entry("fixtures", "fixtures", fixturesHelp),
		Entry("TODO create", "todos create", todosCreateHelp),
	)
	DescribeTable("keeps command-specific and shared model flags in their sections", func(path, heading string, names []string) {
		command, _, err := rootCmd.Find(strings.Fields(path))
		Expect(err).NotTo(HaveOccurred())
		Expect(command.CommandPath()).To(Equal("gavel " + path))
		_, section, found := strings.Cut(command.UsageString(), "\n"+heading+":\n")
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
		Entry("formatting", "todos run", "Formatting", []string{"--json", "--format", "--no-color"}),
	)
})
