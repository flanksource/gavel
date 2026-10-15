// Tests that a project action's finished run is read from the workspace's snapshot store, so it
// shows up when snapshots live in a bucket rather than in .gavel.
package ui

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCompletedProjectActionSnapshotReadsTheStore(t *testing.T) {
	store := blob.Local(t.TempDir())
	previous := snapshots.OpenStore
	snapshots.OpenStore = func(string) (blob.Store, error) { return store, nil }
	t.Cleanup(func() { snapshots.OpenStore = previous })
	workDir := t.TempDir()
	started := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	_, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: started.Add(time.Second)},
		Tests:    []parsers.Test{{Name: "TestAction", Passed: true}},
	}, started.Add(time.Second), "")
	require.NoError(t, err)

	run := testui.NewServer()
	loadCompletedProjectActionSnapshot(workDir, started, run)

	tests := run.Snapshot().Tests
	require.Len(t, tests, 1, "the run in the store is loaded")
	assert.Equal(t, "TestAction", tests[0].Name)
}
