package todos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestDiscoverTODOs_SortsByPriorityAndName(t *testing.T) {
	// Setup: Create .todos/ with files:
	//   - medium-z.md (priority: medium)
	//   - high-b.md (priority: high)
	//   - high-a.md (priority: high)
	//   - low-x.md (priority: low)
	// Expected order: high-a.md, high-b.md, medium-z.md, low-x.md

	tmpDir := t.TempDir()
	todosDir := filepath.Join(tmpDir, ".todos")
	if err := os.Mkdir(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create .todos dir: %v", err)
	}

	// Create test files - each must have a code block for parser to extract metadata
	files := map[string]types.Priority{
		"medium-z.md": types.PriorityMedium,
		"high-b.md":   types.PriorityHigh,
		"high-a.md":   types.PriorityHigh,
		"low-x.md":    types.PriorityLow,
	}

	for name, priority := range files {
		content := "---\npriority: " + string(priority) + "\nstatus: pending\nattempts: 0\nlanguage: go\n---\n\n# TODO: " + name + "\n\n## Verification\n\n```bash\necho test\n```\n"

		path := filepath.Join(todosDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", name, err)
		}
	}

	// Discover and sort TODOs
	todos, err := DiscoverTODOs(todosDir, DiscoveryFilters{})
	if err != nil {
		t.Fatalf("Failed to discover TODOs: %v", err)
	}

	// Verify order
	expectedOrder := []string{"high-a.md", "high-b.md", "medium-z.md", "low-x.md"}
	if len(todos) != len(expectedOrder) {
		t.Fatalf("Expected %d TODOs, got %d", len(expectedOrder), len(todos))
	}

	for i, expected := range expectedOrder {
		actual := filepath.Base(todos[i].FilePath)
		if actual != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, actual)
		}
	}
}

func TestDiscoverTODOs_FiltersByStatus(t *testing.T) {
	// Setup: Create TODOs with status: pending, completed, failed
	// Expected: Only pending and failed returned (skip completed)

	tmpDir := t.TempDir()
	todosDir := filepath.Join(tmpDir, ".todos")
	if err := os.Mkdir(todosDir, 0755); err != nil {
		t.Fatalf("Failed to create .todos dir: %v", err)
	}

	statuses := []types.Status{types.StatusPending, types.StatusCompleted, types.StatusFailed}
	for i, status := range statuses {
		// Each file must have a code block for parser to extract metadata
		content := "---\npriority: high\nstatus: " + string(status) + "\nattempts: 0\nlanguage: go\n---\n\n# TODO: Test\n\n## Verification\n\n```bash\necho test\n```\n"

		name := filepath.Join(todosDir, string(status)+".md")
		if err := os.WriteFile(name, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %d: %v", i, err)
		}
	}

	// Discover with filter to exclude completed
	todos, err := DiscoverTODOs(todosDir, DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		t.Fatalf("Failed to discover TODOs: %v", err)
	}

	// Should only return pending and failed
	if len(todos) != 2 {
		t.Errorf("Expected 2 TODOs, got %d", len(todos))
	}

	for _, todo := range todos {
		if todo.Status == types.StatusCompleted {
			t.Error("Completed TODO should have been filtered out")
		}
	}
}
