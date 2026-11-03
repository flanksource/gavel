package snapshots

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/traces"
)

// run start times, distinct seconds so SavePerRun lands them in separate files.
var (
	tTest     = time.Date(2026, 6, 28, 10, 0, 0, 0, time.UTC)
	tLint     = time.Date(2026, 6, 28, 11, 0, 0, 0, time.UTC)
	tTestLint = time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC)
	// tCorrupt is before tTest so a tTest watermark excludes it.
	tCorrupt = time.Date(2026, 6, 28, 9, 0, 0, 0, time.UTC)
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
	if _, err := SavePerRun(workDir, &snap, started, ""); err != nil {
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

	runs, err := ListRuns(workDir, ListRunsOptions{})
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

	runs, err := ListRuns(workDir, ListRunsOptions{Since: tTest})
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

// A run at or before the watermark must be skipped from its dirent alone,
// without being opened: an incremental sweep re-reads the whole .gavel
// directory, and parsing every snapshot only to discard it by start time made an
// incremental sweep cost the same as a full scan.
func TestListRunsSkipsOlderRunsWithoutReadingThem(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)

	// ListRuns reports a parse failure as an error, so corrupting the oldest run
	// turns "was it opened?" into an observable outcome.
	corrupt, err := SavePerRun(workDir, &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{Started: tCorrupt, Ended: tCorrupt},
	}, tCorrupt, "")
	if err != nil {
		t.Fatalf("SavePerRun: %v", err)
	}
	if err := os.WriteFile(corrupt, []byte("{ not valid json"), 0o600); err != nil {
		t.Fatalf("corrupt run file: %v", err)
	}
	// The mtime is the gate ListRuns reads, and writing the file just now set it
	// to the present. Restore it to the run's own time, which is what a real
	// run-*.json carries.
	if err := os.Chtimes(corrupt, tCorrupt, tCorrupt); err != nil {
		t.Fatalf("set run file mtime: %v", err)
	}

	runs, err := ListRuns(workDir, ListRunsOptions{Since: tTest})
	if err != nil {
		t.Fatalf("ListRuns opened a run at or before the watermark: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("want 2 runs after watermark, got %d", len(runs))
	}

	// Without this the assertion above would also pass if the corrupt file were
	// simply being ignored rather than skipped by start time.
	if _, err := ListRuns(workDir, ListRunsOptions{}); err == nil {
		t.Fatal("a full scan must still fail on the corrupt run file")
	}
}

// A failed package or suite carries the failure of the test under it; only the
// test itself counts.
func TestListRunsCountsLeafTestsOnly(t *testing.T) {
	workDir := t.TempDir()
	writeRun(t, workDir, testui.Snapshot{Tests: []parsers.Test{{
		Name: "pkg/", Failed: true,
		Children: []parsers.Test{{Name: "Suite", Failed: true, Children: leafTests()}},
	}}}, tTest)

	runs, err := ListRuns(workDir, ListRunsOptions{})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if got := runs[0]; got.Total != 3 || got.Failed != 1 || got.Passed != 1 || got.Skipped != 1 {
		t.Errorf("counts = total %d failed %d passed %d skipped %d, want 3/1/1/1", got.Total, got.Failed, got.Passed, got.Skipped)
	}
}

// A listed run's summary is kept in .gavel/runs-index.json, so a run is decoded
// once: run files are immutable and some hold gigabytes of test output.
func TestListRunsReadsSummariesFromTheIndex(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)
	first, err := ListRuns(workDir, ListRunsOptions{})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}

	// Same size and mtime, unreadable content: only the index can answer.
	path := filepath.Join(workDir, Dir, traces.RunID(tTest)+".json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat run: %v", err)
	}
	garbage := make([]byte, info.Size())
	if err := os.WriteFile(path, garbage, 0o600); err != nil {
		t.Fatalf("overwrite run: %v", err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatalf("restore mtime: %v", err)
	}

	second, err := ListRuns(workDir, ListRunsOptions{})
	if err != nil {
		t.Fatalf("ListRuns re-decoded an indexed run: %v", err)
	}
	if !slices.EqualFunc(first, second, func(a, b RunInfo) bool { return a.RunID == b.RunID && a.Failed == b.Failed && a.Kind == b.Kind }) {
		t.Errorf("indexed listing = %+v, want %+v", second, first)
	}

	// A run whose size changed is not the one indexed, so it is read again.
	if err := os.WriteFile(path, []byte("{ not valid json"), 0o600); err != nil {
		t.Fatalf("rewrite run: %v", err)
	}
	if _, err := ListRuns(workDir, ListRunsOptions{}); err == nil {
		t.Fatal("a run that changed since it was indexed must be read again")
	}
}

