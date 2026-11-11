package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/fixtures"
	jsonlint "github.com/flanksource/gavel/linters/json"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"github.com/spf13/pflag"
)

var _ = g.Describe("builtin fixture schemas", func() {
	g.It("validate fixture frontmatter and documents that declare no schema, and reach the editor", func() {
		root := g.GinkgoT().TempDir()
		o.Expect(os.Mkdir(filepath.Join(root, ".git"), 0o700)).To(o.Succeed())
		markdown := filepath.Join(root, "checks.fixture.md")
		o.Expect(os.WriteFile(markdown, []byte("---\ntimeout: 30s\nenv: [not, a, map]\n---\n# Checks\n"), 0o600)).To(o.Succeed())
		document := filepath.Join(root, "plain.md")
		o.Expect(os.WriteFile(document, []byte("---\nenv: [fine]\n---\n"), 0o600)).To(o.Succeed())

		frontmatter, err := jsonlint.PathSchema(markdown)
		o.Expect(err).NotTo(o.HaveOccurred())
		plain, err := jsonlint.PathSchema(document)
		o.Expect(err).NotTo(o.HaveOccurred())
		adapter := jsonlint.New(root)
		adapter.Files = []string{markdown, document}
		violations, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		o.Expect(err).NotTo(o.HaveOccurred())
		editor, err := jsonlint.EditorSchema(context.Background(), []byte("timeout: 30s\n"), markdown, "")
		o.Expect(err).NotTo(o.HaveOccurred())

		o.Expect(frontmatter).To(o.Equal("gavel:fixture-frontmatter"))
		o.Expect(plain).To(o.BeEmpty())
		o.Expect(violations).To(o.ConsistOf(o.SatisfyAll(o.HaveField("File", markdown), o.HaveField("Line", 3))))
		o.Expect(editor).To(o.HaveKey("$defs"))
	})
})

var _ = g.Describe("fixture document schema", func() {
	g.It("retains native CLI flag groups for global reference fields", func() {
		for _, name := range []string{"test", "lint"} {
			cmd, err := schemaCommand(name)
			o.Expect(err).NotTo(o.HaveOccurred())
			var original []string
			cmd.Flags().VisitAll(func(flag *pflag.Flag) { original = append(original, flag.Name) })
			schema := commandSchema(cmd, "", commandSchemaOptions{})
			properties := schema["properties"].(map[string]any)
			var updated []string
			cmd.Flags().VisitAll(func(flag *pflag.Flag) { updated = append(updated, flag.Name) })
			o.Expect(updated).To(o.Equal(original))
			for flag, group := range map[string]string{"format": "Formatting", "json": "Formatting", "loglevel": "Logging", "max-concurrent": "Tasks", "cwd": "Global flags", "db-url": "Global flags"} {
				o.Expect(properties).To(o.HaveKey(flag), name+" --"+flag)
				o.Expect(properties[flag].(fixtureSchemaProperty)["x-cli-group"]).To(o.Equal(group), name+" --"+flag)
			}
		}
	})

	g.It("composes a step oneOf over every registered kind, with the CLI option sets for test and lint", func() {
		doc, err := fixturesSchemaDocument()
		o.Expect(err).NotTo(o.HaveOccurred())
		data, err := json.Marshal(doc["document"])
		o.Expect(err).NotTo(o.HaveOccurred())
		var schema map[string]any
		o.Expect(json.Unmarshal(data, &schema)).To(o.Succeed())

		steps := schema["properties"].(map[string]any)["steps"].(map[string]any)
		branches := map[string]map[string]any{}
		for _, raw := range steps["items"].(map[string]any)["oneOf"].([]any) {
			branch := raw.(map[string]any)
			branches[branch["title"].(string)] = branch["properties"].(map[string]any)
		}
		for _, kind := range fixtures.DefaultKinds.Names() {
			o.Expect(branches).To(o.HaveKey(kind))
			o.Expect(branches[kind]["kind"]).To(o.Equal(map[string]any{"const": kind}))
			o.Expect(branches[kind]).To(o.HaveKey("wait"))
			o.Expect(branches[kind]).To(o.HaveKey("export"))
		}
		o.Expect(branches).To(o.HaveKey("services"))
		o.Expect(branches["test"]).To(o.HaveKey("paths"))
		o.Expect(branches["lint"]).To(o.HaveKey("linters"))
		o.Expect(branches["exec"]).To(o.HaveKey("script"))
		o.Expect(schema["properties"].(map[string]any)["kind"]).To(o.Equal(map[string]any{"const": "Fixture"}))
	})
})

var _ = g.Describe("fixture frontmatter params schema", func() {
	g.It("offers params: as an object JSON Schema the run is started with", func() {
		data, err := json.Marshal(fixtureFrontmatterSchema())
		o.Expect(err).NotTo(o.HaveOccurred())
		var schema map[string]any
		o.Expect(json.Unmarshal(data, &schema)).To(o.Succeed())
		params := schema["properties"].(map[string]any)["params"].(map[string]any)
		o.Expect(params["type"]).To(o.Equal("object"))
		o.Expect(params["properties"].(map[string]any)["type"]).To(o.Equal(map[string]any{"const": "object"}))
		o.Expect(params["required"]).To(o.Equal([]any{"type", "properties"}))
	})
})

var _ = g.Describe("fixture grader runtime schema", func() {
	g.It("exposes the native runtime fields under ai.spec alongside the flat options", func() {
		data, err := json.Marshal(fixtureFrontmatterSchema())
		o.Expect(err).NotTo(o.HaveOccurred())
		var schema map[string]any
		o.Expect(json.Unmarshal(data, &schema)).To(o.Succeed())
		ai := schema["properties"].(map[string]any)["ai"].(map[string]any)["properties"].(map[string]any)
		o.Expect(ai).To(o.HaveKey("spec"))
		properties := ai["spec"].(map[string]any)["properties"].(map[string]any)
		for _, field := range []string{"model", "mode", "effort", "fallbacks", "budget", "permissions", "memory", "toolPreferences", "setup", "sandbox"} {
			o.Expect(properties).To(o.HaveKey(field))
		}
		o.Expect(ai).To(o.HaveKey("model"))
		o.Expect(ai).To(o.HaveKey("cacheTTL"))
		o.Expect(ai).To(o.HaveKey("criteriaSection"))
	})
})
