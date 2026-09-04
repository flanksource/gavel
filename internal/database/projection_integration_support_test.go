package database_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCaptainDatabaseOperatesWithoutGavelSchema(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres projection tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "captain_without_gavel",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	captain, err := captaindb.Open(t.Context(), captaindb.WithDSN(dsn), captaindb.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, captain.Close()) })
	require.NoError(t, captaindb.Migrate(t.Context(), dsn), "Captain-only migration must be repeatable")

	sessionID := uuid.New()
	runID := uuid.New()
	require.NoError(t, captain.Gorm().Exec(`
		INSERT INTO captain_sessions
			(id, source, provider, host_id, lifecycle_status, state_version)
		VALUES (?, 'standalone-test', 'test', 'local', 'running', 1)`, sessionID).Error)
	require.NoError(t, captain.Gorm().Exec(`
		INSERT INTO captain_prompt_runs
			(id, session_id, root_session_id, origin, phase, state, version)
		VALUES (?, ?, ?, 'standalone-test', 'generate', 'running', 1)`,
		runID, sessionID, sessionID).Error)
	require.NoError(t, captain.Gorm().Exec(`
		UPDATE captain_prompt_runs
		SET phase = 'finished', state = 'succeeded', version = 2,
			finished_at = now(), updated_at = now()
		WHERE id = ?`, runID).Error)

	var gavelTables int64
	require.NoError(t, captain.Gorm().Raw(`
		SELECT count(*) FROM pg_class AS relation
		JOIN pg_namespace AS namespace ON namespace.oid = relation.relnamespace
		WHERE namespace.nspname = 'public' AND relation.relname LIKE 'todo_%'
	`).Scan(&gavelTables).Error)
	assert.Zero(t, gavelTables)
	var gavelTriggers int64
	require.NoError(t, captain.Gorm().Raw(`
		SELECT count(*) FROM pg_trigger
		WHERE NOT tgisinternal AND tgname LIKE 'gavel_todo_%'
	`).Scan(&gavelTriggers).Error)
	assert.Zero(t, gavelTriggers)
}

type projectionFixture struct {
	issueID            uuid.UUID
	sessionID          uuid.UUID
	admissionSessionID uuid.UUID
	runID              uuid.UUID
}

