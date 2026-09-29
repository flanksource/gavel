package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var (
	_ todos.RunLifecycleProvider = (*Provider)(nil)
	_ todos.PlanContentProvider  = (*Provider)(nil)
	_ todos.PlanStateProvider    = (*Provider)(nil)
	_ todos.EventProvider        = (*Provider)(nil)
)

// AppendEvent records a lifecycle event on the issue's durable history.
func (p *Provider) AppendEvent(ctx context.Context, todo *types.TODO, event todos.Event) error {
	if strings.TrimSpace(event.Kind) == "" {
		return fmt.Errorf("append event: kind is required")
	}
	return p.appendEvent(ctx, todo, native.EventInput{
		Kind: event.Kind, Actor: mutationActor, Body: event.Body, Payload: event.Payload,
	})
}

// finishAttempt applies one executor result to the issue once its run is over.
// It returns false only for compatibility calls that have no active Captain
// run.
//
// Captain filed the run's terminal state as promptrun.Run finished it. A turn
// gavel did not dispatch — an answer given from Captain's session page —
// leaves the run it continued parked, and is settled here through Captain.
func (p *Provider) finishAttempt(ctx context.Context, todo *types.TODO, result *todos.ExecutionResult) (bool, error) {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		if errors.Is(err, native.ErrNotFound) || errors.Is(err, captaindb.ErrPromptRunNotFound) {
			return false, nil
		}
		return false, err
	}

	if active.link.StepKind == native.StepPlan {
		if active, err = p.persistPlanAttempt(ctx, todo, active, result); err != nil {
			return true, err
		}
	}

	outcome := lifecycle.RunOutcome(result, active.link.StepKind == native.StepVerify)
	filed, err := filedAs(active.run, outcome)
	if err != nil {
		return true, err
	}
	if !terminalPromptRun(active.run.State) && !filed {
		if err := p.settleParkedRun(ctx, active.run, outcome); err != nil {
			return true, err
		}
	}
	if outcome.State != captaindb.PromptRunStateWaiting {
		p.clearPrepared(active.issue.ID, active.run.ID)
	}
	return true, p.reloadTODO(ctx, todo, todo.CWD)
}

// filedAs reports whether Captain already filed the run the way outcome reads
// it: a dispatched run's own outcome, which promptrun.Run wrote. A parked run a
// later turn answered still carries the envelope that parked it.
func filedAs(run *captaindb.PromptRun, outcome promptrun.Outcome) (bool, error) {
	if run.State != outcome.State {
		return false, nil
	}
	stored, err := json.Marshal(run.ResultJSON)
	if err != nil {
		return false, fmt.Errorf("encode result of Captain prompt run %s: %w", run.ID, err)
	}
	reported, err := json.Marshal(outcome.JSON)
	if err != nil {
		return false, fmt.Errorf("encode outcome of Captain prompt run %s: %w", run.ID, err)
	}
	return string(stored) == string(reported), nil
}

// settleParkedRun files the outcome of a run no promptrun.Run is left to
// finish through Captain's Settle: succeeded, failed, cancelled, or parked
// again on the turn's new envelope.
func (p *Provider) settleParkedRun(ctx context.Context, run *captaindb.PromptRun, outcome promptrun.Outcome) error {
	switch outcome.State {
	case captaindb.PromptRunStateFailed:
		outcome.Error = firstNonBlank(outcome.Error, outcome.Text)
	case captaindb.PromptRunStateCancelled:
		outcome.Error = firstNonBlank(outcome.Error, outcome.Text, todos.ErrExecutionCancelled.Error())
	}
	if _, err := promptrun.Settle(ctx, p.captain, run.ID, outcome); err != nil {
		return fmt.Errorf("settle Captain prompt run %s as %s: %w", run.ID, outcome.State, err)
	}
	return nil
}

// failRun fails a run that has not reached a terminal state yet. A run Captain
// already finished keeps the state it was filed with.
func (p *Provider) failRun(ctx context.Context, run *captaindb.PromptRun, reason string) error {
	if terminalPromptRun(run.State) {
		return nil
	}
	if _, err := promptrun.Fail(ctx, p.captain, run.ID, reason); err != nil {
		return fmt.Errorf("fail Captain prompt run %s: %w", run.ID, err)
	}
	return nil
}

