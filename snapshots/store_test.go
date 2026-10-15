// Tests that snapshots go through the workspace's blob store: with storage configured they never
// touch the workspace's .gavel directory, except running.json and the copy --failed reads.
package snapshots

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// useRemoteStore points OpenStore at a store outside workDir, standing in for a bucket.
func useRemoteStore(t *testing.T) (blob.Store, string) {
	t.Helper()
	root := t.TempDir()
	store := blob.Local(root)
	previous := OpenStore
	OpenStore = func(string) (blob.Store, error) { return store, nil }
	t.Cleanup(func() { OpenStore = previous })
	return store, root
}

func readStoreKey(t *testing.T, store blob.Store, key string) []byte {
	t.Helper()
	reader, err := store.Read(context.Background(), key)
	require.NoError(t, err, key)
	defer reader.Close()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	return data
}

func TestSnapshotsGoThroughTheStore(t *testing.T) {
	store, _ := useRemoteStore(t)
	repo := initRepo(t)
	snap := sampleSnapshot()

	_, err := Save(repo, snap)
	require.NoError(t, err)
	started := time.Date(2026, 10, 15, 9, 30, 0, 0, time.UTC)
	_, err = SavePerRun(repo, snap, started, "")
	require.NoError(t, err)
	_, err = SavePerRun(repo, snap, started, "")
	require.NoError(t, err, "a second run in the same second gets its own key")

	objects, err := store.List(context.Background(), "")
	require.NoError(t, err)
	var keys []string
	for _, object := range objects {
		keys = append(keys, object.Key)
	}
	assert.Contains(t, keys, "last.json")
	assert.Contains(t, keys, "run-2026-10-15T09-30-00Z.json")
	assert.Contains(t, keys, "run-2026-10-15T09-30-00Z-2.json")
	assert.NoDirExists(t, filepath.Join(repo, Dir), "nothing is written to the workspace")

	pointer, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	require.NotNil(t, pointer)
	assert.Contains(t, string(readStoreKey(t, store, filepath.Base(pointer.Path))), `"tests"`)

	loaded, err := LoadByPointer(repo, pointer)
	require.NoError(t, err)
	assert.Equal(t, len(snap.Tests), len(loaded.Tests))

	last, err := LastRun(repo)
	require.NoError(t, err)
	require.NotNil(t, last)
	runs, err := ListRuns(repo, time.Time{})
	require.NoError(t, err)
	require.Len(t, runs, 2)

	data, err := ReadRun(repo, runs[0].RunID)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"tests"`)
	data, err = ReadRun(repo, PointerLast)
	require.NoError(t, err, "the last pointer resolves through the store")
	assert.Contains(t, string(data), `"tests"`)

	// --failed reads a file, so the run last names is copied into the workspace.
	resolved, err := ResolveLast(repo)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(repo, Dir, filepath.Base(pointer.Path)), resolved)
	copied, err := os.ReadFile(resolved)
	require.NoError(t, err)
	assert.Equal(t, readStoreKey(t, store, filepath.Base(pointer.Path)), copied)
}

func TestReadRunRejectsRunIDsOutsideTheRunFamily(t *testing.T) {
	useRemoteStore(t)
	repo := initRepo(t)
	for _, runID := range []string{"", "../last", "run-x/../../etc", `run-a\b`, "sha-abc", "other"} {
		_, err := ReadRun(repo, runID)
		assert.ErrorIs(t, err, ErrInvalidRunID, runID)
	}
	_, err := ReadRun(repo, "run-2026-01-01T00-00-00Z")
	assert.ErrorIs(t, err, blob.ErrNotFound, "a valid id with no run")
}

func TestRunningSnapshotStaysInTheWorkspace(t *testing.T) {
	store, _ := useRemoteStore(t)
	repo := initRepo(t)
	require.NoError(t, writeRunningJSON(repo, sampleSnapshot()))
	assert.FileExists(t, filepath.Join(repo, Dir, RunningName))
	exists, err := store.Exists(context.Background(), RunningName)
	require.NoError(t, err)
	assert.False(t, exists, "in-flight state is local")
}

func TestOpenStoreUsesTheHomeStorageConfig(t *testing.T) {
	home, bucket := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage:\n  path: "+bucket+"\n  prefix: gavel\n"), 0o644))
	repo := initRepo(t)

	_, err := Save(repo, sampleSnapshot())
	require.NoError(t, err)

	abs, err := filepath.Abs(repo)
	require.NoError(t, err)
	sum := sha256.Sum256([]byte(abs))
	workspace := hex.EncodeToString(sum[:])[:16]
	assert.FileExists(t, filepath.Join(bucket, "gavel", "workspaces", workspace, "last.json"))
	assert.NoDirExists(t, filepath.Join(repo, Dir))
	assert.Equal(t, filepath.Join(bucket, "gavel", "workspaces", workspace, "last.json"), Location(repo, "last.json"))
}

func TestOpenStoreDefaultsToTheWorkspaceDirectory(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	repo := initRepo(t)
	path, err := Save(repo, sampleSnapshot())
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(repo, Dir, PointerLast+".json"))
	assert.Equal(t, filepath.Join(repo, Dir, filepath.Base(path)), path, "Save reports the file it wrote")
	assert.Equal(t, filepath.Join(repo, Dir, "last.json"), Location(repo, "last.json"))
}

// sampleSnapshot is a snapshot with one passing test.
func sampleSnapshot() *testui.Snapshot {
	return &testui.Snapshot{Tests: []parsers.Test{{Name: "TestA", Passed: true}}}
}
