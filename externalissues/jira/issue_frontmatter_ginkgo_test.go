package jira

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var frontmatterCatalog = map[string]FieldDefinition{
	"resolution": {Name: "Resolution"}, "resolutiondate": {Name: "Resolved", Schema: json.RawMessage(`{"type":"datetime"}`)},
	"duedate": {Name: "Due Date", Schema: json.RawMessage(`{"type":"date"}`)}, "components": {Name: "Components"},
	"statusCategory": {Name: "Status Category"}, "creator": {Name: "Creator"},
	"statuscategorychangedate": {Name: "Status Category Changed", Schema: json.RawMessage(`{"type":"datetime"}`)},
	"customfield_severity":     {Name: "Severities"}, "customfield_countries": {Name: "Countries Impacted"},
	"customfield_sprint": {Name: "Sprint"}, "customfield_points": {Name: "Story Points"}, "customfield_ratio": {Name: "Ratio"},
	"customfield_flag": {Name: "Flagged"}, "customfield_region": {Name: "Region"},
	"customfield_launch": {Name: "Launch", Schema: json.RawMessage(`{"type":"date"}`)}, "fixVersions": {Name: "Fix versions"},
	"customfield_outcomes": {Name: "Business Outcomes"}, "environment": {Name: "Environment"},
	"customfield_empty_list": {Name: "Empty List"}, "customfield_empty_object": {Name: "Empty Object"},
	"customfield_empty_text": {Name: "Empty Text"}, "customfield_null": {Name: "Unset"},
	"watches": {Name: "Watchers"}, "progress": {Name: "Progress"}, "parent": {Name: "Parent"}, "issuelinks": {Name: "Linked Issues"},
}

func frontmatterIssue(mutate func(map[string]json.RawMessage)) Issue {
	data, err := os.ReadFile("testdata/frontmatter.json")
	Expect(err).NotTo(HaveOccurred())
	var payload map[string]json.RawMessage
	Expect(json.Unmarshal(data, &payload)).To(Succeed())
	if mutate != nil {
		mutate(payload)
	}
	data, err = json.Marshal(payload)
	Expect(err).NotTo(HaveOccurred())
	issue, err := ParseIssue(data, ParseOptions{Site: "https://example.test", Catalog: frontmatterCatalog})
	Expect(err).NotTo(HaveOccurred())
	return issue
}

func frontmatterTime(value string) *time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	Expect(err).NotTo(HaveOccurred())
	return &parsed
}

var _ = Describe("issue frontmatter", func() {
	It("promotes core fields and converts every other valued field by its display name", func() {
		block, err := frontmatterIssue(nil).syncMetadata()
		Expect(err).NotTo(HaveOccurred())
		Expect(block).To(Equal(&syncdoc.Issue{
			Type: "Defect", Status: "Analysis", StatusCategory: externalissues.StatusInProgress, Priority: "High",
			Assignee: "Alex Example", Reporter: "Sam Example", Project: "Group Life & Credit Life Testing",
			Created: frontmatterTime("2026-01-13T16:52:35+02:00"), Updated: frontmatterTime("2026-10-28T16:31:47+02:00"),
			Resolved: frontmatterTime("2026-03-01T10:00:00+02:00"), Resolution: "Done", Due: "2026-04-01",
			Labels: []string{"Group_Life", "benefit*coverage"}, Components: []string{"Billing"},
			Fields: map[string]any{
				"Creator": "Sam Example", "Severities": "Sev 2 - Major", "Countries Impacted": []any{"Kenya"},
				"Status Category Changed": "2026-03-03T09:31:51.458+02:00", "Sprint": []any{"Sprint 12"},
				"Story Points": int64(5), "Ratio": 0.5, "Flagged": true, "Region": "Region / East", "Launch": "2026-05-01",
				"Fix versions": []any{"1.2.0"}, "Business Outcomes": "Faster billing", "Watchers": int64(3),
				"Progress": map[string]any{"progress": int64(0), "total": int64(0)},
				"Parent":   "EXAMPLE-1", "Linked Issues": []any{"relates to EXAMPLE-2"},
			},
		}))
	})

	It("omits resolved and resolution while the issue is unresolved", func() {
		block, err := frontmatterIssue(func(payload map[string]json.RawMessage) {
			payload["resolution"], payload["resolutiondate"] = json.RawMessage("null"), json.RawMessage("null")
		}).syncMetadata()
		Expect(err).NotTo(HaveOccurred())
		Expect(block.Resolved).To(BeNil())
		Expect(block.Resolution).To(BeEmpty())
		Expect(block.Fields).NotTo(HaveKey("Resolution"))
	})

	It("fails loudly on a datetime field that does not parse", func() {
		_, err := frontmatterIssue(func(payload map[string]json.RawMessage) {
			payload["statuscategorychangedate"] = json.RawMessage(`"yesterday"`)
		}).syncMetadata()
		Expect(err).To(MatchError(ContainSubstring("Status Category Changed")))
	})

	It("replaces a stale issue block on re-sync and keeps foreign frontmatter", func() {
		root, err := os.OpenRoot(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		stale := "---\ngavel:\n  version: 1\n  source:\n    provider: jira\n    site: https://example.test\n    id: \"1836428\"\n    key: EXAMPLE-7\n  issue:\n    status: Open\n    resolution: Won't Do\n    fields:\n      Removed Field: stale\noipa:\n  classification:\n    product: Group Life\n---\n\nBody\n"
		Expect(root.WriteFile("EXAMPLE-7.md", []byte(stale), 0o600)).To(Succeed())
		doc, err := syncdoc.Read(root, "EXAMPLE-7.md")
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Metadata.Issue.Status).To(Equal("Open"))
		Expect(prepareSyncMetadata(doc, frontmatterIssue(nil), SyncOptions{})).To(Succeed())
		data, err := doc.Markdown("Body")
		Expect(err).NotTo(HaveOccurred())
		Expect(root.WriteFile("EXAMPLE-7.md", data, 0o600)).To(Succeed())
		reread, err := syncdoc.Read(root, "EXAMPLE-7.md")
		Expect(err).NotTo(HaveOccurred())
		Expect(reread.Metadata.Issue.Status).To(Equal("Analysis"))
		Expect(reread.Metadata.Issue.Resolution).To(Equal("Done"))
		Expect(reread.Metadata.Issue.Fields).NotTo(HaveKey("Removed Field"))
		Expect(reread.Extra).To(HaveKey("oipa"))
		Expect(string(data)).To(SatisfyAll(ContainSubstring("product: Group Life"), ContainSubstring("labels: [Group_Life, benefit*coverage]")))
		Expect(strings.Count(string(data), "issue:")).To(Equal(1))
	})
})