func (p *Provider) failPreparedRun(ctx context.Context, todo *types.TODO, reason string) error {
	issueID, err := p.todoID(todo)
	if err != nil {
		return err
	}
	// Only fail a run this process prepared. A TODO this process never
	// dispatched — or whose run it already finished — is not this caller's to
	// end, and has no run to look up.
	if !p.hasPrepared(issueID) {
		return nil
	}
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		return err
	}
	if !p.isPrepared(issueID, active.run.ID) {
		return nil
	}
	if err := p.failRun(ctx, active.run, reason); err != nil {
		return err
	}
	p.clearPrepared(issueID, active.run.ID)
	return p.reloadTODO(ctx, todo, todo.CWD)
}

// decorateExecution projects Captain-owned details into the temporary legacy
// TODO view. The database records remain authoritative; these fields exist only
// so current CLI/API/UI consumers can render the cutover without a second
// provider or a filesystem plan pointer.
func (p *Provider) decorateExecution(ctx context.Context, issue *native.Issue, todo *types.TODO) error {
	if p == nil || p.captain == nil || p.repository == nil || issue == nil || todo == nil {
		return nil
	}
	activeStep := native.StepKind("")
	if issue.ActivePromptRunID != nil {
		// Two point reads rather than one batched overview: see executionIndex
		// for why ListPromptRunOverviews is not usable here.
		run, err := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
		if err != nil {
			return err
		}
		session, err := p.captain.GetSession(ctx, run.SessionID)
		if err != nil {
			return err
		}
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		// session.Provider is the executor/driver identity (e.g. "cmux-claude"),
		// never an LLM model — assigning it to LLM.Model poisons the next run's
		// --model resolution. Only the session id is a legitimate LLM field here.
		if strings.TrimSpace(session.ProviderSessionID) != "" {
			todo.LLM.SessionId = session.ProviderSessionID
		}
		lastRun := run.QueuedAt
		if run.StartedAt != nil {
			lastRun = *run.StartedAt
		}
		if run.FinishedAt != nil {
			lastRun = *run.FinishedAt
		}
		if issue.UpdatedAt.After(lastRun) {
			lastRun = issue.UpdatedAt
		}
		todo.LastRun = &lastRun
		todo.LastRunSummary = strings.TrimSpace(run.ResultText)
		if questions, ok := run.ResultJSON["questions"]; ok {
			todo.Questions = decodeQuestions(questions)
		}
		links, err := p.promptRunLinks(ctx, issue.WorkspaceID, issue.ID)
		if err != nil {
			return err
		}
		todo.Attempts = len(links)
		for _, link := range links {
			if link.PromptRunID != run.ID {
				continue
			}
			activeStep = link.StepKind
			switch link.StepKind {
			case native.StepPlan:
				todo.RunMode = types.ModePlan
			case native.StepVerify:
				todo.RunMode = types.ModeVerify
			default:
				todo.RunMode = types.ModeRun
			}
			break
		}
	}

	if issue.SelectedPlanID == nil {
		return nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return err
	}
	if plan.LatestRevision != nil {
		if plan.LatestRevision.Revision <= 1 {
			todo.PlanStatus = types.PlanNew
		} else {
			todo.PlanStatus = types.PlanUpdated
		}
	}
	// Plan paths are source metadata only in native storage.
	todo.PlanPath = ""
	todo.Status = todoStatusWithPlan(issue.Status, issue.ExecutionState, activeStep, plan.ApprovalState)
	return nil
}

func todoStatusWithPlan(
	status native.IssueStatus,
	execution native.ExecutionState,
	step native.StepKind,
	approval captaindb.PlanApprovalState,
) types.Status {
	projected := todoStatus(status, execution)
	planReviewable := execution == native.ExecutionIdle || (execution == native.ExecutionFailed && step == native.StepPlan)
	if (status != native.StatusOpen && status != native.StatusDraft) || !planReviewable {
		return projected
	}
	switch approval {
	case captaindb.PlanApprovalPending, captaindb.PlanApprovalRevisionRequested:
		return types.StatusReview
	case captaindb.PlanApprovalRejected, captaindb.PlanApprovalApproved:
		return types.StatusPending
	default:
		return projected
	}
}

func decodeQuestions(value any) []types.AgentQuestion {
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var questions []types.AgentQuestion
	if json.Unmarshal(data, &questions) != nil {
		return nil
	}
	return questions
}
