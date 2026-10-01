package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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
	head := strings.TrimSpace(string(gitOut(t, dir, "rev-parse", "HEAD")))

	s := &Server{ghOpts: github.Options{WorkDir: dir}}

	// Missing hash → 400.
	rec := httptest.NewRecorder()
	s.handleTodoCommitDiff(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits/diff", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing hash status = %d, want 400", rec.Code)
	}

	// Malformed hash is rejected before shelling out to git.
	rec = httptest.NewRecorder()
	s.handleTodoCommitDiff(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits/diff?hash=not-a-hash", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad hash status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}

	// A real commit returns its diff.
	rec = httptest.NewRecorder()
	s.handleTodoCommitDiff(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits/diff?hash="+head, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("diff status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoCommitDiffResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal diff: %v", err)
	}
	if !strings.Contains(resp.Diff, "f.txt") {
		t.Fatalf("diff missing file name:\n%s", resp.Diff)
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
