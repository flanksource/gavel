package native

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

type CreateIssueSessionInput struct {
	Issue       CreateIssueInput
	RootSession captaindb.CreateSessionInput
}

type CreatedIssueSession struct {
	Issue *Issue
	Root  *captaindb.Session
}

type UpdateIssueSessionInput struct {
	IssueID              uuid.UUID
	ExpectedIssueVersion int64
	Patch                IssuePatch
	RootSession          captaindb.CreateSessionInput
	Session              captaindb.CreateSessionInput
}

func (c *LaunchCoordinator) CreateIssueWithSession(
	ctx context.Context,
	input CreateIssueSessionInput,
) (*CreatedIssueSession, error) {
	if input.Issue.ID == uuid.Nil || input.RootSession.ID != input.Issue.ID {
		return nil, fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	if input.RootSession.ParentSessionID != nil || input.RootSession.RootSessionID != nil {
		return nil, fmt.Errorf("%w: TODO root session must be canonical", ErrInvalidInput)
	}
	result := &CreatedIssueSession{}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		repositoryTx, err := NewRepository(captainTx.Gorm())
		if err != nil {
			return err
		}
		result.Issue, err = repositoryTx.CreateIssue(ctx, input.Issue)
		if err != nil {
			return err
		}
		result.Root, err = captainTx.CreateOrGetSession(ctx, input.RootSession)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *LaunchCoordinator) UpdateIssueWithSession(
	ctx context.Context,
	input UpdateIssueSessionInput,
) (*Issue, error) {
	if input.IssueID == uuid.Nil || input.RootSession.ID != input.IssueID {
		return nil, fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	if input.RootSession.ParentSessionID != nil || input.RootSession.RootSessionID != nil {
		return nil, fmt.Errorf("%w: TODO root session must be canonical", ErrInvalidInput)
	}
	var issue *Issue
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		repositoryTx, err := NewRepository(captainTx.Gorm())
		if err != nil {
			return err
		}
		issue, err = repositoryTx.UpdateIssue(
			ctx,
			input.IssueID,
			input.ExpectedIssueVersion,
			input.Patch,
		)
		if err != nil {
			return err
		}
		root, err := captainTx.CreateOrGetSession(ctx, input.RootSession)
		if err != nil {
			return err
		}
		if input.Session.ParentSessionID != nil && *input.Session.ParentSessionID != root.ID {
			return fmt.Errorf("%w: operation session parent must be the TODO root", ErrInvalidInput)
		}
		input.Session.ParentSessionID = &root.ID
		_, err = captainTx.CreateOrGetSession(ctx, input.Session)
		return err
	})
	if err != nil {
		return nil, err
	}
	return issue, nil
}
