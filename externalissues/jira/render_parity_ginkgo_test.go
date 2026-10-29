package jira

import (
	"encoding/json"
	"os"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("external issue semantic rendering", func() {
	var issue Issue
	BeforeEach(func() {
		data, err := os.ReadFile("testdata/rendering.json")
		Expect(err).NotTo(HaveOccurred())
		issue, err = ParseIssue(data, ParseOptions{Catalog: map[string]FieldDefinition{
			"customfield_team": {Name: "Team"}, "customfield_doc": {Name: "Instructions"},
			"customfield_null": {Name: "Unset"}, "customfield_false": {Name: "Enabled"},
		}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("renders Jira option objects as their value even in full output", func() {
		view, err := renderFieldValue(json.RawMessage(`{"id":"20714","self":"https://example.test/rest/api/3/customFieldOption/20714","value":"Not Received"}`))
		Expect(err).NotTo(HaveOccurred())
		for _, value := range []api.Textable{view.PrettyShort(), view.PrettyFull()} {
			Expect(value.String()).To(Equal("Not Received"))
			Expect(value.Markdown()).To(Equal("Not Received"))
			Expect(value.HTML()).To(Equal("Not Received"))
		}
	})

	It("preserves both values of a cascading Jira option", func() {
		view, err := renderFieldValue(json.RawMessage(`{"id":"parent","value":"Region","child":{"id":"child","value":"East"}}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(view.PrettyShort().String()).To(Equal("Region / East"))
		Expect(view.PrettyFull().String()).To(Equal("Region / East"))
	})

	It("omits nulls and empty arrays recursively in both human views while preserving exports", func() {
		issue.Fields = map[string]externalissues.Field{
			"null":   {Value: json.RawMessage(`null`)},
			"empty":  {Value: json.RawMessage(`[]`)},
			"nested": {Value: json.RawMessage(`{"missing":null,"empty":[],"items":[null,[],false,0],"zero":0,"disabled":false}`)},
		}
		for _, full := range []bool{false, true} {
			fields, err := issue.renderFields(RenderOptions{Full: full})
			Expect(err).NotTo(HaveOccurred())
			Expect(fields.(api.TextTable).Rows).To(HaveLen(1))
			for _, output := range []string{fields.String(), fields.Markdown(), fields.HTML(), fields.ANSI()} {
				Expect(output).To(SatisfyAll(Not(ContainSubstring("null")), Not(ContainSubstring("empty")), Not(ContainSubstring("missing")), Not(ContainSubstring("[]")), ContainSubstring("false"), ContainSubstring("0")))
			}
		}
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		var decoded Issue
		Expect(json.Unmarshal([]byte(output), &decoded)).To(Succeed())
		Expect(decoded.Fields).To(HaveLen(len(issue.Fields)))
		for id, field := range issue.Fields {
			Expect(string(decoded.Fields[id].Value)).To(MatchJSON(string(field.Value)))
		}
	})

	It("keeps compact values useful without expanding named object metadata", func() {
		output, err := issue.Render(RenderOptions{Full: false})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(SatisfyAll(
			ContainSubstring("[Platform](https://example.test/team)"),
			ContainSubstring(`customfield\_unknown`), ContainSubstring("retained"),
			ContainSubstring("false"), ContainSubstring("9007199254740993"),
			Not(ContainSubstring("team-internal-id")), Not(ContainSubstring("private comment body")),
		))
		Expect(output.HTML()).To(SatisfyAll(
			ContainSubstring("&lt;customers&gt;"), ContainSubstring(`<a href="https://example.test/team">Platform</a>`),
			Not(ContainSubstring("<customers>")), Not(ContainSubstring("\x1b")),
		))
	})

	It("renders linked people, rich fields, attachment tables and comment metadata in full", func() {
		output, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.HTML()).To(SatisfyAll(
			ContainSubstring(`<a href="https://example.test/people/alex">Alex</a>`),
			ContainSubstring("team-internal-id"), ContainSubstring("<code>run()</code>"),
			ContainSubstring("<table"), ContainSubstring("private comment body"),
			ContainSubstring("Updated"), Not(ContainSubstring("\x1b")),
		))
		Expect(output.Markdown()).To(SatisfyAll(ContainSubstring("| ID | FILE |"), ContainSubstring("[report.pdf](https://example.test/report.pdf)"), ContainSubstring("**Updated**:")))
		Expect(output.String()).To(SatisfyAll(ContainSubstring("report.pdf"), ContainSubstring("private comment body")))
	})

	It("includes every attachment as a filename link in compact output", func() {
		issue.Attachments = append(issue.Attachments, externalissues.Attachment{Name: "notes [2].txt", URL: "https://example.test/notes(2).txt"})
		output, err := issue.Render(RenderOptions{Full: false})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(SatisfyAll(
			ContainSubstring("**Attachments**:\n\n- [report.pdf](https://example.test/report.pdf)\n- [notes \\[2\\].txt](https://example.test/notes%282%29.txt)"),
			Not(ContainSubstring("| ID | FILE |")),
		))
		Expect(output.HTML()).To(SatisfyAll(
			ContainSubstring(`<a href="https://example.test/report.pdf">report.pdf</a>`),
			ContainSubstring(`<a href="https://example.test/notes(2).txt">notes [2].txt</a>`),
		))
		Expect(output.String()).To(SatisfyAll(ContainSubstring("report.pdf"), ContainSubstring("notes [2].txt")))
	})

	It("does not mutate normalized JSON when selecting a presentation", func() {
		before, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		for _, full := range []bool{false, true} {
			issue.Full = full
			for _, format := range []string{"markdown", "html", "pretty"} {
				_, err := clicky.Format(issue, clicky.FormatOptions{Format: format})
				Expect(err).NotTo(HaveOccurred())
			}
		}
		after, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(MatchJSON(before))
	})
})
