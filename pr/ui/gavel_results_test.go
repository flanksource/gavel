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
