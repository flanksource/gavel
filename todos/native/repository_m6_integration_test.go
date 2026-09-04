package native_test

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryGlobalReferenceResolution(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	firstWorkspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/global-first"})
	require.NoError(t, err)
	secondWorkspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/global-second"})
	require.NoError(t, err)

	nativeID := uuid.MustParse("11111111-2222-4333-8444-555555555555")
	nativeIssue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		ID:          nativeID,
		WorkspaceID: firstWorkspace.ID,
		Title:       "Native UUID owner",
	})
	require.NoError(t, err)
	aliasOwner, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: secondWorkspace.ID,
		Aliases: []native.AliasInput{
			{Alias: nativeID.String(), Kind: "legacy"},
			{Alias: "tiny", Kind: "legacy"},
		},
		Title: "Exact alias owner",
	})
	require.NoError(t, err)

	resolved, err := repo.GetIssueByGlobalRef(ctx, nativeID.String())
	require.NoError(t, err)
	assert.Equal(t, aliasOwner.ID, resolved.ID, "an exact alias must win before UUID parsing")
	resolved, err = repo.GetIssueByGlobalRef(ctx, "tiny")
	require.NoError(t, err)
	assert.Equal(t, aliasOwner.ID, resolved.ID, "an exact alias may be shorter than the prefix minimum")

	shortUUIDIssue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		ID:          uuid.MustParse("12345678-2222-4333-8444-555555555555"),
		WorkspaceID: firstWorkspace.ID,
		Title:       "Short native UUID",
	})
	require.NoError(t, err)
	resolved, err = repo.GetIssueByGlobalRef(ctx, "12345678")
	require.NoError(t, err)
	assert.Equal(t, shortUUIDIssue.ID, resolved.ID)

	legacyAlias := "e2a3b8c2d0f7c9a98b400dc78e8a94a5"
	legacyIssue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: firstWorkspace.ID,
		Aliases:     []native.AliasInput{{Alias: legacyAlias, Kind: "external"}},
		Title:       "Legacy alias",
	})
	require.NoError(t, err)
	for _, ref := range []string{legacyAlias, legacyAlias[:native.MinShortReferenceLength]} {
		resolved, err = repo.GetIssueByGlobalRef(ctx, ref)
		require.NoError(t, err)
		assert.Equal(t, legacyIssue.ID, resolved.ID)
	}

	for _, workspaceID := range []uuid.UUID{firstWorkspace.ID, secondWorkspace.ID} {
		_, err = repo.CreateIssue(ctx, native.CreateIssueInput{
			WorkspaceID: workspaceID,
			Aliases:     []native.AliasInput{{Alias: "shared-exact", Kind: "legacy"}},
			Title:       "Shared exact alias",
		})
		require.NoError(t, err)
	}
	_, err = repo.GetIssueByGlobalRef(ctx, "shared-exact")
	require.ErrorIs(t, err, native.ErrAmbiguousReference)

	for index, input := range []struct {
		workspaceID uuid.UUID
		alias       string
	}{
		{firstWorkspace.ID, "prefix88-first"},
		{secondWorkspace.ID, "prefix88-second"},
	} {
		_, err = repo.CreateIssue(ctx, native.CreateIssueInput{
			WorkspaceID: input.workspaceID,
			Aliases:     []native.AliasInput{{Alias: input.alias, Kind: "legacy"}},
			Title:       "Shared prefix " + string(rune('A'+index)),
		})
		require.NoError(t, err)
	}
	_, err = repo.GetIssueByGlobalRef(ctx, "prefix88")
	require.ErrorIs(t, err, native.ErrAmbiguousReference)

	_, err = repo.GetIssueByGlobalRef(ctx, "short")
	require.ErrorIs(t, err, native.ErrInvalidInput)
	_, err = repo.GetIssueByGlobalRef(ctx, "missing8")
	require.ErrorIs(t, err, native.ErrNotFound)

	// The native issue still exists even though its UUID-shaped global reference
	// is intentionally shadowed by an exact alias in another workspace.
	byID, err := repo.GetIssue(ctx, nativeIssue.ID)
	require.NoError(t, err)
	assert.Equal(t, nativeIssue.ID, byID.ID)
}

