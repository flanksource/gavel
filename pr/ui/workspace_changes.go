package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/flanksource/gavel/tenant"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
)

type workspaceChange struct {
	Catalog     bool `json:"catalog"`
	Resync      bool
	Err         error
	All         bool
	WorkspaceID uuid.UUID `json:"workspaceId"`
	Workspaces  []uuid.UUID
	RunIDs      []uuid.UUID
	SessionIDs  []uuid.UUID
}

func (d *workspaceData) refresh(ctx context.Context, workspaceIDs ...uuid.UUID) error {
	d.refreshMu.Lock()
	defer d.refreshMu.Unlock()
	ids := workspaceIDs
	if len(ids) == 0 {
		for _, entry := range d.workspaceEntries(nil) {
			if entry.Workspace != nil {
				ids = append(ids, entry.Workspace.ID)
			}
		}
	}
	if len(ids) == 0 {
		d.mu.Lock()
		changed := len(d.snapshot.Workspaces) > 0
		d.snapshot.Workspaces = map[string]workspaceRevision{}
		d.mu.Unlock()
		if changed {
			d.changed.Notify()
		}
		return nil
	}
	var rows []struct {
		WorkspaceID uuid.UUID
		Fingerprint string
	}
	err := d.db.WithContext(ctx).Raw(`SELECT i.workspace_id,
 md5(string_agg(concat_ws('|', i.id::text, i.version::text, i.updated_at::text,
 v.execution_state, v.approval_state, v.plan_revision::text,
 (v.session_summary - 'updatedAt' - 'durationMs')::text), ',' ORDER BY i.id)) AS fingerprint
 FROM todo_issues i JOIN todo_issue_overview v ON v.issue_id=i.id
 WHERE i.workspace_id IN ? AND i.workspace_id IN (SELECT id FROM todo_workspaces WHERE org_id=?)
 GROUP BY i.workspace_id`, ids, d.org).Scan(&rows).Error
	if err != nil {
		return fmt.Errorf("read workspace generations: %w", err)
	}
	fingerprints := map[uuid.UUID]string{}
	for _, row := range rows {
		fingerprints[row.WorkspaceID] = row.Fingerprint
	}
	var labels []struct {
		WorkspaceID *uuid.UUID
		Fingerprint string
	}
	if err := d.db.WithContext(ctx).Raw(`SELECT workspace_id, md5(string_agg(to_jsonb(l)::text, ',' ORDER BY id)) AS fingerprint FROM todo_labels l WHERE org_id=? AND (workspace_id IN ? OR workspace_id IS NULL) GROUP BY workspace_id`, d.org, ids).Scan(&labels).Error; err != nil {
		return err
	}
	global := ""
	taxonomy := map[uuid.UUID]string{}
	for _, row := range labels {
		if row.WorkspaceID == nil {
			global = row.Fingerprint
		} else {
			taxonomy[*row.WorkspaceID] = row.Fingerprint
		}
	}
	var histories []struct {
		WorkspaceID uuid.UUID
		Fingerprint string
	}
	if err := d.db.WithContext(ctx).Raw(`SELECT i.workspace_id, md5(string_agg(concat_ws('|',r.id::text,r.state::text,r.phase::text,r.started_at::text,r.finished_at::text,r.workspace::text,r.result_json::text),',' ORDER BY r.id)) AS fingerprint FROM todo_issue_prompt_runs link JOIN todo_issues i ON i.id=link.issue_id JOIN captain_prompt_runs r ON r.id=link.prompt_run_id WHERE i.workspace_id IN ? AND i.workspace_id IN (SELECT id FROM todo_workspaces WHERE org_id=?) GROUP BY i.workspace_id`, ids, d.org).Scan(&histories).Error; err != nil {
		return err
	}
	for _, row := range histories {
		fingerprints[row.WorkspaceID] += row.Fingerprint
	}
	var assignments []struct {
		WorkspaceID uuid.UUID
		Fingerprint string
	}
	if err := d.db.WithContext(ctx).Raw(`SELECT workspace_id, md5(string_agg(concat_ws('|',id::text,labels::text,visibility::text,created_by_user_id::text),',' ORDER BY id)) AS fingerprint FROM todo_issues WHERE workspace_id IN ? AND workspace_id IN (SELECT id FROM todo_workspaces WHERE org_id=?) GROUP BY workspace_id`, ids, d.org).Scan(&assignments).Error; err != nil {
		return err
	}
	for _, row := range assignments {
		taxonomy[row.WorkspaceID] += row.Fingerprint
	}
	d.mu.Lock()
	changed := false
	next := map[string]workspaceRevision{}
	for dir, entry := range d.entries {
		if entry.Workspace == nil {
			continue
		}
		id := entry.Workspace.ID
		previous, exists := d.snapshot.Workspaces[dir]
		if len(workspaceIDs) > 0 && !slices.Contains(workspaceIDs, id) {
			next[dir] = previous
			continue
		}
		fingerprint := fingerprints[id] + d.diffFingerprints[id]
		if !exists || previous.Fingerprint != fingerprint {
			previous.Revision++
			changed = true
		}
		labelFingerprint := global + taxonomy[id]
		if !exists || previous.LabelFingerprint != labelFingerprint {
			previous.Labels++
			changed = true
		}
		encoded, encodeErr := json.Marshal(entry.Options)
		if encodeErr != nil {
			d.mu.Unlock()
			return encodeErr
		}
		configFingerprint := string(encoded) + d.configChanges[dir]
		if !exists || previous.ConfigFingerprint != configFingerprint {
			previous.Config++
			changed = true
		}
		previous.ID = id
		previous.Dir = dir
		previous.Fingerprint = fingerprint
		previous.LabelFingerprint = labelFingerprint
		previous.ConfigFingerprint = configFingerprint
		next[dir] = previous
	}
	if len(next) != len(d.snapshot.Workspaces) {
		changed = true
	}
	d.snapshot.Workspaces = next
	d.mu.Unlock()
	if changed {
		d.changed.Notify()
	}
	return nil
}

