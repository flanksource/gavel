package native_test

import (
	"os"
	"path/filepath"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestExecutionIntegrationRequiresRepository(t *testing.T) {
	_, err := native.NewExecutionIntegration(nil)
	require.ErrorIs(t, err, native.ErrInvalidInput)
}

func TestExecutionIntegrationAtomicLinksAndReplay(t *testing.T) {
	repo, db, dsn := openExecutionRepository(t)
	ctx := t.Context()

	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey:  "github.com/flanksource/gavel-execution-integration",
		RootPath: "/workspace/gavel-execution-integration",
	})
	require.NoError(t, err)
	issue := createIssue(t, repo, workspace.ID, "execution issue")
	otherIssue := createIssue(t, repo, workspace.ID, "other execution issue")

	integration, err := native.NewExecutionIntegration(repo)
	require.NoError(t, err)
	owner, err := native.LocalOwner()
	require.NoError(t, err)

	runID := insertCaptainPromptRun(t, db)
	originalVersion := issue.Version
	issue, err = integration.ActivatePromptRun(ctx, native.PromptRunAttachment{
		IssueID:              issue.ID,
		PromptRunID:          runID,
		StepKind:             native.StepRun,
		Ordinal:              0,
		ExpectedIssueVersion: originalVersion,
		Actor:                "execution-test",
		Owner:                &owner,
	})
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	assert.Equal(t, runID, *issue.ActivePromptRunID)
	assert.Equal(t, originalVersion+1, issue.Version, "activation mutates once; execution state is derived at read time, not projected")
	assert.Equal(t, native.ExecutionRunning, issue.ExecutionState)

	// A lost-response retry carries the original version. The complete link and
	// pointer make it an exact no-op rather than a version conflict.
	replayed, err := integration.ActivatePromptRun(ctx, native.PromptRunAttachment{
		IssueID:              issue.ID,
		PromptRunID:          runID,
		StepKind:             native.StepRun,
		Ordinal:              0,
		ExpectedIssueVersion: originalVersion,
		Actor:                "execution-test",
		Owner:                &owner,
	})
	require.NoError(t, err)
	assert.Equal(t, issue.Version, replayed.Version)
	events, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Len(t, events, int(issue.Version))

	// One Captain prompt run cannot be overloaded across a grouped TODO run.
	_, err = integration.ActivatePromptRun(ctx, native.PromptRunAttachment{
		IssueID:              otherIssue.ID,
		PromptRunID:          runID,
		StepKind:             native.StepRun,
		Ordinal:              0,
		ExpectedIssueVersion: otherIssue.Version,
		Actor:                "execution-test",
		Owner:                &owner,
	})
	require.ErrorIs(t, err, native.ErrLinkConflict)
	otherIssue, err = repo.GetIssue(ctx, otherIssue.ID)
	require.NoError(t, err)
	assert.Nil(t, otherIssue.ActivePromptRunID)
	otherLinks, err := repo.ListPromptRuns(ctx, otherIssue.ID)
	require.NoError(t, err)
	assert.Empty(t, otherLinks)

	// A conflicting ordinal fails before activation and leaves the prior pointer
	// and issue version intact.
	conflictingRunID := insertCaptainPromptRun(t, db)
	_, err = integration.ActivatePromptRun(ctx, native.PromptRunAttachment{
		IssueID:              issue.ID,
		PromptRunID:          conflictingRunID,
		StepKind:             native.StepRun,
		Ordinal:              0,
		ExpectedIssueVersion: issue.Version,
		Actor:                "execution-test",
		Owner:                &owner,
	})
	require.ErrorIs(t, err, native.ErrLinkConflict)
	afterConflict, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, issue.Version, afterConflict.Version)
	require.NotNil(t, afterConflict.ActivePromptRunID)
	assert.Equal(t, runID, *afterConflict.ActivePromptRunID)
	links, err := repo.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, runID, links[0].PromptRunID)

	beforeClearVersion := issue.Version
	issue, err = repo.SetActivePromptRun(ctx, issue.ID, nil, issue.Version, "execution-test")
	require.NoError(t, err)
	assert.Nil(t, issue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionIdle, issue.ExecutionState)
	// Execution state is read-time only: clearing the pointer is the one
	// mutation, and nothing projects a status back onto the issue.
	assert.Equal(t, beforeClearVersion+1, issue.Version, "pointer clear mutates once")

	planID := insertCaptainPlan(t, db, runID)
	planOriginalVersion := issue.Version
	issue, err = integration.SelectPlan(ctx, native.PlanAttachment{
		IssueID:              issue.ID,
		PlanID:               planID,
		Ordinal:              0,
		ExpectedIssueVersion: planOriginalVersion,
		Actor:                "execution-test",
	})
	require.NoError(t, err)
	require.NotNil(t, issue.SelectedPlanID)
	assert.Equal(t, planID, *issue.SelectedPlanID)
	assert.Equal(t, planOriginalVersion+1, issue.Version)

	planReplay, err := integration.SelectPlan(ctx, native.PlanAttachment{
		IssueID:              issue.ID,
		PlanID:               planID,
		Ordinal:              0,
		ExpectedIssueVersion: planOriginalVersion,
		Actor:                "execution-test",
	})
	require.NoError(t, err)
	assert.Equal(t, issue.Version, planReplay.Version)

	conflictingPlanID := insertCaptainPlan(t, db, runID)
	_, err = integration.SelectPlan(ctx, native.PlanAttachment{
		IssueID:              issue.ID,
		PlanID:               conflictingPlanID,
		Ordinal:              0,
		ExpectedIssueVersion: issue.Version,
		Actor:                "execution-test",
	})
	require.ErrorIs(t, err, native.ErrLinkConflict)
	afterPlanConflict, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, issue.Version, afterPlanConflict.Version)
	require.NotNil(t, afterPlanConflict.SelectedPlanID)
	assert.Equal(t, planID, *afterPlanConflict.SelectedPlanID)
	plans, err := repo.ListPlans(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, plans, 1)
	assert.Equal(t, planID, plans[0].PlanID)

	// A second GORM connection to the same database is still a distinct pool;
	// mixing it with the native repository cannot provide one atomic operation.
	otherPool, err := commonsdb.NewGorm(dsn, commonsdb.DefaultGormConfig())
	require.NoError(t, err)
	otherSQL, err := otherPool.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, otherSQL.Close()) })
	otherCaptain, err := captaindb.Use(otherPool)
	require.NoError(t, err)
	_, err = native.NewLaunchCoordinator(otherCaptain, repo)
	require.ErrorIs(t, err, native.ErrDatabasePoolMismatch)
}

func openExecutionRepository(t *testing.T) (*native.Repository, *gorm.DB, string) {
	t.Helper()
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres execution integration tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_native_execution",
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
	repo, err := native.NewRepository(opened.Gorm())
	require.NoError(t, err)
	return repo, opened.Gorm(), dsn
}

func insertCaptainPromptRun(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	sessionID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO captain_sessions (id, source, provider, host_id)
		VALUES (?, 'test', 'test', 'local')`, sessionID,
	).Error)
	runID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin)
		VALUES (?, ?, ?, 'gavel-test')`, runID, sessionID, sessionID,
	).Error)
	return runID
}

func insertCaptainPlan(t *testing.T, db *gorm.DB, runID uuid.UUID) uuid.UUID {
	t.Helper()
	var sessionText string
	require.NoError(t, db.Raw(`
		SELECT session_id FROM captain_prompt_runs WHERE id = ?`, runID,
	).Scan(&sessionText).Error)
	sessionID, err := uuid.Parse(sessionText)
	require.NoError(t, err)
	planID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO captain_plans (id, source_session_id, source_prompt_run_id, title)
		VALUES (?, ?, ?, 'Gavel integration plan')`, planID, sessionID, runID,
	).Error)
	return planID
}
