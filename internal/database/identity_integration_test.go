package database_test

import (
	"context"
	"errors"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/database/dbtest"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orgResolver struct {
	org api.Organization
	err error
}

func (r orgResolver) ResolveOrganization(context.Context, string) (api.Organization, error) {
	return r.org, r.err
}

func TestIdentitySchemaAndPersistence(t *testing.T) {
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_identity"}).DSN()
	// Exercise additive migration of an existing private cache without claiming it.
	legacy, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	require.NoError(t, legacy.AutoMigrate(&legacyHTTPCacheEntry{}))
	require.NoError(t, legacy.Create(&legacyHTTPCacheEntry{URL: "https://private.example", Method: "GET", StatusCode: 200}).Error)
	legacySQL, err := legacy.DB()
	require.NoError(t, err)
	require.NoError(t, legacySQL.Close())
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	unmigrated, err := database.Open(t.Context())
	require.NoError(t, err)
	ready, err := (api.IdentityStore{DB: unmigrated.Gorm()}).Ready(t.Context())
	require.NoError(t, err)
	require.False(t, ready, "identity tables are absent before migrations")
	require.NoError(t, unmigrated.Close())
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	ready, err = (api.IdentityStore{DB: db.Gorm()}).Ready(t.Context())
	require.NoError(t, err)
	require.True(t, ready, "identity tables exist after migrations")
	roots := []string{"todo_workspaces", "http_cache_entries", "workflow_run_caches", "job_log_caches", "workflow_def_caches", "seen_prs", "test_run_caches", "test_run_cursors", "file_scans", "linter_executions", "debounce_metadata", "task_run_history", "session_owners", "api_token_owners", "git_agent_task_owners", "runtime_profile_owners", "runtime_preset_owners"}
	for _, table := range roots {
		require.True(t, db.Gorm().Migrator().HasColumn(table, "user_id"), table)
		require.True(t, db.Gorm().Migrator().HasColumn(table, "org_id"), table)
		var count int64
		require.NoError(t, db.Gorm().Raw(`SELECT count(*) FROM pg_constraint WHERE conrelid = ('public.' || ?)::regclass AND contype = 'f' AND confrelid IN ('public.users'::regclass, 'public.orgs'::regclass)`, table).Scan(&count).Error)
		require.EqualValues(t, 2, count, table)
	}
	var unattributed int64
	require.NoError(t, db.Gorm().Raw(`SELECT count(*) FROM http_cache_entries WHERE user_id IS NULL AND org_id = ?`, tenant.LocalOrgID).Scan(&unattributed).Error)
	require.EqualValues(t, 1, unattributed, "existing rows belong to the local org with no user")
	store := api.IdentityStore{DB: db.Gorm()}
	identity := api.Identity{ID: "user_1", Issuer: "https://clerk.example", Name: "User", Email: "user@example.com", EmailVerified: true, OrganizationID: "org_1", OrganizationSlug: "team"}
	result, owner, err := store.Sync(t.Context(), identity)
	require.NoError(t, err)
	require.Equal(t, "team", result.OrganizationSlug)
	require.NotEqual(t, uuid.Nil, owner.UserID)
	require.NotEqual(t, uuid.Nil, owner.OrgID)
	require.NotEqual(t, tenant.LocalOrgID, owner.OrgID)
	identity.Name = "Updated User"
	identity.OrganizationSlug = ""
	result, repeatedOwner, err := store.Sync(t.Context(), identity)
	require.NoError(t, err)
	require.Equal(t, owner, repeatedOwner, "upserts retain internal UUIDs")
	require.Equal(t, "team", result.OrganizationSlug, "missing claims must not erase stored slug")
	var user api.User
	require.NoError(t, db.Gorm().First(&user, "id = ?", owner.UserID).Error)
	require.Equal(t, "Updated User", user.Name)
	store.OrganizationResolver = orgResolver{org: api.Organization{ID: "org_1", Slug: "renamed-team", Name: "Renamed Team"}}
	result, renamedOwner, err := store.Sync(t.Context(), identity)
	require.NoError(t, err)
	require.Equal(t, "renamed-team", result.OrganizationSlug)
	require.Equal(t, owner, renamedOwner)
	store.OrganizationResolver = orgResolver{org: api.Organization{ID: "other-org", Slug: "wrong"}}
	_, _, err = store.Sync(t.Context(), identity)
	require.Error(t, err)
	store.OrganizationResolver = orgResolver{err: errors.New("backend unavailable")}
	_, _, err = store.Sync(t.Context(), identity)
	require.Error(t, err)
	store.OrganizationResolver = nil
	other := identity
	other.Issuer = "https://other-clerk.example"
	_, otherOwner, err := store.Sync(t.Context(), other)
	require.NoError(t, err)
	require.NotEqual(t, owner.UserID, otherOwner.UserID)
	require.NotEqual(t, owner.OrgID, otherOwner.OrgID)
	personal := identity
	personal.OrganizationID = ""
	_, personalOwner, err := store.Sync(t.Context(), personal)
	require.NoError(t, err)
	require.Equal(t, tenant.LocalOrgID, personalOwner.OrgID, "an identity with no active org acts for the local org")
	require.Equal(t, owner.UserID, personalOwner.UserID)
	// Both direct ownership references reject nonexistent identities.
	dbtest.RequireConstraintViolation(t, dbtest.Quiet(db.Gorm()).Exec(`UPDATE http_cache_entries SET user_id = ?`, uuid.New()).Error, "http_cache_entries_user_id_fkey")
	dbtest.RequireConstraintViolation(t, dbtest.Quiet(db.Gorm()).Exec(`UPDATE http_cache_entries SET org_id = ?`, uuid.New()).Error, "http_cache_entries_org_id_fkey")
	require.NoError(t, db.Gorm().Exec(`UPDATE http_cache_entries SET user_id = ?, org_id = ?`, owner.UserID, owner.OrgID).Error)
	// An org with owned data cannot be deleted.
	dbtest.RequireConstraintViolation(t, dbtest.Quiet(db.Gorm()).Delete(&api.Org{}, "id = ?", owner.OrgID).Error, "http_cache_entries_org_id_fkey")
	// User deletion preserves historical resources while nulling attribution.
	require.NoError(t, db.Gorm().Delete(&api.User{}, "id = ?", owner.UserID).Error)
	var remaining int64
	require.NoError(t, db.Gorm().Raw(`SELECT count(*) FROM http_cache_entries WHERE user_id IS NULL AND org_id = ?`, owner.OrgID).Scan(&remaining).Error)
	require.EqualValues(t, 1, remaining)
	// Sidecars hold Captain IDs without a foreign key; the next migration sweeps
	// a row whose Captain resource is gone and keeps one whose resource exists.
	live, orphan := uuid.New(), uuid.New()
	require.NoError(t, db.Gorm().Exec(`INSERT INTO captain_sessions (id, source, provider, host_id) VALUES (?, 'test', 'test', 'local')`, live).Error)
	require.NoError(t, db.Gorm().Exec(`INSERT INTO session_owners (session_id, org_id) VALUES (?, ?), (?, ?)`, live, owner.OrgID, orphan, owner.OrgID).Error)
	second, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "schema apply is idempotent across Captain/Gavel ownership")
	require.NoError(t, second.Close())
	var owned []uuid.UUID
	require.NoError(t, db.Gorm().Raw(`SELECT session_id FROM session_owners WHERE session_id IN ?`, []uuid.UUID{live, orphan}).Scan(&owned).Error)
	require.Equal(t, []uuid.UUID{live}, owned)
	require.True(t, db.Gorm().Migrator().HasColumn("orgs", "slug"))
}