func TestRepositorySessionReferenceResolution(t *testing.T) {
	repo, db := openRepository(t)
	ctx := t.Context()
	workspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey: "github.com/example/session-reference",
	})
	require.NoError(t, err)
	issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: workspace.ID,
		Title:       "Session-owned issue",
	})
	require.NoError(t, err)

	providerSessionID := uuid.New()
	orchestrationSessionID := uuid.New()
	transcriptSessionID := uuid.New()
	promptRunID := uuid.New()
	require.NoError(t, db.Exec(`
		INSERT INTO captain_sessions (id, provider_session_id, source, provider, host_id)
		VALUES
			(?, ?, 'gavel', 'headless-codex', 'local'),
			(?, ?, 'codex', '', 'local')`,
		orchestrationSessionID, providerSessionID.String(),
		transcriptSessionID, providerSessionID.String(),
	).Error)
	require.NoError(t, db.Exec(`
		INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin)
		VALUES (?, ?, ?, 'session-reference-test')`,
		promptRunID, orchestrationSessionID, orchestrationSessionID,
	).Error)
	_, err = repo.LinkPromptRun(ctx, issue.ID, promptRunID, native.StepPlan, 0, issue.Version, "test")
	require.NoError(t, err)

	for _, ref := range []string{
		providerSessionID.String(),
		orchestrationSessionID.String(),
		transcriptSessionID.String(),
	} {
		resolved, sessionID, err := repo.GetIssueBySessionRef(ctx, ref)
		require.NoError(t, err)
		assert.Equal(t, issue.ID, resolved.ID)
		assert.Equal(t, providerSessionID.String(), sessionID)
	}

	_, _, err = repo.GetIssueBySessionRef(ctx, uuid.NewString())
	require.ErrorIs(t, err, native.ErrNotFound)
	_, _, err = repo.GetIssueBySessionRef(ctx, "not-a-uuid")
	require.ErrorIs(t, err, native.ErrInvalidInput)
}

