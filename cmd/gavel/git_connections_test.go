// Tests that a local gavel reads repository tokens from its git connections: a command's from the
// connection files alone, a local dashboard's from its catalog too, and a hosted server's never.
package main

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionFiles makes the working directory hold a github connection for acme's repositories,
// with no other GitHub credential on the host.
func connectionFiles(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".connections"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".connections", "acme.yaml"), []byte("apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
		"metadata: {name: acme}\nspec: {properties: {organization: acme}, github: {personalAccessToken: {value: gh_file}}}\n"), 0o600))
	previous := workingDir
	workingDir = dir
	t.Cleanup(func() { workingDir = previous })
}

func TestALocalRunReadsRepositoryTokensFromItsConnections(t *testing.T) {
	connectionFiles(t)
	client, err := github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.NoError(t, err)
	assert.Equal(t, "gh_file", client.Auth.Token)
}

func TestATokenLookupNeverOpensTheDatabase(t *testing.T) {
	connectionFiles(t)
	// A database with no connection catalog, as an older gavel's embedded one is.
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_no_catalog"}).DSN())
	client, err := github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.NoError(t, err)
	assert.Equal(t, "gh_file", client.Auth.Token, "a command reads the connection files alone")
}

func TestALocalDashboardsCatalogSuppliesRepositoryTokens(t *testing.T) {
	connectionFiles(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString([]byte(strings.Repeat("c", 32))))
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_local_catalog"}).DSN())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	store, closeStore, err := connections.OpenStore(context.Background(), connections.Options{DB: opened.Gorm(), Scope: tenant.Local()})
	require.NoError(t, err)
	_, err = store.SaveConnection(context.Background(), &models.Connection{Name: "web", Type: models.ConnectionTypeGithub, Password: "gh_catalog",
		Properties: map[string]string{"organization": "web-team"}})
	require.NoError(t, err)
	require.NoError(t, closeStore())

	useConnectionCatalog(opened.Gorm())
	t.Cleanup(func() { useConnectionCatalog(nil) })
	client, err := github.GitClient(context.Background(), github.Options{Repo: "web-team/site"})
	require.NoError(t, err)
	assert.Equal(t, "gh_catalog", client.Auth.Token)
	client, err = github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.NoError(t, err)
	assert.Equal(t, "gh_file", client.Auth.Token, "the connection files still come first")
}

func TestAHostedServerNeverReadsTheHostsConnections(t *testing.T) {
	connectionFiles(t)
	newHostedFixture(t)
	_, err := github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.ErrorIs(t, err, github.ErrNoToken)
}
