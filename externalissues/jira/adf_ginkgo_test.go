package jira

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ADF", func() {
	It("renders formatted paragraphs, lists, mentions and file references", func() {
		var doc ADFNode
		err := json.Unmarshal([]byte(`{"version":1,"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello ","marks":[{"type":"strong"}]},{"type":"mention","attrs":{"id":"user-1","text":"@Alex"}}]},{"type":"orderedList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"First"}]}]}]},{"type":"mediaSingle","content":[{"type":"media","attrs":{"type":"file","id":"file-1","collection":"","alt":"image.png"}}]}]}`), &doc)
		Expect(err).NotTo(HaveOccurred())
		rendered, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Markdown()).To(ContainSubstring("**Hello **"))
		Expect(rendered.Markdown()).To(ContainSubstring("@Alex"))
		Expect(rendered.Markdown()).To(ContainSubstring("1. First"))
		Expect(rendered.Markdown()).To(ContainSubstring("atlassian-media://"))
	})

	It("reports unknown node types instead of silently dropping content", func() {
		doc := ADFNode{Type: "doc", Version: 1, Content: []ADFNode{{Type: "unknown"}}}
		_, err := doc.Render()
		Expect(err).To(MatchError(ContainSubstring("unknown")))
	})

	It("renders table headings and cells in Markdown", func() {
		var doc ADFNode
		err := json.Unmarshal([]byte(`{"version":1,"type":"doc","content":[{"type":"table","content":[{"type":"tableRow","content":[{"type":"tableHeader","content":[{"type":"paragraph","content":[{"type":"text","text":"Name"}]}]}]},{"type":"tableRow","content":[{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Value"}]}]}]}]}]}`), &doc)
		Expect(err).NotTo(HaveOccurred())
		rendered, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Markdown()).To(ContainSubstring("Name"))
		Expect(rendered.Markdown()).To(ContainSubstring("Value"))
	})
})
