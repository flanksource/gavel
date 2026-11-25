package jira

import (
	"encoding/json"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira comment presentation", func() {
	It("renders one Comments heading, bold author and creation date, body and an HTML rule per comment", func() {
		comments, err := ParseComments(json.RawMessage(`{"comments":[
			{"id":"comment-1","author":{"displayName":"Alex","url":"https://example.test/people/alex","avatarUrls":{"24x24":"https://example.test/avatar.png"}},"created":"2026-01-02T03:04:05+0200","updated":"2026-01-03T11:00:00Z","body":"First comment"},
			{"id":"comment-2","author":{"displayName":"Blair"},"created":"2026-01-02T04:05:06Z","body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Second comment","marks":[{"type":"strong"}]}]}]}}
		]}`))
		Expect(err).NotTo(HaveOccurred())
		before, err := clicky.Format(comments, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		view, err := renderComments(comments, documentRenderOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.TrimSpace(view.Markdown())).To(Equal("## Comments\n\n**Alex commented 2026-01-02 03:04 +02:00**\n\nFirst comment\n\n<hr/>\n\n**Blair commented 2026-01-02 04:05 +00:00**\n\n**Second comment**\n\n<hr/>"))
		Expect(strings.Count(view.HTML(), "<hr/>")).To(Equal(len(comments)))
		Expect(view.HTML()).To(SatisfyAll(ContainSubstring("<h2>Comments</h2>"), Not(ContainSubstring("<h3>")), Not(ContainSubstring("<img"))))
		Expect(view.String()).To(SatisfyAll(ContainSubstring("Alex commented"), ContainSubstring("Second comment"), Not(ContainSubstring("<hr"))))
		issue := Issue{Issue: externalissues.Issue{Key: "EXAMPLE-1", Title: "Comment headings", Comments: comments}}
		full, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Count(full.Markdown(), "## Comments")).To(Equal(1))
		after, err := clicky.Format(comments, clicky.FormatOptions{Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(MatchJSON(before))
	})

	It("uses an available author ID and omits an unavailable creation date", func() {
		comments, err := ParseComments(json.RawMessage(`{"comments":[{"author":{"accountId":"user-1"},"body":"By ID"},{"body":"Anonymous"}]}`))
		Expect(err).NotTo(HaveOccurred())
		view, err := renderComments(comments, documentRenderOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(SatisfyAll(ContainSubstring("**user-1 commented**"), ContainSubstring("**Unknown author commented**")))
	})

	It("does not render a Comments section when no comments are loaded", func() {
		view, err := renderComments(nil, documentRenderOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(BeEmpty())
	})
})
