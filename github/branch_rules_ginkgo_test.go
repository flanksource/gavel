package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("effective branch rules", func() {
	It("retains rules and opaque parameters beyond the first page", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.URL.Path).To(Equal("/repos/acme/widgets/rules/branches/release/stable"))
			Expect(r.URL.Query().Get("per_page")).To(Equal("100"))
			switch r.URL.Query().Get("page") {
			case "1":
				fmt.Fprint(w, "["+strings.TrimSuffix(strings.Repeat(`{"type":"update"},`, 100), ",")+"]")
			case "2":
				fmt.Fprint(w, `[{"type":"required_workflows","ruleset_source_type":"Organization","parameters":{"future_setting":true}}]`)
			default:
				Fail("unexpected rules page")
			}
		}))
		defer server.Close()
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GAVEL_DB", "off")
		rules, err := FetchBranchRules(Options{Repo: "acme/widgets", Token: "test-token"}, "release/stable")
		Expect(err).NotTo(HaveOccurred())
		Expect(rules).To(HaveLen(101))
		Expect(string(rules[100])).To(MatchJSON(`{"type":"required_workflows","ruleset_source_type":"Organization","parameters":{"future_setting":true}}`))
	})
})
