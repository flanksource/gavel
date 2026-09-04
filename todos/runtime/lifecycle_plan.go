package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var (
	errPlanContentMissing = errors.New("plan run produced no durable markdown content")
	resolveSessionPlan    = todos.ResolveSessionPlan
)

// PlanMarkdown returns Captain-owned immutable plan content. Runtime callers
// never read a portable-file plan pointer or silently execute an unapproved
// revision.
func (p *Provider) PlanMarkdown(ctx context.Context, todo *types.TODO, mode types.RunMode) (string, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return "", err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return "", err
	}
	if issue.SelectedPlanID == nil {
		return "", nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return "", err
	}
	switch mode {
	case "", types.ModeRun:
		if plan.ApprovalState != captaindb.PlanApprovalApproved || plan.ApprovedRevision == nil {
			return "", fmt.Errorf("selected Captain plan %s is %s; approve an immutable revision before implementation", plan.ID, plan.ApprovalState)
		}
		return plan.ApprovedRevision.PlanMarkdown, nil
	case types.ModePlan:
		if plan.LatestRevision == nil {
			return "", nil
		}
		return plan.LatestRevision.PlanMarkdown, nil
	case types.ModeVerify:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported TODO run mode %q", mode)
	}
}

// PlanState is the selected plan as the lifecycle sees it. Content is the
// approved revision once one is approved — what an implementation run follows —
// and otherwise the latest revision, what a planning run revises.
func (p *Provider) PlanState(ctx context.Context, todo *types.TODO) (todos.PlanState, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return todos.PlanState{}, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return todos.PlanState{}, err
	}
	if issue.SelectedPlanID == nil {
		return todos.PlanState{}, nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return todos.PlanState{}, err
	}
	state := todos.PlanState{Exists: true, Path: plan.Path}
	if plan.LatestRevision != nil {
		state.Revision = plan.LatestRevision.Revision
		state.Content = plan.LatestRevision.PlanMarkdown
	}
	if plan.ApprovalState == captaindb.PlanApprovalApproved && plan.ApprovedRevision != nil {
		state.Approved = true
		state.Revision = plan.ApprovedRevision.Revision
		state.Content = plan.ApprovedRevision.PlanMarkdown
	}
	return state, nil
}

func successfulPlanAttempt(result *todos.ExecutionResult) bool {
	if result == nil || !result.Success || result.EndStatus == types.EndFailed || result.EndStatus == types.EndAsk {
		return false
	}
	return strings.TrimSpace(result.ErrorMessage) == ""
}

// persistPlanAttempt reconciles a finished plan step with Captain's plan store:
// a new or updated plan becomes a revision of the issue's selected plan, an
// unchanged plan is re-read from its native file in case an earlier attempt
// persisted only a summary, and a failed plan run still records whatever
// content it produced so a person can review it.
func (p *Provider) persistPlanAttempt(ctx context.Context, todo *types.TODO, active *activeRun, result *todos.ExecutionResult) (*activeRun, error) {
	switch {
	case successfulPlanAttempt(result) && result.Plan != nil:
		switch result.Plan.Status {
		case types.PlanNew, types.PlanUpdated:
			if err := p.persistPlanResult(ctx, todo, active, result); err != nil {
				_ = p.failPromptRun(ctx, active, err.Error())
				p.clearPrepared(active.issue.ID, active.run.ID)
				return active, err
			}
			return p.loadActiveRun(ctx, todo)
		case types.PlanUnchanged:
			if active.issue.SelectedPlanID == nil {
				return active, fmt.Errorf("plan run reported unchanged but issue %s has no selected Captain plan", active.issue.ID)
			}
			// Codex plan mode keeps the full Markdown in its native plan file and
			// may return only a short summary in plan.content. An "unchanged"
			// result therefore still needs to reconcile Captain when an earlier
			// attempt persisted that summary instead of the referenced file.
			path := strings.TrimSpace(result.Plan.Path)
			if path == "" {
				return active, nil
			}
			fileMarkdown, _, exists, readErr := todos.ReadPlanFile(path)
			if readErr != nil {
				return active, readErr
			}
			if !exists || strings.TrimSpace(fileMarkdown) == "" {
				return active, nil
			}
			plan, getErr := p.captain.GetPlan(ctx, *active.issue.SelectedPlanID)
			if getErr != nil {
				return active, getErr
			}
			if plan.LatestRevision != nil && normalizePlanResultMarkdown(plan.LatestRevision.PlanMarkdown) == normalizePlanResultMarkdown(fileMarkdown) {
				return active, nil
			}
			if err := p.persistPlanResult(ctx, todo, active, result); err != nil {
				_ = p.failPromptRun(ctx, active, err.Error())
				p.clearPrepared(active.issue.ID, active.run.ID)
				return active, err
			}
			return p.loadActiveRun(ctx, todo)
		}
		return active, nil
	default:
		status := types.PlanUpdated
		if active.issue.SelectedPlanID == nil {
			status = types.PlanNew
		}
		recovered := &todos.ExecutionResult{
			ExecutorName: active.run.Runtime.Driver,
			Plan:         &types.PlanResult{Status: status},
		}
		if result != nil && strings.TrimSpace(result.ExecutorName) != "" {
			recovered.ExecutorName = result.ExecutorName
		}
		if err := p.persistPlanResult(ctx, todo, active, recovered); err != nil {
			if !errors.Is(err, errPlanContentMissing) {
				return active, err
			}
			return active, nil
		}
		return p.loadActiveRun(ctx, todo)
	}
}

