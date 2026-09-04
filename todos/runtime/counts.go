package runtime

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/types"
)

// CountByStatus reports how many of the workspace's issues resolve to each
// derived status.
//
// It is the aggregate counterpart of List: rather than decoding every issue
// body, parsing its markdown and running the per-issue execution decoration,
// it groups on the four columns the status is a function of and folds each
// group through todoStatusWithPlan — the same derivation List uses — so the two
// can never disagree. Only statuses with at least one issue appear in the map.
func (p *Provider) CountByStatus(ctx context.Context) (map[types.Status]int, error) {
	groups, err := p.repository.CountIssuesByStatus(ctx, p.workspace.ID)
	if err != nil {
		return nil, err
	}
	counts := make(map[types.Status]int, len(groups))
	for _, group := range groups {
		status := todoStatusWithPlan(
			group.Status,
			group.ExecutionState,
			group.StepKind,
			captaindb.PlanApprovalState(group.ApprovalState),
		)
		counts[status] += group.Count
	}
	return counts, nil
}
