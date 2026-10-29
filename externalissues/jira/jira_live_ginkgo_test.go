//go:build jira_live

package jira

import (
	"context"
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("live Jira issue rendering", func() {
	It("renders full issue data and its complete comments", func() {
		id := os.Getenv("GAVEL_JIRA_LIVE_ISSUE")
		Expect(id).NotTo(BeEmpty(), "GAVEL_JIRA_LIVE_ISSUE is required")
		client, err := NewClient(context.Background(), ClientOptions{})
		Expect(err).NotTo(HaveOccurred())
		issue, err := client.Get(context.Background(), GetOptions{Key: id, Full: true, NoCache: true, Store: CatalogStore{Dir: GinkgoT().TempDir()}})
		Expect(err).NotTo(HaveOccurred())
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		var exported struct {
			Fields map[string]json.RawMessage `json:"fields"`
		}
		Expect(json.Unmarshal(data, &exported)).To(Succeed())
		Expect(exported.Fields).NotTo(BeEmpty())
		Expect(string(data)).NotTo(ContainSubstring(`"native"`))
		rendered, err := issue.Render(RenderOptions{Full: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(rendered.Markdown()).NotTo(BeEmpty())
		Expect(issue.CommentCount).NotTo(BeNil())
		Expect(*issue.CommentCount).To(Equal(len(issue.Comments)))
		GinkgoWriter.Printf("rendered issue with %d fields and %d comments\n", len(exported.Fields), len(issue.Comments))
	})
})
