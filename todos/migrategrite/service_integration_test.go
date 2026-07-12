package migrategrite_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/migrategrite"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	serviceIssueLinked         = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	serviceIssueMissingSession = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	serviceIssueMissingPlan    = "cccccccccccccccccccccccccccccccc"
	serviceIssueUnresolvedPlan = "dddddddddddddddddddddddddddddddd"
)

func TestServiceImportResolvesCaptainPlansWarningsAndReplaysExactly(t *testing.T) {
	service, repository, captain, gormDB := openMigrationService(t)
	ctx := t.Context()

	linkedSession, linkedRun := createCaptainRun(t, captain, "captain-linked")
	_, missingPlanRun := createCaptainRun(t, captain, "captain-missing-plan")
	_, unresolvedPlanRun := createCaptainRun(t, captain, "captain-unresolved-plan")
	require.NotEqual(t, linkedRun.ID, missingPlanRun.ID)
	require.NotEqual(t, linkedRun.ID, unresolvedPlanRun.ID)

	planRoot := t.TempDir()
	planPath := filepath.Join("plans", "import.md")
	planMarkdown := "# Imported plan\n\n- preserve the source history\n"
	require.NoError(t, os.MkdirAll(filepath.Join(planRoot, "plans"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, planPath), []byte(planMarkdown), 0o600))
	unresolvedPlanPath := filepath.Join("plans", "unresolved.md")
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, unresolvedPlanPath), []byte("# Orphaned readable plan\n"), 0o600))
	authoritativePlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: linkedSession.ID, Path: planPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)

	snapshot := serviceSnapshot(t, linkedSession.ProviderSessionID, "captain-missing", "captain-missing-plan", "captain-unresolved-plan", planPath, unresolvedPlanPath)
	options := migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey:     "github.com/flanksource/gavel-migrategrite-service-test",
			RootPath:    planRoot,
			DisplayName: "migrategrite service test",
		},
		PlanRoot: planRoot,
	}

	first, err := service.Import(ctx, snapshot, options)
	require.NoError(t, err)
	assert.Equal(t, 4, first.Counts.IssuesCreated)
	assert.Equal(t, 16, first.Counts.EventsInserted, "source events, checkpoints, and durable warnings share the event stream")
	assert.Equal(t, 5, first.Counts.WarningsInserted)
	assert.Equal(t, 3, first.Counts.PromptRunLinksInserted)
	assert.Equal(t, 1, first.Counts.PlanLinksInserted)
	assert.Equal(t, 1, first.Counts.ProjectionEventsInserted)
	assert.Equal(t, 3, first.Validation.PromptRunLinkCount)
	assert.Equal(t, 1, first.Validation.PlanLinkCount)
	assert.Equal(t, 5, first.Validation.WarningCount)
	assertWarning(t, first.Warnings, serviceIssueMissingSession, "captain_session_unresolved")
	assertWarning(t, first.Warnings, serviceIssueMissingPlan, "plan_file_missing")
	assertWarning(t, first.Warnings, serviceIssueMissingPlan, "captain_plan_unresolved")
	assertWarning(t, first.Warnings, serviceIssueUnresolvedPlan, "captain_plan_unresolved")
	assert.NotEmpty(t, first.Validation.ImportFingerprint)
	assert.NotEmpty(t, first.Validation.CaptainChecksum)

	linkedByFull, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	linkedByShort, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueLinked[:8])
	require.NoError(t, err)
	assert.Equal(t, linkedByFull.ID, linkedByShort.ID)
	require.NotNil(t, linkedByFull.ActivePromptRunID)
	assert.Equal(t, linkedRun.ID, *linkedByFull.ActivePromptRunID)
	assert.Equal(t, native.ExecutionRunning, linkedByFull.ExecutionState)

	promptLinks, err := repository.ListPromptRuns(ctx, linkedByFull.ID)
	require.NoError(t, err)
	require.Len(t, promptLinks, 1)
	assert.Equal(t, linkedRun.ID, promptLinks[0].PromptRunID)
	assert.Equal(t, native.StepRun, promptLinks[0].StepKind)

	require.NotNil(t, linkedByFull.SelectedPlanID)
	planLinks, err := repository.ListPlans(ctx, linkedByFull.ID)
	require.NoError(t, err)
	require.Len(t, planLinks, 1)
	assert.Equal(t, *linkedByFull.SelectedPlanID, planLinks[0].PlanID)
	plan, err := captain.GetPlan(ctx, *linkedByFull.SelectedPlanID)
	require.NoError(t, err)
	assert.Equal(t, authoritativePlan.ID, plan.ID)
	assert.Equal(t, linkedSession.ID, plan.SourceSessionID)
	assert.Equal(t, planPath, plan.Path)
	revisions, err := captain.ListPlanRevisions(ctx, plan.ID)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
	assert.Equal(t, "# Imported plan\n\n- preserve the source history", revisions[0].PlanMarkdown)
	assert.Equal(t, native.DefaultImportSource, revisions[0].CreatedBy)

	missingSessionIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueMissingSession)
	require.NoError(t, err)
	assert.Nil(t, missingSessionIssue.ActivePromptRunID)
	assertIssueWarningEvent(t, repository, missingSessionIssue.ID, "captain_session_unresolved")
	missingPlanIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueMissingPlan)
	require.NoError(t, err)
	assert.Nil(t, missingPlanIssue.SelectedPlanID)
	assertIssueWarningEvent(t, repository, missingPlanIssue.ID, "plan_file_missing")
	assertIssueWarningEvent(t, repository, missingPlanIssue.ID, "captain_plan_unresolved")
	unresolvedPlanIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueUnresolvedPlan)
	require.NoError(t, err)
	assert.Nil(t, unresolvedPlanIssue.SelectedPlanID)
	assertIssueWarningEvent(t, repository, unresolvedPlanIssue.ID, "captain_plan_unresolved")

	beforeReplay := readServiceCounts(t, gormDB)
	second, err := service.Import(ctx, snapshot, options)
	require.NoError(t, err)
	assert.Zero(t, second.Counts.IssuesCreated)
	assert.Zero(t, second.Counts.IssuesUpdated)
	assert.Zero(t, second.Counts.EventsInserted)
	assert.Zero(t, second.Counts.WarningsInserted)
	assert.Zero(t, second.Counts.ProjectionEventsInserted)
	assert.Zero(t, second.Counts.PromptRunLinksInserted)
	assert.Zero(t, second.Counts.PlanLinksInserted)
	assert.Equal(t, 16, second.Counts.EventsReplayed, "source events, checkpoints, and durable warnings must all replay")
	assert.Equal(t, 5, second.Counts.WarningsReplayed)
	assert.Equal(t, 3, second.Counts.PromptRunLinksReplayed)
	assert.Equal(t, 1, second.Counts.PlanLinksReplayed)
	assert.Equal(t, first.Validation.TargetChecksum, second.Validation.TargetChecksum)
	assert.Equal(t, first.Validation.SourceHash, second.Validation.SourceHash)
	assert.Equal(t, first.Validation.ImportFingerprint, second.Validation.ImportFingerprint)
	assert.Equal(t, first.Validation.CaptainChecksum, second.Validation.CaptainChecksum)
	assert.Equal(t, beforeReplay, readServiceCounts(t, gormDB))

	revisionsAfterReplay, err := captain.ListPlanRevisions(ctx, plan.ID)
	require.NoError(t, err)
	assert.Equal(t, revisions, revisionsAfterReplay)
	assert.Empty(t, second.Rollback.CreatedCaptainPlanIDs)
	assert.Empty(t, second.Rollback.AppendedCaptainRevisionIDs)
}

