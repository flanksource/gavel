package ui

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/flanksource/gavel/github"
)

func TestComputeGavelSummary_ObjectFormat(t *testing.T) {
	input := `{
		"tests": [
			{"name": "TestA", "passed": true},
			{"name": "TestB", "failed": true},
			{"name": "TestC", "skipped": true},
			{"name": "TestD", "passed": true}
		],
		"lint": [
			{
				"linter": "golangci-lint",
				"success": false,
				"violations": [
					{"file": "a.go", "line": 1, "message": "err"},
					{"file": "b.go", "line": 2, "message": "err2"}
				]
			},
			{
				"linter": "eslint",
				"success": true,
				"violations": []
			},
			{
				"linter": "ruff",
				"skipped": true,
				"violations": []
			}
		]
	}`
	s := computeGavelSummary([]byte(input), 42, "https://example.com/art/42")
	if s.Error != "" {
		t.Fatalf("unexpected error: %s", s.Error)
	}
	if s.ArtifactID != 42 {
		t.Errorf("ArtifactID = %d, want 42", s.ArtifactID)
	}
	if s.TestsTotal != 4 {
		t.Errorf("TestsTotal = %d, want 4", s.TestsTotal)
	}
	if s.TestsPassed != 2 {
		t.Errorf("TestsPassed = %d, want 2", s.TestsPassed)
	}
	if s.TestsFailed != 1 {
		t.Errorf("TestsFailed = %d, want 1", s.TestsFailed)
	}
	if s.TestsSkipped != 1 {
		t.Errorf("TestsSkipped = %d, want 1", s.TestsSkipped)
	}
	// ruff is skipped so not counted
	if s.LintLinters != 2 {
		t.Errorf("LintLinters = %d, want 2", s.LintLinters)
	}
	if s.LintViolations != 2 {
		t.Errorf("LintViolations = %d, want 2", s.LintViolations)
	}
	if len(s.TopFailures) != 1 || s.TopFailures[0].Name != "TestB" {
		t.Errorf("TopFailures = %+v, want single TestB failure", s.TopFailures)
	}
	if len(s.TopLintViolations) != 2 {
		t.Errorf("TopLintViolations = %d, want 2", len(s.TopLintViolations))
	}
	if s.TopLintViolations[0].Linter != "golangci-lint" || s.TopLintViolations[0].File != "a.go" {
		t.Errorf("first lint violation = %+v", s.TopLintViolations[0])
	}
}

func TestComputeGavelSummary_TopFailuresCap(t *testing.T) {
	// 7 failing tests — summary must cap at 5 and preserve encounter order.
	tests := `[
		{"name": "F1", "failed": true},
		{"name": "F2", "failed": true},
		{"name": "F3", "failed": true},
		{"name": "F4", "failed": true},
		{"name": "F5", "failed": true},
		{"name": "F6", "failed": true},
		{"name": "F7", "failed": true}
	]`
	s := computeGavelSummary([]byte(tests), 1, "")
	if s.TestsFailed != 7 {
		t.Errorf("TestsFailed = %d, want 7", s.TestsFailed)
	}
	if len(s.TopFailures) != 5 {
		t.Fatalf("TopFailures length = %d, want 5", len(s.TopFailures))
	}
	for i, want := range []string{"F1", "F2", "F3", "F4", "F5"} {
		if s.TopFailures[i].Name != want {
			t.Errorf("TopFailures[%d] = %s, want %s", i, s.TopFailures[i].Name, want)
		}
	}
}

func TestComputeGavelSummary_ArrayFormat(t *testing.T) {
	input := `[
		{"name": "TestX", "passed": true},
		{"name": "TestY", "failed": true},
		{
			"name": "pkg/",
			"children": [
				{"name": "TestZ", "passed": true}
			]
		}
	]`
	s := computeGavelSummary([]byte(input), 99, "")
	if s.Error != "" {
		t.Fatalf("unexpected error: %s", s.Error)
	}
	if s.TestsTotal != 3 {
		t.Errorf("TestsTotal = %d, want 3", s.TestsTotal)
	}
	if s.TestsPassed != 2 {
		t.Errorf("TestsPassed = %d, want 2", s.TestsPassed)
	}
	if s.TestsFailed != 1 {
		t.Errorf("TestsFailed = %d, want 1", s.TestsFailed)
	}
	if s.LintLinters != 0 {
		t.Errorf("LintLinters = %d, want 0", s.LintLinters)
	}
	if s.HasBench {
		t.Error("HasBench should be false")
	}
}

