package main

import (
	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("JSON lint command", func() {
	DescribeTable("binds the schema flag around the linter subcommand", func(args []string) {
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
		add("json", add("lint", root))
		root.SetArgs(args)
		Expect(root.Execute()).To(Succeed())
		Expect(captured.Schema).To(Equal("schema.json"))
		Expect(captured.Files).To(Equal([]string{"data.yaml"}))
		Expect(captured.OptionFields).To(HaveKeyWithValue("schema", true))
	},
		Entry("before subcommand", []string{"lint", "--schema", "schema.json", "json", "data.yaml"}),
		Entry("after subcommand", []string{"lint", "json", "data.yaml", "--schema", "schema.json"}),
	)

	It("publishes schema selection in the fixture command contract", func() {
		lint, err := schemaCommand("lint")
		Expect(err).NotTo(HaveOccurred())
		command, _, err := lint.Find([]string{"json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(command.Name()).To(Equal("json"))
		Expect(command.Flags().Lookup("schema")).NotTo(BeNil())
		doc, err := fixturesSchemaDocument()
		Expect(err).NotTo(HaveOccurred())
		fences := doc["fences"].(map[string]any)
		lintSchema := fences["lint"].(map[string]any)["schema"].(fixtureJSONSchema)
		Expect(lintSchema["properties"]).To(HaveKey("schema"))
	})
})
