// Tests linking GitHub App installations to organizations and syncing their repositories,
// against a migrated database and a fake GitHub API.
package installations_test

import (
	"context"
	"sort"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/githubtest"
	"github.com/flanksource/gavel/github/installations"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fixture struct {
	db         *gorm.DB
	store      installations.Store
	github     *githubtest.Server
	orgA, orgB uuid.UUID
	user       uuid.UUID
}

var acmeAccount = github.InstallationAccount{Login: "acme", Type: "Organization"}

func newFixture(t *testing.T) fixture {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_installations"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	db := opened.Gorm()

	f := fixture{db: db, github: githubtest.New(t), orgA: uuid.New(), orgB: uuid.New(), user: uuid.New()}
	for _, org := range []uuid.UUID{f.orgA, f.orgB} {
		require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Org')`, org, org.String()).Error)
	}
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', 'user_ada')`, f.user).Error)
	f.store = installations.Store{DB: db, App: f.github.App(t)}
	return f
}

func acme(id int64, name string) github.InstallationRepo {
	return github.InstallationRepo{ID: id, FullName: "acme/" + name, DefaultBranch: "main"}
}

func (f fixture) repoRows(t *testing.T, org uuid.UUID) []string {
	t.Helper()
	var rows []struct {
		Name           string
		InstallationID *int64
	}
	require.NoError(t, f.db.Raw(`SELECT name, installation_id FROM repos WHERE org_id = ? AND source = 'github' ORDER BY name`, org).Scan(&rows).Error)
	var out []string
	for _, row := range rows {
		linked := "unlinked"
		if row.InstallationID != nil {
			linked = "linked"
		}
		out = append(out, row.Name+" "+linked)
	}
	return out
}

func TestLinkInstallation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	installation := github.Installation{ID: 42, Account: github.InstallationAccount{Login: "acme", Type: "Organization"}}

	require.NoError(t, f.store.Link(ctx, f.orgA, f.user, installation))
	require.NoError(t, f.store.Link(ctx, f.orgA, f.user, installation), "linking again to the same org is a no-op")
	err := f.store.Link(ctx, f.orgB, f.user, installation)
	require.ErrorIs(t, err, installations.ErrLinkedToAnotherOrg)
	assert.Contains(t, err.Error(), "acme")

	ids, err := f.store.Installations(ctx, f.orgA)
	require.NoError(t, err)
	assert.Equal(t, []int64{42}, ids)
	ids, err = f.store.Installations(ctx, f.orgB)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestSyncRepos(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	require.NoError(t, f.store.Link(ctx, f.orgA, f.user, github.Installation{ID: 42, Account: github.InstallationAccount{Login: "acme", Type: "Organization"}}))

	f.github.Install(42, acmeAccount, acme(1, "api"), acme(2, "web"), acme(3, "old"))
	require.NoError(t, f.store.SyncRepos(ctx, f.orgA, 42))
	assert.Equal(t, []string{"acme/api linked", "acme/old linked", "acme/web linked"}, f.repoRows(t, f.orgA))
	names, err := f.store.Repos(ctx, f.orgA)
	require.NoError(t, err)
	assert.Equal(t, []string{"acme/api", "acme/old", "acme/web"}, names)

	// web is used by a sandbox run, so it stays when the installation drops it.
	var web struct{ ID uuid.UUID }
	require.NoError(t, f.db.Raw(`SELECT id FROM repos WHERE name = 'acme/web'`).Scan(&web).Error)
	require.NoError(t, f.db.Exec(`INSERT INTO sandbox_runs (id, org_id, repo_id, kind, status, backend) VALUES (?, ?, ?, 'test', 'succeeded', 'docker')`,
		uuid.New(), f.orgA, web.ID).Error)

	f.github.Install(42, acmeAccount, acme(1, "api-v2"), acme(4, "new"))
	require.NoError(t, f.store.SyncRepos(ctx, f.orgA, 42))
	assert.Equal(t, []string{"acme/api-v2 linked", "acme/new linked", "acme/web unlinked"}, f.repoRows(t, f.orgA),
		"renamed in place, added, unreferenced removed, referenced kept without its installation")
	names, err = f.store.Repos(ctx, f.orgA)
	require.NoError(t, err)
	sort.Strings(names)
	assert.Equal(t, []string{"acme/api-v2", "acme/new"}, names, "only repos an installation still grants")

	f.github.Install(42, acmeAccount)
	require.NoError(t, f.store.SyncRepos(ctx, f.orgA, 42))
	assert.Equal(t, []string{"acme/web unlinked"}, f.repoRows(t, f.orgA), "an installation that grants nothing keeps nothing")

	err = f.store.SyncRepos(ctx, f.orgB, 42)
	require.ErrorIs(t, err, installations.ErrNotLinked, "an org can't sync an installation it didn't link")
}

func TestTokenSource(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	_, err := f.store.TokenSource(f.orgA)()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connect GitHub", "an org without an installation is told to connect one")

	f.github.Install(42, acmeAccount)
	require.NoError(t, f.store.Link(ctx, f.orgA, f.user, github.Installation{ID: 42, Account: github.InstallationAccount{Login: "acme"}}))
	token, err := f.store.TokenSource(f.orgA)()
	require.NoError(t, err)
	assert.Equal(t, "ghs_42", token)
}

func TestReposByInstallation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	for id, login := range map[int64]string{42: "acme", 43: "globex"} {
		account := github.InstallationAccount{Login: login, Type: "Organization"}
		f.github.Install(id, account, github.InstallationRepo{ID: id * 10, FullName: login + "/api"}, github.InstallationRepo{ID: id*10 + 1, FullName: login + "/web"})
		require.NoError(t, f.store.Link(ctx, f.orgA, f.user, github.Installation{ID: id, Account: account}))
		require.NoError(t, f.store.SyncRepos(ctx, f.orgA, id))
	}

	groups, err := f.store.ReposByInstallation(ctx, f.orgA)
	require.NoError(t, err)
	assert.Equal(t, []installations.InstallationRepos{
		{Installation: 42, Repos: []string{"acme/api", "acme/web"}},
		{Installation: 43, Repos: []string{"globex/api", "globex/web"}},
	}, groups)

	token, err := f.store.InstallationTokenSource(43)()
	require.NoError(t, err)
	assert.Equal(t, "ghs_43", token)

	groups, err = f.store.ReposByInstallation(ctx, f.orgB)
	require.NoError(t, err)
	assert.Empty(t, groups)
}
