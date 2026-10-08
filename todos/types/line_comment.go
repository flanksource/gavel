package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// EventKindComment is a comment appended to a todo's history: free-form, or
	// anchored to one diff line of a run's branch when its payload carries a
	// LineCommentPayload anchor.
	EventKindComment = "comment"
	// EventKindCommentResolved records a comment being resolved or reopened; its
	// payload is a CommentResolution. The latest one for a comment wins.
	EventKindCommentResolved = "comment_resolved"
)

// DiffSide is the side of a diff a line comment is anchored to.
type DiffSide string

const (
	DiffSideOld DiffSide = "old"
	DiffSideNew DiffSide = "new"
)

// LineCommentAnchor pins a comment to one line of a run branch's diff: the file,
// the side and line number, the line's text when commented, and the commit range
// the diff was read from.
type LineCommentAnchor struct {
	Path      string   `json:"path"`
	Side      DiffSide `json:"side"`
	Line      int      `json:"line"`
	LineText  string   `json:"lineText"`
	Base      string   `json:"base,omitempty"`
	Commit    string   `json:"commit"`
	Branch    string   `json:"branch,omitempty"`
	AttemptID string   `json:"attemptId,omitempty"`
}

// Validate checks the fields an anchor cannot be placed without. LineText may be
// empty: a blank line can be commented on.
func (a LineCommentAnchor) Validate() error {
	switch {
	case strings.TrimSpace(a.Path) == "":
		return fmt.Errorf("line comment anchor: path is required")
	case a.Side != DiffSideOld && a.Side != DiffSideNew:
		return fmt.Errorf("line comment anchor: side must be %q or %q, got %q", DiffSideOld, DiffSideNew, a.Side)
	case a.Line <= 0:
		return fmt.Errorf("line comment anchor: line must be greater than zero, got %d", a.Line)
	case strings.TrimSpace(a.Commit) == "":
		return fmt.Errorf("line comment anchor: commit is required")
	}
	return nil
}

// LineCommentPayload is the payload of an anchored comment event.
type LineCommentPayload struct {
	Anchor *LineCommentAnchor `json:"anchor,omitempty"`
}

// CommentResolution is the payload of a comment_resolved event.
type CommentResolution struct {
	CommentID string `json:"commentId"`
	Resolved  bool   `json:"resolved"`
}

// LineComment is one anchored comment with its current resolution.
type LineComment struct {
	Event    ProviderEvent
	Anchor   LineCommentAnchor
	Resolved bool
}

// LineComments folds a todo's history into its anchored comments, in event
// order, with each comment's resolution taken from the latest comment_resolved
// event naming it. A comment without an anchor is not a line comment; an event
// whose payload does not decode is an error, never a comment silently dropped.
func LineComments(events []ProviderEvent) ([]LineComment, error) {
	var comments []LineComment
	index := map[string]int{}
	for _, event := range events {
		switch event.Kind {
		case EventKindComment:
			var payload LineCommentPayload
			if err := decodeEventPayload(event, &payload); err != nil {
				return nil, err
			}
			if payload.Anchor == nil {
				continue
			}
			index[event.ID] = len(comments)
			comments = append(comments, LineComment{Event: event, Anchor: *payload.Anchor})
		case EventKindCommentResolved:
			var resolution CommentResolution
			if err := decodeEventPayload(event, &resolution); err != nil {
				return nil, err
			}
			if i, ok := index[resolution.CommentID]; ok {
				comments[i].Resolved = resolution.Resolved
			}
		}
	}
	return comments, nil
}

func decodeEventPayload(event ProviderEvent, into any) error {
	if len(event.Payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(event.Payload, into); err != nil {
		return fmt.Errorf("decode %s event %s payload: %w", event.Kind, event.ID, err)
	}
	return nil
}
