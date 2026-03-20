package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/cmd/gavel/choose"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// selectTODOs presents an interactive multi-select for the given TODO list.
// Returns the selected TODOs, or nil if the user cancels.
func selectTODOs(todoList []*types.TODO, header string) ([]*types.TODO, error) {
	items := make([]string, len(todoList))
	for i, todo := range todoList {
		items[i] = formatTODOItem(todo)
	}

	detailFunc := func(i int) string {
		if i < 0 || i >= len(todoList) {
			return ""
		}
		return buildDetailOptions(todoList[i])
	}

	selected, err := choose.Run(items,
		choose.WithHeader(header),
		choose.WithLimit(0),
		choose.WithDetailFunc(detailFunc),
	)
	if err != nil {
		return nil, fmt.Errorf("interactive selection failed: %w", err)
	}
	if len(selected) == 0 {
		return nil, nil
	}

	result := make([]*types.TODO, len(selected))
	for i, idx := range selected {
		result[i] = todoList[idx]
	}
	return result, nil
}

func formatTODOItem(todo *types.TODO) string {
	title := todo.Title
	if title == "" {
		title = todo.Filename()
	}

	priority := prioritySymbol(todo.Priority)
	status := statusSymbol(todo.Status)

	var path string
	if len(todo.Path) > 0 {
		path = strings.Join([]string(todo.Path), ", ")
	}

	return choose.FormatTODOListItem(title, priority, status, path)
}

func buildDetailOptions(todo *types.TODO) string {
	title := todo.Title
	if title == "" {
		title = todo.Filename()
	}

	opts := choose.DetailOptions{
		Title:    title,
		Priority: string(todo.Priority),
		Status:   string(todo.Status),
		Attempts: todo.Attempts,
		Language: string(todo.Language),
		Branch:   todo.Branch,
	}

	if todo.LastRun != nil {
		opts.LastRun = humanDuration(time.Since(*todo.LastRun))
	}

	if todo.PR != nil {
		opts.PRNumber = todo.PR.Number
		opts.PRAuthor = todo.PR.CommentAuthor
	}

	if len(todo.Path) > 0 {
		opts.Paths = []string(todo.Path)
	}

	opts.Tests = collectTestNames(todo.Verification)
	opts.Tests = append(opts.Tests, collectTestNames(todo.CustomValidations)...)

	if todo.Implementation != "" {
		opts.Implementation = todo.Implementation
	}

	return choose.FormatTODODetail(opts)
}

func collectTestNames(nodes []*fixtures.FixtureNode) []string {
	var names []string
	for _, node := range nodes {
		collectTestNamesRecursive(node, &names)
	}
	return names
}

func collectTestNamesRecursive(node *fixtures.FixtureNode, names *[]string) {
	if node.Test != nil && node.Test.Name != "" {
		*names = append(*names, node.Test.Name)
	}
	for _, child := range node.Children {
		collectTestNamesRecursive(child, names)
	}
}

func prioritySymbol(p types.Priority) string {
	switch p {
	case types.PriorityHigh:
		return "⚠ HIGH"
	case types.PriorityMedium:
		return "◉ MED"
	case types.PriorityLow:
		return "○ LOW"
	default:
		return ""
	}
}

func statusSymbol(s types.Status) string {
	switch s {
	case types.StatusFailed:
		return "✗ FAIL"
	case types.StatusPending:
		return "● PEND"
	case types.StatusInProgress:
		return "→ RUN"
	case types.StatusCompleted:
		return "✓ DONE"
	case types.StatusSkipped:
		return "⊘ SKIP"
	default:
		return ""
	}
}

func humanDuration(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
