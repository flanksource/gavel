package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunPushHooks_AbortsOnFailure covers the core guarantee: if a pre
// hook fails, RunPushHooks returns a wrapped error and subsequent steps
// do not run. This is what gates whether `gavel test` proceeds to
// testrunner.Run or aborts before tests execute.
func TestRunPushHooks_AbortsOnFailure(t *testing.T) {
	dir := t.TempDir()
	marker := filepath.Join(dir, "second-marker")

	hooks := []verify.HookStep{
		{Name: "fail", Run: "exit 7"},
		{Name: "should-not-run", Run: "touch " + marker},
	}

	err := verify.RunPushHooks(dir, hooks, "pre")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `pre hook "fail" failed`)

	_, statErr := os.Stat(marker)
	assert.True(t, os.IsNotExist(statErr), "second hook should not have run: %v", statErr)
}

func TestRunPushHooks_RunsStepsInOrder(t *testing.T) {
	dir := t.TempDir()
	// Each hook appends its name to the same file; we then assert the
	// file contents match the declared order. Using a file instead of
	// channels keeps the test a simple integration check.
	log := filepath.Join(dir, "log")

	hooks := []verify.HookStep{
		{Name: "first", Run: "echo a >> " + log},
		{Name: "second", Run: "echo b >> " + log},
		{Name: "third", Run: "echo c >> " + log},
	}

	require.NoError(t, verify.RunPushHooks(dir, hooks, "pre"))

	data, err := os.ReadFile(log)
	require.NoError(t, err)
	assert.Equal(t, "a\nb\nc\n", string(data))
}

func TestRunPushHooks_RunsInWorkDir(t *testing.T) {
	// The hook runs in workDir, not the caller's cwd. Verify via `pwd`.
	dir := t.TempDir()
	marker := filepath.Join(dir, "pwd-marker")

	hooks := []verify.HookStep{
		{Name: "check-pwd", Run: "pwd > " + marker},
	}
	require.NoError(t, verify.RunPushHooks(dir, hooks, "pre"))

	data, err := os.ReadFile(marker)
	require.NoError(t, err)
	// macOS tempdirs can be under /var -> /private/var; allow both.
	got := string(data[:len(data)-1])
	assert.Contains(t, []string{dir, "/private" + dir}, got, "hook should have run in %s, got %s", dir, got)
}

func TestRunPushHooks_EmptyNameDefaultsToPhase(t *testing.T) {
	// An empty Name is rendered as the phase ("pre" / "post") in log
	// banners so users can still tell hooks apart. This is only visible
	// in log output — we just exercise the branch by running one.
	dir := t.TempDir()
	marker := filepath.Join(dir, "nameless")
	hooks := []verify.HookStep{
		{Run: "touch " + marker}, // Name intentionally empty
	}
	require.NoError(t, verify.RunPushHooks(dir, hooks, "post"))
	_, err := os.Stat(marker)
	require.NoError(t, err)
}

func TestRunPushHooks_EmptyRunIsSkipped(t *testing.T) {
	// A HookStep with no Run is silently skipped; this mirrors what the
	// YAML parser emits when a user writes `pre: [{name: noop}]` with no
	// command. Must not error out.
	dir := t.TempDir()
	hooks := []verify.HookStep{{Name: "noop"}}
	require.NoError(t, verify.RunPushHooks(dir, hooks, "pre"))
}

// resetHookTestsState clears the package-level hookTests slice between
// tests so one test's state can't leak into another. Mirrors t.Cleanup
// for the other package globals.
func resetHookTestsState(t *testing.T) {
	t.Helper()
	hookTestsMu.Lock()
	hookTests = nil
	hookTestsMu.Unlock()
	t.Cleanup(func() {
		hookTestsMu.Lock()
		hookTests = nil
		hookTestsMu.Unlock()
	})
}

func TestAppendRunningHookTest_StartsPending(t *testing.T) {
	resetHookTestsState(t)
	idx := appendRunningHookTest("pre", "deps", "make deps")
	assert.Equal(t, 0, idx)

	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	require.Len(t, hookTests, 1)
	ht := hookTests[0]
	assert.Equal(t, "deps", ht.Name)
	assert.Equal(t, "hook:pre", ht.Package)
	assert.Equal(t, "make deps", ht.Command)
	assert.True(t, ht.Pending, "new hook should be pending so the UI shows a spinner")
	assert.False(t, ht.Passed)
	assert.False(t, ht.Failed)
}

func TestFinishHookTest_Success(t *testing.T) {
	resetHookTestsState(t)
	idx := appendRunningHookTest("pre", "deps", "make deps")
	finishHookTest(idx, 42*time.Millisecond, "hello\n", nil)

	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	ht := hookTests[0]
	assert.False(t, ht.Pending)
	assert.True(t, ht.Passed)
	assert.False(t, ht.Failed)
	assert.Equal(t, 42*time.Millisecond, ht.Duration)
	assert.Equal(t, "hello\n", ht.Stdout)
	assert.Empty(t, ht.Message)
}

func TestFinishHookTest_Failure(t *testing.T) {
	resetHookTestsState(t)
	idx := appendRunningHookTest("post", "notify", "curl webhook")
	finishHookTest(idx, 100*time.Millisecond, "err: 500\n", fmt.Errorf("exit status 1"))

	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	ht := hookTests[0]
	assert.False(t, ht.Pending)
	assert.False(t, ht.Passed)
	assert.True(t, ht.Failed)
	assert.Equal(t, "exit status 1", ht.Message)
	assert.Equal(t, "err: 500\n", ht.Stdout)
}

func TestMergeHooksWithTests_PrependsHooks(t *testing.T) {
	resetHookTestsState(t)
	// Simulate a running pre hook plus two real tests arriving from the
	// testrunner. The merged view must put hooks first so they stay
	// visible at the top of the UI.
	appendRunningHookTest("pre", "deps", "make deps")
	finishHookTest(0, time.Second, "", nil)

	batch := []parsers.Test{
		{Name: "TestA", Package: "pkg/a", Passed: true},
		{Name: "TestB", Package: "pkg/b", Failed: true},
	}
	merged := mergeHooksWithTests(batch)
	require.Len(t, merged, 3)
	assert.Equal(t, "deps", merged[0].Name)
	assert.Equal(t, "hook:pre", merged[0].Package)
	assert.Equal(t, "TestA", merged[1].Name)
	assert.Equal(t, "TestB", merged[2].Name)
}

func TestMergeHooksWithTests_EmptyHooksReturnsBatchUnchanged(t *testing.T) {
	resetHookTestsState(t)
	batch := []parsers.Test{{Name: "TestA", Passed: true}}
	merged := mergeHooksWithTests(batch)
	require.Len(t, merged, 1)
	assert.Equal(t, "TestA", merged[0].Name)
}
