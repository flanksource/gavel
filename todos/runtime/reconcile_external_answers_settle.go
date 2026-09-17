package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"gorm.io/gorm"
)

// settleExternalTurn applies a finished Captain turn to the gavel run it
// continued.
//
// Readers race to it, so it runs under a transaction-scoped advisory lock keyed
// on the run and re-reads the run inside it: the loser finds the run no longer
// at the version it read and leaves it. The lock guards no rows, so the host's
// writes — made on their own connections — never wait on it.
//
// A settlement is several commits (the prompt run, the attempt, the status, the
// outcome event), so it must not stop half way because the read that noticed it
// went away: it runs detached from the caller's cancellation.
func (p *Provider) settleExternalTurn(ctx context.Context, candidate askCandidate, answer externalAnswer, workDir string) error {
	ctx = context.WithoutCancel(ctx)
	turn, err := externalTurn(candidate, answer)
	if err != nil {
		return err
	}
	host, err := lifecycle.NewHost(p, workDir, lifecycle.HostCLI)
	if err != nil {
		return err
	}
	return p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`,
			"gavel-external-answer:"+candidate.PromptRunID.String()).Error; err != nil {
			return fmt.Errorf("lock prompt run %s for settlement: %w", candidate.PromptRunID, err)
		}
		run, err := p.captain.GetPromptRun(ctx, candidate.PromptRunID)
		if err != nil {
			return err
		}
		if run.State != captaindb.PromptRunStateWaiting || run.Version != candidate.RunVersion {
			return nil
		}
		issue, err := p.repository.GetIssue(ctx, candidate.IssueID)
		if err != nil {
			return err
		}
		if issue.ActivePromptRunID == nil || *issue.ActivePromptRunID != run.ID {
			return nil
		}
		todo, err := p.todoFromIssue(ctx, issue, workDir, false)
		if err != nil {
			return err
		}
		return host.SettleExternalTurn(todos.WithPromptRun(ctx, run.ID), todo, candidate.StepKind, turn)
	})
}

// externalTurn is the Captain turn as the lifecycle reads it.
//
// Captain resumes the session without gavel's output schema, so a reply in
// plain text is the ordinary case, and it is a completed turn: the lifecycle's
// outcomes then land the todo where any completed turn lands it. A structured
// result is taken, in order, from the Captain run's own result JSON, the last
// StructuredOutput call of the turn, or a reply text that is a JSON object
// naming an endStatus. An envelope found that way that does not validate fails
// the turn, as it fails a run gavel dispatched.
func externalTurn(candidate askCandidate, answer externalAnswer) (lifecycle.ExternalTurn, error) {
	turn := lifecycle.ExternalTurn{
		Source: externalAnswerSource, PromptRunID: candidate.PromptRunID,
		Summary: firstNonBlank(answer.TurnText, answer.FinalText), Error: answer.TurnError,
	}
	switch captaindb.PromptRunState(answer.TurnState) {
	case captaindb.PromptRunStateSucceeded:
		turn.State = lifecycle.RunSucceeded
	case captaindb.PromptRunStateFailed:
		turn.State = lifecycle.RunFailed
	case captaindb.PromptRunStateCancelled:
		turn.State = lifecycle.RunCancelled
	default:
		return lifecycle.ExternalTurn{}, fmt.Errorf("the Captain turn on session %s ended in unexpected state %q", candidate.ExecutionSessionID, answer.TurnState)
	}
	for _, source := range []struct {
		name string
		data string
	}{
		{"result JSON", string(answer.TurnResult)},
		{"StructuredOutput call", string(answer.Envelope)},
		{"reply", answer.FinalText},
		{"result text", answer.TurnText},
	} {
		envelope, err := envelopeIn(source.data)
		if err != nil {
			return lifecycle.ExternalTurn{}, fmt.Errorf("decode the Captain turn's %s: %w", source.name, err)
		}
		if envelope != nil {
			turn.Envelope = envelope
			break
		}
	}
	return turn, nil
}

// envelopeIn returns the structured result a turn's text carries: a JSON object
// with an endStatus. Text that is not JSON, or JSON that is not such an object,
// carries none — that is the reply of a turn run without gavel's schema.
func envelopeIn(text string) (map[string]any, error) {
	text = strings.TrimSpace(text)
	if text == "" || text == "null" || !json.Valid([]byte(text)) {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal([]byte(text), &value); err != nil {
		return nil, err
	}
	object, ok := value.(map[string]any)
	if !ok || object["endStatus"] == nil {
		return nil, nil
	}
	return object, nil
}