func (s *Server) prepareWorkspaceDiffs(ctx context.Context, workspaceIDs ...uuid.UUID) error {
	d := s.workspaceData
	entries := d.workspaceEntries(workspaceIDs)
	ctx = tenant.WithScope(ctx, tenant.Scope{OrgID: s.orgID()})
	for _, entry := range entries {
		if entry.Workspace == nil {
			continue
		}
		provider, err := todoruntime.Resolved(ctx, d.db, entry.Options, entry.Workspace)
		if err != nil {
			return err
		}
		stats, readErr := s.prepareRunDiffStats(ctx, provider)
		encoded, err := json.Marshal(stats)
		if err != nil {
			return err
		}
		fingerprint := string(encoded)
		if readErr != nil {
			fingerprint = readErr.Error()
		}
		d.mu.Lock()
		d.diffs[entry.Workspace.ID] = stats
		d.diffErrors[entry.Workspace.ID] = readErr
		d.diffFingerprints[entry.Workspace.ID] = fingerprint
		d.mu.Unlock()
	}
	return nil
}

func (s *Server) reconcileWorkspaceData(ctx context.Context, workspaceIDs ...uuid.UUID) error {
	d := s.workspaceData
	entries := d.workspaceEntries(workspaceIDs)
	providers := make([]*todoruntime.Provider, 0, len(entries))
	ctx = tenant.WithScope(ctx, tenant.Scope{OrgID: s.orgID()})
	for _, entry := range entries {
		if entry.Workspace == nil {
			continue
		}
		provider, err := todoruntime.Resolved(ctx, d.db, entry.Options, entry.Workspace)
		if err != nil {
			return err
		}
		providers = append(providers, provider)
	}
	return todoruntime.ReconcileWorkspaces(ctx, providers)
}
