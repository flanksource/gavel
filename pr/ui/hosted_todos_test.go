// Tests hosted todos: an organization's repositories are its todo workspaces, a todo is
// private to its creator unless made public, and only its creator or an admin changes it.
package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// hostedTodosServer serves org's hosted dashboard, acting for the user and role named in
// the X-Test-User and X-Test-Role headers, as the Clerk middleware would, and returns its URL.
func hostedTodosServer(t *testing.T, db *gorm.DB, org uuid.UUID, dataDir string) string {
	t.Helper()
	srv := &Server{}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.SetHostedWorkspaces(HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}})
	return serveActingAs(t, srv, org)
}

// serveActingAs serves srv acting for the user and role in the X-Test-User and X-Test-Role
// headers, and returns its URL.
func serveActingAs(t *testing.T, srv *Server, org uuid.UUID) string {
	t.Helper()
	handler := srv.Handler()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := uuid.MustParse(r.Header.Get("X-Test-User"))
		scope := tenant.Scope{OrgID: org, UserID: &user, Role: r.Header.Get("X-Test-Role")}
		handler.ServeHTTP(w, r.WithContext(tenant.WithScope(r.Context(), scope)))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

// pushedRepo records a pushed repository of org with one commit and returns its ID.
func pushedRepo(t *testing.T, db *gorm.DB, org uuid.UUID, dataDir, name string) uuid.UUID {
	t.Helper()
	var repo struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source) VALUES (?, ?, 'push') RETURNING id`, org, name).Scan(&repo).Error)
	work := t.TempDir()
	run := func(dir string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
	}
	run(work, "init", "-q", "-b", "main")
	require.NoError(t, os.WriteFile(filepath.Join(work, "README.md"), []byte(name), 0o644))
	run(work, "add", ".")
	run(work, "commit", "-q", "-m", "init")
	run(work, "clone", "-q", "--bare", work, filepath.Join(dataDir, "repos", org.String(), repo.ID.String()+".git"))
	return repo.ID
}

func hostedCall(t *testing.T, base, method, path string, user uuid.UUID, role string, body any) (int, map[string]any) {
	t.Helper()
	var payload bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&payload).Encode(body))
	}
	req, err := http.NewRequest(method, base+path, &payload)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", user.String())
	req.Header.Set("X-Test-Role", role)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var decoded any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&decoded))
	if object, ok := decoded.(map[string]any); ok {
		return resp.StatusCode, object
	}
	return resp.StatusCode, map[string]any{"items": decoded}
}

func listedTitles(response map[string]any) []string {
	var titles []string
	items, _ := response["items"].([]any)
	for _, item := range items {
		titles = append(titles, item.(map[string]any)["title"].(string))
	}
	return titles
}

func TestHostedTodosLiveInTheOrgsRepositories(t *testing.T) {
	db, _, org := openOrgDB(t)
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	base := hostedTodosServer(t, db, org, dataDir)
	workspace := "repo:" + repo.String()

	status, projects := hostedCall(t, base, http.MethodGet, "/api/projects", creator, "org:member", nil)
	require.Equal(t, http.StatusOK, status, projects)
	listed := projects["items"].([]any)
	require.Len(t, listed, 1)
	assert.Equal(t, "app", listed[0].(map[string]any)["name"])
	assert.Equal(t, workspace, listed[0].(map[string]any)["dir"])

	status, created := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Private plan"})
	require.Equal(t, http.StatusCreated, status, created)
	assert.Equal(t, "private", created["visibility"])
	status, shared := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Shared plan", "visibility": "public"})
	require.Equal(t, http.StatusCreated, status, shared)

	_, list := hostedCall(t, base, http.MethodGet, "/api/todos?dir="+workspace, creator, "org:member", nil)
	assert.ElementsMatch(t, []string{"Private plan", "Shared plan"}, listedTitles(list))
	_, list = hostedCall(t, base, http.MethodGet, "/api/todos?dir="+workspace, member, "org:member", nil)
	assert.Equal(t, []string{"Shared plan"}, listedTitles(list))

	status, refused := hostedCall(t, base, http.MethodPatch, "/api/todos/item", member, "org:member", map[string]any{"dir": workspace, "ref": shared["id"], "title": "Taken over"})
	assert.Equal(t, http.StatusForbidden, status)
	assert.Contains(t, refused["error"], "only its creator or an org admin can change it")

	status, refused = hostedCall(t, base, http.MethodGet, "/api/todos?dir=/etc", creator, "org:member", nil)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, `todo workspace "/etc" is not one of this organization's repositories: name it repo:<id>`, refused["error"])

	status, refused = hostedCall(t, base, http.MethodGet, "/api/todos?dir=repo:"+uuid.NewString(), creator, "org:member", nil)
	assert.Equal(t, http.StatusNotFound, status)
	assert.Contains(t, refused["error"], "is not one of this organization's repositories")
}

