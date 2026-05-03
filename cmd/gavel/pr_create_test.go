package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
)

// --- fixture helpers --------------------------------------------------------

type prCreateFixture struct {
	t        *testing.T
	repo     string // working copy
	bare     string // bare repo acting as `origin`
	baseSHA  string // commit on main in working copy
	topicSHA string // feature commit on a branch in working copy
}

func runCmd(t *testing.T, dir, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func newPRCreateFixture(t *testing.T) *prCreateFixture {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "origin.git")
	repo := filepath.Join(root, "work")

	runCmd(t, root, "git", "init", "--bare", "-b", "main", bare)
	runCmd(t, root, "git", "init", "-b", "main", repo)
	runCmd(t, repo, "git", "remote", "add", "origin", bare)

	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, repo, "git", "add", "README.md")
	runCmd(t, repo, "git", "commit", "-m", "chore: initial")
	runCmd(t, repo, "git", "push", "-u", "origin", "main")
	baseSHA := runCmd(t, repo, "git", "rev-parse", "HEAD")

	runCmd(t, repo, "git", "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, repo, "git", "add", "feature.txt")
	runCmd(t, repo, "git", "commit", "-m", "feat: add feature.txt")
	topicSHA := runCmd(t, repo, "git", "rev-parse", "HEAD")
	runCmd(t, repo, "git", "checkout", "main")

	return &prCreateFixture{t: t, repo: repo, bare: bare, baseSHA: baseSHA, topicSHA: topicSHA}
}

// --- tests ------------------------------------------------------------------

func TestSplitBaseRef(t *testing.T) {
	cases := []struct {
		in, branch, ref string
	}{
		{"origin/main", "main", "origin/main"},
		{"main", "main", "main"},
		{"refs/heads/foo", "foo", "refs/heads/foo"},
	}
	for _, c := range cases {
		gotBranch, gotRef := splitBaseRef(c.in)
		if gotBranch != c.branch || gotRef != c.ref {
			t.Errorf("splitBaseRef(%q) = (%q, %q), want (%q, %q)", c.in, gotBranch, gotRef, c.branch, c.ref)
		}
	}
}

func TestPreflightSHAExists(t *testing.T) {
	f := newPRCreateFixture(t)
	if _, err := preflightPRCreate(f.repo, f.topicSHA, prCreateOptions{Base: "origin/main"}); err != nil {
		t.Fatalf("expected real SHA to pass preflight: %v", err)
	}
	if _, err := preflightPRCreate(f.repo, "deadbeef", prCreateOptions{Base: "origin/main"}); err == nil {
		t.Fatal("expected unknown SHA to fail preflight")
	}
}

