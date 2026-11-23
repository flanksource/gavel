package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

const (
	todoSessionType            = "todo"
	todoSessionAddVerification = "add-verification"
)

func (p *Provider) todoRootSessionInput(issue native.CreateIssueInput) captaindb.CreateSessionInput {
	return captaindb.CreateSessionInput{
		ID:            issue.ID,
		Source:        "gavel",
		Provider:      "todos",
		HostID:        "local", // The TODO root can receive operations from any host.
		Project:       p.workspace.RepoKey,
		CWD:           p.workDir,
		Title:         issue.Title,
		InitialPrompt: issue.Body,
		AgentType:     todoSessionType,
		Description:   "Gavel TODO " + issue.ID.String(),
		Metadata:      todoSessionMetadata(issue.ID, ""),
	}
}

func (p *Provider) todoRootSessionInputForIssue(ctx context.Context, issue *native.Issue) (captaindb.CreateSessionInput, error) {
	root := p.todoRootSessionInput(native.CreateIssueInput{ID: issue.ID, Title: issue.Title, Body: issue.Body})
	stored, err := p.captain.GetSession(ctx, issue.ID)
	if errors.Is(err, captaindb.ErrSessionNotFound) {
		return root, nil
	}
	if err != nil {
		return captaindb.CreateSessionInput{}, fmt.Errorf("load Captain root session for TODO %s: %w", issue.ID, err)
	}
	root.HostID = stored.HostID
	return root, nil
}

type todoOperationSessionOptions struct {
	ID        uuid.UUID
	Operation string
	Provider  string
	CWD       string
	Prompt    string
}

func (p *Provider) todoOperationSessionInput(issue *native.Issue, options todoOperationSessionOptions) captaindb.CreateSessionInput {
	options.Operation = strings.TrimSpace(options.Operation)
	return captaindb.CreateSessionInput{
		ID:              options.ID,
		Source:          "gavel",
		Provider:        options.Provider,
		HostID:          captaindb.LocalHostID(),
		ParentSessionID: &issue.ID,
		Project:         p.workspace.RepoKey,
		CWD:             options.CWD,
		Title:           fmt.Sprintf("%s · %s", issue.Title, options.Operation),
		InitialPrompt:   options.Prompt,
		AgentType:       options.Operation,
		Description:     fmt.Sprintf("Gavel TODO %s %s", issue.ID, options.Operation),
		Metadata:        todoSessionMetadata(issue.ID, options.Operation),
	}
}

// suppliedPlanSessions builds the session pair a human-supplied plan hangs off:
// the TODO root, and one deterministic per-issue "plan supplied" session so
// `todos create --plan` and every later `todos edit --plan` reuse a single row
// instead of accumulating one per edit.
func (p *Provider) suppliedPlanSessions(ctx context.Context, issue *native.Issue) (*captaindb.CreateSessionInput, *captaindb.CreateSessionInput, error) {
	root, err := p.todoRootSessionInputForIssue(ctx, issue)
	if err != nil {
		return nil, nil, err
	}
	session := p.todoOperationSessionInput(issue, todoOperationSessionOptions{
		ID:        suppliedPlanSessionID(issue.ID),
		Operation: string(native.StepPlan),
		Provider:  "human",
		CWD:       p.workDir,
		Prompt:    issue.Body,
	})
	return &root, &session, nil
}

func suppliedPlanSessionID(issueID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("gavel-todo-plan:"+issueID.String()+":supplied"))
}

func todoSessionMetadata(issueID uuid.UUID, operation string) map[string]any {
	tags := []string{todoSessionType}
	if operation = strings.TrimSpace(operation); operation != "" {
		tags = append(tags, operation)
	}
	return map[string]any{
		"tags":  tags,
		"links": map[string]string{"todo": issueID.String()},
	}
}
