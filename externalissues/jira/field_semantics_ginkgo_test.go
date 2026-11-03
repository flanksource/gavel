package jira

import (
	"encoding/json"

	"github.com/flanksource/clicky/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira badge and person fields", func() {
	It("normalizes avatar URLs for issue people, comment authors and attachment authors", func() {
		person := `{"accountId":"account-1","displayName":"Alex","self":"https://example.test/user/account-1","avatarUrls":{"48x48":"https://example.test/avatar.png"}}`
		issue, err := ParseIssue(json.RawMessage(`{"key":"EXAMPLE-1","summary":"Avatars","assignee":`+person+`,"reporter":`+person+`,"comments":[{"author":`+person+`,"body":"Ready"}],"attachment":[{"filename":"report.pdf","author":`+person+`}]}`), ParseOptions{})
		Expect(err).NotTo(HaveOccurred())
		for _, value := range []string{issue.Assignee.AvatarURL, issue.Reporter.AvatarURL, issue.Comments[0].Author.AvatarURL, issue.Attachments[0].Author.AvatarURL} {
			Expect(value).To(Equal("https://example.test/avatar.png"))
		}
		data, err := json.Marshal(issue.Assignee)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(MatchJSON(`{"id":"account-1","name":"Alex","url":"https://example.test/user/account-1","avatar_url":"https://example.test/avatar.png"}`))
	})

	It("renders status categories and nested statuses as HTML badges in both field views", func() {
		for _, raw := range []string{
			`{"colorName":"blue-gray","id":2,"key":"new","name":"To Do","self":"https://example.test/statuscategory/2"}`,
			`{"name":"Open","id":"1","statusCategory":{"colorName":"blue-gray","key":"new","name":"To Do"}}`,
		} {
			view, err := renderFieldValue(json.RawMessage(raw))
			Expect(err).NotTo(HaveOccurred())
			for _, value := range []api.Textable{view.PrettyShort(), view.PrettyFull()} {
				for _, output := range []string{value.HTML(), value.Markdown()} {
					Expect(output).To(SatisfyAll(ContainSubstring(`<span`), ContainSubstring(`class="badge`), ContainSubstring("background-color:"), Not(ContainSubstring("colorName")), Not(ContainSubstring("**id**")), Not(ContainSubstring("statuscategory/2"))))
				}
			}
		}
	})

	It("renders creator identities as avatar and name without expanding account metadata", func() {
		view, err := renderFieldValue(json.RawMessage(`{"accountId":"account-1","displayName":"Alex","self":"https://example.test/user/account-1","active":true,"avatarUrls":{"48x48":"https://example.test/avatar48.png","24x24":"https://example.test/avatar24.png"}}`))
		Expect(err).NotTo(HaveOccurred())
		for _, value := range []api.Textable{view.PrettyShort(), view.PrettyFull()} {
			Expect(value.String()).To(Equal("Alex"))
			Expect(value.Markdown()).To(SatisfyAll(ContainSubstring(`<img src="https://example.test/avatar24.png"`), ContainSubstring(`width="24"`), ContainSubstring("Alex"), Not(ContainSubstring("avatarUrls")), Not(ContainSubstring("**active**"))))
			Expect(value.HTML()).To(SatisfyAll(ContainSubstring(`<img`), ContainSubstring(`src="https://example.test/avatar24.png"`), ContainSubstring(`width="24"`), ContainSubstring("Alex"), Not(ContainSubstring("accountId"))))
		}
	})
})
