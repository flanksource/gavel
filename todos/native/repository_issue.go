package native

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/flanksource/gavel/todos/labels"
)

type issueRecord struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	Title             string
	Body              string
	Verification      string
	Labels            pq.StringArray `gorm:"type:text[]"`
	Priority          string
	Status            string
	ExecutionState    string
	ActivePromptRunID *uuid.UUID
	SelectedPlanID    *uuid.UUID
	Version           int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (r issueRecord) issue() *Issue {
	labels := append([]string(nil), r.Labels...)
	if labels == nil {
		labels = []string{}
	}
	return &Issue{
		ID:                r.ID,
		WorkspaceID:       r.WorkspaceID,
		Title:             r.Title,
		Body:              r.Body,
		Verification:      r.Verification,
		Labels:            labels,
		Priority:          Priority(r.Priority),
		Status:            IssueStatus(r.Status),
		ExecutionState:    ExecutionState(r.ExecutionState),
		ActivePromptRunID: r.ActivePromptRunID,
		SelectedPlanID:    r.SelectedPlanID,
		Version:           r.Version,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

const issueColumns = `
	issue.id, issue.workspace_id, issue.title, issue.body, issue.verification,
	issue.labels, issue.priority, issue.status,
	COALESCE(runtime.execution_state, 'idle') AS execution_state,
	issue.active_prompt_run_id, issue.selected_plan_id, issue.version,
	issue.created_at, issue.updated_at`

const issueFrom = `
	todo_issues AS issue
	LEFT JOIN todo_issue_runtime AS runtime ON runtime.issue_id = issue.id`

func (r *Repository) CreateIssue(ctx context.Context, input CreateIssueInput) (*Issue, error) {
	if input.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace ID is required", ErrInvalidInput)
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: issue title is required", ErrInvalidInput)
	}
	priority := input.Priority
	if priority == "" {
		priority = PriorityMedium
	}
	if !priority.valid() {
		return nil, fmt.Errorf("%w: unsupported priority %q", ErrInvalidInput, priority)
	}
	status := input.Status
	if status == "" {
		status = StatusOpen
	}
	if !status.valid() {
		return nil, fmt.Errorf("%w: unsupported status %q", ErrInvalidInput, status)
	}
	aliases, err := normalizeAliases(input.Aliases)
	if err != nil {
		return nil, err
	}
	normalizedLabels := normalizeStrings(input.Labels)
	id := input.ID
	if id == uuid.Nil {
		id = uuid.New()
	}

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workspaceExists bool
		if err := tx.Raw(`SELECT EXISTS(SELECT 1 FROM todo_workspaces WHERE id = ?)`, input.WorkspaceID).Scan(&workspaceExists).Error; err != nil {
			return err
		}
		if !workspaceExists {
			return fmt.Errorf("%w: workspace %s", ErrNotFound, input.WorkspaceID)
		}

		result := tx.Exec(`
			INSERT INTO todo_issues
				(id, workspace_id, title, body, verification, labels, priority, status,
				 version, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, now(), now())`,
			id, input.WorkspaceID, title, input.Body, input.Verification, pq.Array(normalizedLabels),
			priority, status,
		)
		if result.Error != nil {
			return result.Error
		}
		if err := insertAliases(tx, input.WorkspaceID, id, aliases); err != nil {
			return err
		}
		_, err := insertEvent(tx, id, 1, EventInput{
			Kind:  "created",
			Actor: input.Actor,
			Payload: map[string]any{
				"title":     title,
				"labels":    normalizedLabels,
				"priority":  priority,
				"status":    status,
				"workspace": input.WorkspaceID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, id)
}

func (r *Repository) GetIssue(ctx context.Context, id uuid.UUID) (*Issue, error) {
	return getIssue(r.db.WithContext(ctx), `id = ?`, id)
}

// MoveIssueWorkspace transfers one issue between native workspaces without
// changing its identity or deleting its history and Captain links. Issues with
// relationships must be detached first because relationships are
// workspace-scoped.
func (r *Repository) MoveIssueWorkspace(
	ctx context.Context,
	issueID, targetWorkspaceID uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	if issueID == uuid.Nil || targetWorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue and target workspace IDs are required", ErrInvalidInput)
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sourceWorkspaceID, err := issueWorkspace(tx, issueID)
		if err != nil {
			return err
		}
		// Relationship mutations take this same workspace-scoped advisory lock.
		// Taking it before the issue row lock preserves their lock ordering.
		if err := lockWorkspaceRelationships(tx, sourceWorkspaceID); err != nil {
			return err
		}
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if locked.WorkspaceID != sourceWorkspaceID {
			return fmt.Errorf("%w: issue %s workspace changed during move", ErrVersionConflict, issueID)
		}
		if locked.WorkspaceID == targetWorkspaceID {
			return ErrNoChanges
		}

		var targetWorkspace struct{ ID uuid.UUID }
		targetResult := tx.Raw(`SELECT id FROM todo_workspaces WHERE id = ? FOR SHARE`, targetWorkspaceID).Scan(&targetWorkspace)
		if targetResult.Error != nil {
			return targetResult.Error
		}
		if targetResult.RowsAffected == 0 {
			return fmt.Errorf("%w: workspace %s", ErrNotFound, targetWorkspaceID)
		}

		var hasRelationships bool
		if err := tx.Raw(`
			SELECT EXISTS(
				SELECT 1 FROM todo_issue_relationships
				WHERE issue_id = ? OR target_issue_id = ?
			)`, issueID, issueID,
		).Scan(&hasRelationships).Error; err != nil {
			return err
		}
		if hasRelationships {
			return fmt.Errorf("%w: issue %s", ErrIssueHasRelationships, issueID)
		}

		var aliases []Alias
		if err := tx.Raw(`
			SELECT workspace_id, alias, issue_id, COALESCE(kind, '') AS kind, created_at
			FROM todo_issue_aliases
			WHERE issue_id = ?
			ORDER BY alias`, issueID,
		).Scan(&aliases).Error; err != nil {
			return err
		}
		var conflictingAlias string
		conflictResult := tx.Raw(`
			SELECT source.alias
			FROM todo_issue_aliases AS source
			JOIN todo_issue_aliases AS target
			  ON target.workspace_id = ? AND target.alias = source.alias
			WHERE source.issue_id = ?
			LIMIT 1`, targetWorkspaceID, issueID,
		).Scan(&conflictingAlias)
		if conflictResult.Error != nil {
			return conflictResult.Error
		}
		if conflictResult.RowsAffected > 0 {
			return fmt.Errorf("%w: workspace %s alias %q", ErrAliasConflict, targetWorkspaceID, conflictingAlias)
		}

		// The alias foreign key includes workspace_id and is immediate. Remove
		// and reinsert aliases inside this transaction so the issue and aliases
		// never expose different workspaces while retaining alias metadata.
		if err := tx.Exec(`DELETE FROM todo_issue_aliases WHERE issue_id = ?`, issueID).Error; err != nil {
			return err
		}
		moveResult := tx.Exec(`
			UPDATE todo_issues
			SET workspace_id = ?
			WHERE id = ? AND version = ?`, targetWorkspaceID, issueID, expectedVersion,
		)
		if moveResult.Error != nil {
			return moveResult.Error
		}
		if moveResult.RowsAffected != 1 {
			return fmt.Errorf("%w: issue %s expected version %d", ErrVersionConflict, issueID, expectedVersion)
		}
		for _, alias := range aliases {
			result := tx.Exec(`
				INSERT INTO todo_issue_aliases
					(workspace_id, alias, issue_id, kind, created_at)
				VALUES (?, ?, ?, NULLIF(?, ''), ?)`,
				targetWorkspaceID, alias.Alias, issueID, alias.Kind, alias.CreatedAt,
			)
			if result.Error != nil {
				return mapUniqueError(result.Error, ErrAliasConflict, "workspace %s alias %q", targetWorkspaceID, alias.Alias)
			}
		}

		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "workspace_moved",
			Actor: actor,
			Payload: map[string]any{
				"fromWorkspaceId": sourceWorkspaceID,
				"toWorkspaceId":   targetWorkspaceID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) UpdateIssue(ctx context.Context, id uuid.UUID, expectedVersion int64, patch IssuePatch) (*Issue, error) {
	updates := map[string]any{}
	payload := map[string]any{}
	if patch.Title != nil {
		title := strings.TrimSpace(*patch.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: issue title is required", ErrInvalidInput)
		}
		updates["title"] = title
		payload["title"] = title
	}
	if patch.Body != nil {
		updates["body"] = *patch.Body
		payload["body"] = *patch.Body
	}
	if patch.Verification != nil {
		updates["verification"] = *patch.Verification
		payload["verification"] = *patch.Verification
	}
	if patch.Labels != nil {
		normalizedLabels := normalizeStrings(*patch.Labels)
		updates["labels"] = pq.Array(normalizedLabels)
		payload["labels"] = normalizedLabels
	}
	if patch.Priority != nil {
		if !patch.Priority.valid() {
			return nil, fmt.Errorf("%w: unsupported priority %q", ErrInvalidInput, *patch.Priority)
		}
		updates["priority"] = *patch.Priority
		payload["priority"] = *patch.Priority
	}
	if patch.Status != nil {
		if !patch.Status.valid() {
			return nil, fmt.Errorf("%w: unsupported status %q", ErrInvalidInput, *patch.Status)
		}
		updates["status"] = *patch.Status
		payload["status"] = *patch.Status
	}
	if len(updates) == 0 {
		return nil, ErrNoChanges
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, id, expectedVersion)
		if err != nil {
			return err
		}
		current, err := getIssue(tx, `id = ?`, id)
		if err != nil {
			return err
		}
		pruneUnchangedIssueUpdates(current, updates, payload)
		if len(updates) == 0 {
			return nil
		}
		if err := tx.Table("todo_issues").Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:    "updated",
			Actor:   patch.Actor,
			Payload: payload,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, id)
}

func pruneUnchangedIssueUpdates(current *Issue, updates, payload map[string]any) {
	if value, ok := updates["title"].(string); ok && current.Title == value {
		delete(updates, "title")
		delete(payload, "title")
	}
	if value, ok := updates["body"].(string); ok && current.Body == value {
		delete(updates, "body")
		delete(payload, "body")
	}
	if value, ok := updates["verification"].(string); ok && current.Verification == value {
		delete(updates, "verification")
		delete(payload, "verification")
	}
	if value, ok := payload["labels"].([]string); ok && slices.Equal(current.Labels, value) {
		delete(updates, "labels")
		delete(payload, "labels")
	}
	if value, ok := updates["priority"].(Priority); ok && current.Priority == value {
		delete(updates, "priority")
		delete(payload, "priority")
	}
	if value, ok := updates["status"].(IssueStatus); ok && current.Status == value {
		delete(updates, "status")
		delete(payload, "status")
	}
}

// DeleteIssue performs an explicit hard delete. Normal workflow deletion
// should use StatusCancelled so append-only history remains available.
func (r *Repository) DeleteIssue(ctx context.Context, id uuid.UUID, expectedVersion int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockIssue(tx, id, expectedVersion); err != nil {
			return err
		}
		result := tx.Exec(`DELETE FROM todo_issues WHERE id = ?`, id)
		if result.Error != nil {
			return mapDeleteIssueError(result.Error, id)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: issue %s", ErrNotFound, id)
		}
		return nil
	})
}

// normalizeStrings normalizes, dedupes and sorts a label set. It defers to
// labels.Normalize so the form stored in todo_issues.labels is exactly the form
// a label definition is looked up by — and exactly what the todo_labels name
// check enforces in SQL.
func normalizeStrings(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value = labels.Normalize(value); value != "" {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
