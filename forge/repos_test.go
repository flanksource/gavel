// Tests syncing an organization's repositories from its git connections: each connection's
// namespace is listed into repos, followed across renames, and pruned of what it no longer lists.
package forge_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// listing is a fake provider API answering one path with a JSON body the test can change.
type listing struct {
	*httptest.Server
	mu   sync.Mutex
	body string
}

func fakeListing(t *testing.T, path string) *listing {
	t.Helper()
	l := &listing{body: "[]"}
	l.Server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		l.mu.Lock()
		defer l.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(l.body))
	}))
	t.Cleanup(l.Close)
	return l
}

func (l *listing) set(body string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.body = body
}

// hostedOrg is a hosted organization in db.
func hostedOrg(t *testing.T, db *gorm.DB) tenant.Scope {
	t.Helper()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	return tenant.Scope{OrgID: org, Role: "org:admin"}
}

func saveFor(t *testing.T, db *gorm.DB, scope tenant.Scope, connection models.Connection) uuid.UUID {
	t.Helper()
	store, closeStore, err := connections.OpenStore(context.Background(), connections.Options{DB: db, Scope: scope})
	require.NoError(t, err)
	defer func() { require.NoError(t, closeStore()) }()
	saved, err := store.SaveConnection(context.Background(), &connection)
	require.NoError(t, err)
	return saved.ID
}

// heldRepo is one of an organization's repositories as the sync records it.
type heldRepo struct {
	Name, Source, ProviderRepoID, RemoteURL string
	ConnectionID                            *uuid.UUID
	InstallationID                          *int64
}

func held(t *testing.T, db *gorm.DB, org uuid.UUID) []heldRepo {
	t.Helper()
	var rows []heldRepo
	require.NoError(t, db.Raw(`SELECT name, source, coalesce(provider_repo_id, '') AS provider_repo_id, coalesce(remote_url, '') AS remote_url,
		connection_id, installation_id FROM repos WHERE org_id = ? ORDER BY name`, org).Scan(&rows).Error)
	return rows
}

func TestAConnectionBringsTheRepositoriesOfItsNamespace(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	gitlab := fakeListing(t, "/api/v4/groups/acme/projects")
	gitlab.set(`[{"id":31,"path":"api","path_with_namespace":"acme/api","http_url_to_repo":"https://gitlab.example.com/acme/api.git","default_branch":"main","namespace":{"full_path":"acme"}},
		{"id":32,"path":"web","path_with_namespace":"acme/web","http_url_to_repo":"https://gitlab.example.com/acme/web.git","default_branch":"dev","namespace":{"full_path":"acme"}},
		{"id":33,"path":"old","path_with_namespace":"acme/old","http_url_to_repo":"https://gitlab.example.com/acme/old.git","namespace":{"full_path":"acme"}}]`)
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: gitlab.URL, Password: "glpat",
		InsecureTLS: true, Properties: map[string]string{"group": "acme"}})

	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	assert.Equal(t, []heldRepo{
		{Name: "acme/api", Source: "gitlab", ProviderRepoID: "31", RemoteURL: "https://gitlab.example.com/acme/api.git", ConnectionID: &conn},
		{Name: "acme/old", Source: "gitlab", ProviderRepoID: "33", RemoteURL: "https://gitlab.example.com/acme/old.git", ConnectionID: &conn},
		{Name: "acme/web", Source: "gitlab", ProviderRepoID: "32", RemoteURL: "https://gitlab.example.com/acme/web.git", ConnectionID: &conn},
	}, held(t, db, scope.OrgID))
	var branches []string
	require.NoError(t, db.Raw(`SELECT default_branch FROM repos WHERE org_id = ? ORDER BY name`, scope.OrgID).Scan(&branches).Error)
	assert.Equal(t, []string{"main", "main", "dev"}, branches, "a project without a default branch gets main")

	// web is used by a sandbox run, so it stays when the connection stops listing it.
	var web struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`SELECT id FROM repos WHERE org_id = ? AND name = 'acme/web'`, scope.OrgID).Scan(&web).Error)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend) VALUES (?, ?, ?, 'test', 'succeeded', 'docker')`,
		uuid.New(), scope.OrgID, web.ID).Error)
	gitlab.set(`[{"id":31,"path":"api-v2","path_with_namespace":"acme/api-v2","http_url_to_repo":"https://gitlab.example.com/acme/api-v2.git","default_branch":"main","namespace":{"full_path":"acme"}}]`)

	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	assert.Equal(t, []heldRepo{
		{Name: "acme/api-v2", Source: "gitlab", ProviderRepoID: "31", RemoteURL: "https://gitlab.example.com/acme/api-v2.git", ConnectionID: &conn},
		{Name: "acme/web", Source: "gitlab", ProviderRepoID: "32", RemoteURL: "https://gitlab.example.com/acme/web.git"},
	}, held(t, db, scope.OrgID), "renamed in place, unreferenced removed, referenced kept without its connection")
}

func TestAGitHubConnectionSharesTheRepositoryItsInstallationGrants(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (42, ?, 'acme', 'Organization')`, scope.OrgID).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/api', 'github', 101, 42)`, scope.OrgID).Error)
	github := fakeListing(t, "/orgs/acme/repos")
	github.set(`[{"id":101,"name":"api","full_name":"acme/api","clone_url":"https://github.com/acme/api.git","default_branch":"main"},
		{"id":102,"name":"web","full_name":"acme/web","clone_url":"https://github.com/acme/web.git","default_branch":"main"}]`)
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-github", Type: models.ConnectionTypeGithub, URL: github.URL, Password: "ghp",
		InsecureTLS: true, Properties: map[string]string{"organization": "acme"}})

	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	installation := int64(42)
	assert.Equal(t, []heldRepo{
		{Name: "acme/api", Source: "github", ProviderRepoID: "101", RemoteURL: "https://github.com/acme/api.git", ConnectionID: &conn, InstallationID: &installation},
		{Name: "acme/web", Source: "github", ProviderRepoID: "102", RemoteURL: "https://github.com/acme/web.git", ConnectionID: &conn},
	}, held(t, db, scope.OrgID), "one row per GitHub repository, keeping the installation that also grants it")
	var ids []int64
	require.NoError(t, db.Raw(`SELECT github_repo_id FROM repos WHERE org_id = ? ORDER BY github_repo_id`, scope.OrgID).Scan(&ids).Error)
	assert.Equal(t, []int64{101, 102}, ids)

	github.set(`[]`)
	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	assert.Equal(t, []heldRepo{{Name: "acme/api", Source: "github", ProviderRepoID: "101", RemoteURL: "https://github.com/acme/api.git", InstallationID: &installation}},
		held(t, db, scope.OrgID), "the installation still grants api; nothing else brings web")
}

