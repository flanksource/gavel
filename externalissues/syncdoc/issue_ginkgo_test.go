package syncdoc

import (
	"os"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("issue frontmatter block", func() {
	It("round-trips gavel.issue and foreign top-level keys unchanged", func() {
		root, err := os.OpenRoot(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		header := `gavel:
    version: 1
    source:
        provider: jira
        site: https://example.test
        key: EXAMPLE-7
    issue:
        type: Defect
        status: Analysis
        status_category: in_progress
        created: 2026-01-13T16:52:35+02:00
        resolved: 2026-03-01T10:00:00+02:00
        due: "2026-04-01"
        labels: [Group_Life, benefitcoverage]
        fields:
            Countries Impacted:
                - Kenya
            Status Category Changed: "2026-03-03T09:31:51.458+02:00"
            Story Points: 5
    sync:
        synced_at: 0001-01-01T00:00:00Z
        attachments: false
        extract: false
    attachments: {}
oipa:
    classification:
        product: Group Life
`
		original := "---\n" + header + "---\n\nBody\n"
		Expect(root.WriteFile("issue.md", []byte(original), 0o600)).To(Succeed())
		doc, err := Read(root, "issue.md")
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Metadata.Issue.StatusCategory).To(Equal(externalissues.StatusInProgress))
		encoded, err := doc.Markdown("Body")
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(Equal(original))
	})
})
