package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/history"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTestHistoryUsesGlobalCWD(t *testing.T) {
	workDir := t.TempDir()
	prev := workingDir
	workingDir = workDir
	t.Cleanup(func() { workingDir = prev })

	started := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	_, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: started},
		Git:      &testui.SnapshotGit{Root: workDir, Repo: "repo", SHA: "abc"},
		Tests: []parsers.Test{{
			Framework:   parsers.GoTest,
			PackagePath: "./pkg/foo",
			File:        filepath.Join(workDir, "pkg/foo/foo_test.go"),
			Name:        "TestFoo",
			Passed:      true,
			Duration:    time.Millisecond,
		}},
	}, started)
	require.NoError(t, err)

	got, err := runTestHistory(testHistoryOptions{Paths: []string{"pkg/foo"}})
	require.NoError(t, err)

	report, ok := got.(*history.Report)
	require.True(t, ok)
	assert.Equal(t, workDir, report.WorkDir)
	require.Len(t, report.Tests, 1)
	assert.Equal(t, "TestFoo", report.Tests[0].Name)
}

func TestTestHistoryCommandRegistered(t *testing.T) {
	cmd, _, err := testCmd.Find([]string{"history"})
	require.NoError(t, err)
	require.NotNil(t, cmd)
	assert.Equal(t, "history", cmd.Name())
}
