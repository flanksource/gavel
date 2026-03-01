package todos

import (
	"testing"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
)

func todo(path types.StringOrSlice, priority types.Priority, name string) *types.TODO {
	return &types.TODO{
		FilePath:        ".todos/" + name + ".md",
		TODOFrontmatter: types.TODOFrontmatter{Priority: priority, Status: types.StatusPending, Path: path},
	}
}

func TestGroupTODOs_ByFile(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/auth/login.go"}, types.PriorityHigh, "a"),
		todo(types.StringOrSlice{"pkg/auth/login.go"}, types.PriorityLow, "b"),
		todo(types.StringOrSlice{"pkg/api/handler.go"}, types.PriorityMedium, "c"),
		todo(nil, types.PriorityHigh, "d"),
	}

	groups := GroupTODOs(todos, GroupByFile)

	assert.Len(t, groups, 3)
	// High-priority group first
	assert.Equal(t, "pkg/auth/login.go", groups[0].Name)
	assert.Len(t, groups[0].TODOs, 2)
	// Medium-priority group second
	assert.Equal(t, "pkg/api/handler.go", groups[1].Name)
	assert.Len(t, groups[1].TODOs, 1)
	// Ungrouped last
	assert.Equal(t, UngroupedLabel, groups[2].Name)
	assert.Len(t, groups[2].TODOs, 1)
}

func TestGroupTODOs_ByDirectory(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/auth/login.go"}, types.PriorityHigh, "a"),
		todo(types.StringOrSlice{"pkg/auth/session.go"}, types.PriorityLow, "b"),
		todo(types.StringOrSlice{"pkg/api/handler.go"}, types.PriorityMedium, "c"),
	}

	groups := GroupTODOs(todos, GroupByDirectory)

	assert.Len(t, groups, 2)
	assert.Equal(t, "pkg/auth", groups[0].Name)
	assert.Len(t, groups[0].TODOs, 2)
	assert.Equal(t, "pkg/api", groups[1].Name)
	assert.Len(t, groups[1].TODOs, 1)
}

func TestGroupTODOs_MultiPath(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/a.go", "pkg/b.go"}, types.PriorityHigh, "multi"),
	}

	groups := GroupTODOs(todos, GroupByFile)

	assert.Len(t, groups, 2)
	assert.Equal(t, "pkg/a.go", groups[0].Name)
	assert.Equal(t, "pkg/b.go", groups[1].Name)
}

func TestGroupTODOs_None(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/a.go"}, types.PriorityHigh, "a"),
	}

	groups := GroupTODOs(todos, GroupByNone)

	assert.Len(t, groups, 1)
	assert.Empty(t, groups[0].Name)
	assert.Len(t, groups[0].TODOs, 1)
}

func TestGroupTODOs_EmptyGroupBy(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/a.go"}, types.PriorityHigh, "a"),
	}

	groups := GroupTODOs(todos, "")

	assert.Len(t, groups, 1)
	assert.Empty(t, groups[0].Name)
}

func TestGroupTODOs_AllUngrouped(t *testing.T) {
	todos := types.TODOS{
		todo(nil, types.PriorityHigh, "a"),
		todo(nil, types.PriorityLow, "b"),
	}

	groups := GroupTODOs(todos, GroupByFile)

	assert.Len(t, groups, 1)
	assert.Equal(t, UngroupedLabel, groups[0].Name)
	assert.Len(t, groups[0].TODOs, 2)
}

func TestGroupTODOs_All(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"pkg/auth/login.go"}, types.PriorityHigh, "a"),
		todo(types.StringOrSlice{"pkg/api/handler.go"}, types.PriorityMedium, "b"),
		todo(nil, types.PriorityLow, "c"),
	}

	groups := GroupTODOs(todos, GroupByAll)

	assert.Len(t, groups, 1)
	assert.Equal(t, "All TODOs", groups[0].Name)
	assert.Len(t, groups[0].TODOs, 3)
}

func TestGroupTODOs_GroupSortedByPriority(t *testing.T) {
	todos := types.TODOS{
		todo(types.StringOrSlice{"low/a.go"}, types.PriorityLow, "a"),
		todo(types.StringOrSlice{"high/b.go"}, types.PriorityHigh, "b"),
		todo(types.StringOrSlice{"med/c.go"}, types.PriorityMedium, "c"),
	}

	groups := GroupTODOs(todos, GroupByDirectory)

	assert.Equal(t, "high", groups[0].Name)
	assert.Equal(t, "med", groups[1].Name)
	assert.Equal(t, "low", groups[2].Name)
}
