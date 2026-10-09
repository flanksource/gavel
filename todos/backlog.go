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

// BacklogExcerptRunes bounds each entry's body excerpt. It is long enough to tell
// two superficially similar titles apart and short enough that two hundred of them
// do not bury the TODO being triaged. The excerpt is a hint that something is
// worth opening with `gavel todos get`, never the evidence itself.
const BacklogExcerptRunes = 200

// BuildBacklogIndex renders the compact list of other open TODOs a triage run
// uses to spot duplicates. Entries being triaged are excluded — a TODO cannot
// duplicate itself, and seeing its own row invites the agent to say it does.
//
// batch names the TODOs in the same triage request. Those rows are marked, because
// their verdicts are being decided concurrently: a fold naming one that has
// already been closed by its own verdict is refused, and the agent needs to know
// that before it proposes one.
//
// Truncation is reported in the rendered text rather than applied silently: an
// agent told it is seeing the whole backlog will assert "no duplicate" with a
// confidence the truncated list does not support.
func BuildBacklogIndex(candidates []*types.TODO, exclude []*types.TODO, batch []string) string {
	skip := refSet(exclude)
	inBatch := map[string]bool{}
	for _, ref := range batch {
		if ref = strings.TrimSpace(ref); ref != "" {
			inBatch[strings.ToLower(ref)] = true
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
		lines = append(lines, backlogEntry(todo, inBatch))
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

func refSet(todoList []*types.TODO) map[string]bool {
	refs := map[string]bool{}
	for _, todo := range todoList {
		if todo == nil {
			continue
		}
		for _, ref := range []string{todo.ID, todo.ShortID} {
			if ref = strings.TrimSpace(ref); ref != "" {
				refs[ref] = true
			}
		}
	}
	return refs
}

func backlogEntry(todo *types.TODO, inBatch map[string]bool) string {
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
	entry := fmt.Sprintf("- %s  %s  (%s, %s)", ref, title, status, priority)
	if inBatch[strings.ToLower(ref)] || inBatch[strings.ToLower(strings.TrimSpace(todo.ID))] {
		entry += "  [in this batch]"
	}
	if excerpt := backlogExcerpt(todo); excerpt != "" {
		entry += "\n  > " + excerpt
	}
	return entry
}

// backlogExcerpt is the opening of a TODO's description, flattened to one line.
//
// Without it the index is titles only, and two TODOs whose titles both say
// "parser crash" are indistinguishable — the agent either guesses or opens all of
// them. The newlines have to go: a multi-line excerpt breaks the one-entry-per-row
// shape the list depends on.
func backlogExcerpt(todo *types.TODO) string {
	body := strings.Join(strings.Fields(todo.MarkdownBody), " ")
	if body == "" {
		return ""
	}
	runes := []rune(body)
	if len(runes) <= BacklogExcerptRunes {
		return body
	}
	cut := string(runes[:BacklogExcerptRunes])
	// Cutting mid-word reads as a typo rather than a truncation.
	if idx := strings.LastIndex(cut, " "); idx > 0 {
		cut = cut[:idx]
	}
	return strings.TrimRight(cut, " ,;:.") + "…"
}
