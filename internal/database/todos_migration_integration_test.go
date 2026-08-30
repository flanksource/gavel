package database_test

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNativeTodoHCLMigrationAndRepeatedApply(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_todo_migrate",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	require.False(t, db.Disabled())
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	nativeTables := []string{
		"todo_workspaces",
		"todo_workspace_paths",
		"todo_labels",
		"todo_issues",
		"todo_issue_aliases",
		"todo_issue_relationships",
		"todo_issue_events",
		"todo_issue_prompt_runs",
		"todo_issue_plans",
	}
	for _, table := range nativeTables {
		require.True(t, db.Gorm().Migrator().HasTable(table), "%s should exist", table)
	}
	assert.True(t, db.Gorm().Migrator().HasTable("captain_prompt_runs"), "Captain migrations must run before the Gavel bundle")
	assert.True(t, db.Gorm().Migrator().HasTable("captain_plans"), "Captain migrations must share the Gavel database")

	const (
		workspaceOne   = "10000000-0000-0000-0000-000000000001"
		workspaceTwo   = "10000000-0000-0000-0000-000000000002"
		issueOne       = "20000000-0000-0000-0000-000000000001"
		issueTwo       = "20000000-0000-0000-0000-000000000002"
		issueThree     = "20000000-0000-0000-0000-000000000003"
		captainSession = "25000000-0000-0000-0000-000000000001"
		promptRun      = "30000000-0000-0000-0000-000000000001"
		plan           = "40000000-0000-0000-0000-000000000001"
	)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_workspaces (id, repo_key, display_name) VALUES
			(?, 'github.com/flanksource/gavel', 'gavel'),
			(?, NULL, 'non-git workspace')`, workspaceOne, workspaceTwo).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_workspace_paths (workspace_id, path, is_primary) VALUES
			(?, '/work/gavel', true),
			(?, '/work/second', true)`, workspaceOne, workspaceTwo).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_workspace_paths (workspace_id, path, is_primary)
		VALUES (?, '/work/gavel-moved', true)`, workspaceOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_workspace_paths (workspace_id, path, is_primary)
		VALUES (?, '/work/gavel', false)`, workspaceTwo).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_workspaces (repo_key) VALUES (' GitHub.com/Invalid/Repo ')`).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_workspace_paths (workspace_id, path, is_primary)
		VALUES (?, ' /work/not-normalized ', false)`, workspaceOne).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issues (id, workspace_id, title, labels) VALUES
			(?, ?, 'First issue', ARRAY['database', 'todos']),
			(?, ?, 'Second issue', ARRAY[]::text[]),
			(?, ?, 'Other workspace issue', ARRAY[]::text[])`,
		issueOne, workspaceOne, issueTwo, workspaceOne, issueThree, workspaceTwo).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_sessions (id, source, provider, host_id)
		VALUES (?, 'gavel-test', 'test', 'local')`, captainSession).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin)
		VALUES (?, ?, ?, 'gavel-test')`, promptRun, captainSession, captainSession).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_plans (id, source_session_id, source_prompt_run_id, title)
		VALUES (?, ?, ?, 'Migration fixture plan')`, plan, captainSession, promptRun).Error)

	// The alias primary key is the concurrency-safe uniqueness boundary, and
	// the composite foreign key prevents cross-workspace aliases.
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind)
		VALUES (?, 'e2a3b8c2d0f7c9a98b400dc78e8a94a5', ?, 'external')`, workspaceOne, issueOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind)
		VALUES (?, 'e2a3b8c2d0f7c9a98b400dc78e8a94a5', ?, 'external')`, workspaceOne, issueTwo).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind)
		VALUES (?, 'cross-workspace', ?, 'legacy')`, workspaceTwo, issueOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_aliases (workspace_id, alias, issue_id, kind)
		VALUES (?, 'NOT-NORMALIZED', ?, 'EXTERNAL')`, workspaceOne, issueTwo).Error)

	// Active pointers are valid only after the corresponding same-issue link
	// exists, and link rows must reference authoritative Captain records.
	assert.Error(t, db.Gorm().Exec(`
		UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`, promptRun, issueOne).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'run', 0)`, issueOne, promptRun).Error)
	require.NoError(t, db.Gorm().Exec(`
		UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`, promptRun, issueOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'run', 0)`, issueTwo, promptRun).Error)

	// Triage earned its own step kind so a backlog can tell a triage pass from a
	// planning pass; anything outside the four is still refused.
	assertTriageStepKind(t, db.Gorm(), issueTwo, captainSession)

	assert.Error(t, db.Gorm().Exec(`
		UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`, plan, issueOne).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal)
		VALUES (?, ?, 0)`, issueOne, plan).Error)
	require.NoError(t, db.Gorm().Exec(`
		UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`, plan, issueOne).Error)

	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
		VALUES (?, ?, ?, 'depends_on')`, workspaceOne, issueOne, issueOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
		VALUES (?, ?, ?, 'related_to')`, workspaceOne, issueTwo, issueOne).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
		VALUES (?, ?, ?, 'related_to')`, workspaceOne, issueOne, issueTwo).Error)
	// Both endpoint FKs include workspace_id, so direct SQL cannot create an
	// edge whose target belongs to another workspace.
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
		VALUES (?, ?, ?, 'depends_on')`, workspaceOne, issueOne, issueThree).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_relationships (workspace_id, issue_id, target_issue_id, relation)
		VALUES (?, ?, ?, 'depends_on')`, workspaceTwo, issueOne, issueThree).Error)

	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_events (issue_id, sequence, kind, source, source_id)
		VALUES (?, 1, 'created', 'portable-import', 'legacy-event-1')`, issueOne).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_events (issue_id, sequence, kind, source, source_id)
		VALUES (?, 1, 'created', 'portable-import', 'legacy-event-1')`, issueTwo).Error)
	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_events (issue_id, sequence, kind, source)
		VALUES (?, 2, 'Not-Normalized', 'PORTABLE-IMPORT')`, issueOne).Error)

	assert.Error(t, db.Gorm().Exec(`
		INSERT INTO todo_issues (workspace_id, title, priority)
		VALUES (?, 'Invalid priority', 'urgent')`, workspaceOne).Error)

	second, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "reapplying the unchanged HCL bundle should be idempotent")
	require.NoError(t, second.Close())

	var issueCount int64
	require.NoError(t, db.Gorm().Table("todo_issues").Count(&issueCount).Error)
	assert.EqualValues(t, 3, issueCount, "repeated apply must preserve native issue data")

	// Remove only the native surface and prove the declarative bundle recreates it.
	require.NoError(t, db.Gorm().Exec(`DROP TABLE
		todo_issue_events,
		todo_issue_aliases,
		todo_issue_relationships,
		todo_issue_prompt_runs,
		todo_issue_plans,
		todo_issues,
		todo_workspace_paths,
		todo_workspaces CASCADE`).Error)
	fresh, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "the HCL bundle should create the native TODO schema from scratch")
	require.NoError(t, fresh.Close())
	for _, table := range nativeTables {
		require.True(t, db.Gorm().Migrator().HasTable(table), "%s should be recreated", table)
	}
}

// assertTriageStepKind covers the widened step_kind CHECK and the backfill that
// accompanies it. Triage has always been a plan-CLASS run, so before it earned
// its own kind every triage pass was linked as step_kind='plan' and was
// distinguishable only by the prompt name Captain recorded (spec_profile).
func assertTriageStepKind(t *testing.T, db *gorm.DB, issueID, _ string) {
	t.Helper()

	// Each run gets its own root session: captain_prompt_runs_active_root_key
	// allows only one active run per root, so reusing the fixture's session
	// would collide rather than exercise the step kind.
	newTriageRun := func() string {
		session, run := uuid.NewString(), uuid.NewString()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_sessions (id, source, provider, host_id)
			VALUES (?, 'gavel-test', 'test', 'local')`, session).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, spec_profile)
			VALUES (?, ?, ?, 'gavel-test', 'triage')`, run, session, session).Error)
		return run
	}

	triageRun := newTriageRun()
	require.NoError(t, db.Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'triage', 0)`, issueID, triageRun).Error,
		"the widened CHECK must accept triage as its own step kind")

	assert.Error(t, db.Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'compact', 1)`, issueID, uuid.NewString()).Error,
		"the CHECK must still refuse a kind outside the four")

	// The backfill is a one-time reclassification, so re-running its statement
	// against a plan-linked triage run must move it exactly as the migration did.
	legacyRun := newTriageRun()
	require.NoError(t, db.Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'plan', 7)`, issueID, legacyRun).Error)
	require.NoError(t, db.Exec(`
		UPDATE todo_issue_prompt_runs AS link
		SET step_kind = 'triage'
		FROM captain_prompt_runs AS run
		WHERE run.id = link.prompt_run_id
		  AND link.step_kind = 'plan'
		  AND run.spec_profile = 'triage'`).Error)

	var kind string
	require.NoError(t, db.Raw(
		`SELECT step_kind FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, legacyRun).Scan(&kind).Error)
	assert.Equal(t, "triage", kind, "a historical triage run must not keep reporting as a planning pass")
}

// TestTodoStepKindCheckWidensOnAnExistingDatabase covers the upgrade path the
// fresh-install test cannot reach. Atlas creates a new table straight from the
// HCL, so a widened CHECK expression always looks correct on an empty database;
// on a database that already has the table Atlas silently keeps the old
// expression (sqlx.checksSimilarDiff matches the declared CHECK to the live one
// by name and then only compares NO INHERIT, so ModifyCheck is never emitted).
// The schema bundle has to reconcile the CHECK itself, before the backfill
// writes the new kind.
func TestTodoStepKindCheckWidensOnAnExistingDatabase(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_step_kind_widen",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())

	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	// Rewind to the shape a database provisioned before triage had its own step
	// kind is in: the narrow CHECK, a triage pass linked as a planning pass, and
	// no record that the triage scripts ever ran.
	require.NoError(t, db.Gorm().Exec(`
		ALTER TABLE public.todo_issue_prompt_runs
			DROP CONSTRAINT todo_issue_prompt_runs_step_kind_check;
		ALTER TABLE public.todo_issue_prompt_runs
			ADD CONSTRAINT todo_issue_prompt_runs_step_kind_check
			CHECK (step_kind = ANY (ARRAY['plan'::text, 'run'::text, 'verify'::text]))`).Error)

	workspaceID, issueID := uuid.NewString(), uuid.NewString()
	sessionID, legacyRun := uuid.NewString(), uuid.NewString()
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_workspaces (id, repo_key, display_name)
		VALUES (?, 'github.com/flanksource/gavel', 'gavel')`, workspaceID).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issues (id, workspace_id, title) VALUES (?, ?, 'legacy triage pass')`,
		issueID, workspaceID).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_sessions (id, source, provider, host_id)
		VALUES (?, 'gavel-test', 'test', 'local')`, sessionID).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, spec_profile)
		VALUES (?, ?, ?, 'gavel-test', 'triage')`, legacyRun, sessionID, sessionID).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'plan', 0)`, issueID, legacyRun).Error)
	require.NoError(t, db.Gorm().Exec(`
		DELETE FROM schema_migration_scripts
		WHERE scope = 'gavel' AND path IN (
			'102_todo_prompt_run_step_kind.sql',
			'116_backfill_triage_step_kind.sql')`).Error)

	upgraded, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err, "migrating a pre-triage database must widen the CHECK before backfilling it")
	require.NoError(t, upgraded.Close())

	var definition string
	require.NoError(t, db.Gorm().Raw(`
		SELECT pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conrelid = 'public.todo_issue_prompt_runs'::regclass
		  AND conname = 'todo_issue_prompt_runs_step_kind_check'`).Scan(&definition).Error)
	assert.Contains(t, definition, "'triage'::text", "the live CHECK must match the widened HCL declaration")

	var kind string
	require.NoError(t, db.Gorm().Raw(
		`SELECT step_kind FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, legacyRun).Scan(&kind).Error)
	assert.Equal(t, "triage", kind, "the backfill must reclassify the historical triage pass")
}