func TestHostedWorkspacesRunInSandboxes(t *testing.T) {
	db, _, org := openOrgDB(t)
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")

	workspace, err := HostedWorkspaces{DB: db, Sources: &sandbox.Sources{DataDir: dataDir}}.Resolve(t.Context(), org, "repo:"+repo.String())

	require.NoError(t, err)
	assert.True(t, workspace.Options.SandboxedRuns, "a hosted workspace's runs execute in sandboxes")
}

func TestHostedTodoBatchListsTheOrgsWorkspaces(t *testing.T) {
	db, _, org := openOrgDB(t)
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	workspace := "repo:" + pushedRepo(t, db, org, dataDir, "app").String()
	base := hostedTodosServer(t, db, org, dataDir)
	hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Private plan"})
	hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": workspace, "title": "Shared plan", "visibility": "public"})

	status, batch := hostedCall(t, base, http.MethodPost, "/api/todos/batch", member, "org:member", map[string]any{"dirs": []string{workspace, "repo:" + uuid.NewString()}})
	require.Equal(t, http.StatusOK, status, batch)
	results := batch["results"].([]any)
	require.Len(t, results, 2)
	listed := results[0].(map[string]any)
	assert.Equal(t, workspace, listed["dir"])
	assert.Equal(t, []string{"Shared plan"}, listedTitles(listed), "the member sees only the public todo")
	assert.Contains(t, results[1].(map[string]any)["error"].(map[string]any)["message"], "is not one of this organization's repositories")

	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/batch", member, "org:member", map[string]any{"dirs": []string{"/etc"}})
	require.Equal(t, http.StatusOK, status, refused)
	assert.Contains(t, refused["results"].([]any)[0].(map[string]any)["error"].(map[string]any)["message"], `todo workspace "/etc" is not one of this organization's repositories`)
}

func TestAHostedDashboardWithoutWorkspacesReadsNoSessions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	srv := &Server{}
	srv.SetHostedMode(true)
	base := serveActingAs(t, srv, uuid.New())
	user, session := uuid.New(), uuid.NewString()

	for _, call := range []struct {
		method, path string
		body         any
	}{
		{http.MethodGet, "/api/todos/session/approvals?sessionId=" + session, nil},
		{http.MethodPost, "/api/todos/session/approve", map[string]any{"approvalId": uuid.NewString(), "action": "approve"}},
		{http.MethodGet, "/api/todos/session/stats?sessionId=" + session, nil},
		{http.MethodGet, "/api/captain/sessions/" + session, nil},
		{http.MethodGet, "/api/todos/item?ref=ab12cd", nil},
	} {
		status, refused := hostedCall(t, base, call.method, call.path, user, "org:member", call.body)
		assert.Equal(t, http.StatusNotFound, status, "%s %s: %v", call.method, call.path, refused)
		assert.Equal(t, "this hosted dashboard has no todo workspaces", refused["error"], "%s %s", call.method, call.path)
	}
}

func TestHostedTodoFormCreatesTodosWithTheirVisibility(t *testing.T) {
	db, _, org := openOrgDB(t)
	creator := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, creator).Error)
	dataDir := t.TempDir()
	workspace := "repo:" + pushedRepo(t, db, org, dataDir, "app").String()
	base := hostedTodosServer(t, db, org, dataDir)

	status, created := hostedCall(t, base, http.MethodPost, "/api/todos/new", creator, "org:member", map[string]any{"dir": workspace, "title": "From the form", "visibility": "public"})
	require.Equal(t, http.StatusCreated, status, created)
	assert.Equal(t, "public", created["todo"].(map[string]any)["visibility"])

	var form bytes.Buffer
	writer := multipart.NewWriter(&form)
	for name, value := range map[string]string{"dir": workspace, "title": "With a screenshot", "visibility": "public"} {
		require.NoError(t, writer.WriteField(name, value))
	}
	require.NoError(t, writer.Close())
	req, err := http.NewRequest(http.MethodPost, base+"/api/todos/new", &form)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Test-User", creator.String())
	req.Header.Set("X-Test-Role", "org:member")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var multipartCreated map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&multipartCreated))
	require.Equal(t, http.StatusCreated, resp.StatusCode, multipartCreated)
	assert.Equal(t, "public", multipartCreated["todo"].(map[string]any)["visibility"], "a multipart form keeps its visibility")
}