func TestServiceBeforeCommitFailureRollsBackNativeAndCaptainMutations(t *testing.T) {
	service, _, captain, gormDB := openMigrationService(t)
	ctx := t.Context()
	session, _ := createCaptainRun(t, captain, "captain-before-commit")
	planRoot := t.TempDir()
	planPath := filepath.Join("plans", "atomic.md")
	require.NoError(t, os.MkdirAll(filepath.Join(planRoot, "plans"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, planPath), []byte("# Atomic plan\n"), 0o600))
	plan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: session.ID,
		Path:            planPath,
		Variant:         "legacy-authoritative",
	})
	require.NoError(t, err)

	snapshot := singleServiceSnapshot(t, serviceIssueLinked, "Atomic import", "status:pending", session.ProviderSessionID, planPath)
	sentinel := errors.New("artifact fsync failed")
	var callbackReport *migrategrite.ImportReport
	_, err = service.Import(ctx, snapshot, migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey:     "github.com/flanksource/gavel-migrategrite-before-commit",
			RootPath:    planRoot,
			DisplayName: "before commit rollback",
		},
		PlanRoot: planRoot,
		BeforeCommit: func(report *migrategrite.ImportReport) error {
			callbackReport = report
			return sentinel
		},
	})
	require.ErrorIs(t, err, sentinel)
	require.NotNil(t, callbackReport)
	assert.NotEmpty(t, callbackReport.Validation.TargetChecksum)
	assert.NotEmpty(t, callbackReport.Validation.CaptainChecksum)
	assert.Len(t, callbackReport.Rollback.AppendedCaptainRevisionIDs, 1)

	assert.Equal(t, serviceCounts{}, readServiceCounts(t, gormDB))
	var issueCount, workspaceCount int64
	require.NoError(t, gormDB.Table("todo_issues").Count(&issueCount).Error)
	require.NoError(t, gormDB.Table("todo_workspaces").Count(&workspaceCount).Error)
	assert.Zero(t, issueCount)
	assert.Zero(t, workspaceCount)
	revisions, err := captain.ListPlanRevisions(ctx, plan.ID)
	require.NoError(t, err)
	assert.Empty(t, revisions, "Captain revision append must roll back with the native import")
}

func TestServiceDeferredActivePromptRunPreservesDriftGuardUntilFinalProjection(t *testing.T) {
	service, repository, captain, _ := openMigrationService(t)
	ctx := t.Context()
	session, run := createCaptainRun(t, captain, "captain-deferred-active")
	events := []griteexport.Event{
		serviceEvent(t, "deferred-created", serviceIssueLinked, 1_000, "IssueCreated", map[string]any{
			"title": "Deferred active run", "body": "body",
		}),
	}
	snapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 2_000, EventCount: len(events)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Deferred active run", State: "open",
			Labels:    []string{"status:in_progress", "mode:run", "session:" + session.ProviderSessionID},
			CreatedTS: 1_000, UpdatedTS: 2_000,
		}},
		Events: events,
	}
	options := migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey: "github.com/flanksource/gavel-migrategrite-deferred-active", RootPath: t.TempDir(), DisplayName: "deferred active",
		},
		DeferActivePromptRuns: true,
	}
	initial, err := service.Import(ctx, snapshot, options)
	require.NoError(t, err)
	assert.Equal(t, 1, initial.Counts.PromptRunLinksInserted)
	assert.Zero(t, initial.Counts.ProjectionEventsInserted)
	issue, err := repository.GetIssueByRef(ctx, initial.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	assert.Nil(t, issue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionIdle, issue.ExecutionState)
	links, err := repository.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, run.ID, links[0].PromptRunID)
	assert.Equal(t, native.StepRun, links[0].StepKind)

	verifyPhase := captaindb.PromptRunPhaseVerify
	run, err = captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: run.ID, ExpectedVersion: run.Version, Phase: &verifyPhase,
	})
	require.NoError(t, err)
	assert.Equal(t, captaindb.PromptRunStateRunning, run.State)
	afterCaptainDrift, err := repository.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Nil(t, afterCaptainDrift.ActivePromptRunID, "deferred links must not project mutable Captain state before finalization")
	assert.Equal(t, native.ExecutionIdle, afterCaptainDrift.ExecutionState)
	assert.Equal(t, issue.Version, afterCaptainDrift.Version)

	options.DeferActivePromptRuns = false
	options.ExpectedTargetChecksum = initial.Validation.TargetChecksum
	final, err := service.Import(ctx, snapshot, options)
	require.NoError(t, err)
	assert.Equal(t, 1, final.Counts.PromptRunLinksReplayed)
	assert.Equal(t, 1, final.Counts.ProjectionEventsInserted)
	assert.NotEqual(t, initial.Validation.TargetChecksum, final.Validation.TargetChecksum)
	assert.NotEqual(t, initial.Validation.CaptainChecksum, final.Validation.CaptainChecksum)
	issue, err = repository.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	assert.Equal(t, run.ID, *issue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionVerifying, issue.ExecutionState)

	options.ExpectedTargetChecksum = final.Validation.TargetChecksum
	replay, err := service.Import(ctx, snapshot, options)
	require.NoError(t, err)
	assert.Zero(t, replay.Counts.EventsInserted)
	assert.Zero(t, replay.Counts.ProjectionEventsInserted)
	assert.Equal(t, final.Validation.TargetChecksum, replay.Validation.TargetChecksum)
}

