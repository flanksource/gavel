package snapshots

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// run start times, distinct seconds so SavePerRun lands them in separate files.
var (
	tTest     = time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	tLint     = time.Date(2026, 6, 28, 11, 0, 0, 0, time.UTC)
	tTestLint = time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
)

func leafTests() []parsers.Test {
	return []parsers.Test{
		{Name: "pass-a", Passed: true},
		{Name: "fail-b", Failed: true},
		{Name: "skip-c", Skipped: true},
	}
}

func lintResults(violations int) []*linters.LinterResult {
	vs := make([]models.Violation, violations)
	return []*linters.LinterResult{{Linter: "golangci-lint", Violations: vs}}
}

func writeRun(t *testing.T, workDir string, snap testui.Snapshot, started time.Time) {
	t.Helper()
	if _, err := SavePerRun(workDir, &snap, started); err != nil {
		t.Fatalf("SavePerRun(%s): %v", started, err)
	}
}

func seedRuns(t *testing.T, workDir string) {
	t.Helper()
	writeRun(t, workDir, testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: tTest, Ended: tTest, Frameworks: []string{"go test"}},
		Git:      &testui.SnapshotGit{Repo: "gavel", SHA: "abc123"},
		Tests:    leafTests(),
	}, tTest)

	writeRun(t, workDir, testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: tLint, Ended: tLint},
		Status:   testui.SnapshotStatus{LintRun: true},
		Lint:     lintResults(2),
	}, tLint)

	writeRun(t, workDir, testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: tTestLint, Ended: tTestLint, Frameworks: []string{"ginkgo"}},
		Status:   testui.SnapshotStatus{LintRun: true},
		Tests:    leafTests(),
		Lint:     lintResults(1),
	}, tTestLint)
}

func TestListRunsClassifiesAndCountsNewestFirst(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)

	runs, err := ListRuns(workDir, time.Time{})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("want 3 runs, got %d", len(runs))
	}

	// newest-first: test+lint (12:00), lint (11:00), test (10:00).
	wantKinds := []string{RunKindTestLint, RunKindLint, RunKindTest}
	for i, want := range wantKinds {
		if runs[i].Kind != want {
			t.Errorf("runs[%d].Kind = %q, want %q", i, runs[i].Kind, want)
		}
	}
	if !runs[0].Started.Equal(tTestLint) {
		t.Errorf("runs[0].Started = %v, want %v", runs[0].Started, tTestLint)
	}

	// leaf roll-up: 1 pass / 1 fail / 1 skip across the three leaves.
	testRun := runs[2]
	if testRun.Passed != 1 || testRun.Failed != 1 || testRun.Skipped != 1 || testRun.Total != 3 {
		t.Errorf("test run counts = %+v, want passed=1 failed=1 skipped=1 total=3", testRun)
	}
	if testRun.Repo != "gavel" || testRun.SHA != "abc123" {
		t.Errorf("test run git = repo %q sha %q, want gavel/abc123", testRun.Repo, testRun.SHA)
	}

	lintRun := runs[1]
	if lintRun.LintLinters != 1 || lintRun.LintViolations != 2 {
		t.Errorf("lint run = %d linters / %d violations, want 1/2", lintRun.LintLinters, lintRun.LintViolations)
	}
	if lintRun.Total != 0 {
		t.Errorf("lint run Total = %d, want 0 (no tests)", lintRun.Total)
	}
}

func TestListRunsSinceWatermarkSkipsOlderOrEqual(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)

	runs, err := ListRuns(workDir, tTest)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	// since=tTest excludes the run started exactly at tTest (watermark is
	// exclusive), leaving the lint and test+lint runs.
	if len(runs) != 2 {
		t.Fatalf("want 2 runs after watermark, got %d", len(runs))
	}
	for _, r := range runs {
		if !r.Started.After(tTest) {
			t.Errorf("run %s started %v is not after watermark %v", r.RunID, r.Started, tTest)
		}
	}
}

func TestListRunsMissingGavelDirIsEmpty(t *testing.T) {
	runs, err := ListRuns(t.TempDir(), time.Time{})
	if err != nil {
		t.Fatalf("ListRuns on empty workdir: %v", err)
	}
	if runs != nil {
		t.Errorf("want nil runs for missing .gavel, got %v", runs)
	}
}
