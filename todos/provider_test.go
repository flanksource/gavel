package todos

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

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
