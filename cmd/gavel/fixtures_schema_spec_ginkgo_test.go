package main

import (
	"encoding/json"

	"github.com/flanksource/gavel/fixtures"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("fixture document schema", func() {
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
