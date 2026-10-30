package jira

import (
	"encoding/json"

	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ADF", func() {
	It("preserves adjacent inline text when building a compact preview", func() {
		body, err := renderDocument(&externalissues.Document{Format: "adf", Content: json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"read"},{"type":"text","text":"_only","marks":[{"type":"code"}]}]},{"type":"paragraph","content":[{"type":"text","text":"Next paragraph"}]}]}`)}, documentRenderOptions{Preview: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(body.Markdown()).To(Equal(`read\_only Next paragraph`))
		Expect(body.String()).To(Equal("read_only Next paragraph"))
	})

	It("preserves inline code in paragraphs and table cells", func() {
		text := ADFNode{Type: "text", Text: "run()", Marks: []ADFMark{{Type: "code"}}}
		for _, doc := range []ADFNode{
			{Type: "doc", Version: 1, Content: []ADFNode{{Type: "paragraph", Content: []ADFNode{text}}}},
			{Type: "doc", Version: 1, Content: []ADFNode{{Type: "table", Content: []ADFNode{{Type: "tableRow", Content: []ADFNode{{Type: "tableHeader", Content: []ADFNode{{Type: "paragraph", Content: []ADFNode{text}}}}}}}}}},
		} {
			rendered, err := doc.Render()
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered.Markdown()).To(ContainSubstring("`run()`"))
			Expect(rendered.HTML()).To(ContainSubstring("<code>run()</code>"))
		}
	})

	It("preserves nested list indentation", func() {
		var doc ADFNode
		Expect(json.Unmarshal([]byte(`{"type":"doc","version":1,"content":[{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Parent"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"Child"}]}]}]}]}]}]}`), &doc)).To(Succeed())
		rendered, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Markdown()).To(Equal("- Parent\n  - Child"))
	})

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
		output, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(ContainSubstring(`"type":"unknown"`))
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
