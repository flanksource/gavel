package ui

import (
	"context"
	"slices"

	"github.com/google/uuid"
)

func (d *workspaceData) workspaceEntries(ids []uuid.UUID) []workspaceEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	entries := make([]workspaceEntry, 0, len(d.entries))
	for _, entry := range d.entries {
		if len(ids) == 0 || entry.Workspace != nil && slices.Contains(ids, entry.Workspace.ID) {
			entries = append(entries, entry)
		}
	}
	return entries
}

func (d *workspaceData) affectedWorkspaces(ctx context.Context, change workspaceChange) ([]uuid.UUID, error) {
	if change.All || change.Catalog || change.Resync {
		return nil, nil
	}
	ids := slices.Clone(change.Workspaces)
	if change.WorkspaceID != uuid.Nil {
		ids = append(ids, change.WorkspaceID)
	}
	if len(change.RunIDs) > 0 || len(change.SessionIDs) > 0 {
		var linked []uuid.UUID
		err := d.db.WithContext(ctx).Raw(`WITH changed AS MATERIALIZED (
 SELECT id, COALESCE(root_session_id,id) AS root, provider_session_id
 FROM captain_sessions WHERE id IN ?)
 SELECT DISTINCT i.workspace_id FROM todo_issue_prompt_runs l
 JOIN todo_issues i ON i.id=l.issue_id JOIN captain_prompt_runs r ON r.id=l.prompt_run_id
 WHERE i.workspace_id IN (SELECT id FROM todo_workspaces WHERE org_id=?) AND (
 r.id IN ? OR r.session_id IN ? OR r.execution_session_id IN ? OR r.root_session_id IN (SELECT root FROM changed)
 OR EXISTS (SELECT 1 FROM captain_sessions owner WHERE owner.id IN (r.session_id,r.execution_session_id)
 AND (COALESCE(owner.root_session_id,owner.id) IN (SELECT root FROM changed)
 OR owner.provider_session_id IN (SELECT provider_session_id FROM changed WHERE provider_session_id <> ''))))`,
			change.SessionIDs, d.org, change.RunIDs, change.SessionIDs, change.SessionIDs).Scan(&linked).Error
		if err != nil {
			return nil, err
		}
		ids = append(ids, linked...)
	}
	result := make([]uuid.UUID, 0, len(ids))
	for _, entry := range d.workspaceEntries(ids) {
		if entry.Workspace != nil && slices.Contains(ids, entry.Workspace.ID) && !slices.Contains(result, entry.Workspace.ID) {
			result = append(result, entry.Workspace.ID)
		}
	}
	return result, nil
}

func mergeWorkspaceChanges(previous, next workspaceChange) workspaceChange {
	next.Catalog = next.Catalog || previous.Catalog
	next.Resync = next.Resync || previous.Resync
	next.All = next.All || previous.All
	next.Workspaces = append(slices.Clone(next.Workspaces), previous.Workspaces...)
	if previous.WorkspaceID != uuid.Nil {
		next.Workspaces = append(next.Workspaces, previous.WorkspaceID)
	}
	next.RunIDs = append(slices.Clone(next.RunIDs), previous.RunIDs...)
	next.SessionIDs = append(slices.Clone(next.SessionIDs), previous.SessionIDs...)
	return next
}
