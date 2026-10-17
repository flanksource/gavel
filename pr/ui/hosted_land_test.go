// Tests landing a hosted run: it opens a pull request from the run's branch on GitHub through the
// organization's GitHub App, titled from the todo, and lands no other way.
package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostedRunsLandAsPullRequests(t *testing.T) {
	db, _, org := openOrgDB(t)
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	// GitHub's repository, here a local one.
	githubDir := t.TempDir()
	origin := filepath.Join(githubDir, "acme", "app.git")
	work := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	git("init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte("app"), 0o644))
	git("add", ".")
	git("commit", "-q", "-m", "init")
	setup := git("rev-parse", "HEAD")
	git("clone", "-q", "--bare", work, origin)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (2, ?, 'acme', 'Organization')`, org).Error)
	var linked struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/app', 'github', 1, 2) RETURNING id`, org).Scan(&linked).Error)
	workspace := "repo:" + linked.ID.String()

	// GitHub's API, here a local one.
	var mu sync.Mutex
	var opened map[string]any
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/acme/app/pulls", r.URL.Path)
		require.Equal(t, "Bearer ghs_app", r.Header.Get("Authorization"), "the organization's App opens it")
		mu.Lock()
		defer mu.Unlock()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&opened))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number":7,"html_url":"https://github.com/acme/app/pull/7","title":"Fix the build","state":"open"}`))
	}))
	t.Cleanup(api.Close)
	t.Setenv("GITHUB_API_URL", api.URL)

	dataDir := t.TempDir()
	srv := &Server{ghOpts: github.Options{TokenSource: func() (string, error) { return "ghs_app", nil }}}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{
		DataDir: dataDir, GitHubURL: githubDir,
		Token: func(context.Context, int64) (string, error) { return "ghs_install", nil },
	}})
	base := serveActingAs(t, srv, org)
	_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Fix the build", "visibility": "public"})
	todo := uuid.MustParse(created["id"].(string))

	// The run's commits, on the branch serve put them on.
	git("checkout", "-q", "-b", "agent")
	require.NoError(t, os.WriteFile(filepath.Join(work, "fix.go"), []byte("package app\n"), 0o644))
	git("add", ".")
	git("commit", "-q", "-m", "fix the build")
	head := git("rev-parse", "HEAD")
	branch := sandbox.RunBranch(todo, uuid.New())
	git("push", "-q", origin, "HEAD:refs/heads/"+branch)
	session, promptRun := uuid.New(), uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO captain_sessions (id, source, provider, host_id) VALUES (?, 'gavel', 'todos', 'local')`, session).Error)
	workspaceRecord, err := json.Marshal(map[string]any{"cwd": "/workspace/src", "worktree": map[string]any{
		"repo": "/workspace/src", "branch": "agent", "setup": setup, "head": head,
	}})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, workspace) VALUES (?, ?, ?, 'gavel.todos', ?)`,
		promptRun, session, session, string(workspaceRecord)).Error)
	repository, err := native.NewRepository(db)
	require.NoError(t, err)
	issues := repository.ForOrg(org)
	issue, err := issues.GetIssue(context.Background(), todo)
	require.NoError(t, err)
	_, err = issues.LinkPromptRun(context.Background(), todo, promptRun, native.StepRun, 0, issue.Version, "test")
	require.NoError(t, err)

	land := map[string]any{"dir": workspace, "ref": todo.String(), "via": "pr", "base": "main"}
	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/land", member, "org:member", land)
	assert.Equal(t, http.StatusForbidden, status, "another member can't land it: %v", refused)

	status, refused = hostedCall(t, base, http.MethodPost, "/api/todos/land", creator, "org:member", map[string]any{"dir": workspace, "ref": todo.String(), "via": "merge"})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "a hosted run lands as a pull request; land it with via pr", refused["error"])

	status, landed := hostedCall(t, base, http.MethodPost, "/api/todos/land", creator, "org:member", land)
	require.Equal(t, http.StatusOK, status, landed)
	landing := landed["landing"].(map[string]any)
	assert.Equal(t, "pr", landing["via"])
	assert.Equal(t, float64(7), landing["prNumber"])
	assert.Equal(t, "https://github.com/acme/app/pull/7", landing["prUrl"])
	assert.Equal(t, head, landing["landedSha"])
	assert.Equal(t, "main", landing["targetBranch"])
	assert.Equal(t, float64(1), landing["commitCount"])
	mu.Lock()
	assert.Equal(t, branch, opened["head"], "from the run's branch")
	assert.Equal(t, "main", opened["base"])
	assert.Equal(t, "Fix the build", opened["title"], "titled from the todo")
	assert.Contains(t, opened["body"], "fix the build", "described from the run's commits")
	mu.Unlock()

	status, refused = hostedCall(t, base, http.MethodPost, "/api/todos/land", creator, "org:member", land)
	assert.Equal(t, http.StatusConflict, status, "it landed once: %v", refused)

	pushed := "repo:" + pushedRepo(t, db, org, dataDir, "tool").String()
	_, other := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": pushed, "title": "Tool fix"})
	status, refused = hostedCall(t, base, http.MethodPost, "/api/todos/land", creator, "org:member", map[string]any{"dir": pushed, "ref": other["id"], "via": "pr"})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "repository tool was pushed to gavel, not linked from GitHub, so its runs can't land as pull requests", refused["error"])
}
