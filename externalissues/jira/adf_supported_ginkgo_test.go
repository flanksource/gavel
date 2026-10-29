package jira

import (
	"bytes"
	"encoding/json"

	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ADF supported content", func() {
	var warnings bytes.Buffer
	BeforeEach(func() {
		previous := logger.StandardLogger()
		logger.Use(&warnings)
		DeferCleanup(func() { logger.SetLogger(previous) })
	})

	DescribeTable("renders standard content without unsupported warnings", func(source, markdown, html, plain string) {
		var node ADFNode
		Expect(json.Unmarshal([]byte(source), &node)).To(Succeed())
		view, err := node.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(ContainSubstring(markdown))
		Expect(view.HTML()).To(ContainSubstring(html))
		Expect(view.String()).To(ContainSubstring(plain))
		Expect(view.ANSI()).NotTo(ContainSubstring("<span"))
		Expect(view.ANSI()).NotTo(ContainSubstring("<u>"))
		Expect(view.ANSI()).NotTo(ContainSubstring("<sup>"))
		Expect(warnings.String()).To(BeEmpty())
	},
		Entry("Jira card", `{"type":"inlineCard","attrs":{"url":"https://example.test/browse/EXAMPLE-2"}}`, "[EXAMPLE-2](https://example.test/browse/EXAMPLE-2)", `href="https://example.test/browse/EXAMPLE-2"`, "EXAMPLE-2"),
		Entry("external card", `{"type":"inlineCard","attrs":{"url":"https://example.test/docs"}}`, "[https://example.test/docs]", `href="https://example.test/docs"`, "https://example.test/docs"),
		Entry("card data", `{"type":"inlineCard","attrs":{"data":{"@id":"https://example.test/docs","name":"Guide"}}}`, "[Guide](https://example.test/docs)", ">Guide</a>", "Guide"),
		Entry("inline file", `{"type":"paragraph","content":[{"type":"text","text":"See "},{"type":"mediaInline","attrs":{"id":"file-1","collection":"","type":"file","alt":"report.pdf"}},{"type":"text","text":" now"}]}`, "See [report.pdf](atlassian-media:///file-1) now", ">report.pdf</a> now", "See report.pdf now"),
		Entry("date", `{"type":"date","attrs":{"timestamp":"1767225600000"}}`, "2026-01-01", "2026-01-01", "2026-01-01"),
		Entry("rule", `{"type":"rule"}`, "---", "<hr", ""),
		Entry("panel", `{"type":"panel","attrs":{"panelType":"warning"},"content":[{"type":"paragraph","content":[{"type":"text","text":"Check this"}]}]}`, "> **Warning**", "<blockquote>", "Check this"),
		Entry("underline with emphasis", `{"type":"text","text":"Amount < limit","marks":[{"type":"strong"},{"type":"underline"}]}`, "<u>**Amount", "<u><strong>Amount &lt; limit</strong></u>", "Amount < limit"),
		Entry("superscript", `{"type":"text","text":"2","marks":[{"type":"subsup","attrs":{"type":"sup"}}]}`, "<sup>2</sup>", "<sup>2</sup>", "2"),
		Entry("subscript", `{"type":"text","text":"2","marks":[{"type":"subsup","attrs":{"type":"sub"}}]}`, "<sub>2</sub>", "<sub>2</sub>", "2"),
		Entry("media border", `{"type":"media","attrs":{"id":"file-1","alt":"image.png"},"marks":[{"type":"border","attrs":{"size":2,"color":"#172b4d"}}]}`, "border:2px solid #172b4d", "border:2px solid #172b4d", "image.png"),
		Entry("native task list", `{"type":"taskList","attrs":{"localId":"list-1"},"content":[{"type":"taskItem","attrs":{"localId":"item-1","state":"DONE"},"content":[{"type":"text","text":"Reviewed"}]},{"type":"taskItem","attrs":{"localId":"item-2","state":"TODO"},"content":[{"type":"text","text":"Deploy"}]}]}`, "- [x] Reviewed\n- [ ] Deploy", `type="checkbox" disabled checked`, "Deploy"),
	)

	DescribeTable("retains content and warns for invalid attributes", func(source string) {
		view := renderADF(json.RawMessage(source), "issue EXAMPLE-1 description $")
		Expect(view.Markdown()).NotTo(BeEmpty())
		Expect(warnings.String()).To(ContainSubstring("issue EXAMPLE-1 description $"))
	},
		Entry("missing card URL", `{"type":"inlineCard","attrs":{}}`),
		Entry("invalid date", `{"type":"date","attrs":{"timestamp":"yesterday"}}`),
		Entry("invalid panel", `{"type":"panel","attrs":{"panelType":"future"},"content":[{"type":"text","text":"Retained"}]}`),
		Entry("invalid task state", `{"type":"taskItem","attrs":{"state":"MAYBE"},"content":[{"type":"text","text":"Retained"}]}`),
		Entry("invalid subscript", `{"type":"text","text":"Retained","marks":[{"type":"subsup","attrs":{"type":"future"}}]}`),
		Entry("invalid border color", `{"type":"media","attrs":{"id":"file-1"},"marks":[{"type":"border","attrs":{"size":2,"color":"red;display:none"}}]}`),
		Entry("invalid border size", `{"type":"media","attrs":{"id":"file-1"},"marks":[{"type":"border","attrs":{"size":100,"color":"#172b4d"}}]}`),
	)
})