func TestRepositoryMoveIssueWorkspace(t *testing.T) {
	repo, db := openRepository(t)
	ctx := t.Context()
	sourceWorkspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/move-source"})
	require.NoError(t, err)
	targetWorkspace, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{RepoKey: "github.com/example/move-target"})
	require.NoError(t, err)

	issue, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: sourceWorkspace.ID,
		Aliases: []native.AliasInput{
			{Alias: "legacy-move-ref", Kind: "legacy"},
			{Alias: "jira-move-42", Kind: "external"},
		},
		Title: "Move with durable state",
		Actor: "creator",
	})
	require.NoError(t, err)
	promptRunID := insertCaptainPromptRun(t, db)
	_, err = repo.LinkPromptRun(ctx, issue.ID, promptRunID, native.StepRun, 0, issue.Version, "runner")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	issue, err = repo.SetActivePromptRun(ctx, issue.ID, &promptRunID, issue.Version, "runner")
	require.NoError(t, err)
	planID := insertCaptainPlan(t, db, promptRunID)
	_, err = repo.LinkPlan(ctx, issue.ID, planID, 0, issue.Version, "planner")
	require.NoError(t, err)
	issue, err = repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	issue, err = repo.SelectPlan(ctx, issue.ID, &planID, issue.Version, "planner")
	require.NoError(t, err)

	aliasesBefore, err := repo.ListAliases(ctx, issue.ID)
	require.NoError(t, err)
	eventsBefore, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	promptRunsBefore, err := repo.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	plansBefore, err := repo.ListPlans(ctx, issue.ID)
	require.NoError(t, err)
	versionBefore := issue.Version

	moved, err := repo.MoveIssueWorkspace(ctx, issue.ID, targetWorkspace.ID, issue.Version, "transfer-agent")
	require.NoError(t, err)
	assert.Equal(t, issue.ID, moved.ID)
	assert.Equal(t, targetWorkspace.ID, moved.WorkspaceID)
	assert.Equal(t, versionBefore+1, moved.Version)
	assert.Equal(t, issue.CreatedAt, moved.CreatedAt)
	assert.Equal(t, issue.ActivePromptRunID, moved.ActivePromptRunID)
	assert.Equal(t, issue.SelectedPlanID, moved.SelectedPlanID)

	aliasesAfter, err := repo.ListAliases(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, aliasesAfter, len(aliasesBefore))
	for index := range aliasesBefore {
		assert.Equal(t, aliasesBefore[index].Alias, aliasesAfter[index].Alias)
		assert.Equal(t, aliasesBefore[index].Kind, aliasesAfter[index].Kind)
		assert.Equal(t, aliasesBefore[index].IssueID, aliasesAfter[index].IssueID)
		assert.Equal(t, aliasesBefore[index].CreatedAt, aliasesAfter[index].CreatedAt)
		assert.Equal(t, targetWorkspace.ID, aliasesAfter[index].WorkspaceID)
	}
	promptRunsAfter, err := repo.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, promptRunsBefore, promptRunsAfter)
	plansAfter, err := repo.ListPlans(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, plansBefore, plansAfter)

	eventsAfter, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, eventsAfter, len(eventsBefore)+1)
	lastEvent := eventsAfter[len(eventsAfter)-1]
	assert.Equal(t, "workspace_moved", lastEvent.Kind)
	assert.Equal(t, "transfer-agent", lastEvent.Actor)
	assert.Equal(t, int64(len(eventsAfter)), lastEvent.Sequence)
	var movePayload map[string]string
	require.NoError(t, json.Unmarshal(lastEvent.Payload, &movePayload))
	assert.Equal(t, sourceWorkspace.ID.String(), movePayload["fromWorkspaceId"])
	assert.Equal(t, targetWorkspace.ID.String(), movePayload["toWorkspaceId"])

	sourceIssues, err := repo.ListIssues(ctx, sourceWorkspace.ID)
	require.NoError(t, err)
	assert.Empty(t, sourceIssues)
	targetIssues, err := repo.ListIssues(ctx, targetWorkspace.ID)
	require.NoError(t, err)
	require.Len(t, targetIssues, 1)
	assert.Equal(t, issue.ID, targetIssues[0].ID)
	_, err = repo.GetIssueByRef(ctx, sourceWorkspace.ID, "legacy-move-ref")
	require.ErrorIs(t, err, native.ErrNotFound)
	resolved, err := repo.GetIssueByRef(ctx, targetWorkspace.ID, "legacy-move-ref")
	require.NoError(t, err)
	assert.Equal(t, issue.ID, resolved.ID)

	_, err = repo.MoveIssueWorkspace(ctx, moved.ID, targetWorkspace.ID, moved.Version, "transfer-agent")
	require.ErrorIs(t, err, native.ErrNoChanges)
	_, err = repo.MoveIssueWorkspace(ctx, moved.ID, sourceWorkspace.ID, moved.Version-1, "transfer-agent")
	require.ErrorIs(t, err, native.ErrVersionConflict)
	_, err = repo.MoveIssueWorkspace(ctx, moved.ID, uuid.New(), moved.Version, "transfer-agent")
	require.ErrorIs(t, err, native.ErrNotFound)

	collisionOwner, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: targetWorkspace.ID,
		Aliases:     []native.AliasInput{{Alias: "collision-ref", Kind: "legacy"}},
		Title:       "Collision owner",
	})
	require.NoError(t, err)
	collisionMover, err := repo.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID: sourceWorkspace.ID,
		Aliases:     []native.AliasInput{{Alias: "collision-ref", Kind: "external"}},
		Title:       "Collision mover",
	})
	require.NoError(t, err)
	_, err = repo.MoveIssueWorkspace(ctx, collisionMover.ID, targetWorkspace.ID, collisionMover.Version, "transfer-agent")
	require.ErrorIs(t, err, native.ErrAliasConflict)
	unchangedCollisionMover, err := repo.GetIssue(ctx, collisionMover.ID)
	require.NoError(t, err)
	assert.Equal(t, sourceWorkspace.ID, unchangedCollisionMover.WorkspaceID)
	assert.Equal(t, collisionMover.Version, unchangedCollisionMover.Version)
	resolved, err = repo.GetIssueByRef(ctx, targetWorkspace.ID, "collision-ref")
	require.NoError(t, err)
	assert.Equal(t, collisionOwner.ID, resolved.ID)
	resolved, err = repo.GetIssueByRef(ctx, sourceWorkspace.ID, "collision-ref")
	require.NoError(t, err)
	assert.Equal(t, collisionMover.ID, resolved.ID)

	relationshipMover := createIssue(t, repo, sourceWorkspace.ID, "Relationship mover")
	dependent := createIssue(t, repo, sourceWorkspace.ID, "Relationship dependent")
	_, err = repo.AddRelationship(ctx, dependent.ID, relationshipMover.ID, native.RelationshipDependsOn, dependent.Version, "linker")
	require.NoError(t, err)
	_, err = repo.MoveIssueWorkspace(ctx, relationshipMover.ID, targetWorkspace.ID, relationshipMover.Version, "transfer-agent")
	require.ErrorIs(t, err, native.ErrIssueHasRelationships)
	unchangedRelationshipMover, err := repo.GetIssue(ctx, relationshipMover.ID)
	require.NoError(t, err)
	assert.Equal(t, sourceWorkspace.ID, unchangedRelationshipMover.WorkspaceID)
	assert.Equal(t, relationshipMover.Version, unchangedRelationshipMover.Version)

	finalMoved, err := repo.GetIssue(ctx, moved.ID)
	require.NoError(t, err)
	assert.Equal(t, moved.Version, finalMoved.Version, "failed move attempts must not mutate the issue")
	finalEvents, err := repo.ListEvents(ctx, moved.ID)
	require.NoError(t, err)
	assert.Len(t, finalEvents, len(eventsAfter), "failed move attempts must not append events")
}
