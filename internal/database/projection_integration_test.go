package database_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCaptainProjectionLifecycleAndMigrationIdempotence(t *testing.T) {
	db := openProjectionDatabase(t)
	ctx := t.Context()

	before := projectionCatalogObjects(t, db)
	require.NotEmpty(t, before)
	second, err := database.Open(ctx, database.WithMigrations())
	require.NoError(t, err, "reapplying the projection bundle must be idempotent")
	require.NoError(t, second.Close())
	after := projectionCatalogObjects(t, db)
	assert.Equal(t, before, after, "repeated apply must not drop and recreate SQL-owned objects")

	var scriptRows int64
	require.NoError(t, db.Raw(`
		SELECT count(*) FROM schema_migration_scripts
		WHERE scope = 'gavel'
		  AND path IN (
		    '100_todo_captain_constraints.sql',
		    '105_view_todo_plan_revisions.sql',
		    '110_todo_projection_functions.sql',
		    '111_todo_projection_triggers.sql',
		    '112_view_todo_issue_runtime.sql'
		  )
	`).Scan(&scriptRows).Error)
	assert.EqualValues(t, 5, scriptRows)

	workspaceID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO todo_workspaces (id, repo_key)
		VALUES (?, 'github.com/flanksource/projection-test')`, workspaceID).Error)

	t.Run("run verify failure success cancellation and stale versions", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "gavel fixture", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")

		// Replaying the same authoritative snapshot is a complete no-op.
		var changed int
		require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`, fixture.runID).Scan(&changed).Error)
		assert.Zero(t, changed)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET phase = 'verify', state = 'running', version = 1, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "verifying", 2, "execution_state_changed")

		// Captain rejects an out-of-order optimistic update before it can
		// overwrite the authoritative row or regress the Gavel projection.
		captain, err := captaindb.Use(db)
		require.NoError(t, err)
		generate := captaindb.PromptRunPhaseGenerate
		running := captaindb.PromptRunStateRunning
		_, err = captain.UpdatePromptRun(t.Context(), captaindb.UpdatePromptRunInput{
			ID: fixture.runID, ExpectedVersion: 0, Phase: &generate, State: &running,
		})
		require.ErrorIs(t, err, captaindb.ErrPromptRunConflict)
		assertProjection(t, db, fixture.issueID, "open", "verifying", 2, "execution_state_changed")

		// A failed verification reopens a previously verified issue.
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, fixture.issueID).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', version = 3, finished_at = now(), updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "verification_failed", 3, "verification_failed")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'cancelled', phase = 'finished', version = 4, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "idle", 4, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 5, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "idle", 5, "verification_succeeded")
		assertProjectionAuditVersions(t, db, fixture.issueID, fixture.runID, 4)
	})

	t.Run("requests session health and descendant deletion", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")

		questionID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_turn_requests
				(id, session_id, prompt_run_id, kind, state, request, version)
			VALUES (?, ?, ?, 'question', 'pending', '{}'::jsonb, 0)`,
			questionID, fixture.sessionID, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting", 2, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_turn_requests
			SET state = 'answered', response = '{"answer":"yes"}'::jsonb,
				version = 1, resolved_at = now()
			WHERE id = ?`, questionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 3, "execution_state_changed")

		approvalID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_turn_requests
				(id, session_id, prompt_run_id, kind, state, request, version)
			VALUES (?, ?, ?, 'tool_approval', 'pending', '{}'::jsonb, 0)`,
			approvalID, fixture.sessionID, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting", 4, "execution_state_changed")
		require.NoError(t, db.Exec(`DELETE FROM captain_turn_requests WHERE id = ?`, approvalID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 5, "execution_state_changed")

		childID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_sessions
				(id, source, provider, host_id, parent_session_id, root_session_id,
				 activity_state, health_state, state_version)
			VALUES (?, 'test', 'test', 'local', ?, ?, 'working', 'stalled', 0)`,
			childID, fixture.sessionID, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled", 6, "execution_state_changed")

		// Removing the descendant that supplied the stalled signal recomputes
		// from its surviving parent instead of leaving the issue stale.
		require.NoError(t, db.Exec(`DELETE FROM captain_sessions WHERE id = ?`, childID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 7, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'ask', state_version = 1, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting", 8, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'working', health_state = 'stalled',
				state_version = 2, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled", 9, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET health_state = 'zombie', state_version = 3, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "failed", 10, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET state_version = 2, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "failed", 10, "execution_state_changed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET lifecycle_status = 'cancelled', health_state = 'healthy',
				state_version = 4, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "idle", 11, "execution_state_changed")
	})

	t.Run("agent activity is monotonic without issue versions or events", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		var before struct {
			UpdatedAt  time.Time
			Version    int64
			EventCount int64
		}
		require.NoError(t, db.Raw(`
			SELECT issue.updated_at, issue.version,
			       (SELECT count(*) FROM todo_issue_events event WHERE event.issue_id = issue.id) AS event_count
			FROM todo_issues issue WHERE issue.id = ?`, fixture.issueID).Scan(&before).Error)

		activityAt := before.UpdatedAt.Add(2 * time.Minute)
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'ask', state_version = state_version + 1,
			    state_observed_at = ?, last_activity_at = ?, updated_at = ?
			WHERE id = ?`, activityAt, activityAt, activityAt, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting", 0, "execution_state_changed")

		var after struct {
			UpdatedAt  time.Time
			Version    int64
			EventCount int64
		}
		require.NoError(t, db.Raw(`
			SELECT issue.updated_at, issue.version,
			       (SELECT count(*) FROM todo_issue_events event WHERE event.issue_id = issue.id) AS event_count
			FROM todo_issues issue WHERE issue.id = ?`, fixture.issueID).Scan(&after).Error)
		assert.Equal(t, activityAt.UTC(), after.UpdatedAt.UTC())
		assert.Equal(t, before.Version, after.Version)
		assert.Equal(t, before.EventCount, after.EventCount)

		older := activityAt.Add(-time.Minute)
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions SET last_activity_at = ?, updated_at = ? WHERE id = ?`,
			older, older, fixture.sessionID).Error)
		var watermark time.Time
		require.NoError(t, db.Raw(`SELECT updated_at FROM todo_issues WHERE id = ?`, fixture.issueID).Scan(&watermark).Error)
		assert.Equal(t, activityAt.UTC(), watermark.UTC(), "older Captain observations must not regress issue activity")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'working', health_state = 'healthy',
			    state_version = state_version + 1, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 0, "execution_state_changed")

		// Admission state is intentionally ignored by runtime state derivation.
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'approval', health_state = 'zombie',
			    state_version = state_version + 1, updated_at = now()
			WHERE id = ?`, fixture.admissionSessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 0, "execution_state_changed")
	})

	t.Run("planning verify and explicit no-fixture policy", func(t *testing.T) {
		plan := newProjectionFixture(t, db, workspaceID, "plan", "plan fixture must not verify", `{
			"workflow":{"autoVerifyWithoutFixture":true}
		}`)
		assertProjection(t, db, plan.issueID, "open", "planning", 1, "execution_state_changed")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, plan.runID).Error)
		assertProjection(t, db, plan.issueID, "open", "idle", 2, "execution_state_changed")

		verify := newProjectionFixture(t, db, workspaceID, "verify", "", `{}`)
		assertProjection(t, db, verify.issueID, "open", "verifying", 1, "execution_state_changed")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, verify.runID).Error)
		assertProjection(t, db, verify.issueID, "verified", "idle", 2, "verification_succeeded")

		noPolicy := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, noPolicy.issueID).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, noPolicy.runID).Error)
		assertProjection(t, db, noPolicy.issueID, "open", "idle", 2, "verification_required")

		autoVerify := newProjectionFixture(t, db, workspaceID, "run", "", `{
			"workflow":{"autoVerifyWithoutFixture":true}
		}`)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, autoVerify.runID).Error)
		assertProjection(t, db, autoVerify.issueID, "verified", "idle", 2, "verification_succeeded")

		malformedPolicy := newProjectionFixture(t, db, workspaceID, "run", "", `{
			"workflow":{"autoVerifyWithoutFixture":"not-a-boolean"}
		}`)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, malformedPolicy.runID).Error)
		assertProjection(t, db, malformedPolicy.issueID, "open", "idle", 2, "execution_state_changed")

		stringTruePolicy := newProjectionFixture(t, db, workspaceID, "run", "", `{
			"workflow":{"autoVerifyWithoutFixture":"true"}
		}`)
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, stringTruePolicy.issueID).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, stringTruePolicy.runID).Error)
		assertProjection(t, db, stringTruePolicy.issueID, "open", "idle", 2, "verification_required")
	})

	t.Run("terminal policy inputs reproject without prompt version bump", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs
				SET state = 'succeeded', phase = 'finished', version = 1,
					finished_at = now(), updated_at = now()
				WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "idle", 2, "execution_state_changed")

		var runVersion int64
		require.NoError(t, db.Raw(`
				SELECT version FROM captain_prompt_runs WHERE id = ?`, fixture.runID,
		).Scan(&runVersion).Error)
		require.EqualValues(t, 1, runVersion)

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs
				SET rendered_spec = '{"workflow":{"autoVerifyWithoutFixture":true}}'::jsonb
				WHERE id = ?`, fixture.runID).Error)
		assertPromptRunVersion(t, db, fixture.runID, runVersion)
		assertProjection(t, db, fixture.issueID, "verified", "idle", 3, "verification_succeeded")

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs SET rendered_spec = '{}'::jsonb WHERE id = ?`,
			fixture.runID).Error)
		assertPromptRunVersion(t, db, fixture.runID, runVersion)
		assertProjection(t, db, fixture.issueID, "open", "idle", 4, "verification_required")

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs SET verification_markdown = 'gavel fixture' WHERE id = ?`,
			fixture.runID).Error)
		assertPromptRunVersion(t, db, fixture.runID, runVersion)
		assertProjection(t, db, fixture.issueID, "verified", "idle", 5, "verification_succeeded")
	})

	t.Run("finished verification failure retains iteration provenance", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "gavel fixture", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, fixture.issueID).Error)

		// A finished failure without durable verifier evidence is a generic run
		// failure, even though this run supplied verification Markdown.
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "failed", 2, "execution_state_changed")

		// The latest failed iteration carries durable verification provenance.
		// Its trigger corrects projection after cleanup advanced the phase.
		require.NoError(t, db.Exec(`
			INSERT INTO captain_prompt_run_iterations
				(id, prompt_run_id, iteration, state, verification_result,
				 started_at, finished_at)
			VALUES (?, ?, 0, 'failed', '{"success":false}'::jsonb, now(), now())`,
			uuid.New(), fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "verification_failed", 3, "verification_failed")

		// A later successful iteration supersedes the failed attempt. An
		// unrelated post-run failure must remain generic.
		require.NoError(t, db.Exec(`
			INSERT INTO captain_prompt_run_iterations
				(id, prompt_run_id, iteration, state, verification_result,
				 started_at, finished_at)
			VALUES (?, ?, 1, 'succeeded', '{"success":true}'::jsonb, now(), now())`,
			uuid.New(), fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "failed", 4, "execution_state_changed")
	})

	t.Run("session root reassignment reprojects the old aggregate", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")

		otherRootID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_sessions (id, source, provider, host_id)
			VALUES (?, 'test', 'test', 'local')`, otherRootID).Error)
		childID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_sessions
				(id, source, provider, host_id, parent_session_id, root_session_id,
				 activity_state, health_state)
			VALUES (?, 'test', 'test', 'local', ?, ?, 'working', 'stalled')`,
			childID, fixture.sessionID, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled", 2, "execution_state_changed")

		// Parent/root changes do not increment Captain state_version, so they
		// must be explicit trigger columns and route both the old and new roots.
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET parent_session_id = ?, root_session_id = ?
			WHERE id = ?`, otherRootID, otherRootID, childID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running", 3, "execution_state_changed")
	})

	t.Run("prompt run reassignment reprojects without prompt version bump", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")
		require.NoError(t, db.Exec(`
				UPDATE captain_sessions
				SET health_state = 'stalled', state_version = 1, updated_at = now()
				WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled", 2, "execution_state_changed")

		otherRootID := uuid.New()
		require.NoError(t, db.Exec(`
				INSERT INTO captain_sessions
					(id, source, provider, host_id, lifecycle_status, activity_state, health_state)
				VALUES (?, 'test', 'test', 'local', 'running', 'working', 'healthy')`,
			otherRootID).Error)
		var runVersion int64
		require.NoError(t, db.Raw(`
				SELECT version FROM captain_prompt_runs WHERE id = ?`, fixture.runID,
		).Scan(&runVersion).Error)

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs
				SET session_id = ?, root_session_id = ?
				WHERE id = ?`, otherRootID, otherRootID, fixture.runID).Error)
		assertPromptRunVersion(t, db, fixture.runID, runVersion)
		assertProjection(t, db, fixture.issueID, "open", "running", 3, "execution_state_changed")
	})

	t.Run("active pointer clearing is explicitly projectable", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running", 1, "execution_state_changed")
		require.NoError(t, db.Exec(`
			UPDATE todo_issues SET active_prompt_run_id = NULL WHERE id = ?`, fixture.issueID).Error)
		var changed bool
		require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_issue(?)`, fixture.issueID).Scan(&changed).Error)
		assert.False(t, changed, "clearing transient runtime state is not a durable issue mutation")
		assertProjection(t, db, fixture.issueID, "open", "idle", 2, "execution_state_changed")
	})

	t.Run("foreign keys and durable plan revision view", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "plan", "", `{}`)
		planID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_plans
				(id, source_session_id, source_prompt_run_id, title, path)
			VALUES (?, ?, ?, 'Durable plan', '/tmp/deleted-plan.md')`,
			planID, fixture.sessionID, fixture.runID).Error)
		revisionID := uuid.New()
		const planMarkdown = "# Durable plan\n\nCaptain revision content survives without its file."
		require.NoError(t, db.Exec(`
			INSERT INTO captain_plan_revisions
				(id, plan_id, revision, plan_markdown, content_hash, created_by)
			VALUES (?, ?, 1, ?, 'revision-hash-1', 'test')`,
			revisionID, planID, planMarkdown).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_plans
			SET approved_revision_id = ?, approval_state = 'approved',
				approval_created_at = now(), updated_at = now()
			WHERE id = ?`, revisionID, planID).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal)
			VALUES (?, ?, 0)`, fixture.issueID, planID).Error)
		require.NoError(t, db.Exec(`
			UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`,
			planID, fixture.issueID).Error)

		var detail struct {
			PlanMarkdown string
			Approved     bool
			Selected     bool
		}
		require.NoError(t, db.Raw(`
			SELECT plan_markdown, approved, selected
			FROM todo_issue_plan_revision_details
			WHERE issue_id = ? AND plan_id = ? AND revision = 1`,
			fixture.issueID, planID).Scan(&detail).Error)
		assert.Equal(t, planMarkdown, detail.PlanMarkdown)
		assert.True(t, detail.Approved)
		assert.True(t, detail.Selected)

		pendingPlanID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_plans
				(id, source_session_id, source_prompt_run_id, title)
			VALUES (?, ?, ?, 'Pending alternative')`,
			pendingPlanID, fixture.sessionID, fixture.runID).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO captain_plan_revisions
				(id, plan_id, revision, plan_markdown, content_hash)
			VALUES (?, ?, 1, '# Pending alternative', 'pending-revision-hash')`,
			uuid.New(), pendingPlanID).Error)
		require.NoError(t, db.Exec(`
			INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal)
			VALUES (?, ?, 1)`, fixture.issueID, pendingPlanID).Error)
		var pendingDetail struct {
			Approved bool
			Selected bool
		}
		require.NoError(t, db.Raw(`
			SELECT approved, selected
			FROM todo_issue_plan_revision_details
			WHERE issue_id = ? AND plan_id = ? AND revision = 1`,
			fixture.issueID, pendingPlanID).Scan(&pendingDetail).Error)
		assert.False(t, pendingDetail.Approved)
		assert.False(t, pendingDetail.Selected)

		assert.Error(t, db.Exec(`DELETE FROM captain_plans WHERE id = ?`, planID).Error,
			"a linked Captain plan must be explicitly unlinked before deletion")
		assert.Error(t, db.Exec(`DELETE FROM captain_prompt_runs WHERE id = ?`, fixture.runID).Error,
			"a linked Captain prompt run must be explicitly unlinked before deletion")
		assert.Error(t, db.Exec(`
			INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
			VALUES (?, ?, 'run', 99)`, fixture.issueID, uuid.New()).Error)
		assert.Error(t, db.Exec(`
			INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal)
			VALUES (?, ?, 99)`, fixture.issueID, uuid.New()).Error)
	})
}

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
	var changed int
	require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`, fixture.runID).Scan(&changed).Error)
	require.Zero(t, changed)
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

