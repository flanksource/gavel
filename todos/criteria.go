package todos

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

const acceptanceCriteriaSection = "Acceptance Criteria"

// criteriaInsertBefore keeps the criteria section above the operational sections
// when it is first added to a body that already has them.
var criteriaInsertBefore = []string{
	"Steps to Reproduce", "Implementation", "Verification",
	"Custom Validations", "Latest Failure", "Verification Result",
	"Attempts", "Failure History",
}

// ParseAcceptanceCriteria reads the "## Acceptance Criteria" checklist from a
// markdown body into one criterion per task-list item, preserving each item's
// checked state.
func ParseAcceptanceCriteria(body string) []types.AcceptanceCriterion {
	var out []types.AcceptanceCriterion
	for _, line := range sectionLines(body, acceptanceCriteriaSection) {
		text, done, ok := parseChecklistItem(line)
		if !ok {
			continue
		}
		out = append(out, types.AcceptanceCriterion{Text: text, Done: done})
	}
	return out
}

// RenderCriteriaSection renders the criteria as a "## Acceptance Criteria"
// checklist that round-trips through ParseAcceptanceCriteria.
func RenderCriteriaSection(criteria []types.AcceptanceCriterion) string {
	var b strings.Builder
	b.WriteString("## " + acceptanceCriteriaSection + "\n\n")
	for _, c := range criteria {
		box := "[ ]"
		if c.Done {
			box = "[x]"
		}
		fmt.Fprintf(&b, "- %s %s\n", box, c.Text)
	}
	return b.String()
}

// UpsertCriteriaSection replaces (or inserts) the criteria section in body.
func UpsertCriteriaSection(body string, criteria []types.AcceptanceCriterion) string {
	return ReplaceOrAppendSection(body, acceptanceCriteriaSection, RenderCriteriaSection(criteria), criteriaInsertBefore...)
}

// sectionLines returns the lines inside the "## <header>" section (excluding the
// heading), stopping at the next "## " heading or end of body.
func sectionLines(body, header string) []string {
	want := "## " + header
	var out []string
	inSection := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == want || strings.HasPrefix(trimmed, want+" ") {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			out = append(out, line)
		}
	}
	return out
}

// parseChecklistItem extracts the text and checked state from a markdown list
// item ("- [ ] text", "- [x] text", or "- text"). ok is false for non-items.
func parseChecklistItem(line string) (text string, done bool, ok bool) {
	trimmed := strings.TrimSpace(line)
	for _, bullet := range []string{"- ", "* "} {
		if !strings.HasPrefix(trimmed, bullet) {
			continue
		}
		rest := strings.TrimSpace(trimmed[len(bullet):])
		switch {
		case strings.HasPrefix(rest, "[ ] "):
			return strings.TrimSpace(rest[4:]), false, true
		case strings.HasPrefix(rest, "[x] "), strings.HasPrefix(rest, "[X] "):
			return strings.TrimSpace(rest[4:]), true, true
		default:
			if rest != "" {
				return rest, false, true
			}
		}
	}
	return "", false, false
}
