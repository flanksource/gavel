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
	It("resolves a separate plan-class prompt and required output", func() {
		catalog, err := NewCatalog(verify.TodosConfig{})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		def, err := catalog.Lookup("triage.new")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(def.Class).To(gomega.Equal(types.ModePlan))
		gomega.Expect(def.Envelope).To(gomega.Equal(EnvelopeTriageNew))
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("parser", "Parser crashes")}, Options{Prompt: "triage.new", Envelope: EnvelopeTriageNew})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect([]string{string(req.Model.Mode), req.Model.Name, string(req.Model.Effort)}).To(gomega.Equal([]string{string(api.ModeAPI), "gpt-6-luna", string(api.EffortHigh)}))
		schema, err := EnvelopeSchemaJSON(EnvelopeTriageNew)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		var document struct {
			Defs map[string]struct {
				Required   []string
				Properties map[string]struct {
					MinItems  int
					MinLength int
				}
			} `json:"$defs"`
		}
		gomega.Expect(json.Unmarshal(schema, &document)).To(gomega.Succeed())
		gomega.Expect(document.Defs["TriageNewEnvelope"].Required).To(gomega.ContainElements("title", "labels", "action"))
		gomega.Expect(document.Defs["TriageNewEnvelope"].Properties["labels"].MinItems).To(gomega.Equal(1))
		gomega.Expect(document.Defs["TriageNewEnvelope"].Properties["title"].MinLength).To(gomega.Equal(1))
	})
})
