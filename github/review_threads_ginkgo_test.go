package github

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("review thread pagination", func() {
	It("retains native thread IDs and feedback beyond the first page", func() {
		requests := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			requests++
			var query struct {
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&query)).To(Succeed())
			Expect(query.Variables["id"]).To(Equal("PR_node"))
			Expect(query.Variables["after"]).To(Equal("cursor-1"))
			_, err := w.Write([]byte(`{"data":{"node":{"reviewThreads":{"pageInfo":{"hasNextPage":false},"nodes":[{"id":"thread-2","viewerCanResolve":true,"comments":{"nodes":[{"databaseId":2,"body":"Validate empty input"}]}}]}}}}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		pr := &graphQLPR{ID: "PR_node", ReviewThreads: graphQLReviewThreads{PageInfo: graphQLPageInfo{HasNextPage: true, EndCursor: "cursor-1"}, Nodes: []graphQLReviewThreadNode{{ID: "thread-1"}}}}
		Expect(completeReviewThreads("test-token", pr)).To(Succeed())
		Expect(requests).To(Equal(1))
		Expect(pr.ReviewThreads.Nodes).To(HaveLen(2))
		comments := pr.toPRInfo().ReviewThreads
		Expect(comments).To(HaveLen(1))
		Expect(comments[0].ThreadID).To(Equal("thread-2"))
		Expect(comments[0].CanResolve).To(BeTrue())
	})
})
