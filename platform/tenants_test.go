// Tests the platform dashboard's tenant list against a migrated database: every
// organization of the Clerk instance with its members, repositories, installations and runs.
package platform_test

import (
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const issuer = "https://clerk.example"

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_platform"}).DSN()
	t.Setenv(database.EnvDSN, dsn)
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	return opened.Gorm()
}

func addOrg(t *testing.T, db *gorm.DB, orgIssuer, clerkID, slug, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, slug, name) VALUES (?, ?, ?, ?, ?)`, id, orgIssuer, clerkID, slug, name).Error)
	return id
}

func addMember(t *testing.T, db *gorm.DB, org uuid.UUID) {
	t.Helper()
	user := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, ?, ?)`, user, issuer, "user_"+user.String()).Error)
	require.NoError(t, db.Exec(`INSERT INTO org_memberships (user_id, org_id, role) VALUES (?, ?, 'org:member')`, user, org).Error)
}

func addRun(t *testing.T, db *gorm.DB, org uuid.UUID, kind, status string, created time.Time, started, finished *time.Time) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, created_at, started_at, finished_at) VALUES (?, ?, ?, ?, 'substrate', ?, ?, ?)`,
		uuid.New(), org, kind, status, created, started, finished).Error)
}

func TestTenants(t *testing.T) {
	db := openDB(t)
	now := time.Now()
	alpha := addOrg(t, db, issuer, "org_alpha", "alpha", "Alpha")
	beta := addOrg(t, db, issuer, "org_beta", "beta", "Beta")
	addOrg(t, db, "https://other-clerk.example", "org_other", "other", "Another Clerk instance")

	addMember(t, db, alpha)
	addMember(t, db, alpha)
	addMember(t, db, beta)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (42, ?, 'alpha-gh', 'Organization')`, alpha).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, installation_id) VALUES (?, 'alpha-gh/api', 'github', 1, 42)`, alpha).Error)
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source) VALUES (?, 'scratch', 'push')`, alpha).Error)
	lastRun := now.Add(-time.Hour)
	addRun(t, db, alpha, "agent", "running", lastRun, nil, nil)
	addRun(t, db, alpha, "test", "succeeded", now.Add(-48*time.Hour), nil, nil)
	addRun(t, db, alpha, "lint", "failed", now.Add(-10*24*time.Hour), nil, nil)

	tenants, err := platform.Store{DB: db, Issuer: issuer}.Tenants(t.Context(), now.Add(-7*24*time.Hour))
	require.NoError(t, err)
	require.Len(t, tenants, 2, "only the Clerk instance's organizations, not the local one or another instance's: %+v", tenants)

	got := tenants[0]
	assert.Equal(t, alpha, got.ID)
	assert.Equal(t, "org_alpha", got.ClerkOrgID)
	assert.Equal(t, "alpha", got.Slug)
	assert.Equal(t, "Alpha", got.Name)
	assert.Equal(t, 2, got.Members)
	assert.Equal(t, 2, got.Repos)
	assert.Equal(t, 1, got.PushedRepos)
	assert.Equal(t, 1, got.GitHubInstallations)
	assert.Equal(t, 2, got.SandboxRuns, "runs since the window started")
	require.NotNil(t, got.LastRunAt)
	assert.WithinDuration(t, lastRun, *got.LastRunAt, time.Millisecond)

	assert.Equal(t, platform.Tenant{ID: beta, ClerkOrgID: "org_beta", Slug: "beta", Name: "Beta", CreatedAt: tenants[1].CreatedAt, Members: 1}, tenants[1],
		"an organization with nothing but a member")
}
