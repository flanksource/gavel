package jira

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Markdown comments", func() {
	It("preserves headings, links, inline code and nested lists in Jira ADF", func() {
		doc, err := MarkdownToADF("# Instructions\n\nUse **bold** and [guide](https://example.test) with `run()`.\n\n- Parent\n  - Child\n\n```go\nrun()\n```\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Type).To(Equal("doc"))
		Expect(doc.Version).To(Equal(1))
		data, err := json.Marshal(doc)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring(`"type":"bulletList"`), ContainSubstring(`"type":"codeBlock"`), ContainSubstring(`"type":"link"`), ContainSubstring(`"href":"https://example.test"`)))
		rendered, err := doc.Render()
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Markdown()).To(SatisfyAll(ContainSubstring("# Instructions"), ContainSubstring("**bold**"), ContainSubstring("`run()`"), ContainSubstring("Child")))
	})
	It("rejects empty bodies and unsupported Markdown instead of losing content", func() {
		_, err := MarkdownToADF(" \n")
		Expect(err).To(MatchError(ContainSubstring("must not be empty")))
		_, err = MarkdownToADF("- [x] Task")
		Expect(err).To(MatchError(ContainSubstring("task lists")))
	})
})
