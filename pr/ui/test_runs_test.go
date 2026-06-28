package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGroupRuns(t *testing.T) {
	projects := []Project{
		{Name: "with-runs", Dir: "/ws/a"},
		{Name: "no-runs", Dir: "/ws/b"},
	}
	byDir := map[string][]testRunView{
		"/ws/a": {{RunID: "run-1", Kind: "test", Total: 6, Passed: 5, Failed: 1}},
	}

	got := groupRuns(projects, byDir)
	if len(got) != len(projects) {
		t.Fatalf("groupRuns returned %d projects, want %d", len(got), len(projects))
	}
	if len(got[0].Runs) != 1 || got[0].Runs[0].RunID != "run-1" {
		t.Errorf("with-runs runs = %+v, want one run-1", got[0].Runs)
	}

	// A project absent from byDir must get a non-nil empty slice: a nil slice
	// marshals to JSON null and crashes the client on `runs.length`.
	if got[1].Runs == nil {
		t.Fatalf("no-runs project got nil Runs; must be a non-nil slice")
	}
	b, err := json.Marshal(testRunsResponse{Projects: got})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), `"runs":null`) {
		t.Errorf("response serialized a null runs field: %s", b)
	}
}

func TestResolveRunPath(t *testing.T) {
	dir := t.TempDir()
	gavelDir := filepath.Join(dir, ".gavel")
	if err := os.MkdirAll(gavelDir, 0o755); err != nil {
		t.Fatalf("mkdir .gavel: %v", err)
	}
	const stem = "run-2026-06-28T10-43-42Z"
	runPath := filepath.Join(gavelDir, stem+".json")
	if err := os.WriteFile(runPath, []byte(`{"tests":[]}`), 0o644); err != nil {
		t.Fatalf("write run file: %v", err)
	}
	// A sibling secret outside .gavel that a traversal attempt would target.
	if err := os.WriteFile(filepath.Join(dir, "secret.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	t.Run("valid run stem resolves to the file", func(t *testing.T) {
		got, err := resolveRunPath(dir, stem)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != runPath {
			t.Errorf("path = %q, want %q", got, runPath)
		}
	})

	t.Run("missing run returns empty path, no error", func(t *testing.T) {
		got, err := resolveRunPath(dir, "run-9999-01-01T00-00-00Z")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "" {
			t.Errorf("path = %q, want empty for missing run", got)
		}
	})

	// Every malformed/hostile runId must be rejected as a bad request (error),
	// never resolved to a path — defends the endpoint against path traversal and
	// reading non-run files out of .gavel.
	for _, bad := range []string{
		"",
		"../secret",
		"run-../../etc/passwd",
		"sub/run-x",
		`run-x\..\secret`,
		"sha-abc123", // not a per-run file
		"last",       // pointer file, not a run
		"run-x/../secret",
	} {
		t.Run("rejects "+bad, func(t *testing.T) {
			got, err := resolveRunPath(dir, bad)
			if err == nil {
				t.Errorf("runId %q: want error, got path %q", bad, got)
			}
			if got != "" {
				t.Errorf("runId %q: want empty path, got %q", bad, got)
			}
		})
	}
}
