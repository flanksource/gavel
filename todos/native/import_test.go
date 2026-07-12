package native_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	importIssueA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	importIssueB = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func TestApplyImportIsAtomicIdempotentAndPreservesLegacyState(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	batch := native.ImportBatch{
		Workspace: native.ImportWorkspace{
			RepoKey:     " GitHub.com/Flanksource/Import-Test ",
			RootPath:    "/workspace/legacy-import",
			DisplayName: "Legacy import",
			CreatedAt:   base.Add(-time.Hour),
			UpdatedAt:   base.Add(10 * time.Minute),
		},
		Issues: []native.ImportIssue{
			{
				SourceID:       importIssueB,
				Title:          "Dependency",
				Body:           "Dependency body",
				Labels:         []string{" TODOS ", "database"},
				Priority:       native.PriorityMedium,
				Status:         native.StatusDraft,
				ExecutionState: native.ExecutionIdle,
				CreatedAt:      base.Add(2 * time.Minute),
				UpdatedAt:      base.Add(3 * time.Minute),
			},
			{
				SourceID:       importIssueA,
				Title:          "Imported issue",
				Body:           "Description\n\n## Verification\n\n```exec\ntrue\n```",
				Verification:   "```exec\ntrue\n```",
				Labels:         []string{" Todos ", "DATABASE", "todos"},
				Priority:       native.PriorityHigh,
				Status:         native.StatusOpen,
				ExecutionState: native.ExecutionIdle,
				CreatedAt:      base,
				UpdatedAt:      base.Add(8 * time.Minute),
			},
		},
		Events: []native.ImportEvent{
			{
				IssueSourceID: importIssueA,
				SourceID:      "event-a-comment",
				Kind:          "comment",
				Actor:         "legacy-user",
				Body:          "checkpoint",
				Payload:       json.RawMessage(`{"body":"checkpoint","n":2}`),
				CreatedAt:     base.Add(6 * time.Minute),
			},
			{
				IssueSourceID: importIssueB,
				SourceID:      "event-b-created",
				Kind:          "created",
				Payload:       json.RawMessage(`{"title":"Dependency"}`),
				CreatedAt:     base.Add(2 * time.Minute),
			},
			{
				IssueSourceID: importIssueA,
				SourceID:      "event-a-created",
				Kind:          "created",
				Payload:       json.RawMessage(`{"title":"Imported issue"}`),
				CreatedAt:     base,
			},
		},
		Relationships: []native.ImportRelationship{{
			IssueSourceID:       importIssueA,
			TargetIssueSourceID: importIssueB,
			Relation:            native.RelationshipDependsOn,
			CreatedAt:           base.Add(7 * time.Minute),
		}},
		Warnings: []native.ImportWarning{{
			IssueSourceID: importIssueA,
			Code:          "missing_plan_file",
			Message:       "legacy plan file is missing",
			Payload:       json.RawMessage(`{"path":"/missing/plan.md"}`),
			CreatedAt:     base.Add(8 * time.Minute),
		}},
	}

	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	assert.Equal(t, 1, first.Counts.WorkspaceCreated)
	assert.Equal(t, 2, first.Counts.IssuesCreated)
	assert.Equal(t, 2, first.Counts.AliasesInserted)
	assert.Equal(t, 4, first.Counts.EventsInserted)
	assert.Equal(t, 1, first.Counts.WarningsInserted)
	assert.Equal(t, 1, first.Counts.RelationshipsInserted)
	require.NotEmpty(t, first.Checksum)

	byFull, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	byShort, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA[:8])
	require.NoError(t, err)
	assert.Equal(t, byFull.ID, byShort.ID)
	assert.Equal(t, "Imported issue", byFull.Title)
	assert.Equal(t, "Description\n\n## Verification\n\n```exec\ntrue\n```", byFull.Body)
	assert.Equal(t, "```exec\ntrue\n```", byFull.Verification)
	assert.Equal(t, []string{"database", "todos"}, byFull.Labels)
	assert.Equal(t, native.StatusOpen, byFull.Status)
	assert.True(t, byFull.CreatedAt.Equal(base))
	assert.True(t, byFull.UpdatedAt.Equal(base.Add(8*time.Minute)))
	assert.EqualValues(t, 3, byFull.Version, "two source events plus one warning advance version")

	events, err := repo.ListEvents(ctx, byFull.ID)
	require.NoError(t, err)
	require.Len(t, events, 3, "the relationship import must not synthesize another event")
	assert.Equal(t, []int64{1, 2, 3}, []int64{events[0].Sequence, events[1].Sequence, events[2].Sequence})
	assert.Equal(t, "event-a-created", events[0].SourceID)
	assert.Equal(t, "event-a-comment", events[1].SourceID)
	assert.Equal(t, "migration_warning", events[2].Kind)
	assert.True(t, events[0].CreatedAt.Equal(base))

	relationships, err := repo.ListRelationships(ctx, byFull.ID)
	require.NoError(t, err)
	require.Len(t, relationships, 1)
	assert.Equal(t, native.RelationshipDependsOn, relationships[0].Relation)

	second, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	assert.Zero(t, second.Counts.WorkspaceCreated)
	assert.Zero(t, second.Counts.WorkspaceUpdated)
	assert.Zero(t, second.Counts.IssuesCreated)
	assert.Zero(t, second.Counts.IssuesUpdated)
	assert.Zero(t, second.Counts.AliasesInserted)
	assert.Zero(t, second.Counts.EventsInserted)
	assert.Equal(t, 4, second.Counts.EventsReplayed)
	assert.Equal(t, 1, second.Counts.WarningsReplayed)
	assert.Zero(t, second.Counts.RelationshipsInserted)
	assert.Equal(t, 1, second.Counts.RelationshipsReplayed)
	assert.Equal(t, first.Checksum, second.Checksum)

	afterReplay, err := repo.GetIssue(ctx, byFull.ID)
	require.NoError(t, err)
	assert.Equal(t, byFull.Version, afterReplay.Version)
	assert.True(t, afterReplay.CreatedAt.Equal(base))
	assert.True(t, afterReplay.UpdatedAt.Equal(base.Add(8*time.Minute)))

	removed := batch
	removed.Relationships = nil
	removed.RelationshipDeletes = append([]native.ImportRelationship(nil), batch.Relationships...)
	third, err := repo.ApplyImport(ctx, removed)
	require.NoError(t, err)
	assert.Equal(t, 1, third.Counts.RelationshipsDeleted)
	assert.Len(t, third.Rollback.DeletedRelationships, 1)
	assert.NotEqual(t, second.Checksum, third.Checksum)
	relationships, err = repo.ListRelationships(ctx, byFull.ID)
	require.NoError(t, err)
	assert.Empty(t, relationships)

	fourth, err := repo.ApplyImport(ctx, removed)
	require.NoError(t, err)
	assert.Zero(t, fourth.Counts.RelationshipsDeleted)
	assert.Equal(t, 1, fourth.Counts.RelationshipDeletesReplayed)
	assert.Equal(t, third.Checksum, fourth.Checksum)
}