// newProjectionFixture links one queued prompt run of the named step to a fresh
// open issue: the admission root is the source='gavel' session, the monitored
// agent session shares its provider identity.
func newProjectionFixture(
	t *testing.T,
	db *gorm.DB,
	workspaceID uuid.UUID,
	stepKind string,
	verification string,
	renderedSpec string,
) projectionFixture {
	t.Helper()
	fixture := projectionFixture{
		issueID:            uuid.New(),
		sessionID:          uuid.New(),
		admissionSessionID: uuid.New(),
		runID:              uuid.New(),
	}
	providerSessionID := uuid.New().String()
	require.True(t, json.Valid([]byte(renderedSpec)), "test rendered spec must be valid JSON")
	require.NoError(t, db.Exec(`
		INSERT INTO todo_issues (id, workspace_id, title)
		VALUES (?, ?, ?)`, fixture.issueID, workspaceID, "Projection "+stepKind).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO captain_sessions
			(id, source, provider, provider_session_id, host_id,
			 lifecycle_status, activity_state, health_state)
		VALUES (?, 'gavel', 'test', ?, 'local', 'created', 'idle', 'healthy')`,
		fixture.admissionSessionID, providerSessionID).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO captain_sessions
			(id, source, provider_session_id, host_id,
			 lifecycle_status, activity_state, health_state)
		VALUES (?, 'codex', ?, 'local', 'running', 'working', 'healthy')`,
		fixture.sessionID, providerSessionID).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO captain_prompt_runs
			(id, session_id, root_session_id, origin, verification_markdown,
			 rendered_spec, phase, state, version)
		VALUES (?, ?, ?, 'gavel-test', NULLIF(?, ''), ?::jsonb, 'queued', 'pending', 0)`,
		fixture.runID, fixture.admissionSessionID, fixture.admissionSessionID, verification, renderedSpec).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, ?, 0)`, fixture.issueID, fixture.runID, stepKind).Error)
	require.NoError(t, db.Exec(`
		UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`,
		fixture.runID, fixture.issueID).Error)
	// The run's insert fired the projection before the issue pointed at it, so
	// this is the first projection that reaches the issue: it advances the
	// issue's activity watermark to the run's and reports that one issue moved.
	var changed int
	require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`, fixture.runID).Scan(&changed).Error)
	require.Equal(t, 1, changed, "the first projection carries the run's activity onto its issue")
	return fixture
}

func openProjectionDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres projection tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_captain_projection",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })
	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	return opened.Gorm()
}

type catalogObject struct {
	Kind string
	Name string
	OID  int64
}

func projectionCatalogObjects(t *testing.T, db *gorm.DB) []catalogObject {
	t.Helper()
	var objects []catalogObject
	require.NoError(t, db.Raw(`
		SELECT 'constraint' AS kind, con.conname AS name, con.oid::bigint AS oid
		FROM pg_constraint AS con
		WHERE con.conname IN (
			'todo_issue_prompt_runs_captain_prompt_run_fkey',
			'todo_issue_plans_captain_plan_fkey'
		)
		UNION ALL
		SELECT 'trigger', trg.tgname, trg.oid::bigint
		FROM pg_trigger AS trg
		WHERE NOT trg.tgisinternal
			  AND trg.tgname IN (
				'gavel_todo_prompt_run_projection',
				'gavel_todo_prompt_run_iteration_projection',
				'gavel_todo_session_projection',
				'gavel_todo_session_delete_projection',
			'gavel_todo_turn_request_projection'
		  )
		UNION ALL
		SELECT 'view', relation.relname, relation.oid::bigint
		FROM pg_class AS relation
		JOIN pg_namespace AS namespace ON namespace.oid = relation.relnamespace
		WHERE namespace.nspname = 'public'
		  AND relation.relname IN ('todo_issue_plan_revision_details', 'todo_issue_runtime')
		ORDER BY kind, name
	`).Scan(&objects).Error)
	require.Len(t, objects, 9)
	return objects
}

// assertProjection reads the issue the way the dashboard does — durable status
// from the row, execution state derived through todo_issue_runtime — and pins
// the projection's one invariant: it never appends an issue event. Durable
// status has a single writer, the lifecycle host, and it is the only thing that
// records events; nothing in these flows goes through it, so the log stays
// empty however Captain's rows move.
func assertProjection(t *testing.T, db *gorm.DB, issueID uuid.UUID, wantStatus, wantExecution string) {
	t.Helper()
	var issue struct {
		Status         string
		ExecutionState string
		Events         int64
	}
	require.NoError(t, db.Raw(`
		SELECT issue.status, runtime.execution_state,
		       (SELECT count(*) FROM todo_issue_events event
		         WHERE event.issue_id = issue.id) AS events
		FROM todo_issues issue
		JOIN todo_issue_runtime runtime ON runtime.issue_id = issue.id
		WHERE issue.id = ?`, issueID).Scan(&issue).Error)
	assert.Equal(t, wantStatus, issue.Status, "durable status belongs to the lifecycle host")
	assert.Equal(t, wantExecution, issue.ExecutionState)
	assert.Zero(t, issue.Events, "the projection must not append issue events")
}

func assertPromptRunVersion(t *testing.T, db *gorm.DB, runID uuid.UUID, want int64) {
	t.Helper()
	var got int64
	require.NoError(t, db.Raw(`
		SELECT version FROM captain_prompt_runs WHERE id = ?`, runID,
	).Scan(&got).Error)
	assert.Equal(t, want, got)
}
