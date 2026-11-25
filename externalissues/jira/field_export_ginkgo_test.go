package jira

import (
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("named field values and field tables", func() {
	It("exports direct values under normalized names without native payloads or schema wrappers", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Readable export","customfield_1":false,"customfield_2":0,"customfield_unknown":null,"comments":[{"id":"comment-1","body":"Ready"}],"attachment":[{"filename":"report.pdf","content":"https://example.test/report.pdf"}]}`), ParseOptions{Catalog: map[string]FieldDefinition{
			"customfield_1": {Name: "Team", Schema: json.RawMessage(`{"type":"boolean"}`)},
			"customfield_2": {Name: "Team"},
		}})
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		var decoded struct {
			Fields map[string]json.RawMessage `json:"fields"`
		}
		Expect(json.Unmarshal([]byte(output), &decoded)).To(Succeed())
		Expect(decoded.Fields).To(Equal(map[string]json.RawMessage{
			"team_1": json.RawMessage(`false`), "team_2": json.RawMessage(`0`), "customfield_unknown": json.RawMessage(`null`),
		}))
		Expect(output).NotTo(MatchRegexp(`"(?:native|schema|source|scheme)"\s*:`))
		var restored Issue
		Expect(json.Unmarshal([]byte(output), &restored)).To(Succeed())
		roundTrip, err := json.Marshal(restored)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(roundTrip)).To(MatchJSON(output))
	})

	It("does not overwrite normalized name collisions or numbered field names", func() {
		issue := Issue{Issue: externalissues.Issue{Fields: map[string]externalissues.Field{
			"one":   {Name: "Team", Value: json.RawMessage(`"first"`)},
			"two":   {Name: "Team", Value: json.RawMessage(`"second"`)},
			"three": {Name: "Team_1", Value: json.RawMessage(`"reserved"`)},
			"four":  {Name: "A/B", Value: json.RawMessage(`"slash"`)},
			"five":  {Name: "A B", Value: json.RawMessage(`"space"`)},
		}}}
		output, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		var decoded struct {
			Fields map[string]string `json:"fields"`
		}
		Expect(json.Unmarshal(output, &decoded)).To(Succeed())
		Expect(decoded.Fields).To(Equal(map[string]string{"team_2": "first", "team_3": "second", "team_1": "reserved", "a_b_1": "space", "a_b_2": "slash"}))
	})

	It("renders populated fields as a two-column table with rich values", func() {
		issue := Issue{Issue: externalissues.Issue{Fields: map[string]externalissues.Field{
			"votes": {Name: "Votes", Value: json.RawMessage(`{"self":"https://example.test/votes","votes":0,"hasVoted":false}`)},
			"note":  {Name: "Notes", Format: "markdown", Value: json.RawMessage(`"First | row\n\nSecond with ` + "`code`" + `"`)},
			"empty": {Name: "Empty", Value: json.RawMessage(`[]`)},
		}}}
		for _, full := range []bool{false, true} {
			fields, err := issue.renderFields(RenderOptions{Full: full})
			Expect(err).NotTo(HaveOccurred())
			Expect(fields.(api.TextTable).Rows).To(HaveLen(2))
			Expect(fields.Markdown()).To(SatisfyAll(ContainSubstring("| FIELD | VALUE |"), ContainSubstring(`First \| row`), ContainSubstring("<br>"), ContainSubstring("[0](https://example.test/votes)")))
			Expect(fields.HTML()).To(SatisfyAll(ContainSubstring("<table"), ContainSubstring("<code>code</code>"), ContainSubstring(`<a href="https://example.test/votes">0</a>`)))
		}
	})
})
