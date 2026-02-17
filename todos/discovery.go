package todos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// DiscoveryFilters specifies criteria for filtering TODOs during discovery.
// Filters can include or exclude specific statuses to narrow down the set of TODOs to process.
type DiscoveryFilters struct {
	IncludeStatuses []types.Status
	ExcludeStatuses []types.Status
}

// IsEmpty returns true if no filters are configured.
func (filter DiscoveryFilters) IsEmpty() bool {
	return len(filter.IncludeStatuses) == 0 && len(filter.ExcludeStatuses) == 0
}

// Matches returns true if the given TODO matches the filter criteria.
// A TODO is excluded if its status matches any ExcludeStatuses.
// If IncludeStatuses is non-empty, the TODO must match at least one status to be included.
func (filters DiscoveryFilters) Matches(todo *types.TODO) bool {
	status := todo.Status

	// Check exclude filters
	for _, excludeStatus := range filters.ExcludeStatuses {
		if status == excludeStatus {
			return false
		}
	}

	// If include filters specified, check them
	if len(filters.IncludeStatuses) > 0 {
		included := false
		for _, includeStatus := range filters.IncludeStatuses {
			if status == includeStatus {
				included = true
				break
			}
		}
		return included
	}

	return true
}

// DiscoverTODOs recursively discovers all TODO markdown files in the specified directory,
// parses them, applies filters, and returns a sorted list of matching TODOs.
// TODOs are sorted by priority (high to low) using the TODOS.Sort() method.
// Files that fail to parse are silently skipped.
func DiscoverTODOs(dir string, filters DiscoveryFilters) (types.TODOS, error) {
	var todos types.TODOS

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		todo, err := ParseTODO(path)
		if err != nil {
			return nil
		}
		if filters.Matches(todo) {
			todos = append(todos, todo)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	todos.Sort()
	return todos, nil
}
