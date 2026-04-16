package ui

import (
	"encoding/json"
	"testing"
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
