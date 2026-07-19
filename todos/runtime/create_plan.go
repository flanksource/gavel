package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
)

func (p *Provider) createIssueWithPlan(
	ctx context.Context,
	issueInput native.CreateIssueInput,
	request todos.CreatePlanRequest,
) (*native.Issue, error) {
	markdown := strings.TrimSpace(request.Markdown)
	if markdown == "" {
		return nil, fmt.Errorf("plan markdown is required")
	}
	input := native.CreateIssuePlanInput{
		Issue: issueInput,
		Session: captaindb.CreateSessionInput{
			Source: "gavel", Provider: "human", HostID: captaindb.LocalHostID(),
			Project: p.workspace.RepoKey, CWD: p.workDir, Title: issueInput.Title,
			InitialPrompt: issueInput.Body, AgentType: "human",
			Description: "Plan supplied by gavel todos create",
		},
		Plan: captaindb.CreatePlanInput{
			Title: issueInput.Title, Variant: "primary", SpecProfile: "gavel.todo.plan",
		},
		Revision: captaindb.AppendPlanRevisionInput{PlanMarkdown: markdown, CreatedBy: "human"},
		Actor:    mutationActor,
	}
	if request.Approved {
		input.Approval = &native.InitialPlanApproval{
			ApprovedBy: "human",
			Comment:    "Reviewed and approved by gavel todos create",
		}
	}
	created, err := p.coordinator.CreateIssueWithPlan(ctx, input)
	if err != nil {
		return nil, err
	}
	return created.Issue, nil
}
