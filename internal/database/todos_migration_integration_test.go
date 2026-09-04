package database_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

	// Lifecycle steps are project-defined data, so step_kind is an open string
	// column. Only the shape of the name is still a database concern.
	assertOpenStepKind(t, db.Gorm(), issueTwo)

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

// assertOpenStepKind covers the open step_kind domain and the triage backfill
// that predates it. The lifecycle definition (todos/lifecycle/todos.yaml) owns
// which step names exist and a project may add its own, so the database only
// enforces the shape of the name: non-empty, lower-cased and trimmed.
func assertOpenStepKind(t *testing.T, db *gorm.DB, issueID string) {
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
	link := func(kind string, ordinal int, runID string) error {
		return db.Exec(`
			INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
			VALUES (?, ?, ?, ?)`, issueID, runID, kind, ordinal).Error
	}

	triageRun := newTriageRun()
	require.NoError(t, link("triage", 0, triageRun),
		"a built-in lifecycle step must still be accepted")
	require.NoError(t, link("custom-step", 1, newTriageRun()),
		"a project-defined lifecycle step is the whole point of the open domain")

	assertStepKindRejected(t, link("", 2, newTriageRun()),
		"an unnamed step is not a step")
	assertStepKindRejected(t, link("Run", 3, newTriageRun()),
		"a step recorded under a spelling readers never match is a silent orphan")

	// The backfill is a one-time reclassification, so re-running its statement
	// against a plan-linked triage run must move it exactly as the migration did.
	legacyRun := newTriageRun()
	require.NoError(t, link("plan", 7, legacyRun))
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

// assertStepKindRejected demands the step_kind CHECK specifically, so a link
// that fails for an unrelated reason (a missing run, a duplicate ordinal)
// cannot masquerade as domain enforcement.
func assertStepKindRejected(t *testing.T, err error, why string) {
	t.Helper()
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr, why)
	assert.Equal(t, "23514", pgErr.Code, why)
	assert.Equal(t, "todo_issue_prompt_runs_step_kind_check", pgErr.ConstraintName, why)
}

// TestTodoStepKindCheckOpensOnAnExistingDatabase covers the upgrade path the
// fresh-install test cannot reach. Atlas creates a new table straight from the
// HCL, so a changed CHECK expression always looks correct on an empty database;
// on a database that already has the table Atlas silently keeps the old
// expression (sqlx.checksSimilarDiff matches the declared CHECK to the live one
// by name and then only compares NO INHERIT, so ModifyCheck is never emitted).
// The schema bundle has to reconcile the CHECK itself: 102 widens it so the 116
// backfill can write 'triage', and 140 then opens the domain entirely. Applying
// 140 before 116 would leave the live constraint on 102's closed enum, so the
// order is asserted from what actually landed.
func TestTodoStepKindCheckOpensOnAnExistingDatabase(t *testing.T) {
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
	assert.Contains(t, definition, "btrim(step_kind)",
		"the live CHECK must end on the open domain todos.hcl declares")
	assert.NotContains(t, definition, "'triage'::text",
		"140 must land after 102, or the closed enum survives the upgrade")

	var kind string
	require.NoError(t, db.Gorm().Raw(
		`SELECT step_kind FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, legacyRun).Scan(&kind).Error)
	assert.Equal(t, "triage", kind, "the backfill must reclassify the historical triage pass")

	// dependsOn is what orders 140 behind 116; the recorded apply times are the
	// evidence that the declared dependency actually held on this database.
	var applyOrder []struct {
		Path      string
		UpdatedAt time.Time
	}
	require.NoError(t, db.Gorm().Raw(`
		SELECT path, updated_at FROM schema_migration_scripts
		WHERE scope = 'gavel' AND path IN (
			'102_todo_prompt_run_step_kind.sql',
			'116_backfill_triage_step_kind.sql',
			'140_todo_prompt_run_step_open.sql')
		ORDER BY updated_at`).Scan(&applyOrder).Error)
	require.Len(t, applyOrder, 3, "all three step-kind scripts must have re-applied")
	assert.Equal(t, []string{
		"102_todo_prompt_run_step_kind.sql",
		"116_backfill_triage_step_kind.sql",
		"140_todo_prompt_run_step_open.sql",
	}, []string{applyOrder[0].Path, applyOrder[1].Path, applyOrder[2].Path})

	// A project-defined step is the reason the domain opened; it must be
	// insertable on an upgraded database, not only a freshly created one.
	upgradedSession, upgradedRun := uuid.NewString(), uuid.NewString()
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_sessions (id, source, provider, host_id)
		VALUES (?, 'gavel-test', 'test', 'local')`, upgradedSession).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin)
		VALUES (?, ?, ?, 'gavel-test')`, upgradedRun, upgradedSession, upgradedSession).Error)
	require.NoError(t, db.Gorm().Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, 'custom-step', 1)`, issueID, upgradedRun).Error)
}

