// Tests serving every hosted organization from the gh CLI's login: the repositories the server
// names (or every one the login sees) are synced into each organization and searched with its token.
package sharedaccount_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/github/sharedaccount"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fixture struct {
	db         *gorm.DB
	github     *githubtest.Server
	orgA, orgB uuid.UUID
}

func cliToken() (string, error) { return "gho_cli", nil }

func newFixture(t *testing.T) fixture {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_sharedaccount"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	f := fixture{db: opened.Gorm(), github: githubtest.New(t), orgA: uuid.New(), orgB: uuid.New()}
	for _, org := range []uuid.UUID{f.orgA, f.orgB} {
		require.NoError(t, f.db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Org')`, org, org.String()).Error)
	}
	f.github.GrantUserRepos("gho_cli", acme(1, "api"), acme(2, "web"), acme(3, "tool"))
	return f
}

func acme(id int64, name string) github.InstallationRepo {
	return github.InstallationRepo{ID: id, FullName: "acme/" + name, DefaultBranch: "main"}
}

func (f fixture) repoNames(t *testing.T, org uuid.UUID) []string {
	t.Helper()
	var names []string
	require.NoError(t, f.db.Raw(`SELECT name FROM repos WHERE org_id = ? AND source = 'github' AND installation_id IS NULL ORDER BY name`, org).
		Scan(&names).Error)
	return names
}

func TestNamedRepositoriesReachEveryOrganization(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	account := sharedaccount.New(f.db, []string{"acme/api", "acme/web"}, cliToken)

	for _, org := range []uuid.UUID{f.orgA, f.orgB} {
		require.NoError(t, account.SyncAll(ctx, org))
		assert.Equal(t, []string{"acme/api", "acme/web"}, f.repoNames(t, org), "every organization holds the named repositories")
		groups, err := account.SearchGroups(ctx, org)
		require.NoError(t, err)
		require.Len(t, groups, 1)
		assert.Equal(t, []string{"acme/api", "acme/web"}, groups[0].Repos)
		token, err := groups[0].Token()
		require.NoError(t, err)
		assert.Equal(t, "gho_cli", token)
	}

	token, err := account.RepoToken(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, "gho_cli", token, "a repository without an installation is reached with the gh CLI's login")
	token, err = account.TokenSource(f.orgA)()
	require.NoError(t, err)
	assert.Equal(t, "gho_cli", token)
}

func TestEveryRepositoryTheLoginSees(t *testing.T) {
	f := newFixture(t)
	account := sharedaccount.New(f.db, []string{sharedaccount.AllRepos}, cliToken)
	require.NoError(t, account.SyncAll(context.Background(), f.orgA))
	assert.Equal(t, []string{"acme/api", "acme/tool", "acme/web"}, f.repoNames(t, f.orgA))
}

func TestARepositoryTheLoginCantSee(t *testing.T) {
	f := newFixture(t)
	account := sharedaccount.New(f.db, []string{"acme/api", "acme/secret"}, cliToken)
	err := account.SyncAll(context.Background(), f.orgA)
	require.EqualError(t, err, "read GitHub repository acme/secret with the gh CLI's login: HTTP 404: Not Found")
}

func TestRepositoriesTheServerNoLongerNamesAreRemoved(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	require.NoError(t, sharedaccount.New(f.db, []string{"acme/api", "acme/web", "acme/tool"}, cliToken).SyncAll(ctx, f.orgA))

	// web is used by a sandbox run, so it stays when the server stops naming it.
	var web struct{ ID uuid.UUID }
	require.NoError(t, f.db.Raw(`SELECT id FROM repos WHERE org_id = ? AND name = 'acme/web'`, f.orgA).Scan(&web).Error)
	require.NoError(t, f.db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend) VALUES (?, ?, ?, 'test', 'succeeded', 'docker')`,
		uuid.New(), f.orgA, web.ID).Error)

	account := sharedaccount.New(f.db, []string{"acme/api"}, cliToken)
	require.NoError(t, account.SyncAll(ctx, f.orgA))
	assert.Equal(t, []string{"acme/api", "acme/web"}, f.repoNames(t, f.orgA), "tool is removed; web is kept for its run")
	groups, err := account.SearchGroups(ctx, f.orgA)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, []string{"acme/api"}, groups[0].Repos, "only the repositories the server still names are searched")
}

func TestATokenFailureSaysWhereTheTokenComesFrom(t *testing.T) {
	f := newFixture(t)
	failing := func() (string, error) { return "", sharedaccount.ErrNoLogin }
	err := sharedaccount.New(f.db, []string{"acme/api"}, failing).SyncAll(context.Background(), f.orgA)
	require.ErrorIs(t, err, sharedaccount.ErrNoLogin)
	assert.EqualError(t, err, "read GitHub repository acme/api with the gh CLI's login: the gh CLI has no GitHub login; run gh auth login on the server")
}

func TestGHTokenReadsTheGHCLIsLogin(t *testing.T) {
	bin := t.TempDir()
	t.Setenv("PATH", bin)
	_, err := sharedaccount.GHToken()
	require.ErrorIs(t, err, sharedaccount.ErrNoLogin, "no gh on PATH")

	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\n[ \"$1 $2\" = 'auth token' ] && echo gho_cli\n"), 0o755))
	token, err := sharedaccount.GHToken()
	require.NoError(t, err)
	assert.Equal(t, "gho_cli", token)

	require.NoError(t, os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nexit 1\n"), 0o755))
	_, err = sharedaccount.GHToken()
	require.ErrorIs(t, err, sharedaccount.ErrNoLogin, "gh is installed but not logged in")
}
