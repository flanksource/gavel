package todos

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestFileProviderEditUpdatesTitleAndBody(t *testing.T) {
	workDir := t.TempDir()
	provider := NewFileProvider(workDir, "")
	ctx := context.Background()

	todo, err := provider.Create(ctx, CreateRequest{
		Title:    "Original title",
		Body:     "Original body",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	newTitle := "Updated title"
	newBody := "## Updated\n\nThe new body content."
	if err := provider.Edit(ctx, todo, EditRequest{Title: &newTitle, Body: &newBody}); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}
	if todo.Title != "Updated title" {
		t.Fatalf("in-memory title = %q, want updated", todo.Title)
	}

	reloaded, err := provider.Get(ctx, todo.FilePath)
	if err != nil {
		t.Fatalf("Get after edit failed: %v", err)
	}
	if reloaded.Title != "Updated title" {
		t.Fatalf("persisted title = %q, want updated", reloaded.Title)
	}
	if !strings.Contains(reloaded.MarkdownBody, "The new body content.") {
		t.Fatalf("persisted body missing edit: %q", reloaded.MarkdownBody)
	}
	if strings.Contains(reloaded.MarkdownBody, "Original body") {
		t.Fatalf("old body should be replaced: %q", reloaded.MarkdownBody)
	}
	// Priority/status frontmatter must survive a content edit.
	if reloaded.Priority != types.PriorityMedium || reloaded.Status != types.StatusPending {
		t.Fatalf("edit clobbered frontmatter state: %+v", reloaded.TODOFrontmatter)
	}
}

func TestFileProviderEditRejectsEmptyTitle(t *testing.T) {
	workDir := t.TempDir()
	provider := NewFileProvider(workDir, "")
	ctx := context.Background()
	todo, err := provider.Create(ctx, CreateRequest{Title: "Keep me", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	blank := "   "
	if err := provider.Edit(ctx, todo, EditRequest{Title: &blank}); err == nil {
		t.Fatal("expected error for empty title")
	}
	if err := provider.Edit(ctx, todo, EditRequest{}); err == nil {
		t.Fatal("expected error for empty edit")
	}
}

func TestFileProviderCommentAppendsSection(t *testing.T) {
	workDir := t.TempDir()
	provider := NewFileProvider(workDir, "")
	ctx := context.Background()

	todo, err := provider.Create(ctx, CreateRequest{
		Title:  "Discuss me",
		Body:   "Initial body.",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := provider.Comment(ctx, todo, "first observation"); err != nil {
		t.Fatalf("Comment failed: %v", err)
	}
	if err := provider.Comment(ctx, todo, "second observation"); err != nil {
		t.Fatalf("second Comment failed: %v", err)
	}

	reloaded, err := provider.Get(ctx, todo.FilePath)
	if err != nil {
		t.Fatalf("Get after comment failed: %v", err)
	}
	body := reloaded.MarkdownBody
	if strings.Count(body, "## Comments") != 1 {
		t.Fatalf("expected a single Comments section, got: %q", body)
	}
	if !strings.Contains(body, "first observation") || !strings.Contains(body, "second observation") {
		t.Fatalf("comments not persisted: %q", body)
	}
	if !strings.Contains(body, "Initial body.") {
		t.Fatalf("original body lost after commenting: %q", body)
	}

	if err := provider.Comment(ctx, todo, "  "); err == nil {
		t.Fatal("expected error for blank comment")
	}
}

func TestFileProviderCreateGetListDelete(t *testing.T) {
	workDir := t.TempDir()
	provider := NewFileProvider(workDir, "")

	todo, err := provider.Create(context.Background(), CreateRequest{
		Title:    "Fix UI",
		Body:     "Implement the workspace todo view.",
		Priority: types.PriorityHigh,
		Status:   types.StatusInProgress,
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if !strings.HasSuffix(todo.FilePath, filepath.Join(".todos", "fix-ui.md")) {
		t.Fatalf("unexpected file path: %s", todo.FilePath)
	}
	if todo.Title != "Fix UI" || todo.Priority != types.PriorityHigh || todo.Status != types.StatusInProgress {
		t.Fatalf("unexpected TODO fields: %+v", todo)
	}
	if !strings.Contains(todo.MarkdownBody, "workspace todo view") {
		t.Fatalf("markdown body was not preserved: %q", todo.MarkdownBody)
	}

	items, err := provider.List(context.Background(), DiscoveryFilters{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Fix UI" {
		t.Fatalf("unexpected list result: %+v", items)
	}

	if err := provider.Delete(context.Background(), todo); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := os.Stat(todo.FilePath); !os.IsNotExist(err) {
		t.Fatalf("expected TODO file to be removed, stat err=%v", err)
	}
}
