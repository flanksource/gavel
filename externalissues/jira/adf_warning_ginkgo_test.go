package jira

import (
	"bytes"
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ADF warnings", func() {
	var warnings bytes.Buffer

	BeforeEach(func() {
		warnings.Reset()
		previous := logger.StandardLogger()
		logger.Use(&warnings)
		DeferCleanup(func() { logger.SetLogger(previous) })
	})

	It("preserves sibling text and supported marks while warning with comment context", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Warning example","comments":[{"id":"comment-1","author":{"displayName":"Alex"},"body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Before "},{"type":"text","text":"marked","marks":[{"type":"strong"},{"type":"futureMark"},{"type":"em"}]},{"type":"futureNode","content":[{"type":"text","text":" retained child"}]},{"type":"text","text":" after"}]}]}}]}`), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		issue.Full = true
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("Before "), ContainSubstring("**marked**"), ContainSubstring("retained child"), ContainSubstring(" after")))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("comment comment-1 by Alex"), ContainSubstring("$.content[0].content[1]"), ContainSubstring("futureMark"), ContainSubstring("futureNode")))
	})

	It("preserves raw documents that cannot be decoded and warns during rendering", func() {
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Malformed ADF","description":{"type":"doc","version":1,"content":{"unexpected":"retained raw"}}}`), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		issue.Full = true
		output, err := clicky.Format(issue, clicky.FormatOptions{Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("```json"), ContainSubstring("retained raw")))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("issue EXAMPLE-1 description"), ContainSubstring("decode ADF")))
	})

	It("retains an unsupported leaf as JSON instead of dropping it", func() {
		doc := ADFNode{Type: "doc", Version: 1, Content: []ADFNode{{Type: "futureLeaf", Attrs: json.RawMessage(`{"label":"retained metadata"}`)}}}
		output, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(SatisfyAll(ContainSubstring("futureLeaf"), ContainSubstring("retained metadata")))
		Expect(warnings.String()).To(ContainSubstring("$.content[0]"))
	})

	It("keeps unsupported inline leaves inside their surrounding paragraph", func() {
		output, err := renderADF(json.RawMessage(`{"type":"paragraph","content":[{"type":"text","text":"Before "},{"type":"futureLeaf","attrs":{"label":"Retained"}},{"type":"text","text":" after"}]}`), adfRenderOptions{Path: "$"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(Equal("Before `{\"type\":\"futureLeaf\",\"attrs\":{\"label\":\"Retained\"}}` after"))
		Expect(output.HTML()).To(ContainSubstring("<code>"))
	})

	DescribeTable("preserves text after structural ADF errors", func(content string) {
		output, err := renderADF(json.RawMessage(content), adfRenderOptions{Path: "example $"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output.Markdown()).To(ContainSubstring("Retained"))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("example $")))
	},
		Entry("document version", `{"type":"doc","version":2,"content":[{"type":"text","text":"Retained"}]}`),
		Entry("heading level", `{"type":"heading","attrs":{"level":9},"content":[{"type":"text","text":"Retained"}]}`),
		Entry("invalid attributes", `{"type":"paragraph","attrs":123,"content":[{"type":"text","text":"Retained"}]}`),
		Entry("invalid list child", `{"type":"bulletList","content":[{"type":"text","text":"Retained"}]}`),
		Entry("invalid table row", `{"type":"table","content":[{"type":"paragraph","content":[{"type":"text","text":"Retained"}]}]}`),
		Entry("invalid code block child", `{"type":"codeBlock","content":[{"type":"paragraph","content":[{"type":"text","text":"Retained"}]}]}`),
		Entry("invalid link", `{"type":"text","text":"Retained","marks":[{"type":"link","attrs":{}}]}`),
	)

	It("warns about malformed custom-field ADF with its field ID", func() {
		fields, err := MapFields(map[string]json.RawMessage{"customfield_example": json.RawMessage(`{"type":"doc","content":{"text":"Retained"}}`)}, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fields["customfield_example"].Value)).To(ContainSubstring("Retained"))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("customfield_example"), ContainSubstring("decode ADF")))
	})
})