func TestServiceTwoPassStaleModeWaitsForAuthoritativeRunStart(t *testing.T) {
	service, repository, captain, _ := openMigrationService(t)
	ctx := t.Context()
	session, run := createCaptainRun(t, captain, "captain-stale-mode-two-pass")
	initialEvents := []griteexport.Event{
		serviceEvent(t, "stale-created", serviceIssueLinked, 100, "IssueCreated", map[string]any{
			"title": "Stale mode two-pass", "body": "body",
		}),
		serviceEvent(t, "stale-plan-mode", serviceIssueLinked, 200, "LabelAdded", map[string]any{"label": "mode:plan"}),
		serviceEvent(t, "stale-new-session", serviceIssueLinked, 300, "LabelAdded", map[string]any{"label": "session:" + session.ProviderSessionID}),
	}
	initialSnapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 350, EventCount: len(initialEvents)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Stale mode two-pass", State: "open",
			Labels:    []string{"status:in_progress", "mode:plan", "session:" + session.ProviderSessionID},
			CreatedTS: 100, UpdatedTS: 300,
		}},
		Events: initialEvents,
	}
	options := migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey: "github.com/flanksource/gavel-migrategrite-stale-mode-two-pass", RootPath: t.TempDir(), DisplayName: "stale mode two-pass",
		},
		DeferActivePromptRuns: true,
	}
	initial, err := service.Import(ctx, initialSnapshot, options)
	require.NoError(t, err)
	assertWarning(t, initial.Warnings, serviceIssueLinked, "captain_step_unknown")
	assertWarning(t, initial.Warnings, serviceIssueLinked, "captain_live_run_missing")
	assert.Zero(t, initial.Counts.PromptRunLinksInserted)
	assert.Zero(t, initial.Counts.ProjectionEventsInserted)
	assert.Empty(t, initial.Rollback.Native.InsertedPromptRunLinks)
	issue, err := repository.GetIssueByRef(ctx, initial.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	assert.Nil(t, issue.ActivePromptRunID)
	initialLinks, err := repository.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	assert.Empty(t, initialLinks)

	finalEvents := append([]griteexport.Event(nil), initialEvents...)
	finalEvents = append(finalEvents, serviceEvent(t, "authoritative-run-start", serviceIssueLinked, 400, "CommentAdded", map[string]any{
		"body": "**Todo run started**\n\n- **Session ID:** `" + session.ProviderSessionID + "`\n- **Mode:** `run`\n- **Resolved Model:** `default`\n- **Effort:** `default`",
	}))
	finalSnapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 450, EventCount: len(finalEvents)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Stale mode two-pass", State: "open",
			Labels:    []string{"status:in_progress", "mode:plan", "session:" + session.ProviderSessionID},
			CreatedTS: 100, UpdatedTS: 400, CommentCount: 1,
		}},
		Events: finalEvents,
	}
	options.DeferActivePromptRuns = false
	options.ExpectedTargetChecksum = initial.Validation.TargetChecksum
	final, err := service.Import(ctx, finalSnapshot, options)
	require.NoError(t, err)
	assertNoWarning(t, final.Warnings, serviceIssueLinked, "captain_step_unknown")
	assertNoWarning(t, final.Warnings, serviceIssueLinked, "captain_live_run_missing")
	assert.Equal(t, 1, final.Counts.PromptRunLinksInserted)
	assert.Equal(t, 1, final.Counts.ProjectionEventsInserted)
	assert.Len(t, final.Rollback.Native.InsertedPromptRunLinks, 1)
	assert.Equal(t, native.StepRun, final.Rollback.Native.InsertedPromptRunLinks[0].StepKind)
	issue, err = repository.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	assert.Equal(t, run.ID, *issue.ActivePromptRunID)
	links, err := repository.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, run.ID, links[0].PromptRunID)
	assert.Equal(t, native.StepRun, links[0].StepKind)

	combined := migrategrite.MergeRollback(initial.Rollback, final.Rollback)
	assert.Len(t, combined.Native.InsertedPromptRunLinks, 1, "only the authoritative final link belongs in the combined inverse")
	assert.Equal(t, native.StepRun, combined.Native.InsertedPromptRunLinks[0].StepKind)
}

