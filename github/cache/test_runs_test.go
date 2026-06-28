package cache

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestRunCacheDisabledNoOp(t *testing.T) {
	// A disabled store answers reads as misses and drops writes, so the PR UI
	// can call Shared() without branching on whether postgres is configured.
	s := &Store{disabled: true}

	ts, syncedAt, err := s.LoadTestRunCursor(context.Background(), "/ws")
	require.NoError(t, err)
	assert.Zero(t, ts)
	assert.True(t, syncedAt.IsZero())

	runs, err := s.ListTestRuns(context.Background())
	require.NoError(t, err)
	assert.Empty(t, runs)

	path, err := s.GetTestRunPath(context.Background(), "/ws", "run-x")
	require.NoError(t, err)
	assert.Equal(t, "", path)

	assert.NoError(t, s.UpsertTestRuns(context.Background(), "/ws",
		[]TestRunCache{{RunID: "run-x", Path: "/ws/.gavel/run-x.json"}}, 1))
}

func TestTestRunCacheRoundtripIntegration(t *testing.T) {
	dsn := os.Getenv(EnvDSN)
	if dsn == "" {
		t.Skipf("set %s=postgres://... to run integration tests", EnvDSN)
	}
	t.Setenv(EnvDSN, dsn)
	t.Setenv(EnvDisable, "")

	store, err := Open()
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	const ws = "/tmp/test-run-cache-test"
	t.Cleanup(func() {
		_ = store.gorm().Where("workspace_dir = ?", ws).Delete(&TestRunCache{})
		_ = store.gorm().Where("workspace_dir = ?", ws).Delete(&TestRunCursor{})
	})

	ctx := context.Background()
	rows := []TestRunCache{
		{RunID: "run-1", Path: ws + "/.gavel/run-1.json", Kind: "test", StartedTS: 100, Passed: 5, Failed: 1, Total: 6},
		{RunID: "run-2", Path: ws + "/.gavel/run-2.json", Kind: "lint", StartedTS: 200, LintLinters: 2, LintViolations: 3},
	}
	require.NoError(t, store.UpsertTestRuns(ctx, ws, rows, 200))

	cursor, syncedAt, err := store.LoadTestRunCursor(ctx, ws)
	require.NoError(t, err)
	assert.Equal(t, int64(200), cursor)
	assert.False(t, syncedAt.IsZero())

	listed, err := store.ListTestRunsForDir(ctx, ws)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	// newest-first by StartedTS.
	assert.Equal(t, "run-2", listed[0].RunID)
	assert.Equal(t, "test", listed[1].Kind)

	path, err := store.GetTestRunPath(ctx, ws, "run-1")
	require.NoError(t, err)
	assert.Equal(t, ws+"/.gavel/run-1.json", path)

	missing, err := store.GetTestRunPath(ctx, ws, "run-nope")
	require.NoError(t, err)
	assert.Equal(t, "", missing)

	// Re-upsert overwrites in place (no duplicate row) and advances the cursor.
	require.NoError(t, store.UpsertTestRuns(ctx, ws,
		[]TestRunCache{{RunID: "run-1", Path: ws + "/.gavel/run-1.json", Kind: "test", StartedTS: 100, Passed: 6, Total: 6}}, 250))
	again, err := store.ListTestRunsForDir(ctx, ws)
	require.NoError(t, err)
	assert.Len(t, again, 2, "re-upsert must not duplicate run-1")

	cursor, _, err = store.LoadTestRunCursor(ctx, ws)
	require.NoError(t, err)
	assert.Equal(t, int64(250), cursor)
}
