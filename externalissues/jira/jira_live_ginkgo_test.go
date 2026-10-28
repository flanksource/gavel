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
		raw, err := (Client{}).Get(context.Background(), id)
		Expect(err).NotTo(HaveOccurred())
		issue, err := ParseIssue(raw)
		Expect(err).NotTo(HaveOccurred())
		var fields map[string]json.RawMessage
		Expect(json.Unmarshal(raw, &fields)).To(Succeed())
		Expect(issue.Pretty().Markdown()).NotTo(BeEmpty())
		comments, err := ParseComments(raw)
		Expect(err).NotTo(HaveOccurred())
		var thread []json.RawMessage
		Expect(json.Unmarshal(comments.Raw, &thread)).To(Succeed())
		GinkgoWriter.Printf("rendered issue with %d fields and %d comments\n", len(fields), len(thread))
	})
})
