package github

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitHub repository queries", func() {
	const query = "org:acme topic:mission-control"
	It("paginates queries then merges, deduplicates and excludes repositories", func() {
		var server *httptest.Server
		calls := 0
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			calls++
			Expect(r.URL.Path).To(Equal("/search/repositories"))
			Expect(r.URL.Query().Get("q")).To(Equal(query))
			if calls == 1 {
				w.Header().Set("Link", fmt.Sprintf(`<%s/search/repositories?q=org%%3Aacme+topic%%3Amission-control&page=2>; rel="next"`, server.URL))
				fmt.Fprint(w, `{"total_count":3,"incomplete_results":false,"items":[{"full_name":"ACME/service"},{"full_name":"acme/old-api"}]}`)
			} else {
				Expect(r.URL.Query().Get("page")).To(Equal("2"))
				fmt.Fprint(w, `{"total_count":3,"incomplete_results":false,"items":[{"full_name":"acme/extra"}]}`)
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		repos, err := (Options{Token: "test-token"}).AuditRepositories(context.Background(), RepositorySelection{Repos: []string{query, "acme/service", query}, ExcludeRepos: []string{"ACME/old-*"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(Equal([]string{"acme/extra", "acme/service"}))
		Expect(calls).To(Equal(2))
	})
	DescribeTable("surfaces search failures and retains other selections", func(status int, body string, want string) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status); fmt.Fprint(w, body) }))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		repos, err := (Options{Token: "test-token"}).AuditRepositories(context.Background(), RepositorySelection{Repos: []string{query, "personal/service"}})
		Expect(err).To(MatchError(ContainSubstring(want)))
		Expect(repos).To(ContainElement("personal/service"))
	}, Entry("API failure", 422, `{"message":"Validation failed"}`, "HTTP 422"),
		Entry("incomplete", 200, `{"total_count":1,"incomplete_results":true,"items":[{"full_name":"acme/service"}]}`, "incomplete"),
		Entry("capped", 200, `{"total_count":1001,"incomplete_results":false,"items":[{"full_name":"acme/service"}]}`, "1,000"),
		Entry("invalid response", 200, `{"total_count":1,"items":null}`, "invalid"))
	It("does not fall back to the checkout when a query has no matches", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{"total_count":0,"incomplete_results":false,"items":[]}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		repos, err := (Options{Token: "test-token", WorkDir: GinkgoT().TempDir()}).AuditRepositories(context.Background(), RepositorySelection{Repos: []string{query}})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(BeEmpty())
	})
	It("bypasses queries for an explicit CLI repository", func() {
		repos, err := (Options{Repo: "personal/service"}).AuditRepositories(context.Background(), RepositorySelection{Repos: []string{query}})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(Equal([]string{"personal/service"}))
	})
	It("honors cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := (Options{Token: "test-token"}).AuditRepositories(ctx, RepositorySelection{Repos: []string{query}})
		Expect(err).To(HaveOccurred())
	})
})
