package migrategrite

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/flanksource/gavel/todos/griteexport"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArtifactsAreSecureImmutableAndFinalize(t *testing.T) {
	initialSource := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	initialSnapshot := decodeSnapshot(t, string(initialSource))
	initialDocument, err := Normalize(initialSnapshot)
	require.NoError(t, err)
	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	report := &ImportReport{
		WorkspaceID: workspaceID,
		Watermark:   griteexport.WatermarkFor(initialSnapshot.Events),
		Validation: ImportValidation{
			SourceHash: initialDocument.SourceHash, ImportFingerprint: "initial-fingerprint",
			TargetChecksum: "target-initial", CaptainChecksum: "captain-initial",
		},
		Warnings: []Warning{{Code: "warning", IssueID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Message: "explicit warning"}},
		Rollback: Rollback{Native: native.ImportRollback{CreatedIssueIDs: []uuid.UUID{uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")}}},
	}
	dir := filepath.Join(t.TempDir(), "run")
	pending, err := PrepareArtifactGeneration(dir)
	require.NoError(t, err)
	require.NoError(t, WritePreparedSource(pending, "source-initial.json", initialSource))
	require.NoError(t, WriteInitialArtifacts(pending, initialSource, initialSnapshot, report, ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", RepoKey: "github.com/example/repo", BinaryVersion: "test",
	}))
	require.NoError(t, PublishArtifactGeneration(pending, dir))
	require.ErrorIs(t, WriteInitialArtifacts(dir, initialSource, initialSnapshot, report), ErrArtifactCommitted)
	require.NoError(t, VerifyArtifactChecksums(dir))

	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())
	for _, name := range []string{"source-initial.json", "manifest.json", "validation-initial.json", "report-initial.json", "rollback.json", "rollback.md", "checksums.sha256"} {
		info, err := os.Stat(filepath.Join(dir, name))
		require.NoError(t, err, name)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), name)
	}
	checksums, err := os.ReadFile(filepath.Join(dir, "checksums.sha256"))
	require.NoError(t, err)
	assert.Contains(t, string(checksums), "  source-initial.json\n")
	assert.NotContains(t, string(checksums), "checksums.sha256")

	finalSource := []byte(`{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	finalSnapshot := decodeSnapshot(t, string(finalSource))
	merged, err := MergeSnapshots(initialSnapshot, finalSnapshot)
	require.NoError(t, err)
	finalDocument, err := Normalize(merged)
	require.NoError(t, err)
	probeSource := []byte(`{"meta":{"schema_version":1,"generated_ts":3000,"event_count":0},"issues":[],"events":[]}`)
	finalReport := *report
	finalReport.Validation.SourceHash = finalDocument.SourceHash
	finalReport.Validation.ImportFingerprint = "final-fingerprint"
	finalReport.Validation.TargetChecksum = "target-final"
	finalReport.Validation.CaptainChecksum = "captain-final"
	combined := MergeRollback(report.Rollback, Rollback{CreatedCaptainPlanIDs: []uuid.UUID{uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")}})
	finalDir := dir + "-final"
	finalPending, err := PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, WritePreparedSource(finalPending, "source-final-delta.json", finalSource))
	require.NoError(t, WritePreparedSource(finalPending, "source-final-probe.json", probeSource))
	require.NoError(t, WriteFinalArtifacts(finalPending, dir, finalSource, finalSnapshot, &finalReport, combined, ArtifactProvenance{
		SourceMode: "live", SourceRoot: "/repo", CursorMS: 10, FrozenValidated: true, ProbeRawHash: SHA256Bytes(probeSource),
		WALHead: []byte(`{"segment":2,"offset":10}`), ProbeWALHead: []byte(`{"offset":10,"segment":2}`),
	}))
	require.NoError(t, PublishArtifactGeneration(finalPending, finalDir))
	require.NoError(t, VerifyArtifactChecksums(finalDir))

	manifest, err := LoadArtifactManifest(finalDir)
	require.NoError(t, err)
	assert.Equal(t, "final", manifest.State)
	assert.True(t, manifest.Frozen)
	assert.Equal(t, initialDocument.SourceHash, manifest.InitialSourceHash)
	assert.Equal(t, finalDocument.SourceHash, manifest.FinalSourceHash)
	assert.Equal(t, SHA256Bytes(probeSource), manifest.FinalProbeRawHash)
	assert.Equal(t, "live", manifest.FinalSourceMode)
	assert.JSONEq(t, `{"segment":2,"offset":10}`, string(manifest.FinalWALHead))
	assert.JSONEq(t, `{"segment":2,"offset":10}`, string(manifest.FinalProbeWALHead))
	probeCopy, err := os.ReadFile(filepath.Join(finalDir, "source-final-probe.json"))
	require.NoError(t, err)
	assert.Equal(t, probeSource, probeCopy)
	loadedRollback, err := LoadRollback(finalDir)
	require.NoError(t, err)
	assert.Len(t, loadedRollback.Native.CreatedIssueIDs, 1)
	assert.Len(t, loadedRollback.CreatedCaptainPlanIDs, 1)

	_, err = PrepareArtifactGeneration(finalDir)
	require.ErrorIs(t, err, ErrArtifactCommitted)
	initialManifest, err := LoadArtifactManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, "initial", initialManifest.State, "finalization must not mutate initial evidence")
}

func TestFinalArtifactRequiresFrozenProbeAndRevalidatesItsContent(t *testing.T) {
	initialSource := []byte(`{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	initialSnapshot := decodeSnapshot(t, string(initialSource))
	initialDocument, err := Normalize(initialSnapshot)
	require.NoError(t, err)
	workspaceID := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	initialReport := &ImportReport{
		WorkspaceID: workspaceID,
		Validation: ImportValidation{
			SourceHash: initialDocument.SourceHash, ImportFingerprint: "initial",
			TargetChecksum: "target", CaptainChecksum: "captain",
		},
	}
	root := t.TempDir()
	initialDir := filepath.Join(root, "initial")
	pending, err := PrepareArtifactGeneration(initialDir)
	require.NoError(t, err)
	require.NoError(t, WritePreparedSource(pending, "source-initial.json", initialSource))
	require.NoError(t, WriteInitialArtifacts(pending, initialSource, initialSnapshot, initialReport, ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo",
	}))
	require.NoError(t, PublishArtifactGeneration(pending, initialDir))

	finalSource := []byte(`{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	finalSnapshot := decodeSnapshot(t, string(finalSource))
	merged, err := MergeSnapshots(initialSnapshot, finalSnapshot)
	require.NoError(t, err)
	finalDocument, err := Normalize(merged)
	require.NoError(t, err)
	finalReport := *initialReport
	finalReport.Validation.SourceHash = finalDocument.SourceHash
	finalReport.Validation.ImportFingerprint = "final"

	missingProbeDir := filepath.Join(root, "missing-probe")
	missingProbe, err := PrepareArtifactGeneration(missingProbeDir)
	require.NoError(t, err)
	require.NoError(t, WritePreparedSource(missingProbe, "source-final-delta.json", finalSource))
	err = WriteFinalArtifacts(missingProbe, initialDir, finalSource, finalSnapshot, &finalReport, Rollback{}, ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", FrozenValidated: true,
	})
	require.ErrorContains(t, err, "probe")
	require.NoError(t, DiscardArtifactGeneration(missingProbe))

	mismatchedWALDir := filepath.Join(root, "mismatched-wal")
	mismatchedWAL, err := PrepareArtifactGeneration(mismatchedWALDir)
	require.NoError(t, err)
	probeSource := []byte(`{"meta":{"schema_version":1,"generated_ts":3000,"event_count":0},"issues":[],"events":[]}`)
	require.NoError(t, WritePreparedSource(mismatchedWAL, "source-final-delta.json", finalSource))
	require.NoError(t, WritePreparedSource(mismatchedWAL, "source-final-probe.json", probeSource))
	err = WriteFinalArtifacts(mismatchedWAL, initialDir, finalSource, finalSnapshot, &finalReport, Rollback{}, ArtifactProvenance{
		SourceMode: "live", SourceRoot: "/repo", FrozenValidated: true,
		WALHead: []byte(`"delta"`), ProbeWALHead: []byte(`"probe"`),
	})
	require.ErrorContains(t, err, "WAL heads differ")
	require.NoError(t, DiscardArtifactGeneration(mismatchedWAL))

	finalDir := filepath.Join(root, "final")
	finalPending, err := PrepareArtifactGeneration(finalDir)
	require.NoError(t, err)
	require.NoError(t, WritePreparedSource(finalPending, "source-final-delta.json", finalSource))
	require.NoError(t, WritePreparedSource(finalPending, "source-final-probe.json", probeSource))
	require.NoError(t, WriteFinalArtifacts(finalPending, initialDir, finalSource, finalSnapshot, &finalReport, Rollback{}, ArtifactProvenance{
		SourceMode: "file", SourceRoot: "/repo", FrozenValidated: true,
	}))
	require.NoError(t, PublishArtifactGeneration(finalPending, finalDir))

	changedProbe := []byte(`{"meta":{"schema_version":1,"generated_ts":4000,"event_count":0},"issues":[{"issue_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","title":"changed","state":"open"}],"events":[]}`)
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "source-final-probe.json"), changedProbe, 0o600))
	manifest, err := LoadArtifactManifest(finalDir)
	require.NoError(t, err)
	manifest.FinalProbeRawHash = SHA256Bytes(changedProbe)
	require.NoError(t, writeJSONArtifact(filepath.Join(finalDir, "manifest.json"), manifest))
	require.NoError(t, writeArtifactChecksums(finalDir))
	require.ErrorContains(t, VerifyArtifactChecksums(finalDir), "snapshot changed")
}

func TestValidateFrozenWALHeads(t *testing.T) {
	require.NoError(t, ValidateFrozenWALHeads(
		[]byte(`{"segment":2,"offset":10}`),
		[]byte(`{"offset":10,"segment":2}`),
	))
	require.ErrorContains(t, ValidateFrozenWALHeads(nil, []byte(`"head"`)), "missing")
	require.ErrorContains(t, ValidateFrozenWALHeads([]byte(`null`), []byte(`null`)), "missing")
	require.ErrorContains(t, ValidateFrozenWALHeads([]byte(`"head-a"`), []byte(`"head-b"`)), "not frozen")
}

func TestPrepareArtifactGenerationRefusesPreservedPendingSibling(t *testing.T) {
	workspaceDir := t.TempDir()
	firstTarget := filepath.Join(workspaceDir, "initial-source-one")
	secondTarget := filepath.Join(workspaceDir, "initial-source-two")
	pending, err := PrepareArtifactGeneration(firstTarget)
	require.NoError(t, err)

	_, err = PrepareArtifactGeneration(secondTarget)
	require.ErrorIs(t, err, ErrArtifactPending)
	require.ErrorContains(t, err, pending)
	require.ErrorContains(t, err, "publish or explicitly discard")

	require.NoError(t, DiscardArtifactGeneration(pending))
	replacement, err := PrepareArtifactGeneration(secondTarget)
	require.NoError(t, err)
	require.NoError(t, DiscardArtifactGeneration(replacement))
}

func TestPrepareArtifactGenerationAtomicallyReservesWorkspace(t *testing.T) {
	workspaceDir := t.TempDir()
	targets := []string{
		filepath.Join(workspaceDir, "initial-source-one"),
		filepath.Join(workspaceDir, "initial-source-two"),
	}
	type result struct {
		pending string
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, len(targets))
	var wait sync.WaitGroup
	for _, target := range targets {
		wait.Add(1)
		go func(target string) {
			defer wait.Done()
			<-start
			pending, err := PrepareArtifactGeneration(target)
			results <- result{pending: pending, err: err}
		}(target)
	}
	close(start)
	wait.Wait()
	close(results)

	var winner string
	losers := 0
	for result := range results {
		if result.err == nil {
			require.Empty(t, winner, "only one caller may reserve the workspace")
			winner = result.pending
			continue
		}
		require.ErrorIs(t, result.err, ErrArtifactPending)
		require.ErrorContains(t, result.err, ".workspace.pending-generation")
		losers++
	}
	require.NotEmpty(t, winner)
	assert.Equal(t, 1, losers)
	entries, err := os.ReadDir(workspaceDir)
	require.NoError(t, err)
	pendingCount := 0
	for _, entry := range entries {
		if strings.Contains(entry.Name(), ".pending-") {
			pendingCount++
		}
	}
	assert.Equal(t, 1, pendingCount)
	require.NoError(t, DiscardArtifactGeneration(winner))
	_, err = os.Lstat(winner)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestArtifactRunIDAndRollbackMergeAreStable(t *testing.T) {
	snapshot := decodeSnapshot(t, `{"meta":{"schema_version":1,"generated_ts":1000,"event_count":0},"issues":[],"events":[]}`)
	later := decodeSnapshot(t, `{"meta":{"schema_version":1,"generated_ts":2000,"event_count":0},"issues":[],"events":[]}`)
	assert.Equal(t, "initial-0123456789abcdef", ArtifactRunID(snapshot, "0123456789abcdef"))
	assert.Equal(t, ArtifactRunID(snapshot, "0123456789abcdef"), ArtifactRunID(later, "0123456789abcdef"))
	assert.False(t, strings.Contains(ArtifactDirectory("/tmp/root", "github.com/flanksource/gavel", "run"), "github.com/"))

	issueID := uuid.New()
	first := Rollback{Native: native.ImportRollback{
		CreatedIssueIDs: []uuid.UUID{issueID},
		InsertedEvents:  []native.ImportEventKey{{Source: "grite-import", SourceID: "a"}},
	}}
	second := Rollback{Native: native.ImportRollback{
		CreatedIssueIDs: []uuid.UUID{issueID},
		InsertedEvents:  []native.ImportEventKey{{Source: "grite-import", SourceID: "a"}, {Source: "grite-import", SourceID: "b"}},
	}}
	merged := MergeRollback(first, second)
	assert.Equal(t, []uuid.UUID{issueID}, merged.Native.CreatedIssueIDs)
	assert.Len(t, merged.Native.InsertedEvents, 2)
}
