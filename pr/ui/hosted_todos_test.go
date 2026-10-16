// Tests hosted todos: an organization's repositories are its todo workspaces, a todo is
// private to its creator unless made public, and only its creator or an admin changes it.
package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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
