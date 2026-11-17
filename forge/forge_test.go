// Tests finding the git connection that holds a repository's credentials: the most specific
// connection owns it, file layers come before the local catalog, and the catalog needs its key.
package forge_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	dbcontext "github.com/flanksource/commons-db/context"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const header = "apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"

// writeConnection puts a connection file named name into dir.
func writeConnection(t *testing.T, dir, name, spec string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(header+"metadata: {name: "+name+"}\nspec: "+spec+"\n"), 0o600))
}

// workspace is a directory whose .connections layer the test writes, with empty global and app layers.
func workspace(t *testing.T) string {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	return t.TempDir()
}

func repo(t *testing.T, raw string) git.RepoURL {
	t.Helper()
	parsed, err := git.ParseRepoURL(raw, "")
	require.NoError(t, err)
	return parsed
}

// password is the password of the connection that owns target, or "" for none.
func password(t *testing.T, target git.RepoURL, sources ...forge.Source) string {
	t.Helper()
	found, err := forge.Find(dbcontext.NewContext(context.Background()), target, sources...)
	require.NoError(t, err)
	if found == nil {
		return ""
	}
	return found.Password
}

func localSources(t *testing.T, dir string, db *gorm.DB) []forge.Source {
	t.Helper()
	sources, closeSources, err := forge.Local(context.Background(), dir, db)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, closeSources()) })
	return sources
}

func TestTheMostSpecificConnectionOwnsARepository(t *testing.T) {
	dir := workspace(t)
	cwd := filepath.Join(dir, ".connections")
	writeConnection(t, cwd, "everything", "{github: {personalAccessToken: {value: gh_all}}}")
	writeConnection(t, cwd, "acme", "{properties: {organization: acme}, github: {personalAccessToken: {value: gh_acme}}}")
	writeConnection(t, cwd, "tool", "{git: {url: 'https://github.com/acme/tool.git', password: {value: gh_tool}}}")
	writeConnection(t, cwd, "lab", "{properties: {group: team}, gitlab: {url: 'https://gitlab.example.com', personalAccessToken: {value: gl_team}}}")
	writeConnection(t, cwd, "ado", "{properties: {project: web}, azureDevops: {url: 'https://dev.azure.com/contoso', organization: contoso, personalAccessToken: {value: ado_web}}}")
	sources := localSources(t, dir, nil)

	for _, tc := range []struct {
		name   string
		target git.RepoURL
		want   string
	}{
		{"a repository in a connection's namespace", repo(t, "https://github.com/acme/api"), "gh_acme"},
		{"the one repository a git connection names", repo(t, "git@github.com:acme/tool.git"), "gh_tool"},
		{"a repository only the host's connection owns", repo(t, "https://github.com/other/site"), "gh_all"},
		{"an owner's repositories", git.RepoURL{Scheme: "https", Host: "github.com", Owner: "acme"}, "gh_acme"},
		{"every repository on the host", git.RepoURL{Scheme: "https", Host: "github.com"}, "gh_all"},
		{"a self-hosted GitLab group's repository", repo(t, "https://gitlab.example.com/team/svc.git"), "gl_team"},
		{"an Azure DevOps project's repository", repo(t, "https://dev.azure.com/contoso/web/_git/site"), "ado_web"},
		{"an Azure DevOps project no connection names", repo(t, "https://dev.azure.com/contoso/ops/_git/site"), ""},
		{"a host no connection names", repo(t, "https://bitbucket.org/acme/api"), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, password(t, tc.target, sources...))
		})
	}
}

// catalog is a database holding the local organization's encrypted connections, keyed by
// GAVEL_SECRET_KEY.
func catalog(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_forge"}).DSN())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	return opened.Gorm()
}

func save(t *testing.T, db *gorm.DB, connection models.Connection) {
	t.Helper()
	store, closeStore, err := connections.OpenStore(context.Background(), connections.Options{DB: db, Scope: tenant.Local()})
	require.NoError(t, err)
	defer func() { require.NoError(t, closeStore()) }()
	_, err = store.SaveConnection(context.Background(), &connection)
	require.NoError(t, err)
}

func TestTheLocalCatalogsGitConnections(t *testing.T) {
	db := catalog(t)
	save(t, db, models.Connection{Name: "acme", Type: models.ConnectionTypeGithub, Password: "gh_catalog", Properties: map[string]string{"organization": "acme"}})
	save(t, db, models.Connection{Name: "orders", Type: models.ConnectionTypePostgres, URL: "postgres://db/orders"})
	dir := workspace(t)

	sources := localSources(t, dir, db)
	assert.Equal(t, "gh_catalog", password(t, repo(t, "https://github.com/acme/api"), sources...))

	writeConnection(t, filepath.Join(dir, ".connections"), "acme", "{properties: {organization: acme}, github: {personalAccessToken: {value: gh_file}}}")
	sources = localSources(t, dir, db)
	assert.Equal(t, "gh_file", password(t, repo(t, "https://github.com/acme/api"), sources...), "the file layers come before the catalog")
}

func TestACatalogWithGitConnectionsNeedsItsKey(t *testing.T) {
	db := catalog(t)
	save(t, db, models.Connection{Name: "acme", Type: models.ConnectionTypeGithub, Password: "gh_catalog"})
	t.Setenv("GAVEL_SECRET_KEY", "")

	_, _, err := forge.Local(context.Background(), workspace(t), db)
	require.EqualError(t, err, "the local connection catalog holds git connections, but its key is unavailable: encrypted connections require a keeper connection or GAVEL_SECRET_KEY (at least 32 base64-encoded bytes)")
}

func TestACatalogWithoutGitConnectionsNeedsNoKey(t *testing.T) {
	db := catalog(t)
	save(t, db, models.Connection{Name: "orders", Type: models.ConnectionTypePostgres, URL: "postgres://db/orders"})
	t.Setenv("GAVEL_SECRET_KEY", "")

	dir := workspace(t)
	writeConnection(t, filepath.Join(dir, ".connections"), "everything", "{github: {personalAccessToken: {value: gh_all}}}")
	assert.Equal(t, "gh_all", password(t, repo(t, "https://github.com/acme/api"), localSources(t, dir, db)...))
}
