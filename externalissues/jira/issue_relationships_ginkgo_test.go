package jira

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira issue relationships and checklists", func() {
	var issue Issue
	BeforeEach(func() {
		data, err := os.ReadFile("testdata/relationships.json")
		Expect(err).NotTo(HaveOccurred())
		issue, err = ParseIssue(data, ParseOptions{Site: "https://example.test", Catalog: map[string]FieldDefinition{
			"resolution": {Name: "Resolution"}, "parent": {Name: "Parent"}, "issuelinks": {Name: "Linked Issues"},
			"customfield_ready": {Name: "Readiness Checklist"},
			"customfield_epic":  {Name: "Epic Link", Schema: json.RawMessage(`{"custom":"com.pyxis.greenhopper.jira:gh-epic-link"}`)},
		}})
		Expect(err).NotTo(HaveOccurred())
	})

	It("keeps styled custom fields readable without HTML tags in terminal output", func() {
		fields, err := MapFields(map[string]json.RawMessage{
			"customfield_styled": json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Amount < limit","marks":[{"type":"underline"}]}]}]}`),
		}, nil)
		Expect(err).NotTo(HaveOccurred())
		issue.Fields = externalissues.Fields(fields)
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(ContainSubstring("<u>"))
		for _, output := range []string{view.String(), view.ANSI()} {
			Expect(output).To(SatisfyAll(ContainSubstring("Amount < limit"), Not(ContainSubstring("<u>")), Not(ContainSubstring("</u>"))))
		}
	})

	It("renders leading horizontal rules in rich fields without treating the body as YAML frontmatter", func() {
		fields, err := MapFields(map[string]json.RawMessage{
			"customfield_rules": json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"rule"},{"type":"paragraph","content":[{"type":"text","text":"Status: [literal"}]},{"type":"rule"}]}`),
		}, nil)
		Expect(err).NotTo(HaveOccurred())
		issue.Fields = externalissues.Fields(fields)
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.HTML()).To(SatisfyAll(ContainSubstring("<hr"), ContainSubstring("Status: [literal")))
	})

	It("uses local links in relationships, descriptions, comments and rich fields", func() {
		root := GinkgoT().TempDir()
		resolver := localLinkOptions{Site: issue.Site, Path: filepath.Join(root, "10001", "EXAMPLE-1.md"), Files: map[string]string{
			"EXAMPLE-2": filepath.Join(root, "10001", "EXAMPLE-2.md"),
			"EXAMPLE-3": filepath.Join(root, "10002", "EXAMPLE-3.md"),
		}}
		view, err := issue.Render(RenderOptions{Full: true, ResolveIssueURL: resolver.resolve})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(SatisfyAll(
			ContainSubstring("[EXAMPLE-2 — Refunds](EXAMPLE-2.md)"),
			ContainSubstring("[EXAMPLE-3 — Loan top up](../10002/EXAMPLE-3.md)"),
			ContainSubstring("[EXAMPLE-2](EXAMPLE-2.md)"),
			ContainSubstring("[EXAMPLE-3](../10002/EXAMPLE-3.md)"),
			ContainSubstring("[EXAMPLE-4](https://example.test/browse/EXAMPLE-4)"),
		))
	})

	It("renders concise resolution and directional relationships without changing exported values", func() {
		before, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		for _, full := range []bool{false, true} {
			issue.Full = full
			output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(SatisfyAll(
				ContainSubstring("Cancelled"),
				ContainSubstring("[EXAMPLE-2 — Refunds](https://example.test/browse/EXAMPLE-2)"),
				ContainSubstring("is required for"), ContainSubstring("depends on"),
				ContainSubstring("[EXAMPLE-3 — Loan top up](https://example.test/browse/EXAMPLE-3)"),
				ContainSubstring("[EXAMPLE-4 — Collections](https://example.test/browse/EXAMPLE-4)"),
				Not(ContainSubstring("**self**")), Not(ContainSubstring("**inwardIssue**")), Not(ContainSubstring("avatarId")),
			))
			if full {
				Expect(output).To(ContainSubstring("No work performed"))
			}
		}
		after, err := clicky.Format(issue, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(MatchJSON(before))
	})

	It("renders checklist groups and individual task states outside the full field table", func() {
		issue.Full = true
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(
			ContainSubstring("## Readiness Checklist"), ContainSubstring("**Ready for analysis**"),
			ContainSubstring("- [ ] Contacts identified\n- [x] **Wireframes** linked"),
			Not(ContainSubstring("| Readiness Checklist |")), Not(ContainSubstring(":cross")),
		))
		html, err := clicky.Format(issue, clicky.FormatOptions{Format: "html"})
		Expect(err).NotTo(HaveOccurred())
		Expect(html).To(ContainSubstring(`type="checkbox" disabled checked`))
	})
})
