package snapshots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecorderWritesAndRemovesOnSuccess(t *testing.T) {
	repo := initRepo(t)
	r := NewRecorder(repo)
	r.interval = 20 * time.Millisecond
	r.Start()

	snap := &testui.Snapshot{
		Status: testui.SnapshotStatus{Running: true},
		Tests:  []parsers.Test{{Name: "TestFoo", Passed: true}},
	}
	r.Update(snap)

	path := filepath.Join(repo, Dir, RunningName)
	require.Eventually(t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, time.Second, 10*time.Millisecond, "running.json should appear")

	r.Stop(true)
	_, err := os.Stat(path)
	assert.True(t, os.IsNotExist(err), "running.json should be removed on success")
}

func TestRecorderKeepsFileOnFailure(t *testing.T) {
	repo := initRepo(t)
	r := NewRecorder(repo)
	r.interval = 20 * time.Millisecond
	r.Start()

	snap := &testui.Snapshot{
		Status: testui.SnapshotStatus{Running: true},
		Tests:  []parsers.Test{{Name: "TestBar", Failed: true}},
	}
	r.Update(snap)

	path := filepath.Join(repo, Dir, RunningName)
	require.Eventually(t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, time.Second, 10*time.Millisecond)

	r.Stop(false)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "running.json should be retained on failure")

	var got testui.Snapshot
	require.NoError(t, json.Unmarshal(data, &got))
	require.Len(t, got.Tests, 1)
	assert.Equal(t, "TestBar", got.Tests[0].Name)
	assert.True(t, got.Tests[0].Failed)
}

func TestRecorderSkipsTickWhenNothingChanged(t *testing.T) {
	repo := initRepo(t)
	r := NewRecorder(repo)
	r.interval = 15 * time.Millisecond
	r.Start()
	defer r.Stop(true)

	snap := &testui.Snapshot{Status: testui.SnapshotStatus{Running: true}}
	r.Update(snap)

	path := filepath.Join(repo, Dir, RunningName)
	require.Eventually(t, func() bool {
		_, err := os.Stat(path)
		return err == nil
	}, time.Second, 5*time.Millisecond)

	stat1, err := os.Stat(path)
	require.NoError(t, err)
	mtime1 := stat1.ModTime()

	// Several ticks pass with no Update — file mtime must NOT advance.
	time.Sleep(80 * time.Millisecond)
	stat2, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, mtime1, stat2.ModTime(), "no-op ticks must not rewrite running.json")
}

func TestRecorderStopIdempotent(t *testing.T) {
	repo := initRepo(t)
	r := NewRecorder(repo)
	r.interval = 20 * time.Millisecond
	r.Start()
	r.Update(&testui.Snapshot{})
	r.Stop(true)
	// A second Stop must not panic or block.
	r.Stop(true)
}

func TestSavePerRunWritesTimestampedFile(t *testing.T) {
	repo := initRepo(t)
	snap := &testui.Snapshot{
		Status: testui.SnapshotStatus{LintRun: true},
		Tests:  []parsers.Test{{Name: "TestQux", Passed: true}},
	}
	started := time.Date(2026, 4, 27, 10, 43, 42, 0, time.UTC)

	path, err := SavePerRun(repo, snap, started)
	require.NoError(t, err)
	assert.Equal(t,
		filepath.Join(repo, Dir, "run-2026-04-27T10-43-42Z.json"),
		path,
	)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var got testui.Snapshot
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, got.Status.LintRun)
	require.Len(t, got.Tests, 1)
	assert.Equal(t, "TestQux", got.Tests[0].Name)
}

func TestSavePerRunDoesNotTouchPointers(t *testing.T) {
	repo := initRepo(t)
	snap := &testui.Snapshot{}

	_, err := SavePerRun(repo, snap, time.Now())
	require.NoError(t, err)

	p, err := LoadPointer(repo, PointerLast)
	require.NoError(t, err)
	assert.Nil(t, p, "SavePerRun must not write last.json")
}

func TestSavePerRunRejectsNilSnapshot(t *testing.T) {
	repo := initRepo(t)
	_, err := SavePerRun(repo, nil, time.Now())
	require.Error(t, err)
}

func TestSavePerRunRejectsEmptyWorkDir(t *testing.T) {
	_, err := SavePerRun("", &testui.Snapshot{}, time.Now())
	require.Error(t, err)
}
