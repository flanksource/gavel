package status

import (
	"context"
	"errors"
	"fmt"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestGatherWithFreshSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(context.Context, string) (string, string, error) {
		return "deadbeef", "", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "deadbeef", Path: ".gavel/sha-deadbeef.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return &testui.Snapshot{
			Tests: []parsers.Test{{File: "a.go", Failed: true}},
			Lint: []*linters.LinterResult{{
				Violations: []models.Violation{{File: "a.go", Severity: "error"}},
			}},
		}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	assert.False(t, result.ResultsStale)
	assert.Equal(t, 1, result.Files[0].TestStatus.Failed)
	assert.Equal(t, 1, result.Files[0].LintStatus.Errors)
}

func TestGatherWithStaleSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(context.Context, string) (string, string, error) {
		return "newsha", "ab12cd34", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "oldsha", Path: ".gavel/sha-oldsha.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return &testui.Snapshot{
			Tests: []parsers.Test{{File: "a.go", Passed: true}},
		}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	assert.True(t, result.ResultsStale)
	assert.Equal(t, "oldsha", result.ResultsSHA)
	assert.Equal(t, 1, result.Files[0].TestStatus.Passed)
	assert.True(t, result.Files[0].ResultsStale)
}

func TestGatherNoSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(context.Context, string) (string, string, error) {
		return "deadbeef", "", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return nil, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		t.Fatal("loadSnapshotFunc should not be called when pointer is nil")
		return nil, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	assert.False(t, result.ResultsStale)
	assert.Empty(t, result.ResultsSHA)
	assert.Equal(t, 0, result.Files[0].TestStatus.Failed)
}

func TestGatherStalePointerMissingSnapshotFile(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(context.Context, string) (string, string, error) {
		return "newsha", "ab12cd34", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "oldsha", Path: ".gavel/sha-oldsha-deadbeef.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return nil, fmt.Errorf("read snapshot .gavel/sha-oldsha-deadbeef.json: %w", os.ErrNotExist)
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Files[0].TestStatus.Failed)
	assert.Equal(t, 0, result.Files[0].TestStatus.Passed)
	assert.False(t, result.Files[0].ResultsStale)
}

func TestGatherSnapshotLoadErrorPropagates(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	wantErr := errors.New("decode snapshot: invalid JSON")
	restore := stubSnapshot(func(context.Context, string) (string, string, error) {
		return "newsha", "ab12cd34", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "oldsha", Path: ".gavel/sha-oldsha-deadbeef.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return nil, wantErr
	})
	defer restore()

	_, err := Gather(repo, Options{NoRepomap: true})
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
}

func stubSnapshot(
	id func(context.Context, string) (string, string, error),
	load func(string, string) (*snapshots.Pointer, error),
	snap func(string, *snapshots.Pointer) (*testui.Snapshot, error),
) func() {
	prevID, prevLoad, prevSnap := snapshotIDFunc, loadPointerFunc, loadSnapshotFunc
	snapshotIDFunc = id
	loadPointerFunc = load
	loadSnapshotFunc = snap
	return func() {
		snapshotIDFunc = prevID
		loadPointerFunc = prevLoad
		loadSnapshotFunc = prevSnap
	}
}
