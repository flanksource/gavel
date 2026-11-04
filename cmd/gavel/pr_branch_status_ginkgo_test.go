package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("repository-scoped default branch status", func() {
	BeforeEach(func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			var request struct {
				Query     string         `json:"query"`
				Variables map[string]any `json:"variables"`
			}
			Expect(json.NewDecoder(r.Body).Decode(&request)).To(Succeed())
			Expect(request.Query).To(ContainSubstring("repository(owner:"))
			Expect(request.Variables).To(HaveKeyWithValue("owner", "acme"))
			Expect(request.Variables).To(HaveKeyWithValue("name", "widgets"))
			fmt.Fprint(w, `{"data":{"repository":{"nameWithOwner":"acme/widgets","url":"https://github.com/acme/widgets","defaultBranchRef":{"name":"main","target":{"oid":"merged-sha","statusCheckRollup":{"state":"SUCCESS","contexts":{"nodes":[{"__typename":"CheckRun","name":"deploy","status":"COMPLETED","conclusion":"SUCCESS","detailsUrl":"https://github.com/acme/widgets/actions/runs/101/job/7"}]}}}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
		GinkgoT().Setenv("GAVEL_DB", "off")
	})

	It("queries only the selected repository and retains its commit and check URLs", func() {
		result, err := runOrgBranchStatus(PRListOptions{Org: "acme", Repos: []string{"acme/widgets"}})
		Expect(err).NotTo(HaveOccurred())
		status := result.(github.OrgBranchStatusResults)
		Expect(status.Repos).To(HaveLen(1))
		Expect(status.Repos[0].HeadSHA).To(Equal("merged-sha"))
		Expect(status.Repos[0].Checks).To(HaveLen(1))
		Expect(status.Repos[0].Checks[0].DetailsURL).To(Equal("https://github.com/acme/widgets/actions/runs/101/job/7"))
	})

	It("applies the failing-only filter to selected repositories", func() {
		result, err := runOrgBranchStatus(PRListOptions{Org: "acme", Repos: []string{"acme/widgets"}, FailedOnly: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.(github.OrgBranchStatusResults).Repos).To(BeEmpty())
	})
})
