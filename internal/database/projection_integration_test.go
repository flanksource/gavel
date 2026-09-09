package database_test

import (
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verifyOnlySpec is the rendered spec the lifecycle's verify-only step
// dispatches: a definition of done and no prompt.
const verifyOnlySpec = `{"workflow":{"verify":{"fixture":"# checks"}}}`

// The projection derives transient execution state from Captain's rows and
// propagates activity timestamps. It never writes durable status: that column
// has exactly one writer, the Go lifecycle host, which records a
// lifecycle_outcome event alongside every status it decides. Every flow below
// therefore expects the status it started with and an empty event log, however
// the prompt run, its sessions, requests and iterations move.
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
		assertProjection(t, db, fixture.issueID, "open", "running")

		// Replaying the same authoritative snapshot is a complete no-op.
		var changed int
		require.NoError(t, db.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`, fixture.runID).Scan(&changed).Error)
		assert.Zero(t, changed)
		assertProjection(t, db, fixture.issueID, "open", "running")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET phase = 'verify', state = 'running', version = 1, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "verifying")

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
		assertProjection(t, db, fixture.issueID, "open", "verifying")

		// The host verified the issue. A failed verification run is visible as
		// transient state; whether it reopens the issue is the host's decision.
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, fixture.issueID).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', version = 3, finished_at = now(), updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "verification_failed")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'cancelled', phase = 'finished', version = 4, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "idle")

		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 5, updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "idle")
	})

	t.Run("requests session health and descendant deletion", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")

		questionID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_turn_requests
				(id, session_id, prompt_run_id, kind, state, request, version)
			VALUES (?, ?, ?, 'question', 'pending', '{}'::jsonb, 0)`,
			questionID, fixture.sessionID, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting")

		require.NoError(t, db.Exec(`
			UPDATE captain_turn_requests
			SET state = 'answered', response = '{"answer":"yes"}'::jsonb,
				version = 1, resolved_at = now()
			WHERE id = ?`, questionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")

		// A tool approval is addressed by prompt run and tool call — Captain's
		// captain_turn_requests_tool_approval_identity check refuses one without.
		approvalID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_turn_requests
				(id, session_id, prompt_run_id, tool_call_id, kind, state, request, version)
			VALUES (?, ?, ?, 'toolu_projection_bash', 'tool_approval', 'pending', '{}'::jsonb, 0)`,
			approvalID, fixture.sessionID, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting")
		require.NoError(t, db.Exec(`DELETE FROM captain_turn_requests WHERE id = ?`, approvalID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")

		childID := uuid.New()
		require.NoError(t, db.Exec(`
			INSERT INTO captain_sessions
				(id, source, provider, host_id, parent_session_id, root_session_id,
				 activity_state, health_state, state_version)
			VALUES (?, 'test', 'test', 'local', ?, ?, 'working', 'stalled', 0)`,
			childID, fixture.sessionID, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled")

		// Removing the descendant that supplied the stalled signal recomputes
		// from its surviving parent instead of leaving the issue stale.
		require.NoError(t, db.Exec(`DELETE FROM captain_sessions WHERE id = ?`, childID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'ask', state_version = 1, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "waiting")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'working', health_state = 'stalled',
				state_version = 2, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET health_state = 'zombie', state_version = 3, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "failed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET state_version = 2, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "failed")

		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET lifecycle_status = 'cancelled', health_state = 'healthy',
				state_version = 4, updated_at = now()
			WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "idle")
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
		assertProjection(t, db, fixture.issueID, "open", "waiting")

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
		assertProjection(t, db, fixture.issueID, "open", "running")

		// Admission state is intentionally ignored by runtime state derivation.
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET activity_state = 'approval', health_state = 'zombie',
			    state_version = state_version + 1, updated_at = now()
			WHERE id = ?`, fixture.admissionSessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")
	})

	t.Run("planning and verify steps derive transient state only", func(t *testing.T) {
		// A planning pass is recognised by the plan-mode spec it ran with, and
		// finishing it changes nothing durable: the host reads the plan it left.
		plan := newProjectionFixture(t, db, workspaceID, "plan", "plan fixture must not verify", `{
			"permissions":{"mode":"plan"}
		}`)
		assertProjection(t, db, plan.issueID, "open", "planning")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, plan.runID).Error)
		assertProjection(t, db, plan.issueID, "open", "idle")

		// A verification pass is recognised by the spec it ran with — a definition
		// of done and no prompt, which is what the verify-only step dispatches —
		// so its failure is a failed verification. A finished verify step used to
		// promote the issue to verified from SQL; the verdict now travels through
		// the host's outcome.
		verify := newProjectionFixture(t, db, workspaceID, "verify", "", verifyOnlySpec)
		assertProjection(t, db, verify.issueID, "open", "verifying")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, verify.runID).Error)
		assertProjection(t, db, verify.issueID, "open", "verification_failed")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', version = 2, updated_at = now()
			WHERE id = ?`, verify.runID).Error)
		assertProjection(t, db, verify.issueID, "open", "idle")

		// Nor does an implementation run without a fixture reopen a verified
		// issue any more (the retired verification_required transition).
		noFixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, noFixture.issueID).Error)
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'succeeded', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, noFixture.runID).Error)
		assertProjection(t, db, noFixture.issueID, "verified", "idle")
	})

	t.Run("rendered spec changes are derived at read time", func(t *testing.T) {
		// Execution state is computed from the current rows on every read, so
		// the spec a run was (re)admitted with is never cached on the issue.
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs
				SET rendered_spec = '{"permissions":{"mode":"plan"}}'::jsonb
				WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "planning")

		require.NoError(t, db.Exec(`
				UPDATE captain_prompt_runs SET rendered_spec = '{}'::jsonb WHERE id = ?`,
			fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")
	})

	t.Run("finished verification failure retains iteration provenance", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "gavel fixture", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")
		require.NoError(t, db.Exec(`UPDATE todo_issues SET status = 'verified' WHERE id = ?`, fixture.issueID).Error)

		// A finished failure without durable verifier evidence is a generic run
		// failure, even though this run supplied verification Markdown.
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET state = 'failed', phase = 'finished', version = 1,
				finished_at = now(), updated_at = now()
			WHERE id = ?`, fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "failed")

		// The latest failed iteration carries durable verification provenance,
		// which the derivation reads after cleanup advanced the phase.
		require.NoError(t, db.Exec(`
			INSERT INTO captain_prompt_run_iterations
				(id, prompt_run_id, iteration, state, verification_result,
				 started_at, finished_at)
			VALUES (?, ?, 0, 'failed', '{"success":false}'::jsonb, now(), now())`,
			uuid.New(), fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "verification_failed")

		// A later successful iteration supersedes the failed attempt. An
		// unrelated post-run failure must remain generic.
		require.NoError(t, db.Exec(`
			INSERT INTO captain_prompt_run_iterations
				(id, prompt_run_id, iteration, state, verification_result,
				 started_at, finished_at)
			VALUES (?, ?, 1, 'succeeded', '{"success":true}'::jsonb, now(), now())`,
			uuid.New(), fixture.runID).Error)
		assertProjection(t, db, fixture.issueID, "verified", "failed")
	})

	t.Run("session root reassignment reprojects the old aggregate", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")

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
		assertProjection(t, db, fixture.issueID, "open", "stalled")

		// Parent/root changes do not increment Captain state_version, so they
		// must be explicit trigger columns and route both the old and new roots.
		require.NoError(t, db.Exec(`
			UPDATE captain_sessions
			SET parent_session_id = ?, root_session_id = ?
			WHERE id = ?`, otherRootID, otherRootID, childID).Error)
		assertProjection(t, db, fixture.issueID, "open", "running")
	})

	t.Run("prompt run reassignment reprojects without prompt version bump", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")
		require.NoError(t, db.Exec(`
				UPDATE captain_sessions
				SET health_state = 'stalled', state_version = 1, updated_at = now()
				WHERE id = ?`, fixture.sessionID).Error)
		assertProjection(t, db, fixture.issueID, "open", "stalled")

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
		assertProjection(t, db, fixture.issueID, "open", "running")
	})

	t.Run("clearing the active pointer reads as idle without a projection call", func(t *testing.T) {
		fixture := newProjectionFixture(t, db, workspaceID, "run", "", `{}`)
		assertProjection(t, db, fixture.issueID, "open", "running")
		require.NoError(t, db.Exec(`
			UPDATE todo_issues SET active_prompt_run_id = NULL WHERE id = ?`, fixture.issueID).Error)
		assertProjection(t, db, fixture.issueID, "open", "idle")

		// The per-issue projection function was a stub every caller ignored; the
		// bundle drops it rather than keeping a contract nobody depends on.
		var stub *string
		require.NoError(t, db.Raw(`SELECT to_regprocedure('public.gavel_project_todo_issue(uuid)')::text`).Scan(&stub).Error)
		assert.Nil(t, stub, "gavel_project_todo_issue must be dropped from the database")
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
