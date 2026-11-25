package github

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

var _ = Describe("GitHub audit", func() {
	It("follows alert links and preserves CodeQL branch locations", func() {
		calls := 0
		var server *httptest.Server
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			calls++
			if calls == 1 {
				w.Header().Set("Link", fmt.Sprintf(`<%s/repos/acme/service/code-scanning/alerts?after=next>; rel="next"`, server.URL))
			} else {
				Expect(r.URL.Query().Get("after")).To(Equal("next"))
			}
			fmt.Fprintf(w, `[{"number":%d,"state":"open","tool":{"name":"CodeQL"},"rule":{"id":"go/path-injection","description":"Unsafe path","severity":"warning"},"html_url":"https://github.com/acme/service/security/code-scanning/1","most_recent_instance":{"ref":"refs/heads/main","message":{"text":"Untrusted value"},"location":{"path":"service.go","start_line":3,"end_line":4}}}]`, calls)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		found, err := fetchAuditAlerts(Options{Repo: "acme/service", Token: "test-token"}, "code-scanning")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(HaveLen(2))
		Expect(found[0].Severity).To(Equal(models.SeverityWarning))
		Expect(found[0].Locations[0].Ref).To(Equal("refs/heads/main"))
		Expect(found[0].Line).To(Equal(3))
		Expect(found[0].EndLine).To(Equal(4))
	})
	It("honors cancellation", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := Audit(ctx, AuditOptions{Client: Options{Repo: "acme/service", Token: "test-token"}, PR: &model.PRInfo{NodeID: "pr-node", StatusCheckRollup: model.StatusChecks{{NodeID: "check-node", Name: "CodeQL", AppSlug: "github-advanced-security"}}}})
		Expect(err).To(HaveOccurred())
	})
	It("reports repository PR health and security alerts with partial-source errors", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if strings.Contains(r.URL.Path, "dependabot") {
				fmt.Fprint(w, `[{"number":1,"state":"open","html_url":"https://github.com/acme/service/security/dependabot/1","dependency":{"manifest_path":"go.mod","package":{"name":"example.org/dependency","ecosystem":"go"}},"security_advisory":{"ghsa_id":"GHSA-example","summary":"Unsafe input"},"security_vulnerability":{"vulnerable_version_range":"< 2.0","first_patched_version":{"identifier":"2.0"}}}]`)
				return
			}
			if strings.Contains(r.URL.Path, "code-scanning") {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"message":"Forbidden"}`)
				return
			}
			var req struct{ Query string }
			Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
			if strings.Contains(req.Query, "pullRequests(states") {
				fmt.Fprint(w, `{"data":{"repository":{"pullRequests":{"nodes":[{"id":"pr-1","number":1,"title":"Fix service","url":"https://github.com/acme/service/pull/1","mergeable":"CONFLICTING","updatedAt":"2026-01-01T00:00:00Z","commits":{"nodes":[{"commit":{"id":"head-1"}}]}}],"pageInfo":{"hasNextPage":false}}}}}`)
				return
			}
			fmt.Fprint(w, `{"data":{"node":{"statusCheckRollup":{"contexts":{"nodes":[{"__typename":"CheckRun","name":"Test","conclusion":"FAILURE","isRequired":false}],"pageInfo":{"hasNextPage":false}}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		found, err := Audit(context.Background(), AuditOptions{Client: Options{Repo: "acme/service", Token: "test-token"}, Now: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)})
		Expect(err).To(MatchError(ContainSubstring("code-scanning alerts HTTP 403")))
		Expect(found).To(HaveLen(4))
		Expect(found[0].Rule.Method).To(Equal("pr-stale"))
		Expect(found[1].Rule.Method).To(Equal("pr-conflicting"))
		Expect(found[2].Severity).To(Equal(models.SeverityWarning))
		Expect(found[3].File).To(Equal("go.mod"))
		Expect(*found[3].Message).To(ContainSubstring("patched in 2.0"))
	})
	It("retains PRs and checks beyond the first page", func() {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			calls++
			var req struct {
				Query     string
				Variables map[string]any
			}
			Expect(json.NewDecoder(r.Body).Decode(&req)).To(Succeed())
			if strings.Contains(req.Query, "pullRequests(states") {
				if calls == 1 {
					fmt.Fprint(w, `{"data":{"repository":{"pullRequests":{"nodes":[{"number":1}],"pageInfo":{"hasNextPage":true,"endCursor":"next"}}}}}`)
					return
				}
				Expect(req.Variables["after"]).To(Equal("next"))
				fmt.Fprint(w, `{"data":{"repository":{"pullRequests":{"nodes":[{"number":2}],"pageInfo":{"hasNextPage":false}}}}}`)
				return
			}
			if calls == 3 {
				fmt.Fprint(w, `{"data":{"node":{"statusCheckRollup":{"contexts":{"nodes":[],"pageInfo":{"hasNextPage":true,"endCursor":"later"}}}}}}`)
				return
			}
			Expect(req.Variables["after"]).To(Equal("later"))
			fmt.Fprint(w, `{"data":{"node":{"statusCheckRollup":{"contexts":{"nodes":[{"__typename":"CheckRun","name":"Test","isRequired":true,"conclusion":"FAILURE"}],"pageInfo":{"hasNextPage":false}}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		opts := Options{Repo: "acme/service", Token: "test-token"}
		prs, err := fetchAuditPRs(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(prs).To(HaveLen(2))
		Expect(json.Unmarshal([]byte(`{"id":"pr-1","number":1,"commits":{"nodes":[{"commit":{"id":"head-1"}}]}}`), &prs[0])).To(Succeed())
		checks, err := fetchAuditChecks(opts, prs[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(checks).To(HaveLen(1))
		Expect(*checks[0].Required).To(BeTrue())
	})
	DescribeTable("scopes CodeQL findings and uses the originating check requirement", func(required bool, want models.ViolationSeverity) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			calls++
			Expect(r.URL.Path).To(Equal("/graphql"))
			if calls == 1 {
				fmt.Fprintf(w, `{"data":{"node":{"isRequired":%t}}}`, required)
				return
			}
			fmt.Fprint(w, `{"data":{"node":{"annotations":{"nodes":[{"path":"service.go","message":"Unsafe path","annotationLevel":"FAILURE","location":{"start":{"line":1},"end":{"line":1}}}],"pageInfo":{"hasNextPage":false}}}}}`)
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		pr := &model.PRInfo{NodeID: "pr-node", Number: 7, StatusCheckRollup: model.StatusChecks{
			{NodeID: "codeql-node", Name: "CodeQL", AppSlug: "github-advanced-security", Conclusion: "FAILURE"},
			{NodeID: "test-node", Name: "Test", Conclusion: "FAILURE"},
		}}
		found, err := Audit(context.Background(), AuditOptions{Client: Options{Token: "test-token"}, PR: pr})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(HaveLen(1))
		Expect(found[0].Severity).To(Equal(want))
		Expect(calls).To(Equal(2))
	}, Entry("required", true, models.SeverityError), Entry("optional despite another failing check", false, models.SeverityWarning))
})
