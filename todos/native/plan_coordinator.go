package native

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/plans"
	"github.com/flanksource/captain/pkg/sessiontree"
	"github.com/google/uuid"
)

// PlanSelectionAttachment is Gavel-owned selection metadata. The selected
// plan ID is always taken from Captain's plan operation outcome.
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

// PersistPlanInput is one plan persistence: the Captain plan identity, its new
// immutable content, and the Gavel-owned selection.
//
// RootSession/Session are required exactly when Plan.SourceSessionID is unset —
// a plan an agent run produced already has its run's session, while a
// human-authored first plan has no session yet and Captain refuses to create a
// plan without one. Captain ensures them inside the plan transaction so a
// failed revision leaves no orphaned session behind.
type PersistPlanInput struct {
	Plan        captaindb.CreatePlanInput
	Revision    captaindb.AppendPlanRevisionInput
	Attachment  PlanSelectionAttachment
	RootSession *captaindb.CreateSessionInput
	Session     *captaindb.CreateSessionInput
}

// ApprovedPlanSelection is the cross-owner approval result. Captain remains
// authoritative for approval while Gavel stores only the selected plan link.
type ApprovedPlanSelection struct {
	Plan  *captaindb.Plan
	Issue *Issue
}

// ReviewedPlan is the atomic result of a non-approval plan decision and its
// corresponding Gavel selection/event update.
type ReviewedPlan struct {
	Plan  *captaindb.Plan
	Issue *Issue
}

