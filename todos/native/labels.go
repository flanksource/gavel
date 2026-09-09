package native

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/flanksource/gavel/todos/labels"
)

// ErrLabelNotFound is returned when a definition to delete does not exist.
var ErrLabelNotFound = errors.New("native todo label definition not found")

// LabelDefinition is one stored label presentation. WorkspaceID is nil for a
// global definition applying to every workspace.
type LabelDefinition struct {
	ID          uuid.UUID  `json:"id" gorm:"column:id"`
	WorkspaceID *uuid.UUID `json:"workspaceId,omitempty" gorm:"column:workspace_id"`
	Name        string     `json:"name" gorm:"column:name"`
	Color       string     `json:"color" gorm:"column:color"`
	Icon        string     `json:"icon,omitempty" gorm:"column:icon"`
	Description string     `json:"description,omitempty" gorm:"column:description"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"column:updated_at"`
}

// LabelDefinitionInput upserts one definition. A nil WorkspaceID writes the
// global row.
type LabelDefinitionInput struct {
	WorkspaceID *uuid.UUID
	Name        string
	Color       string
	Icon        string
	Description string
}

const labelColumns = `id, workspace_id, name, color, icon, description, created_at, updated_at`

// ListLabelDefinitions reads a workspace's definitions and every global one in a
// single query.
//
// This is the only read path, and it is deliberately set-at-a-time: callers
// build one labels.Resolver per request and resolve every issue against it, so
// listing a backlog never issues a query per label. A per-row lookup here is the
// N+1 that made /api/projects take 46 seconds.
func (r *Repository) ListLabelDefinitions(ctx context.Context, workspaceID uuid.UUID) ([]LabelDefinition, error) {
	var rows []LabelDefinition
	// Global rows come first so that folding the result into a map by
	// last-one-wins yields the same precedence the resolver applies (workspace
	// shadows global). Ordering them last would silently invert it for any
	// caller that does the obvious thing.
	err := r.db.WithContext(ctx).Raw(`
		SELECT `+labelColumns+`
		FROM todo_labels
		WHERE workspace_id = ? OR workspace_id IS NULL
		ORDER BY workspace_id NULLS FIRST, name`, workspaceID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ListGlobalLabelDefinitions reads only the global rows.
func (r *Repository) ListGlobalLabelDefinitions(ctx context.Context) ([]LabelDefinition, error) {
	var rows []LabelDefinition
	err := r.db.WithContext(ctx).Raw(`
		SELECT ` + labelColumns + `
		FROM todo_labels
		WHERE workspace_id IS NULL
		ORDER BY name`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SetLabelDefinition upserts by scope and name, so saving from an editor is
// idempotent.
//
// The two partial unique indexes mean two different ON CONFLICT targets — a
// single statement cannot serve both scopes.
func (r *Repository) SetLabelDefinition(ctx context.Context, input LabelDefinitionInput) (*LabelDefinition, error) {
	definition := labels.Definition{
		Name:        labels.Normalize(input.Name),
		Color:       labels.Color(labels.Normalize(input.Color)),
		Icon:        labels.Normalize(input.Icon),
		Description: input.Description,
	}
	if err := definition.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, err)
	}
	if input.WorkspaceID != nil && *input.WorkspaceID == uuid.Nil {
		return nil, fmt.Errorf("%w: workspace ID cannot be the nil UUID", ErrInvalidInput)
	}

	query := `
		INSERT INTO todo_labels (workspace_id, name, color, icon, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, now(), now())
		ON CONFLICT (workspace_id, name) WHERE workspace_id IS NOT NULL
		DO UPDATE SET color = EXCLUDED.color, icon = EXCLUDED.icon,
		              description = EXCLUDED.description, updated_at = now()
		RETURNING ` + labelColumns
	if input.WorkspaceID == nil {
		query = `
		INSERT INTO todo_labels (workspace_id, name, color, icon, description, created_at, updated_at)
		VALUES (NULL, ?, ?, ?, ?, now(), now())
		ON CONFLICT (name) WHERE workspace_id IS NULL
		DO UPDATE SET color = EXCLUDED.color, icon = EXCLUDED.icon,
		              description = EXCLUDED.description, updated_at = now()
		RETURNING ` + labelColumns
	}

	args := []any{definition.Name, string(definition.Color), definition.Icon, definition.Description}
	if input.WorkspaceID != nil {
		args = append([]any{*input.WorkspaceID}, args...)
	}

	var row LabelDefinition
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// DeleteLabelDefinition removes one label, and what "remove" means depends on
// the scope it is removed from.
//
// A workspace-scoped removal retires the label from that project: the stored
// row goes AND the label is stripped from every TODO in the workspace. That is
// the whole point of a per-project taxonomy — a label removed from the project
// should not linger on forty TODOs, resurrecting itself in every facet the next
// time someone opens the backlog.
//
// A global removal drops only the shared presentation. It cannot strip TODO
// content, because "global" spans every workspace on the machine and a colour
// edit must never become a cross-project data deletion. The label falls back to
// its built-in or hashed hue and stays on its TODOs.
//
// Both halves of a workspace removal run in one transaction: a definition that
// survived a failed strip would leave the project half-retired.
func (r *Repository) DeleteLabelDefinition(ctx context.Context, workspaceID *uuid.UUID, name string) (labels.Removal, error) {
	normalized := labels.Normalize(name)
	if normalized == "" {
		return labels.Removal{}, fmt.Errorf("%w: label name is required", ErrInvalidInput)
	}
	removal := labels.Removal{Name: normalized}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		query := `DELETE FROM todo_labels WHERE workspace_id = ? AND name = ?`
		args := []any{workspaceID, normalized}
		if workspaceID == nil {
			query = `DELETE FROM todo_labels WHERE workspace_id IS NULL AND name = ?`
			args = []any{normalized}
		}
		deleted := tx.Exec(query, args...)
		if deleted.Error != nil {
			return deleted.Error
		}
		removal.Definition = deleted.RowsAffected > 0

		if workspaceID == nil {
			return nil
		}
		stripped, err := stripLabelFromIssues(tx, *workspaceID, normalized)
		if err != nil {
			return err
		}
		removal.Todos = stripped
		return nil
	})
	if err != nil {
		return labels.Removal{}, err
	}

	// A name that resolved through a built-in has no stored row, so "no row
	// deleted" is only a miss when no TODO carried the label either.
	if removal.Empty() {
		return removal, fmt.Errorf("%w: %s", ErrLabelNotFound, normalized)
	}
	return removal, nil
}

// stripLabelFromIssues removes one label from every issue in a workspace and
// records the edit in each issue's history.
//
// It is one statement rather than a read-modify-write per issue: retiring a
// label from a large backlog is a single set-based UPDATE, not N transactions
// through the issue-patch path. The cost of doing it here is that the append-only
// invariants the patch path maintains have to be maintained by hand — the
// version bump each optimistic-lock reader depends on, and one 'updated' event
// per issue carrying the resulting label set, so history never shows a label
// vanishing with no event behind it.
func stripLabelFromIssues(tx *gorm.DB, workspaceID uuid.UUID, label string) (int64, error) {
	// The event's sequence is read before this statement's own inserts are
	// visible, which is exactly right: each issue gets one event here, so
	// max+1 cannot collide with itself.
	result := tx.Exec(`
		WITH stripped AS (
			UPDATE todo_issues
			   SET labels = array_remove(labels, ?),
			       version = version + 1,
			       updated_at = now()
			 WHERE workspace_id = ?
			   AND labels @> ARRAY[?]::text[]
			RETURNING id, labels
		)
		INSERT INTO todo_issue_events (issue_id, sequence, kind, payload, source, created_at)
		SELECT stripped.id,
		       COALESCE((
		           SELECT MAX(event.sequence)
		           FROM todo_issue_events AS event
		           WHERE event.issue_id = stripped.id
		       ), 0) + 1,
		       'updated',
		       jsonb_build_object('labels', to_jsonb(stripped.labels)),
		       'gavel',
		       now()
		FROM stripped`, label, workspaceID, label)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// CountIssuesByLabel returns how many issues in the workspace carry each label,
// including labels nothing defines. One aggregate over the unnested array, never
// a query per label — it feeds the dashboard's tag facet counts and the Todos
// column of `gavel todos labels list`.
func (r *Repository) CountIssuesByLabel(ctx context.Context, workspaceID uuid.UUID) (map[string]int, error) {
	var rows []struct {
		Label string `gorm:"column:label"`
		Count int    `gorm:"column:count"`
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT label, COUNT(*) AS count
		FROM todo_issues AS issue, unnest(issue.labels) AS label
		WHERE issue.workspace_id = ?
		GROUP BY label
		ORDER BY label`, workspaceID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.Label] = row.Count
	}
	return counts, nil
}
