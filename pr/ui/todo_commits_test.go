package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
)

// commitInDir is a tiny helper that runs git in dir, failing the test on error.
func gitInDir(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestHandleTodoCommitDiff(t *testing.T) {
	dir := t.TempDir()
	gitInDir(t, dir, "init", "-q")
	gitInDir(t, dir, "config", "user.email", "test@example.com")
	gitInDir(t, dir, "config", "user.name", "Test User")
	gitInDir(t, dir, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("a\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	gitInDir(t, dir, "add", "f.txt")
	gitInDir(t, dir, "commit", "-m", "feat: f")
	base := strings.TrimSpace(string(gitOut(t, dir, "rev-parse", "HEAD")))
	if err := os.WriteFile(filepath.Join(dir, "g.txt"), []byte("g\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	gitInDir(t, dir, "add", "g.txt")
	gitInDir(t, dir, "commit", "-m", "feat: g")
	head := strings.TrimSpace(string(gitOut(t, dir, "rev-parse", "HEAD")))
	const missing = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

	s := &Server{ghOpts: github.Options{WorkDir: dir}}
	get := func(handler http.HandlerFunc, query string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits/diff?"+query, nil))
		return rec
	}

	statusCases := []struct {
		name, query string
		want        int
	}{
		{"missing hash", "", http.StatusBadRequest},
		{"malformed hash", "hash=not-a-hash", http.StatusBadRequest},
		{"malformed base", "hash=" + head + "&base=not-a-hash", http.StatusBadRequest},
		{"option-shaped file", "hash=" + head + "&file=--output=x", http.StatusBadRequest},
		{"unreachable hash", "hash=" + missing, http.StatusGone},
		{"unreachable base", "hash=" + head + "&base=" + missing, http.StatusGone},
	}
	for _, tc := range statusCases {
		for name, handler := range map[string]http.HandlerFunc{"diff": s.handleTodoCommitDiff, "files": s.handleTodoCommitFiles} {
			if rec := get(handler, tc.query); rec.Code != tc.want {
				t.Fatalf("%s %s: status = %d, want %d; body = %q", name, tc.name, rec.Code, tc.want, rec.Body.String())
			}
		}
	}
	if body := get(s.handleTodoCommitDiff, "hash="+missing).Body.String(); !strings.Contains(body, missing) {
		t.Fatalf("410 body must name the missing commit: %q", body)
	}

	// A single commit returns its plain diff in the GitDiffPayload shape.
	rec := get(s.handleTodoCommitDiff, "hash="+head+"&file=g.txt")
	if rec.Code != http.StatusOK {
		t.Fatalf("diff status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal diff: %v", err)
	}
	diff, _ := payload["diff"].(string)
	delete(payload, "diff")
	want := map[string]any{"truncated": false, "binary": false, "path": "g.txt", "commit": head}
	if !reflect.DeepEqual(payload, want) || !strings.HasPrefix(diff, "diff --git a/g.txt b/g.txt\n") {
		t.Fatalf("payload = %v diff = %q, want %v and a g.txt diff", payload, diff, want)
	}

	// base turns the request into the base..hash range.
	var ranged todoCommitDiffResponse
	if err := json.Unmarshal(get(s.handleTodoCommitDiff, "hash="+head+"&base="+base).Body.Bytes(), &ranged); err != nil {
		t.Fatalf("unmarshal range diff: %v", err)
	}
	if !strings.Contains(ranged.Diff, "g.txt") || strings.Contains(ranged.Diff, "f.txt") {
		t.Fatalf("range diff must hold only g.txt:\n%s", ranged.Diff)
	}
	var files todoCommitFilesResponse
	if err := json.Unmarshal(get(s.handleTodoCommitFiles, "hash="+head+"&base="+base).Body.Bytes(), &files); err != nil {
		t.Fatalf("unmarshal files: %v", err)
	}
	if files.Hash != head || files.Base != base || len(files.Files) != 1 || files.Files[0].Path != "g.txt" {
		t.Fatalf("files response = %+v, want hash %s base %s and g.txt", files, head, base)
	}
}

func gitOut(t *testing.T, dir string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return out
}
