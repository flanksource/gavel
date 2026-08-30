package todos

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// MaxBacklogEntries caps the duplicate-detection index. A very large backlog
// would otherwise crowd out the TODO actually being triaged, and the tail of a
// long list is the part a model attends to least.
const MaxBacklogEntries = 200

// BuildBacklogIndex renders the compact list of other open TODOs a triage run
// uses to spot duplicates. Entries being triaged are excluded — a TODO cannot
// duplicate itself, and seeing its own row invites the agent to say it does.
//
// Truncation is reported in the rendered text rather than applied silently: an
// agent told it is seeing the whole backlog will assert "no duplicate" with a
// confidence the truncated list does not support.
func BuildBacklogIndex(candidates []*types.TODO, exclude []*types.TODO) string {
	skip := map[string]bool{}
	for _, todo := range exclude {
		if todo == nil {
			continue
		}
		for _, ref := range []string{todo.ID, todo.ShortID} {
			if ref = strings.TrimSpace(ref); ref != "" {
				skip[ref] = true
			}
		}
	}

	var lines []string
	truncated := 0
	for _, todo := range candidates {
		if todo == nil || skip[todo.ID] || skip[todo.ShortID] {
			continue
		}
		if len(lines) >= MaxBacklogEntries {
			truncated++
			continue
		}
		lines = append(lines, backlogEntry(todo))
	}
	if len(lines) == 0 {
		return ""
	}
	index := strings.Join(lines, "\n")
	if truncated > 0 {
		index += fmt.Sprintf("\n\n(%d further TODOs are not listed; treat this index as partial when judging duplicates.)", truncated)
	}
	return index
}

func backlogEntry(todo *types.TODO) string {
	ref := strings.TrimSpace(todo.ShortID)
	if ref == "" {
		ref = todo.ID
	}
	title := strings.TrimSpace(todo.Title)
	if title == "" {
		title = "(untitled)"
	}
	status := strings.TrimSpace(string(todo.Status))
	if status == "" {
		status = "unknown"
	}
	priority := strings.TrimSpace(string(todo.Priority))
	if priority == "" {
		priority = "unset"
	}
	return fmt.Sprintf("- %s  %s  (%s, %s)", ref, title, status, priority)
}
