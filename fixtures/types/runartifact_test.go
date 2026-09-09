package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runArtifactTests(failures int) []parsers.Test {
	tests := []parsers.Test{
		{Name: "pass-a", Suite: []string{"suite"}, Passed: true},
		{Name: "skip-b", Skipped: true},
		{Name: "warn-c", Warned: true},
	}
	for i := 0; i < failures; i++ {
		tests = append(tests, parsers.Test{
			Name:    fmt.Sprintf("fail-%d", i),
			Suite:   []string{"suite"},
			Failed:  true,
			Message: fmt.Sprintf("boom %d", i),
		})
	}
	return tests
}

func msg(s string) *string { return &s }

func TestSaveRunArtifactWritesSnapshotAndCounts(t *testing.T) {
	workDir := t.TempDir()
	started := time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir,
		Kind:    "test",
		Label:   "Verify DoD",
		Started: started,
		Tests:   runArtifactTests(2),
	})
	require.NoError(t, err)
	require.NotNil(t, artifact)

	assert.Equal(t, "test", artifact.Kind)
	assert.Equal(t, "run-2026-07-30T09-00-00Z-verify-dod", artifact.RunID)
	assert.Equal(t, filepath.Join(workDir, snapshots.Dir, artifact.RunID+".json"), artifact.Path)

	// counts mirror parsers.Tests.Sum()
	want := parsers.Tests(runArtifactTests(2)).Sum()
	assert.Equal(t, want.Total, artifact.Total)
	assert.Equal(t, want.Passed, artifact.Passed)
	assert.Equal(t, want.Failed, artifact.Failed)
	assert.Equal(t, want.Warned, artifact.Warned)
	assert.Equal(t, want.Skipped, artifact.Skipped)

	require.Len(t, artifact.Failures, 2)
	assert.Equal(t, "fail-0", artifact.Failures[0].Name)
	assert.Equal(t, "suite", artifact.Failures[0].Suite)
	assert.Equal(t, "failed", artifact.Failures[0].Status)
	assert.Equal(t, "boom 0", artifact.Failures[0].Message)
	assert.Zero(t, artifact.Truncated)

	data, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)
	var snap testui.Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))
	assert.Len(t, snap.Tests, len(runArtifactTests(2)))
	require.NotNil(t, snap.Metadata)
	assert.Equal(t, "test", snap.Metadata.Kind)
	assert.True(t, snap.Metadata.Started.Equal(started))
	assert.False(t, snap.Status.Running)
	assert.False(t, snap.Status.LintRun)
}

func TestSaveRunArtifactDoesNotWritePointers(t *testing.T) {
	workDir := t.TempDir()

	_, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir,
		Kind:    "test",
		Started: time.Now(),
		Tests:   runArtifactTests(1),
	})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(workDir, snapshots.Dir, "last.json"))
	assert.True(t, os.IsNotExist(err), "a verification run must not overwrite last.json")
}

func TestSaveRunArtifactCapsRecordedFailures(t *testing.T) {
	workDir := t.TempDir()
	overflow := 4

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir,
		Kind:    "test",
		Started: time.Now(),
		Tests:   runArtifactTests(maxRecordedFailures + overflow),
	})
	require.NoError(t, err)

	assert.Len(t, artifact.Failures, maxRecordedFailures)
	assert.Equal(t, overflow, artifact.Truncated)
	assert.Equal(t, maxRecordedFailures+overflow, artifact.Failed, "counts stay complete even when the list is capped")
}

func TestSaveRunArtifactRecordsLint(t *testing.T) {
	workDir := t.TempDir()

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir,
		Kind:    "lint",
		Started: time.Now(),
		Lint: []*linters.LinterResult{
			{Linter: "golangci-lint", Violations: []models.Violation{
				{File: "a.go", Line: 12, Message: msg("unused variable")},
			}},
			{Linter: "eslint"},
			{Linter: "ruff", Skipped: true},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, 1, artifact.LintViolations)
	assert.Equal(t, []string{"golangci-lint", "eslint"}, artifact.Linters, "skipped linters are not reported as run")
	require.Len(t, artifact.Failures, 1)
	assert.Equal(t, "golangci-lint a.go:12", artifact.Failures[0].Name)
	assert.Equal(t, "unused variable", artifact.Failures[0].Message)

	data, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)
	var snap testui.Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))
	assert.True(t, snap.Status.LintRun)
	assert.Len(t, snap.Lint, 3)
}

func TestSaveRunArtifactRecordsEngineError(t *testing.T) {
	workDir := t.TempDir()

	artifact, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir,
		Kind:    "test",
		Started: time.Now(),
		Err:     errors.New("build failed: package foo does not compile"),
	})
	require.NoError(t, err, "a crashed engine still records why there are no results")

	assert.Equal(t, "build failed: package foo does not compile", artifact.Error)
	assert.Zero(t, artifact.Total)

	data, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)
	var snap testui.Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))
	assert.Equal(t, "build failed: package foo does not compile", snap.Error)
}

func TestSaveRunArtifactRequiresWorkDir(t *testing.T) {
	_, err := saveRunArtifact(runArtifactOptions{Kind: "test", Started: time.Now()})
	require.Error(t, err, "an unwritable artifact must fail the step loudly, not be dropped")
}

func TestSaveRunArtifactConcurrentStepsDoNotCollide(t *testing.T) {
	workDir := t.TempDir()
	started := time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC)

	first, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir, Kind: "test", Label: "step", Started: started, Tests: runArtifactTests(0),
	})
	require.NoError(t, err)
	second, err := saveRunArtifact(runArtifactOptions{
		WorkDir: workDir, Kind: "test", Label: "step", Started: started, Tests: runArtifactTests(1),
	})
	require.NoError(t, err)

	assert.NotEqual(t, first.RunID, second.RunID, "two steps starting in the same second must not overwrite each other")
}
