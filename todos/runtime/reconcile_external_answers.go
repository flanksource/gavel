package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// externalAnswerSource is who answers a parked run outside gavel: Captain's
// session page resumes the run's own provider session with the answer as the
// next user turn, and records the turn on that transcript.
const externalAnswerSource = native.AskAnswerSourceCaptain

// reconcileScope narrows reconciliation to one workspace (List, CountByStatus)
// or one issue (Get). WorkDir is the directory the lifecycle host that settles
// a finished turn loads its configuration from.
type reconcileScope struct {
	WorkspaceID uuid.UUID
	IssueID     uuid.UUID
	WorkDir     string
}

// askCandidate is an issue whose active run is parked on an ask outcome gavel
// recorded for that run, with a provider transcript an answer could land on.
type askCandidate struct {
	IssueID            uuid.UUID
	WorkspaceID        uuid.UUID
	PromptRunID        uuid.UUID
	RunVersion         int64
	ExecutionSessionID uuid.UUID
	StepKind           string
	ParkedAt           time.Time
}

// reconcileExternalAnswers notices answers given outside gavel and records them
// on the issues they answer. Gavel is not involved when Captain resumes a parked
// run's session, so its reads are where the TODO learns of it: a live answer
// records one ask_answered event (the projection then reads the run as running),
// and a finished Captain turn settles the run through the lifecycle's outcomes.
//
// A single issue's read fails when its reconciliation does. A workspace read
// does not fail every TODO for one that could not be reconciled: that failure
// is logged against the TODO, whose own read then reports it.
//
// It reports whether any parked run was examined, so a caller holding an issue
// it read before re-reads it. A workspace with nothing parked costs the one
// candidate query.
func (p *Provider) reconcileExternalAnswers(ctx context.Context, scope reconcileScope) (bool, error) {
	candidates, err := p.askCandidates(ctx, scope)
	if err != nil {
		return false, fmt.Errorf("list TODOs awaiting an answer: %w", err)
	}
	reconciled := false
	for _, candidate := range candidates {
		if p.isPrepared(candidate.IssueID, candidate.PromptRunID) {
			// This provider dispatched or resumed the run and still drives it; an
			// answer it sent must not be read back as someone else's.
			continue
		}
		reconciled = true
		err := p.reconcileAskCandidate(ctx, candidate, scope.WorkDir)
		if err == nil {
			continue
		}
		err = fmt.Errorf("reconcile external answer for TODO %s: %w", candidate.IssueID, err)
		if scope.IssueID != uuid.Nil {
			return reconciled, err
		}
		logger.Errorf("%v", err)
	}
	return reconciled, nil
}

// reconcileIssue reconciles the one issue a Get resolved, and re-reads it when
// the issue was parked on an ask.
func (p *Provider) reconcileIssue(ctx context.Context, issue *native.Issue, workDir string) (*native.Issue, error) {
	reconciled, err := p.reconcileExternalAnswers(ctx, reconcileScope{IssueID: issue.ID, WorkDir: workDir})
	if err != nil || !reconciled {
		return issue, err
	}
	return p.repository.GetIssue(ctx, issue.ID)
}

func (p *Provider) askCandidates(ctx context.Context, scope reconcileScope) ([]askCandidate, error) {
	filter, arg := "issue.workspace_id = ?", scope.WorkspaceID
	if scope.IssueID != uuid.Nil {
		filter, arg = "issue.id = ?", scope.IssueID
	}
	var candidates []askCandidate
	err := p.db.WithContext(ctx).Raw(`
		SELECT issue.id AS issue_id, issue.workspace_id, run.id AS prompt_run_id, run.version AS run_version,
		       run.execution_session_id, link.step_kind, parked.created_at AS parked_at
		FROM todo_issues issue
		JOIN todo_issue_prompt_runs link
		  ON link.issue_id = issue.id AND link.prompt_run_id = issue.active_prompt_run_id
		JOIN captain_prompt_runs run ON run.id = link.prompt_run_id
		JOIN LATERAL (
		  SELECT event.created_at FROM todo_issue_events event
		  WHERE event.issue_id = issue.id AND event.kind = ?
		    AND event.payload->>'status' = 'ask'
		    AND event.payload->>'promptRunId' = run.id::text
		  ORDER BY event.sequence DESC LIMIT 1
		) parked ON true
		WHERE `+filter+`
		  AND run.state = 'waiting' AND run.result_json->>'endStatus' = 'ask'
		  AND run.execution_session_id IS NOT NULL`,
		lifecycle.EventLifecycleOutcome, arg,
	).Scan(&candidates).Error
	return candidates, err
}

// reconcileAskCandidate records what Captain did with one parked run.
func (p *Provider) reconcileAskCandidate(ctx context.Context, candidate askCandidate, workDir string) error {
	answer, err := p.detectExternalAnswer(ctx, candidate)
	if err != nil || answer == nil {
		return err
	}
	if !answer.AnswerRecorded {
		if _, err := p.coordinator.RecordAskAnswer(ctx, native.AskAnswerInput{
			IssueID: candidate.IssueID, PromptRunID: candidate.PromptRunID, ExpectedRunVersion: candidate.RunVersion,
			Source: externalAnswerSource, Actor: externalAnswerSource,
			SourceID: fmt.Sprintf("ask-answer:%s:%s", candidate.PromptRunID, answer.MessageID),
			Body:     "**Answer (Captain):** " + strings.TrimSpace(answer.MessageText),
		}); err != nil {
			return err
		}
	}
	if answer.turnFinished() {
		if err := p.settleExternalTurn(ctx, candidate, *answer, workDir); err != nil {
			return err
		}
	}
	p.dropPhaseRuns(candidate.WorkspaceID)
	return nil
}

// recordResumedAnswer records gavel's own answer to a parked run as the resumed
// turn is admitted: before it is dispatched, so the turn's next ask can never be
// ordered before it, and bound to the version of the run it resumes, so a run
// that moved meanwhile refuses the resume. The resumed message is what tells a
// Captain answer apart from gavel's own on the shared transcript. It returns the
// issue as the event left it.
func (p *Provider) recordResumedAnswer(ctx context.Context, issue *native.Issue, run *captaindb.PromptRun, message string) (*native.Issue, error) {
	message = strings.TrimSpace(message)
	if message == "" || run.State != captaindb.PromptRunStateWaiting || run.ResultJSON["endStatus"] != "ask" {
		return issue, nil
	}
	recorded, err := p.coordinator.RecordAskAnswer(ctx, native.AskAnswerInput{
		IssueID: issue.ID, PromptRunID: run.ID, ExpectedRunVersion: run.Version,
		Source: mutationActor, Actor: mutationActor, Body: "**Answer:** " + message,
	})
	if err != nil {
		return nil, err
	}
	if !recorded {
		return nil, fmt.Errorf("%w: prompt run %s of issue %s changed while its answer was being admitted",
			native.ErrVersionConflict, run.ID, issue.ID)
	}
	return p.repository.GetIssue(ctx, issue.ID)
}
