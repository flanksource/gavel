package native_test

import (
	"testing"

	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// linkPhaseRun attaches a Captain prompt run to an issue under one step kind,
// bypassing ExecutionIntegration so a test can seed several finished phases
// without driving each through activation.
func linkPhaseRun(t *testing.T, db *gorm.DB, issueID uuid.UUID, step native.StepKind, ordinal int) uuid.UUID {
	t.Helper()
	runID := insertCaptainPromptRun(t, db)
	require.NoError(t, db.Exec(`
		INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal)
		VALUES (?, ?, ?, ?)`, issueID, runID, string(step), ordinal,
	).Error)
	return runID
}

// setRunState finishes a seeded run. queued_at is backdated with the rest
// because captain_prompt_runs_time_order requires queued_at <= started_at <=
// finished_at, and queued_at defaults to now().
func setRunState(t *testing.T, db *gorm.DB, runID uuid.UUID, state, phase string) {
	t.Helper()
	require.NoError(t, db.Exec(`
		UPDATE captain_prompt_runs
		SET state = ?, phase = ?,
		    queued_at  = now() - interval '6 minutes',
		    started_at = now() - interval '5 minutes',
		    finished_at = now()
		WHERE id = ?`, state, phase, runID,
	).Error)
}

func phasesByKind(runs []native.IssuePhaseRun, issueID uuid.UUID) map[native.StepKind]native.IssuePhaseRun {
	byKind := map[native.StepKind]native.IssuePhaseRun{}
	for _, run := range runs {
		if run.IssueID == issueID {
			byKind[run.Phase] = run
		}
	}
	return byKind
}

func TestListIssuePhaseRuns(t *testing.T) {
	repo, db, _ := openExecutionRepository(t)
	ctx := t.Context()

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey:  "github.com/flanksource/gavel-phase-runs",
		RootPath: "/workspace/gavel-phase-runs",
	})
	require.NoError(t, err)

	t.Run("returns the latest run per phase", func(t *testing.T) {
		issue := createIssue(t, repo, workspace.ID, "four phases")

		setRunState(t, db, linkPhaseRun(t, db, issue.ID, native.StepPlan, 0), "succeeded", "finished")
		setRunState(t, db, linkPhaseRun(t, db, issue.ID, native.StepTriage, 0), "succeeded", "finished")
		setRunState(t, db, linkPhaseRun(t, db, issue.ID, native.StepVerify, 0), "failed", "finished")
		// Two run-phase attempts: only the later ordinal should surface.
		setRunState(t, db, linkPhaseRun(t, db, issue.ID, native.StepRun, 0), "failed", "finished")
		latestRun := linkPhaseRun(t, db, issue.ID, native.StepRun, 1)
		setRunState(t, db, latestRun, "succeeded", "finished")

		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)

		byKind := phasesByKind(runs, issue.ID)
		assert.ElementsMatch(t,
			[]native.StepKind{native.StepPlan, native.StepRun, native.StepTriage, native.StepVerify},
			maps(byKind),
		)
		assert.Equal(t, "succeeded", byKind[native.StepRun].State, "the later ordinal wins")
		assert.Equal(t, "failed", byKind[native.StepVerify].State)
		// A finished run has a measurable duration; the view derives it from
		// started_at/finished_at rather than storing it.
		require.NotNil(t, byKind[native.StepRun].DurationSeconds)
		assert.Greater(t, *byKind[native.StepRun].DurationSeconds, 0.0)
	})

	// Most verification never produces a standalone verify run — it happens as
	// phase='verify' inside the run. A verify column that only read
	// step_kind='verify' would be empty for almost every todo.
	t.Run("folds a run that reached verification into the verify phase", func(t *testing.T) {
		issue := createIssue(t, repo, workspace.ID, "verify inside run")
		setRunState(t, db, linkPhaseRun(t, db, issue.ID, native.StepRun, 0), "failed", "verify")

		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)

		byKind := phasesByKind(runs, issue.ID)
		require.Contains(t, byKind, native.StepVerify, "run-phase verification must surface as verify")
		assert.Equal(t, "verify", byKind[native.StepVerify].RunPhase)
		assert.Contains(t, byKind, native.StepRun, "and still count as the run phase")
	})

	// The verify phase draws rows from two step kinds, and ordinals are numbered
	// per kind — so a standalone verify at ordinal 0 must still beat an older
	// run at ordinal 3. Ordering by ordinal alone would pick the run.
	t.Run("prefers the most recent verification across step kinds", func(t *testing.T) {
		issue := createIssue(t, repo, workspace.ID, "verify ordering")

		olderRun := linkPhaseRun(t, db, issue.ID, native.StepRun, 3)
		setRunState(t, db, olderRun, "failed", "verify")
		require.NoError(t, db.Exec(`
			UPDATE captain_prompt_runs
			SET queued_at = now() - interval '2 hours',
			    started_at = now() - interval '2 hours',
			    finished_at = now() - interval '110 minutes'
			WHERE id = ?`, olderRun).Error)

		newerVerify := linkPhaseRun(t, db, issue.ID, native.StepVerify, 0)
		setRunState(t, db, newerVerify, "succeeded", "finished")

		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)
		assert.Equal(t, "succeeded", phasesByKind(runs, issue.ID)[native.StepVerify].State,
			"the later standalone verify must win over the older run's verify phase")
	})

	t.Run("marks the issue's active run", func(t *testing.T) {
		issue := createIssue(t, repo, workspace.ID, "active run")
		runID := linkPhaseRun(t, db, issue.ID, native.StepRun, 0)
		require.NoError(t, db.Exec(
			`UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`, runID, issue.ID).Error)

		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)
		assert.True(t, phasesByKind(runs, issue.ID)[native.StepRun].Active)
	})

	t.Run("omits an issue that has never run", func(t *testing.T) {
		issue := createIssue(t, repo, workspace.ID, "never run")
		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)
		assert.Empty(t, phasesByKind(runs, issue.ID))
	})

	// The whole point of the aggregate: a workspace costs one query however many
	// todos it holds. Resolving phases per row is the N+1 that made
	// /api/projects take 46 seconds.
	t.Run("scopes to one workspace", func(t *testing.T) {
		other, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey:  "github.com/flanksource/gavel-phase-runs-other",
			RootPath: "/workspace/gavel-phase-runs-other",
		})
		require.NoError(t, err)
		foreign := createIssue(t, repo, other.ID, "foreign issue")
		setRunState(t, db, linkPhaseRun(t, db, foreign.ID, native.StepRun, 0), "succeeded", "finished")

		runs, err := repo.ListIssuePhaseRuns(ctx, workspace.ID)
		require.NoError(t, err)
		for _, run := range runs {
			assert.NotEqual(t, foreign.ID, run.IssueID)
		}
	})
}

func maps(byKind map[native.StepKind]native.IssuePhaseRun) []native.StepKind {
	kinds := make([]native.StepKind, 0, len(byKind))
	for kind := range byKind {
		kinds = append(kinds, kind)
	}
	return kinds
}
