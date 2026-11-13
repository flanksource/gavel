package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/gavel/prwatch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR status merge rules", func() {
	It("fetches the PR base branch's effective rules when requested", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch r.URL.Path {
			case "/graphql":
				fmt.Fprint(w, `{"data":{"repository":{"pullRequest":{"id":"pr-7","number":7,"state":"OPEN","baseRefName":"main"}}}}`)
			case "/repos/acme/widgets/rules/branches/main":
				fmt.Fprint(w, `[{"type":"pull_request","parameters":{"required_approving_review_count":1}}]`)
			default:
				Fail("unexpected status request: " + r.URL.Path)
			}
		}))
		defer server.Close()
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
		GinkgoT().Setenv("GAVEL_DB", "off")
		DeferCleanup(func() { exitCode = 0 })
		result, err := runPRStatus(PRStatusOptions{Repo: "acme/widgets", Args: []string{"7"}, Rules: true, Interval: "1s"})
		Expect(err).NotTo(HaveOccurred())
		rules := result.(*prwatch.PRWatchResult).PR.MergeRules
		Expect(rules).To(HaveLen(1))
		Expect(string(rules[0])).To(MatchJSON(`{"type":"pull_request","parameters":{"required_approving_review_count":1}}`))
	})
})
