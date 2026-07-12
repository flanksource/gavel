package todos

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestValidateRuntimeProviderOnlyAcceptsDatabase(t *testing.T) {
	for _, value := range []string{"", ProviderDB, " DB "} {
		if err := ValidateRuntimeProvider(value); err != nil {
			t.Fatalf("ValidateRuntimeProvider(%q) = %v", value, err)
		}
	}
	for _, value := range []string{ProviderGrite, ProviderFiles, "auto", "unknown"} {
		if err := ValidateRuntimeProvider(value); !errors.Is(err, ErrProviderRetired) {
			t.Fatalf("ValidateRuntimeProvider(%q) = %v, want ErrProviderRetired", value, err)
		}
	}
}

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

func TestFileProviderCreateAndEditGeneratedMetadata(t *testing.T) {
	workDir := t.TempDir()
	provider := NewFileProvider(workDir, "")
	ctx := context.Background()

	todo, err := provider.Create(ctx, CreateRequest{
		Title:    "TODO: source task",
		Body:     "Generated body.",
		Status:   types.StatusPending,
		Path:     types.StringOrSlice{"main.go:1"},
		Labels:   []string{"source:todo", "source-id:abc123"},
		Metadata: map[string]any{"source": "code-comment", "source_id": "abc123"},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	reloaded, err := provider.Get(ctx, todo.FilePath)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(reloaded.Path) != 1 || reloaded.Path[0] != "main.go:1" {
		t.Fatalf("path = %#v, want main.go:1", reloaded.Path)
	}
	if reloaded.Metadata["source"] != "code-comment" || reloaded.Metadata["source_id"] != "abc123" {
		t.Fatalf("metadata not preserved: %#v", reloaded.Metadata)
	}

	path := types.StringOrSlice{"main.go:2"}
	if err := provider.Edit(ctx, reloaded, EditRequest{
		Path:     &path,
		Labels:   []string{"source:todo", "source-id:abc123"},
		Metadata: map[string]any{"source_id": "abc123", "source_marker": "todo"},
	}); err != nil {
		t.Fatalf("Edit failed: %v", err)
	}
	edited, err := provider.Get(ctx, todo.FilePath)
	if err != nil {
		t.Fatalf("Get after edit failed: %v", err)
	}
	if len(edited.Path) != 1 || edited.Path[0] != "main.go:2" {
		t.Fatalf("edited path = %#v, want main.go:2", edited.Path)
	}
	if edited.Metadata["source_marker"] != "todo" {
		t.Fatalf("metadata edit missing source_marker: %#v", edited.Metadata)
	}
}