func TestComputeGavelSummary_WithBench(t *testing.T) {
	input := `{
		"tests": [],
		"bench": {
			"threshold": 5.0,
			"deltas": [
				{"name": "BenchA", "delta_pct": 10.5, "significant": true},
				{"name": "BenchB", "delta_pct": 2.0, "significant": true},
				{"name": "BenchC", "delta_pct": -3.0, "significant": true}
			],
			"geomean_delta": 3.2,
			"has_regression": true
		}
	}`
	s := computeGavelSummary([]byte(input), 1, "")
	if s.Error != "" {
		t.Fatalf("unexpected error: %s", s.Error)
	}
	if !s.HasBench {
		t.Error("HasBench should be true")
	}
	if s.BenchRegressions != 1 {
		t.Errorf("BenchRegressions = %d, want 1 (only BenchA exceeds threshold 5.0)", s.BenchRegressions)
	}
}

func TestComputeGavelSummary_InvalidJSON(t *testing.T) {
	s := computeGavelSummary([]byte(`{invalid`), 1, "")
	if s.Error == "" {
		t.Error("expected error for invalid JSON")
	}
}

func TestGavelResultJSON_UnmarshalArray(t *testing.T) {
	var g gavelResultJSON
	if err := json.Unmarshal([]byte(`[{"name":"T1","passed":true}]`), &g); err != nil {
		t.Fatal(err)
	}
	if len(g.Tests) != 1 || g.Tests[0].Name != "T1" {
		t.Errorf("got %+v", g.Tests)
	}
}

func TestGavelResultJSON_UnmarshalObject(t *testing.T) {
	var g gavelResultJSON
	if err := json.Unmarshal([]byte(`{"tests":[{"name":"T2"}],"lint":[]}`), &g); err != nil {
		t.Fatal(err)
	}
	if len(g.Tests) != 1 || g.Tests[0].Name != "T2" {
		t.Errorf("got %+v", g.Tests)
	}
}

// TestSnapshotIncludesGavelResults ensures that setGavelSummary cached
// entries surface in snapshotLocked's payload, but only for PRs still
// present in the current snapshot (prevents unbounded growth when PRs
// close/drop out of the view).
func TestSnapshotIncludesGavelResults(t *testing.T) {
	s := &Server{
		prs: github.PRSearchResults{
			{Repo: "owner/a", Number: 1},
			{Repo: "owner/b", Number: 2},
		},
		gavelCache: map[string]*GavelResultsSummary{},
	}

	s.setGavelSummary("owner/a", 1, &GavelResultsSummary{TestsTotal: 3, TestsFailed: 1, TestsPassed: 2})
	// "owner/c" is not in the current snapshot — the stale cache entry
	// must NOT leak into the wire payload.
	s.setGavelSummary("owner/c", 9, &GavelResultsSummary{TestsTotal: 1, TestsPassed: 1})

	s.mu.RLock()
	snap := s.snapshotLocked()
	s.mu.RUnlock()

	if len(snap.GavelResults) != 1 {
		t.Fatalf("expected 1 entry in snapshot.GavelResults, got %d (%v)", len(snap.GavelResults), snap.GavelResults)
	}
	got, ok := snap.GavelResults["owner/a#1"]
	if !ok {
		t.Fatalf("missing owner/a#1 entry: %v", snap.GavelResults)
	}
	if got.TestsFailed != 1 || got.TestsPassed != 2 {
		t.Errorf("cached summary lost fields: %+v", got)
	}
	if _, leaked := snap.GavelResults["owner/c#9"]; leaked {
		t.Errorf("stale cache entry leaked into snapshot: %v", snap.GavelResults)
	}

	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Contains(b, []byte(`"gavelResults"`)) {
		t.Errorf("marshaled snapshot missing gavelResults field: %s", b)
	}
}

// TestMergeGavelResultsConcatenates exercises the per-job merge: when one job
// uploads multiple gavel-* artifacts (e.g. matrix shards or a separate bench
// JSON), all of their tests/lint/bench data must aggregate into one logical
// payload. Pass + Fail + Skip counts must equal the sum across inputs.
func TestMergeGavelResultsConcatenates(t *testing.T) {
	a := gavelResultJSON{}
	if err := json.Unmarshal([]byte(`{"tests":[{"name":"A1","passed":true},{"name":"A2","failed":true}]}`), &a); err != nil {
		t.Fatal(err)
	}
	b := gavelResultJSON{}
	if err := json.Unmarshal([]byte(`{"tests":[{"name":"B1","passed":true}],"lint":[{"linter":"lint1","success":false,"violations":[{"file":"x.go","line":1,"message":"x"}]}]}`), &b); err != nil {
		t.Fatal(err)
	}

	merged := mergeGavelResults(a, b)
	if len(merged.Tests) != 3 {
		t.Errorf("merged tests = %d, want 3", len(merged.Tests))
	}
	if len(merged.Lint) != 1 {
		t.Errorf("merged lint = %d, want 1", len(merged.Lint))
	}

	summary := &GavelResultsSummary{}
	applyResultsToSummary(merged, summary)
	if summary.TestsPassed != 2 || summary.TestsFailed != 1 || summary.TestsTotal != 3 {
		t.Errorf("summary counts wrong: passed=%d failed=%d total=%d",
			summary.TestsPassed, summary.TestsFailed, summary.TestsTotal)
	}
	if summary.LintViolations != 1 {
		t.Errorf("LintViolations = %d, want 1", summary.LintViolations)
	}
}

