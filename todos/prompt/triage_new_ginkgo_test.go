package prompt

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("Triage new prompt", func() {
	It("resolves a separate plan-class backlog prompt under its own config key", func() {
		catalog, err := NewCatalog(verify.TodosConfig{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		def, err := catalog.Lookup("triage.new")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect([]any{def.Class, def.ConfigKey, def.Backlog}).To(gomega.Equal([]any{types.ModePlan, "triageNew", true}))
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("parser", "Parser crashes")}, Options{Prompt: "triage.new", RawFixture: true})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect([]string{string(req.Model.Mode), req.Model.Name, string(req.Model.Effort)}).To(gomega.Equal([]string{string(api.ModeAPI), "gpt-6-luna", string(api.EffortHigh)}))
	})
})

var _ = Describe("EnvelopeSchemaJSON", func() {
	It("is the closed {success, event, comment} object with the step's events as the event enum", func() {
		events := []string{"kept", "duplicate"}
		schema, err := EnvelopeSchemaJSON(events)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var document map[string]any
		gomega.Expect(json.Unmarshal(schema, &document)).To(gomega.Succeed())
		gomega.Expect(document["required"]).To(gomega.ConsistOf("success", "event", "comment"))
		gomega.Expect(document["additionalProperties"]).To(gomega.BeFalse())
		properties := document["properties"].(map[string]any)
		gomega.Expect(properties).To(gomega.HaveLen(3))
		gomega.Expect(properties["event"].(map[string]any)["enum"]).To(gomega.Equal([]any{"kept", "duplicate"}))
		gomega.Expect(properties["comment"].(map[string]any)["minLength"]).To(gomega.BeEquivalentTo(1))
	})

	It("refuses a step without events", func() {
		_, err := EnvelopeSchemaJSON(nil)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