func TestApplyImportEventConflictRollsBackEarlierBatchMutations(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.February, 3, 4, 5, 6, 0, time.UTC)
	batch := singleIssueImportBatch(base, "Original title")
	batch.Events = []native.ImportEvent{{
		IssueSourceID: importIssueA,
		SourceID:      "event-original",
		Kind:          "created",
		Body:          "original body",
		Payload:       json.RawMessage(`{"value":"original"}`),
		CreatedAt:     base.Add(2 * time.Minute),
	}}
	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	issue, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	require.EqualValues(t, 1, issue.Version)

	conflict := singleIssueImportBatch(base, "Must roll back")
	conflict.Events = []native.ImportEvent{
		{
			IssueSourceID: importIssueA,
			SourceID:      "event-new-before-conflict",
			Kind:          "comment",
			Body:          "this insert must roll back",
			CreatedAt:     base.Add(time.Minute),
		},
		{
			IssueSourceID: importIssueA,
			SourceID:      "event-original",
			Kind:          "created",
			Body:          "different body",
			Payload:       json.RawMessage(`{"value":"different"}`),
			CreatedAt:     base.Add(2 * time.Minute),
		},
	}
	_, err = repo.ApplyImport(ctx, conflict)
	require.ErrorIs(t, err, native.ErrEventConflict)

	after, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, "Original title", after.Title)
	assert.Equal(t, issue.Version, after.Version)
	events, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "event-original", events[0].SourceID)
}

