package todos

import (
	"context"

	"github.com/flanksource/gavel/todos/types"
)

// ToolWrite is one write a todo tool made during a run: which tool, and which
// TODO it changed. The run's history entry lists them so it shows what changed.
type ToolWrite struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
}

// MergeRequest folds Retired into Survivor: the survivor is rewritten with the
// combined content and every retired TODO is linked to it and closed.
type MergeRequest struct {
	Survivor     *types.TODO
	Retired      []*types.TODO
	Title        string
	Body         string
	Verification string
	Plan         string
	Rationale    string
	// Preview validates the merge exactly as an apply would and writes nothing.
	Preview bool
}

// MergeProvider merges TODOs through the store's merge path, refusing a merge
// that would lose acceptance criteria, attachments, a fixture or a plan, or
// that spans TODOs linked to different PRs.
type MergeProvider interface {
	MergeInto(ctx context.Context, request MergeRequest) error
}
