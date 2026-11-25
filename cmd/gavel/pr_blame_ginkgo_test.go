package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/pr/blame"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gavel pr blame", func() {
	var dir, base, head, original, bob string
	var requests []string
	var authorStatus int
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "%s", out)
		return strings.TrimSpace(string(out))
	}
	write := func(path, content string) {
		Expect(os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, path), []byte(content), 0644)).To(Succeed())
	}
	commit := func(author string) string {
		git("add", "--", "source.txt", ".github/CODEOWNERS")
		git("-c", "user.name="+author, "-c", "user.email="+strings.ToLower(author)+"@example.com", "commit", "-qm", author)
		return git("rev-parse", "HEAD")
	}
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		git("init", "-q")
		git("config", "commit.gpgsign", "false")
		git("remote", "add", "origin", "https://github.com/acme/widgets.git")
		write("source.txt", "one\ntwo\nthree\n")
		write(".github/CODEOWNERS", "* @alice @author\n")
		original = commit("Alice")
		write("source.txt", "one\nbob\nthree\n")
		bob = commit("Bob")
		write("source.txt", "one\nbob\nauthor\n")
		base = commit("Author")
		write("source.txt", "one\nupdated\nauthor\n")
		head = commit("Author")
		write("source.txt", "uncommitted\n")
		previous := workingDir
		workingDir = dir
		DeferCleanup(func() { workingDir = previous })
		requests, authorStatus = nil, http.StatusOK
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Or(Equal(http.MethodGet), Equal(http.MethodPost)))
			requests = append(requests, r.URL.Path)
			if r.URL.Path == "/graphql" {
				var request struct {
					Variables map[string]any `json:"variables"`
				}
				Expect(json.NewDecoder(r.Body).Decode(&request)).To(Succeed())
				Expect(request.Variables["owner"]).To(Equal("acme"))
				Expect(request.Variables["repo"]).To(Equal("widgets"))
				Expect(request.Variables["number"]).To(Or(BeNil(), BeNumerically("==", 7)))
				pr := map[string]any{"number": 7, "title": "Update source", "url": "https://github.com/acme/widgets/pull/7", "author": map[string]any{"login": "author"},
					"headRefOid": head, "baseRefName": "main", "baseRef": map[string]any{"target": map[string]any{"oid": base}}}
				Expect(json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"repository": map[string]any{"pullRequest": pr, "pullRequests": map[string]any{"nodes": []any{pr}}}}})).To(Succeed())
				return
			}
			if authorStatus != http.StatusOK {
				w.WriteHeader(authorStatus)
				fmt.Fprint(w, `{"message":"identity lookup denied"}`)
				return
			}
			sha := strings.TrimPrefix(r.URL.Path, "/repos/acme/widgets/commits/")
			login, found := map[string]string{original: "alice", bob: "bob", base: "author"}[sha]
			Expect(found).To(BeTrue(), "unexpected request %s", r.URL.Path)
			Expect(json.NewEncoder(w).Encode(map[string]any{"sha": sha, "author": map[string]any{"login": login}})).To(Succeed())
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		GinkgoT().Setenv("GITHUB_TOKEN", "test-token")
		GinkgoT().Setenv("GAVEL_DB", "off")
	})

	It("resolves PR references and combines GitHub identities with ownership without touching local edits", func() {
		for _, args := range [][]string{nil, {"7"}, {"acme/widgets", "7"}, {"acme/widgets#7"}, {"https://github.com/acme/widgets/pull/7"}} {
			value, err := runPRBlame(PRBlameOptions{Args: args, Limit: 5, HistoryLimit: 100})
			Expect(err).NotTo(HaveOccurred())
			result := value.(*blame.Recommendations)
			Expect(result.Reviewers).To(HaveLen(2))
			Expect(result.Reviewers[0].Login).To(Equal("@alice"))
			Expect(result.Reviewers[0].CodeOwnerFiles).To(Equal(1))
			Expect(result.Reviewers[0].Commits).To(Equal(1))
			Expect(result.Reviewers[1].Login).To(Equal("@bob"))
			Expect(result.Reviewers[1].Lines).To(Equal(1))
			data, err := json.Marshal(result)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring(`"mergeBase":"` + base + `"`))
			Expect(result.Pretty().String()).To(ContainSubstring("source.txt"))
			Expect(result.Pretty().String()).NotTo(ContainSubstring("@author"))
		}
		content, err := os.ReadFile(filepath.Join(dir, "source.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("uncommitted\n"))
		Expect(git("rev-parse", "HEAD")).To(Equal(head))
	})

	It("applies explicit exclusions and reports how many candidates the limit omits", func() {
		value, err := runPRBlame(PRBlameOptions{Args: []string{"7"}, Limit: 1, HistoryLimit: 100})
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(*blame.Recommendations).TotalReviewers).To(Equal(2))
		Expect(value.(*blame.Recommendations).Reviewers).To(HaveLen(1))
		value, err = runPRBlame(PRBlameOptions{Args: []string{"7"}, Limit: 5, HistoryLimit: 100, Exclude: []string{"alice"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(value.(*blame.Recommendations).Reviewers).To(HaveLen(1))
		Expect(value.(*blame.Recommendations).Reviewers[0].Login).To(Equal("@bob"))
	})

	It("fails on identity API errors instead of returning incomplete recommendations", func() {
		authorStatus = http.StatusForbidden
		_, err := runPRBlame(PRBlameOptions{Args: []string{"7"}, Limit: 5, HistoryLimit: 100})
		Expect(err).To(MatchError(ContainSubstring("403")))
	})

	It("rejects invalid limits and non-PR targets before API calls", func() {
		for _, opts := range []PRBlameOptions{
			{Limit: 0, HistoryLimit: 100}, {Limit: 5, HistoryLimit: -1},
			{Limit: 5, HistoryLimit: 100, Args: []string{"https://github.com/acme/widgets/actions/runs/1"}},
			{Limit: 5, HistoryLimit: 100, Since: "invalid-date"},
		} {
			_, err := runPRBlame(opts)
			Expect(err).To(HaveOccurred())
		}
		Expect(requests).To(BeEmpty())
	})

	It("passes the since cutoff to GitHub analysis without filtering out CODEOWNERS", func() {
		value, err := runPRBlame(PRBlameOptions{Args: []string{"7"}, Limit: 5, HistoryLimit: 100, Since: "2100-01-01"})
		Expect(err).NotTo(HaveOccurred())
		result := value.(*blame.Recommendations)
		Expect(result.Reviewers).To(HaveLen(1))
		Expect(result.Reviewers[0].Login).To(Equal("@alice"))
		Expect(result.Reviewers[0].Commits).To(BeZero())
		Expect(result.Since).NotTo(BeNil())
		Expect(result.Since.Format("2006-01-02")).To(Equal("2100-01-01"))
		Expect(requests).To(Equal([]string{"/graphql"}))
	})

	It("refuses a repository that does not match any local remote", func() {
		git("remote", "set-url", "origin", "https://github.com/acme/other.git")
		_, err := runPRBlame(PRBlameOptions{Repo: "acme/widgets", Args: []string{"7"}, Limit: 5, HistoryLimit: 100})
		Expect(err).To(MatchError(ContainSubstring("no git remote points at")))
		Expect(requests).To(Equal([]string{"/graphql"}))
	})
})