func TestApplyImportRejectsExpectedChecksumDriftBeforeMutation(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.April, 5, 6, 7, 8, 0, time.UTC)
	batch := singleIssueImportBatch(base, "Imported title")
	batch.Fingerprint = strings.Repeat("a", 64)

	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	issue, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	driftTitle := "Concurrent native edit"
	drifted, err := repo.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{
		Title: &driftTitle,
		Actor: "drift-test",
	})
	require.NoError(t, err)
	eventsBefore, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)

	retry := batch
	retry.ExpectedChecksum = first.Checksum
	_, err = repo.ApplyImport(ctx, retry)
	require.ErrorIs(t, err, native.ErrImportConflict)
	require.ErrorContains(t, err, "workspace checksum drifted")

	after, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, drifted.Title, after.Title, "drift rejection must not overwrite the concurrent edit")
	assert.Equal(t, drifted.Version, after.Version)
	eventsAfter, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, eventsBefore, eventsAfter)
}

func TestApplyImportWithOmittedIDReusesExistingRandomWorkspace(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	existing, err := repo.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey:  "github.com/flanksource/import-existing-random",
		RootPath: "/workspace/import-existing-random",
	})
	require.NoError(t, err)
	assert.NotEqual(t,
		native.DeterministicImportWorkspaceID(native.ImportWorkspace{RepoKey: existing.RepoKey, RootPath: existing.RootPath}),
		existing.ID,
		"repository-created workspace should exercise non-derived ID reuse",
	)

	batch := relationshipImportBatch(time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC))
	batch.Workspace = native.ImportWorkspace{
		RepoKey:  " GitHub.COM/Flanksource/Import-Existing-Random ",
		RootPath: "/workspace/./import-existing-random",
	}
	result, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	assert.Equal(t, existing.ID, result.WorkspaceID)
}

func TestApplyImportExactReadbackRejectsUnexpectedSourceEvent(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.May, 6, 7, 8, 9, 0, time.UTC)
	batch := singleIssueImportBatch(base, "Exact target")

	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	issue, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	_, err = repo.AppendEvent(ctx, issue.ID, issue.Version, native.EventInput{
		Kind:     "comment",
		Actor:    "target-drift",
		Body:     "unexpected import-owned event",
		Source:   native.DefaultImportSource,
		SourceID: "rogue-source-event",
	})
	require.NoError(t, err)
	drifted, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)

	_, err = repo.ApplyImport(ctx, batch)
	require.ErrorIs(t, err, native.ErrImportConflict)
	require.ErrorContains(t, err, "unexpected source event rogue-source-event")

	after, err := repo.GetIssue(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, drifted, after, "readback failure must roll back attempted snapshot rewrites")
	events, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "rogue-source-event", events[0].SourceID)
}

func TestApplyImportRejectsReorderedEqualTimestampSourceEvents(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.May, 7, 8, 9, 10, 0, time.UTC)
	batch := singleIssueImportBatch(base, "Ordered source history")
	eventTime := base.Add(time.Minute)
	batch.Events = []native.ImportEvent{
		{
			IssueSourceID: importIssueA, SourceID: "event-first", Order: 0,
			Kind: "comment", Body: "first", Payload: json.RawMessage(`{"body":"first"}`), CreatedAt: eventTime,
		},
		{
			IssueSourceID: importIssueA, SourceID: "event-second", Order: 1,
			Kind: "comment", Body: "second", Payload: json.RawMessage(`{"body":"second"}`), CreatedAt: eventTime,
		},
	}

	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	issue, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	events, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, events, 2)
	assert.Equal(t, []string{"event-first", "event-second"}, []string{events[0].SourceID, events[1].SourceID})

	reordered := batch
	reordered.ExpectedChecksum = first.Checksum
	reordered.Events = append([]native.ImportEvent(nil), batch.Events...)
	reordered.Events[0], reordered.Events[1] = reordered.Events[1], reordered.Events[0]
	reordered.Events[0].Order = 0
	reordered.Events[1].Order = 1
	_, err = repo.ApplyImport(ctx, reordered)
	require.ErrorIs(t, err, native.ErrImportConflict)
	require.ErrorContains(t, err, "source event order")

	after, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, events, after, "order conflict must not mutate the persisted history")
}

