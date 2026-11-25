package main

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("embedded project action schemas", func() {
	provider := projectActionCommandProvider{}
	DescribeTable("excludes AI settings outside the native runtime bar", func(name string) {
		definition, err := provider.Schema("commit")
		Expect(err).NotTo(HaveOccurred())
		Expect(definition.Schema["properties"]).NotTo(HaveKey(name))
		_, err = provider.Args("commit", map[string]any{"files": []any{"one.go"}, name: "override"})
		Expect(err).To(MatchError(ContainSubstring("unsupported commit option")))
	}, Entry("group model", "group-model"), Entry("fallbacks", "fallback"), Entry("temperature", "temperature"), Entry("cache", "no-cache"))
	DescribeTable("publishes v2 with only action options", func(action, positional string) {
		definition, err := provider.Schema(action)
		Expect(err).NotTo(HaveOccurred())
		Expect(definition.SchemaVersion).To(Equal(2))
		properties := definition.Schema["properties"].(map[string]any)
		Expect(properties).To(HaveKey(positional))
		for _, excluded := range []string{"cwd", "db-url", "format", "json", "max-concurrent", "ui", "work-dir"} {
			Expect(properties).NotTo(HaveKey(excluded))
		}
	}, Entry("commit", "commit", "files"), Entry("lint", "lint", "files"), Entry("test", "test", "paths"))

	It("keeps an omitted repair limit out of the real commit binder", func() {
		args, err := provider.Args("commit", map[string]any{"files": []any{"one.go"}, "message": "fix: selected"})
		Expect(err).NotTo(HaveOccurred())
		var captured CommitOptions
		root := &cobra.Command{Use: "probe", SilenceErrors: true, SilenceUsage: true}
		registerCommitCommand(root, func(options CommitOptions) (any, error) { captured = options; return nil, nil })
		root.SetArgs(append([]string{"commit"}, args...))
		Expect(root.Execute()).To(Succeed())
		Expect(captured.AIFixMaxIterations).To(BeZero())
		Expect(captured.Args).To(Equal([]string{"one.go"}))
	})

	DescribeTable("rejects invalid options before execution", func(action string, options map[string]any, field string) {
		_, err := provider.Args(action, options)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(field))
	},
		Entry("zero repair limit", "commit", map[string]any{"files": []any{"one.go"}, "ai-fix-max-iterations": 0}, "ai-fix-max-iterations"),
		Entry("fractional repair limit", "commit", map[string]any{"files": []any{"one.go"}, "ai-fix-max-iterations": 1.5}, "ai-fix-max-iterations"),
		Entry("incompatible message", "commit", map[string]any{"files": []any{"one.go"}, "commit-all": true, "message": "fix: selected"}, "message"),
		Entry("summary bound", "lint", map[string]any{"summary-limit": 0}, "summary-limit"),
		Entry("unknown framework", "test", map[string]any{"framework": []any{"unknown-runner"}}, "framework"),
		Entry("invalid node count", "test", map[string]any{"nodes": -2}, "nodes"),
	)

	It("preserves valid automatic counts and explicit false", func() {
		args, err := provider.Args("test", map[string]any{"nodes": -1, "concurrency": 0, "recursive": false})
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(ConsistOf("--nodes=-1", "--concurrency=0", "--recursive=false"))
	})

	DescribeTable("covers every property in serialization order and matches Cobra types", func(action, positional string) {
		definition, err := provider.Schema(action)
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(definition.Schema)
		Expect(err).NotTo(HaveOccurred())
		var document struct {
			Properties map[string]struct {
				Type string `json:"type"`
			} `json:"properties"`
			Order []string `json:"x-order"`
		}
		Expect(json.Unmarshal(encoded, &document)).To(Succeed())
		Expect(document.Order).To(HaveLen(len(document.Properties)))
		command, err := schemaCommand(action)
		Expect(err).NotTo(HaveOccurred())
		for name, property := range document.Properties {
			Expect(document.Order).To(ContainElement(name))
			if name == positional {
				continue
			}
			flag := command.Flags().Lookup(name)
			Expect(flag).NotTo(BeNil(), "%s --%s", action, name)
			expected := map[string]string{"bool": "boolean", "int": "integer", "duration": "string", "string": "string", "stringSlice": "array", "stringArray": "array"}
			Expect(property.Type).To(Equal(expected[flag.Value.Type()]), "%s --%s", action, name)
		}
	}, Entry("commit", "commit", "files"), Entry("lint", "lint", "files"), Entry("test", "test", "paths"))
})
