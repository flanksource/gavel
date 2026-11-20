// Tests an Azure DevOps pull request's commits, changed files and patches, read from git through
// its connection: the commits since the merge base and the patch from it.
package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cgi"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/commons-db/git"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// azureRepoWithPR is a bare repository "rules" whose branch tighten adds rules.go and edits
// README.md on top of main, and returns the directory holding it and tighten's commit.
func azureRepoWithPR(t *testing.T) (string, string) {
	t.Helper()
	root, work := t.TempDir(), t.TempDir()
	run := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Ada Lovelace", "GIT_AUTHOR_EMAIL=ada@example.com",
			"GIT_COMMITTER_NAME=Ada Lovelace", "GIT_COMMITTER_EMAIL=ada@example.com", "GIT_CONFIG_GLOBAL=/dev/null")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
		return strings.TrimSpace(string(out))
	}
	write := func(name, content string) {
		require.NoError(t, os.WriteFile(filepath.Join(work, name), []byte(content), 0o644))
	}
	run(work, "init", "-q", "-b", "main")
	write("README.md", "rules\n")
	run(work, "add", "-A")
	run(work, "commit", "-q", "-m", "Start the rules")
	run(work, "checkout", "-q", "-b", "tighten")
	write("rules.go", "package rules\n")
	write("README.md", "tight rules\n")
	run(work, "add", "-A")
	run(work, "commit", "-q", "-m", "Tighten the rules")
	head := run(work, "rev-parse", "HEAD")
	run(root, "clone", "-q", "--bare", work, filepath.Join(root, "rules"))
	return root, head
}

// fakeAzureWithGit is fakeAzure that also serves the repositories under root over git's smart
// HTTP protocol at /acme/Platform/_git/{name}.
func fakeAzureWithGit(t *testing.T, routes map[string]string, root string) *httptest.Server {
	t.Helper()
	gitPath, err := exec.LookPath("git")
	require.NoError(t, err)
	backend := &cgi.Handler{Path: gitPath, Args: []string{"http-backend"}, Root: "/acme/Platform/_git",
		Env: []string{"GIT_PROJECT_ROOT=" + root, "GIT_HTTP_EXPORT_ALL=1"}}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/acme/Platform/_git/") {
			backend.ServeHTTP(w, r)
			return
		}
		body, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			http.Error(w, "no route "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func diffRoutes() map[string]string {
	return map[string]string{
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9": `{"pullRequestId":9,"status":"active","title":"Tighten the rules",
			"sourceRefName":"refs/heads/tighten","targetRefName":"refs/heads/main",
			"repository":{"id":"rules-id","name":"rules","project":{"name":"Platform"}}}`,
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9/threads":   `{"value":[]}`,
		"GET /acme/Platform/_apis/build/builds":                                    `{"value":[]}`,
		"GET /acme/_apis/projects/Platform":                                        `{"id":"platform-id"}`,
		"GET /acme/Platform/_apis/policy/evaluations":                              `{"value":[]}`,
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9/reviewers": `{"value":[]}`,
		"GET /acme/Platform/_apis/git/repositories/rules-id/refs":                  `{"value":[{"name":"refs/heads/tighten","objectId":"0"},{"name":"refs/heads/main","objectId":"0"}]}`,
	}
}

func TestAnAzurePullRequestsCommitsAndFilesComeFromGit(t *testing.T) {
	root, head := azureRepoWithPR(t)
	server := fakeAzureWithGit(t, diffRoutes(), root)
	remote := azureRemote(t, server)
	cache := &git.Cache{Root: t.TempDir()}

	detail, err := connectionPRDetail(context.Background(), remote, 9, cache)
	require.NoError(t, err)
	pr := detail.PR
	require.Len(t, pr.PRCommits, 1)
	assert.Equal(t, head, pr.PRCommits[0].OID)
	assert.Equal(t, "Tighten the rules", pr.PRCommits[0].MessageHeadline)
	assert.Equal(t, "Ada Lovelace", pr.PRCommits[0].AuthorName)
	assert.Equal(t, []model.PRFileInfo{
		{Path: "README.md", Additions: 1, Deletions: 1, ChangeType: "MODIFIED"},
		{Path: "rules.go", Additions: 1, ChangeType: "ADDED"},
	}, pr.PRFiles)
	assert.Equal(t, 2, pr.Additions)
	assert.Equal(t, 1, pr.Deletions)
	assert.Equal(t, 2, pr.ChangedFiles)
}

func TestTheDashboardServesAnAzurePullRequestsPatchesFromGit(t *testing.T) {
	root, head := azureRepoWithPR(t)
	server := fakeAzureWithGit(t, diffRoutes(), root)
	s := NewServer(gavelctx.New(t.Context()), 0, github.Options{}, SearchConfig{})
	s.publishResults(github.PRSearchResults{{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 9,
		Source: "tighten", Target: "main", URL: server.URL + "/acme/Platform/_git/rules/pullrequest/9"}})
	s.SetPRRemote(func(context.Context, github.PRListItem) (*git.Repository, error) { return azureRemote(t, server), nil })
	s.SetGitCache(&git.Cache{Root: t.TempDir()})
	diff := func(handler http.HandlerFunc, query string) prDiffResponse {
		t.Helper()
		recorder := httptest.NewRecorder()
		handler(recorder, httptest.NewRequest(http.MethodGet, "/api/prs/diff?"+query, nil))
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		var response prDiffResponse
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
		return response
	}

	file := diff(s.handlePRFileDiff, "repo=acme/Platform/rules&number=9&path=rules.go")
	assert.Contains(t, file.Diff, "+package rules")
	assert.NotContains(t, file.Diff, "README.md", "only the file asked")

	commit := diff(s.handlePRCommitDiff, "repo=acme/Platform/rules&sha="+head)
	assert.Contains(t, commit.Diff, "+package rules")
	assert.Contains(t, commit.Diff, "+tight rules")
}
