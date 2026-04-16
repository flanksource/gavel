//go:build unix

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteInitialLockfile_RoundTrip validates the on-disk lockfile shape
// and that writeInitialLockfile + waitForHandoff form a compatible pair.
// The fork handoff depends on the child being able to observe pid==0
// (just-written) vs. pid!=0 (child reported ready) through readLockfilePID.
func TestWriteInitialLockfile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "port-1234.lock")
	require.NoError(t, writeInitialLockfile(path, 1234, "http://localhost:1234"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var payload uiLockfilePayload
	require.NoError(t, json.Unmarshal(data, &payload))
	assert.Equal(t, 1234, payload.Port)
	assert.Equal(t, "http://localhost:1234", payload.URL)
	assert.Equal(t, 0, payload.PID, "initial pid must be 0 so the child can flip it")
	assert.NotEmpty(t, payload.Deadline)

	// waitForHandoff must not return while pid is still 0.
	_, _, err = readLockfilePID(path)
	require.NoError(t, err)
	_, ok, _ := readLockfilePID(path)
	assert.False(t, ok, "pid==0 should not count as handoff complete")
}

// TestWaitForHandoff_ObservesChildFlip simulates the child's side of the
// handoff by writing a non-zero PID into the lockfile mid-wait, then
// asserts waitForHandoff observes the flip and returns the right PID.
func TestWaitForHandoff_ObservesChildFlip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "port-4321.lock")
	require.NoError(t, writeInitialLockfile(path, 4321, "http://localhost:4321"))

	// Flip the lockfile in a goroutine ~100ms into the parent's wait.
	go func() {
		time.Sleep(100 * time.Millisecond)
		payload := uiLockfilePayload{Port: 4321, URL: "http://localhost:4321", PID: 9999}
		data, _ := json.Marshal(payload)
		_ = os.WriteFile(path, data, 0o600)
	}()

	start := time.Now()
	pid, err := waitForHandoff(path, 2*time.Second)
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, 9999, pid)
	assert.Less(t, elapsed, 500*time.Millisecond, "should notice the flip within poll interval")
}

func TestWaitForHandoff_DeadlineExpires(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "port-5555.lock")
	require.NoError(t, writeInitialLockfile(path, 5555, "http://localhost:5555"))

	start := time.Now()
	_, err := waitForHandoff(path, 200*time.Millisecond)
	elapsed := time.Since(start)
	require.Error(t, err, "expected handoff deadline error")
	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond)
	assert.Less(t, elapsed, 500*time.Millisecond)
}

func TestWriteSnapshotJSON_ReadByLoadResults(t *testing.T) {
	clicky.ClearGlobalTasks()
	t.Cleanup(clicky.ClearGlobalTasks)

	// The fork parent writes the snapshot and the child's loadResults reads
	// it — they must agree on the JSON shape. This test pins that contract.
	dir := t.TempDir()
	path := filepath.Join(dir, "results.json")

	tests := []parsers.Test{
		{Name: "TestA", Passed: true, Framework: parsers.GoTest},
		{Name: "TestB", Failed: true, Framework: parsers.GoTest, Message: "boom"},
	}
	lint := []*linters.LinterResult{{Linter: "golangci-lint"}}
	require.NoError(t, writeSnapshotJSON(path, testui.Snapshot{
		Git:    &testui.SnapshotGit{Root: "/tmp/repo", Repo: "repo"},
		Status: testui.SnapshotStatus{Running: false, LintRun: true},
		Tests:  tests,
		Lint:   lint,
	}))

	srv := testui.NewServer()
	require.NoError(t, loadResults(srv, path))
	require.Equal(t, "/tmp/repo", srv.GitRoot())

	handler := srv.Handler()
	req, _ := http.NewRequest("GET", "/api/tests", nil)
	rw := &recordingResponse{header: http.Header{}}
	handler.ServeHTTP(rw, req)

	var got map[string]any
	require.NoError(t, json.Unmarshal(rw.body, &got))
	loaded, _ := got["tests"].([]any)
	assert.Len(t, loaded, 2)
}
