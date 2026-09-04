package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
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

// The Verification tab addresses a workspace by directory, not by stored
// project name — a TODO's run artifacts must be readable without the workspace
// being registered as a project.
func TestHandleTestRunAcceptsDir(t *testing.T) {
	dir := t.TempDir()
	const stem = "run-2026-07-30T09-00-00Z-verify"
	body := `{"tests":[{"name":"TestFoo","failed":true}]}`
	if err := os.MkdirAll(filepath.Join(dir, ".gavel"), 0o755); err != nil {
		t.Fatalf("mkdir .gavel: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gavel", stem+".json"), []byte(body), 0o644); err != nil {
		t.Fatalf("write run file: %v", err)
	}

	s := &Server{}
	call := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		s.handleTestRun(w, httptest.NewRequest(http.MethodGet, "/api/tests/run?"+query, nil))
		return w
	}

	t.Run("dir and runId stream the snapshot", func(t *testing.T) {
		w := call("dir=" + dir + "&runId=" + stem)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (%s)", w.Code, w.Body.String())
		}
		if got := w.Body.String(); got != body {
			t.Errorf("body = %q, want %q", got, body)
		}
	})

	// A todo surface on the server's default workspace sends `dir=` with no value;
	// that must resolve to the work dir, not read as "no dir supplied".
	t.Run("bare dir resolves to the server work dir", func(t *testing.T) {
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		w := httptest.NewRecorder()
		server.handleTestRun(w, httptest.NewRequest(http.MethodGet, "/api/tests/run?dir=&runId="+stem, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200 (%s)", w.Code, w.Body.String())
		}
		if got := w.Body.String(); got != body {
			t.Errorf("body = %q, want %q", got, body)
		}
	})

	t.Run("unknown run is 404", func(t *testing.T) {
		if w := call("dir=" + dir + "&runId=run-9999-01-01T00-00-00Z"); w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	// Ambiguity is a client bug, not something to resolve with a precedence rule.
	for name, query := range map[string]string{
		"neither project nor dir": "runId=" + stem,
		"both project and dir":    "project=demo&dir=" + dir + "&runId=" + stem,
	} {
		t.Run(name+" is 400", func(t *testing.T) {
			if w := call(query); w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (%s)", w.Code, w.Body.String())
			}
		})
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
