// Tests the upgrade from main's single-tenant schema to the organization-scoped schema.
// Existing rows must land in the local organization, and keys must allow per-org duplicates.
package database_test

import (
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"os"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/commons-db/migrate"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/database/dbtest"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// tenantTables are the tables whose rows each belong to exactly one organization.
var tenantTables = []string{
	"todo_workspaces", "todo_labels",
	"session_owners", "api_token_owners", "runtime_profile_owners", "runtime_preset_owners", "git_agent_task_owners",
	"http_cache_entries", "workflow_run_caches", "job_log_caches", "workflow_def_caches", "seen_prs", "pull_requests",
	"test_run_caches", "test_run_cursors",
	"file_scans", "violations", "linter_executions", "debounce_metadata", "task_run_history",
	"git_repos", "repos", "github_installations", "sandbox_runs", "org_memberships",
}

func TestTenancyMigrationFromMainSchema(t *testing.T) {
	dsn := commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_tenancy"}).DSN()
	ctx := t.Context()

	// The schema as main's binary leaves it: testdata/schema_main is a snapshot of
	// main's internal/database/schema, applied with main's exclusions.
	require.NoError(t, captaindb.Migrate(ctx, dsn))
	require.NoError(t, migrate.Apply(ctx, dsn, os.DirFS("testdata"),
		migrate.WithDir("schema_main"),
		migrate.WithName("gavel"),
		migrate.WithExclude(
			"captain_*",
			"todo_issue_plan_revision_details",
			"todo_issue_runtime",
		),
	))
	legacy, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	require.NoError(t, legacy.Exec(`
		INSERT INTO seen_prs (repo, number, seen_at) VALUES ('a/b', 1, now());
		INSERT INTO pull_requests (provider, repo, number, url, title, synced_at) VALUES ('github', 'a/b', 1, 'https://github.com/a/b/pull/1', 'Fix', now());
		INSERT INTO file_scans (file_path, last_scan_time, file_mod_time, file_hash) VALUES ('/src/x.go', 1, 1, 'hash');
		INSERT INTO violations (file_path, line, "column", source, severity, fixable, fix_applicability) VALUES ('/src/x.go', 1, 1, 'lint', 'error', false, '');
		INSERT INTO todo_workspaces (repo_key) VALUES ('github.com/a/b');
		INSERT INTO git_repos (root_dir, common_dir) VALUES ('/src/repo', '/src/repo/.git');
	`).Error)
	legacySQL, err := legacy.DB()
	require.NoError(t, err)
	require.NoError(t, legacySQL.Close())

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	db, err := database.Open(ctx, database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	gdb := db.Gorm()

	var local struct{ Issuer, ClerkOrgID, Slug, Name string }
	require.NoError(t, gdb.Raw(`SELECT issuer, clerk_org_id, slug, name FROM orgs WHERE id = ?`, tenant.LocalOrgID).Scan(&local).Error)
	require.Equal(t, "local", local.Issuer)
	require.Equal(t, "local", local.ClerkOrgID)
	require.Equal(t, "local", local.Slug)
	require.Equal(t, "Local", local.Name)

	for _, table := range []string{"seen_prs", "pull_requests", "file_scans", "violations", "todo_workspaces", "git_repos"} {
		var orgs []uuid.UUID
		require.NoError(t, gdb.Raw(`SELECT org_id FROM `+table).Scan(&orgs).Error, table)
		require.Equal(t, []uuid.UUID{tenant.LocalOrgID}, orgs, "existing %s rows belong to the local org", table)
	}

	for _, table := range tenantTables {
		var nullable string
		require.NoError(t, gdb.Raw(`SELECT is_nullable FROM information_schema.columns WHERE table_schema = 'public' AND table_name = ? AND column_name = 'org_id'`, table).Scan(&nullable).Error, table)
		require.Equal(t, "NO", nullable, "%s.org_id is NOT NULL", table)
		var orgFKs int64
		require.NoError(t, gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conrelid = ('public.' || ?)::regclass AND contype = 'f' AND confrelid = 'public.orgs'::regclass`, table).Scan(&orgFKs).Error, table)
		require.EqualValues(t, 1, orgFKs, "%s has one org foreign key", table)
	}

	dbtest.RequireConstraintViolation(t,
		dbtest.Quiet(gdb).Exec(`INSERT INTO seen_prs (org_id, repo, number, seen_at) VALUES (?, 'a/b', 2, now())`, uuid.New()).Error,
		"seen_prs_org_id_fkey")

	other := uuid.New()
	require.NoError(t, gdb.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, slug, name) VALUES (?, 'https://clerk.example', 'org_other', 'other', 'Other')`, other).Error)
	require.NoError(t, gdb.Exec(`INSERT INTO seen_prs (org_id, repo, number, seen_at) VALUES (?, 'a/b', 1, now())`, other).Error,
		"another org can hold the same pull request key")
	require.NoError(t, gdb.Exec(`INSERT INTO pull_requests (org_id, provider, repo, number, url, title, synced_at) VALUES (?, 'github', 'a/b', 1, 'https://github.com/a/b/pull/1', 'Fix', now())`, other).Error,
		"another org can index the same pull request")
	require.NoError(t, gdb.Exec(`INSERT INTO todo_workspaces (org_id, repo_key) VALUES (?, 'github.com/a/b')`, other).Error,
		"another org can hold the same workspace repo key")
	require.NoError(t, gdb.Exec(`INSERT INTO todo_labels (org_id, name, color) VALUES (?, 'bug', 'red'), (?, 'bug', 'red')`, tenant.LocalOrgID, other).Error,
		"global labels with the same name coexist across orgs")
	dbtest.RequireConstraintViolation(t,
		dbtest.Quiet(gdb).Exec(`INSERT INTO violations (org_id, file_path, line, "column", source, severity, fixable, fix_applicability) VALUES (?, '/src/x.go', 1, 1, 'lint', 'error', false, '')`, other).Error,
		"violations_file_scan_fkey")
	require.NoError(t, gdb.Exec(`INSERT INTO file_scans (org_id, file_path, last_scan_time, file_mod_time, file_hash) VALUES (?, '/src/x.go', 1, 1, 'hash')`, other).Error,
		"another org can scan the same path")
	require.NoError(t, gdb.Exec(`INSERT INTO violations (org_id, file_path, line, "column", source, severity, fixable, fix_applicability) VALUES (?, '/src/x.go', 1, 1, 'lint', 'error', false, '')`, other).Error)
	require.NoError(t, gdb.Exec(`DELETE FROM file_scans WHERE org_id = ?`, tenant.LocalOrgID).Error)
	var remaining []uuid.UUID
	require.NoError(t, gdb.Raw(`SELECT org_id FROM violations`).Scan(&remaining).Error)
	require.Equal(t, []uuid.UUID{other}, remaining, "deleting a file scan cascades to its violations in the same org only")

	second, err := database.Open(ctx, database.WithMigrations())
	require.NoError(t, err, "re-running migrations is a no-op")
	require.NoError(t, second.Close())
}
