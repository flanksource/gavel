package jira

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira issue rendering", func() {
	It("includes description, comments, attachments and custom fields in Markdown", func() {
		issue, err := ParseIssue([]byte(`{"key":"EXAMPLE-1","summary":"Example","url":"https://example.com/EXAMPLE-1","customfield_1":"Custom value","description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Description text"}]}]},"comments":[{"author":{"displayName":"Alex"},"created":"2026-01-01","body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Comment text"}]}]}}],"attachment":[{"filename":"report.pdf","content":"https://example.com/report.pdf"}]}`))
		Expect(err).NotTo(HaveOccurred())
		output := issue.Pretty().Markdown()
		Expect(output).To(ContainSubstring("Description text"))
		Expect(output).To(ContainSubstring("Comment text"))
		Expect(output).To(ContainSubstring("report.pdf"))
		Expect(output).To(ContainSubstring("https://example.com/report.pdf"))
		Expect(output).To(ContainSubstring("Custom value"))
	})
})