func TestServicePlanClearCombinedWarningsAndEmptyRevisionlessPlan(t *testing.T) {
	service, repository, captain, _ := openMigrationService(t)
	ctx := t.Context()
	clearSession, _ := createCaptainRun(t, captain, "captain-plan-clear")
	emptySession, _ := createCaptainRun(t, captain, "captain-plan-empty")
	pathlessSession, _ := createCaptainRun(t, captain, "captain-plan-pathless")
	planRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(planRoot, "plans"), 0o700))
	clearPath := filepath.Join("plans", "clear.md")
	emptyPath := filepath.Join("plans", "empty.md")
	successorPath := filepath.Join("plans", "empty-successor.md")
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, clearPath), []byte("# Retained after clear\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, emptyPath), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, successorPath), []byte("# Resolved successor\n"), 0o600))
	clearPlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: clearSession.ID, Path: clearPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)
	emptyPlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: emptySession.ID, Path: emptyPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)
	successorPlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: emptySession.ID, Path: successorPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)
	pathlessPlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: pathlessSession.ID, Variant: "inline-authoritative",
	})
	require.NoError(t, err)
	_, created, err := captain.AppendPlanRevisionWithResult(ctx, captaindb.AppendPlanRevisionInput{
		PlanID: pathlessPlan.ID, PlanMarkdown: "# Existing inline plan", CreatedBy: "migration-test",
	})
	require.NoError(t, err)
	require.True(t, created)

	initialEvents := []griteexport.Event{
		serviceEvent(t, "edge-clear-created", serviceIssueLinked, 1_000, "IssueCreated", map[string]any{"title": "Clear", "body": "body"}),
		serviceEvent(t, "edge-clear-plan", serviceIssueLinked, 2_000, "CommentAdded", map[string]any{
			"body": "<!-- gavel:state {\"planPath\":\"" + clearPath + "\"} -->",
		}),
		serviceEvent(t, "edge-missing-created", serviceIssueMissingSession, 1_000, "IssueCreated", map[string]any{"title": "Missing", "body": "body"}),
		serviceEvent(t, "edge-missing-plan", serviceIssueMissingSession, 2_000, "CommentAdded", map[string]any{
			"body": "<!-- gavel:state {\"planPath\":\"plans/missing-both.md\"} -->",
		}),
		serviceEvent(t, "edge-empty-created", serviceIssueMissingPlan, 1_000, "IssueCreated", map[string]any{"title": "Empty", "body": "body"}),
		serviceEvent(t, "edge-empty-plan", serviceIssueMissingPlan, 2_000, "CommentAdded", map[string]any{
			"body": "<!-- gavel:state {\"planPath\":\"" + emptyPath + "\"} -->",
		}),
		serviceEvent(t, "edge-pathless-created", serviceIssueUnresolvedPlan, 1_000, "IssueCreated", map[string]any{"title": "Pathless", "body": "body"}),
		serviceEvent(t, "edge-pathless-status", serviceIssueUnresolvedPlan, 1_500, "LabelAdded", map[string]any{"label": "plan:new"}),
		serviceEvent(t, "edge-pathless-plan", serviceIssueUnresolvedPlan, 2_000, "CommentAdded", map[string]any{
			"body": "<!-- gavel:state {} -->",
		}),
	}
	issues := []griteexport.Issue{
		{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Clear", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "session:" + clearSession.ProviderSessionID},
			CreatedTS: 1_000, UpdatedTS: 2_000, CommentCount: 1,
		},
		{
			IssueID: griteexport.ID(serviceIssueMissingSession), Title: "Missing", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "session:captain-does-not-exist"},
			CreatedTS: 1_000, UpdatedTS: 2_000, CommentCount: 1,
		},
		{
			IssueID: griteexport.ID(serviceIssueMissingPlan), Title: "Empty", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "session:" + emptySession.ProviderSessionID},
			CreatedTS: 1_000, UpdatedTS: 2_000, CommentCount: 1,
		},
		{
			IssueID: griteexport.ID(serviceIssueUnresolvedPlan), Title: "Pathless", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "plan:new", "session:" + pathlessSession.ProviderSessionID},
			CreatedTS: 1_000, UpdatedTS: 2_000, CommentCount: 1,
		},
	}
	initial := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 2_000, EventCount: len(initialEvents)},
		Issues: issues, Events: initialEvents,
	}
	options := migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey: "github.com/flanksource/gavel-migrategrite-plan-edges", RootPath: planRoot, DisplayName: "plan edges",
		},
		PlanRoot: planRoot,
	}
	first, err := service.Import(ctx, initial, options)
	require.NoError(t, err)
	assertWarning(t, first.Warnings, serviceIssueMissingSession, "captain_session_unresolved")
	assertWarning(t, first.Warnings, serviceIssueMissingSession, "plan_file_missing")
	assertWarning(t, first.Warnings, serviceIssueMissingPlan, "plan_content_empty")

	clearIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	require.NotNil(t, clearIssue.SelectedPlanID)
	assert.Equal(t, clearPlan.ID, *clearIssue.SelectedPlanID)
	emptyIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueMissingPlan)
	require.NoError(t, err)
	assert.Nil(t, emptyIssue.SelectedPlanID)
	emptyLinks, err := repository.ListPlans(ctx, emptyIssue.ID)
	require.NoError(t, err)
	assert.Empty(t, emptyLinks, "an empty file must not attach a revisionless authoritative plan")
	revisions, err := captain.ListPlanRevisions(ctx, emptyPlan.ID)
	require.NoError(t, err)
	assert.Empty(t, revisions)
	pathlessIssue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueUnresolvedPlan)
	require.NoError(t, err)
	require.NotNil(t, pathlessIssue.SelectedPlanID)
	assert.Equal(t, pathlessPlan.ID, *pathlessIssue.SelectedPlanID)
	pathlessLinks, err := repository.ListPlans(ctx, pathlessIssue.ID)
	require.NoError(t, err)
	require.Len(t, pathlessLinks, 1)
	assert.Equal(t, pathlessPlan.ID, pathlessLinks[0].PlanID)
	assertNoWarning(t, first.Warnings, serviceIssueUnresolvedPlan, "plan_file_missing")

	finalEvents := append([]griteexport.Event(nil), initialEvents...)
	finalEvents = append(finalEvents, serviceEvent(t, "edge-clear-selection", serviceIssueLinked, 2_000, "CommentAdded", map[string]any{
		"body": "<!-- gavel:state {} -->",
	}))
	finalEvents = append(finalEvents, serviceEvent(t, "edge-empty-successor", serviceIssueMissingPlan, 2_000, "CommentAdded", map[string]any{
		"body": "<!-- gavel:state {\"planPath\":\"" + successorPath + "\"} -->",
	}))
	finalIssues := append([]griteexport.Issue(nil), issues...)
	finalIssues[0].CommentCount = 2
	finalIssues[2].CommentCount = 2
	final := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 3_000, EventCount: len(finalEvents)},
		Issues: finalIssues, Events: finalEvents,
	}
	options.ExpectedTargetChecksum = first.Validation.TargetChecksum
	second, err := service.Import(ctx, final, options)
	require.NoError(t, err)
	clearIssue, err = repository.GetIssueByRef(ctx, second.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	assert.Nil(t, clearIssue.SelectedPlanID)
	clearLinks, err := repository.ListPlans(ctx, clearIssue.ID)
	require.NoError(t, err)
	require.Len(t, clearLinks, 1)
	assert.Equal(t, clearPlan.ID, clearLinks[0].PlanID)
	assertNoWarning(t, second.Warnings, serviceIssueLinked, "plan_file_missing")
	emptyIssue, err = repository.GetIssueByRef(ctx, second.WorkspaceID, serviceIssueMissingPlan)
	require.NoError(t, err)
	require.NotNil(t, emptyIssue.SelectedPlanID)
	assert.Equal(t, successorPlan.ID, *emptyIssue.SelectedPlanID)
	emptyLinks, err = repository.ListPlans(ctx, emptyIssue.ID)
	require.NoError(t, err)
	require.Len(t, emptyLinks, 1)
	assert.Equal(t, successorPlan.ID, emptyLinks[0].PlanID)
	assert.Equal(t, 1, emptyLinks[0].Ordinal, "the unresolved first historical plan reserves ordinal zero")

	require.NoError(t, os.WriteFile(filepath.Join(planRoot, emptyPath), []byte("# Repaired predecessor\n"), 0o600))
	options.ExpectedTargetChecksum = second.Validation.TargetChecksum
	third, err := service.Import(ctx, final, options)
	require.NoError(t, err)
	assert.Equal(t, 1, third.Counts.PlanLinksInserted)
	emptyLinks, err = repository.ListPlans(ctx, emptyIssue.ID)
	require.NoError(t, err)
	require.Len(t, emptyLinks, 2)
	assert.Equal(t, emptyPlan.ID, emptyLinks[0].PlanID)
	assert.Equal(t, 0, emptyLinks[0].Ordinal)
	assert.Equal(t, successorPlan.ID, emptyLinks[1].PlanID)
	assert.Equal(t, 1, emptyLinks[1].Ordinal, "repairing an older plan must not renumber its resolved successor")
	emptyIssue, err = repository.GetIssueByRef(ctx, third.WorkspaceID, serviceIssueMissingPlan)
	require.NoError(t, err)
	require.NotNil(t, emptyIssue.SelectedPlanID)
	assert.Equal(t, successorPlan.ID, *emptyIssue.SelectedPlanID)

	options.ExpectedTargetChecksum = third.Validation.TargetChecksum
	fourth, err := service.Import(ctx, final, options)
	require.NoError(t, err)
	assert.Zero(t, fourth.Counts.EventsInserted)
	assert.Zero(t, fourth.Counts.PlanLinksInserted)
	assert.Equal(t, third.Validation.TargetChecksum, fourth.Validation.TargetChecksum)
	assert.Empty(t, fourth.Rollback.AppendedCaptainRevisionIDs)
}

