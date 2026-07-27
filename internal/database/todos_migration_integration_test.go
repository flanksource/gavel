package database_test

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