func TestApplyImportFinalRelationshipRemovalReplaysWithHistoricalAuditEvents(t *testing.T) {
	repo, _ := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.June, 7, 8, 9, 10, 0, time.UTC)
	initial := relationshipImportBatch(base)
	initial.Fingerprint = strings.Repeat("a", 64)
	initial.Warnings = []native.ImportWarning{{
		IssueSourceID: importIssueA,
		SourceID:      "warning:" + strings.Repeat("c", 64),
		Code:          "initial_only_warning",
		Message:       "warning present in the initial snapshot only",
		CreatedAt:     base.Add(time.Minute),
	}}

	first, err := repo.ApplyImport(ctx, initial)
	require.NoError(t, err)
	assert.Equal(t, 1, first.Counts.RelationshipsInserted)
	assert.Equal(t, 5, first.Counts.EventsInserted, "two source events, one warning, and two checkpoints")

	final := initial
	final.Fingerprint = strings.Repeat("b", 64)
	final.ExpectedChecksum = first.Checksum
	final.Relationships = nil
	final.RelationshipDeletes = append([]native.ImportRelationship(nil), initial.Relationships...)
	final.Warnings = nil
	second, err := repo.ApplyImport(ctx, final)
	require.NoError(t, err)
	assert.Equal(t, 1, second.Counts.RelationshipsDeleted)
	assert.Len(t, second.Rollback.DeletedRelationships, 1)
	assert.Equal(t, 2, second.Counts.EventsInserted, "the final fingerprint adds one checkpoint per issue")
	assert.Equal(t, 2, second.Counts.EventsReplayed, "immutable source events replay exactly")
	assert.NotEqual(t, first.Checksum, second.Checksum)

	issueA, err := repo.GetIssueByRef(ctx, second.WorkspaceID, importIssueA)
	require.NoError(t, err)
	relationships, err := repo.ListRelationships(ctx, issueA.ID)
	require.NoError(t, err)
	assert.Empty(t, relationships)
	events, err := repo.ListEvents(ctx, issueA.ID)
	require.NoError(t, err)
	require.Len(t, events, 4)
	assert.Equal(t, []string{
		"event-a-created",
		"warning:" + strings.Repeat("c", 64),
		"checkpoint:" + strings.Repeat("a", 64) + ":" + importIssueA,
		"checkpoint:" + strings.Repeat("b", 64) + ":" + importIssueA,
	}, []string{events[0].SourceID, events[1].SourceID, events[2].SourceID, events[3].SourceID})
	for index, fingerprint := range []string{strings.Repeat("a", 64), strings.Repeat("b", 64)} {
		checkpoint := events[index+2]
		assert.Equal(t, "migration_checkpoint", checkpoint.Kind)
		assert.Equal(t, native.DefaultImportSource, checkpoint.Actor)
		var payload map[string]string
		require.NoError(t, json.Unmarshal(checkpoint.Payload, &payload))
		assert.Equal(t, fingerprint, payload["fingerprint"])
	}

	replay := final
	replay.ExpectedChecksum = second.Checksum
	third, err := repo.ApplyImport(ctx, replay)
	require.NoError(t, err)
	assert.Zero(t, third.Counts.EventsInserted)
	assert.Equal(t, 4, third.Counts.EventsReplayed)
	assert.Zero(t, third.Counts.RelationshipsDeleted)
	assert.Equal(t, 1, third.Counts.RelationshipDeletesReplayed)
	assert.Equal(t, second.Checksum, third.Checksum)
}