func TestAConnectionThatFailsDoesntStopTheOthers(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	gitlab := fakeListing(t, "/api/v4/groups/acme/projects")
	gitlab.set(`[{"id":31,"path":"api","path_with_namespace":"acme/api","http_url_to_repo":"https://gitlab.example.com/acme/api.git","namespace":{"full_path":"acme"}}]`)
	saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: gitlab.URL, Password: "glpat",
		InsecureTLS: true, Properties: map[string]string{"group": "acme"}})
	missing := fakeListing(t, "/nothing")
	saveFor(t, db, scope, models.Connection{Name: "gone", Type: models.ConnectionTypeGitlab, URL: missing.URL, Password: "glpat",
		InsecureTLS: true, Properties: map[string]string{"group": "gone"}})

	err := forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{})
	require.ErrorContains(t, err, `list the repositories of connection "gone": `)
	var names []string
	for _, repo := range held(t, db, scope.OrgID) {
		names = append(names, repo.Name)
	}
	sort.Strings(names)
	assert.Equal(t, []string{"acme/api"}, names, "the working connection still synced")
}

func TestAnOrgWithoutGitConnectionsSyncsWithoutAKey(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	t.Setenv("GAVEL_SECRET_KEY", "")
	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	assert.Empty(t, held(t, db, scope.OrgID))
}

func TestDeletingAConnectionRemovesTheRepositoriesOnlyItBrings(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	gitlab := fakeListing(t, "/api/v4/groups/acme/projects")
	gitlab.set(`[{"id":31,"path":"api","path_with_namespace":"acme/api","http_url_to_repo":"https://gitlab.example.com/acme/api.git","namespace":{"full_path":"acme"}},
		{"id":32,"path":"web","path_with_namespace":"acme/web","http_url_to_repo":"https://gitlab.example.com/acme/web.git","namespace":{"full_path":"acme"}}]`)
	saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: gitlab.URL, Password: "glpat",
		InsecureTLS: true, Properties: map[string]string{"group": "acme"}})
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (42, ?, 'acme', 'Organization')`, scope.OrgID).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'acme/app', 'github', 101, 42)`, scope.OrgID).Error)
	github := fakeListing(t, "/orgs/acme/repos")
	github.set(`[{"id":101,"name":"app","full_name":"acme/app","clone_url":"https://github.com/acme/app.git","default_branch":"main"}]`)
	saveFor(t, db, scope, models.Connection{Name: "acme-github", Type: models.ConnectionTypeGithub, URL: github.URL, Password: "ghp",
		InsecureTLS: true, Properties: map[string]string{"organization": "acme"}})
	require.NoError(t, forge.SyncRepos(context.Background(), db, scope, git.ClientOptions{}))
	var web struct{ ID uuid.UUID }
	require.NoError(t, db.Raw(`SELECT id FROM repos WHERE org_id = ? AND name = 'acme/web'`, scope.OrgID).Scan(&web).Error)
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend) VALUES (?, ?, ?, 'test', 'succeeded', 'docker')`,
		uuid.New(), scope.OrgID, web.ID).Error)

	require.NoError(t, forge.DeleteConnection(context.Background(), db, scope, "acme-gitlab"))
	require.NoError(t, forge.DeleteConnection(context.Background(), db, scope, "acme-github"))
	installation := int64(42)
	assert.Equal(t, []heldRepo{
		{Name: "acme/app", Source: "github", ProviderRepoID: "101", RemoteURL: "https://github.com/acme/app.git", InstallationID: &installation},
		{Name: "acme/web", Source: "gitlab", ProviderRepoID: "32", RemoteURL: "https://gitlab.example.com/acme/web.git"},
	}, held(t, db, scope.OrgID), "api is gone; web is kept for its run and app for its installation, both without the connection")
	var left int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM connections WHERE org_id = ?`, scope.OrgID).Scan(&left).Error)
	assert.Zero(t, left)

	require.ErrorContains(t, forge.DeleteConnection(context.Background(), db, scope, "acme-gitlab"), `connection "acme-gitlab"`)
}
