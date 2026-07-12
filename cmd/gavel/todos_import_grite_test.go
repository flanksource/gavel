package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/migrategrite"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportGriteCommandIsDisposableAndDoesNotChangeProviderFlags(t *testing.T) {
	assert.True(t, todosImportGriteCmd.Hidden)
	assert.Same(t, todosCmd, todosImportGriteCmd.Parent())
	for _, name := range []string{"input", "probe-input", "workspace", "artifact-dir", "finalize", "plan-root", "frozen", "strict"} {
		assert.NotNil(t, todosImportGriteCmd.Flags().Lookup(name), name)
	}
	provider := todosCmd.PersistentFlags().Lookup("provider")
	require.NotNil(t, provider)
	assert.Equal(t, "grite", provider.DefValue, "milestone 5 must not cut over runtime provider routing")
}

func TestValidateImportGriteOptions(t *testing.T) {
	assert.NoError(t, validateImportGriteOptions(importGriteOptions{}))
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Input: "initial.json"}), "workspace")
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Frozen: true}), "only valid")
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Finalize: "run"}), "requires --frozen")
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Finalize: "run", Frozen: true, ArtifactDir: "other"}), "cannot be combined")
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Finalize: "run", Frozen: true, Workspace: "repo"}), "workspace")
	require.ErrorContains(t, validateImportGriteOptions(importGriteOptions{Finalize: "run", Frozen: true, Input: "delta.json"}), "probe-input")
	assert.NoError(t, validateImportGriteOptions(importGriteOptions{Finalize: "run", Frozen: true, Input: "delta.json", ProbeInput: "probe.json"}))
}

func TestLoadGriteSourceFileModeDoesNotInvokeGrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")
	raw := []byte(`{
		"meta":{"schema_version":1,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open"}],
		"events":[{"event_id":"event","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"x","body":"body"}}}]
	}`)
	require.NoError(t, os.WriteFile(path, raw, 0o600))
	source, err := loadGriteSource(context.Background(), dir, path, 42)
	require.NoError(t, err)
	assert.False(t, source.Live)
	assert.Equal(t, raw, source.Raw)
	assert.Equal(t, int64(42), source.CursorMS)
	assert.Len(t, source.Snapshot.Events, 1)
}

func TestWorkspaceArtifactKey(t *testing.T) {
	id := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	assert.Equal(t, id.String(), workspaceArtifactKey(id))
	assert.Equal(t, "workspace", workspaceArtifactKey(uuid.Nil))
	first := native.NormalizeImportWorkspace(native.ImportWorkspace{
		RepoKey: " GitHub.COM/Flanksource/Gavel ", RootPath: "/tmp/work/../gavel", DisplayName: " Gavel ",
	})
	first.ID = native.DeterministicImportWorkspaceID(first)
	retry := native.NormalizeImportWorkspace(native.ImportWorkspace{
		RepoKey: "github.com/flanksource/gavel", RootPath: "/tmp/gavel", DisplayName: "Gavel",
	})
	retry.ID = native.DeterministicImportWorkspaceID(retry)
	assert.Equal(t, first, retry)
	assert.Equal(t, workspaceArtifactKey(first.ID), workspaceArtifactKey(retry.ID))
	assert.Equal(t,
		native.DeterministicImportWorkspaceID(native.ImportWorkspace{RootPath: "/tmp/work/../gavel"}),
		native.DeterministicImportWorkspaceID(native.ImportWorkspace{RootPath: "/tmp/gavel"}),
	)
}

func TestValidateOfflineGriteProbeRequiresASeparateLaterCapture(t *testing.T) {
	delta := loadedGriteSource{
		Raw:      []byte("delta"),
		Snapshot: griteexport.Snapshot{Meta: griteexport.Meta{GeneratedTS: 100}},
	}
	probe := loadedGriteSource{
		Raw:      []byte("probe"),
		Snapshot: griteexport.Snapshot{Meta: griteexport.Meta{GeneratedTS: 101}},
	}
	require.ErrorContains(t, validateOfflineGriteProbe("/repo", "delta.json", "delta.json", delta, probe), "separately captured")

	identical := probe
	identical.Raw = append([]byte(nil), delta.Raw...)
	require.ErrorContains(t, validateOfflineGriteProbe("/repo", "delta.json", "probe.json", delta, identical), "byte-identical")

	stale := probe
	stale.Snapshot.Meta.GeneratedTS = delta.Snapshot.Meta.GeneratedTS
	require.ErrorContains(t, validateOfflineGriteProbe("/repo", "delta.json", "probe.json", delta, stale), "later")
	require.NoError(t, validateOfflineGriteProbe("/repo", "delta.json", "probe.json", delta, probe))
}