func (p *Provider) persistPlanResult(ctx context.Context, todo *types.TODO, active *activeRun, result *todos.ExecutionResult) error {
	markdown, path, err := planResultContent(result, planResolutionSessionID(todo, active.run))
	if err != nil {
		return err
	}
	currentIssue, err := p.repository.GetIssue(ctx, active.issue.ID)
	if err != nil {
		return err
	}
	planInput := captaindb.CreatePlanInput{
		SourceSessionID:   active.run.SessionID,
		SourcePromptRunID: &active.run.ID,
		Title:             currentIssue.Title,
		Path:              path,
		Variant:           "primary",
		SpecProfile:       "gavel.todo.plan",
	}
	ordinal := 0
	if currentIssue.SelectedPlanID != nil {
		existing, err := p.captain.GetPlan(ctx, *currentIssue.SelectedPlanID)
		if err != nil {
			return err
		}
		planInput = captaindb.CreatePlanInput{
			ID: existing.ID, SourceSessionID: existing.SourceSessionID,
			Title: existing.Title, Path: path, SpecProfile: existing.SpecProfile,
		}
		links, err := p.repository.ListPlans(ctx, currentIssue.ID)
		if err != nil {
			return err
		}
		for _, link := range links {
			if link.PlanID == existing.ID {
				ordinal = link.Ordinal
				break
			}
		}
	}
	persisted, err := p.coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
		Plan: planInput,
		Revision: captaindb.AppendPlanRevisionInput{
			PlanMarkdown: markdown,
			CreatedBy:    strings.TrimSpace(result.ExecutorName),
		},
		Attachment: native.PlanSelectionAttachment{
			IssueID: currentIssue.ID, Ordinal: ordinal,
			ExpectedIssueVersion: currentIssue.Version, Actor: mutationActor,
		},
	})
	if err != nil {
		return fmt.Errorf("persist Captain plan revision: %w", err)
	}
	return p.replaceTODO(ctx, todo, persisted.Issue, todo.CWD)
}

func planResultContent(result *todos.ExecutionResult, sessionID string) (content, path string, err error) {
	if result != nil && result.Plan != nil {
		path = strings.TrimSpace(result.Plan.Path)
	}
	// The native plan file is authoritative when the agent supplies one.
	// Codex commonly puts the detailed plan there while plan.content is only a
	// short completion summary, so preferring inline content truncates the
	// immutable Captain revision and the dashboard's Plan tab.
	if path != "" {
		read, _, exists, readErr := todos.ReadPlanFile(path)
		if readErr != nil {
			return "", path, readErr
		}
		if exists && strings.TrimSpace(read) != "" {
			return strings.TrimSpace(read), path, nil
		}
	}
	if result != nil && result.Plan != nil {
		if content = strings.TrimSpace(result.Plan.Content); content != "" {
			return content, path, nil
		}
	}
	resolvedPath, resolved := resolveSessionPlan(sessionID)
	if strings.TrimSpace(resolved) != "" {
		if path == "" {
			path = resolvedPath
		}
		return strings.TrimSpace(resolved), path, nil
	}
	return "", path, fmt.Errorf("%w for session %q", errPlanContentMissing, strings.TrimSpace(sessionID))
}

func planResolutionSessionID(todo *types.TODO, run *captaindb.PromptRun) string {
	if run != nil && run.ExecutionSessionID != nil {
		return run.ExecutionSessionID.String()
	}
	if todo != nil && todo.LLM != nil && strings.TrimSpace(todo.LLM.SessionId) != "" {
		return strings.TrimSpace(todo.LLM.SessionId)
	}
	if run != nil {
		return run.SessionID.String()
	}
	return ""
}

func normalizePlanResultMarkdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	markdown = strings.ReplaceAll(markdown, "\r", "\n")
	return strings.TrimSpace(markdown)
}

func (p *Provider) attachInputPlan(ctx context.Context, issue *native.Issue, mode types.RunMode, input *captaindb.CreatePromptRunInput) error {
	if issue.SelectedPlanID == nil || mode == types.ModeVerify {
		return nil
	}
	plan, err := p.captain.GetPlan(ctx, *issue.SelectedPlanID)
	if err != nil {
		return err
	}
	var revision *captaindb.PlanRevision
	switch mode {
	case types.ModeRun:
		if plan.ApprovalState != captaindb.PlanApprovalApproved || plan.ApprovedRevision == nil {
			return fmt.Errorf("selected Captain plan %s is %s; approve an immutable revision before implementation", plan.ID, plan.ApprovalState)
		}
		revision = plan.ApprovedRevision
	case types.ModePlan:
		revision = plan.LatestRevision
	}
	if revision != nil {
		planID := plan.ID
		revisionID := revision.ID
		input.InputPlanID = &planID
		input.InputPlanRevisionID = &revisionID
	}
	return nil
}
