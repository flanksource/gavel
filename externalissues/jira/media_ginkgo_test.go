package jira

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/externalissues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mediaDocument(attrs string) *externalissues.Document {
	return &externalissues.Document{Format: "adf", Content: json.RawMessage(fmt.Sprintf(`{"type":"doc","version":1,"content":[{"type":"mediaSingle","content":[{"type":"media","attrs":%s}]}]}`, attrs))}
}

var _ = Describe("Jira inline attachment media", func() {
	var issue Issue
	var warnings bytes.Buffer
	BeforeEach(func() {
		warnings.Reset()
		previous := logger.StandardLogger()
		logger.Use(&warnings)
		DeferCleanup(func() { logger.SetLogger(previous) })
		issue = Issue{Issue: externalissues.Issue{Key: "EXAMPLE-1", Title: "Inline media", Attachments: []externalissues.Attachment{
			{ID: "20001", Name: "screen #1.png", URL: "EXAMPLE-1.attachments/20001/screen%20%231.png", MIMEType: "image/png"},
			{ID: "20002", Name: "report.pdf", URL: "EXAMPLE-1.attachments/20002/report.pdf", MIMEType: "application/pdf"},
		}}}
	})

	It("renders images in descriptions, comments and rich fields without changing structured exports", func() {
		issue.Description = mediaDocument(`{"id":"media-uuid","type":"file","alt":"screen #1.png","collection":""}`)
		issue.Comments = []externalissues.Comment{{ID: "comment-1", Author: externalissues.Person{Name: "Alex"}, Body: issue.Description}}
		issue.Fields = map[string]externalissues.Field{"customfield_1": {ID: "customfield_1", Name: "Screenshot", Format: "markdown", Value: json.RawMessage(`"[screen #1.png](atlassian-media:///media-uuid)"`)}}
		before, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		const image = "![screen \\#1.png](EXAMPLE-1.attachments/20001/screen%20%231.png)"
		Expect(view.Markdown()).To(ContainSubstring(image))
		Expect(view.HTML()).To(ContainSubstring(`<img src="EXAMPLE-1.attachments/20001/screen%20%231.png" alt="screen #1.png"`))
		Expect(view.Markdown()).NotTo(ContainSubstring("atlassian-media:"))
		after, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})

	DescribeTable("resolves native IDs and unique filenames with the correct media kind", func(attrs, expected string) {
		issue.Description = mediaDocument(attrs)
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(ContainSubstring(expected))
		Expect(view.Markdown()).NotTo(ContainSubstring("atlassian-media:"))
	},
		Entry("image UUID and filename", `{"id":"media-uuid","type":"file","alt":"screen #1.png"}`, "![screen \\#1.png](EXAMPLE-1.attachments/20001/screen%20%231.png)"),
		Entry("native image ID without alt", `{"id":"20001","type":"file"}`, "![screen \\#1.png](EXAMPLE-1.attachments/20001/screen%20%231.png)"),
		Entry("non-image file", `{"id":"file-uuid","type":"file","alt":"report.pdf"}`, "[report.pdf](EXAMPLE-1.attachments/20002/report.pdf)"),
	)

	It("rejects ambiguous filenames instead of choosing an image", func() {
		issue.Description = mediaDocument(`{"id":"media-uuid","type":"file","alt":"screen #1.png"}`)
		issue.Attachments = append(issue.Attachments, externalissues.Attachment{ID: "20003", Name: "screen #1.png", URL: "other.png", MIMEType: "image/png"})
		_, err := issue.Render(RenderOptions{Full: true})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("ambiguous"), ContainSubstring("screen #1.png"), ContainSubstring("media-uuid"))))
	})

	It("uses a native ID even when filenames are duplicated", func() {
		issue.Description = mediaDocument(`{"id":"20001","type":"file","alt":"screen #1.png"}`)
		issue.Attachments = append(issue.Attachments, externalissues.Attachment{ID: "20003", Name: "screen #1.png", URL: "other.png", MIMEType: "image/png"})
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(ContainSubstring("![screen \\#1.png](EXAMPLE-1.attachments/20001/screen%20%231.png)"))
	})

	It("retains unresolved media references and ordinary user-authored image links", func() {
		issue.Description = mediaDocument(`{"id":"missing-uuid","type":"file","alt":"removed.png"}`)
		issue.Comments = []externalissues.Comment{{ID: "comment-1", Body: &externalissues.Document{Format: "adf", Content: json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"screenshot","marks":[{"type":"link","attrs":{"href":"https://example.test/screen.png"}}]}]}]}`)}}}
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(SatisfyAll(ContainSubstring("[removed.png](atlassian-media:///missing-uuid)"), ContainSubstring("[screenshot](https://example.test/screen.png)"), Not(ContainSubstring("![screenshot]"))))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("missing-uuid"), ContainSubstring("removed.png")))
	})

	It("resolves images before applying media border formatting", func() {
		issue.Description = &externalissues.Document{Format: "adf", Content: json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"id":"media-uuid","alt":"screen #1.png"},"marks":[{"type":"border","attrs":{"size":2,"color":"#172b4d"}}]}]}]}`)}
		view, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.Markdown()).To(SatisfyAll(ContainSubstring("![screen \\#1.png](EXAMPLE-1.attachments/20001/screen%20%231.png)"), ContainSubstring("border:2px solid #172b4d"), Not(ContainSubstring("atlassian-media:"))))
		Expect(view.HTML()).To(ContainSubstring(`<img src="EXAMPLE-1.attachments/20001/screen%20%231.png"`))
	})

	It("fails when a resolved attachment has no content URL", func() {
		issue.Description = mediaDocument(`{"id":"20001","type":"file"}`)
		issue.Attachments[0].URL = ""
		_, err := issue.Render(RenderOptions{Full: true})
		Expect(err).To(MatchError(ContainSubstring("has no content URL")))
	})
})
