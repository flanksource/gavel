package database

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCutoverArtifactCommitsPrivateVerifiedGeneration(t *testing.T) {
	target := filepath.Join(t.TempDir(), "private", "captain-legacy", "generation-1")
	createdAt := time.Date(2026, time.July, 13, 1, 2, 3, 0, time.UTC)
	dump := []byte{'P', 'G', 'D', 'M', 'P', 0, 1, 2, 3}
	request := CutoverArtifactRequest{
		Cutover:    "captain-legacy-session-v1",
		Generation: "generation-1",
		CreatedAt:  createdAt,
		Metadata:   map[string]any{"database": "gavel", "legacyRows": 37},
		Report:     map[string]any{"validated": true, "sessionRows": 37},
		Rollback:   map[string]any{"archiveTable": "captain_sessions_legacy_v1"},
		Snapshots: []CutoverArtifactSnapshot{
			{Name: "legacy-schema.sql", ContentType: "application/sql", Data: []byte("CREATE TABLE captain_sessions (...);\n")},
			{Name: "legacy-data.dump", ContentType: "application/vnd.postgresql.custom-dump", Data: dump},
		},
	}

	result, err := WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	assert.Equal(t, target, result.Directory)
	assert.Equal(t, createdAt, result.Manifest.CreatedAt)
	assert.Equal(t, request.Cutover, result.Manifest.Cutover)
	require.NoError(t, assertPrivateMode(target, 0o700))

	entries, err := os.ReadDir(target)
	require.NoError(t, err)
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
		require.NoError(t, assertPrivateMode(filepath.Join(target, entry.Name()), 0o600))
	}
	sort.Strings(names)
	assert.Equal(t, []string{
		"checksums.sha256", "legacy-data.dump", "legacy-schema.sql", "manifest.json",
		"report.json", "rollback.json", "rollback.md",
	}, names)

	actualDump, err := os.ReadFile(filepath.Join(target, "legacy-data.dump"))
	require.NoError(t, err)
	assert.Equal(t, dump, actualDump, "opaque pg_dump bytes must not be normalized")

	manifest, err := VerifyCutoverArtifact(target)
	require.NoError(t, err)
	assert.Equal(t, CutoverArtifactSchemaVersion, manifest.SchemaVersion)
	assert.Len(t, manifest.Files, 5)
	assert.Len(t, result.Checksums, 6, "checksums cover manifest plus five payload files")
	assert.NotContains(t, result.Checksums, cutoverChecksumsFile)
	for _, file := range manifest.Files {
		assert.Equal(t, result.Checksums[file.Name], file.SHA256)
	}

	var report map[string]any
	reportData, err := os.ReadFile(filepath.Join(target, cutoverReportFile))
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(reportData, &report))
	assert.Equal(t, float64(37), report["sessionRows"])
}

func TestWriteCutoverArtifactRefusesCommittedGeneration(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "generation")
	request := validCutoverArtifactRequest()

	_, err := WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	original, err := os.ReadFile(filepath.Join(target, "legacy.dump"))
	require.NoError(t, err)
	request.Snapshots[0].Data = []byte("different")
	_, err = WriteCutoverArtifact(target, request)
	require.ErrorIs(t, err, ErrCutoverArtifactCommitted)
	current, err := os.ReadFile(filepath.Join(target, "legacy.dump"))
	require.NoError(t, err)
	assert.Equal(t, original, current)
}

func TestWriteCutoverArtifactPromotesCompleteMatchingPendingGeneration(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "generation")
	pending := filepath.Join(root, ".generation.pending-generation")
	request := validCutoverArtifactRequest()
	request.CreatedAt = time.Time{}

	first, err := WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	require.NoError(t, os.Rename(target, pending), "simulate a crash after the pending generation was fully synced")

	recovered, err := WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	assert.Equal(t, target, recovered.Directory)
	assert.Equal(t, first.Manifest, recovered.Manifest)
	assert.Equal(t, first.Checksums, recovered.Checksums)
	_, err = VerifyCutoverArtifact(target)
	require.NoError(t, err)
	_, err = os.Lstat(pending)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestWriteCutoverArtifactQuarantinesIncompletePendingAndCompletesRetry(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "generation")
	pending := filepath.Join(root, ".generation.pending-generation")
	require.NoError(t, os.Mkdir(pending, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(pending, "partial-report.json"), []byte("interrupted evidence"), 0o600))

	result, err := WriteCutoverArtifact(target, validCutoverArtifactRequest())
	require.NoError(t, err)
	assert.Equal(t, target, result.Directory)
	_, err = VerifyCutoverArtifact(target)
	require.NoError(t, err)
	_, err = os.Lstat(pending)
	assert.ErrorIs(t, err, os.ErrNotExist)

	evidence, err := filepath.Glob(filepath.Join(root, ".generation.incomplete-evidence-*"))
	require.NoError(t, err)
	require.Len(t, evidence, 1)
	preserved, err := os.ReadFile(filepath.Join(evidence[0], "partial-report.json"))
	require.NoError(t, err)
	assert.Equal(t, "interrupted evidence", string(preserved))
}