func TestExistingInitialArtifactRejectsDifferentWorkspaceID(t *testing.T) {
	raw := []byte(`{
		"meta":{"schema_version":1,"generated_ts":1000,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open"}],
		"events":[{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"x","body":"body"}}}]
	}`)
	snapshot, err := griteexport.DecodeFile(raw)
	require.NoError(t, err)
	document, err := migrategrite.Normalize(snapshot)
	require.NoError(t, err)
	artifactWorkspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	report := &migrategrite.ImportReport{
		WorkspaceID: artifactWorkspaceID,
		Watermark:   document.Watermark,
		Validation: migrategrite.ImportValidation{
			SourceHash: document.SourceHash, ImportFingerprint: "fingerprint",
			TargetChecksum: "target", CaptainChecksum: "captain",
		},
	}
	dir := filepath.Join(t.TempDir(), "artifact")
	pending, err := migrategrite.PrepareArtifactGeneration(dir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-initial.json", raw))
	require.NoError(t, migrategrite.WriteInitialArtifacts(pending, raw, snapshot, report, migrategrite.ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", RepoKey: "github.com/example/repo",
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(pending, dir))

	source := loadedGriteSource{Snapshot: snapshot, Raw: raw}
	matching := native.ImportWorkspace{ID: artifactWorkspaceID, RepoKey: "github.com/example/repo", RootPath: "/repo"}
	loaded, exists, err := existingInitialArtifact(dir, source, document, matching, "/repo")
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, artifactWorkspaceID, loaded.WorkspaceID)

	different := matching
	different.ID = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	_, exists, err = existingInitialArtifact(dir, source, document, different, "/repo")
	require.Error(t, err)
	assert.False(t, exists)
}

func TestWorkspaceInitialArtifactReusesSemanticLiveBaselineAndRejectsSecondSource(t *testing.T) {
	firstRaw := []byte(`{
		"meta":{"schema_version":1,"generated_ts":1000,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open"}],
		"events":[{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"x","body":"body"}}}]
	}`)
	firstSnapshot, err := griteexport.DecodeFile(firstRaw)
	require.NoError(t, err)
	firstDocument, err := migrategrite.Normalize(firstSnapshot)
	require.NoError(t, err)
	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	createdIssueID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	report := &migrategrite.ImportReport{
		WorkspaceID: workspaceID,
		Watermark:   firstDocument.Watermark,
		Validation: migrategrite.ImportValidation{
			SourceHash: firstDocument.SourceHash, ImportFingerprint: "fingerprint",
			TargetChecksum: "target", CaptainChecksum: "captain",
		},
		Rollback: migrategrite.Rollback{Native: native.ImportRollback{CreatedIssueIDs: []uuid.UUID{createdIssueID}}},
	}
	artifactRoot := t.TempDir()
	firstDir := migrategrite.ArtifactDirectory(
		artifactRoot, "workspace", migrategrite.ArtifactRunID(firstSnapshot, firstDocument.SourceHash),
	)
	pending, err := migrategrite.PrepareArtifactGeneration(firstDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-initial.json", firstRaw))
	require.NoError(t, migrategrite.WriteInitialArtifacts(pending, firstRaw, firstSnapshot, report, migrategrite.ArtifactProvenance{
		SourceMode: "live", SourceRoot: "/repo", RepoKey: "github.com/example/repo", WALHead: []byte(`"head-one"`),
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(pending, firstDir))

	laterRaw := []byte(`{
		"meta":{"schema_version":1,"generated_ts":2000,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"x","state":"open"}],
		"events":[{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"x","body":"body"}}}]
	}`)
	laterSnapshot, err := griteexport.DecodeFile(laterRaw)
	require.NoError(t, err)
	laterDocument, err := migrategrite.Normalize(laterSnapshot)
	require.NoError(t, err)
	assert.Equal(t, firstDocument.SourceHash, laterDocument.SourceHash)
	laterDir := migrategrite.ArtifactDirectory(
		artifactRoot, "workspace", migrategrite.ArtifactRunID(laterSnapshot, laterDocument.SourceHash),
	)
	assert.Equal(t, firstDir, laterDir, "generation time must not create a second semantic baseline")
	preservedPending, err := migrategrite.PrepareArtifactGeneration(filepath.Join(filepath.Dir(laterDir), "unpublished-final"))
	require.NoError(t, err)
	_, exists, err := resolveWorkspaceInitialArtifact(filepath.Dir(laterDir), laterDocument.SourceHash)
	require.ErrorIs(t, err, migrategrite.ErrArtifactPending)
	require.ErrorContains(t, err, preservedPending)
	require.ErrorContains(t, err, "publish or explicitly discard")
	assert.False(t, exists)
	require.NoError(t, migrategrite.DiscardArtifactGeneration(preservedPending))
	baseline, exists, err := resolveWorkspaceInitialArtifact(filepath.Dir(laterDir), laterDocument.SourceHash)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, firstDir, baseline)
	loaded, exists, err := existingInitialArtifact(
		baseline,
		loadedGriteSource{Snapshot: laterSnapshot, Raw: laterRaw, Live: true},
		laterDocument,
		native.ImportWorkspace{ID: workspaceID, RepoKey: "github.com/example/repo", RootPath: "/repo"},
		"/repo",
	)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, []uuid.UUID{createdIssueID}, loaded.Rollback.Native.CreatedIssueIDs, "reuse must retain the original rollback")
	manifest, err := migrategrite.LoadArtifactManifest(baseline)
	require.NoError(t, err)
	assert.Equal(t, migrategrite.SHA256Bytes(firstRaw), manifest.InitialRawHash)
	assert.NotEqual(t, migrategrite.SHA256Bytes(laterRaw), manifest.InitialRawHash, "reuse must not overwrite original evidence")

	differentRaw := []byte(`{
		"meta":{"schema_version":1,"generated_ts":3000,"event_count":1},
		"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"changed","state":"open"}],
		"events":[{"event_id":"created","issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","ts_unix_ms":1,"kind":{"IssueCreated":{"title":"changed","body":"body"}}}]
	}`)
	differentSnapshot, err := griteexport.DecodeFile(differentRaw)
	require.NoError(t, err)
	differentDocument, err := migrategrite.Normalize(differentSnapshot)
	require.NoError(t, err)
	_, exists, err = resolveWorkspaceInitialArtifact(filepath.Dir(laterDir), differentDocument.SourceHash)
	require.ErrorContains(t, err, "refuse a second rollback baseline")
	require.ErrorContains(t, err, "--finalize "+firstDir)
	assert.False(t, exists)
	entries, err := os.ReadDir(filepath.Dir(firstDir))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestResolveWorkspaceInitialArtifactRejectsMultipleAndTamperedBaselines(t *testing.T) {
	raw := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)

	multipleWorkspace := filepath.Join(t.TempDir(), "workspace")
	document := writeCLIInitialArtifactFixture(t, filepath.Join(multipleWorkspace, "first"), raw)
	writeCLIInitialArtifactFixture(t, filepath.Join(multipleWorkspace, "second"), raw)
	_, exists, err := resolveWorkspaceInitialArtifact(multipleWorkspace, document.SourceHash)
	require.ErrorContains(t, err, "multiple committed initial Grite baselines")
	assert.False(t, exists)

	tamperedWorkspace := filepath.Join(t.TempDir(), "workspace")
	tamperedDir := filepath.Join(tamperedWorkspace, "initial")
	document = writeCLIInitialArtifactFixture(t, tamperedDir, raw)
	reportPath := filepath.Join(tamperedDir, "report-initial.json")
	reportRaw, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(reportPath, append(reportRaw, ' '), 0o600))
	_, exists, err = resolveWorkspaceInitialArtifact(tamperedWorkspace, document.SourceHash)
	require.ErrorContains(t, err, "checksum mismatch")
	assert.False(t, exists)

	symlinkWorkspace := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.MkdirAll(symlinkWorkspace, 0o700))
	require.NoError(t, os.Symlink(tamperedDir, filepath.Join(symlinkWorkspace, "linked-generation")))
	_, exists, err = resolveWorkspaceInitialArtifact(symlinkWorkspace, document.SourceHash)
	require.ErrorContains(t, err, "not a real directory")
	assert.False(t, exists)
}