// An index written for another indexVersion holds summaries derived another
// way, so its runs are summarized again.
func TestListRunsRecomputesAnIndexOfAnotherVersion(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)
	if _, err := ListRuns(workDir, ListRunsOptions{}); err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	indexPath := filepath.Join(workDir, Dir, IndexName)
	var file indexFile
	data, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("decode index: %v", err)
	}
	const staleFailed = 99
	for name, entry := range file.Runs {
		entry.Info.Failed = staleFailed
		file.Runs[name] = entry
	}
	file.Version = indexVersion - 1
	if data, err = json.Marshal(file); err != nil {
		t.Fatalf("encode index: %v", err)
	}
	if err := os.WriteFile(indexPath, data, 0o600); err != nil {
		t.Fatalf("write index: %v", err)
	}

	runs, err := ListRuns(workDir, ListRunsOptions{})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	for _, run := range runs {
		if run.Failed == staleFailed {
			t.Errorf("run %s served a summary from an index of another version", run.RunID)
		}
	}
}

// Paging walks newest-first by RunID: Limit caps a page, Before resumes after
// the oldest RunID of the previous page.
func TestListRunsPagesNewestFirstByRunID(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)

	first, err := ListRuns(workDir, ListRunsOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ListRuns page 1: %v", err)
	}
	second, err := ListRuns(workDir, ListRunsOptions{Limit: 2, Before: traces.RunID(tLint)})
	if err != nil {
		t.Fatalf("ListRuns page 2: %v", err)
	}
	ids := func(runs []RunInfo) []string {
		var out []string
		for _, r := range runs {
			out = append(out, r.RunID)
		}
		return out
	}
	if want := []string{traces.RunID(tTestLint), traces.RunID(tLint)}; !slices.Equal(ids(first), want) {
		t.Errorf("page 1 = %v, want %v", ids(first), want)
	}
	if want := []string{traces.RunID(tTest)}; !slices.Equal(ids(second), want) {
		t.Errorf("page 2 = %v, want %v", ids(second), want)
	}
}

func TestLatestRunsCombinedRunCoversTestAndLint(t *testing.T) {
	workDir := t.TempDir()
	seedRuns(t, workDir)

	paths, err := LatestRuns(workDir)
	if err != nil {
		t.Fatalf("LatestRuns: %v", err)
	}
	want := filepath.Join(workDir, Dir, traces.RunID(tTestLint)+".json")
	if len(paths) != 1 || paths[0] != want {
		t.Fatalf("LatestRuns = %v, want [%s]", paths, want)
	}
}

// The newest test run and the newest lint run are separate files; older runs
// past both must not be opened, since .gavel holds hundreds of multi-MB runs.
func TestLatestRunsStopsOnceTestAndLintFound(t *testing.T) {
	workDir := t.TempDir()
	tOlder := tTest.Add(-time.Hour)
	corrupt, err := SavePerRun(workDir, &testui.Snapshot{}, tOlder, "")
	if err != nil {
		t.Fatalf("SavePerRun: %v", err)
	}
	if err := os.WriteFile(corrupt, []byte("{ not valid json"), 0o600); err != nil {
		t.Fatalf("corrupt run file: %v", err)
	}
	writeRun(t, workDir, testui.Snapshot{Tests: leafTests()}, tTest)
	writeRun(t, workDir, testui.Snapshot{Status: testui.SnapshotStatus{LintRun: true}, Lint: lintResults(1)}, tLint)
	// An empty run (aborted before any test or linter ran) is neither.
	writeRun(t, workDir, testui.Snapshot{}, tTestLint)

	paths, err := LatestRuns(workDir)
	if err != nil {
		t.Fatalf("LatestRuns opened a run older than the latest test and lint runs: %v", err)
	}
	want := []string{
		filepath.Join(workDir, Dir, traces.RunID(tLint)+".json"),
		filepath.Join(workDir, Dir, traces.RunID(tTest)+".json"),
	}
	if !slices.Equal(paths, want) {
		t.Fatalf("LatestRuns = %v, want %v", paths, want)
	}
}

func TestLatestRunsMissingGavelDirIsEmpty(t *testing.T) {
	paths, err := LatestRuns(t.TempDir())
	if err != nil {
		t.Fatalf("LatestRuns on empty workdir: %v", err)
	}
	if paths != nil {
		t.Errorf("want nil paths for missing .gavel, got %v", paths)
	}
}

func TestListRunsMissingGavelDirIsEmpty(t *testing.T) {
	runs, err := ListRuns(t.TempDir(), ListRunsOptions{})
	if err != nil {
		t.Fatalf("ListRuns on empty workdir: %v", err)
	}
	if runs != nil {
		t.Errorf("want nil runs for missing .gavel, got %v", runs)
	}
}
