package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
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