func TestServicePlanUnchangedRetainsAuthoritativePriorSession(t *testing.T) {
	service, repository, captain, _ := openMigrationService(t)
	ctx := t.Context()
	originalSession, _ := createCaptainRun(t, captain, "captain-plan-original")
	laterSession, _ := createCaptainRun(t, captain, "captain-plan-unchanged")
	planRoot := t.TempDir()
	planPath := filepath.Join("plans", "retained.md")
	require.NoError(t, os.MkdirAll(filepath.Join(planRoot, "plans"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, planPath), []byte("# Retained plan\n\nThe existing plan still stands.\n"), 0o600))
	authoritativePlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: originalSession.ID, Path: planPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)

	events := []griteexport.Event{
		serviceEvent(t, "unchanged-created", serviceIssueLinked, 100, "IssueCreated", map[string]any{"title": "Retained ownership", "body": "body"}),
		serviceEvent(t, "unchanged-plan-mode", serviceIssueLinked, 200, "LabelAdded", map[string]any{"label": "mode:plan"}),
		serviceEvent(t, "unchanged-original-session", serviceIssueLinked, 300, "LabelAdded", map[string]any{"label": "session:" + originalSession.ProviderSessionID}),
		serviceEvent(t, "unchanged-original-status", serviceIssueLinked, 350, "LabelAdded", map[string]any{"label": "plan:new"}),
		serviceEvent(t, "unchanged-original-marker", serviceIssueLinked, 400, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n**Plan:** `" + planPath + "`\n\n<!-- gavel:state {\"planPath\":\"" + planPath + "\"} -->",
		}),
		serviceEvent(t, "unchanged-original-session-removed", serviceIssueLinked, 500, "LabelRemoved", map[string]any{"label": "session:" + originalSession.ProviderSessionID}),
		serviceEvent(t, "unchanged-later-session", serviceIssueLinked, 600, "LabelAdded", map[string]any{"label": "session:" + laterSession.ProviderSessionID}),
		serviceEvent(t, "unchanged-original-status-removed", serviceIssueLinked, 700, "LabelRemoved", map[string]any{"label": "plan:new"}),
		serviceEvent(t, "unchanged-later-status", serviceIssueLinked, 800, "LabelAdded", map[string]any{"label": "plan:unchanged"}),
		serviceEvent(t, "unchanged-later-marker", serviceIssueLinked, 900, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n**Summary:** Existing plan still stands.\n\n**Plan:** `" + planPath + "`\n\n<!-- gavel:state {\"planPath\":\"" + planPath + "\",\"summary\":\"Existing plan still stands.\"} -->",
		}),
	}
	snapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 1_000, EventCount: len(events)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Retained ownership", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "plan:unchanged", "session:" + laterSession.ProviderSessionID},
			CreatedTS: 100, UpdatedTS: 900, CommentCount: 2,
		}},
		Events: events,
	}
	report, err := service.Import(ctx, snapshot, migrategrite.ImportOptions{
		Workspace: native.ImportWorkspace{
			RepoKey: "github.com/flanksource/gavel-migrategrite-plan-unchanged", RootPath: planRoot, DisplayName: "plan unchanged ownership",
		},
		PlanRoot: planRoot,
	})
	require.NoError(t, err)
	assertNoWarning(t, report.Warnings, serviceIssueLinked, "captain_plan_unresolved")
	assertNoWarning(t, report.Warnings, serviceIssueLinked, "plan_file_missing")
	assert.Equal(t, 1, report.Counts.PlanLinksInserted)
	assert.Equal(t, 1, report.Validation.PlanLinkCount)

	issue, err := repository.GetIssueByRef(ctx, report.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	require.NotNil(t, issue.SelectedPlanID)
	assert.Equal(t, authoritativePlan.ID, *issue.SelectedPlanID)
	links, err := repository.ListPlans(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, authoritativePlan.ID, links[0].PlanID)
	revisions, err := captain.ListPlanRevisions(ctx, authoritativePlan.ID)
	require.NoError(t, err)
	require.Len(t, revisions, 1)
	assert.Equal(t, "# Retained plan\n\nThe existing plan still stands.", revisions[0].PlanMarkdown)
}

