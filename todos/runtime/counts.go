package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
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
	counts, err := countWorkspaces(ctx, map[uuid.UUID]*Provider{p.workspace.ID: p})
	if err != nil {
		return nil, err
	}
	return counts[p.workspace.ID], nil
}

// countWorkspaces is CountByStatus for many workspaces at once, each read
// through its own provider: it reconciles answers given outside gavel across
// all of them, then counts all of them in one grouped query. The number of
// queries does not depend on how many workspaces are counted; only a parked ask
// that has to be examined costs one of its own.
//
// Every requested workspace appears in the result, with an empty map when it
// has no issues; asking for none issues no query.
func countWorkspaces(ctx context.Context, providers map[uuid.UUID]*Provider) (map[uuid.UUID]map[types.Status]int, error) {
	if len(providers) == 0 {
		return map[uuid.UUID]map[types.Status]int{}, nil
	}
	if err := reconcileWorkspaceAnswers(ctx, providers); err != nil {
		return nil, err
	}
	ids := slices.Collect(maps.Keys(providers))
	groups, err := providers[ids[0]].repository.CountIssuesByStatus(ctx, ids)
	if err != nil {
		return nil, err
	}
	counts := make(map[uuid.UUID]map[types.Status]int, len(ids))
	for _, id := range ids {
		counts[id] = map[types.Status]int{}
	}
	for _, group := range groups {
		byStatus, ok := counts[group.WorkspaceID]
		if !ok {
			return nil, fmt.Errorf("count TODOs: grouped count returned workspace %s, which was not asked for", group.WorkspaceID)
		}
		status := todoStatusWithPlan(
			group.Status,
			group.ExecutionState,
			group.StepKind,
			captaindb.PlanApprovalState(group.ApprovalState),
		)
		byStatus[status] += group.Count
	}
	return counts, nil
}