func TestIdentityStoreSyncsMemberships(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_memberships"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	store := api.IdentityStore{DB: db.Gorm()}
	_, owner, err := store.Sync(t.Context(), api.Identity{ID: "user_1", Issuer: "https://clerk.example"})
	require.NoError(t, err)

	type row struct{ ClerkOrgID, Slug, Name, Role string }
	memberships := func() []row {
		t.Helper()
		var rows []row
		require.NoError(t, db.Gorm().Raw(`
			SELECT org.clerk_org_id, COALESCE(org.slug, '') AS slug, org.name, membership.role
			FROM org_memberships AS membership JOIN orgs AS org ON org.id = membership.org_id
			WHERE membership.user_id = ? ORDER BY org.clerk_org_id`, owner.UserID).Scan(&rows).Error)
		return rows
	}

	require.NoError(t, store.SyncMemberships(t.Context(), owner.UserID, []api.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha", Role: "org:admin"},
		{OrgID: "org_b", OrgSlug: "beta", OrgName: "Beta", Role: "org:member"},
	}))
	assert.Equal(t, []row{{"org_a", "alpha", "Alpha", "org:admin"}, {"org_b", "beta", "Beta", "org:member"}}, memberships())

	// A resync replaces the set: dropped orgs lose the membership, roles and names update.
	require.NoError(t, store.SyncMemberships(t.Context(), owner.UserID, []api.Membership{
		{OrgID: "org_a", OrgSlug: "alpha", OrgName: "Alpha Renamed", Role: "org:member"},
	}))
	assert.Equal(t, []row{{"org_a", "alpha", "Alpha Renamed", "org:member"}}, memberships())
	var orgs int64
	require.NoError(t, db.Gorm().Raw(`SELECT count(*) FROM orgs WHERE issuer = 'https://clerk.example'`).Scan(&orgs).Error)
	assert.EqualValues(t, 2, orgs, "a dropped membership keeps its org")

	require.NoError(t, store.SyncMemberships(t.Context(), owner.UserID, nil))
	assert.Empty(t, memberships())
	require.Error(t, store.SyncMemberships(t.Context(), uuid.New(), nil), "an unknown user is rejected")
}
