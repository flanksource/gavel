package runtime

import (
	"context"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
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
		Issue:       issueInput,
		RootSession: p.todoRootSessionInput(issueInput),
		Session: p.todoOperationSessionInput(&native.Issue{
			ID: issueInput.ID, Title: issueInput.Title,
		}, todoOperationSessionOptions{
			ID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-plan:"+issueInput.ID.String()+":supplied")),
			Operation: string(native.StepPlan), Provider: "human", CWD: p.workDir, Prompt: issueInput.Body,
		}),
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
