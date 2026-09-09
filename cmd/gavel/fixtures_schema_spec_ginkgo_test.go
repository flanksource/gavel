package main

import (
	"encoding/json"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

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