func TestApplyImportMaterializesResolvedCaptainLinksWithExactReplay(t *testing.T) {
	repo, db := openRepository(t)
	ctx := t.Context()
	base := time.Date(2025, time.March, 4, 5, 6, 7, 0, time.UTC)
	runID := insertCaptainPromptRun(t, db)
	planID := insertCaptainPlan(t, db, runID)
	batch := singleIssueImportBatch(base, "Linked issue")
	batch.Issues[0].ActivePromptRunID = &runID
	batch.Issues[0].SelectedPlanID = &planID
	batch.PromptRunLinks = []native.ImportPromptRunLink{{
		IssueSourceID: importIssueA,
		PromptRunID:   runID,
		StepKind:      native.StepRun,
		Ordinal:       0,
		CreatedAt:     base.Add(time.Minute),
	}}
	batch.PlanLinks = []native.ImportPlanLink{{
		IssueSourceID: importIssueA,
		PlanID:        planID,
		Ordinal:       0,
		CreatedAt:     base.Add(time.Minute),
	}}

	first, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	assert.Equal(t, 1, first.Counts.PromptRunLinksInserted)
	assert.Equal(t, 1, first.Counts.PlanLinksInserted)
	assert.Equal(t, 1, first.Counts.ProjectionEventsInserted)
	require.Len(t, first.Rollback.InsertedEvents, 1)
	assert.Equal(t, "captain-projection", first.Rollback.InsertedEvents[0].Source)
	issue, err := repo.GetIssueByRef(ctx, first.WorkspaceID, importIssueA)
	require.NoError(t, err)
	require.NotNil(t, issue.ActivePromptRunID)
	require.NotNil(t, issue.SelectedPlanID)
	assert.Equal(t, runID, *issue.ActivePromptRunID)
	assert.Equal(t, planID, *issue.SelectedPlanID)
	assert.Equal(t, native.ExecutionRunning, issue.ExecutionState, "active Captain state must be projected during import")
	assert.EqualValues(t, 1, issue.Version, "only Captain's authoritative projection may add an event")
	firstEvents, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, firstEvents, 1)

	second, err := repo.ApplyImport(ctx, batch)
	require.NoError(t, err)
	assert.Zero(t, second.Counts.IssuesUpdated)
	assert.Zero(t, second.Counts.ProjectionEventsInserted)
	assert.Equal(t, 1, second.Counts.PromptRunLinksReplayed)
	assert.Equal(t, 1, second.Counts.PlanLinksReplayed)
	assert.Equal(t, first.Checksum, second.Checksum)
	secondEvents, err := repo.ListEvents(ctx, issue.ID)
	require.NoError(t, err)
	assert.Len(t, secondEvents, len(firstEvents), "exact replay must not manufacture another projection event")

	conflictingRunID := insertCaptainPromptRun(t, db)
	conflict := singleIssueImportBatch(base, "Linked issue")
	conflict.Issues[0].ActivePromptRunID = &runID
	conflict.Issues[0].SelectedPlanID = &planID
	conflict.PromptRunLinks = []native.ImportPromptRunLink{{
		IssueSourceID: importIssueA,
		PromptRunID:   conflictingRunID,
		StepKind:      native.StepRun,
		Ordinal:       0,
		CreatedAt:     base.Add(2 * time.Minute),
	}}
	_, err = repo.ApplyImport(ctx, conflict)
	require.ErrorIs(t, err, native.ErrLinkConflict)
	links, err := repo.ListPromptRuns(ctx, issue.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, runID, links[0].PromptRunID)
}

func singleIssueImportBatch(base time.Time, title string) native.ImportBatch {
	return native.ImportBatch{
		Workspace: native.ImportWorkspace{
			RepoKey:     "github.com/flanksource/import-single",
			RootPath:    "/workspace/import-single",
			DisplayName: "Import single",
			CreatedAt:   base.Add(-time.Hour),
			UpdatedAt:   base,
		},
		Issues: []native.ImportIssue{{
			ID:             uuid.Nil,
			SourceID:       importIssueA,
			Title:          title,
			Body:           "legacy body",
			Verification:   "```exec\ntrue\n```",
			Labels:         []string{"database", "todos"},
			Priority:       native.PriorityHigh,
			Status:         native.StatusOpen,
			ExecutionState: native.ExecutionIdle,
			CreatedAt:      base,
			UpdatedAt:      base.Add(5 * time.Minute),
		}},
	}
}

func relationshipImportBatch(base time.Time) native.ImportBatch {
	return native.ImportBatch{
		Workspace: native.ImportWorkspace{
			RepoKey:     "github.com/flanksource/import-relationship-final",
			RootPath:    "/workspace/import-relationship-final",
			DisplayName: "Relationship final import",
			CreatedAt:   base.Add(-time.Hour),
			UpdatedAt:   base.Add(10 * time.Minute),
		},
		Issues: []native.ImportIssue{
			{
				SourceID: importIssueA, Title: "Dependent", Body: "body",
				Priority: native.PriorityHigh, Status: native.StatusOpen, ExecutionState: native.ExecutionIdle,
				CreatedAt: base, UpdatedAt: base.Add(5 * time.Minute),
			},
			{
				SourceID: importIssueB, Title: "Dependency", Body: "body",
				Priority: native.PriorityMedium, Status: native.StatusOpen, ExecutionState: native.ExecutionIdle,
				CreatedAt: base.Add(time.Minute), UpdatedAt: base.Add(6 * time.Minute),
			},
		},
		Events: []native.ImportEvent{
			{IssueSourceID: importIssueA, SourceID: "event-a-created", Kind: "created", CreatedAt: base},
			{IssueSourceID: importIssueB, SourceID: "event-b-created", Kind: "created", CreatedAt: base.Add(time.Minute)},
		},
		Relationships: []native.ImportRelationship{{
			IssueSourceID: importIssueA, TargetIssueSourceID: importIssueB,
			Relation: native.RelationshipDependsOn, CreatedAt: base.Add(2 * time.Minute),
		}},
	}
}
