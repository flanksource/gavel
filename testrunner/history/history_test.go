package history

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAggregatesLeafExecutions(t *testing.T) {
	workDir := t.TempDir()
	ts1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	ts3 := time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)

	writeRun(t, workDir, ts1, []parsers.Test{
		{
			Name:     "pkg/foo/",
			Failed:   true,
			Children: []parsers.Test{fooTest(workDir, "TestFoo", true, false, 10*time.Millisecond)},
		},
		{PackagePath: "./pkg/foo", Name: "TestSkipped", Skipped: true},
		{PackagePath: "./pkg/foo", Name: "TestPending", Pending: true},
	})
	writeRun(t, workDir, ts2, []parsers.Test{
		fooTest(workDir, "TestFoo", false, true, 30*time.Millisecond),
		{
			Framework:   parsers.GoTest,
			PackagePath: "./pkg/bar",
			File:        filepath.Join(workDir, "pkg/bar/bar_test.go"),
			Name:        "TestTimeout",
			TimedOut:    true,
			Duration:    time.Second,
		},
	})
	writeRun(t, workDir, ts3, []parsers.Test{
		fooTest(workDir, "TestBar", true, false, 5*time.Millisecond),
	})

	report, err := Load(Options{WorkDir: workDir})
	require.NoError(t, err)
	require.Equal(t, 3, report.RunCount)
	require.Len(t, report.Tests, 3)
	assert.Equal(t, ts1, report.FirstRunAt)
	assert.Equal(t, ts3, report.LastRunAt)

	foo := findEntry(t, report, "TestFoo")
	assert.Equal(t, parsers.GoTest, foo.Framework)
	assert.Equal(t, cleanWorkDir(workDir), foo.WorkDir)
	assert.Equal(t, "pkg/foo", foo.PackagePath)
	assert.Equal(t, "pkg/foo/foo_test.go", foo.File)
	assert.Equal(t, []string{"Math"}, foo.Suite)
	assert.Equal(t, 2, foo.ExecutionCount)
	assert.Equal(t, 1, foo.PassCount)
	assert.Equal(t, 1, foo.FailCount)
	assert.Equal(t, 0.5, foo.PassRate)
	assert.Equal(t, 10*time.Millisecond, foo.MinDuration)
	assert.Equal(t, 20*time.Millisecond, foo.AvgDuration)
	assert.Equal(t, 30*time.Millisecond, foo.MaxDuration)
	assert.Equal(t, ts1, foo.AddedAt)
	assert.Equal(t, ts1, foo.LastPassedAt)
	assert.Equal(t, ts2, foo.LastFailedAt)

	timeout := findEntry(t, report, "TestTimeout")
	assert.Equal(t, 1, timeout.ExecutionCount)
	assert.Equal(t, 0, timeout.PassCount)
	assert.Equal(t, 1, timeout.FailCount)
	assert.Equal(t, 0.0, timeout.PassRate)
	assert.Equal(t, ts2, timeout.LastFailedAt)
}

func TestLoadFiltersByPackageAndFile(t *testing.T) {
	workDir := t.TempDir()
	ts := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	writeRun(t, workDir, ts, []parsers.Test{
		fooTest(workDir, "TestFoo", true, false, time.Millisecond),
		{
			Framework:   parsers.GoTest,
			PackagePath: "./pkg/bar",
			File:        filepath.Join(workDir, "pkg/bar/bar_test.go"),
			Name:        "TestBar",
			Passed:      true,
			Duration:    time.Millisecond,
		},
	})

	report, err := Load(Options{WorkDir: workDir, Paths: []string{"pkg/foo"}})
	require.NoError(t, err)
	require.Len(t, report.Tests, 1)
	assert.Equal(t, "TestFoo", report.Tests[0].Name)

	report, err = Load(Options{WorkDir: workDir, Paths: []string{"pkg/bar/bar_test.go"}})
	require.NoError(t, err)
	require.Len(t, report.Tests, 1)
	assert.Equal(t, "TestBar", report.Tests[0].Name)
}

func TestLoadUsesFilenameTimestampWhenMetadataMissing(t *testing.T) {
	workDir := t.TempDir()
	started := time.Date(2026, 4, 27, 10, 43, 42, 0, time.UTC)
	path, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
		Tests: []parsers.Test{fooTest(workDir, "TestFoo", true, false, time.Millisecond)},
	}, started)
	require.NoError(t, err)
	require.Contains(t, filepath.Base(path), "2026-04-27T10-43-42Z")

	report, err := Load(Options{WorkDir: workDir})
	require.NoError(t, err)
	assert.Equal(t, started, report.FirstRunAt)
	assert.Equal(t, started, report.Tests[0].AddedAt)
}

func TestLoadNoHistory(t *testing.T) {
	_, err := Load(Options{WorkDir: t.TempDir()})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoHistory))
}

func TestLoadCorruptSnapshotReportsPath(t *testing.T) {
	workDir := t.TempDir()
	dir := filepath.Join(workDir, snapshots.Dir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "run-2026-01-01T00-00-00Z.json")
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o644))

	_, err := Load(Options{WorkDir: workDir})
	require.Error(t, err)
	assert.Contains(t, err.Error(), path)
}

func TestReportPrettyRendersHistoryRows(t *testing.T) {
	workDir := t.TempDir()
	ts1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	writeRun(t, workDir, ts1, []parsers.Test{fooTest(workDir, "TestFoo", true, false, 10*time.Millisecond)})
	writeRun(t, workDir, ts2, []parsers.Test{fooTest(workDir, "TestFoo", false, true, 20*time.Millisecond)})

	report, err := Load(Options{WorkDir: workDir})
	require.NoError(t, err)
	out := clicky.MustFormat(report, clicky.FormatOptions{Pretty: true})

	for _, want := range []string{
		"Test history",
		"pkg/foo",
		"foo_test.go",
		"TestFoo",
		"exec 2",
		"pass 50%",
		"last pass 2026-01-01",
		"last fail 2026-01-02",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected rendered history to contain %q, got:\n%s", want, out)
		}
	}
}

func writeRun(t *testing.T, workDir string, started time.Time, tests []parsers.Test) {
	t.Helper()
	_, err := snapshots.SavePerRun(workDir, &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: started},
		Git:      &testui.SnapshotGit{Root: workDir, Repo: "repo", SHA: "abc"},
		Status:   testui.SnapshotStatus{Running: false},
		Tests:    tests,
	}, started)
	require.NoError(t, err)
}

func fooTest(workDir, name string, passed, failed bool, duration time.Duration) parsers.Test {
	return parsers.Test{
		Framework:   parsers.GoTest,
		PackagePath: "./pkg/foo",
		File:        filepath.Join(workDir, "pkg/foo/foo_test.go"),
		Line:        12,
		Suite:       []string{"Math"},
		Name:        name,
		Passed:      passed,
		Failed:      failed,
		Duration:    duration,
	}
}

func findEntry(t *testing.T, report *Report, name string) Entry {
	t.Helper()
	for _, entry := range report.Tests {
		if entry.Name == name {
			return entry
		}
	}
	t.Fatalf("entry %q not found in %#v", name, report.Tests)
	return Entry{}
}
