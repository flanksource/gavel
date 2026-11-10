package native

import (
	"context"

	"github.com/google/uuid"
)

func (r *Repository) LabelDefinitionsForWorkspaces(ctx context.Context, ids []uuid.UUID) ([]LabelDefinition, error) {
	var rows []LabelDefinition
	err := r.db.WithContext(ctx).Raw(`SELECT `+labelColumns+` FROM todo_labels
  WHERE org_id = ? AND (workspace_id IN ? OR workspace_id IS NULL)
  ORDER BY workspace_id NULLS FIRST, name`, r.org, ids).Scan(&rows).Error
	return rows, err
}

func (r *Repository) LabelCountsForWorkspaces(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]map[string]int, error) {
	var rows []struct {
		WorkspaceID uuid.UUID
		Label       string
		Count       int
	}
	visible, args := visibleIssue(ctx, "issue")
	err := r.db.WithContext(ctx).Raw(`SELECT issue.workspace_id, label, count(*) AS count
  FROM todo_issues issue, unnest(issue.labels) AS label
  WHERE issue.workspace_id IN ? AND `+inOrgWorkspace("issue.workspace_id")+` AND `+visible+`
  GROUP BY issue.workspace_id, label`, append([]any{ids, r.org}, args...)...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]map[string]int, len(ids))
	for _, row := range rows {
		if result[row.WorkspaceID] == nil {
			result[row.WorkspaceID] = map[string]int{}
		}
		result[row.WorkspaceID][row.Label] = row.Count
	}
	return result, nil
}
