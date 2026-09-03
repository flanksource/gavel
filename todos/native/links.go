package native

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// promptRunLinkSelect is the single column list every PromptRunLink read uses,
// so a new link column cannot be added to one query and forgotten in another.
const promptRunLinkSelect = `
	SELECT issue_id, prompt_run_id, step_kind, ordinal, created_at,
	       owner_host_id, owner_pid, owner_started_at, owner_token, owner_heartbeat_at`

func (r *Repository) LinkPromptRun(
	ctx context.Context,
	issueID, promptRunID uuid.UUID,
	stepKind StepKind,
	ordinal int,
	expectedVersion int64,
	actor string,
) (*PromptRunLink, error) {
	if issueID == uuid.Nil || promptRunID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue and prompt run IDs are required", ErrInvalidInput)
	}
	if !stepKind.valid() {
		return nil, fmt.Errorf("%w: unsupported prompt step %q", ErrInvalidInput, stepKind)
	}
	if ordinal < 0 {
		return nil, fmt.Errorf("%w: prompt run ordinal cannot be negative", ErrInvalidInput)
	}

	var link PromptRunLink
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		result := tx.Raw(`
			INSERT INTO todo_issue_prompt_runs
				(issue_id, prompt_run_id, step_kind, ordinal, created_at)
			VALUES (?, ?, ?, ?, now())
			RETURNING issue_id, prompt_run_id, step_kind, ordinal, created_at,
			          owner_host_id, owner_pid, owner_started_at, owner_token, owner_heartbeat_at`,
			issueID, promptRunID, stepKind, ordinal,
		).Scan(&link)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrLinkConflict, "prompt run %s", promptRunID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "prompt_run_linked",
			Actor: actor,
			Payload: map[string]any{
				"promptRunId": promptRunID,
				"stepKind":    stepKind,
				"ordinal":     ordinal,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *Repository) SetActivePromptRun(
	ctx context.Context,
	issueID uuid.UUID,
	promptRunID *uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		var current struct {
			ActivePromptRunID *uuid.UUID
		}
		if err := tx.Raw(`SELECT active_prompt_run_id FROM todo_issues WHERE id = ?`, issueID).Scan(&current).Error; err != nil {
			return err
		}
		if sameUUIDPointer(current.ActivePromptRunID, promptRunID) {
			if promptRunID == nil {
				return projectIssue(tx, issueID)
			}
			return nil
		}
		if promptRunID != nil {
			if *promptRunID == uuid.Nil {
				return fmt.Errorf("%w: active prompt run ID cannot be nil UUID", ErrInvalidInput)
			}
			var linked bool
			if err := tx.Raw(`
				SELECT EXISTS(
					SELECT 1 FROM todo_issue_prompt_runs
					WHERE issue_id = ? AND prompt_run_id = ?
				)`, issueID, *promptRunID,
			).Scan(&linked).Error; err != nil {
				return err
			}
			if !linked {
				return fmt.Errorf("%w: prompt run %s is not linked to issue %s", ErrLinkConflict, *promptRunID, issueID)
			}
		}
		if err := tx.Exec(`UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`, promptRunID, issueID).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "active_prompt_run_changed",
			Actor: actor,
			Payload: map[string]any{
				"promptRunId": promptRunID,
			},
		})
		if err != nil {
			return err
		}
		if promptRunID == nil {
			return projectIssue(tx, issueID)
		}
		return projectPromptRun(tx, *promptRunID)
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) UnlinkPromptRun(
	ctx context.Context,
	issueID, promptRunID uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
			UPDATE todo_issues SET active_prompt_run_id = NULL
			WHERE id = ? AND active_prompt_run_id = ?`, issueID, promptRunID,
		).Error; err != nil {
			return err
		}
		result := tx.Exec(`
			DELETE FROM todo_issue_prompt_runs
			WHERE issue_id = ? AND prompt_run_id = ?`, issueID, promptRunID,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: prompt run %s is not linked to issue %s", ErrNotFound, promptRunID, issueID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "prompt_run_unlinked",
			Actor: actor,
			Payload: map[string]any{
				"promptRunId": promptRunID,
			},
		})
		if err != nil {
			return err
		}
		return projectIssue(tx, issueID)
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) ListPromptRuns(ctx context.Context, issueID uuid.UUID) ([]PromptRunLink, error) {
	var links []PromptRunLink
	result := r.db.WithContext(ctx).Raw(promptRunLinkSelect+`
		FROM todo_issue_prompt_runs
		WHERE issue_id = ? ORDER BY step_kind, ordinal, created_at`, issueID,
	).Scan(&links)
	return links, result.Error
}

// ListPromptRunLinks reads a whole workspace's prompt run links in one query,
// keyed by issue, in the same order ListPromptRuns returns them. Listing
// decorates every issue with its attempt count and active step, so the
// per-issue ListPromptRuns is an N+1 across the entire backlog — the same
// reason ListAliasesByKind exists alongside ListAliases.
func (r *Repository) ListPromptRunLinks(ctx context.Context, workspaceID uuid.UUID) (map[uuid.UUID][]PromptRunLink, error) {
	var links []PromptRunLink
	result := r.db.WithContext(ctx).Raw(promptRunLinkSelect+`
		FROM todo_issue_prompt_runs AS link
		WHERE EXISTS (
			SELECT 1 FROM todo_issues AS issue
			WHERE issue.id = link.issue_id AND issue.workspace_id = ?
		)
		ORDER BY issue_id, step_kind, ordinal, created_at`, workspaceID,
	).Scan(&links)
	if result.Error != nil {
		return nil, result.Error
	}
	byIssue := make(map[uuid.UUID][]PromptRunLink, len(links))
	for _, link := range links {
		byIssue[link.IssueID] = append(byIssue[link.IssueID], link)
	}
	return byIssue, nil
}

func (r *Repository) LinkPlan(
	ctx context.Context,
	issueID, planID uuid.UUID,
	ordinal int,
	expectedVersion int64,
	actor string,
) (*PlanLink, error) {
	if issueID == uuid.Nil || planID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue and plan IDs are required", ErrInvalidInput)
	}
	if ordinal < 0 {
		return nil, fmt.Errorf("%w: plan ordinal cannot be negative", ErrInvalidInput)
	}

	var link PlanLink
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		result := tx.Raw(`
			INSERT INTO todo_issue_plans
				(issue_id, plan_id, ordinal, created_at)
			VALUES (?, ?, ?, now())
			RETURNING issue_id, plan_id, ordinal, created_at`,
			issueID, planID, ordinal,
		).Scan(&link)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrLinkConflict, "plan %s", planID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "plan_linked",
			Actor: actor,
			Payload: map[string]any{
				"planId":  planID,
				"ordinal": ordinal,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *Repository) SelectPlan(
	ctx context.Context,
	issueID uuid.UUID,
	planID *uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		var current struct {
			SelectedPlanID *uuid.UUID
		}
		if err := tx.Raw(`SELECT selected_plan_id FROM todo_issues WHERE id = ?`, issueID).Scan(&current).Error; err != nil {
			return err
		}
		if sameUUIDPointer(current.SelectedPlanID, planID) {
			return nil
		}
		if planID != nil {
			if *planID == uuid.Nil {
				return fmt.Errorf("%w: selected plan ID cannot be nil UUID", ErrInvalidInput)
			}
			var linked bool
			if err := tx.Raw(`
				SELECT EXISTS(
					SELECT 1 FROM todo_issue_plans
					WHERE issue_id = ? AND plan_id = ?
				)`, issueID, *planID,
			).Scan(&linked).Error; err != nil {
				return err
			}
			if !linked {
				return fmt.Errorf("%w: plan %s is not linked to issue %s", ErrLinkConflict, *planID, issueID)
			}
		}
		if err := tx.Exec(`UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`, planID, issueID).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "selected_plan_changed",
			Actor: actor,
			Payload: map[string]any{
				"planId": planID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) UnlinkPlan(
	ctx context.Context,
	issueID, planID uuid.UUID,
	expectedVersion int64,
	actor string,
) (*Issue, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		if err := tx.Exec(`
			UPDATE todo_issues SET selected_plan_id = NULL
			WHERE id = ? AND selected_plan_id = ?`, issueID, planID,
		).Error; err != nil {
			return err
		}
		result := tx.Exec(`
			DELETE FROM todo_issue_plans
			WHERE issue_id = ? AND plan_id = ?`, issueID, planID,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: plan %s is not linked to issue %s", ErrNotFound, planID, issueID)
		}
		_, err = recordMutation(tx, locked, EventInput{
			Kind:  "plan_unlinked",
			Actor: actor,
			Payload: map[string]any{
				"planId": planID,
			},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return r.GetIssue(ctx, issueID)
}

func (r *Repository) ListPlans(ctx context.Context, issueID uuid.UUID) ([]PlanLink, error) {
	var links []PlanLink
	result := r.db.WithContext(ctx).Raw(`
		SELECT issue_id, plan_id, ordinal, created_at
		FROM todo_issue_plans
		WHERE issue_id = ? ORDER BY ordinal, created_at`, issueID,
	).Scan(&links)
	return links, result.Error
}

func sameUUIDPointer(left, right *uuid.UUID) bool {
	switch {
	case left == nil || right == nil:
		return left == nil && right == nil
	default:
		return *left == *right
	}
}
