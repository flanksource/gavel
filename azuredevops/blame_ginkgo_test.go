package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/pr/blame"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure PR reviewer recommendations", func() {
	var dir, base, head, snapshotHead string
	var client *Client
	var policyStatus, policyResponse string
	var policyHTTP int
	var requests []string
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "%s", out)
		return strings.TrimSpace(string(out))
	}
	write := func(value string) {
		Expect(os.WriteFile(filepath.Join(dir, "source.txt"), []byte(value), 0644)).To(Succeed())
	}
	commit := func(name, email string) string {
		git("add", "--", "source.txt")
		git("-c", "user.name="+name, "-c", "user.email="+email, "commit", "-qm", "change")
		return git("rev-parse", "HEAD")
	}
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		git("init", "-q", "-b", "feature")
		git("config", "commit.gpgsign", "false")
		git("remote", "add", "origin", "https://dev.azure.com/acme/product/_git/service")
		write("one\ntwo\nthree\n")
		commit("Alice", "alice@example.com")
		write("one\nbob\nthree\n")
		commit("Bob", "bob@example.com")
		write("one\nbob\nauthor\n")
		base = commit("Author", "author@example.com")
		write("one\nchanged\nnew\n")
		head = commit("Author", "author@example.com")
		snapshotHead = head
		write("uncommitted edits\n")
		requests, policyStatus = nil, "approved"
		policyHTTP = http.StatusOK
		policyResponse = ""
		client = testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal(http.MethodGet))
			requests = append(requests, r.URL.Path)
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprintf(w, `{"pullRequestId":27,"title":"Update source","status":"active","sourceRefName":"refs/heads/feature","targetRefName":"refs/heads/main","lastMergeSourceCommit":{"commitId":%q},"lastMergeTargetCommit":{"commitId":%q},"createdBy":{"id":"author-id","displayName":"Author","uniqueName":"author@example.com"}}`, snapshotHead, base)
			case strings.HasSuffix(r.URL.Path, "/pullrequests"):
				Expect(r.URL.Query().Get("searchCriteria.sourceRefName")).To(Equal("refs/heads/feature"))
				fmt.Fprint(w, `{"value":[{"pullRequestId":27}]}`)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				ref, sha := "refs/heads/main", base
				if r.URL.Query().Get("filter") == "heads/feature" {
					ref, sha = "refs/heads/feature", head
				}
				Expect(json.NewEncoder(w).Encode(map[string]any{"value": []any{map[string]string{"name": ref, "objectId": sha}}})).To(Succeed())
			case strings.HasSuffix(r.URL.Path, "/reviewers"):
				fmt.Fprint(w, `{"value":[{"id":"bob-id","displayName":"Bob","uniqueName":"bob@example.com"},{"id":"team-id","displayName":"Platform team","isContainer":true},{"id":"author-id","displayName":"Author","uniqueName":"author@example.com"}]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				Expect(r.URL.Query().Get("artifactId")).To(Equal("vstfs:///CodeReview/CodeReviewId/project-id/27"))
				if policyHTTP != http.StatusOK {
					w.WriteHeader(policyHTTP)
					fmt.Fprint(w, `{"message":"Policy read denied"}`)
					return
				}
				if policyResponse != "" {
					fmt.Fprint(w, policyResponse)
					return
				}
				fmt.Fprintf(w, `{"value":[{"status":%q,"configuration":{"id":17,"isEnabled":true,"isBlocking":true,"type":{"id":"fd2167ab-b0be-447a-8ec8-39368250530e","displayName":"Required reviewers"},"settings":{"requiredReviewerIds":["bob-id","team-id","author-id"],"filenamePatterns":["/source.txt"]}}}]}`, policyStatus)
			default:
				Expect(strings.Contains(r.URL.Path, "/git/repositories/")).To(BeTrue(), r.URL.Path)
				fmt.Fprint(w, fixtureRepository)
			}
		})
	})
	options := func() blame.Request {
		return blame.Request{Context: context.Background(), Dir: dir, Number: 27, Limit: 5, HistoryLimit: 100}
	}

	It("combines applicable reviewer policies and Git history while excluding the PR author", func() {
		result, err := client.RecommendReviewers(options())
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Provider).To(Equal("azuredevops"))
		Expect(result.Reviewers).To(HaveLen(3))
		Expect(result.Reviewers[0].Email).To(Equal("bob@example.com"))
		Expect(result.Reviewers[0].Lines).To(Equal(1))
		Expect(result.Reviewers[0].Commits).To(Equal(1))
		Expect(result.Reviewers[0].Policies).To(Equal([]blame.Policy{{ID: 17, Name: "Required reviewers", Required: true, Status: "approved", Patterns: []string{"/source.txt"}}}))
		Expect(result.Reviewers[1].ID).To(Equal("team-id"))
		Expect(result.Reviewers[1].Name).To(Equal("Platform team"))
		Expect(result.Reviewers[1].Files).To(BeEmpty())
		Expect(result.Reviewers[2].Email).To(Equal("alice@example.com"))
		Expect(result.Pretty().String()).To(ContainSubstring("Required reviewers"))
		Expect(result.Pretty().String()).NotTo(ContainSubstring("author@example.com"))
		content, err := os.ReadFile(filepath.Join(dir, "source.txt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal("uncommitted edits\n"))
		Expect(git("rev-parse", "HEAD")).To(Equal(head))
	})

	It("resolves the current branch's PR and preserves limit and email exclusion semantics", func() {
		opts := options()
		opts.Number, opts.Limit, opts.Exclude = 0, 1, []string{"bob@example.com"}
		result, err := client.RecommendReviewers(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Number).To(Equal(27))
		Expect(result.TotalReviewers).To(Equal(2))
		Expect(result.Reviewers).To(HaveLen(1))
		Expect(result.Reviewers[0].ID).To(Equal("team-id"))
	})

	It("does not recommend reviewers from policies ADO says do not apply", func() {
		policyStatus = "notApplicable"
		result, err := client.RecommendReviewers(options())
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Reviewers).To(HaveLen(2))
		Expect(result.Reviewers[0].Email).To(Equal("bob@example.com"))
		Expect(result.Reviewers[0].Policies).To(BeEmpty())
	})

	It("recommends queued human reviewers while preserving the pending evaluation status", func() {
		policyStatus = "queued"
		result, err := client.RecommendReviewers(options())
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Reviewers[0].Policies[0].Status).To(Equal("queued"))
	})

	It("does not silently ignore malformed reviewer settings", func() {
		policyResponse = `{"value":[{"status":"approved","configuration":{"id":17,"isEnabled":true,"isBlocking":false,"type":{"id":"fd2167ab-b0be-447a-8ec8-39368250530e","displayName":"Required reviewers"},"settings":{}}}]}`
		_, err := client.RecommendReviewers(options())
		Expect(err).To(MatchError(ContainSubstring("requiredReviewerIds")))
	})

	It("refuses an unrelated local checkout before reading PR policies", func() {
		git("remote", "set-url", "origin", "https://dev.azure.com/acme/product/_git/other")
		_, err := client.RecommendReviewers(options())
		Expect(err).To(MatchError(ContainSubstring("no git remote points at")))
		Expect(requests).NotTo(ContainElement(ContainSubstring("evaluations")))
	})

	It("refuses policy evidence when the source branch has advanced beyond the evaluated PR snapshot", func() {
		snapshotHead = base
		_, err := client.RecommendReviewers(options())
		Expect(err).To(MatchError(ContainSubstring("advanced beyond its merge snapshot")))
		Expect(requests).NotTo(ContainElement(ContainSubstring("evaluations")))
	})

	It("keeps optional reviewer policies distinguishable from required policies", func() {
		policyResponse = `{"value":[{"status":"approved","configuration":{"id":18,"isEnabled":true,"isBlocking":false,"type":{"id":"fd2167ab-b0be-447a-8ec8-39368250530e","displayName":"Required reviewers"},"settings":{"requiredReviewerIds":["team-id"]}}}]}`
		result, err := client.RecommendReviewers(options())
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Reviewers[0].ID).To(Equal("team-id"))
		Expect(result.Reviewers[0].Policies).To(Equal([]blame.Policy{{ID: 18, Name: "Required reviewers", Required: false, Status: "approved"}}))
	})

	It("reports policy permission failures instead of returning history-only recommendations", func() {
		policyHTTP = http.StatusForbidden
		_, err := client.RecommendReviewers(options())
		Expect(err).To(MatchError(ContainSubstring("Policy read denied")))
	})

	It("honors the since cutoff for Git evidence while retaining native reviewer policies", func() {
		opts := options()
		opts.Since = time.Date(2100, time.January, 1, 0, 0, 0, 0, time.UTC)
		result, err := client.RecommendReviewers(opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Reviewers).To(HaveLen(2))
		for _, reviewer := range result.Reviewers {
			Expect(reviewer.Policies).To(HaveLen(1))
			Expect(reviewer.Commits).To(BeZero())
			Expect(reviewer.Lines).To(BeZero())
		}
		Expect(result.Since).To(Equal(&opts.Since))
	})
})
