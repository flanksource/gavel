package jira

import (
	"encoding/json"
	"strings"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("normalized issues", func() {
	It("exports typed core fields and direct extension values keyed by name", func() {
		issue, err := ParseIssue(json.RawMessage(`{"id":"10001","key":"EXAMPLE-42","summary":"Retry imports","url":"https://example.test/browse/EXAMPLE-42","status":{"id":"3","name":"In Progress","statusCategory":{"key":"indeterminate","name":"In Progress"}},"assignee":{"accountId":"user-1","displayName":"Alex"},"project":{"id":"10","key":"EXAMPLE","name":"Platform"},"issuetype":{"id":"2","name":"Story"},"priority":{"id":"1","name":"High"},"created":"2026-01-01T12:00:00.000+0000","labels":["imports"],"customfield_1":{"id":"team-1","name":"Platform"}}`), ParseOptions{Catalog: map[string]FieldDefinition{"customfield_1": {ID: "customfield_1", Name: "Team"}}})
		Expect(err).NotTo(HaveOccurred())
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		var result map[string]any
		Expect(json.Unmarshal(data, &result)).To(Succeed())
		Expect(result).To(SatisfyAll(
			HaveKeyWithValue("provider", "jira"),
			HaveKeyWithValue("id", "10001"),
			HaveKeyWithValue("title", "Retry imports"),
			HaveKeyWithValue("status", HaveKeyWithValue("category", "in_progress")),
			HaveKeyWithValue("assignee", HaveKeyWithValue("id", "user-1")),
			HaveKeyWithValue("created", "2026-01-01T12:00:00Z"),
			HaveKeyWithValue("fields", HaveKeyWithValue("team", HaveKeyWithValue("name", "Platform"))),
		))
	})

	It("preserves JSON access when rich content cannot be rendered", func() {
		raw := json.RawMessage(`{"key":"EXAMPLE-1","summary":"Unsupported document","description":{"type":"doc","version":1,"content":[{"type":"unknown"}]},"comments":[{"body":{"type":"doc","version":1,"content":[{"type":"unknown"}]}}]}`)
		issue, err := ParseIssue(raw, ParseOptions{Catalog: nil})
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Valid([]byte(output))).To(BeTrue())
		output, err = clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("unknown"))
		rows, err := ParseList(json.RawMessage(`{"issues":[`+string(raw)+`]}`), ListOptions{Catalogs: nil})
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(1))
	})

	It("round trips the model through JSON without cached presentation state", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Original","description":{"type":"doc","version":1,"content":[]}}`), ParseOptions{Catalog: nil})
		Expect(err).NotTo(HaveOccurred())
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		var restored Issue
		Expect(json.Unmarshal(data, &restored)).To(Succeed())
		roundTrip, err := json.Marshal(restored)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(roundTrip)).To(MatchJSON(data))
		output, err := clicky.Format(restored, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("Original"))
	})

	It("keeps compact output bounded and omits thread bodies and raw payloads", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Compact","status":{"name":"Open","id":"1"},"description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"`+strings.Repeat("description ", 200)+`"}]}]},"comments":[{"author":{"displayName":"Alex"},"body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"private thread body"}]}]}}],"attachment":[{"filename":"report.pdf","content":"https://example.test/report.pdf"}]}`), ParseOptions{Catalog: nil})
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("1 comment"), ContainSubstring("1 attachment"), Not(ContainSubstring("private thread body")), Not(ContainSubstring("<details>")), Not(ContainSubstring("```json"))))
		Expect(len([]rune(output))).To(BeNumerically("<", 500))
	})

	It("keeps extension identity unchanged after renames and duplicate discoveries", func() {
		fields := map[string]json.RawMessage{"customfield_1": json.RawMessage(`"Platform"`)}
		for _, catalog := range []map[string]FieldDefinition{
			{"customfield_1": {ID: "customfield_1", Name: "Team"}},
			{"customfield_1": {ID: "customfield_1", Name: "Owner"}, "customfield_2": {ID: "customfield_2", Name: "Owner"}},
		} {
			mapped, err := MapFields(fields, catalog)
			Expect(err).NotTo(HaveOccurred())
			Expect(mapped).To(HaveKey("customfield_1"))
		}
	})

	It("rejects invalid core timestamps with issue and field context", func() {
		_, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Invalid time","created":"yesterday"}`), ParseOptions{Catalog: nil})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("EXAMPLE-1"), ContainSubstring("created"))))
	})

	DescribeTable("exports unsupported content without invoking renderers", func(field string) {
		raw := json.RawMessage(`{"key":"EXAMPLE-1","summary":"Future document",` + field + `}`)
		issue, err := ParseIssue(raw, ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		issue.Full = true
		_, err = clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring(`"type":"unknown"`))
	},
		Entry("comment document", `"comments":[{"id":"comment-1","body":{"type":"doc","version":1,"content":[{"type":"unknown"}]}}]`),
	)

	It("renders changes to the typed model after a JSON round trip", func() {
		var issue Issue
		Expect(json.Unmarshal([]byte(`{"provider":"jira","key":"EXAMPLE-1","title":"Updated title","fields":{}}`), &issue)).To(Succeed())
		issue.Title = "Current title"
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("Current title"), Not(ContainSubstring("Updated title"))))
	})

	It("exports comments independently of their rich-text renderer", func() {
		comments, err := ParseComments(json.RawMessage(`{"comments":[{"id":"comment-1","author":{"accountId":"user-1","displayName":"Alex"},"body":{"type":"doc","version":1,"content":[{"type":"unknown"}]}}]}`))
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(comments, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("user-1"))
		output, err = clicky.Format(comments, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("Alex commented"), ContainSubstring(`"type":"unknown"`)))
	})

	It("distinguishes unloaded collections from known counts", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Partial record","comment":{"total":12}}`), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("12 comments"), Not(ContainSubstring("0 attachments"))))
	})

	It("omits native payloads from standalone comments and attachments", func() {
		comment := json.RawMessage(`{"id":"comment-1","visibility":{"type":"role","value":"Developers"},"body":{"type":"doc","version":1,"content":[]}}`)
		comments, err := ParseComments(json.RawMessage(`{"comments":[` + string(comment) + `]}`))
		Expect(err).NotTo(HaveOccurred())
		attachments, err := ParseAttachments(json.RawMessage(`[{"id":"file-1","filename":"report.pdf","content":"https://example.test/report.pdf","thumbnail":"https://example.test/thumbnail"}]`))
		Expect(err).NotTo(HaveOccurred())
		commentJSON, err := json.Marshal(comments)
		Expect(err).NotTo(HaveOccurred())
		attachmentJSON, err := json.Marshal(attachments)
		Expect(err).NotTo(HaveOccurred())
		var commentOutput, attachmentOutput []map[string]any
		Expect(json.Unmarshal(commentJSON, &commentOutput)).To(Succeed())
		Expect(json.Unmarshal(attachmentJSON, &attachmentOutput)).To(Succeed())
		Expect(commentOutput).To(ConsistOf(SatisfyAll(HaveKeyWithValue("id", "comment-1"), Not(HaveKey("native")))))
		Expect(attachmentOutput).To(ConsistOf(SatisfyAll(HaveKeyWithValue("name", "report.pdf"), HaveKeyWithValue("url", "https://example.test/report.pdf"), Not(HaveKey("native")))))
	})
})
