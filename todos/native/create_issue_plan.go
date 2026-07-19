package native

import (
	"context"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
)

type InitialPlanApproval struct {
	ApprovedBy string
	Comment    string
}

type CreateIssuePlanInput struct {
	Issue    CreateIssueInput
	Session  captaindb.CreateSessionInput
	Plan     captaindb.CreatePlanInput
	Revision captaindb.AppendPlanRevisionInput
	Approval *InitialPlanApproval
	Actor    string
}

// CreateIssueWithPlan creates a native issue and its manually supplied Captain
// plan in one shared transaction. A failed session, revision, selection, or
// approval leaves no issue or provenance records behind.
func (c *LaunchCoordinator) CreateIssueWithPlan(ctx context.Context, input CreateIssuePlanInput) (*PersistedPlan, error) {
	result := &PersistedPlan{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		repositoryTx, err := NewRepository(captainTx.Gorm())
		if err != nil {
			return err
		}
		issue, err := repositoryTx.CreateIssue(ctx, input.Issue)
		if err != nil {
			return err
		}
		session, err := captainTx.CreateOrGetSession(ctx, input.Session)
		if err != nil {
			return err
		}
		input.Plan.SourceSessionID = session.ID
		plan, err := captainTx.CreateOrGetPlan(ctx, input.Plan)
		if err != nil {
			return err
		}
		input.Revision.PlanID = plan.ID
		revision, err := captainTx.AppendPlanRevision(ctx, input.Revision)
		if err != nil {
			return err
		}
		plan, mutation, err := approveInitialPlan(ctx, captainTx, plan, revision, input)
		if err != nil {
			return err
		}
		locked, err := lockExecutionIssue(captainTx.Gorm(), issue.ID)
		if err != nil {
			return err
		}
		if err := selectPlanLocked(captainTx.Gorm(), locked, PlanAttachment{
			IssueID: issue.ID, PlanID: plan.ID, ExpectedIssueVersion: issue.Version, Actor: input.Actor,
		}, mutation); err != nil {
			return err
		}
		result.Plan, result.Revision = plan, revision
		result.Issue, err = getIssue(captainTx.Gorm(), "id = ?", issue.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func approveInitialPlan(
	ctx context.Context,
	db *captaindb.DB,
	plan *captaindb.Plan,
	revision *captaindb.PlanRevision,
	input CreateIssuePlanInput,
) (*captaindb.Plan, *EventInput, error) {
	kind := "plan_created_and_selected"
	payload := map[string]any{"planId": plan.ID, "revisionId": revision.ID, "ordinal": 0}
	if input.Approval != nil {
		var err error
		plan, err = db.ApprovePlanRevision(ctx, captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID,
			ApprovedBy: input.Approval.ApprovedBy, Comment: input.Approval.Comment,
		})
		if err != nil {
			return nil, nil, err
		}
		kind = "plan_approved_and_selected"
		payload["approvedBy"] = strings.TrimSpace(input.Approval.ApprovedBy)
		payload["comment"] = strings.TrimSpace(input.Approval.Comment)
	}
	return plan, &EventInput{Kind: kind, Actor: input.Actor, Payload: payload}, nil
}
