package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var (
	_ todos.PlanReviewProvider   = (*Provider)(nil)
	_ todos.PlanRecoveryProvider = (*Provider)(nil)
	_ todos.PlanRevisionProvider = (*Provider)(nil)
)

func (p *Provider) RecoverPlan(ctx context.Context, todo *types.TODO) (*types.TODO, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	links, err := p.repository.ListPromptRuns(ctx, issueID)
	if err != nil {
		return nil, err
	}
	var latest *native.PromptRunLink
	for i := range links {
		link := &links[i]
		if link.StepKind != native.StepPlan {
			continue
		}
		if latest == nil || link.Ordinal > latest.Ordinal || (link.Ordinal == latest.Ordinal && link.CreatedAt.After(latest.CreatedAt)) {
			latest = link
		}
	}
	if latest == nil {
		return nil, fmt.Errorf("native TODO %s has no plan-step run to recover", issueID)
	}
	run, err := p.captain.GetPromptRun(ctx, latest.PromptRunID)
	if err != nil {
		return nil, err
	}
	if !terminalPromptRun(run.State) {
		return nil, fmt.Errorf("Captain plan run %s is still %s", run.ID, run.State)
	}
	status := types.PlanUpdated
	if issue.SelectedPlanID == nil {
		status = types.PlanNew
	}
	result := &todos.ExecutionResult{
		ExecutorName: run.Runtime.Driver,
		Plan:         &types.PlanResult{Status: status},
	}
	if err := p.persistPlanResult(ctx, todo, &activeRun{issue: issue, link: latest, run: run}, result); err != nil {
		return nil, err
	}
	return todo, nil
}

func (p *Provider) SavePlanRevision(ctx context.Context, todo *types.TODO, markdown, actor string) (*types.TODO, error) {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return nil, fmt.Errorf("plan markdown is required")
	}
	issue, plan, ordinal, err := p.selectedPlan(ctx, todo)
	if err != nil {
		return nil, err
	}
	actor = reviewActor(actor)
	persisted, err := p.coordinator.PersistAndSelectPlan(ctx, captaindb.CreatePlanInput{
		ID: plan.ID, SourceSessionID: plan.SourceSessionID,
		Title: plan.Title, Path: plan.Path, SpecProfile: plan.SpecProfile,
	}, captaindb.AppendPlanRevisionInput{
		PlanMarkdown: markdown, CreatedBy: actor,
	}, native.PlanSelectionAttachment{
		IssueID: issue.ID, Ordinal: ordinal,
		ExpectedIssueVersion: issue.Version, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	mapped, err := p.todoFromIssue(ctx, persisted.Issue, todo.CWD, true)
	if err != nil {
		return nil, err
	}
	*todo = *mapped
	return todo, nil
}

func (p *Provider) ApprovePlan(ctx context.Context, todo *types.TODO, actor, comment string) (*types.TODO, error) {
	issue, plan, ordinal, err := p.selectedPlan(ctx, todo)
	if err != nil {
		return nil, err
	}
	if plan.LatestRevision == nil {
		return nil, fmt.Errorf("Captain plan %s has no immutable revision to approve", plan.ID)
	}
	actor = reviewActor(actor)
	approved, err := p.coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
		PlanID: plan.ID, RevisionID: plan.LatestRevision.ID,
		ApprovedBy: actor, Comment: strings.TrimSpace(comment),
	}, native.PlanSelectionAttachment{
		IssueID: issue.ID, Ordinal: ordinal,
		ExpectedIssueVersion: issue.Version, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	mapped, err := p.todoFromIssue(ctx, approved.Issue, todo.CWD, true)
	if err != nil {
		return nil, err
	}
	*todo = *mapped
	return todo, nil
}

func (p *Provider) RejectPlan(ctx context.Context, todo *types.TODO, actor, comment string) (*types.TODO, error) {
	return p.setPlanReview(ctx, todo, captaindb.PlanApprovalRejected, actor, comment)
}

func (p *Provider) RequestPlanRevision(ctx context.Context, todo *types.TODO, actor, feedback string) (*types.TODO, error) {
	if strings.TrimSpace(feedback) == "" {
		return nil, fmt.Errorf("plan revision feedback is required")
	}
	return p.setPlanReview(ctx, todo, captaindb.PlanApprovalRevisionRequested, actor, feedback)
}

func (p *Provider) setPlanReview(
	ctx context.Context,
	todo *types.TODO,
	state captaindb.PlanApprovalState,
	actor, comment string,
) (*types.TODO, error) {
	issue, plan, ordinal, err := p.selectedPlan(ctx, todo)
	if err != nil {
		return nil, err
	}
	actor = reviewActor(actor)
	reviewed, err := p.coordinator.ReviewAndSelectPlan(ctx, captaindb.SetPlanReviewStateInput{
		PlanID: plan.ID, State: state, Actor: actor, Comment: strings.TrimSpace(comment),
	}, native.PlanSelectionAttachment{
		IssueID: issue.ID, Ordinal: ordinal,
		ExpectedIssueVersion: issue.Version, Actor: actor,
	})
	if err != nil {
		return nil, err
	}
	mapped, err := p.todoFromIssue(ctx, reviewed.Issue, todo.CWD, true)
	if err != nil {
		return nil, err
	}
	*todo = *mapped
	return todo, nil
}

func (p *Provider) selectedPlan(ctx context.Context, todo *types.TODO) (*native.Issue, *captaindb.Plan, int, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, nil, 0, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, nil, 0, err
	}
	if todo.Version != issue.Version {
		return nil, nil, 0, fmt.Errorf("%w: issue %s expected version %d, current version %d", native.ErrVersionConflict, issue.ID, todo.Version, issue.Version)
	}
	if issue.SelectedPlanID == nil {
		return nil, nil, 0, fmt.Errorf("native TODO %s has no selected Captain plan", issue.ID)
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return nil, nil, 0, err
	}
	links, err := p.repository.ListPlans(ctx, issue.ID)
	if err != nil {
		return nil, nil, 0, err
	}
	for _, link := range links {
		if link.PlanID == plan.ID {
			return issue, plan, link.Ordinal, nil
		}
	}
	return nil, nil, 0, fmt.Errorf("%w: selected plan %s is not linked to issue %s", native.ErrLinkConflict, plan.ID, issue.ID)
}

func reviewActor(actor string) string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return "human"
	}
	return actor
}