func TestServiceRelationshipRemoveReaddRetainsStableLifecycleTimestamp(t *testing.T) {
	service, repository, _, _ := openMigrationService(t)
	ctx := t.Context()
	initialEvents := []griteexport.Event{
		serviceEvent(t, "relationship-readd-a-created", serviceIssueLinked, 100, "IssueCreated", map[string]any{"title": "Dependent", "body": "body"}),
		serviceEvent(t, "relationship-readd-b-created", serviceIssueMissingSession, 110, "IssueCreated", map[string]any{"title": "Dependency", "body": "body"}),
		serviceEvent(t, "relationship-readd-added", serviceIssueLinked, 200, "DependencyAdded", map[string]any{
			"dep_type": "depends_on", "target": serviceIssueMissingSession,
		}),
	}
	issues := []griteexport.Issue{
		{
			IssueID: griteexport.ID(serviceIssueLinked), Title: "Dependent", State: "open", Labels: []string{"status:pending"},
			CreatedTS: 100, UpdatedTS: 200,
		},
		{
			IssueID: griteexport.ID(serviceIssueMissingSession), Title: "Dependency", State: "open", Labels: []string{"status:pending"},
			CreatedTS: 110, UpdatedTS: 200,
		},
	}
	initial := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 200, EventCount: len(initialEvents)},
		Issues: issues, Events: initialEvents,
	}
	options := migrategrite.ImportOptions{Workspace: native.ImportWorkspace{
		RepoKey: "github.com/flanksource/gavel-migrategrite-relationship-readd", RootPath: t.TempDir(), DisplayName: "relationship re-add",
	}}
	first, err := service.Import(ctx, initial, options)
	require.NoError(t, err)
	assert.Equal(t, 1, first.Counts.RelationshipsInserted)
	issue, err := repository.GetIssueByRef(ctx, first.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	relationships, err := repository.ListRelationships(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, relationships, 1)
	assert.Equal(t, int64(200), relationships[0].CreatedAt.UnixMilli())

	finalEvents := append([]griteexport.Event(nil), initialEvents...)
	finalEvents = append(finalEvents,
		serviceEvent(t, "relationship-readd-removed", serviceIssueLinked, 300, "DependencyRemoved", map[string]any{
			"dep_type": "depends_on", "target": serviceIssueMissingSession,
		}),
		serviceEvent(t, "relationship-readd-added-again", serviceIssueLinked, 400, "DependencyAdded", map[string]any{
			"dep_type": "depends_on", "target": serviceIssueMissingSession,
		}),
	)
	finalIssues := append([]griteexport.Issue(nil), issues...)
	finalIssues[0].UpdatedTS = 400
	final := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 400, EventCount: len(finalEvents)},
		Issues: finalIssues, Events: finalEvents,
	}
	options.ExpectedTargetChecksum = first.Validation.TargetChecksum
	second, err := service.Import(ctx, final, options)
	require.NoError(t, err)
	assert.Zero(t, second.Counts.RelationshipsInserted)
	assert.Zero(t, second.Counts.RelationshipsDeleted)
	assert.Equal(t, 1, second.Counts.RelationshipsReplayed)
	relationships, err = repository.ListRelationships(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, relationships, 1)
	assert.Equal(t, int64(200), relationships[0].CreatedAt.UnixMilli(), "remove/re-add keeps the first durable add timestamp")

	options.ExpectedTargetChecksum = second.Validation.TargetChecksum
	third, err := service.Import(ctx, final, options)
	require.NoError(t, err)
	assert.Zero(t, third.Counts.EventsInserted)
	assert.Equal(t, 1, third.Counts.RelationshipsReplayed)
	assert.Equal(t, second.Validation.TargetChecksum, third.Validation.TargetChecksum)
}

func TestServiceSelectsNewestQualifyingRunBeforeNewerTerminalRun(t *testing.T) {
	service, repository, captain, _ := openMigrationService(t)
	ctx := t.Context()
	runStart := func(identity, mode string) map[string]any {
		return map[string]any{
			"body": "**Todo run started**\n\n- **Session ID:** `" + identity + "`\n- **Mode:** `" + mode + "`\n- **Resolved Model:** `default`\n- **Effort:** `default`",
		}
	}
	runningSession, olderRunning, newerRunningTerminal := createMaskedQualifyingRunPair(
		t, captain, "captain-running-mask", captaindb.PromptRunStateRunning,
	)
	waitingSession, olderWaiting, newerWaitingTerminal := createMaskedQualifyingRunPair(
		t, captain, "captain-waiting-mask", captaindb.PromptRunStateWaiting,
	)
	terminalMaskSession, olderTerminalMaskedLive, newerTerminalMask := createMaskedQualifyingRunPair(
		t, captain, "captain-terminal-mask-only", captaindb.PromptRunStateRunning,
	)
	planSession, planRun := createCaptainRun(t, captain, "captain-newer-plan-stage")
	require.NoError(t, captain.Gorm().WithContext(ctx).Exec(
		`UPDATE captain_prompt_runs SET created_at = ? WHERE id = ?`,
		newerRunningTerminal.CreatedAt.Add(time.Minute), planRun.ID,
	).Error)
	planRun, err := captain.GetPromptRun(ctx, planRun.ID)
	require.NoError(t, err)
	require.True(t, olderRunning.CreatedAt.Before(newerRunningTerminal.CreatedAt))
	require.True(t, olderWaiting.CreatedAt.Before(newerWaitingTerminal.CreatedAt))
	require.True(t, olderTerminalMaskedLive.CreatedAt.Before(newerTerminalMask.CreatedAt))
	require.True(t, olderRunning.CreatedAt.Before(planRun.CreatedAt))

	events := []griteexport.Event{
		serviceEvent(t, "event-running-created", serviceIssueLinked, 1_000, "IssueCreated", map[string]any{
			"title": "Running issue", "body": "body",
		}),
		serviceEvent(t, "event-running-mode", serviceIssueLinked, 1_100, "LabelAdded", map[string]any{
			"label": "mode:run",
		}),
		serviceEvent(t, "event-running-session", serviceIssueLinked, 1_200, "LabelAdded", map[string]any{
			"label": "session:" + runningSession.ProviderSessionID,
		}),
		serviceEvent(t, "event-running-start", serviceIssueLinked, 1_250, "CommentAdded", runStart(runningSession.ProviderSessionID, "run")),
		serviceEvent(t, "event-running-mode-removed", serviceIssueLinked, 1_300, "LabelRemoved", map[string]any{
			"label": "mode:run",
		}),
		serviceEvent(t, "event-plan-mode", serviceIssueLinked, 1_400, "LabelAdded", map[string]any{
			"label": "mode:plan",
		}),
		serviceEvent(t, "event-plan-session", serviceIssueLinked, 1_500, "LabelAdded", map[string]any{
			"label": "session:" + planSession.ProviderSessionID,
		}),
		serviceEvent(t, "event-plan-start", serviceIssueLinked, 1_550, "CommentAdded", runStart(planSession.ProviderSessionID, "plan")),
		serviceEvent(t, "event-waiting-created", serviceIssueMissingSession, 2_000, "IssueCreated", map[string]any{
			"title": "Waiting issue", "body": "body",
		}),
		serviceEvent(t, "event-waiting-session", serviceIssueMissingSession, 2_050, "LabelAdded", map[string]any{
			"label": "session:" + waitingSession.ProviderSessionID,
		}),
		serviceEvent(t, "event-waiting-start", serviceIssueMissingSession, 2_075, "CommentAdded", runStart(waitingSession.ProviderSessionID, "run")),
		serviceEvent(t, "event-terminal-mask-created", serviceIssueMissingPlan, 2_100, "IssueCreated", map[string]any{
			"title": "Terminal mask issue", "body": "body",
		}),
		serviceEvent(t, "event-terminal-mask-mode", serviceIssueMissingPlan, 2_200, "LabelAdded", map[string]any{
			"label": "mode:run",
		}),
		serviceEvent(t, "event-terminal-mask-session", serviceIssueMissingPlan, 2_300, "LabelAdded", map[string]any{
			"label": "session:" + terminalMaskSession.ProviderSessionID,
		}),
		serviceEvent(t, "event-terminal-mask-start", serviceIssueMissingPlan, 2_350, "CommentAdded", runStart(terminalMaskSession.ProviderSessionID, "run")),
	}
	snapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 4_000, EventCount: len(events)},
		Issues: []griteexport.Issue{
			{
				IssueID: griteexport.ID(serviceIssueLinked), Title: "Running issue", State: "open",
				Labels:    []string{"status:in_progress", "mode:plan", "session:" + planSession.ProviderSessionID},
				CreatedTS: 1_000, UpdatedTS: 3_000, CommentCount: 2,
			},
			{
				IssueID: griteexport.ID(serviceIssueMissingSession), Title: "Waiting issue", State: "open",
				Labels:    []string{"status:ask", "mode:run", "session:" + waitingSession.ProviderSessionID},
				CreatedTS: 2_000, UpdatedTS: 4_000, CommentCount: 1,
			},
			{
				IssueID: griteexport.ID(serviceIssueMissingPlan), Title: "Terminal mask issue", State: "open",
				Labels:    []string{"status:in_progress", "mode:run", "session:" + terminalMaskSession.ProviderSessionID},
				CreatedTS: 2_100, UpdatedTS: 4_000, CommentCount: 1,
			},
		},
		Events: events,
	}
	report, err := service.Import(ctx, snapshot, migrategrite.ImportOptions{Workspace: native.ImportWorkspace{
		RepoKey:     "github.com/flanksource/gavel-migrategrite-qualifying-runs",
		RootPath:    t.TempDir(),
		DisplayName: "qualifying runs",
	}})
	require.NoError(t, err)
	assertNoWarning(t, report.Warnings, serviceIssueLinked, "captain_live_run_missing")
	assertNoWarning(t, report.Warnings, serviceIssueMissingSession, "captain_live_request_missing")
	assertNoWarning(t, report.Warnings, serviceIssueMissingPlan, "captain_live_run_missing")

	runningIssue, err := repository.GetIssueByRef(ctx, report.WorkspaceID, serviceIssueLinked)
	require.NoError(t, err)
	require.NotNil(t, runningIssue.ActivePromptRunID)
	assert.Equal(t, planRun.ID, *runningIssue.ActivePromptRunID)
	assert.NotEqual(t, olderRunning.ID, *runningIssue.ActivePromptRunID)
	assert.NotEqual(t, newerRunningTerminal.ID, *runningIssue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionPlanning, runningIssue.ExecutionState)
	waitingIssue, err := repository.GetIssueByRef(ctx, report.WorkspaceID, serviceIssueMissingSession)
	require.NoError(t, err)
	require.NotNil(t, waitingIssue.ActivePromptRunID)
	assert.Equal(t, olderWaiting.ID, *waitingIssue.ActivePromptRunID)
	assert.NotEqual(t, newerWaitingTerminal.ID, *waitingIssue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionWaiting, waitingIssue.ExecutionState)
	terminalMaskIssue, err := repository.GetIssueByRef(ctx, report.WorkspaceID, serviceIssueMissingPlan)
	require.NoError(t, err)
	require.NotNil(t, terminalMaskIssue.ActivePromptRunID)
	assert.Equal(t, olderTerminalMaskedLive.ID, *terminalMaskIssue.ActivePromptRunID)
	assert.NotEqual(t, newerTerminalMask.ID, *terminalMaskIssue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionRunning, terminalMaskIssue.ExecutionState)
}

