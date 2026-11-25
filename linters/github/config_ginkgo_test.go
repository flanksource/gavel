package githublint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGitHubLinter(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "GitHub linter") }

var _ = Describe("GitHub linter settings", func() {
	var workDir string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		workDir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("lint:\n  linters:\n    github:\n      config:\n        orgs: [acme]\n        repos: [personal/service]\n        excludeRepos: [acme/legacy-*]\n        staleAfter: 240h\n"), 0600)).To(Succeed())
	})
	It("loads settings and applies native CLI overrides", func() {
		l := New(workDir)
		opts, selection, err := l.configuration()
		Expect(err).NotTo(HaveOccurred())
		Expect(selection).To(Equal(github.RepositorySelection{Orgs: []string{"acme"}, Repos: []string{"personal/service"}, ExcludeRepos: []string{"acme/legacy-*"}}))
		Expect(opts.StaleAfter).To(Equal(240 * time.Hour))
		l.ExtraArgs = []string{"--repo", "acme/service", "--stale-after", "48h"}
		opts, selection, err = l.configuration()
		Expect(err).NotTo(HaveOccurred())
		Expect(opts.Client.Repo).To(Equal("acme/service"))
		Expect(opts.StaleAfter).To(Equal(48 * time.Hour))
		Expect(selection.ExcludeRepos).To(Equal([]string{"acme/legacy-*"}))
	})
	It("bypasses repository configuration when invoked for PR status", func() {
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("invalid: ["), 0600)).To(Succeed())
		l := New(workDir)
		l.AuditOptions.PR = &model.PRInfo{}
		opts, selection, err := l.configuration()
		Expect(err).NotTo(HaveOccurred())
		Expect(selection).To(Equal(github.RepositorySelection{}))
		Expect(opts.PR).To(BeIdenticalTo(l.AuditOptions.PR))
	})
	It("preserves a successful skip and its reason through the lint runner", func() {
		l := New(workDir)
		result := linters.RunLinterWithTask(context.Background(), nil, l, linters.RunOptions{WorkDir: workDir, ExtraArgs: []string{"--repo", "acme/legacy-api"}})
		Expect(result.Skipped).To(BeTrue())
		Expect(result.Success).To(BeTrue())
		Expect(result.Error).To(Equal("no repositories remain after selection and exclusions"))
		Expect(result.Violations).To(BeEmpty())
	})
	It("audits the configured union and applies the configured staleness threshold", func() {
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("lint:\n  linters:\n    github:\n      config:\n        orgs: [acme]\n        repos: ['org:acme topic:mission-control', personal/service]\n        excludeRepos: [acme/legacy-*]\n        staleAfter: 240h\n"), 0600)).To(Succeed())
		var repositories []string
		searches := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case r.URL.Path == "/search/repositories":
				searches++
				Expect(r.URL.Query().Get("q")).To(Equal("org:acme topic:mission-control"))
				fmt.Fprint(w, `{"total_count":1,"incomplete_results":false,"items":[{"full_name":"personal/service"}]}`)
			case r.URL.Path == "/orgs/acme/repos":
				fmt.Fprint(w, `[{"full_name":"acme/service"},{"full_name":"acme/legacy-api"}]`)
			case r.URL.Path == "/graphql":
				var req struct {
					Query     string
					Variables map[string]any
				}
				Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
				if strings.Contains(req.Query, "pullRequests(states") {
					Expect(req.Variables).To(HaveKeyWithValue("owner", BeAssignableToTypeOf("")))
					Expect(req.Variables).To(HaveKeyWithValue("repo", BeAssignableToTypeOf("")))
					repo := req.Variables["owner"].(string) + "/" + req.Variables["repo"].(string)
					repositories = append(repositories, repo)
					if repo == "acme/service" {
						fmt.Fprint(w, `{"data":{"repository":{"pullRequests":{"nodes":[{"id":"pr-1","number":1,"updatedAt":"2026-01-01T00:00:00Z","commits":{"nodes":[{"commit":{"id":"head-1"}}]}}],"pageInfo":{"hasNextPage":false}}}}}`)
						return
					}
					fmt.Fprint(w, `{"data":{"repository":{"pullRequests":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}`)
					return
				}
				fmt.Fprint(w, `{"data":{"node":{"statusCheckRollup":{"contexts":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}}}`)
			default:
				Expect(r.URL.Path).To(Or(Equal("/repos/acme/service/dependabot/alerts"), Equal("/repos/acme/service/code-scanning/alerts"), Equal("/repos/personal/service/dependabot/alerts"), Equal("/repos/personal/service/code-scanning/alerts")))
				fmt.Fprint(w, `[]`)
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		l := New(workDir)
		l.AuditOptions = github.AuditOptions{Client: github.Options{Token: "test-token"}, Now: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)}
		result := linters.RunLinterWithTask(context.Background(), nil, l, linters.RunOptions{WorkDir: workDir})
		Expect(result.Error).To(BeEmpty())
		Expect(result.Success).To(BeTrue())
		Expect(repositories).To(Equal([]string{"acme/service", "personal/service"}))
		Expect(searches).To(Equal(1))
		Expect(result.Violations).To(HaveLen(1))
		Expect(result.Violations[0].Rule.Method).To(Equal("pr-stale"))
		Expect(*result.Violations[0].Message).To(HavePrefix("acme/service: "))
	})
})