// TestMergeGavelResultsKeepsFirstBench documents that bench comparisons are
// not summed across artifacts — the first non-nil one wins. Two artifacts
// each carrying their own bench struct is rare, but if it happens we want a
// deterministic answer (no double-counting regressions, no panic).
func TestMergeGavelResultsKeepsFirstBench(t *testing.T) {
	a := gavelResultJSON{}
	b := gavelResultJSON{}
	if err := json.Unmarshal([]byte(`{"bench":{"threshold":1.05}}`), &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"bench":{"threshold":2.0}}`), &b); err != nil {
		t.Fatal(err)
	}
	merged := mergeGavelResults(a, b)
	if merged.Bench == nil || merged.Bench.Threshold != 1.05 {
		t.Errorf("expected first bench to win (threshold=1.05), got %+v", merged.Bench)
	}
}

// TestJobNameFromArtifactsDistinct asserts that distinct artifact names from
// the same job are joined into a comma-separated label so the UI can show
// "gavel-results, gavel-bench" instead of just "gavel-results".
func TestJobNameFromArtifactsDistinct(t *testing.T) {
	got := jobNameFromArtifacts([]github.GavelArtifact{
		{Name: "gavel-results"},
		{Name: "gavel-bench"},
		{Name: "gavel-results"}, // duplicate must be deduped
	})
	if got != "gavel-results, gavel-bench" {
		t.Errorf("jobNameFromArtifacts = %q, want %q", got, "gavel-results, gavel-bench")
	}
}

// TestSummarizeGavelArtifactsGroupsByRun verifies that the PR-level rollup
// groups artifacts from the same workflow run into a single GavelJobSummary
// entry. We can't exercise the network-fetch path here (DownloadArtifactFiles
// would call out to GitHub), so we use an empty artifact set per run — the
// grouping/RunID handling is what matters.
func TestSummarizeGavelArtifactsGroupsByRun(t *testing.T) {
	// Fake out the github layer by passing only metadata: the function will
	// try to download and fail, recording an error on each job. That's
	// enough to assert the grouping shape.
	arts := []github.GavelArtifact{
		{Name: "gavel-results", ID: 1, RunID: 100, URL: "u1"},
		{Name: "gavel-bench", ID: 2, RunID: 100, URL: "u2"},
		{Name: "gavel-results", ID: 3, RunID: 200, URL: "u3"},
	}
	// Without a token, DownloadArtifactFiles errors immediately at
	// opts.token() — no network call. The merge logic still has to walk
	// every group and produce the right number of Job entries.
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	pr := summarizeGavelArtifacts(github.Options{Repo: "o/r"}, arts)
	if pr == nil {
		t.Fatal("summarizeGavelArtifacts returned nil")
	}
	if len(pr.Jobs) != 2 {
		t.Fatalf("Jobs = %d, want 2 (one per RunID)", len(pr.Jobs))
	}
	// Job for RunID 100 must merge both artifact IDs.
	var run100 *GavelJobSummary
	for i := range pr.Jobs {
		if pr.Jobs[i].RunID == 100 {
			run100 = &pr.Jobs[i]
		}
	}
	if run100 == nil {
		t.Fatalf("missing RunID 100 in Jobs: %+v", pr.Jobs)
	}
	if len(run100.ArtifactIDs) != 2 {
		t.Errorf("RunID 100 ArtifactIDs = %v, want 2 entries", run100.ArtifactIDs)
	}
	if run100.JobName != "gavel-results, gavel-bench" {
		t.Errorf("RunID 100 JobName = %q", run100.JobName)
	}
	// Each job recorded an error (no token); the PR-level counts should
	// therefore be zero rather than misleading partials.
	if pr.TestsTotal != 0 {
		t.Errorf("expected 0 tests (all downloads failed), got %d", pr.TestsTotal)
	}
}

// TestSummarizeGavelArtifactsEmpty asserts that an empty input returns nil
// so the caller can short-circuit emit("gavel", ...) and not show an empty
// "Gavel results" section in the UI.
func TestSummarizeGavelArtifactsEmpty(t *testing.T) {
	if got := summarizeGavelArtifacts(github.Options{}, nil); got != nil {
		t.Errorf("expected nil for empty input, got %+v", got)
	}
}
