package native

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PromptRunAttachment is the explicit identity returned by Captain for a
// single-issue execution. PromptRunID is Captain's durable prompt-run UUID; it
// is deliberately not inferred from a session ID or a timestamp.
type PromptRunAttachment struct {
	IssueID              uuid.UUID
	PromptRunID          uuid.UUID
	StepKind             StepKind
	Ordinal              int
	ExpectedIssueVersion int64
	Actor                string
}

// PlanAttachment is the explicit durable Captain plan selected for an issue.
// PlanID identifies captain_plans, whose revisions remain Captain-owned.
type PlanAttachment struct {
	IssueID              uuid.UUID
	PlanID               uuid.UUID
	Ordinal              int
	ExpectedIssueVersion int64
	Actor                string
}

// ExecutionIntegration is Gavel's application seam for connecting one native
// issue to Captain-owned execution records. It never creates or updates a
// Captain record. Callers must pass the IDs returned by Captain directly.
//
// Each operation is intentionally single-issue. A grouped Captain run must not
// be fanned out through this seam: todo_issue_prompt_runs enforces that a
// prompt run belongs to exactly one issue until grouped execution gets its own
// approved model.
type ExecutionIntegration struct {
	repository *Repository
}

func NewExecutionIntegration(repository *Repository) (*ExecutionIntegration, error) {
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("%w: native repository is nil", ErrInvalidInput)
	}
	return &ExecutionIntegration{repository: repository}, nil
}

// ActivatePromptRun atomically inserts the issue-to-run link and makes that
// linked run active. The pointer is written only after the link insert succeeds;
// any later failure rolls the link and pointer back together.
//
// Replaying an already-completed attachment is a no-op even when the caller is
// retrying with the original issue version after losing the first response.
func (s *ExecutionIntegration) ActivatePromptRun(ctx context.Context, input PromptRunAttachment) (*Issue, error) {
	if input.IssueID == uuid.Nil || input.PromptRunID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue and prompt run IDs are required", ErrInvalidInput)
	}
	if !input.StepKind.valid() {
		return nil, fmt.Errorf("%w: unsupported prompt step %q", ErrInvalidInput, input.StepKind)
	}
	if input.Ordinal < 0 {
		return nil, fmt.Errorf("%w: prompt run ordinal cannot be negative", ErrInvalidInput)
	}

	err := s.repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		issue, err := lockExecutionIssue(tx, input.IssueID)
		if err != nil {
			return err
		}

		var existing PromptRunLink
		result := tx.Raw(`
			SELECT issue_id, prompt_run_id, step_kind, ordinal, created_at
			FROM todo_issue_prompt_runs WHERE prompt_run_id = ?`, input.PromptRunID,
		).Scan(&existing)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			exact := existing.IssueID == input.IssueID &&
				existing.StepKind == input.StepKind && existing.Ordinal == input.Ordinal
			if !exact {
				return fmt.Errorf("%w: prompt run %s already has a different issue attachment", ErrLinkConflict, input.PromptRunID)
			}
			if sameUUIDPointer(issue.ActivePromptRunID, &input.PromptRunID) {
				return nil
			}
		}
		if issue.Version != input.ExpectedIssueVersion {
			return versionConflict(input.IssueID, input.ExpectedIssueVersion, issue.Version)
		}

		if result.RowsAffected == 0 {
			insert := tx.Exec(`
				INSERT INTO todo_issue_prompt_runs
					(issue_id, prompt_run_id, step_kind, ordinal, created_at)
				VALUES (?, ?, ?, ?, now())`,
				input.IssueID, input.PromptRunID, input.StepKind, input.Ordinal,
			)
			if insert.Error != nil {
				return mapUniqueError(insert.Error, ErrLinkConflict, "prompt run %s", input.PromptRunID)
			}
		}
		if err := tx.Exec(`
			UPDATE todo_issues SET active_prompt_run_id = ? WHERE id = ?`,
			input.PromptRunID, input.IssueID,
		).Error; err != nil {
			return err
		}
		_, err = recordMutation(tx, issue.lockedIssue(), EventInput{
			Kind:  "prompt_run_activated",
			Actor: input.Actor,
			Payload: map[string]any{
				"promptRunId": input.PromptRunID,
				"stepKind":    input.StepKind,
				"ordinal":     input.Ordinal,
			},
		})
		if err != nil {
			return err
		}

		// The Captain row usually exists before this attachment, so its INSERT
		// trigger could not find an issue to project. Reconcile the run's current
		// state now, in the attachment transaction, instead of waiting for a
		// later Captain update that may never arrive.
		return projectPromptRun(tx, input.PromptRunID)
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetIssue(ctx, input.IssueID)
}

// SelectPlan atomically inserts the issue-to-plan link and selects it. Plan
// content is not accepted here: the durable body and revisions live in Captain.
// An exact replay is a no-op, including a retry carrying the original version.
func (s *ExecutionIntegration) SelectPlan(ctx context.Context, input PlanAttachment) (*Issue, error) {
	if err := validatePlanAttachment(input); err != nil {
		return nil, err
	}

	err := s.repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		issue, err := lockExecutionIssue(tx, input.IssueID)
		if err != nil {
			return err
		}
		return selectPlanLocked(tx, issue, input, nil)
	})
	if err != nil {
		return nil, err
	}
	return s.repository.GetIssue(ctx, input.IssueID)
}