// PersistAndSelectPlan saves the plan revision through Captain's plans service
// and, in its Link, selects the plan on one native issue. Captain takes the
// plan row lock first; the issue row is locked inside the link. An exact
// content replay whose selection already holds is a no-op even with a stale
// issue version; any other stale caller rolls the revision back.
func (c *LaunchCoordinator) PersistAndSelectPlan(ctx context.Context, input PersistPlanInput) (*PersistedPlan, error) {
	if input.Attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	if err := validatePlanSessionBootstrap(input); err != nil {
		return nil, err
	}
	result := &PersistedPlan{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		planInput := input.Plan
		if planInput.SourceSessionID == uuid.Nil {
			sessions, err := sessiontree.EnsureTx(ctx, captainTx, []captaindb.CreateSessionInput{*input.RootSession, *input.Session})
			if err != nil {
				return err
			}
			planInput.SourceSessionID = sessions[1].ID
		}
		_, err := plans.Save(ctx, captainTx, plans.SaveInput{
			Plan: planInput, Markdown: input.Revision.PlanMarkdown,
			Feedback: input.Revision.Feedback, CreatedBy: input.Revision.CreatedBy,
		}, func(ctx context.Context, tx *captaindb.DB, outcome plans.Outcome) error {
			if requested := input.Revision.PlanID; requested != uuid.Nil && requested != outcome.Plan.ID {
				return fmt.Errorf("%w: revision plan %s does not match the authoritative plan %s", ErrInvalidInput, requested, outcome.Plan.ID)
			}
			var mutation *EventInput
			if outcome.Changed {
				mutation = &EventInput{Actor: input.Attachment.Actor, Payload: map[string]any{
					"planId": outcome.Plan.ID, "revisionId": outcome.Revision.ID,
					"revision": outcome.Revision.Revision, "ordinal": input.Attachment.Ordinal,
				}}
			}
			issue, err := linkPlanSelection(tx, input.Attachment, outcome, mutation, "plan_revision_persisted", "plan_persisted_and_selected")
			if err != nil {
				return err
			}
			result.Plan, result.Revision, result.Issue = outcome.Plan, outcome.Revision, issue
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// validatePlanSessionBootstrap rejects a plan with neither a source session nor
// sessions to create one from, and a root session that is not the issue's.
func validatePlanSessionBootstrap(input PersistPlanInput) error {
	if input.Plan.SourceSessionID != uuid.Nil {
		return nil
	}
	if input.RootSession == nil || input.Session == nil {
		return fmt.Errorf("%w: a plan without a source session requires root and operation session inputs", ErrInvalidInput)
	}
	if input.RootSession.ID != input.Attachment.IssueID {
		return fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	return nil
}

// ApproveAndSelectPlan approves one immutable revision through Captain's plans
// service and selects that plan on one native issue in its Link. If Gavel's
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
	_, err := plans.Approve(ctx, c.captain, plans.ApproveInput{
		PlanID: approval.PlanID, RevisionID: approval.RevisionID,
		ApprovedBy: approval.ApprovedBy, Comment: approval.Comment,
	}, func(ctx context.Context, tx *captaindb.DB, outcome plans.Outcome) error {
		var mutation *EventInput
		if outcome.Changed {
			mutation = approvalEvent(outcome.Plan, attachment)
		}
		issue, err := linkPlanSelection(tx, attachment, outcome, mutation, "plan_approval_changed", "plan_approved_and_selected")
		if err != nil {
			return err
		}
		result.Plan, result.Issue = outcome.Plan, issue
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func approvalEvent(plan *captaindb.Plan, attachment PlanSelectionAttachment) *EventInput {
	return &EventInput{Actor: attachment.Actor, Payload: map[string]any{
		"planId": plan.ID, "revisionId": plan.ApprovedRevisionID,
		"approvedBy": strings.TrimSpace(plan.ApprovedBy), "comment": strings.TrimSpace(plan.ApprovalComment),
		"ordinal": attachment.Ordinal,
	}}
}

// linkPlanSelection is the Gavel half of a plan Save/Approve: it locks the
// issue (after Captain's plan lock) and selects the plan. A Captain replay
// whose selection already holds returns the current issue without touching it;
// otherwise mutation, when set, is recorded under changedKind if the plan was
// already selected and selectedKind if this link selects it.
func linkPlanSelection(
	tx *captaindb.DB,
	attachment PlanSelectionAttachment,
	outcome plans.Outcome,
	mutation *EventInput,
	changedKind, selectedKind string,
) (*Issue, error) {
	gormTx := tx.Gorm()
	issue, err := lockExecutionIssue(gormTx, attachment.IssueID)
	if err != nil {
		return nil, err
	}
	selectionExact, err := planSelectionExact(gormTx, issue, outcome.Plan.ID, attachment.Ordinal)
	if err != nil {
		return nil, err
	}
	if !outcome.Changed && selectionExact {
		return getIssue(gormTx, "id = ?", attachment.IssueID)
	}
	if mutation != nil {
		mutation.Kind = changedKind
		if !selectionExact {
			mutation.Kind = selectedKind
		}
	}
	if err := selectPlanLocked(gormTx, issue, PlanAttachment{
		IssueID: attachment.IssueID, PlanID: outcome.Plan.ID, Ordinal: attachment.Ordinal,
		ExpectedIssueVersion: attachment.ExpectedIssueVersion, Actor: attachment.Actor,
	}, mutation); err != nil {
		return nil, err
	}
	return getIssue(gormTx, "id = ?", attachment.IssueID)
}

// ReviewPlan records a non-approval review decision through Captain's plans
// service and applies its selection state in the Link. Rejected plans remain
// linked as history but are deselected; pending and revision-requested plans
// remain selected. Exact retries are mutation-free in both owners.
func (c *LaunchCoordinator) ReviewPlan(
	ctx context.Context,
	review captaindb.SetPlanReviewStateInput,
	attachment PlanSelectionAttachment,
) (*ReviewedPlan, error) {
	if attachment.IssueID == uuid.Nil {
		return nil, fmt.Errorf("%w: issue ID is required", ErrInvalidInput)
	}
	result := &ReviewedPlan{}
	_, err := plans.Review(ctx, c.captain, review, func(ctx context.Context, tx *captaindb.DB, outcome plans.Outcome) error {
		issue, err := linkPlanReview(tx, review, attachment, outcome)
		if err != nil {
			return err
		}
		result.Plan, result.Issue = outcome.Plan, issue
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func linkPlanReview(
	tx *captaindb.DB,
	review captaindb.SetPlanReviewStateInput,
	attachment PlanSelectionAttachment,
	outcome plans.Outcome,
) (*Issue, error) {
	gormTx := tx.Gorm()
	issue, err := lockExecutionIssue(gormTx, attachment.IssueID)
	if err != nil {
		return nil, err
	}
	selectionExact, err := planReviewSelectionExact(gormTx, issue, outcome.Plan.ID, attachment.Ordinal, review.State)
	if err != nil {
		return nil, err
	}
	if !outcome.Changed && selectionExact {
		return getIssue(gormTx, "id = ?", attachment.IssueID)
	}
	mutation := &EventInput{Kind: reviewEventKind(review.State), Actor: attachment.Actor, Payload: map[string]any{
		"planId": outcome.Plan.ID, "state": review.State,
		"actor": strings.TrimSpace(review.Actor), "comment": strings.TrimSpace(review.Comment),
		"ordinal": attachment.Ordinal,
	}}
	planAttachment := PlanAttachment{
		IssueID: attachment.IssueID, PlanID: outcome.Plan.ID, Ordinal: attachment.Ordinal,
		ExpectedIssueVersion: attachment.ExpectedIssueVersion, Actor: attachment.Actor,
	}
	if review.State == captaindb.PlanApprovalRejected {
		err = deselectPlanLocked(gormTx, issue, planAttachment, mutation)
	} else {
		err = selectPlanLocked(gormTx, issue, planAttachment, mutation)
	}
	if err != nil {
		return nil, err
	}
	return getIssue(gormTx, "id = ?", attachment.IssueID)
}

func reviewEventKind(state captaindb.PlanApprovalState) string {
	switch state {
	case captaindb.PlanApprovalRejected:
		return "plan_rejected"
	case captaindb.PlanApprovalRevisionRequested:
		return "plan_revision_requested"
	case captaindb.PlanApprovalPending:
		return "plan_review_pending"
	}
	return "plan_review_changed"
}
