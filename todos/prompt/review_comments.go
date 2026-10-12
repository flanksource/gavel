package prompt

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// validateReviewComments refuses todos whose line comments cannot be read. The
// section renderer cannot return an error, so Render checks first: a run must
// not start with its review comments silently missing.
func validateReviewComments(todoList []*types.TODO) error {
	for _, todo := range todoList {
		if _, err := types.LineComments(todo.ProviderEvents); err != nil {
			return fmt.Errorf("todo %s review comments: %w", todo.ID, err)
		}
	}
	return nil
}

// anchoredComment reports whether a comment event is a line comment, which the
// review comments section renders instead of the plain comments section.
func anchoredComment(event types.ProviderEvent) bool {
	if len(event.Payload) == 0 {
		return false
	}
	var payload types.LineCommentPayload
	return json.Unmarshal(event.Payload, &payload) == nil && payload.Anchor != nil
}

// buildReviewCommentsSection renders the unresolved line comments reviewers left
// on a run's diff, grouped by file and ordered by line, each with the line it
// was left on. Resolved comments are done and stay out of the prompt.
func buildReviewCommentsSection(events []types.ProviderEvent) string {
	comments, err := types.LineComments(events)
	if err != nil {
		// Unreachable through Render, which refuses first; a caller that renders
		// sections directly sees the failure instead of a missing section.
		return fmt.Sprintf("## Review comments\n\nThe review comments could not be read: %v\n\n", err)
	}
	var open []types.LineComment
	for _, comment := range comments {
		if !comment.Resolved {
			open = append(open, comment)
		}
	}
	if len(open) == 0 {
		return ""
	}
	sort.SliceStable(open, func(i, j int) bool {
		if open[i].Anchor.Path != open[j].Anchor.Path {
			return open[i].Anchor.Path < open[j].Anchor.Path
		}
		return open[i].Anchor.Line < open[j].Anchor.Line
	})
	var section strings.Builder
	section.WriteString("## Review comments\n\n")
	for _, comment := range open {
		anchor := comment.Anchor
		author := comment.Event.Actor
		if author == "" {
			author = "unknown"
		}
		fmt.Fprintf(&section, "### `%s:%d` (%s side, at %s)\n\n```%s\n%s\n```\n\n**%s:** %s\n\n",
			anchor.Path, anchor.Line, anchor.Side, shortSHA(anchor.Commit),
			langFromExt(filepath.Ext(anchor.Path)), anchor.LineText, author, strings.TrimSpace(comment.Event.Body))
	}
	return section.String()
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}
