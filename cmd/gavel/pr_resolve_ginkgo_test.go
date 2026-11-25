package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel pr resolve", func() {
	var mutations int
	BeforeEach(func() {
		mutations = 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var request struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&request)).To(Succeed())
			if strings.Contains(request.Query, "resolveReviewThread(") {
				mutations++
				Expect(request.Variables["input"]).To(Equal(map[string]any{"threadId": "thread-7"}))
				fmt.Fprint(w, `{"data":{"resolveReviewThread":{"thread":{"id":"thread-7","isResolved":true}}}}`)
				return
			}
			fmt.Fprint(w, `{"data":{"repository":{"pullRequest":{"id":"pr-7","number":7,"state":"OPEN","reviewThreads":{"nodes":[{"id":"thread-7","viewerCanResolve":true,"isResolved":false,"comments":{"nodes":[{"databaseId":11,"body":"Build first"}]}}]}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
		GinkgoT().Setenv("GAVEL_DB", "off")
	})

	It("resolves the selected thread after verifying it belongs to the PR", func() {
		_, err := runPRResolve(PRResolveOptions{Repo: "acme/widgets", Args: []string{"7"}, Thread: "thread-7"})
		Expect(err).NotTo(HaveOccurred())
		Expect(mutations).To(Equal(1))
	})

	It("rejects a thread from another PR without sending a mutation", func() {
		_, err := runPRResolve(PRResolveOptions{Repo: "acme/widgets", Args: []string{"7"}, Thread: "foreign-thread"})
		Expect(err).To(MatchError(ContainSubstring("does not belong")))
		Expect(mutations).To(BeZero())
	})

	It("requires an explicit thread before contacting the provider", func() {
		_, err := runPRResolve(PRResolveOptions{Repo: "acme/widgets", Args: []string{"7"}})
		Expect(err).To(MatchError(ContainSubstring("--thread")))
		Expect(mutations).To(BeZero())
	})
})