// TestTodoProjectionDefersStatusToLifecycleHost pins the two consequences of a
// data-driven lifecycle on the SQL projection: durable status now has exactly
// one writer (the Go host's OnOutcome, which records a lifecycle_outcome
// event), and no derivation may key on a hard-coded step name, because a
// project names its own steps.
func TestTodoProjectionDefersStatusToLifecycleHost(t *testing.T) {
	db := openProjectionDatabase(t)
	workspaceID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO todo_workspaces (id, repo_key)
		VALUES (?, 'github.com/flanksource/projection-status')`, workspaceID).Error)

	t.Run("a succeeded verified run leaves durable status untouched", func(t *testing.T) {
		// Exactly the shape the projection used to promote to 'verified': a
		// finished, succeeded 'run' step carrying fixture verification Markdown.
		fixture := newProjectionFixture(t, db, workspaceID, "run", "gavel fixture", `{}`)
		before := todoStatusSnapshot(t, db, fixture.issueID)
		require.Equal(t, "open", before.Status)

		// The prompt-run trigger fires gavel_project_todo_prompt_run, which is
		// the path the projection used to write status through.
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, fixture.runID).Error)

		var changed int
		require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`,
			fixture.runID).Scan(&changed).Error)
		assert.Zero(t, changed, "replaying the projection advances no activity watermark and mutates nothing")

		after := todoStatusSnapshot(t, db, fixture.issueID)
		assert.Equal(t, before.Status, after.Status, "status belongs to the lifecycle host")
		assert.Equal(t, before.Version, after.Version, "no status write means no version bump")
		assert.Equal(t, before.Events, after.Events, "no verification_* event may be appended")

		var verificationEvents int64
		require.NoError(t, db.Raw(`
			SELECT count(*) FROM todo_issue_events
			WHERE issue_id = ? AND kind IN
				('verification_succeeded', 'verification_failed', 'verification_required')`,
			fixture.issueID).Scan(&verificationEvents).Error)
		assert.Zero(t, verificationEvents)

		// Transient state is still derived, so dropping the status writer did not
		// blind the runtime view.
		assert.Equal(t, "idle", executionState(t, db, fixture.issueID))
	})

	t.Run("planning is derived from the rendered spec permission mode", func(t *testing.T) {
		planning := newProjectionFixture(t, db, workspaceID, "shape-it", "", `{
			"permissions":{"mode":"plan"}
		}`)
		assert.Equal(t, "planning", executionState(t, db, planning.issueID),
			"a plan-mode run is a planning pass whatever its project named the step")

		named := newProjectionFixture(t, db, workspaceID, "plan", "", `{}`)
		assert.Equal(t, "running", executionState(t, db, named.issueID),
			"a step merely named 'plan' is not evidence the agent was asked to plan")
	})

	t.Run("verification is derived from the rendered spec shape", func(t *testing.T) {
		checking := newProjectionFixture(t, db, workspaceID, "check-it", "", verifyOnlySpec)
		assert.Equal(t, "verifying", executionState(t, db, checking.issueID),
			"a spec with a definition of done and no prompt is a verification pass whatever its step is named")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, checking.runID).Error)
		assert.Equal(t, "verification_failed", executionState(t, db, checking.issueID))

		named := newProjectionFixture(t, db, workspaceID, "verify", "", `{"prompt":{"user":"implement it"}}`)
		assert.Equal(t, "running", executionState(t, db, named.issueID),
			"a step merely named 'verify' that was handed a prompt is an agent turn, not a verification")
	})
}

type todoStatus struct {
	Status  string
	Version int64
	Events  int64
}

func todoStatusSnapshot(t *testing.T, db *gorm.DB, issueID uuid.UUID) todoStatus {
	t.Helper()
	var snapshot todoStatus
	require.NoError(t, db.Raw(`
		SELECT issue.status, issue.version,
		       (SELECT count(*) FROM todo_issue_events event
		         WHERE event.issue_id = issue.id) AS events
		FROM todo_issues issue WHERE issue.id = ?`, issueID).Scan(&snapshot).Error)
	return snapshot
}

func executionState(t *testing.T, db *gorm.DB, issueID uuid.UUID) string {
	t.Helper()
	var state string
	require.NoError(t, db.Raw(`
		SELECT execution_state FROM todo_issue_runtime WHERE issue_id = ?`,
		issueID).Scan(&state).Error)
	return state
}
