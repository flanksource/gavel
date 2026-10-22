// Tests landing a hosted run: it opens a pull request from the run's branch on GitHub through the
// organization's GitHub App, titled as the user asks, leaving out the files they chose.
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
	"gorm.io/gorm"
)

// hostedLanding is a hosted server whose organization links acme/app from GitHub, a local
// repository here, with a GitHub API that records the pull requests it opens.
type hostedLanding struct {
	db                *gorm.DB
	org               uuid.UUID
	creator, member   uuid.UUID
	base, workspace   string
	origin, dataDir   string
	work              string
	setup             string
	git               func(args ...string) string
	mu                sync.Mutex
	opened            map[string]any
	refusePullRequest bool
}

func newHostedLanding(t *testing.T) *hostedLanding {
	db, _, org := openOrgDB(t)
	h := &hostedLanding{db: db, org: org, creator: uuid.New(), member: uuid.New()}
	for name, id := range map[string]uuid.UUID{"ada": h.creator, "grace": h.member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	// GitHub's repository, here a local one.
	githubDir := t.TempDir()
	h.origin = filepath.Join(githubDir, "acme", "app.git")
	h.work = t.TempDir()
	h.git = func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = h.work
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	h.git("init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(h.work, "README.md"), []byte("app"), 0o644))
	h.git("add", ".")
	h.git("commit", "-q", "-m", "init")
	h.setup = h.git("rev-parse", "HEAD")
	h.git("clone", "-q", "--bare", h.work, h.origin)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (2, ?, 'acme', 'Organization')`, org).Error)
	var linked struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/app', 'github', 1, 2) RETURNING id`, org).Scan(&linked).Error)
	h.workspace = "repo:" + linked.ID.String()

	// GitHub's API, here a local one.
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/acme/app/pulls", r.URL.Path)
		require.Equal(t, "Bearer ghs_app", r.Header.Get("Authorization"), "the organization's App opens it")
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.refusePullRequest {
			http.Error(w, `{"message":"Validation Failed"}`, http.StatusUnprocessableEntity)
			return
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&h.opened))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"number":7,"html_url":"https://github.com/acme/app/pull/7","title":"Fix the build","state":"open"}`))
	}))
	t.Cleanup(api.Close)
	t.Setenv("GITHUB_API_URL", api.URL)

	h.dataDir = t.TempDir()
	srv := &Server{ghOpts: github.Options{TokenSource: func() (string, error) { return "ghs_app", nil }}}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{
		DataDir: h.dataDir, GitHubURL: githubDir,
		Token: func(context.Context, int64) (string, error) { return "ghs_install", nil },
	}})
	h.base = serveActingAs(t, srv, org)
	return h
}

// run makes a public todo titled title whose run committed files on the branch serve put its
// commits on, and returns the todo, the branch and the run's head.
func (h *hostedLanding) run(t *testing.T, title string, files ...string) (uuid.UUID, string, string) {
	_, created := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.creator, "org:member", map[string]any{"dir": h.workspace, "title": title, "visibility": "public"})
	todo := uuid.MustParse(created["id"].(string))
	h.git("checkout", "-q", "-B", "agent", h.setup)
	for _, file := range files {
		require.NoError(t, os.WriteFile(filepath.Join(h.work, file), []byte("written by the agent\n"), 0o644))
		h.git("add", ".")
		h.git("commit", "-q", "-m", "add "+file)
	}
	head := h.git("rev-parse", "HEAD")
	branch := sandbox.RunBranch(todo, uuid.New())
	h.git("push", "-q", h.origin, "HEAD:refs/heads/"+branch)
	session, promptRun := uuid.New(), uuid.New()
	require.NoError(t, h.db.Exec(`INSERT INTO captain_sessions (id, source, provider, host_id) VALUES (?, 'gavel', 'todos', 'local')`, session).Error)
	workspaceRecord, err := json.Marshal(map[string]any{"cwd": "/workspace/src", "worktree": map[string]any{
		"repo": "/workspace/src", "branch": "agent", "setup": h.setup, "head": head,
	}})
	require.NoError(t, err)
	require.NoError(t, h.db.Exec(`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, workspace) VALUES (?, ?, ?, 'gavel.todos', ?)`,
		promptRun, session, session, string(workspaceRecord)).Error)
	repository, err := native.NewRepository(h.db)
	require.NoError(t, err)
	issues := repository.ForOrg(h.org)
	issue, err := issues.GetIssue(context.Background(), todo)
	require.NoError(t, err)
	_, err = issues.LinkPromptRun(context.Background(), todo, promptRun, native.StepRun, 0, issue.Version, "test")
	require.NoError(t, err)
	return todo, branch, head
}

func TestHostedRunsLandAsPullRequests(t *testing.T) {
	h := newHostedLanding(t)
	todo, branch, head := h.run(t, "Fix the build", "fix.go")

	land := map[string]any{"dir": h.workspace, "ref": todo.String(), "via": "pr", "base": "main"}
	status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.member, "org:member", land)
	assert.Equal(t, http.StatusForbidden, status, "another member can't land it: %v", refused)

	status, refused = hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.creator, "org:member", map[string]any{"dir": h.workspace, "ref": todo.String(), "via": "merge"})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "a hosted run lands as a pull request; land it with via pr", refused["error"])

	status, landed := hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.creator, "org:member", land)
	require.Equal(t, http.StatusOK, status, landed)
	landing := landed["landing"].(map[string]any)
	assert.Equal(t, "pr", landing["via"])
	assert.Equal(t, float64(7), landing["prNumber"])
	assert.Equal(t, "https://github.com/acme/app/pull/7", landing["prUrl"])
	assert.Equal(t, head, landing["landedSha"])
	assert.Equal(t, "main", landing["targetBranch"])
	assert.Equal(t, float64(1), landing["commitCount"])
	h.mu.Lock()
	assert.Equal(t, branch, h.opened["head"], "from the run's branch")
	assert.Equal(t, "main", h.opened["base"])
	assert.Equal(t, "Fix the build", h.opened["title"], "titled from the todo")
	assert.Contains(t, h.opened["body"], "add fix.go", "described from the run's commits")
	assert.NotContains(t, h.opened["body"], "Left out")
	h.mu.Unlock()

	status, refused = hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.creator, "org:member", land)
	assert.Equal(t, http.StatusConflict, status, "it landed once: %v", refused)

	pushed := "repo:" + pushedRepo(t, h.db, h.org, h.dataDir, "tool").String()
	_, other := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.creator, "org:member", map[string]any{"dir": pushed, "title": "Tool fix"})
	status, refused = hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.creator, "org:member", map[string]any{"dir": pushed, "ref": other["id"], "via": "pr"})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "repository tool was pushed to gavel, not linked from GitHub, so its runs can't land as pull requests", refused["error"])
}

func TestHostedLandingTakesItsTitleAndLeavesFilesOut(t *testing.T) {
	h := newHostedLanding(t)
	todo, branch, head := h.run(t, "Rate limit /login", "ratelimit.go", "notes.md")
	land := func(fields map[string]any) (int, map[string]any) {
		body := map[string]any{"dir": h.workspace, "ref": todo.String(), "via": "pr", "base": "main"}
		for key, value := range fields {
			body[key] = value
		}
		return hostedCall(t, h.base, http.MethodPost, "/api/todos/land", h.creator, "org:member", body)
	}

	status, refused := land(map[string]any{"leaveOut": []string{"ratelimit.go", "notes.md"}})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "can't leave those files out: that is every file the run changed, leaving nothing to land", refused["error"])
	status, refused = land(map[string]any{"leaveOut": []string{"main.go"}})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, "can't leave those files out: the run didn't change main.go", refused["error"])

	h.mu.Lock()
	h.refusePullRequest = true
	h.mu.Unlock()
	status, refused = land(map[string]any{"leaveOut": []string{"notes.md"}})
	assert.NotEqual(t, http.StatusOK, status, refused)
	assert.Equal(t, head, h.git("--git-dir", h.origin, "rev-parse", "refs/heads/"+branch), "a landing that failed takes its leave-out commit back off the branch")
	h.mu.Lock()
	h.refusePullRequest = false
	h.mu.Unlock()

	status, landed := land(map[string]any{"title": "  Limit login attempts  ", "leaveOut": []string{"notes.md"}})
	require.Equal(t, http.StatusOK, status, landed)
	landing := landed["landing"].(map[string]any)
	tip := h.git("--git-dir", h.origin, "rev-parse", "refs/heads/"+branch)
	assert.Equal(t, head, h.git("--git-dir", h.origin, "rev-parse", tip+"^"), "the leave-out commit is on top of the run's")
	assert.Equal(t, "ratelimit.go", h.git("--git-dir", h.origin, "diff", "--name-only", h.setup, tip), "the PR carries only the kept file")
	assert.Equal(t, tip, landing["landedSha"])
	assert.Equal(t, float64(3), landing["commitCount"], "the run's two commits and the leave-out one")
	h.mu.Lock()
	defer h.mu.Unlock()
	assert.Equal(t, branch, h.opened["head"], "from the run's own branch")
	assert.Equal(t, "Limit login attempts", h.opened["title"], "titled as the user asked")
	assert.Equal(t, "Lands todo "+todo.String()[:8]+": Rate limit /login\n\n## Commits\n\n- add ratelimit.go\n- add notes.md\n- Leave out notes.md\n\n## Left out\n\n- notes.md\n", h.opened["body"])
}

func TestHostedPRLinksReadThroughTheOrgsApp(t *testing.T) {
	db, store, org := openOrgDB(t)
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	workspace := "repo:" + pushedRepo(t, db, org, dataDir, "app").String()
	t.Setenv("GITHUB_TOKEN", "serve-own-token")
	var mu sync.Mutex
	var tokens []string
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		tokens = append(tokens, r.Header.Get("Authorization"))
		mu.Unlock()
		http.Error(w, `{"message":"unavailable"}`, http.StatusServiceUnavailable)
	}))
	t.Cleanup(api.Close)
	t.Setenv("GITHUB_API_URL", api.URL)
	srv := &Server{ghOpts: github.Options{TokenSource: func() (string, error) { return "ghs_app", nil }}, githubCache: store}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	base := serveActingAs(t, srv, org)
	const prURL = "https://github.com/acme/app/pull/7"
	for title, visibility := range map[string]string{"Private": "private", "Shared": "public"} {
		_, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": title, "visibility": visibility})
		status, linked := hostedCall(t, base, http.MethodPost, "/api/todos/pr-links", creator, "org:member", map[string]any{"dir": workspace, "id": created["id"], "url": prURL})
		require.Equal(t, http.StatusCreated, status, linked)
	}

	status, body := hostedRaw(t, base, http.MethodGet, "/api/prs/todos?dir="+workspace+"&url="+prURL, member)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, "Shared")
	assert.NotContains(t, body, "Private", "another member doesn't see a private todo linked to the PR")

	status, body = hostedRaw(t, base, http.MethodGet, "/api/prs/search?q=fix", member)
	require.Equal(t, http.StatusOK, status, body)

	status, body = hostedRaw(t, base, http.MethodGet, "/api/prs/linked?url="+prURL, member)
	assert.Equal(t, http.StatusBadGateway, status, body)
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, tokens, "the PR was read from GitHub")
	for _, token := range tokens {
		assert.Equal(t, "Bearer ghs_app", token, "through the organization's App, never serve's own token")
	}
}
