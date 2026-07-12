package migrategrite_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/migrategrite"
	"github.com/flanksource/gavel/todos/native"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	artifactAcceptanceIssueA = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	artifactAcceptanceIssueB = "ffffffffffffffffffffffffffffffff"
)

func TestArtifactServiceAcceptanceInitialToFrozenFinal(t *testing.T) {
	service, repository, captain, gormDB := openMigrationService(t)
	ctx := t.Context()
	planRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(planRoot, "plans"), 0o700))

	session, _ := createCaptainRun(t, captain, "captain-artifact-acceptance")
	planAPath := filepath.Join("plans", "artifact-a.md")
	planBPath := filepath.Join("plans", "artifact-b.md")
	planAMarkdown := "# Plan A\n\n1. Establish the initial relationship.\n"
	planBMarkdown := "# Plan B\n\n1. Remove the obsolete relationship.\n"
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, planAPath), []byte(planAMarkdown), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(planRoot, planBPath), []byte(planBMarkdown), 0o600))
	planA, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: session.ID, Path: planAPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)
	planB, err := captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
		SourceSessionID: session.ID, Path: planBPath, Variant: "legacy-authoritative",
	})
	require.NoError(t, err)

	initial, delta, probe := artifactAcceptanceSnapshots(t, session.ProviderSessionID, planAPath, planBPath)
	require.NoError(t, migrategrite.ValidateFullSnapshotHistory(initial))
	initialRaw := marshalArtifactAcceptanceSnapshot(t, initial)
	deltaRaw := marshalArtifactAcceptanceSnapshot(t, delta)
	probeRaw := marshalArtifactAcceptanceSnapshot(t, probe)

	workspace := native.ImportWorkspace{
		RepoKey:     "github.com/flanksource/gavel-migrategrite-artifact-acceptance",
		RootPath:    planRoot,
		DisplayName: "artifact acceptance",
	}
	artifactRoot := filepath.Join(t.TempDir(), "artifacts")
	initialDir := filepath.Join(artifactRoot, "initial")
	initialPending, err := migrategrite.PrepareArtifactGeneration(initialDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(initialPending, "source-initial.json", initialRaw))
	initialCallbackCalled := false
	initialReport, err := service.Import(ctx, initial, migrategrite.ImportOptions{
		Workspace:             workspace,
		PlanRoot:              planRoot,
		DeferActivePromptRuns: true,
		RequireNoPriorImport:  true,
		BeforeCommit: func(report *migrategrite.ImportReport) error {
			initialCallbackCalled = true
			return migrategrite.WriteInitialArtifacts(initialPending, initialRaw, initial, report, migrategrite.ArtifactProvenance{
				SourceMode: "file", SourceRoot: planRoot,
			})
		},
	})
	require.NoError(t, err)
	require.True(t, initialCallbackCalled)
	require.NoError(t, migrategrite.PublishArtifactGeneration(initialPending, initialDir))
	require.NoError(t, migrategrite.VerifyArtifactChecksums(initialDir))
	assert.Equal(t, 2, initialReport.Counts.IssuesCreated)
	assert.Equal(t, 1, initialReport.Counts.RelationshipsInserted)
	assert.Equal(t, 1, initialReport.Counts.PlanLinksInserted)
	assert.Len(t, initialReport.Rollback.AppendedCaptainRevisionIDs, 1)

	initialRollback, err := migrategrite.LoadRollback(initialDir)
	require.NoError(t, err)
	assert.Equal(t, initialReport.Rollback, initialRollback)

	merged, err := migrategrite.MergeSnapshots(initial, delta)
	require.NoError(t, err)
	require.NoError(t, migrategrite.ValidateFullSnapshotHistory(merged))
	require.Greater(t, probe.Meta.GeneratedTS, delta.Meta.GeneratedTS)
	require.NoError(t, migrategrite.ValidateFrozenProbe(merged, probe))
	finalDocument, err := migrategrite.Normalize(merged)
	require.NoError(t, err)

	finalDir := filepath.Join(artifactRoot, "final")
	finalPending, err := migrategrite.PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(finalPending, "source-initial.json", initialRaw))
	require.NoError(t, migrategrite.WritePreparedSource(finalPending, "source-final-delta.json", deltaRaw))
	require.NoError(t, migrategrite.WritePreparedSource(finalPending, "source-final-probe.json", probeRaw))
	workspace.ID = initialReport.WorkspaceID
	var combinedRollback migrategrite.Rollback
	finalCallbackCalled := false
	finalReport, err := service.Import(ctx, merged, migrategrite.ImportOptions{
		Workspace:              workspace,
		PlanRoot:               planRoot,
		ExpectedTargetChecksum: initialReport.Validation.TargetChecksum,
		BeforeCommit: func(report *migrategrite.ImportReport) error {
			finalCallbackCalled = true
			combinedRollback = migrategrite.MergeRollback(initialRollback, report.Rollback)
			return migrategrite.WriteFinalArtifacts(
				finalPending, initialDir, deltaRaw, delta, report, combinedRollback,
				migrategrite.ArtifactProvenance{
					SourceMode: "file", SourceRoot: planRoot,
					CursorMS: initialReport.Watermark.SinceMS(), FrozenValidated: true,
					ProbeRawHash: migrategrite.SHA256Bytes(probeRaw),
				},
			)
		},
	})
	require.NoError(t, err)
	require.True(t, finalCallbackCalled)
	require.NoError(t, migrategrite.PublishArtifactGeneration(finalPending, finalDir))
	require.NoError(t, migrategrite.VerifyArtifactChecksums(finalDir))
	assert.Equal(t, 1, finalReport.Counts.RelationshipsDeleted)
	assert.Equal(t, 1, finalReport.Counts.PlanLinksInserted)
	assert.Equal(t, 1, finalReport.Counts.PlanLinksReplayed)
	assert.Len(t, finalReport.Rollback.AppendedCaptainRevisionIDs, 1)

	issueA, err := repository.GetIssueByRef(ctx, finalReport.WorkspaceID, artifactAcceptanceIssueA)
	require.NoError(t, err)
	require.NotNil(t, issueA.SelectedPlanID)
	assert.Equal(t, planB.ID, *issueA.SelectedPlanID)
	planLinks, err := repository.ListPlans(ctx, issueA.ID)
	require.NoError(t, err)
	require.Len(t, planLinks, 2)
	assert.Equal(t, []string{planA.ID.String(), planB.ID.String()}, []string{planLinks[0].PlanID.String(), planLinks[1].PlanID.String()})
	relationships, err := repository.ListRelationships(ctx, issueA.ID)
	require.NoError(t, err)
	assert.Empty(t, relationships)
	planARevisions, err := captain.ListPlanRevisions(ctx, planA.ID)
	require.NoError(t, err)
	require.Len(t, planARevisions, 1)
	assert.Equal(t, "# Plan A\n\n1. Establish the initial relationship.", planARevisions[0].PlanMarkdown)
	planBRevisions, err := captain.ListPlanRevisions(ctx, planB.ID)
	require.NoError(t, err)
	require.Len(t, planBRevisions, 1)
	assert.Equal(t, "# Plan B\n\n1. Remove the obsolete relationship.", planBRevisions[0].PlanMarkdown)

	storedRollback, err := migrategrite.LoadRollback(finalDir)
	require.NoError(t, err)
	combinedRollbackJSON, err := json.Marshal(combinedRollback)
	require.NoError(t, err)
	storedRollbackJSON, err := json.Marshal(storedRollback)
	require.NoError(t, err)
	assert.JSONEq(t, string(combinedRollbackJSON), string(storedRollbackJSON))
	assert.Len(t, storedRollback.Native.CreatedIssueIDs, 2)
	assert.Len(t, storedRollback.Native.InsertedPlanLinks, 2)
	assert.Empty(t, storedRollback.Native.InsertedRelationships, "the initial insert and final delete cancel in the combined inverse")
	assert.Empty(t, storedRollback.Native.DeletedRelationships, "the initial insert and final delete cancel in the combined inverse")
	assert.Len(t, storedRollback.AppendedCaptainRevisionIDs, 2)

	manifest, err := migrategrite.LoadArtifactManifest(finalDir)
	require.NoError(t, err)
	assert.Equal(t, "final", manifest.State)
	assert.True(t, manifest.Frozen)
	assert.Equal(t, initialReport.WorkspaceID, manifest.WorkspaceID)
	assert.Equal(t, finalDocument.SourceHash, manifest.FinalSourceHash)
	assert.Equal(t, migrategrite.SHA256Bytes(deltaRaw), manifest.FinalRawHash)
	assert.Equal(t, migrategrite.SHA256Bytes(probeRaw), manifest.FinalProbeRawHash)
	assertArtifactAcceptanceFile(t, finalDir, "source-initial.json", initialRaw)
	assertArtifactAcceptanceFile(t, finalDir, "source-final-delta.json", deltaRaw)
	assertArtifactAcceptanceFile(t, finalDir, "source-final-probe.json", probeRaw)
	storedFinalReport, err := migrategrite.LoadImportReport(finalDir, "final")
	require.NoError(t, err)
	assert.Equal(t, finalReport.Validation, storedFinalReport.Validation)
	assert.Equal(t, finalReport.WorkspaceID, storedFinalReport.WorkspaceID)

	beforeReplay := readServiceCounts(t, gormDB)
	replay, err := service.Import(ctx, merged, migrategrite.ImportOptions{
		Workspace: workspace, PlanRoot: planRoot, ExpectedTargetChecksum: finalReport.Validation.TargetChecksum,
	})
	require.NoError(t, err)
	assert.Zero(t, replay.Counts.WorkspaceCreated)
	assert.Zero(t, replay.Counts.WorkspaceUpdated)
	assert.Zero(t, replay.Counts.IssuesCreated)
	assert.Zero(t, replay.Counts.IssuesUpdated)
	assert.Zero(t, replay.Counts.EventsInserted)
	assert.Zero(t, replay.Counts.ProjectionEventsInserted)
	assert.Zero(t, replay.Counts.RelationshipsInserted)
	assert.Zero(t, replay.Counts.RelationshipsDeleted)
	assert.Equal(t, 1, replay.Counts.RelationshipDeletesReplayed)
	assert.Zero(t, replay.Counts.PlanLinksInserted)
	assert.Equal(t, 2, replay.Counts.PlanLinksReplayed)
	assert.Equal(t, finalReport.Validation.TargetChecksum, replay.Validation.TargetChecksum)
	assert.Equal(t, finalReport.Validation.CaptainChecksum, replay.Validation.CaptainChecksum)
	assert.Empty(t, replay.Rollback.AppendedCaptainRevisionIDs)
	assert.Equal(t, beforeReplay, readServiceCounts(t, gormDB))
	require.NoError(t, migrategrite.VerifyArtifactChecksums(finalDir))
}