func validatePlanAttachment(input PlanAttachment) error {
	if input.IssueID == uuid.Nil || input.PlanID == uuid.Nil {
		return fmt.Errorf("%w: issue and plan IDs are required", ErrInvalidInput)
	}
	if input.Ordinal < 0 {
		return fmt.Errorf("%w: plan ordinal cannot be negative", ErrInvalidInput)
	}
	return nil
}

// selectPlanLocked applies a plan selection while the caller holds the issue's
// FOR UPDATE lock. captainMutation is non-nil when Captain changed durable plan
// state in the same outer transaction. In that case an already-selected plan
// still records a Gavel mutation/version; only a truly unchanged selection may
// use the stale exact-replay fast path.
func selectPlanLocked(tx *gorm.DB, issue *executionIssue, input PlanAttachment, captainMutation *EventInput) error {
	if err := validatePlanAttachment(input); err != nil {
		return err
	}
	var existing PlanLink
	result := tx.Raw(`
		SELECT issue_id, plan_id, ordinal, created_at
		FROM todo_issue_plans WHERE issue_id = ? AND plan_id = ?`,
		input.IssueID, input.PlanID,
	).Scan(&existing)
	if result.Error != nil {
		return result.Error
	}
	exactSelection := false
	if result.RowsAffected > 0 {
		if existing.Ordinal != input.Ordinal {
			return fmt.Errorf("%w: plan %s already has ordinal %d", ErrLinkConflict, input.PlanID, existing.Ordinal)
		}
		exactSelection = sameUUIDPointer(issue.SelectedPlanID, &input.PlanID)
		if exactSelection && captainMutation == nil {
			return nil
		}
	}
	if issue.Version != input.ExpectedIssueVersion {
		return versionConflict(input.IssueID, input.ExpectedIssueVersion, issue.Version)
	}

	if result.RowsAffected == 0 {
		insert := tx.Exec(`
			INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal, created_at)
			VALUES (?, ?, ?, now())`, input.IssueID, input.PlanID, input.Ordinal,
		)
		if insert.Error != nil {
			return mapUniqueError(insert.Error, ErrLinkConflict, "plan %s", input.PlanID)
		}
	}
	if !exactSelection {
		if err := tx.Exec(`
			UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`,
			input.PlanID, input.IssueID,
		).Error; err != nil {
			return err
		}
	}
	event := EventInput{
		Kind:  "plan_selected",
		Actor: input.Actor,
		Payload: map[string]any{
			"planId":  input.PlanID,
			"ordinal": input.Ordinal,
		},
	}
	if captainMutation != nil {
		event = *captainMutation
	}
	_, err := recordMutation(tx, issue.lockedIssue(), event)
	return err
}

func planSelectionExact(tx *gorm.DB, issue *executionIssue, planID uuid.UUID, ordinal int) (bool, error) {
	if issue.SelectedPlanID == nil || *issue.SelectedPlanID != planID {
		return false, nil
	}
	var count int64
	err := tx.Raw(`
		SELECT COUNT(*) FROM todo_issue_plans
		WHERE issue_id = ? AND plan_id = ? AND ordinal = ?`, issue.ID, planID, ordinal,
	).Scan(&count).Error
	return count == 1, err
}

type executionIssue struct {
	ID                uuid.UUID
	WorkspaceID       uuid.UUID
	Version           int64
	ActivePromptRunID *uuid.UUID
	SelectedPlanID    *uuid.UUID
}

func lockExecutionIssue(tx *gorm.DB, id uuid.UUID) (*executionIssue, error) {
	var issue executionIssue
	result := tx.Raw(`
		SELECT id, workspace_id, version, active_prompt_run_id, selected_plan_id
		FROM todo_issues WHERE id = ? FOR UPDATE`, id,
	).Scan(&issue)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: issue %s", ErrNotFound, id)
	}
	return &issue, nil
}

func (i *executionIssue) lockedIssue() *lockedIssue {
	return &lockedIssue{ID: i.ID, WorkspaceID: i.WorkspaceID, Version: i.Version}
}

func versionConflict(id uuid.UUID, expected, current int64) error {
	return fmt.Errorf("%w: issue %s expected version %d, current version %d", ErrVersionConflict, id, expected, current)
}

func projectPromptRun(tx *gorm.DB, promptRunID uuid.UUID) error {
	var changed int
	return tx.Raw(`SELECT public.gavel_project_todo_prompt_run(?)`, promptRunID).Scan(&changed).Error
}

func projectIssue(tx *gorm.DB, issueID uuid.UUID) error {
	var changed bool
	return tx.Raw(`SELECT public.gavel_project_todo_issue(?)`, issueID).Scan(&changed).Error
}
