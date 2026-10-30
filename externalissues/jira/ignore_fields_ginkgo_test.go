package jira

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira ignored fields", func() {
	const raw = `{"id":"1","key":"EXAMPLE-1","summary":"Visible title","url":"https://example.test/browse/EXAMPLE-1","project":{"id":"10001","name":"Visible project"},"description":"Hidden description","priority":{"name":"Hidden priority"},"labels":["hidden-label"],"status":{"name":"Hidden status"},"assignee":{"displayName":"Hidden assignee"},"reporter":{"displayName":"Hidden reporter"},"issuetype":{"name":"Hidden type"},"created":"2026-01-01T00:00:00Z","updated":"2026-01-02T00:00:00Z","comments":[{"body":"Hidden comment"}],"comment":{"total":1},"attachment":[{"filename":"hidden.pdf","content":"https://example.test/hidden.pdf"}],"customfield_1":"Hidden custom one","customfield_2":"Hidden custom two","unknown":"Unknown value"}`
	var catalog map[string]FieldDefinition

	BeforeEach(func() {
		catalog = map[string]FieldDefinition{
			"customfield_1": {ID: "customfield_1", Name: "Internal notes"},
			"customfield_2": {ID: "customfield_2", Name: "Internal notes"},
			"comment":       {ID: "comment", Name: "Comments"},
			"attachment":    {ID: "attachment", Name: "Attachments"},
		}
	})

	DescribeTable("matches field IDs and names with commons semantics",
		func(patterns, retained []string) {
			issue, err := ParseIssue([]byte(raw), ParseOptions{Catalog: catalog, IgnoreFields: patterns})
			Expect(err).NotTo(HaveOccurred())
			Expect(issue.Fields).To(HaveLen(len(retained)))
			for _, id := range retained {
				Expect(issue.Fields).To(HaveKey(id))
			}
		},
		Entry("unset configuration", nil, []string{"customfield_1", "customfield_2", "unknown"}),
		Entry("empty configuration", []string{}, []string{"customfield_1", "customfield_2", "unknown"}),
		Entry("case insensitive ID", []string{"CUSTOMFIELD_1"}, []string{"customfield_2", "unknown"}),
		Entry("ID wildcard", []string{"customfield_*"}, []string{"unknown"}),
		Entry("case insensitive display name", []string{"INTERNAL NOTES"}, []string{"unknown"}),
		Entry("display name prefix", []string{"Internal*"}, []string{"unknown"}),
		Entry("display name suffix", []string{"*notes"}, []string{"unknown"}),
		Entry("display name contains", []string{"*ternal*"}, []string{"unknown"}),
		Entry("mixed name and ID", []string{"Internal notes", "unknown"}, []string{}),
		Entry("comma separated patterns", []string{"customfield_1,unknown"}, []string{"customfield_2"}),
		Entry("negated wildcard retains every field", []string{"!*"}, []string{"customfield_1", "customfield_2", "unknown"}),
		Entry("missing catalog name still matches ID", []string{"UNKNOWN"}, []string{"customfield_1", "customfield_2"}),
		Entry("exported keys are not display names", []string{"internal_notes"}, []string{"customfield_1", "customfield_2", "unknown"}),
		Entry("unmatched pattern", []string{"missing"}, []string{"customfield_1", "customfield_2", "unknown"}),
		Entry("middle wildcard retains commons behavior", []string{"custom*1"}, []string{"customfield_1", "customfield_2", "unknown"}),
	)

	It("retains identity and project while omitting optional fields from every representation", func() {
		issue, err := ParseIssue([]byte(raw), ParseOptions{Catalog: catalog, IgnoreFields: []string{"*"}})
		Expect(err).NotTo(HaveOccurred())
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(MatchJSON(`{"provider":"jira","id":"1","key":"EXAMPLE-1","title":"Visible title","url":"https://example.test/browse/EXAMPLE-1","project":{"id":"10001","name":"Visible project"},"fields":{}}`))
		rendered, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		for _, output := range []string{rendered.String(), rendered.Markdown(), rendered.HTML()} {
			Expect(output).To(SatisfyAll(ContainSubstring("Visible title"), Not(ContainSubstring("Hidden")), Not(ContainSubstring("hidden")), Not(ContainSubstring("Unknown value"))))
		}
	})

	It("omits comment content and counts when the catalog name matches", func() {
		issue, err := ParseIssue([]byte(raw), ParseOptions{Catalog: catalog, IgnoreFields: []string{"Comments"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Comments).To(BeEmpty())
		Expect(issue.CommentCount).To(BeNil())
		Expect(issue.Attachments).To(HaveLen(1))
	})

	It("filters ignored built-ins before decoding invalid values", func() {
		input := []byte(`{"key":"EXAMPLE-1","summary":"Visible title","priority":123,"status":{"statusCategory":{"key":"unexpected"}}}`)
		_, err := ParseIssue(input, ParseOptions{})
		Expect(err).To(HaveOccurred())
		issue, err := ParseIssue(input, ParseOptions{IgnoreFields: []string{"priority", "status"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Priority).To(BeNil())
		Expect(issue.Status).To(BeNil())
	})

	It("passes exclusions through ParseList with each issue's catalog", func() {
		input := json.RawMessage(`{"issues":[{"key":"EXAMPLE-1","summary":"First","customfield_1":"Hidden"},{"key":"EXAMPLE-2","summary":"Second","customfield_1":"Visible"}]}`)
		issues, err := ParseList(input, ListOptions{IgnoreFields: []string{"Internal*"}, Catalogs: map[string]map[string]FieldDefinition{
			"EXAMPLE-1": catalog,
			"EXAMPLE-2": {"customfield_1": {ID: "customfield_1", Name: "Public notes"}},
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(HaveLen(2))
		Expect(issues[0].Fields).To(BeEmpty())
		Expect(issues[1].Fields).To(HaveKey("customfield_1"))
	})
})