func TestHostedTodosMoveAndTakeFixturesForTheirOwner(t *testing.T) {
	db, _, org := openOrgDB(t)
	creator, member := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": creator, "grace": member} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	dataDir := t.TempDir()
	app := "repo:" + pushedRepo(t, db, org, dataDir, "app").String()
	api := "repo:" + pushedRepo(t, db, org, dataDir, "api").String()
	base := hostedTodosServer(t, db, org, dataDir)
	_, shared := hostedCall(t, base, http.MethodPost, "/api/todos", creator, "org:member", map[string]any{"dir": app, "title": "Shared", "visibility": "public"})
	ref := shared["id"].(string)

	fixture := map[string]any{"dir": app, "ref": ref, "fixture": "```bash\nmake test\n```"}
	status, refused := hostedCall(t, base, http.MethodPost, "/api/todos/verification/fixture", member, "org:member", fixture)
	assert.Equal(t, http.StatusForbidden, status, "another member can't change its fixture: %v", refused)
	status, saved := hostedCall(t, base, http.MethodPost, "/api/todos/verification/fixture", creator, "org:member", fixture)
	require.Equal(t, http.StatusOK, status, saved)

	move := map[string]any{"ref": ref, "fromDir": app, "toDir": api}
	status, refused = hostedCall(t, base, http.MethodPost, "/api/todos/transfer", member, "org:member", move)
	assert.Equal(t, http.StatusForbidden, status, "another member can't move it: %v", refused)
	status, moved := hostedCall(t, base, http.MethodPost, "/api/todos/transfer", creator, "org:member", move)
	require.Equal(t, http.StatusOK, status, moved)
	_, list := hostedCall(t, base, http.MethodGet, "/api/todos?dir="+api, member, "org:member", nil)
	assert.Equal(t, []string{"Shared"}, listedTitles(list), "it moved, still public")
}

func TestHostedRunCommitsAreReadFromTheOrgsRepository(t *testing.T) {
	db, _, org := openOrgDB(t)
	member := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_grace')`, member).Error)
	dataDir := t.TempDir()
	repo := pushedRepo(t, db, org, dataDir, "app")
	workspace := "repo:" + repo.String()
	base := hostedTodosServer(t, db, org, dataDir)

	// A run's commits reach the repository on a gavel/ branch, as a sandbox job uploads them.
	bare := filepath.Join(dataDir, "repos", org.String(), repo.String()+".git")
	work := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = work
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	git("clone", "-q", bare, ".")
	require.NoError(t, os.WriteFile(filepath.Join(work, "fix.go"), []byte("package app\n"), 0o644))
	git("add", ".")
	git("commit", "-q", "-m", "fix the build")
	sha := git("rev-parse", "HEAD")
	git("push", "-q", "origin", "HEAD:refs/heads/gavel/abcd1234-ef567890")

	status, files := hostedCall(t, base, http.MethodGet, "/api/todos/commits/files?dir="+workspace+"&hash="+sha, member, "org:member", nil)
	require.Equal(t, http.StatusOK, status, files)
	assert.Contains(t, fmt.Sprint(files["files"]), "fix.go")

	status, diff := hostedCall(t, base, http.MethodGet, "/api/todos/commits/diff?dir="+workspace+"&hash="+sha, member, "org:member", nil)
	require.Equal(t, http.StatusOK, status, diff)
	assert.Contains(t, diff["diff"], "+package app")

	status, refused := hostedCall(t, base, http.MethodGet, "/api/todos/commits/diff?dir=/etc&hash="+sha, member, "org:member", nil)
	assert.Equal(t, http.StatusBadRequest, status, refused)
}