func TestConcurrentInitialImportsAcrossArtifactRootsKeepOneRollbackBaseline(t *testing.T) {
	service, _, _, _ := openMigrationService(t)
	ctx := t.Context()
	events := []griteexport.Event{
		serviceEvent(t, "concurrent-created", artifactAcceptanceIssueA, 100, "IssueCreated", map[string]any{
			"title": "Concurrent baseline", "body": "body",
		}),
	}
	snapshot := griteexport.Snapshot{
		Meta: griteexport.Meta{SchemaVersion: 1, GeneratedTS: 200, EventCount: len(events)},
		Issues: []griteexport.Issue{{
			IssueID: griteexport.ID(artifactAcceptanceIssueA), Title: "Concurrent baseline", State: "open",
			Labels: []string{"status:pending"}, CreatedTS: 100, UpdatedTS: 100,
		}},
		Events: events,
	}
	document, err := migrategrite.Normalize(snapshot)
	require.NoError(t, err)
	raw := marshalArtifactAcceptanceSnapshot(t, snapshot)
	workspace := native.NormalizeImportWorkspace(native.ImportWorkspace{
		RepoKey:  " GitHub.COM/Flanksource/Gavel-Concurrent-Baseline ",
		RootPath: t.TempDir(), DisplayName: " concurrent baseline ",
	})
	workspace.ID = native.DeterministicImportWorkspaceID(workspace)
	roots := []string{t.TempDir(), t.TempDir()}
	targets := make([]string, len(roots))
	pendings := make([]string, len(roots))
	for i, root := range roots {
		targets[i] = migrategrite.ArtifactDirectory(root, workspace.ID.String(), migrategrite.ArtifactRunID(snapshot, document.SourceHash))
		pendings[i], err = migrategrite.PrepareArtifactGeneration(targets[i])
		require.NoError(t, err)
		require.NoError(t, migrategrite.WritePreparedSource(pendings[i], "source-initial.json", raw))
	}

	type importResult struct {
		index  int
		report *migrategrite.ImportReport
		err    error
	}
	start := make(chan struct{})
	results := make(chan importResult, len(roots))
	var wait sync.WaitGroup
	for i := range roots {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			report, importErr := service.Import(ctx, snapshot, migrategrite.ImportOptions{
				Workspace: workspace, RequireNoPriorImport: true,
				BeforeCommit: func(report *migrategrite.ImportReport) error {
					return migrategrite.WriteInitialArtifacts(pendings[index], raw, snapshot, report, migrategrite.ArtifactProvenance{
						SourceMode: "file", SourceRoot: workspace.RootPath, RepoKey: workspace.RepoKey,
					})
				},
			})
			if importErr == nil {
				importErr = migrategrite.PublishArtifactGeneration(pendings[index], targets[index])
			} else {
				_ = migrategrite.DiscardArtifactGeneration(pendings[index])
			}
			results <- importResult{index: index, report: report, err: importErr}
		}(i)
	}
	close(start)
	wait.Wait()
	close(results)

	winner := -1
	loser := -1
	for result := range results {
		if result.err == nil {
			require.Equal(t, -1, winner, "exactly one import may establish the baseline")
			winner = result.index
			require.NotNil(t, result.report)
			continue
		}
		require.ErrorIs(t, result.err, native.ErrImportConflict)
		loser = result.index
	}
	require.NotEqual(t, -1, winner)
	require.NotEqual(t, -1, loser)
	require.NoError(t, migrategrite.VerifyArtifactChecksums(targets[winner]))
	rollback, err := migrategrite.LoadRollback(targets[winner])
	require.NoError(t, err)
	assert.Len(t, rollback.Native.CreatedIssueIDs, 1, "the only committed baseline retains the authoritative inverse")
	_, err = os.Lstat(targets[loser])
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Lstat(pendings[loser])
	require.ErrorIs(t, err, os.ErrNotExist)
}

