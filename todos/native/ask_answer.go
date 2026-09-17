package native

import (
	"context"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// EventAskAnswered is the history event kind recording that the questions an
// active prompt run parked on were answered. Both answer paths record it: the
// dashboard's own answer, and one given from Captain's session page that a read
// reconciles afterwards. The projection reads a waiting run with a Captain answer
// newer than its latest ask as running.
const EventAskAnswered = "ask_answered"

// AskAnswerSourceCaptain is the source and actor of an answer given from
// Captain's session page and reconciled by a gavel read.
const AskAnswerSourceCaptain = "captain"

// AskAnswerInput is one answer to the questions an issue's active prompt run
// asked.
type AskAnswerInput struct {
	IssueID     uuid.UUID
	PromptRunID uuid.UUID
	// ExpectedRunVersion is the version of the waiting run the answer answers:
	// the one a reconciler read, or the one a resume is admitted against. A run
	// that is no longer waiting at it is not recorded against.
	ExpectedRunVersion int64
	Source             string
	SourceID           string
	Actor              string
	Body               string
}

// errAskAnswerStale marks a run that is no longer the one the answer was read
// against. It never leaves this file: callers see recorded=false.
var errAskAnswerStale = errors.New("ask answer is stale")

// RecordAskAnswer appends an ask_answered event under the issue lock. It reports
// false, without error, when the answer was already recorded (the event's source
// identity exists) or the run it answers moved on in the meantime: both mean
// another reader got there first.
func (c *LaunchCoordinator) RecordAskAnswer(ctx context.Context, input AskAnswerInput) (bool, error) {
	if input.IssueID == uuid.Nil || input.PromptRunID == uuid.Nil {
		return false, fmt.Errorf("%w: an ask answer needs an issue and a prompt run", ErrInvalidInput)
	}
	if input.ExpectedRunVersion <= 0 {
		return false, fmt.Errorf("%w: an ask answer needs the version of the run it answers", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Source) == "" || strings.TrimSpace(input.Body) == "" {
		return false, fmt.Errorf("%w: an ask answer needs a source and a body", ErrInvalidInput)
	}
	err := c.captain.Transaction(ctx, func(captainTx *captaindb.DB) error {
		tx := captainTx.Gorm()
		issue, err := lockExecutionIssue(tx, input.IssueID)
		if err != nil {
			return err
		}
		if issue.ActivePromptRunID == nil || *issue.ActivePromptRunID != input.PromptRunID {
			return errAskAnswerStale
		}
		run, err := captainTx.GetPromptRun(ctx, input.PromptRunID)
		if err != nil {
			return err
		}
		if run.State != captaindb.PromptRunStateWaiting || run.Version != input.ExpectedRunVersion {
			return errAskAnswerStale
		}
		if input.SourceID != "" {
			// Every answer to this issue is recorded under the lock just taken, so
			// this read settles a race the unique source index would otherwise
			// settle by failing — and logging — the losing insert.
			var exists bool
			if err := tx.Raw(`SELECT EXISTS (SELECT 1 FROM todo_issue_events WHERE source = ? AND source_id = ?)`,
				normalizeToken(input.Source), strings.TrimSpace(input.SourceID)).Scan(&exists).Error; err != nil {
				return err
			}
			if exists {
				return ErrEventConflict
			}
		}
		_, err = recordMutation(tx, issue.lockedIssue(), EventInput{
			Kind: EventAskAnswered, Actor: input.Actor, Body: input.Body,
			Source: input.Source, SourceID: input.SourceID,
			Payload: map[string]any{
				"promptRunId": input.PromptRunID.String(), "runVersion": input.ExpectedRunVersion, "source": input.Source,
			},
		})
		return err
	})
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, errAskAnswerStale), errors.Is(err, ErrEventConflict):
		return false, nil
	default:
		return false, fmt.Errorf("record ask answer on issue %s: %w", input.IssueID, err)
	}
}
