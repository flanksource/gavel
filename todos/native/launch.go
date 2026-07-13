package native

import (
	"context"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrDatabasePoolMismatch = errors.New("Captain and native TODO repositories use different database pools")

// PromptRunLaunchAttachment contains only Gavel-owned attachment metadata. The
// authoritative prompt-run UUID is obtained from Captain inside the launch
// transaction and cannot be supplied or guessed by the caller.
type PromptRunLaunchAttachment struct {
	IssueID              uuid.UUID
	StepKind             StepKind
	Ordinal              int
	ExpectedIssueVersion int64
	Actor                string
}

// PromptRunLaunch combines Captain's authoritative records with the native
// issue after all three were committed by one database transaction.
type PromptRunLaunch struct {
	Session       *captaindb.Session
	PromptRun     *captaindb.PromptRun
	Issue         *Issue
	DispatchOwned bool
}

// PlanSelectionAttachment is Gavel-owned selection metadata. The selected
// plan ID is always taken from Captain's CreateOrGetPlan result.
type PlanSelectionAttachment struct {
	IssueID              uuid.UUID
	Ordinal              int
	ExpectedIssueVersion int64
	Actor                string
}

// PersistedPlan combines Captain's durable plan/revision with the native issue
// that selected it.
type PersistedPlan struct {
	Plan     *captaindb.Plan
	Revision *captaindb.PlanRevision
	Issue    *Issue
}

// ApprovedPlanSelection is the cross-owner approval result. Captain remains
// authoritative for approval while Gavel stores only the selected plan link.
type ApprovedPlanSelection struct {
	Plan  *captaindb.Plan
	Issue *Issue
}

// ReviewedPlanSelection is the atomic result of a non-approval plan decision
// (pending, rejected, or revision requested) and the corresponding Gavel plan
// selection/event update.
type ReviewedPlanSelection struct {
	Plan  *captaindb.Plan
	Issue *Issue
}

// LaunchCoordinator is the native-only execution boundary. Only the
// PostgreSQL runtime constructs one.
//
// Captain and Gavel must share the exact underlying *sql.DB pool. This lets the
// coordinator use Captain's transaction handle for both owners' writes without
// Captain importing Gavel or writing a Gavel table itself.
type LaunchCoordinator struct {
	captain    *captaindb.DB
	repository *Repository
}

func NewLaunchCoordinator(captain *captaindb.DB, repository *Repository) (*LaunchCoordinator, error) {
	if captain == nil || captain.Gorm() == nil {
		return nil, fmt.Errorf("%w: Captain database is nil", ErrInvalidInput)
	}
	if repository == nil || repository.db == nil {
		return nil, fmt.Errorf("%w: native repository is nil", ErrInvalidInput)
	}
	captainSQL, captainErr := captain.Gorm().DB()
	repositorySQL, repositoryErr := repository.db.DB()
	if captainErr != nil || repositoryErr != nil || captainSQL != repositorySQL {
		return nil, ErrDatabasePoolMismatch
	}
	return &LaunchCoordinator{captain: captain, repository: repository}, nil
}

// LaunchPromptRun creates (or idempotently resolves) the Captain session and
// prompt run, then attaches and activates that run on exactly one native issue.
// Any failure rolls back newly-created Captain records together with the Gavel
// link and pointer. The current Captain state is projected before commit by
// ExecutionIntegration.ActivatePromptRun.
func (c *LaunchCoordinator) LaunchPromptRun(
	ctx context.Context,
	sessionInput captaindb.CreateSessionInput,
	promptRunInput captaindb.CreatePromptRunInput,
	attachment PromptRunLaunchAttachment,
) (*PromptRunLaunch, error) {
	if attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	requestedSessionID := promptRunInput.SessionID
	result := &PromptRunLaunch{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		session, err := captainTx.CreateOrGetSession(ctx, sessionInput)
		if err != nil {
			return err
		}
		if requestedSessionID != uuid.Nil && requestedSessionID != session.ID {
			return fmt.Errorf("%w: prompt run session does not match the authoritative session", ErrInvalidInput)
		}
		promptRunInput.SessionID = session.ID
		promptRun, err := captainTx.CreatePromptRun(ctx, promptRunInput)
		if err != nil {
			return err
		}

		repositoryTx, err := NewRepository(captainTx.Gorm())
		if err != nil {
			return err
		}
		integration, err := NewExecutionIntegration(repositoryTx)
		if err != nil {
			return err
		}
		issue, dispatchOwned, err := integration.activatePromptRun(ctx, PromptRunAttachment{
			IssueID:              attachment.IssueID,
			PromptRunID:          promptRun.ID,
			StepKind:             attachment.StepKind,
			Ordinal:              attachment.Ordinal,
			ExpectedIssueVersion: attachment.ExpectedIssueVersion,
			Actor:                attachment.Actor,
		})
		if err != nil {
			return err
		}
		result.Session = session
		result.PromptRun = promptRun
		result.Issue = issue
		result.DispatchOwned = dispatchOwned
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// PersistAndSelectPlan creates (or resolves) a Captain plan, appends its
// immutable content revision, and links/selects the plan on one native issue in
// a single transaction. A filesystem path in planInput remains optional source
// metadata; revisionInput.PlanMarkdown is the durable content.
func (c *LaunchCoordinator) PersistAndSelectPlan(
	ctx context.Context,
	planInput captaindb.CreatePlanInput,
	revisionInput captaindb.AppendPlanRevisionInput,
	attachment PlanSelectionAttachment,
) (*PersistedPlan, error) {
	if attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	requestedPlanID := revisionInput.PlanID
	result := &PersistedPlan{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		tx := captainTx.Gorm()
		issue, err := lockExecutionIssue(tx, attachment.IssueID)
		if err != nil {
			return err
		}
		priorPlan, err := existingPlanForCreate(ctx, captainTx, planInput)
		if err != nil {
			return err
		}
		var priorRevision *captaindb.PlanRevision
		priorRevisionIDs := map[uuid.UUID]struct{}{}
		selectionWasExact := false
		if priorPlan != nil {
			if err := lockCaptainPlan(tx, priorPlan.ID); err != nil {
				return err
			}
			// The identity lookup above only discovers the row to lock. Re-read
			// after acquiring the Captain row lock so revision exactness and the
			// changed/no-op decision share one stable snapshot.
			priorPlan, err = captainTx.GetPlan(ctx, priorPlan.ID)
			if err != nil {
				return err
			}
			if err := validateExistingPlanIdentity(priorPlan, planInput); err != nil {
				return err
			}
			if requestedPlanID != uuid.Nil && requestedPlanID != priorPlan.ID {
				return fmt.Errorf("%w: revision plan does not match the authoritative plan", ErrInvalidInput)
			}
			priorRevision, priorRevisionIDs, err = existingPlanRevision(ctx, captainTx, priorPlan.ID, revisionInput)
			if err != nil {
				return err
			}
			selectionWasExact, err = planSelectionExact(tx, issue, priorPlan.ID, attachment.Ordinal)
			if err != nil {
				return err
			}
		}
		if issue.Version != attachment.ExpectedIssueVersion {
			if priorPlan != nil && priorRevision != nil && selectionWasExact {
				currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
				if err != nil {
					return err
				}
				result.Plan = priorPlan
				result.Revision = priorRevision
				result.Issue = currentIssue
				return nil
			}
			return versionConflict(attachment.IssueID, attachment.ExpectedIssueVersion, issue.Version)
		}

		plan, err := captainTx.CreateOrGetPlan(ctx, planInput)
		if err != nil {
			return err
		}
		if requestedPlanID != uuid.Nil && requestedPlanID != plan.ID {
			return fmt.Errorf("%w: revision plan does not match the authoritative plan", ErrInvalidInput)
		}
		revisionInput.PlanID = plan.ID
		revision, err := captainTx.AppendPlanRevision(ctx, revisionInput)
		if err != nil {
			return err
		}
		_, revisionExisted := priorRevisionIDs[revision.ID]
		if !revisionExisted && priorPlan != nil && priorPlan.ApprovalState != captaindb.PlanApprovalPending {
			// A new immutable revision is never implicitly approved by an older
			// decision. Reset review state in the same shared transaction that
			// appends and selects the new content.
			plan, err = captainTx.SetPlanReviewState(ctx, captaindb.SetPlanReviewStateInput{
				PlanID: plan.ID, State: captaindb.PlanApprovalPending,
				Actor: attachment.Actor,
			})
			if err != nil {
				return err
			}
		}
		captainChanged := priorPlan == nil || !revisionExisted
		var mutation *EventInput
		if captainChanged {
			kind := "plan_revision_persisted"
			if !selectionWasExact {
				kind = "plan_persisted_and_selected"
			}
			mutation = &EventInput{
				Kind:  kind,
				Actor: attachment.Actor,
				Payload: map[string]any{
					"planId":     plan.ID,
					"revisionId": revision.ID,
					"revision":   revision.Revision,
					"ordinal":    attachment.Ordinal,
				},
			}
		}
		if err := selectPlanLocked(tx, issue, PlanAttachment{
			IssueID:              attachment.IssueID,
			PlanID:               plan.ID,
			Ordinal:              attachment.Ordinal,
			ExpectedIssueVersion: attachment.ExpectedIssueVersion,
			Actor:                attachment.Actor,
		}, mutation); err != nil {
			return err
		}
		plan, err = captainTx.GetPlan(ctx, plan.ID)
		if err != nil {
			return err
		}
		currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
		if err != nil {
			return err
		}
		result.Plan = plan
		result.Revision = revision
		result.Issue = currentIssue
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApproveAndSelectPlan updates Captain's approval state and links/selects that
// exact plan on one native issue in the same shared transaction. If Gavel's
// expected issue version loses a race, the Captain approval is rolled back.
// Replaying the same approval and selection is a no-op in both owners.
func (c *LaunchCoordinator) ApproveAndSelectPlan(
	ctx context.Context,
	approval captaindb.ApprovePlanRevisionInput,
	attachment PlanSelectionAttachment,
) (*ApprovedPlanSelection, error) {
	if attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	result := &ApprovedPlanSelection{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		tx := captainTx.Gorm()
		issue, err := lockExecutionIssue(tx, attachment.IssueID)
		if err != nil {
			return err
		}
		if err := lockCaptainPlan(tx, approval.PlanID); err != nil {
			return err
		}
		priorPlan, err := captainTx.GetPlan(ctx, approval.PlanID)
		if err != nil {
			return err
		}
		approvalWasExact := planApprovalExact(priorPlan, approval)
		selectionWasExact, err := planSelectionExact(tx, issue, priorPlan.ID, attachment.Ordinal)
		if err != nil {
			return err
		}
		if issue.Version != attachment.ExpectedIssueVersion {
			if approvalWasExact && selectionWasExact {
				currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
				if err != nil {
					return err
				}
				result.Plan = priorPlan
				result.Issue = currentIssue
				return nil
			}
			return versionConflict(attachment.IssueID, attachment.ExpectedIssueVersion, issue.Version)
		}

		plan, err := captainTx.ApprovePlanRevision(ctx, approval)
		if err != nil {
			return err
		}
		var mutation *EventInput
		if !approvalWasExact {
			kind := "plan_approval_changed"
			if !selectionWasExact {
				kind = "plan_approved_and_selected"
			}
			mutation = &EventInput{
				Kind:  kind,
				Actor: attachment.Actor,
				Payload: map[string]any{
					"planId":     plan.ID,
					"revisionId": approval.RevisionID,
					"approvedBy": strings.TrimSpace(approval.ApprovedBy),
					"comment":    strings.TrimSpace(approval.Comment),
					"ordinal":    attachment.Ordinal,
				},
			}
		}
		if err := selectPlanLocked(tx, issue, PlanAttachment{
			IssueID:              attachment.IssueID,
			PlanID:               plan.ID,
			Ordinal:              attachment.Ordinal,
			ExpectedIssueVersion: attachment.ExpectedIssueVersion,
			Actor:                attachment.Actor,
		}, mutation); err != nil {
			return err
		}
		currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
		if err != nil {
			return err
		}
		result.Plan = plan
		result.Issue = currentIssue
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReviewAndSelectPlan records a non-approval Captain review decision and keeps
// the exact plan selected on the Gavel issue in the same shared transaction.
// Exact retries are mutation-free in both owners.
func (c *LaunchCoordinator) ReviewAndSelectPlan(
	ctx context.Context,
	review captaindb.SetPlanReviewStateInput,
	attachment PlanSelectionAttachment,
) (*ReviewedPlanSelection, error) {
	if attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	result := &ReviewedPlanSelection{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		tx := captainTx.Gorm()
		issue, err := lockExecutionIssue(tx, attachment.IssueID)
		if err != nil {
			return err
		}
		if err := lockCaptainPlan(tx, review.PlanID); err != nil {
			return err
		}
		priorPlan, err := captainTx.GetPlan(ctx, review.PlanID)
		if err != nil {
			return err
		}
		reviewWasExact := planReviewExact(priorPlan, review)
		selectionWasExact, err := planSelectionExact(tx, issue, priorPlan.ID, attachment.Ordinal)
		if err != nil {
			return err
		}
		if issue.Version != attachment.ExpectedIssueVersion {
			if reviewWasExact && selectionWasExact {
				currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
				if err != nil {
					return err
				}
				result.Plan = priorPlan
				result.Issue = currentIssue
				return nil
			}
			return versionConflict(attachment.IssueID, attachment.ExpectedIssueVersion, issue.Version)
		}

		plan, err := captainTx.SetPlanReviewState(ctx, review)
		if err != nil {
			return err
		}
		var mutation *EventInput
		if !reviewWasExact {
			kind := "plan_review_changed"
			switch review.State {
			case captaindb.PlanApprovalRejected:
				kind = "plan_rejected"
			case captaindb.PlanApprovalRevisionRequested:
				kind = "plan_revision_requested"
			case captaindb.PlanApprovalPending:
				kind = "plan_review_pending"
			}
			mutation = &EventInput{
				Kind:  kind,
				Actor: attachment.Actor,
				Payload: map[string]any{
					"planId":  plan.ID,
					"state":   review.State,
					"actor":   strings.TrimSpace(review.Actor),
					"comment": strings.TrimSpace(review.Comment),
					"ordinal": attachment.Ordinal,
				},
			}
		}
		if err := selectPlanLocked(tx, issue, PlanAttachment{
			IssueID:              attachment.IssueID,
			PlanID:               plan.ID,
			Ordinal:              attachment.Ordinal,
			ExpectedIssueVersion: attachment.ExpectedIssueVersion,
			Actor:                attachment.Actor,
		}, mutation); err != nil {
			return err
		}
		currentIssue, err := getIssue(tx, "id = ?", attachment.IssueID)
		if err != nil {
			return err
		}
		result.Plan = plan
		result.Issue = currentIssue
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func existingPlanForCreate(ctx context.Context, db *captaindb.DB, input captaindb.CreatePlanInput) (*captaindb.Plan, error) {
	if input.SourcePromptRunID != nil && strings.TrimSpace(input.Variant) != "" {
		variant := strings.TrimSpace(input.Variant)
		plans, err := db.ListPlans(ctx, captaindb.PlanFilter{
			SourcePromptRunID: input.SourcePromptRunID,
			Variant:           &variant,
		})
		if err != nil {
			return nil, err
		}
		if len(plans) > 0 {
			plan := &plans[0]
			if err := validateExistingPlanIdentity(plan, input); err != nil {
				return nil, err
			}
			return plan, nil
		}
		return nil, nil
	}
	if input.ID == uuid.Nil {
		return nil, nil
	}
	plan, err := db.GetPlan(ctx, input.ID)
	if errors.Is(err, captaindb.ErrPlanNotFound) {
		return nil, nil
	}
	return plan, err
}

func validateExistingPlanIdentity(plan *captaindb.Plan, input captaindb.CreatePlanInput) error {
	if plan == nil {
		return nil
	}
	if plan.SourceSessionID != input.SourceSessionID {
		return fmt.Errorf("%w: existing plan belongs to session %s", captaindb.ErrPlanConflict, plan.SourceSessionID)
	}
	if input.ID != uuid.Nil && plan.ID != input.ID {
		return fmt.Errorf("%w: existing prompt-run variant has plan ID %s, not %s", captaindb.ErrPlanConflict, plan.ID, input.ID)
	}
	return nil
}

func lockCaptainPlan(tx *gorm.DB, planID uuid.UUID) error {
	var locked struct {
		ID string
	}
	result := tx.Raw(`SELECT id::text AS id FROM captain_plans WHERE id = ? FOR UPDATE`, planID).Scan(&locked)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %s", captaindb.ErrPlanNotFound, planID)
	}
	return nil
}

func existingPlanRevision(
	ctx context.Context,
	db *captaindb.DB,
	planID uuid.UUID,
	input captaindb.AppendPlanRevisionInput,
) (*captaindb.PlanRevision, map[uuid.UUID]struct{}, error) {
	revisions, err := db.ListPlanRevisions(ctx, planID)
	if err != nil {
		return nil, nil, err
	}
	ids := make(map[uuid.UUID]struct{}, len(revisions))
	requestedMarkdown := normalizePlanMarkdown(input.PlanMarkdown)
	requestedFeedback := strings.TrimSpace(input.Feedback)
	requestedBy := strings.TrimSpace(input.CreatedBy)
	var exact *captaindb.PlanRevision
	for i := range revisions {
		revision := &revisions[i]
		ids[revision.ID] = struct{}{}
		if normalizePlanMarkdown(revision.PlanMarkdown) == requestedMarkdown &&
			strings.TrimSpace(revision.Feedback) == requestedFeedback &&
			strings.TrimSpace(revision.CreatedBy) == requestedBy {
			exact = revision
		}
	}
	return exact, ids, nil
}

func normalizePlanMarkdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = strings.ReplaceAll(markdown, "\r", "\n")
	return strings.TrimSpace(markdown)
}

func planApprovalExact(plan *captaindb.Plan, input captaindb.ApprovePlanRevisionInput) bool {
	return plan != nil &&
		plan.ApprovalState == captaindb.PlanApprovalApproved &&
		plan.ApprovedRevisionID != nil && *plan.ApprovedRevisionID == input.RevisionID &&
		strings.TrimSpace(plan.ApprovedBy) == strings.TrimSpace(input.ApprovedBy) &&
		strings.TrimSpace(plan.ApprovalComment) == strings.TrimSpace(input.Comment)
}

func planReviewExact(plan *captaindb.Plan, input captaindb.SetPlanReviewStateInput) bool {
	return plan != nil &&
		plan.ApprovalState == input.State &&
		plan.ApprovedRevisionID == nil &&
		strings.TrimSpace(plan.ApprovedBy) == strings.TrimSpace(input.Actor) &&
		strings.TrimSpace(plan.ApprovalComment) == strings.TrimSpace(input.Comment)
}
