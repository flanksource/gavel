package jira

import (
	"encoding/json"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira field mapping", func() {
	It("keeps IDs, names and values for duplicate names and unknown fields", func() {
		fields := map[string]json.RawMessage{
			"customfield_1": json.RawMessage(`"first"`),
			"customfield_2": json.RawMessage(`{"name":"second"}`),
			"customfield_3": json.RawMessage(`[1,2]`),
		}
		catalog := map[string]FieldDefinition{
			"customfield_1": {ID: "customfield_1", Name: "Team"},
			"customfield_2": {ID: "customfield_2", Name: "Team"},
		}
		mapped, err := MapFields(fields, catalog)
		Expect(err).NotTo(HaveOccurred())
		Expect(mapped).To(Equal(map[string]externalissues.Field{
			"customfield_1": {ID: "customfield_1", Name: "Team", Value: json.RawMessage(`"first"`)},
			"customfield_2": {ID: "customfield_2", Name: "Team", Value: json.RawMessage(`{"name":"second"}`)},
			"customfield_3": {ID: "customfield_3", Value: json.RawMessage(`[1,2]`)},
		}))
	})

	It("keeps returned list fields in the structured result", func() {
		data := json.RawMessage(`{"issues":[{"key":"EXAMPLE-1","summary":"Example","status":{"name":"Open"},"customfield_1":{"name":"Platform"},"description":{"version":1,"type":"doc","content":[]}}]}`)
		catalogs := map[string]map[string]FieldDefinition{"EXAMPLE-1": {"customfield_1": {ID: "customfield_1", Name: "Team"}}}
		rows, err := ParseList(data, ListOptions{Catalogs: catalogs})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Fields["customfield_1"].Name).To(Equal("Team"))
		Expect(string(rows[0].Fields["customfield_1"].Value)).To(MatchJSON(`{"name":"Platform"}`))
		encoded, err := json.Marshal(rows[0])
		Expect(err).NotTo(HaveOccurred())
		var result map[string]json.RawMessage
		Expect(json.Unmarshal(encoded, &result)).To(Succeed())
		Expect(result).To(HaveKey("description"))
	})

	It("keeps the same ID when only one duplicate has a value", func() {
		fields := map[string]json.RawMessage{"customfield_1": json.RawMessage(`"first"`)}
		catalog := map[string]FieldDefinition{
			"customfield_1": {ID: "customfield_1", Name: "Team"},
			"customfield_2": {ID: "customfield_2", Name: "Team"},
		}
		mapped, err := MapFields(fields, catalog)
		Expect(err).NotTo(HaveOccurred())
		Expect(mapped).To(HaveKey("customfield_1"))
	})

	It("renders a named object without discarding its other properties", func() {
		value, err := renderFieldValue(json.RawMessage(`{"name":"Platform","id":"team-1","nested":{"detail":"important"}}`))
		Expect(err).NotTo(HaveOccurred())
		markdown := value.PrettyFull().Markdown()
		Expect(markdown).To(ContainSubstring("Platform"))
		Expect(markdown).To(ContainSubstring("team-1"))
		Expect(markdown).To(ContainSubstring("important"))
	})

	It("keeps both fields when a name matches a disambiguated key", func() {
		fields := map[string]json.RawMessage{"field1": json.RawMessage(`1`), "field2": json.RawMessage(`2`), "field3": json.RawMessage(`3`)}
		catalog := map[string]FieldDefinition{"field1": {ID: "field1", Name: "Team"}, "field2": {ID: "field2", Name: "Team"}, "field3": {ID: "field3", Name: "Team [field1]"}}
		mapped, err := MapFields(fields, catalog)
		Expect(err).NotTo(HaveOccurred())
		Expect(mapped).To(HaveLen(3))
		Expect(mapped["field1"].Name).To(Equal("Team"))
		Expect(mapped["field3"].Name).To(Equal("Team [field1]"))
	})
})
