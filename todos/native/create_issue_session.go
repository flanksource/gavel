package native

import (
	"context"
	"fmt"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/sessiontree"
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

// CreateIssueWithSession has Captain ensure the TODO root session and creates
// the native issue in the same transaction through the session-tree link.
func (c *LaunchCoordinator) CreateIssueWithSession(
	ctx context.Context,
	input CreateIssueSessionInput,
) (*CreatedIssueSession, error) {
	if input.Issue.ID == uuid.Nil || input.RootSession.ID != input.Issue.ID {
		return nil, fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	result := &CreatedIssueSession{}
	sessions, err := sessiontree.Ensure(ctx, c.captain, []captaindb.CreateSessionInput{input.RootSession},
		func(ctx context.Context, tx *captaindb.DB, _ []*captaindb.Session) error {
			repositoryTx, err := NewRepository(tx.Gorm())
			if err != nil {
				return err
			}
			result.Issue, err = repositoryTx.CreateIssue(ctx, input.Issue)
			return err
		})
	if err != nil {
		return nil, err
	}
	result.Root = sessions[0]
	return result, nil
}

// UpdateIssueWithSession has Captain ensure the TODO root and one operation
// session under it, and applies the issue patch in the same transaction
// through the session-tree link.
func (c *LaunchCoordinator) UpdateIssueWithSession(
	ctx context.Context,
	input UpdateIssueSessionInput,
) (*Issue, error) {
	if input.IssueID == uuid.Nil || input.RootSession.ID != input.IssueID {
		return nil, fmt.Errorf("%w: issue and root session must share a non-empty ID", ErrInvalidInput)
	}
	var issue *Issue
	_, err := sessiontree.Ensure(ctx, c.captain, []captaindb.CreateSessionInput{input.RootSession, input.Session},
		func(ctx context.Context, tx *captaindb.DB, _ []*captaindb.Session) error {
			repositoryTx, err := NewRepository(tx.Gorm())
			if err != nil {
				return err
			}
			issue, err = repositoryTx.UpdateIssue(ctx, input.IssueID, input.ExpectedIssueVersion, input.Patch)
			return err
		})
	if err != nil {
		return nil, err
	}
	return issue, nil
}
