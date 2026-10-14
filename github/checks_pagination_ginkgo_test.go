package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("complete head check evidence", func() {
	It("retains failing checks beyond the first page and missing required contexts", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"data":{"node":{"statusCheckRollup":{"contexts":{"nodes":[{"__typename":"CheckRun","name":"later failure","status":"COMPLETED","conclusion":"FAILURE"}]}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		pr := &graphQLPR{Commits: graphQLCommits{Nodes: []graphQLCommitNode{{Commit: graphQLCommit{ID: "commit-node", StatusCheckRollup: &graphQLStatusCheckRollup{Contexts: graphQLContexts{PageInfo: graphQLPageInfo{HasNextPage: true, EndCursor: "cursor"}}}}}}}}
		Expect(completeHeadChecks("token", pr)).To(Succeed())
		Expect(pr.toPRInfo().StatusCheckRollup.HasFailure()).To(BeTrue())
		Expect(pr.toPRInfo().StatusCheckRollup).To(HaveLen(1))
		Expect(json.Unmarshal([]byte(`{"branchProtectionRule":{"requiredStatusCheckContexts":["later failure","required test"]}}`), &pr.BaseRef)).To(Succeed())
		checks := pr.toPRInfo().StatusCheckRollup
		Expect(checks).To(HaveLen(2))
		Expect(checks.AllComplete()).To(BeFalse())
		Expect(checks[1].Name).To(Equal("required test"))
		Expect(checks[1].Status).To(Equal("QUEUED"))
	})
})
