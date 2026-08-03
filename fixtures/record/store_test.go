package record

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCreate(t *testing.T, store *Store, label string, kind Kind) Result {
	t.Helper()
	file, res, err := store.Create(label, kind)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return res
}

func TestStoreCreateNamesArtifacts(t *testing.T) {
	workDir := t.TempDir()
	started := time.Date(2026, 8, 2, 10, 43, 42, 0, time.UTC)
	store := NewStore(workDir, started)

	res := mustCreate(t, store, "GIT_ROOT_DIR follows the worktree", KindANSI)

	assert.Equal(t, filepath.Join(workDir, Dir), filepath.Dir(res.Path))
	assert.Equal(t, "rec-2026-08-02T10-43-42Z-git-root-dir-follows-the-worktree-ansi", res.ID)
	assert.Equal(t, res.ID+".cast.json", filepath.Base(res.Path))
	assert.Equal(t, "asciinema-v2", res.Format)
	assert.FileExists(t, res.Path)
}

func TestStoreCreateIsLazy(t *testing.T) {
	workDir := t.TempDir()
	NewStore(workDir, time.Now())

	// A run whose fixtures declare no `record:` must leave no directory behind.
	_, err := os.Stat(filepath.Join(workDir, Dir))
	assert.True(t, os.IsNotExist(err), "the recordings dir should not exist until something is recorded")
}

func TestStoreCreateSuffixesCollisions(t *testing.T) {
	store := NewStore(t.TempDir(), time.Now())

	first := mustCreate(t, store, "same name", KindHTTP)
	second := mustCreate(t, store, "same name", KindHTTP)

	assert.NotEqual(t, first.Path, second.Path, "two fixtures with one name must not overwrite each other")
	assert.Equal(t, first.ID+"-2", second.ID)
	assert.FileExists(t, first.Path)
}

func TestStoreCreateSeparatesKinds(t *testing.T) {
	store := NewStore(t.TempDir(), time.Now())

	har := mustCreate(t, store, "call github", KindHTTP)
	sql := mustCreate(t, store, "call github", KindSQL)

	assert.Equal(t, ".har", filepath.Ext(har.Path), "a HAR keeps the extension devtools opens")
	assert.Equal(t, "har-1.2", har.Format)
	assert.True(t, filepath.Ext(sql.Path) == ".jsonl")
	assert.Equal(t, "jsonl", sql.Format)
}

func TestPruneRetiresWholeRuns(t *testing.T) {
	workDir := t.TempDir()
	base := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)

	runs := make([][]Result, 0, 4)
	for i := 0; i < 4; i++ {
		store := NewStore(workDir, base.Add(time.Duration(i)*time.Hour))
		// Each run emits two artifacts, so a file-count retention would leave a
		// run half-deleted.
		runs = append(runs, []Result{
			mustCreate(t, store, "fixture", KindANSI),
			mustCreate(t, store, "fixture", KindHTTP),
		})
	}

	require.NoError(t, Prune(workDir, 2))

	for _, res := range append(runs[0], runs[1]...) {
		assert.NoFileExists(t, res.Path, "the two oldest runs should be gone entirely")
	}
	for _, res := range append(runs[2], runs[3]...) {
		assert.FileExists(t, res.Path, "the two newest runs should survive intact")
	}
}

func TestPruneIgnoresMissingDirAndForeignFiles(t *testing.T) {
	workDir := t.TempDir()
	require.NoError(t, Prune(workDir, 3), "pruning a run that recorded nothing is not an error")

	dir := filepath.Join(workDir, Dir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	foreign := filepath.Join(dir, "notes.txt")
	require.NoError(t, os.WriteFile(foreign, []byte("hand-written"), 0o644))

	require.NoError(t, Prune(workDir, 0))
	assert.FileExists(t, foreign, "prune must only touch files it wrote")
}
