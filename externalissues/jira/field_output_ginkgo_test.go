package jira

import (
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("issue field output", func() {
	It("uses a table for known and unknown values without dropping colliding labels", func() {
		issue := Issue{Issue: externalissues.Issue{Fields: map[string]externalissues.Field{
			"field1":              {ID: "field1", Name: "Team", Value: json.RawMessage(`"Platform"`)},
			"field2":              {ID: "field2", Name: "Team", Value: json.RawMessage(`false`)},
			"field3":              {ID: "field3", Name: "Team [field1]", Value: json.RawMessage(`0`)},
			"customfield_unknown": {ID: "customfield_unknown", Value: json.RawMessage(`{"nested":"retained"}`)},
			"customfield_null":    {ID: "customfield_null", Value: json.RawMessage(`null`)},
		}}}
		fields, err := issue.renderFields(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(fields.(api.TextTable).Rows).To(HaveLen(4))
		Expect(fields.Markdown()).To(ContainSubstring(`| Team \[field1\] \[field3\] | 0 |`))
		Expect(fields.String()).To(SatisfyAll(ContainSubstring("Platform"), ContainSubstring("false"), ContainSubstring("0"), ContainSubstring("customfield_unknown"), ContainSubstring("retained"), Not(ContainSubstring("customfield_null"))))
		compact, err := issue.renderFields(RenderOptions{Full: false})
		Expect(err).NotTo(HaveOccurred())
		Expect(compact.String()).To(SatisfyAll(ContainSubstring("customfield_unknown"), ContainSubstring("false"), Not(ContainSubstring("customfield_null"))))
	})

	It("converts named and unknown custom-field documents to Markdown in JSON and list output", func() {
		const doc = `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"run()","marks":[{"type":"code"}]}]}]}`
		raw := json.RawMessage(`{"key":"EXAMPLE-1","summary":"Documents","customfield_1":` + doc + `,"customfield_unknown":` + doc + `}`)
		catalog := map[string]FieldDefinition{"customfield_1": {ID: "customfield_1", Name: "Instructions"}}
		issue, err := ParseIssue(raw, ParseOptions{Catalog: catalog})
		Expect(err).NotTo(HaveOccurred())
		rows, err := ParseList(json.RawMessage(`{"issues":[`+string(raw)+`]}`), ListOptions{Catalogs: map[string]map[string]FieldDefinition{"EXAMPLE-1": catalog}})
		Expect(err).NotTo(HaveOccurred())
		for _, issue := range []Issue{issue, rows[0]} {
			output, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
			Expect(err).NotTo(HaveOccurred())
			var decoded struct {
				Fields map[string]json.RawMessage `json:"fields"`
			}
			Expect(json.Unmarshal([]byte(output), &decoded)).To(Succeed())
			for _, key := range []string{"instructions", "customfield_unknown"} {
				Expect(string(decoded.Fields[key])).To(MatchJSON("\"`run()`\""))
			}
			output, err = clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(SatisfyAll(ContainSubstring("Instructions"), ContainSubstring(`customfield\_unknown`), ContainSubstring("`run()`")))
		}
	})

	It("retains unsupported custom-field content without failing conversion", func() {
		fields, err := MapFields(map[string]json.RawMessage{"customfield_broken": json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"unknown"}]}`)}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fields["customfield_broken"].Value)).To(ContainSubstring("unknown"))
	})

	It("keeps literal placeholders visible after converting a custom-field document", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Placeholders","customfield_unknown":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"For <customers> use *literal*"}]}]}}`), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(issue.Fields["customfield_unknown"].Value)).To(MatchJSON(`"For \\<customers\\> use \\*literal\\*"`))
		rendered, err := issue.Render(RenderOptions{Full: false})
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.String()).To(ContainSubstring("For <customers> use *literal*"))
	})

	It("converts nested documents while preserving numeric precision", func() {
		const raw = `{"key":"EXAMPLE-1","summary":"Nested docs","customfield_unknown":{"count":9007199254740993,"items":[{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Instructions"}]}]},null,false]}}`
		issue, err := ParseIssue(json.RawMessage(raw), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(issue.Fields["customfield_unknown"].Value)).To(MatchJSON(`{"count":9007199254740993,"items":["Instructions",null,false]}`))
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("9007199254740993"), Not(ContainSubstring(`"native"`))))
	})
})