func assertProjection(
	t *testing.T,
	db *gorm.DB,
	issueID uuid.UUID,
	wantStatus string,
	wantExecution string,
	wantVersion int64,
	wantLastKind string,
) {
	t.Helper()
	var issue struct {
		Status         string
		ExecutionState string
		Version        int64
	}
	require.NoError(t, db.Raw(`
		SELECT issue.status, runtime.execution_state, issue.version
		FROM todo_issues issue
		JOIN todo_issue_runtime runtime ON runtime.issue_id = issue.id
		WHERE issue.id = ?`, issueID).Scan(&issue).Error)
	assert.Equal(t, wantStatus, issue.Status)
	assert.Equal(t, wantExecution, issue.ExecutionState)
	_ = wantVersion // Historical projection versions were transient-state changes.
	if wantLastKind == "execution_state_changed" {
		return
	}

	var event struct {
		Kind   string
		Source string
	}
	require.NoError(t, db.Raw(`
		SELECT kind, source FROM todo_issue_events
		WHERE issue_id = ? ORDER BY sequence DESC LIMIT 1`, issueID).Scan(&event).Error)
	assert.Equal(t, wantLastKind, event.Kind)
	assert.Equal(t, "captain-projection", event.Source)
	var eventCount int64
	require.NoError(t, db.Raw(`
		SELECT count(*) FROM todo_issue_events WHERE issue_id = ?`, issueID).Scan(&eventCount).Error)
	assert.Equal(t, issue.Version, eventCount, "durable status versions must have exactly one visible event")
}

func assertPromptRunVersion(t *testing.T, db *gorm.DB, runID uuid.UUID, want int64) {
	t.Helper()
	var got int64
	require.NoError(t, db.Raw(`
		SELECT version FROM captain_prompt_runs WHERE id = ?`, runID,
	).Scan(&got).Error)
	assert.Equal(t, want, got)
}

func assertProjectionAuditVersions(t *testing.T, db *gorm.DB, issueID, runID uuid.UUID, runVersion int64) {
	t.Helper()
	var event struct {
		Payload  []byte
		SourceID string
	}
	require.NoError(t, db.Raw(`
		SELECT payload, source_id FROM todo_issue_events
		WHERE issue_id = ? ORDER BY sequence DESC LIMIT 1`, issueID).Scan(&event).Error)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(event.Payload, &payload))
	assert.Equal(t, runID.String(), payload["promptRunId"])
	assert.Equal(t, float64(runVersion), payload["promptRunVersion"])
	assert.Contains(t, event.SourceID, "prompt-run:"+runID.String()+":version:")
}
