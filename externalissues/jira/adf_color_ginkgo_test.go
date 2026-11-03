package jira

import (
	"bytes"
	"encoding/json"

	"github.com/flanksource/clicky/markdown"
	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ADF textColor", func() {
	var warnings bytes.Buffer
	BeforeEach(func() {
		warnings.Reset()
		previous := logger.StandardLogger()
		logger.Use(&warnings)
		DeferCleanup(func() { logger.SetLogger(previous) })
	})

	It("preserves color, strong formatting and literal text in paragraphs and tables", func() {
		for _, marks := range [][]ADFMark{
			{{Type: "strong"}, {Type: "textColor", Attrs: json.RawMessage(`{"color":"#ff5630"}`)}},
			{{Type: "textColor", Attrs: json.RawMessage(`{"color":"#ff5630"}`)}, {Type: "strong"}},
		} {
			paragraph := ADFNode{Type: "paragraph", Content: []ADFNode{{Type: "text", Text: "Amount < limit & total", Marks: marks}}}
			for _, content := range []ADFNode{paragraph, {Type: "table", Content: []ADFNode{{Type: "tableRow", Content: []ADFNode{{Type: "tableCell", Content: []ADFNode{paragraph}}}}}}} {
				doc := ADFNode{Type: "doc", Version: 1, Content: []ADFNode{content}}
				rendered, err := doc.Render()
				Expect(err).NotTo(HaveOccurred())
				Expect(rendered.HTML()).To(SatisfyAll(ContainSubstring(`style="color:#ff5630"`), ContainSubstring("<strong>"), ContainSubstring("&lt; limit &amp; total")))
				Expect(rendered.Markdown()).To(SatisfyAll(ContainSubstring(`style="color:#ff5630"`), ContainSubstring("**")))
				Expect(documentPreview(rendered.(markdown.Node))).To(ContainSubstring("Amount < limit & total"))
				Expect(documentPreview(rendered.(markdown.Node))).NotTo(ContainSubstring("<span"))
			}
		}
		Expect(warnings.String()).To(BeEmpty())
	})

	It("warns about invalid colors with their ADF path and retains text", func() {
		for _, attrs := range []json.RawMessage{nil, json.RawMessage(`{}`), json.RawMessage(`null`), json.RawMessage(`{"color":123}`), json.RawMessage(`{"color":"red"}`), json.RawMessage(`{"color":"#ff5630;display:none"}`)} {
			doc := ADFNode{Type: "doc", Version: 1, Content: []ADFNode{{Type: "paragraph", Content: []ADFNode{{Type: "text", Text: "Example", Marks: []ADFMark{{Type: "textColor", Attrs: attrs}}}}}}}
			warnings.Reset()
			rendered, err := doc.Render()
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered.Markdown()).To(Equal("Example"))
			Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("$.content[0].content[0]"), ContainSubstring("textColor")))
		}
	})

	It("warns about textColor on non-text nodes and unknown marks without losing content", func() {
		for _, node := range []ADFNode{
			{Type: "paragraph", Content: []ADFNode{{Type: "text", Text: "Example"}}, Marks: []ADFMark{{Type: "textColor", Attrs: json.RawMessage(`{"color":"#bf2600"}`)}}},
			{Type: "text", Text: "Example", Marks: []ADFMark{{Type: "unknown"}}},
		} {
			warnings.Reset()
			rendered, err := node.Render()
			Expect(err).NotTo(HaveOccurred())
			Expect(rendered.Markdown()).To(Equal("Example"))
			Expect(warnings.String()).To(ContainSubstring("WRN"))
		}
	})
})