func artifactAcceptanceSnapshots(t *testing.T, sessionIdentity, planAPath, planBPath string) (griteexport.Snapshot, griteexport.Snapshot, griteexport.Snapshot) {
	t.Helper()
	createdA := serviceEvent(t, "artifact-a-created", artifactAcceptanceIssueA, 1_000, "IssueCreated", map[string]any{
		"title": "Artifact issue A", "body": "acceptance body",
	})
	createdB := serviceEvent(t, "artifact-b-created", artifactAcceptanceIssueB, 1_100, "IssueCreated", map[string]any{
		"title": "Artifact issue B", "body": "target body",
	})
	modeAdded := serviceEvent(t, "artifact-mode-plan", artifactAcceptanceIssueA, 2_000, "LabelAdded", map[string]any{"label": "mode:plan"})
	sessionAdded := serviceEvent(t, "artifact-session", artifactAcceptanceIssueA, 2_100, "LabelAdded", map[string]any{"label": "session:" + sessionIdentity})
	relationshipAdded := serviceEvent(t, "artifact-relationship-added", artifactAcceptanceIssueA, 2_500, "DependencyAdded", map[string]any{
		"dep_type": "depends_on", "target": artifactAcceptanceIssueB,
	})
	planA := serviceEvent(t, "artifact-plan-a", artifactAcceptanceIssueA, 3_000, "CommentAdded", map[string]any{
		"body": "<!-- gavel:state {\"planPath\":\"" + planAPath + "\"} -->",
	})
	planB := serviceEvent(t, "artifact-plan-b", artifactAcceptanceIssueA, 3_000, "CommentAdded", map[string]any{
		"body": "<!-- gavel:state {\"planPath\":\"" + planBPath + "\"} -->",
	})
	relationshipRemoved := serviceEvent(t, "artifact-relationship-removed", artifactAcceptanceIssueA, 4_000, "DependencyRemoved", map[string]any{
		"dep_type": "depends_on", "target": artifactAcceptanceIssueB,
	})

	initialIssues := []griteexport.Issue{
		{
			IssueID: griteexport.ID(artifactAcceptanceIssueA), Title: "Artifact issue A", State: "open",
			Labels:    []string{"status:pending", "mode:plan", "session:" + sessionIdentity},
			CreatedTS: 1_000, UpdatedTS: 3_000, CommentCount: 1,
		},
		{
			IssueID: griteexport.ID(artifactAcceptanceIssueB), Title: "Artifact issue B", State: "open",
			Labels: []string{"status:pending"}, CreatedTS: 1_100, UpdatedTS: 2_500,
		},
	}
	initialEvents := []griteexport.Event{createdA, createdB, modeAdded, sessionAdded, relationshipAdded, planA}
	initial := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 4_000, EventCount: len(initialEvents)},
		Issues: initialIssues, Events: initialEvents,
	}

	finalIssues := append([]griteexport.Issue(nil), initialIssues...)
	finalIssues[0].UpdatedTS = 4_000
	finalIssues[0].CommentCount = 2
	deltaEvents := []griteexport.Event{planA, planB, relationshipRemoved}
	delta := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 5_000, EventCount: len(deltaEvents)},
		Issues: finalIssues, Events: deltaEvents,
	}
	probeEvents := []griteexport.Event{relationshipRemoved}
	probe := griteexport.Snapshot{
		Meta:   griteexport.Meta{SchemaVersion: 1, GeneratedTS: 6_000, EventCount: len(probeEvents)},
		Issues: append([]griteexport.Issue(nil), finalIssues...), Events: probeEvents,
	}
	return initial, delta, probe
}

func marshalArtifactAcceptanceSnapshot(t *testing.T, snapshot griteexport.Snapshot) []byte {
	t.Helper()
	data, err := json.MarshalIndent(snapshot, "", "  ")
	require.NoError(t, err)
	return append(data, '\n')
}

func assertArtifactAcceptanceFile(t *testing.T, dir, name string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	assert.Equal(t, want, got)
}
