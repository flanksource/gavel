package native

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/plans"
	"github.com/flanksource/captain/pkg/sessiontree"
	"github.com/google/uuid"
)

type InitialPlanApproval struct {
	ApprovedBy string
	Comment    string
}

type CreateIssuePlanInput struct {
	Issue       CreateIssueInput
	RootSession captaindb.CreateSessionInput
	Session     captaindb.CreateSessionInput
	Plan        captaindb.CreatePlanInput
	Revision    captaindb.AppendPlanRevisionInput
	Approval    *InitialPlanApproval
	Actor       string
}

// CreateIssueWithPlan creates a native issue and its manually supplied Captain
// plan in one Captain transaction: Captain ensures the TODO root and plan
// session, saves the plan (and approves its latest revision when requested),
// and Gavel creates and selects the issue in the plan links. A failed session,
// revision, selection, or approval leaves no issue or provenance behind.
func (c *LaunchCoordinator) CreateIssueWithPlan(ctx context.Context, input CreateIssuePlanInput) (*PersistedPlan, error) {
	if input.Issue.ID == uuid.Nil || input.RootSession.ID != input.Issue.ID {
		return nil, fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	result := &PersistedPlan{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		saved, issue, err := saveCreatedIssuePlan(ctx, captainTx, input)
		if err != nil {
			return err
		}
		result.Plan, result.Revision, result.Issue = saved.Plan, saved.Revision, issue
		if input.Approval == nil {
			return nil
		}
		result.Plan, result.Issue, err = approveCreatedIssuePlan(ctx, captainTx, input, saved.Plan.ID, issue)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// saveCreatedIssuePlan ensures the session tree, saves the plan, and creates the
// issue in the plan link. Without an initial approval the link also selects
// the plan; with one, selection waits for the approval link.
func saveCreatedIssuePlan(ctx context.Context, captainTx *captaindb.DB, input CreateIssuePlanInput) (plans.Outcome, *Issue, error) {
	sessions, err := sessiontree.EnsureTx(ctx, captainTx, []captaindb.CreateSessionInput{input.RootSession, input.Session})
	if err != nil {
		return plans.Outcome{}, nil, err
	}
	planInput := input.Plan
	planInput.SourceSessionID = sessions[1].ID
	var issue *Issue
	saved, err := plans.Save(ctx, captainTx, plans.SaveInput{
		Plan: planInput, Markdown: input.Revision.PlanMarkdown,
		Feedback: input.Revision.Feedback, CreatedBy: input.Revision.CreatedBy,
	}, func(ctx context.Context, tx *captaindb.DB, outcome plans.Outcome) error {
		repositoryTx, err := NewRepository(tx.Gorm())
		if err != nil {
			return err
		}
		if issue, err = repositoryTx.CreateIssue(ctx, input.Issue); err != nil {
			return err
		}
		if input.Approval != nil {
			return nil
		}
		issue, err = selectCreatedPlan(tx, issue, outcome.Plan, &EventInput{
			Kind: "plan_created_and_selected", Actor: input.Actor,
			Payload: map[string]any{"planId": outcome.Plan.ID, "revisionId": outcome.Revision.ID, "ordinal": 0},
		})
		return err
	})
	return saved, issue, err
}

func approveCreatedIssuePlan(
	ctx context.Context,
	captainTx *captaindb.DB,
	input CreateIssuePlanInput,
	planID uuid.UUID,
	created *Issue,
) (*captaindb.Plan, *Issue, error) {
	var issue *Issue
	approved, err := plans.Approve(ctx, captainTx, plans.ApproveInput{
		PlanID: planID, Latest: true,
		ApprovedBy: input.Approval.ApprovedBy, Comment: input.Approval.Comment,
	}, func(_ context.Context, tx *captaindb.DB, outcome plans.Outcome) error {
		event := approvalEvent(outcome.Plan, PlanSelectionAttachment{Actor: input.Actor})
		event.Kind = "plan_approved_and_selected"
		var err error
		issue, err = selectCreatedPlan(tx, created, outcome.Plan, event)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	return approved.Plan, issue, nil
}

func selectCreatedPlan(tx *captaindb.DB, created *Issue, plan *captaindb.Plan, event *EventInput) (*Issue, error) {
	locked, err := lockExecutionIssue(tx.Gorm(), created.ID)
	if err != nil {
		return nil, err
	}
	if err := selectPlanLocked(tx.Gorm(), locked, PlanAttachment{
		IssueID: created.ID, PlanID: plan.ID, ExpectedIssueVersion: created.Version, Actor: event.Actor,
	}, event); err != nil {
		return nil, err
	}
	return getIssue(tx.Gorm(), "id = ?", created.ID)
}