func TestPreflightRejectsMergeCommitWithoutMainline(t *testing.T) {
	f := newPRCreateFixture(t)
	runCmd(t, f.repo, "git", "checkout", "-b", "other")
	if err := os.WriteFile(filepath.Join(f.repo, "other.txt"), []byte("o\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, f.repo, "git", "add", "other.txt")
	runCmd(t, f.repo, "git", "commit", "-m", "chore: other branch")
	runCmd(t, f.repo, "git", "checkout", "main")
	runCmd(t, f.repo, "git", "merge", "--no-ff", "-m", "merge: other", "other")
	mergeSHA := runCmd(t, f.repo, "git", "rev-parse", "HEAD")

	_, err := preflightPRCreate(f.repo, mergeSHA, prCreateOptions{Base: "origin/main"})
	if err == nil || !strings.Contains(err.Error(), "--mainline") {
		t.Fatalf("expected --mainline error, got %v", err)
	}
	if _, err := preflightPRCreate(f.repo, mergeSHA, prCreateOptions{Base: "origin/main", Mainline: 1}); err != nil {
		t.Fatalf("preflight with --mainline=1: %v", err)
	}
}

func TestPreflightRejectsMidRebase(t *testing.T) {
	f := newPRCreateFixture(t)
	gitDir := strings.TrimSpace(runCmd(t, f.repo, "git", "rev-parse", "--git-dir"))
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(f.repo, gitDir)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "REBASE_HEAD"), []byte(f.topicSHA), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := preflightPRCreate(f.repo, f.topicSHA, prCreateOptions{Base: "origin/main"})
	if err == nil || !strings.Contains(err.Error(), "REBASE_HEAD") {
		t.Fatalf("expected refusal mentioning REBASE_HEAD, got %v", err)
	}
}

func TestEndToEndLocalNoGitHub(t *testing.T) {
	f := newPRCreateFixture(t)

	var capturedInput github.CreatePRInput
	var capturedURL string
	deps := prCreateDeps{
		createPR: func(_ github.Options, in github.CreatePRInput) (*github.CreatePRResult, error) {
			capturedInput = in
			return &github.CreatePRResult{Number: 7, URL: "https://example/pr/7", Title: in.Title, Base: in.Base}, nil
		},
		openBrowser: func(url string) { capturedURL = url },
		generateContent: func(_ context.Context, _ commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			return commitpkg.PRContent{Title: "feat: add feature.txt", Body: "body", Branch: "feat/add-feature"}, nil
		},
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	if err := os.Chdir(f.repo); err != nil {
		t.Fatal(err)
	}

	err := runPRCreateWithDeps(context.Background(), f.topicSHA, prCreateOptions{Base: "origin/main"}, deps)
	if err != nil {
		t.Fatalf("runPRCreateWithDeps: %v", err)
	}

	if capturedInput.Title != "feat: add feature.txt" {
		t.Errorf("Title = %q, want %q", capturedInput.Title, "feat: add feature.txt")
	}
	if capturedInput.Base != "main" {
		t.Errorf("Base = %q, want %q (stripped origin/)", capturedInput.Base, "main")
	}
	if !strings.HasPrefix(capturedInput.Head, "feat/add-feature-") {
		t.Errorf("Head = %q, want prefix feat/add-feature-", capturedInput.Head)
	}
	if capturedURL != "https://example/pr/7" {
		t.Errorf("openBrowser url = %q", capturedURL)
	}

	scratch := filepath.Join(f.repo, ".tmp", "pr-create")
	entries, _ := os.ReadDir(scratch)
	if len(entries) != 0 {
		t.Errorf("worktree dir not cleaned: %v", entries)
	}

	wtList := runCmd(t, f.repo, "git", "worktree", "list", "--porcelain")
	if strings.Contains(wtList, "/.tmp/pr-create/") {
		t.Errorf("worktree still registered:\n%s", wtList)
	}

	branches := runCmd(t, f.bare, "git", "branch", "--list")
	if !strings.Contains(branches, capturedInput.Head) {
		t.Errorf("topic branch %q not pushed to bare:\n%s", capturedInput.Head, branches)
	}
}

func TestCleanupRefusesUnknownPath(t *testing.T) {
	f := newPRCreateFixture(t)
	err := assertSafeWorktreePath(f.repo, "/etc")
	if err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("expected outside-scratch refusal, got %v", err)
	}
}

func TestCleanupRefusesMissingMarker(t *testing.T) {
	f := newPRCreateFixture(t)
	wt := filepath.Join(f.repo, ".tmp", "pr-create", "abc12345-x")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	err := assertSafeWorktreePath(f.repo, wt)
	if err == nil || !strings.Contains(err.Error(), "marker missing") {
		t.Fatalf("expected marker-missing refusal, got %v", err)
	}
}

func TestCleanupRefusesPathOutsideTmp(t *testing.T) {
	f := newPRCreateFixture(t)
	bad := filepath.Join(f.repo, "not-tmp", "abc")
	if err := os.MkdirAll(bad, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeMarker(bad, strings.Repeat("a", 40), f.repo); err != nil {
		t.Fatal(err)
	}
	err := assertSafeWorktreePath(f.repo, bad)
	if err == nil || !strings.Contains(err.Error(), "outside") {
		t.Fatalf("expected outside-scratch refusal, got %v", err)
	}
}

func TestCherryPickConflictLeavesWorktree(t *testing.T) {
	f := newPRCreateFixture(t)

	// Make conflicting change on main against feature.txt's content.
	runCmd(t, f.repo, "git", "checkout", "main")
	if err := os.WriteFile(filepath.Join(f.repo, "feature.txt"), []byte("conflict\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCmd(t, f.repo, "git", "add", "feature.txt")
	runCmd(t, f.repo, "git", "commit", "-m", "chore: conflicting")
	runCmd(t, f.repo, "git", "push", "origin", "main")

	deps := prCreateDeps{
		createPR: func(github.Options, github.CreatePRInput) (*github.CreatePRResult, error) {
			t.Fatal("createPR should not be called on conflict")
			return nil, nil
		},
		openBrowser: func(string) { t.Fatal("openBrowser should not be called on conflict") },
		generateContent: func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			t.Fatal("generateContent should not be called on conflict")
			return commitpkg.PRContent{}, nil
		},
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	_ = os.Chdir(f.repo)

	err := runPRCreateWithDeps(context.Background(), f.topicSHA, prCreateOptions{Base: "origin/main"}, deps)
	if err == nil || !strings.Contains(err.Error(), "cherry-pick") {
		t.Fatalf("expected cherry-pick error, got %v", err)
	}

	scratch := filepath.Join(f.repo, ".tmp", "pr-create")
	entries, _ := os.ReadDir(scratch)
	if len(entries) != 1 {
		t.Fatalf("expected 1 leftover worktree, got %d entries: %v", len(entries), entries)
	}
	markerPath := filepath.Join(scratch, entries[0].Name(), prCreateMarkerFile)
	if _, err := os.Stat(markerPath); err != nil {
		t.Errorf("marker file gone after conflict: %v", err)
	}

	// Cleanup so t.TempDir teardown doesn't choke on the registered worktree.
	t.Cleanup(func() {
		_ = removePRCreateWorktree(f.repo, filepath.Join(scratch, entries[0].Name()))
	})
}

func TestRerunGetsFreshBranch(t *testing.T) {
	f := newPRCreateFixture(t)

	var heads []string
	deps := prCreateDeps{
		createPR: func(_ github.Options, in github.CreatePRInput) (*github.CreatePRResult, error) {
			heads = append(heads, in.Head)
			return &github.CreatePRResult{Number: len(heads), URL: "u", Base: in.Base}, nil
		},
		openBrowser: func(string) {},
		generateContent: func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			return commitpkg.PRContent{Title: "feat: x", Body: "", Branch: "feat/x"}, nil
		},
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	_ = os.Chdir(f.repo)

	for i := 0; i < 2; i++ {
		if err := runPRCreateWithDeps(context.Background(), f.topicSHA, prCreateOptions{Base: "origin/main"}, deps); err != nil {
			t.Fatalf("run %d: %v", i, err)
		}
	}
	if len(heads) != 2 {
		t.Fatalf("expected 2 createPR calls, got %d", len(heads))
	}
	if heads[0] == heads[1] {
		t.Errorf("expected distinct topic branches, got %q twice", heads[0])
	}
}

func TestLLMUnavailableSkipsPush(t *testing.T) {
	f := newPRCreateFixture(t)

	deps := prCreateDeps{
		createPR: func(github.Options, github.CreatePRInput) (*github.CreatePRResult, error) {
			t.Fatal("createPR should not be called when LLM is unavailable")
			return nil, nil
		},
		openBrowser: func(string) { t.Fatal("openBrowser should not be called") },
		generateContent: func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			return commitpkg.PRContent{}, commitpkg.ErrLLMUnavailable
		},
	}

	prevWD, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	_ = os.Chdir(f.repo)

	err := runPRCreateWithDeps(context.Background(), f.topicSHA, prCreateOptions{Base: "origin/main"}, deps)
	if err == nil || !errors.Is(err, commitpkg.ErrLLMUnavailable) {
		t.Fatalf("expected ErrLLMUnavailable, got %v", err)
	}

	branches := runCmd(t, f.bare, "git", "branch", "--list")
	if strings.Contains(branches, "feat/") || strings.Contains(branches, "gavel/pr-create") {
		t.Errorf("no topic branch should be pushed when LLM fails:\n%s", branches)
	}
}