func TestWriteCutoverArtifactQuarantinesCompleteMismatchedPendingAndWritesCurrentRequest(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "generation")
	pending := filepath.Join(root, ".generation.pending-generation")
	request := validCutoverArtifactRequest()

	_, err := WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	require.NoError(t, os.Rename(target, pending))
	request.Snapshots[0].Data = []byte("current pg_dump bytes")

	_, err = WriteCutoverArtifact(target, request)
	require.NoError(t, err)
	current, err := os.ReadFile(filepath.Join(target, "legacy.dump"))
	require.NoError(t, err)
	assert.Equal(t, request.Snapshots[0].Data, current)

	evidence, err := filepath.Glob(filepath.Join(root, ".generation.incomplete-evidence-*"))
	require.NoError(t, err)
	require.Len(t, evidence, 1)
	_, err = VerifyCutoverArtifact(evidence[0])
	require.NoError(t, err, "a complete but mismatched prior attempt remains valid evidence")
}

func TestWriteCutoverArtifactRejectsUnsafeOrIncompleteInputBeforeCreatingTarget(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*CutoverArtifactRequest)
	}{
		{name: "missing cutover", mutate: func(r *CutoverArtifactRequest) { r.Cutover = "" }},
		{name: "missing report", mutate: func(r *CutoverArtifactRequest) { r.Report = nil }},
		{name: "missing rollback", mutate: func(r *CutoverArtifactRequest) { r.Rollback = nil }},
		{name: "missing snapshot", mutate: func(r *CutoverArtifactRequest) { r.Snapshots = nil }},
		{name: "path traversal", mutate: func(r *CutoverArtifactRequest) { r.Snapshots[0].Name = "../legacy.dump" }},
		{name: "reserved filename", mutate: func(r *CutoverArtifactRequest) { r.Snapshots[0].Name = cutoverManifestFile }},
		{name: "duplicate filename", mutate: func(r *CutoverArtifactRequest) {
			r.Snapshots = append(r.Snapshots, r.Snapshots[0])
		}},
		{name: "unencodable report", mutate: func(r *CutoverArtifactRequest) { r.Report = make(chan int) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "artifact")
			request := validCutoverArtifactRequest()
			test.mutate(&request)
			_, err := WriteCutoverArtifact(target, request)
			require.Error(t, err)
			_, statErr := os.Lstat(target)
			assert.ErrorIs(t, statErr, os.ErrNotExist)
		})
	}
}

func TestVerifyCutoverArtifactDetectsTamperingAndPrivacyRegression(t *testing.T) {
	t.Run("content", func(t *testing.T) {
		target := filepath.Join(t.TempDir(), "artifact")
		_, err := WriteCutoverArtifact(target, validCutoverArtifactRequest())
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(target, "legacy.dump"), []byte("tampered"), 0o600))
		_, err = VerifyCutoverArtifact(target)
		require.ErrorIs(t, err, ErrCutoverArtifactInvalid)
		require.ErrorContains(t, err, "checksum mismatch")
	})

	t.Run("file mode", func(t *testing.T) {
		target := filepath.Join(t.TempDir(), "artifact")
		_, err := WriteCutoverArtifact(target, validCutoverArtifactRequest())
		require.NoError(t, err)
		require.NoError(t, os.Chmod(filepath.Join(target, cutoverReportFile), 0o644))
		_, err = VerifyCutoverArtifact(target)
		require.ErrorIs(t, err, ErrCutoverArtifactInvalid)
		require.ErrorContains(t, err, "0600")
	})

	t.Run("directory mode", func(t *testing.T) {
		target := filepath.Join(t.TempDir(), "artifact")
		_, err := WriteCutoverArtifact(target, validCutoverArtifactRequest())
		require.NoError(t, err)
		require.NoError(t, os.Chmod(target, 0o755))
		_, err = VerifyCutoverArtifact(target)
		require.ErrorIs(t, err, ErrCutoverArtifactInvalid)
		require.ErrorContains(t, err, "0700")
	})
}

func validCutoverArtifactRequest() CutoverArtifactRequest {
	return CutoverArtifactRequest{
		Cutover:   "captain-legacy-session-v1",
		CreatedAt: time.Date(2026, time.July, 13, 1, 2, 3, 0, time.UTC),
		Report:    map[string]any{"validated": true},
		Rollback:  map[string]any{"steps": []string{"restore archive"}},
		Snapshots: []CutoverArtifactSnapshot{
			{Name: "legacy.dump", Data: []byte("pg_dump bytes")},
		},
	}
}

func assertPrivateMode(path string, mode os.FileMode) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != mode {
		return errors.New(path + " mode is " + info.Mode().Perm().String() + ", want " + mode.String())
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New(path + " is a symlink")
	}
	if strings.HasSuffix(path, ".json") && !info.Mode().IsRegular() {
		return errors.New(path + " is not regular")
	}
	return nil
}
