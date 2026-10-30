package jira

import (
	"github.com/flanksource/clicky/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("checklist line conversion", func() {
	DescribeTable("preserves non-checklist lines and only converts leading state markers", func(source, expected string) {
		doc, err := markdown.ParseString(source)
		Expect(err).NotTo(HaveOccurred())
		Expect(checklistDocument(doc.Root).Markdown()).To(Equal(expected))
	},
		Entry("inline cross", "A failed check :cross\\_mark: remains text", "A failed check :cross\\_mark: remains text"),
		Entry("formatted items", ":cross\\_mark: First  \n:white\\_check\\_mark: **Second**", "- [ ] First\n- [x] **Second**"),
		Entry("group labels", "**Analysis**\n\n❌ Contacts\n\n**Planning**\n\n✅ Reviewed", "**Analysis**\n\n- [ ] Contacts\n\n**Planning**\n\n- [x] Reviewed"),
		Entry("native checkboxes", "- [ ] Pending\n- [x] Done", "- [ ] Pending\n- [x] Done"),
	)
})
