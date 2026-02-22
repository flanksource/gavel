package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func TestSyncTodos_CreatesForFailingChecks(t *testing.T) {
	tmpDir := t.TempDir()

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{
			"no-code-duplication": {Pass: false, Evidence: []verify.Evidence{
				{File: "pkg/handler.go", Line: 42, Message: "duplicated logic"},
			}},
			"tests-added": {Pass: true},
		},
		Ratings: map[string]verify.RatingResult{},
	}

	syncResult, err := SyncTodos(result, SyncOptions{
		TodosDir:       tmpDir,
		ScoreThreshold: 80,
		RepoPath:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syncResult.Created) != 1 {
		t.Fatalf("expected 1 created, got %d: %v", len(syncResult.Created), syncResult.Created)
	}
	if syncResult.Created[0] != "no-code-duplication.md" {
		t.Errorf("expected no-code-duplication.md, got %s", syncResult.Created[0])
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "no-code-duplication.md"))
	if err != nil {
		t.Fatalf("failed to read created file: %v", err)
	}
	s := string(content)

	for _, want := range []string{"priority: high", "status: pending", "No Code Duplication", "pkg/handler.go"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in content:\n%s", want, s)
		}
	}
}

func TestSyncTodos_CreatesForLowRatings(t *testing.T) {
	tmpDir := t.TempDir()

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{},
		Ratings: map[string]verify.RatingResult{
			"duplication": {Score: 55, Findings: []verify.Evidence{
				{File: "pkg/util.go", Message: "high duplication"},
			}},
			"security": {Score: 90},
		},
	}

	syncResult, err := SyncTodos(result, SyncOptions{
		TodosDir:       tmpDir,
		ScoreThreshold: 80,
		RepoPath:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syncResult.Created) != 1 {
		t.Fatalf("expected 1 created, got %d", len(syncResult.Created))
	}
	if syncResult.Created[0] != "rating-duplication.md" {
		t.Errorf("expected rating-duplication.md, got %s", syncResult.Created[0])
	}

	content, err := os.ReadFile(filepath.Join(tmpDir, "rating-duplication.md"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	s := string(content)

	if !strings.Contains(s, "priority: medium") {
		t.Errorf("expected medium priority in:\n%s", s)
	}
	if !strings.Contains(s, "score 55") {
		t.Errorf("expected score in title:\n%s", s)
	}
}

func TestSyncTodos_CompletesPassingChecks(t *testing.T) {
	tmpDir := t.TempDir()

	// Pre-create a TODO for a check that now passes
	existingTodo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    "No Code Duplication",
			Priority: types.PriorityHigh,
			Status:   types.StatusPending,
		},
	}
	todoPath := filepath.Join(tmpDir, "no-code-duplication.md")
	if err := todos.WriteTODOFile(todoPath, existingTodo); err != nil {
		t.Fatalf("failed to create existing todo: %v", err)
	}

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{
			"no-code-duplication": {Pass: true},
		},
		Ratings: map[string]verify.RatingResult{},
	}

	syncResult, err := SyncTodos(result, SyncOptions{
		TodosDir:       tmpDir,
		ScoreThreshold: 80,
		RepoPath:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syncResult.Completed) != 1 {
		t.Fatalf("expected 1 completed, got %d", len(syncResult.Completed))
	}

	fm, err := todos.ReadTODOState(todoPath)
	if err != nil {
		t.Fatalf("failed to read state: %v", err)
	}
	if fm.Status != types.StatusCompleted {
		t.Errorf("expected completed, got %s", fm.Status)
	}
}

func TestSyncTodos_UpdatesExistingFailingCheck(t *testing.T) {
	tmpDir := t.TempDir()

	existingTodo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    "Tests Added",
			Priority: types.PriorityHigh,
			Status:   types.StatusPending,
		},
	}
	todoPath := filepath.Join(tmpDir, "tests-added.md")
	if err := todos.WriteTODOFile(todoPath, existingTodo); err != nil {
		t.Fatalf("failed to create existing todo: %v", err)
	}

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{
			"tests-added": {Pass: false, Evidence: []verify.Evidence{
				{File: "pkg/new.go", Message: "no tests"},
			}},
		},
		Ratings: map[string]verify.RatingResult{},
	}

	syncResult, err := SyncTodos(result, SyncOptions{
		TodosDir:       tmpDir,
		ScoreThreshold: 80,
		RepoPath:       t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(syncResult.Updated) != 1 {
		t.Fatalf("expected 1 updated, got %d", len(syncResult.Updated))
	}

	// Status should remain pending (not changed)
	fm, err := todos.ReadTODOState(todoPath)
	if err != nil {
		t.Fatalf("failed to read state: %v", err)
	}
	if fm.Status != types.StatusPending {
		t.Errorf("expected pending, got %s", fm.Status)
	}
}

func TestSyncTodos_DetectsGoLanguage(t *testing.T) {
	repoDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoDir, "go.mod"), []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}
	todosDir := t.TempDir()

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{
			"tests-added": {Pass: false},
		},
		Ratings: map[string]verify.RatingResult{},
	}

	_, err := SyncTodos(result, SyncOptions{
		TodosDir: todosDir, ScoreThreshold: 80, RepoPath: repoDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fm, err := todos.ReadTODOState(filepath.Join(todosDir, "tests-added.md"))
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if fm.Language != types.LanguageGo {
		t.Errorf("expected go, got %s", fm.Language)
	}
}

func TestSyncTodos_NoChangesWhenAllPass(t *testing.T) {
	tmpDir := t.TempDir()

	result := &verify.VerifyResult{
		Checks: map[string]verify.CheckResult{
			"tests-added":         {Pass: true},
			"no-code-duplication": {Pass: true},
		},
		Ratings: map[string]verify.RatingResult{
			"security": {Score: 95},
		},
	}

	syncResult, err := SyncTodos(result, SyncOptions{
		TodosDir: tmpDir, ScoreThreshold: 80, RepoPath: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	total := len(syncResult.Created) + len(syncResult.Updated) + len(syncResult.Completed)
	if total != 0 {
		t.Errorf("expected no changes, got %d", total)
	}
}

func TestHumanize(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"no-code-duplication", "No Code Duplication"},
		{"tests-added", "Tests Added"},
		{"security", "Security"},
	}
	for _, tt := range tests {
		got := humanize(tt.input)
		if got != tt.want {
			t.Errorf("humanize(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEvidencePaths(t *testing.T) {
	evidence := []verify.Evidence{
		{File: "pkg/a.go", Message: "issue"},
		{File: "pkg/b.go", Message: "issue"},
		{File: "pkg/a.go", Message: "another issue"},
		{Message: "no file"},
	}
	paths := evidencePaths(evidence)
	if len(paths) != 2 {
		t.Errorf("expected 2 unique paths, got %d: %v", len(paths), paths)
	}
}
