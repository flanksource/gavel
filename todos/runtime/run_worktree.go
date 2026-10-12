package runtime

import (
	"context"

	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var _ lifecycle.RunWorktreeProvider = (*Provider)(nil)

// LatestRunWorktree is the worktree the todo's newest run step recorded, with
// its landing: the commit a verify step checks.
func (p *Provider) LatestRunWorktree(ctx context.Context, todo *types.TODO) (*native.RunWorktree, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	return p.repository.LatestRunWorktree(ctx, p.captain, issueID)
}