func openMigrationService(t *testing.T) (*migrategrite.Service, *native.Repository, *captaindb.DB, *gorm.DB) {
	t.Helper()
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres migration service tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(t.TempDir(), "postgres"),
		Database: "gavel_migrategrite_service",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })

	service, err := migrategrite.NewService(opened.Gorm())
	require.NoError(t, err)
	repository, err := native.NewRepository(opened.Gorm())
	require.NoError(t, err)
	captain, err := captaindb.Use(opened.Gorm())
	require.NoError(t, err)
	return service, repository, captain, opened.Gorm()
}

func createCaptainRun(t *testing.T, captain *captaindb.DB, providerSessionID string) (*captaindb.Session, *captaindb.PromptRun) {
	t.Helper()
	session, err := captain.CreateOrGetSession(t.Context(), captaindb.CreateSessionInput{
		ProviderSessionID: providerSessionID,
		Source:            "migration-test",
		Provider:          "test",
		HostID:            "local",
	})
	require.NoError(t, err)
	run, err := captain.CreatePromptRun(t.Context(), captaindb.CreatePromptRunInput{
		SessionID: session.ID, AdmissionKey: "migration-test:" + providerSessionID,
	})
	require.NoError(t, err)
	phase := captaindb.PromptRunPhaseGenerate
	state := captaindb.PromptRunStateRunning
	run, err = captain.UpdatePromptRun(t.Context(), captaindb.UpdatePromptRunInput{
		ID: run.ID, ExpectedVersion: run.Version, Phase: &phase, State: &state,
	})
	require.NoError(t, err)
	return session, run
}

func createMaskedQualifyingRunPair(
	t *testing.T,
	captain *captaindb.DB,
	providerSessionID string,
	qualifyingState captaindb.PromptRunState,
) (*captaindb.Session, *captaindb.PromptRun, *captaindb.PromptRun) {
	t.Helper()
	ctx := t.Context()
	session, err := captain.CreateOrGetSession(ctx, captaindb.CreateSessionInput{
		ProviderSessionID: providerSessionID,
		Source:            "migration-test",
		Provider:          "test",
		HostID:            "local",
	})
	require.NoError(t, err)
	older, err := captain.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{
		SessionID: session.ID, AdmissionKey: "migration-test:" + providerSessionID + ":older",
	})
	require.NoError(t, err)
	finished := captaindb.PromptRunPhaseFinished
	succeeded := captaindb.PromptRunStateSucceeded
	older, err = captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: older.ID, ExpectedVersion: older.Version, Phase: &finished, State: &succeeded,
	})
	require.NoError(t, err)
	newer, err := captain.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{
		SessionID: session.ID, AdmissionKey: "migration-test:" + providerSessionID + ":newer",
	})
	require.NoError(t, err)
	newer, err = captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: newer.ID, ExpectedVersion: newer.Version, Phase: &finished, State: &succeeded,
	})
	require.NoError(t, err)
	phase := captaindb.PromptRunPhaseGenerate
	if qualifyingState == captaindb.PromptRunStateWaiting {
		phase = captaindb.PromptRunPhaseFeedback
	}
	older, err = captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: older.ID, ExpectedVersion: older.Version, Phase: &phase, State: &qualifyingState,
	})
	require.NoError(t, err)
	base := time.Date(2025, time.July, 8, 9, 10, 11, 0, time.UTC)
	require.NoError(t, captain.Gorm().WithContext(ctx).Exec(
		`UPDATE captain_prompt_runs SET created_at = CAST(? AS timestamptz) WHERE id = ?`, base, older.ID,
	).Error)
	require.NoError(t, captain.Gorm().WithContext(ctx).Exec(
		`UPDATE captain_prompt_runs SET created_at = CAST(? AS timestamptz) WHERE id = ?`, base.Add(time.Minute), newer.ID,
	).Error)
	older, err = captain.GetPromptRun(ctx, older.ID)
	require.NoError(t, err)
	newer, err = captain.GetPromptRun(ctx, newer.ID)
	require.NoError(t, err)
	return session, older, newer
}

