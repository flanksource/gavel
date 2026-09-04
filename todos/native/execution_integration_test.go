package native_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

	coordinatedIssue := createIssue(t, repo, workspace.ID, "coordinated launch")
	captain, err := captaindb.Use(db)
	require.NoError(t, err)
	coordinator, err := native.NewLaunchCoordinator(captain, repo)
	require.NoError(t, err)
	sessionInput := captaindb.CreateSessionInput{
		ProviderSessionID: "provider-session-coordinator",
		Source:            "gavel",
		Provider:          "test",
		CWD:               "/workspace/gavel-execution-integration",
	}
	// The projection classifies a run by what it was asked to do, never by the
	// step's name: a plan run is one whose rendered spec runs in plan mode.
	promptInput := captaindb.CreatePromptRunInput{
		AdmissionKey:   "gavel:coordinated-launch:plan:0",
		Origin:         "gavel.todos.plan",
		PromptMarkdown: "Draft a plan",
		RenderedSpec:   map[string]any{"permissions": map[string]any{"mode": "plan"}},
	}
	attachment := native.PromptRunLaunchAttachment{
		IssueID:              coordinatedIssue.ID,
		StepKind:             native.StepPlan,
		Ordinal:              0,
		ExpectedIssueVersion: coordinatedIssue.Version,
		Actor:                "execution-test",
		Owner:                &owner,
	}
	launchInput := native.PromptRunLaunchInput{
		RootSession: captaindb.CreateSessionInput{
			ID: coordinatedIssue.ID, Source: "gavel", Provider: "todos",
			CWD: "/workspace/gavel-execution-integration",
		},
		Session: sessionInput, PromptRun: promptInput, Attachment: attachment,
	}
	launch, err := coordinator.LaunchPromptRun(ctx, launchInput)
	require.NoError(t, err)
	require.NotNil(t, launch.Session)
	require.NotNil(t, launch.PromptRun)
	require.NotNil(t, launch.Issue)
	assert.True(t, launch.DispatchOwned)
	assert.Equal(t, launch.Session.ID, launch.PromptRun.SessionID)
	require.NotNil(t, launch.Issue.ActivePromptRunID)
	assert.Equal(t, launch.PromptRun.ID, *launch.Issue.ActivePromptRunID)
	assert.Equal(t, native.ExecutionPlanning, launch.Issue.ExecutionState)
	assert.Equal(t, coordinatedIssue.Version+1, launch.Issue.Version, "attaching the run is the one mutation; execution state is derived at read time")

	// The Captain admission key and native exact-replay guard make the whole
	// operation retryable with the original issue version.
	replay, err := coordinator.LaunchPromptRun(ctx, launchInput)
	require.NoError(t, err)
	assert.Equal(t, launch.Session.ID, replay.Session.ID)
	assert.Equal(t, launch.PromptRun.ID, replay.PromptRun.ID)
	assert.Equal(t, launch.Issue.Version, replay.Issue.Version)
	assert.False(t, replay.DispatchOwned)

	sourceRunID := launch.PromptRun.ID
	planInput := captaindb.CreatePlanInput{
		SourceSessionID:   launch.Session.ID,
		SourcePromptRunID: &sourceRunID,
		Variant:           "primary",
		Title:             "Durable coordinated plan",
		Path:              "/tmp/deleted-agent-plan.md",
	}
	revisionInput := captaindb.AppendPlanRevisionInput{
		PlanMarkdown: "# Durable plan\n\n1. Implement the native seam.",
		CreatedBy:    "execution-test",
	}
	planAttachment := native.PlanSelectionAttachment{
		IssueID:              coordinatedIssue.ID,
		Ordinal:              0,
		ExpectedIssueVersion: launch.Issue.Version,
		Actor:                "execution-test",
	}
	persisted, err := coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: planInput, Revision: revisionInput, Attachment: planAttachment})
	require.NoError(t, err)
	require.NotNil(t, persisted.Plan)
	require.NotNil(t, persisted.Revision)
	require.NotNil(t, persisted.Issue.SelectedPlanID)
	assert.Equal(t, persisted.Plan.ID, persisted.Revision.PlanID)
	assert.Equal(t, persisted.Plan.ID, *persisted.Issue.SelectedPlanID)
	assert.Equal(t, revisionInput.PlanMarkdown, persisted.Revision.PlanMarkdown)
	require.NotNil(t, persisted.Plan.LatestRevision)
	assert.Equal(t, persisted.Revision.ID, persisted.Plan.LatestRevision.ID)
	assert.Equal(t, launch.Issue.Version+1, persisted.Issue.Version)

	// Both Captain's content hash and Gavel's exact-selection guard are
	// idempotent with the original issue version.
	persistedReplay, err := coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: planInput, Revision: revisionInput, Attachment: planAttachment})
	require.NoError(t, err)
	assert.Equal(t, persisted.Plan.ID, persistedReplay.Plan.ID)
	assert.Equal(t, persisted.Revision.ID, persistedReplay.Revision.ID)
	assert.Equal(t, persisted.Issue.Version, persistedReplay.Issue.Version)

	wrongSessionPlan := planInput
	wrongSessionPlan.SourceSessionID = uuid.New()
	_, err = coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: wrongSessionPlan, Revision: revisionInput, Attachment: planAttachment})
	require.ErrorIs(t, err, captaindb.ErrPlanConflict, "stale exact replay must still validate immutable session identity")
	wrongIDPlan := planInput
	wrongIDPlan.ID = uuid.New()
	_, err = coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: wrongIDPlan, Revision: revisionInput, Attachment: planAttachment})
	require.ErrorIs(t, err, captaindb.ErrPlanConflict, "stale exact replay must still validate a supplied plan ID")

	secondRevisionInput := captaindb.AppendPlanRevisionInput{
		PlanMarkdown: "# Durable plan v2\n\n2. Record same-plan changes in Gavel.",
		CreatedBy:    "execution-test",
	}
	_, err = coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: planInput, Revision: secondRevisionInput, Attachment: planAttachment})
	require.ErrorIs(t, err, native.ErrVersionConflict, "a stale caller cannot append to the selected plan")
	revisionsAfterStale, err := captain.ListPlanRevisions(ctx, persisted.Plan.ID)
	require.NoError(t, err)
	require.Len(t, revisionsAfterStale, 1)
	issueAfterStaleRevision, err := repo.GetIssue(ctx, coordinatedIssue.ID)
	require.NoError(t, err)
	assert.Equal(t, persisted.Issue.Version, issueAfterStaleRevision.Version)

	secondRevisionAttachment := planAttachment
	secondRevisionAttachment.ExpectedIssueVersion = persisted.Issue.Version
	secondRevision, err := coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: planInput, Revision: secondRevisionInput, Attachment: secondRevisionAttachment})
	require.NoError(t, err)
	assert.Equal(t, persisted.Plan.ID, secondRevision.Plan.ID)
	assert.NotEqual(t, persisted.Revision.ID, secondRevision.Revision.ID)
	require.NotNil(t, secondRevision.Issue.SelectedPlanID)
	assert.Equal(t, persisted.Plan.ID, *secondRevision.Issue.SelectedPlanID)
	assert.Equal(t, persisted.Issue.Version+1, secondRevision.Issue.Version)
	secondRevisionEvents, err := repo.ListEvents(ctx, coordinatedIssue.ID)
	require.NoError(t, err)
	require.NotEmpty(t, secondRevisionEvents)
	assert.Equal(t, "plan_revision_persisted", secondRevisionEvents[len(secondRevisionEvents)-1].Kind)

	secondRevisionReplay, err := coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: planInput, Revision: secondRevisionInput, Attachment: secondRevisionAttachment})
	require.NoError(t, err)
	assert.Equal(t, secondRevision.Revision.ID, secondRevisionReplay.Revision.ID)
	assert.Equal(t, secondRevision.Issue.Version, secondRevisionReplay.Issue.Version)

	// An independent Captain writer can win the plan lock, but the coordinator
	// must snapshot revisions only after that lock is released. It then observes
	// an exact no-op instead of attributing the independent revision to Gavel.
	independentRevisionInput := captaindb.AppendPlanRevisionInput{
		PlanID:       persisted.Plan.ID,
		PlanMarkdown: "# Durable plan v3\n\n3. Written independently by Captain.",
		CreatedBy:    "captain-independent",
	}
	independentRevisionReady := make(chan struct{})
	releaseIndependentRevision := make(chan struct{})
	t.Cleanup(func() {
		select {
		case releaseIndependentRevision <- struct{}{}:
		default:
		}
	})
	independentRevisionDone := make(chan error, 1)
	var independentRevision *captaindb.PlanRevision
	go func() {
		independentRevisionDone <- captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
			var appendErr error
			independentRevision, appendErr = captainTx.AppendPlanRevision(ctx, independentRevisionInput)
			if appendErr != nil {
				return appendErr
			}
			close(independentRevisionReady)
			<-releaseIndependentRevision
			return nil
		})
	}()
	select {
	case <-independentRevisionReady:
	case independentErr := <-independentRevisionDone:
		require.NoError(t, independentErr)
		t.Fatal("independent revision transaction ended before acquiring its plan lock")
	}
	revisionRaceEventsBefore, err := repo.ListEvents(ctx, coordinatedIssue.ID)
	require.NoError(t, err)
	revisionRaceIssueBefore, err := repo.GetIssue(ctx, coordinatedIssue.ID)
	require.NoError(t, err)
	revisionCoordinatorDone := make(chan struct {
		result *native.PersistedPlan
		err    error
	}, 1)
	go func() {
		result, persistErr := coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
			Plan:     planInput,
			Revision: independentRevisionInput,
			Attachment: native.PlanSelectionAttachment{
				IssueID:              coordinatedIssue.ID,
				Ordinal:              0,
				ExpectedIssueVersion: revisionRaceIssueBefore.Version,
				Actor:                "execution-test",
			},
		})
		revisionCoordinatorDone <- struct {
			result *native.PersistedPlan
			err    error
		}{result: result, err: persistErr}
	}()
	waitForBlockedDatabaseLock(t, db)
	releaseIndependentRevision <- struct{}{}
	require.NoError(t, <-independentRevisionDone)
	revisionRaceResult := <-revisionCoordinatorDone
	require.NoError(t, revisionRaceResult.err)
	require.NotNil(t, independentRevision)
	assert.Equal(t, independentRevision.ID, revisionRaceResult.result.Revision.ID)
	assert.Equal(t, revisionRaceIssueBefore.Version, revisionRaceResult.result.Issue.Version)
	revisionRaceEventsAfter, err := repo.ListEvents(ctx, coordinatedIssue.ID)
	require.NoError(t, err)
	assert.Len(t, revisionRaceEventsAfter, len(revisionRaceEventsBefore))

	approvalIssue := createIssue(t, repo, workspace.ID, "approval selection")
	approvalPlan, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID:   launch.Session.ID,
		SourcePromptRunID: &sourceRunID,
		Variant:           "approval-selection",
		Title:             "Approval selection plan",
	})
	require.NoError(t, err)
	approvalRevision, err := captain.AppendPlanRevision(ctx, captaindb.AppendPlanRevisionInput{
		PlanID:       approvalPlan.ID,
		PlanMarkdown: "# Approved plan\n\nShip the transaction boundary.",
		CreatedBy:    "execution-test",
	})
	require.NoError(t, err)
	approvalInput := captaindb.ApprovePlanRevisionInput{
		PlanID:     approvalPlan.ID,
		RevisionID: approvalRevision.ID,
		ApprovedBy: "reviewer",
		Comment:    "approved",
	}
	approvalAttachment := native.PlanSelectionAttachment{
		IssueID:              approvalIssue.ID,
		Ordinal:              0,
		ExpectedIssueVersion: approvalIssue.Version - 1,
		Actor:                "reviewer",
	}
	_, err = coordinator.ApproveAndSelectPlan(ctx, approvalInput, approvalAttachment)
	require.ErrorIs(t, err, native.ErrVersionConflict)
	approvalAfterRollback, err := captain.GetPlan(ctx, approvalPlan.ID)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PlanApprovalPending, approvalAfterRollback.ApprovalState)
	assert.Nil(t, approvalAfterRollback.ApprovedRevisionID)
	approvalIssue, err = repo.GetIssue(ctx, approvalIssue.ID)
	require.NoError(t, err)
	assert.Nil(t, approvalIssue.SelectedPlanID)
	approvalLinks, err := repo.ListPlans(ctx, approvalIssue.ID)
	require.NoError(t, err)
	assert.Empty(t, approvalLinks)

	approvalAttachment.ExpectedIssueVersion = approvalIssue.Version
	approved, err := coordinator.ApproveAndSelectPlan(ctx, approvalInput, approvalAttachment)
	require.NoError(t, err)
	assert.Equal(t, captaindb.PlanApprovalApproved, approved.Plan.ApprovalState)
	require.NotNil(t, approved.Plan.ApprovedRevisionID)
	assert.Equal(t, approvalRevision.ID, *approved.Plan.ApprovedRevisionID)
	require.NotNil(t, approved.Issue.SelectedPlanID)
	assert.Equal(t, approvalPlan.ID, *approved.Issue.SelectedPlanID)
	assert.Equal(t, approvalIssue.Version+1, approved.Issue.Version)

	approvedReplay, err := coordinator.ApproveAndSelectPlan(ctx, approvalInput, approvalAttachment)
	require.NoError(t, err)
	assert.Equal(t, approved.Plan.ID, approvedReplay.Plan.ID)
	assert.Equal(t, approved.Issue.Version, approvedReplay.Issue.Version)

	changedApproval := approvalInput
	changedApproval.ApprovedBy = "second-reviewer"
	changedApproval.Comment = "approved after another review"
	_, err = coordinator.ApproveAndSelectPlan(ctx, changedApproval, approvalAttachment)
	require.ErrorIs(t, err, native.ErrVersionConflict, "a stale caller cannot change approval on the selected plan")
	approvalAfterStaleChange, err := captain.GetPlan(ctx, approvalPlan.ID)
	require.NoError(t, err)
	assert.Equal(t, approvalInput.ApprovedBy, approvalAfterStaleChange.ApprovedBy)
	assert.Equal(t, approvalInput.Comment, approvalAfterStaleChange.ApprovalComment)
	approvalIssueAfterStale, err := repo.GetIssue(ctx, approvalIssue.ID)
	require.NoError(t, err)
	assert.Equal(t, approved.Issue.Version, approvalIssueAfterStale.Version)

	changedApprovalAttachment := approvalAttachment
	changedApprovalAttachment.ExpectedIssueVersion = approved.Issue.Version
	changedApproved, err := coordinator.ApproveAndSelectPlan(ctx, changedApproval, changedApprovalAttachment)
	require.NoError(t, err)
	assert.Equal(t, changedApproval.ApprovedBy, changedApproved.Plan.ApprovedBy)
	assert.Equal(t, changedApproval.Comment, changedApproved.Plan.ApprovalComment)
	assert.Equal(t, approved.Issue.Version+1, changedApproved.Issue.Version)
	changedApprovalEvents, err := repo.ListEvents(ctx, approvalIssue.ID)
	require.NoError(t, err)
	require.NotEmpty(t, changedApprovalEvents)
	assert.Equal(t, "plan_approval_changed", changedApprovalEvents[len(changedApprovalEvents)-1].Kind)

	changedApprovalReplay, err := coordinator.ApproveAndSelectPlan(ctx, changedApproval, changedApprovalAttachment)
	require.NoError(t, err)
	assert.Equal(t, changedApproved.Issue.Version, changedApprovalReplay.Issue.Version)

	// Approval exactness uses the same issue->plan lock order. If Captain changes
	// approval independently first, Gavel observes the committed state and emits
	// no event claiming that independent change as its own mutation.
	independentApproval := changedApproval
	independentApproval.ApprovedBy = "captain-independent"
	independentApproval.Comment = "approved independently"
	independentApprovalReady := make(chan struct{})
	releaseIndependentApproval := make(chan struct{})
	t.Cleanup(func() {
		select {
		case releaseIndependentApproval <- struct{}{}:
		default:
		}
	})
	independentApprovalDone := make(chan error, 1)
	go func() {
		independentApprovalDone <- captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
			if _, approvalErr := captainTx.ApprovePlanRevision(ctx, independentApproval); approvalErr != nil {
				return approvalErr
			}
			close(independentApprovalReady)
			<-releaseIndependentApproval
			return nil
		})
	}()
	select {
	case <-independentApprovalReady:
	case independentErr := <-independentApprovalDone:
		require.NoError(t, independentErr)
		t.Fatal("independent approval transaction ended before acquiring its plan lock")
	}
	approvalRaceEventsBefore, err := repo.ListEvents(ctx, approvalIssue.ID)
	require.NoError(t, err)
	approvalRaceIssueBefore, err := repo.GetIssue(ctx, approvalIssue.ID)
	require.NoError(t, err)
	approvalCoordinatorDone := make(chan struct {
		result *native.ApprovedPlanSelection
		err    error
	}, 1)
	go func() {
		result, approvalErr := coordinator.ApproveAndSelectPlan(ctx, independentApproval, native.PlanSelectionAttachment{
			IssueID:              approvalIssue.ID,
			Ordinal:              0,
			ExpectedIssueVersion: approvalRaceIssueBefore.Version,
			Actor:                "execution-test",
		})
		approvalCoordinatorDone <- struct {
			result *native.ApprovedPlanSelection
			err    error
		}{result: result, err: approvalErr}
	}()
	waitForBlockedDatabaseLock(t, db)
	releaseIndependentApproval <- struct{}{}
	require.NoError(t, <-independentApprovalDone)
	approvalRaceResult := <-approvalCoordinatorDone
	require.NoError(t, approvalRaceResult.err)
	assert.Equal(t, independentApproval.ApprovedBy, approvalRaceResult.result.Plan.ApprovedBy)
	assert.Equal(t, approvalRaceIssueBefore.Version, approvalRaceResult.result.Issue.Version)
	approvalRaceEventsAfter, err := repo.ListEvents(ctx, approvalIssue.ID)
	require.NoError(t, err)
	assert.Len(t, approvalRaceEventsAfter, len(approvalRaceEventsBefore))

	raceIssue := createIssue(t, repo, workspace.ID, "approval race")
	racePlans := make([]*captaindb.Plan, 2)
	raceRevisions := make([]*captaindb.PlanRevision, 2)
	for i := range racePlans {
		racePlans[i], err = captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
			SourceSessionID:   launch.Session.ID,
			SourcePromptRunID: &sourceRunID,
			Variant:           "approval-race-" + string(rune('a'+i)),
		})
		require.NoError(t, err)
		raceRevisions[i], err = captain.AppendPlanRevision(ctx, captaindb.AppendPlanRevisionInput{
			PlanID:       racePlans[i].ID,
			PlanMarkdown: "# Race candidate " + string(rune('A'+i)),
		})
		require.NoError(t, err)
	}
	raceErrors := runConcurrently(
		func() error {
			_, err := coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
				PlanID: racePlans[0].ID, RevisionID: raceRevisions[0].ID, ApprovedBy: "reviewer-a",
			}, native.PlanSelectionAttachment{
				IssueID: raceIssue.ID, Ordinal: 0, ExpectedIssueVersion: raceIssue.Version, Actor: "reviewer-a",
			})
			return err
		},
		func() error {
			_, err := coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
				PlanID: racePlans[1].ID, RevisionID: raceRevisions[1].ID, ApprovedBy: "reviewer-b",
			}, native.PlanSelectionAttachment{
				IssueID: raceIssue.ID, Ordinal: 0, ExpectedIssueVersion: raceIssue.Version, Actor: "reviewer-b",
			})
			return err
		},
	)
	assertExactlyOne(t, raceErrors, nil, native.ErrVersionConflict)
	raceIssue, err = repo.GetIssue(ctx, raceIssue.ID)
	require.NoError(t, err)
	require.NotNil(t, raceIssue.SelectedPlanID)
	approvedRacePlans := 0
	for _, candidate := range racePlans {
		stored, getErr := captain.GetPlan(ctx, candidate.ID)
		require.NoError(t, getErr)
		if stored.ApprovalState == captaindb.PlanApprovalApproved {
			approvedRacePlans++
			assert.Equal(t, candidate.ID, *raceIssue.SelectedPlanID)
		} else {
			assert.Equal(t, captaindb.PlanApprovalPending, stored.ApprovalState)
		}
	}
	assert.Equal(t, 1, approvedRacePlans)
	raceLinks, err := repo.ListPlans(ctx, raceIssue.ID)
	require.NoError(t, err)
	require.Len(t, raceLinks, 1)
	assert.Equal(t, *raceIssue.SelectedPlanID, raceLinks[0].PlanID)

	// A native attachment failure rolls back the Captain rows created earlier
	// in the same coordinator transaction.
	failedIssue := createIssue(t, repo, workspace.ID, "failed coordinated launch")
	failedSession := captaindb.CreateSessionInput{
		ProviderSessionID: "provider-session-rolled-back",
		Source:            "gavel",
		Provider:          "test",
	}
	failedAdmissionKey := "gavel:rolled-back-launch:run:0"
	_, err = coordinator.LaunchPromptRun(ctx, native.PromptRunLaunchInput{
		RootSession: captaindb.CreateSessionInput{
			ID: failedIssue.ID, Source: "gavel", Provider: "todos",
		},
		Session: failedSession,
		PromptRun: captaindb.CreatePromptRunInput{
			AdmissionKey: failedAdmissionKey,
			Origin:       "gavel.todos.run",
		},
		Attachment: native.PromptRunLaunchAttachment{
			IssueID:              failedIssue.ID,
			StepKind:             native.StepRun,
			Ordinal:              0,
			ExpectedIssueVersion: failedIssue.Version - 1,
			Actor:                "execution-test",
		},
	})
	require.ErrorIs(t, err, native.ErrVersionConflict)
	_, err = captain.GetSessionByIdentity(ctx, failedSession.ProviderSessionID, failedSession.Source, failedSession.Provider, "local")
	require.ErrorIs(t, err, captaindb.ErrSessionNotFound)
	var failedRuns int64
	require.NoError(t, db.Table("captain_prompt_runs").Where("admission_key = ?", failedAdmissionKey).Count(&failedRuns).Error)
	assert.Zero(t, failedRuns)
	failedIssue, err = repo.GetIssue(ctx, failedIssue.ID)
	require.NoError(t, err)
	assert.Nil(t, failedIssue.ActivePromptRunID)

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

func waitForBlockedDatabaseLock(t *testing.T, db *gorm.DB) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiters int64
		err := db.Raw(`
			SELECT COUNT(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`,
		).Scan(&waiters).Error
		require.NoError(t, err)
		if waiters > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out waiting for coordinator to block on the Captain plan row lock")
}