func TestResolveWorkspaceInitialArtifactRejectsFinalOnlyMigration(t *testing.T) {
	workspaceDir := filepath.Join(t.TempDir(), "workspace")
	initialDir := filepath.Join(workspaceDir, "initial")
	initialRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	initialDocument := writeCLIInitialArtifactFixture(t, initialDir, initialRaw)
	initialSnapshot, _, err := migrategrite.LoadInitialSnapshot(initialDir)
	require.NoError(t, err)
	initialReport, err := migrategrite.LoadImportReport(initialDir, "initial")
	require.NoError(t, err)

	deltaRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	delta, err := griteexport.DecodeFile(deltaRaw)
	require.NoError(t, err)
	merged, err := migrategrite.MergeSnapshots(initialSnapshot, delta)
	require.NoError(t, err)
	finalDocument, err := migrategrite.Normalize(merged)
	require.NoError(t, err)
	finalReport := *initialReport
	finalReport.Validation.SourceHash = finalDocument.SourceHash
	finalReport.Validation.ImportFingerprint = "final-fingerprint"
	finalReport.Validation.TargetChecksum = "final-target"
	finalReport.Validation.CaptainChecksum = "final-captain"
	probeRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":3000,"event_count":0},"issues":[],"events":[]}`)
	finalDir := filepath.Join(workspaceDir, "final")
	pending, err := migrategrite.PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-final-delta.json", deltaRaw))
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-final-probe.json", probeRaw))
	require.NoError(t, migrategrite.WriteFinalArtifacts(pending, initialDir, deltaRaw, delta, &finalReport, migrategrite.Rollback{}, migrategrite.ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", FrozenValidated: true,
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(pending, finalDir))
	require.NoError(t, migrategrite.DiscardArtifactGeneration(initialDir), "simulate loss of the original initial directory")

	_, exists, err := resolveWorkspaceInitialArtifact(workspaceDir, initialDocument.SourceHash)
	require.ErrorContains(t, err, "already finalized")
	require.ErrorContains(t, err, finalDir)
	assert.False(t, exists)
}

func TestCompareExistingArtifactReplayRequiresAnExactMutationFreeReadback(t *testing.T) {
	expected := &migrategrite.ImportReport{
		WorkspaceID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Watermark: griteexport.Watermark{
			TimestampMS: 42,
			EventIDs:    []griteexport.ID{"event-a", "event-b"},
		},
		Validation: migrategrite.ImportValidation{
			SourceHash: "source", ImportFingerprint: "fingerprint",
			TargetChecksum: "target", CaptainChecksum: "captain",
		},
	}
	actual := *expected
	require.NoError(t, compareExistingArtifactReplay(expected, &actual))

	mutating := actual
	mutating.Counts.EventsInserted = 1
	require.ErrorContains(t, compareExistingArtifactReplay(expected, &mutating), "would mutate")

	missingCaptainRevision := actual
	missingCaptainRevision.Rollback.AppendedCaptainRevisionIDs = []uuid.UUID{uuid.New()}
	require.ErrorContains(t, compareExistingArtifactReplay(expected, &missingCaptainRevision), "would mutate")

	driftedCaptain := actual
	driftedCaptain.Validation.CaptainChecksum = "changed"
	require.ErrorContains(t, compareExistingArtifactReplay(expected, &driftedCaptain), "Captain checksum changed")

	driftedWatermark := actual
	driftedWatermark.Watermark.EventIDs = []griteexport.ID{"event-a"}
	require.ErrorContains(t, compareExistingArtifactReplay(expected, &driftedWatermark), "watermark changed")
}

func TestExistingFinalArtifactRequiresTheExactFrozenProbe(t *testing.T) {
	initialRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	initial, err := griteexport.DecodeFile(initialRaw)
	require.NoError(t, err)
	initialDocument, err := migrategrite.Normalize(initial)
	require.NoError(t, err)
	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	initialReport := &migrategrite.ImportReport{
		WorkspaceID: workspaceID,
		Validation: migrategrite.ImportValidation{
			SourceHash: initialDocument.SourceHash, ImportFingerprint: "initial-fingerprint",
			TargetChecksum: "initial-target", CaptainChecksum: "initial-captain",
		},
	}
	root := t.TempDir()
	initialDir := filepath.Join(root, "initial")
	initialPending, err := migrategrite.PrepareArtifactGeneration(initialDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(initialPending, "source-initial.json", initialRaw))
	require.NoError(t, migrategrite.WriteInitialArtifacts(initialPending, initialRaw, initial, initialReport, migrategrite.ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo",
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(initialPending, initialDir))

	deltaRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	deltaSnapshot, err := griteexport.DecodeFile(deltaRaw)
	require.NoError(t, err)
	merged, err := migrategrite.MergeSnapshots(initial, deltaSnapshot)
	require.NoError(t, err)
	finalDocument, err := migrategrite.Normalize(merged)
	require.NoError(t, err)
	finalReport := *initialReport
	finalReport.Validation.SourceHash = finalDocument.SourceHash
	finalReport.Validation.ImportFingerprint = "final-fingerprint"
	finalReport.Validation.TargetChecksum = "final-target"
	finalReport.Validation.CaptainChecksum = "final-captain"
	probeRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":3000,"event_count":0},"issues":[],"events":[]}`)
	probeSnapshot, err := griteexport.DecodeFile(probeRaw)
	require.NoError(t, err)
	finalDir := filepath.Join(root, "final")
	finalPending, err := migrategrite.PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(finalPending, "source-final-delta.json", deltaRaw))
	require.NoError(t, migrategrite.WritePreparedSource(finalPending, "source-final-probe.json", probeRaw))
	require.NoError(t, migrategrite.WriteFinalArtifacts(finalPending, initialDir, deltaRaw, deltaSnapshot, &finalReport, migrategrite.Rollback{}, migrategrite.ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", FrozenValidated: true,
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(finalPending, finalDir))

	delta := loadedGriteSource{Snapshot: deltaSnapshot, Raw: deltaRaw}
	probe := loadedGriteSource{Snapshot: probeSnapshot, Raw: probeRaw}
	loaded, exists, err := existingFinalArtifact(finalDir, delta, probe, finalDocument, workspaceID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, finalReport.Validation.TargetChecksum, loaded.Validation.TargetChecksum)

	probe.Raw = []byte(`{"meta":{"schema_version":1,"generated_ts":4000,"event_count":0},"issues":[],"events":[]}`)
	_, exists, err = existingFinalArtifact(finalDir, delta, probe, finalDocument, workspaceID)
	require.ErrorContains(t, err, "different source content/state")
	assert.False(t, exists)
}

func TestExistingFinalArtifactReusesSemanticLiveFinalAndPreservesEvidence(t *testing.T) {
	root := t.TempDir()
	initialDir := filepath.Join(root, "initial")
	initialRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	writeCLIInitialArtifactFixture(t, initialDir, initialRaw)
	initialSnapshot, _, err := migrategrite.LoadInitialSnapshot(initialDir)
	require.NoError(t, err)
	initialReport, err := migrategrite.LoadImportReport(initialDir, "initial")
	require.NoError(t, err)

	deltaRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	deltaSnapshot, err := griteexport.DecodeFile(deltaRaw)
	require.NoError(t, err)
	merged, err := migrategrite.MergeSnapshots(initialSnapshot, deltaSnapshot)
	require.NoError(t, err)
	finalDocument, err := migrategrite.Normalize(merged)
	require.NoError(t, err)
	finalReport := *initialReport
	finalReport.Validation.SourceHash = finalDocument.SourceHash
	finalReport.Validation.ImportFingerprint = "live-final-fingerprint"
	finalReport.Validation.TargetChecksum = "live-final-target"
	finalReport.Validation.CaptainChecksum = "live-final-captain"
	probeRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":3000,"event_count":0},"issues":[],"events":[]}`)
	rollbackPlanID := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	originalRollback := migrategrite.Rollback{CreatedCaptainPlanIDs: []uuid.UUID{rollbackPlanID}}
	finalDir := filepath.Join(root, "final")
	pending, err := migrategrite.PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-final-delta.json", deltaRaw))
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-final-probe.json", probeRaw))
	require.NoError(t, migrategrite.WriteFinalArtifacts(pending, initialDir, deltaRaw, deltaSnapshot, &finalReport, originalRollback, migrategrite.ArtifactProvenance{
		SourceMode: "live", SourceRoot: "/repo", FrozenValidated: true,
		WALHead: []byte(`{"segment":2,"offset":10}`), ProbeWALHead: []byte(`{"offset":10,"segment":2}`),
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(pending, finalDir))

	laterDeltaRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":4000,"event_count":0},"issues":[],"events":[]}`)
	laterDeltaSnapshot, err := griteexport.DecodeFile(laterDeltaRaw)
	require.NoError(t, err)
	laterMerged, err := migrategrite.MergeSnapshots(initialSnapshot, laterDeltaSnapshot)
	require.NoError(t, err)
	laterDocument, err := migrategrite.Normalize(laterMerged)
	require.NoError(t, err)
	assert.Equal(t, finalDocument.SourceHash, laterDocument.SourceHash)
	laterProbeRaw := []byte(`{"meta":{"schema_version":1,"generated_ts":5000,"event_count":0},"issues":[],"events":[]}`)
	laterProbeSnapshot, err := griteexport.DecodeFile(laterProbeRaw)
	require.NoError(t, err)
	laterDelta := loadedGriteSource{
		Snapshot: laterDeltaSnapshot, Raw: laterDeltaRaw, Live: true,
		Result: griteexport.Result{WALHead: []byte(`{"offset":10,"segment":2}`)},
	}
	laterProbe := loadedGriteSource{
		Snapshot: laterProbeSnapshot, Raw: laterProbeRaw, Live: true,
		Result: griteexport.Result{WALHead: []byte(`{"segment":2,"offset":10}`)},
	}
	loaded, exists, err := existingFinalArtifact(finalDir, laterDelta, laterProbe, laterDocument, finalReport.WorkspaceID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, finalReport.Validation.TargetChecksum, loaded.Validation.TargetChecksum)
	storedRollback, err := migrategrite.LoadRollback(finalDir)
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{rollbackPlanID}, storedRollback.CreatedCaptainPlanIDs)
	manifest, err := migrategrite.LoadArtifactManifest(finalDir)
	require.NoError(t, err)
	assert.Equal(t, migrategrite.SHA256Bytes(deltaRaw), manifest.FinalRawHash)
	assert.Equal(t, migrategrite.SHA256Bytes(probeRaw), manifest.FinalProbeRawHash)
	assert.NotEqual(t, migrategrite.SHA256Bytes(laterDeltaRaw), manifest.FinalRawHash)
	assert.NotEqual(t, migrategrite.SHA256Bytes(laterProbeRaw), manifest.FinalProbeRawHash)

	laterDelta.Result.WALHead = []byte(`{"segment":3,"offset":10}`)
	laterProbe.Result.WALHead = []byte(`{"segment":3,"offset":10}`)
	_, exists, err = existingFinalArtifact(finalDir, laterDelta, laterProbe, laterDocument, finalReport.WorkspaceID)
	require.ErrorContains(t, err, "WAL binding")
	assert.False(t, exists)
}

func writeCLIInitialArtifactFixture(t *testing.T, target string, raw []byte) migrategrite.Document {
	t.Helper()
	snapshot, err := griteexport.DecodeFile(raw)
	require.NoError(t, err)
	document, err := migrategrite.Normalize(snapshot)
	require.NoError(t, err)
	report := &migrategrite.ImportReport{
		WorkspaceID: uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"),
		Watermark:   document.Watermark,
		Validation: migrategrite.ImportValidation{
			SourceHash: document.SourceHash, ImportFingerprint: "fixture-fingerprint",
			TargetChecksum: "fixture-target", CaptainChecksum: "fixture-captain",
		},
	}
	pending, err := migrategrite.PrepareArtifactGeneration(target)
	require.NoError(t, err)
	require.NoError(t, migrategrite.WritePreparedSource(pending, "source-initial.json", raw))
	require.NoError(t, migrategrite.WriteInitialArtifacts(pending, raw, snapshot, report, migrategrite.ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo",
	}))
	require.NoError(t, migrategrite.PublishArtifactGeneration(pending, target))
	return document
}
