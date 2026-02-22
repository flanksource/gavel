package todos

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/todos/types"
)

const (
	GroupByFile      = "file"
	GroupByDirectory = "directory"
	GroupByNone      = "none"
	UngroupedLabel   = "Ungrouped"
)

type TODOGroup struct {
	Name  string      `json:"name"`
	TODOs types.TODOS `json:"todos"`
}

// GroupedRow is a virtual row for a single unified table.
// It's either a group header (GroupName set) or a TODO item.
type GroupedRow struct {
	GroupName string
	TODO      *types.TODO
}

func (r GroupedRow) PrettyRow(opts interface{}) map[string]api.Text {
	if r.GroupName != "" {
		return map[string]api.Text{
			"Title":    clicky.Text(fmt.Sprintf("── %s ──", r.GroupName), "text-blue-600 font-bold order-1"),
			"Status":   clicky.Text("", "order-2"),
			"Priority": clicky.Text("", "order-3"),
		}
	}
	return r.TODO.PrettyRow(opts)
}

// FlattenGrouped converts groups into a single flat slice with group header rows injected.
func FlattenGrouped(groups []TODOGroup) []GroupedRow {
	var rows []GroupedRow
	for _, g := range groups {
		if g.Name != "" {
			rows = append(rows, GroupedRow{GroupName: g.Name})
		}
		for _, t := range g.TODOs {
			rows = append(rows, GroupedRow{TODO: t})
		}
	}
	return rows
}

// GroupTODOs groups TODOs by the specified groupBy strategy.
// Returns ordered groups sorted by highest priority TODO in each group.
// TODOs without a path are placed in an "Ungrouped" group at the end.
func GroupTODOs(todos types.TODOS, groupBy string) []TODOGroup {
	if groupBy == "" || groupBy == GroupByNone {
		return []TODOGroup{{Name: "", TODOs: todos}}
	}

	grouped := map[string]types.TODOS{}
	var ungrouped types.TODOS

	for _, todo := range todos {
		keys := groupKeys(todo, groupBy)
		if len(keys) == 0 {
			ungrouped = append(ungrouped, todo)
			continue
		}
		for _, key := range keys {
			grouped[key] = append(grouped[key], todo)
		}
	}

	groups := make([]TODOGroup, 0, len(grouped)+1)
	for name, items := range grouped {
		items.Sort()
		groups = append(groups, TODOGroup{Name: name, TODOs: items})
	}

	sort.Slice(groups, func(i, j int) bool {
		pi := highestPriority(groups[i].TODOs)
		pj := highestPriority(groups[j].TODOs)
		if pi != pj {
			return pi < pj
		}
		return groups[i].Name < groups[j].Name
	})

	if len(ungrouped) > 0 {
		ungrouped.Sort()
		groups = append(groups, TODOGroup{Name: UngroupedLabel, TODOs: ungrouped})
	}

	return groups
}

func groupKeys(todo *types.TODO, groupBy string) []string {
	if len(todo.Path) == 0 {
		return nil
	}
	keys := make([]string, 0, len(todo.Path))
	for _, p := range todo.Path {
		switch groupBy {
		case GroupByFile:
			keys = append(keys, p)
		case GroupByDirectory:
			keys = append(keys, filepath.Dir(p))
		}
	}
	return keys
}

func highestPriority(todos types.TODOS) int {
	best := 999
	for _, t := range todos {
		if p := priorityRank(t.Priority); p < best {
			best = p
		}
	}
	return best
}

func priorityRank(p types.Priority) int {
	switch p {
	case types.PriorityHigh:
		return 0
	case types.PriorityMedium:
		return 1
	case types.PriorityLow:
		return 2
	default:
		return 999
	}
}
