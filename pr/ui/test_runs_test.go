package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/snapshots"
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

func TestHandleTestRunReadsRunsByID(t *testing.T) {
	dir := t.TempDir()
	gavelDir := filepath.Join(dir, ".gavel")
	if err := os.MkdirAll(gavelDir, 0o755); err != nil {
		t.Fatalf("mkdir .gavel: %v", err)
	}
	const stem = "run-2026-06-28T10-43-42Z"
	if err := os.WriteFile(filepath.Join(gavelDir, stem+".json"), []byte(`{"tests":[]}`), 0o644); err != nil {
		t.Fatalf("write run file: %v", err)
	}
	// A sibling secret outside .gavel that a traversal attempt would target.
	if err := os.WriteFile(filepath.Join(dir, "secret.json"), []byte(`{"secret":true}`), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	// A symlink planted inside .gavel that points at the secret.
	if err := os.Symlink(filepath.Join(dir, "secret.json"), filepath.Join(gavelDir, "run-link.json")); err != nil {
		t.Fatalf("plant symlink: %v", err)
	}
	s := &Server{}
	call := func(runID string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		s.handleTestRun(w, httptest.NewRequest(http.MethodGet, "/api/tests/run?dir="+url.QueryEscape(dir)+"&runId="+url.QueryEscape(runID), nil))
		return w
	}

	t.Run("a run id streams its snapshot", func(t *testing.T) {
		w := call(stem)
		if w.Code != http.StatusOK || w.Body.String() != `{"tests":[]}` {
			t.Fatalf("status = %d, body = %q", w.Code, w.Body.String())
		}
	})

	t.Run("a missing run is 404", func(t *testing.T) {
		if w := call("run-9999-01-01T00-00-00Z"); w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404 (%s)", w.Code, w.Body.String())
		}
	})

	t.Run("a symlink out of .gavel is not followed", func(t *testing.T) {
		w := call("run-link")
		if w.Code == http.StatusOK || strings.Contains(w.Body.String(), "secret") {
			t.Errorf("status = %d, body = %q: the symlink was followed", w.Code, w.Body.String())
		}
	})

	// Every malformed or hostile runId must be a bad request, never resolved to a
	// file — defends the endpoint against path traversal and reading non-run files
	// out of .gavel.
	for _, bad := range []string{
		"",
		"../secret",
		"run-../../etc/passwd",
		"sub/run-x",
		`run-x\..\secret`,
		"sha-abc123", // not a per-run snapshot
		"run-x/../secret",
	} {
		t.Run("rejects "+bad, func(t *testing.T) {
			if w := call(bad); w.Code != http.StatusBadRequest {
				t.Errorf("runId %q: status = %d, want 400 (%s)", bad, w.Code, w.Body.String())
			}
		})
	}
}

func TestHandleTestRunReadsTheSnapshotStore(t *testing.T) {
	store := blob.Local(t.TempDir())
	previous := snapshots.OpenStore
	snapshots.OpenStore = func(string) (blob.Store, error) { return store, nil }
	t.Cleanup(func() { snapshots.OpenStore = previous })
	const stem = "run-2026-06-28T10-43-42Z"
	if err := store.Write(context.Background(), stem+".json", strings.NewReader(`{"tests":[]}`)); err != nil {
		t.Fatalf("write run: %v", err)
	}
	dir := t.TempDir()
	w := httptest.NewRecorder()
	(&Server{}).handleTestRun(w, httptest.NewRequest(http.MethodGet, "/api/tests/run?dir="+url.QueryEscape(dir)+"&runId="+stem, nil))
	if w.Code != http.StatusOK || w.Body.String() != `{"tests":[]}` {
		t.Fatalf("status = %d, body = %q", w.Code, w.Body.String())
	}
}
