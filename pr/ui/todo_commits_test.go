package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestCollectTodoCommits(t *testing.T) {
	dir := t.TempDir()
	gitInDir(t, dir, "init", "-q")
	gitInDir(t, dir, "config", "user.email", "test@example.com")
	gitInDir(t, dir, "config", "user.name", "Test User")
	gitInDir(t, dir, "config", "commit.gpgsign", "false")
	gitInDir(t, dir, "remote", "add", "origin", "git@github.com:flanksource/gavel.git")
	gitInDir(t, dir, "commit", "--allow-empty", "-m", "feat: implement thing\n\nGavel-Issue-Id: ISSUE-42")
	gitInDir(t, dir, "commit", "--allow-empty", "-m", "chore: unrelated")

	commits, err := collectTodoCommits(dir, "ISSUE-42")
	if err != nil {
		t.Fatalf("collectTodoCommits: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("got %d commits, want 1: %+v", len(commits), commits)
	}
	got := commits[0]
	if got.Subject != "feat: implement thing" {
		t.Fatalf("subject = %q, want %q", got.Subject, "feat: implement thing")
	}
	if got.Author != "Test User" {
		t.Fatalf("author = %q, want %q", got.Author, "Test User")
	}
	if len(got.ShortHash) != 7 || got.Hash[:7] != got.ShortHash {
		t.Fatalf("shortHash = %q, want 7-char prefix of %q", got.ShortHash, got.Hash)
	}
	wantURL := "https://github.com/flanksource/gavel/commit/" + got.Hash
	if got.URL != wantURL {
		t.Fatalf("url = %q, want %q", got.URL, wantURL)
	}

	// An empty issue id (file-backed todos) returns no commits without scanning.
	empty, err := collectTodoCommits(dir, "")
	if err != nil {
		t.Fatalf("collectTodoCommits(empty): %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("got %d commits for empty id, want 0", len(empty))
	}
}

func TestHandleTodoCommits(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	// Missing ref is a 400.
	rec := httptest.NewRecorder()
	s.handleTodoCommits(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits?provider=todos", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing ref status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}

	// Create a file-backed todo (no id), then ask for its linked commits: the
	// endpoint resolves the todo and returns an empty list (file todos carry no
	// Gavel-Issue-Id to match against).
	rec = httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos?provider=todos", strings.NewReader(`{"title":"Wire commits"}`)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want 201; body = %q", rec.Code, rec.Body.String())
	}
	var created todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}

	rec = httptest.NewRecorder()
	s.handleTodoCommits(rec, httptest.NewRequest(http.MethodGet, "/api/todos/commits?provider=todos&ref="+url.QueryEscape(created.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("commits status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoCommitsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal commits: %v", err)
	}
	if resp.IssueID != "" {
		t.Fatalf("issueId = %q, want empty for a file todo", resp.IssueID)
	}
	if len(resp.Commits) != 0 {
		t.Fatalf("got %d commits, want 0 for a file todo", len(resp.Commits))
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

func TestCommitDiffStats(t *testing.T) {
	dir := t.TempDir()
	gitInDir(t, dir, "init", "-q")
	gitInDir(t, dir, "config", "user.email", "test@example.com")
	gitInDir(t, dir, "config", "user.name", "Test User")
	gitInDir(t, dir, "config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatalf("write a.txt: %v", err)
	}
	gitInDir(t, dir, "add", "a.txt")
	gitInDir(t, dir, "commit", "-m", "feat: a\n\nGavel-Issue-Id: ISSUE-7")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatalf("rewrite a.txt: %v", err)
	}
	gitInDir(t, dir, "add", "a.txt")
	gitInDir(t, dir, "commit", "-m", "fix: trim a\n\nGavel-Issue-Id: ISSUE-7")

	// No cache DB is configured in tests, so this exercises the direct-compute path.
	stats := commitDiffStats(context.Background(), dir)
	got := diffStatFor(stats, "ISSUE-7")
	if got == nil {
		t.Fatalf("diffStatFor(ISSUE-7) = nil, want stats; map = %+v", stats)
	}
	if got.Commits != 2 || got.Files != 1 || got.Adds != 2 || got.Dels != 1 {
		t.Fatalf("ISSUE-7 diff = %+v, want {Commits:2 Files:1 Adds:2 Dels:1}", got)
	}

	// An id with no commits, and an empty id, both omit the diff entirely.
	if d := diffStatFor(stats, "ISSUE-404"); d != nil {
		t.Fatalf("diffStatFor(unknown) = %+v, want nil", d)
	}
	if d := diffStatFor(stats, ""); d != nil {
		t.Fatalf("diffStatFor(empty) = %+v, want nil", d)
	}
}

func TestCollectTodoCommitsNoRemote(t *testing.T) {
	dir := t.TempDir()
	gitInDir(t, dir, "init", "-q")
	gitInDir(t, dir, "config", "user.email", "test@example.com")
	gitInDir(t, dir, "config", "user.name", "Test User")
	gitInDir(t, dir, "config", "commit.gpgsign", "false")
	gitInDir(t, dir, "commit", "--allow-empty", "-m", "feat: local\n\nGavel-Issue-Id: LOCAL-1")

	commits, err := collectTodoCommits(dir, "LOCAL-1")
	if err != nil {
		t.Fatalf("collectTodoCommits: %v", err)
	}
	if len(commits) != 1 {
		t.Fatalf("got %d commits, want 1", len(commits))
	}
	// No origin remote → no web link, but the commit metadata is still returned.
	if commits[0].URL != "" {
		t.Fatalf("url = %q, want empty for local-only repo", commits[0].URL)
	}
}