func singleServiceSnapshot(
	t *testing.T,
	issueID, title, statusLabel, sessionIdentity, planPath string,
) griteexport.Snapshot {
	t.Helper()
	events := []griteexport.Event{
		serviceEvent(t, "event-single-created", issueID, 1_000, "IssueCreated", map[string]any{
			"title": title, "body": "body",
		}),
		serviceEvent(t, "event-single-plan", issueID, 2_000, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n<!-- gavel:state {\"planPath\":\"" + planPath + "\"} -->",
		}),
	}
	return griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 2_000, EventCount: len(events)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(issueID), Title: title, State: "open",
			Labels:    []string{statusLabel, "mode:plan", "session:" + sessionIdentity},
			CreatedTS: 1_000, UpdatedTS: 2_000, CommentCount: 1,
		}},
		Events: events,
	}
}

func serviceSnapshot(t *testing.T, linkedSession, missingSession, missingPlanSession, unresolvedPlanSession, planPath, unresolvedPlanPath string) griteexport.Snapshot {
	t.Helper()
	events := []griteexport.Event{
		serviceEvent(t, "event-linked-created", serviceIssueLinked, 1_000, "IssueCreated", map[string]any{
			"title": "Linked issue", "body": "body",
		}),
		serviceEvent(t, "event-linked-plan", serviceIssueLinked, 2_000, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n<!-- gavel:state {\"planPath\":\"" + planPath + "\"} -->",
		}),
		serviceEvent(t, "event-missing-session-created", serviceIssueMissingSession, 3_000, "IssueCreated", map[string]any{
			"title": "Missing session", "body": "body",
		}),
		serviceEvent(t, "event-missing-plan-created", serviceIssueMissingPlan, 4_000, "IssueCreated", map[string]any{
			"title": "Missing plan", "body": "body",
		}),
		serviceEvent(t, "event-missing-plan", serviceIssueMissingPlan, 5_000, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n<!-- gavel:state {\"planPath\":\"plans/missing.md\"} -->",
		}),
		serviceEvent(t, "event-unresolved-plan-created", serviceIssueUnresolvedPlan, 5_100, "IssueCreated", map[string]any{
			"title": "Readable unresolved plan", "body": "body",
		}),
		serviceEvent(t, "event-unresolved-plan", serviceIssueUnresolvedPlan, 5_200, "CommentAdded", map[string]any{
			"body": "**Agent state**\n\n<!-- gavel:state {\"planPath\":\"" + unresolvedPlanPath + "\"} -->",
		}),
	}
	return griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 8_000, EventCount: len(events)},
		Issues: []griteexport.Issue{
			{
				IssueID: griteexport.ID(serviceIssueLinked), Title: "Linked issue", State: "open",
				Labels:    []string{"status:in_progress", "mode:run", "session:" + linkedSession},
				CreatedTS: 1_000, UpdatedTS: 6_000, CommentCount: 1,
			},
			{
				IssueID: griteexport.ID(serviceIssueMissingSession), Title: "Missing session", State: "open",
				Labels:    []string{"status:in_progress", "mode:run", "session:" + missingSession},
				CreatedTS: 3_000, UpdatedTS: 7_000,
			},
			{
				IssueID: griteexport.ID(serviceIssueMissingPlan), Title: "Missing plan", State: "open",
				Labels:    []string{"status:pending", "mode:plan", "session:" + missingPlanSession},
				CreatedTS: 4_000, UpdatedTS: 8_000, CommentCount: 1,
			},
			{
				IssueID: griteexport.ID(serviceIssueUnresolvedPlan), Title: "Readable unresolved plan", State: "open",
				Labels:    []string{"status:pending", "mode:plan", "session:" + unresolvedPlanSession},
				CreatedTS: 5_100, UpdatedTS: 8_100, CommentCount: 1,
			},
		},
		Events: events,
	}
}

func serviceEvent(t *testing.T, eventID, issueID string, timestampMS int64, name string, payload any) griteexport.Event {
	t.Helper()
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return griteexport.Event{
		EventID: griteexport.ID(eventID), IssueID: griteexport.ID(issueID), Actor: "migration-test",
		TimestampMS: timestampMS, Kind: griteexport.Kind{name: raw},
	}
}

func assertWarning(t *testing.T, warnings []migrategrite.Warning, issueID, code string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.IssueID == issueID && warning.Code == code {
			return
		}
	}
	assert.Fail(t, "warning not found", "issue=%s code=%s warnings=%v", issueID, code, warnings)
}

func assertNoWarning(t *testing.T, warnings []migrategrite.Warning, issueID, code string) {
	t.Helper()
	for _, warning := range warnings {
		if warning.IssueID == issueID && warning.Code == code {
			assert.Fail(t, "unexpected warning", "issue=%s code=%s warning=%v", issueID, code, warning)
		}
	}
}

func assertIssueWarningEvent(t *testing.T, repository *native.Repository, issueID uuid.UUID, code string) {
	t.Helper()
	events, err := repository.ListEvents(t.Context(), issueID)
	require.NoError(t, err)
	for _, event := range events {
		if event.Kind != "migration_warning" {
			continue
		}
		var payload struct {
			Code string `json:"code"`
		}
		require.NoError(t, json.Unmarshal(event.Payload, &payload))
		if payload.Code == code {
			return
		}
	}
	assert.Fail(t, "migration warning event not found", "issue=%s code=%s events=%v", issueID, code, events)
}

type serviceCounts struct {
	Events         int64
	PromptRunLinks int64
	PlanLinks      int64
	PlanRevisions  int64
}

func readServiceCounts(t *testing.T, db *gorm.DB) serviceCounts {
	t.Helper()
	var counts serviceCounts
	require.NoError(t, db.Table("todo_issue_events").Count(&counts.Events).Error)
	require.NoError(t, db.Table("todo_issue_prompt_runs").Count(&counts.PromptRunLinks).Error)
	require.NoError(t, db.Table("todo_issue_plans").Count(&counts.PlanLinks).Error)
	require.NoError(t, db.Table("captain_plan_revisions").Count(&counts.PlanRevisions).Error)
	return counts
}
