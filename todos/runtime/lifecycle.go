package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var (
	_ todos.RunLifecycleProvider = (*Provider)(nil)
	_ todos.RunProgressProvider  = (*Provider)(nil)
	_ todos.RunNoticeProvider    = (*Provider)(nil)
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

// finishAttempt projects one executor result into Captain. It returns false
// only for compatibility calls that have no active Captain run.
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

	state, phase, _, _, reason := terminalState(result, active.link.StepKind)
	resultText := ""
	resultJSON := executionResultJSON(result)
	errorText := ""
	if result != nil {
		resultText = strings.TrimSpace(result.Summary)
		errorText = strings.TrimSpace(result.ErrorMessage)
	}
	if state == captaindb.PromptRunStateFailed && errorText == "" {
		errorText = reason
	}
	if !terminalPromptRun(active.run.State) || active.run.State == captaindb.PromptRunStateWaiting {
		updatedRun, updateErr := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
			ID: active.run.ID, ExpectedVersion: active.run.Version,
			State: &state, Phase: &phase, ResultText: &resultText,
			ResultJSON: &resultJSON, Error: &errorText,
		})
		if updateErr != nil {
			return true, fmt.Errorf("finish Captain prompt run: %w", updateErr)
		}
		active.run = updatedRun
	}
	if state != captaindb.PromptRunStateWaiting {
		p.clearPrepared(active.issue.ID, active.run.ID)
	}
	return true, p.reloadTODO(ctx, todo, todo.CWD)
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
	if err := p.failPromptRun(ctx, active, reason); err != nil {
		return err
	}
	p.clearPrepared(issueID, active.run.ID)
	return p.reloadTODO(ctx, todo, todo.CWD)
}

func (p *Provider) failPromptRun(ctx context.Context, active *activeRun, reason string) error {
	state := captaindb.PromptRunStateFailed
	phase := active.run.Phase
	if phase == captaindb.PromptRunPhaseQueued || phase == captaindb.PromptRunPhasePreRun {
		phase = captaindb.PromptRunPhaseGenerate
	}
	if !terminalPromptRun(active.run.State) {
		if _, err := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
			ID: active.run.ID, ExpectedVersion: active.run.Version,
			State: &state, Phase: &phase, Error: &reason,
		}); err != nil {
			return err
		}
	}
	return nil
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

func terminalState(result *todos.ExecutionResult, step native.StepKind) (
	captaindb.PromptRunState,
	captaindb.PromptRunPhase,
	captaindb.SessionLifecycleStatus,
	captaindb.SessionActivityState,
	string,
) {
	phase := captaindb.PromptRunPhaseFinished
	if result == nil {
		return captaindb.PromptRunStateFailed, captaindb.PromptRunPhaseGenerate,
			captaindb.SessionLifecycleFailed, captaindb.SessionActivityIdle, "agent run returned no result"
	}
	if result.Cancelled {
		reason := strings.TrimSpace(result.Summary)
		if reason == "" {
			reason = strings.TrimSpace(result.ErrorMessage)
		}
		if reason == "" {
			reason = todos.ErrExecutionCancelled.Error()
		}
		return captaindb.PromptRunStateCancelled, phase,
			captaindb.SessionLifecycleCancelled, captaindb.SessionActivityIdle, reason
	}
	if result.EndStatus == types.EndAsk {
		return captaindb.PromptRunStateWaiting, captaindb.PromptRunPhaseGenerate,
			captaindb.SessionLifecycleRunning, captaindb.SessionActivityAsk, strings.TrimSpace(result.Summary)
	}
	failedVerification := result.DoD != nil && result.DoD.Ran && !result.DoD.Passed
	if failedVerification || !result.Success || result.EndStatus == types.EndFailed || strings.TrimSpace(result.ErrorMessage) != "" {
		if failedVerification || step == native.StepVerify {
			phase = captaindb.PromptRunPhaseVerify
		} else {
			phase = captaindb.PromptRunPhaseGenerate
		}
		reason := strings.TrimSpace(result.ErrorMessage)
		if reason == "" {
			reason = strings.TrimSpace(result.Summary)
		}
		if reason == "" {
			reason = "agent run failed"
		}
		return captaindb.PromptRunStateFailed, phase,
			captaindb.SessionLifecycleFailed, captaindb.SessionActivityIdle, reason
	}
	return captaindb.PromptRunStateSucceeded, phase,
		captaindb.SessionLifecycleSucceeded, captaindb.SessionActivityIdle, strings.TrimSpace(result.Summary)
}

func executionResultJSON(result *todos.ExecutionResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	out := map[string]any{
		"success": result.Success, "skipped": result.Skipped, "cancelled": result.Cancelled,
		"executor": result.ExecutorName, "tokens": result.TokensUsed,
		"costUsd": result.CostUSD, "turns": result.NumTurns,
		"summary": result.Summary, "endStatus": result.EndStatus,
		"commit": result.CommitSHA, "questions": result.Questions,
	}
	if result.Plan != nil {
		out["plan"] = map[string]any{"status": result.Plan.Status, "path": result.Plan.Path}
	}
	if result.DoD != nil {
		out["definitionOfDone"] = result.DoD
	}
	return out
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
