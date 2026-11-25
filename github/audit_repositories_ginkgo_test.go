package github

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("GitHub repository selection", func() {
	It("uses the checkout origin when no selectors are configured", func() {
		checkout := newConflictRepo("git@github.com:Acme/Service.git")
		repos, err := (Options{WorkDir: checkout.dir}).AuditRepositories(context.Background(), RepositorySelection{})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(Equal([]string{"acme/service"}))
	})
	It("fails when neither selectors nor a checkout origin are available", func() {
		_, err := (Options{WorkDir: GinkgoT().TempDir()}).AuditRepositories(context.Background(), RepositorySelection{})
		Expect(err).To(HaveOccurred())
	})
	It("combines paginated organizations and explicit repos then excludes and deduplicates", func() {
		var server *httptest.Server
		calls := 0
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			calls++
			Expect(r.URL.Path).To(Equal("/orgs/acme/repos"))
			if calls == 1 {
				w.Header().Set("Link", fmt.Sprintf(`<%s/orgs/acme/repos?page=2>; rel="next"`, server.URL))
				rows := make([]map[string]any, 100)
				for i := range rows {
					rows[i] = map[string]any{"full_name": fmt.Sprintf("acme/service-%03d", i)}
				}
				Expect(json.NewEncoder(w).Encode(rows)).To(Succeed())
				return
			}
			fmt.Fprint(w, `[{"full_name":"acme/last","archived":true,"fork":true},{"full_name":"acme/legacy-api"}]`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		repos, err := (Options{Token: "test-token"}).AuditRepositories(context.Background(), RepositorySelection{Orgs: []string{"acme", "ACME"}, Repos: []string{"ACME/last", "personal/service", "ignored-org/service"}, ExcludeRepos: []string{"ACME/legacy-*", "ignored-org/*"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(HaveLen(102))
		Expect(repos).To(ContainElements("acme/last", "personal/service"))
		Expect(calls).To(Equal(2))
	})
	It("applies exclusions to an explicit CLI repo without organization discovery", func() {
		repos, err := (Options{Repo: "acme/service"}).AuditRepositories(context.Background(), RepositorySelection{Orgs: []string{"inaccessible"}, Repos: []string{"personal/service"}, ExcludeRepos: []string{"acme/*"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(repos).To(BeEmpty())
	})
	It("retains explicit repositories when an organization is inaccessible", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		repos, err := (Options{Token: "test-token"}).AuditRepositories(context.Background(), RepositorySelection{Orgs: []string{"acme"}, Repos: []string{"personal/service"}})
		Expect(err).To(HaveOccurred())
		Expect(repos).To(Equal([]string{"personal/service"}))
	})
})
