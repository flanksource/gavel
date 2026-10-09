package prcreate

import (
	"context"
	"encoding/json"
	"fmt"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/provider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type azureTestTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (t azureTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u := *req.URL
	clone.URL = &u
	clone.URL.Scheme, clone.URL.Host = t.target.Scheme, t.target.Host
	return t.next.RoundTrip(clone)
}

var _ = Describe("Azure creation orchestration", func() {
	It("rejects an omitted project that resolves to a different origin repository", func() {
		f := newRepoFixture()
		git(f.repo, "remote", "set-url", "origin", "https://dev.azure.com/acme/product/_git/service")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			project := "product"
			if r.URL.Path == "/acme/_apis/git/repositories/service" {
				project = "other"
			}
			fmt.Fprintf(w, `{"id":"repo-id","name":"service","defaultBranch":"refs/heads/main","project":{"id":"project-id","name":%q}}`, project)
		}))
		DeferCleanup(server.Close)
		target, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		old := http.DefaultTransport
		http.DefaultTransport = azureTestTransport{target: target, next: old}
		DeferCleanup(func() { http.DefaultTransport = old })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
		deps := DefaultDeps()
		deps.GenerateContent = func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			Fail("AI must not be called")
			return commitpkg.PRContent{}, nil
		}
		_, err = Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Repo: "https://dev.azure.com/acme/_git/service", Deps: deps})
		Expect(err).To(MatchError(ContainSubstring("does not match origin")))
		Expect(f.scratchEntries()).To(BeEmpty())
	})
	It("pushes ordered cherry-picks to a bare origin and creates a draft with the shared provider", func() {
		f := newRepoFixture()
		git(f.repo, "remote", "set-url", "--push", "origin", f.bare)
		git(f.repo, "remote", "set-url", "origin", "https://dev.azure.com/acme/product/_git/service")
		git(f.repo, "checkout", "feature")
		second := commitFile(f.repo, "second.txt", "second\n", "feat: second")
		git(f.repo, "checkout", "main")
		var payload struct {
			Source string `json:"sourceRefName"`
			Target string `json:"targetRefName"`
			Draft  bool   `json:"isDraft"`
			Title  string `json:"title"`
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if r.Method == http.MethodGet {
				fmt.Fprint(w, `{"id":"repo-id","name":"service","defaultBranch":"refs/heads/main","project":{"id":"project-id","name":"product"}}`)
				return
			}
			Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
			Expect(git(f.bare, "show", payload.Source+":second.txt")).To(Equal("second"))
			fmt.Fprint(w, `{"pullRequestId":27,"title":"feat: feature","status":"active"}`)
		}))
		DeferCleanup(server.Close)
		target, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		old := http.DefaultTransport
		http.DefaultTransport = azureTestTransport{target: target, next: old}
		DeferCleanup(func() { http.DefaultTransport = old })
		GinkgoT().Setenv("AZURE_DEVOPS_EXT_PAT", "fixture-pat")
		deps := DefaultDeps()
		deps.GenerateContent = func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			return commitpkg.PRContent{Title: "feat: feature", Body: "why", Branch: "feat/topic"}, nil
		}
		result, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA, second}, Base: testBase, Draft: true, Deps: deps})
		Expect(err).NotTo(HaveOccurred())
		Expect(payload.Source).To(Equal("refs/heads/" + result.TopicBranch))
		Expect(payload.Target).To(Equal("refs/heads/main"))
		Expect(payload.Draft).To(BeTrue())
		Expect(result.PR.URL).To(Equal("https://dev.azure.com/acme/product/_git/service/pullrequest/27"))
		Expect(git(f.bare, "log", "--format=%s", "-2", result.TopicBranch)).To(Equal("feat: second\nfeat: add feature.txt"))
		Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(f.baseSHA))
		Expect(f.scratchEntries()).To(BeEmpty())
	})
	It("rejects a different repository before creating worktrees or calling AI", func() {
		f := newRepoFixture()
		git(f.repo, "remote", "set-url", "origin", "https://dev.azure.com/acme/product/_git/service")
		deps := DefaultDeps()
		deps.GenerateContent = func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			Fail("AI must not be called")
			return commitpkg.PRContent{}, nil
		}
		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Repo: "https://dev.azure.com/acme/product/_git/other", Deps: deps})
		Expect(err).To(MatchError(ContainSubstring("does not match origin")))
		Expect(f.scratchEntries()).To(BeEmpty())
	})
	It("stops on failed provider preflight without worktrees or pushes", func() {
		f := newRepoFixture()
		deps := failingDeps()
		deps.Preflight = func(context.Context, provider.Options) error { return fmt.Errorf("HTTP 403") }
		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Deps: deps})
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
		Expect(f.scratchEntries()).To(BeEmpty())
		Expect(git(f.bare, "for-each-ref", "--format=%(refname)", "refs/heads")).To(Equal("refs/heads/main"))
	})
})
