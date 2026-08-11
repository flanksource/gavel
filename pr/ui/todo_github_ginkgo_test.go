package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// todoAliasProvider adds external-reference support to the in-memory test
// provider so the GitHub push handler has the seam it requires, and stands in
// for the durable Captain plan the native store would resolve.
type todoAliasProvider struct {
	*uiTestTODOProvider
	aliases []todos.TodoAlias
	plan    string
}

func (p *todoAliasProvider) Aliases(context.Context, *types.TODO) ([]todos.TodoAlias, error) {
	return p.aliases, nil
}

func (p *todoAliasProvider) AddAlias(_ context.Context, _ *types.TODO, alias todos.TodoAlias) error {
	p.aliases = append(p.aliases, alias)
	return nil
}

func (p *todoAliasProvider) PlanMarkdown(context.Context, *types.TODO, types.RunMode) (string, error) {
	return p.plan, nil
}

// initGitRepoWithOrigin makes dir a git repo whose origin remote is repo, which
// is what the push resolves the target from when no explicit repo is supplied.
func initGitRepoWithOrigin(dir, repo string) {
	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", "https://github.com/" + repo + ".git"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		Expect(cmd.Run()).To(Succeed(), "git %v", args)
	}
}

// stubGitHubAPI stands in for api.github.com, recording the single issue write
// the push makes so the request shape can be asserted.
type stubGitHubAPI struct {
	method  string
	path    string
	payload map[string]any
}

func (s *stubGitHubAPI) start(status int, response string) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.method, s.path = r.Method, r.URL.Path
		body, _ := io.ReadAll(r.Body)
		Expect(json.Unmarshal(body, &s.payload)).To(Succeed())
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	DeferCleanup(server.Close)
	GinkgoT().Setenv("GITHUB_API_URL", server.URL)
	GinkgoT().Setenv("GITHUB_TOKEN", "tok")
}

func postTodoGitHubPush(payload todoGitHubPushPayload) *httptest.ResponseRecorder {
	body, err := json.Marshal(payload)
	Expect(err).NotTo(HaveOccurred())
	request := httptest.NewRequest(http.MethodPost, "/api/todos/github", bytes.NewReader(body))
	recorder := httptest.NewRecorder()
	(&Server{ghOpts: github.Options{Repo: "dashboard/prs"}}).Handler().ServeHTTP(recorder, request)
	return recorder
}

var _ = Describe("todo GitHub push API", func() {
	var (
		originalOpenTodoProvider func(context.Context, string) (todos.Provider, error)
		provider                 *todoAliasProvider
		dir                      string
		todo                     *types.TODO
	)

	BeforeEach(func() {
		originalOpenTodoProvider = openTodoProvider
		dir = filepath.Clean(GinkgoT().TempDir())
		provider = &todoAliasProvider{uiTestTODOProvider: uiTestProviderFor(dir)}
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }

		var err error
		todo, err = provider.Create(GinkgoT().Context(), todos.CreateRequest{
			Title: "Fix the parser", Body: "trailing commas are dropped", Status: types.StatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() { openTodoProvider = originalOpenTodoProvider })

	It("opens the issue on the todo workspace's own remote, not the dashboard's repo", func() {
		initGitRepoWithOrigin(dir, "acme/api")
		api := &stubGitHubAPI{}
		api.start(http.StatusCreated, `{"number":11,"html_url":"https://github.com/acme/api/issues/11"}`)

		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir, BaseURL: "https://gavel.example.com"})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		// The dashboard's own ghOpts.Repo ("dashboard/prs") must not leak in.
		Expect(api.method).To(Equal(http.MethodPost))
		Expect(api.path).To(Equal("/repos/acme/api/issues"))
		Expect(api.payload["title"]).To(Equal("Fix the parser"))

		var response todoGitHubPushResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Repo).To(Equal("acme/api"))
		Expect(response.Number).To(Equal(11))
		Expect(response.URL).To(Equal("https://github.com/acme/api/issues/11"))
		Expect(response.Alias).To(Equal("acme/api#11"))
		Expect(response.Updated).To(BeFalse())
		Expect(provider.aliases).To(ConsistOf(todos.TodoAlias{Alias: "acme/api#11", Kind: githubpush.AliasKind}))
	})

	It("carries the todo's plan into the issue body", func() {
		initGitRepoWithOrigin(dir, "acme/api")
		provider.plan = "1. Rewrite the tokenizer."
		api := &stubGitHubAPI{}
		api.start(http.StatusCreated, `{"number":11,"html_url":"https://github.com/acme/api/issues/11"}`)

		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(api.payload["body"]).To(ContainSubstring("# Plan\n\n1. Rewrite the tokenizer."))
	})

	It("rewrites the linked issue when the request asks for an update", func() {
		provider.aliases = []todos.TodoAlias{{Alias: "acme/api#3", Kind: githubpush.AliasKind}}
		api := &stubGitHubAPI{}
		api.start(http.StatusOK, `{"number":3,"html_url":"https://github.com/acme/api/issues/3"}`)

		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir, Update: true})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(api.method).To(Equal(http.MethodPatch))
		Expect(api.path).To(Equal("/repos/acme/api/issues/3"))

		var response todoGitHubPushResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Updated).To(BeTrue())
		Expect(response.Alias).To(Equal("acme/api#3"))
		// The link already existed, so the push must not record it twice.
		Expect(provider.aliases).To(HaveLen(1))
	})

	It("links and rewrites an issue named explicitly", func() {
		api := &stubGitHubAPI{}
		api.start(http.StatusOK, `{"number":412,"html_url":"https://github.com/acme/api/issues/412"}`)

		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir, Issue: "acme/api#412"})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(api.path).To(Equal("/repos/acme/api/issues/412"))
		Expect(provider.aliases).To(ConsistOf(todos.TodoAlias{Alias: "acme/api#412", Kind: githubpush.AliasKind}))
	})

	It("answers 409 when the todo is already linked", func() {
		provider.aliases = []todos.TodoAlias{{Alias: "acme/api#3", Kind: githubpush.AliasKind}}

		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir})

		Expect(recorder.Code).To(Equal(http.StatusConflict))
		Expect(recorder.Body.String()).To(ContainSubstring("acme/api#3"))
	})

	It("rejects a request without a ref", func() {
		recorder := postTodoGitHubPush(todoGitHubPushPayload{Dir: dir})

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("ref is required"))
	})

	It("rejects a malformed base URL before touching GitHub", func() {
		recorder := postTodoGitHubPush(todoGitHubPushPayload{Ref: todo.ID, Dir: dir, BaseURL: "gavel.example.com"})

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("absolute http(s) origin"))
	})
})
