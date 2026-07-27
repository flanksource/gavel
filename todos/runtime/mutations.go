package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// Delete preserves the issue and its history by transitioning it to cancelled.
func (p *Provider) Delete(ctx context.Context, todo *types.TODO) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	status := native.StatusCancelled
	issue, err := p.repository.UpdateIssue(ctx, id, version, native.IssuePatch{
		Status: &status,
		Actor:  mutationActor,
	})
	if err != nil {
		return err
	}
	return p.replaceTODO(ctx, todo, issue, p.workDir)
}

func (p *Provider) Edit(ctx context.Context, todo *types.TODO, edit todos.EditRequest) error {
	if edit.IsEmpty() {
		return fmt.Errorf("nothing to edit: title, body, or labels are required")
	}
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	patch := native.IssuePatch{Actor: mutationActor}
	hasNativeChange := false
	if edit.Title != nil {
		patch.Title = edit.Title
		hasNativeChange = true
	}
	if edit.Body != nil {
		patch.Body = edit.Body
		verification := todos.ExtractVerificationFixture(*edit.Body)
		patch.Verification = &verification
		hasNativeChange = true
	}
	if len(edit.Labels) > 0 {
		labels := append([]string(nil), edit.Labels...)
		patch.Labels = &labels
		hasNativeChange = true
	}
	if !hasNativeChange {
		return fmt.Errorf("native TODO storage supports title, body, and label edits; path and arbitrary metadata require explicit import/export")
	}
	issue, err := p.repository.UpdateIssue(ctx, id, version, patch)
	if err != nil {
		return err
	}
	return p.replaceTODO(ctx, todo, issue, p.workDir)
}

func (p *Provider) Comment(ctx context.Context, todo *types.TODO, body string) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	if _, err := p.repository.AddComment(ctx, id, version, mutationActor, body); err != nil {
		return err
	}
	return p.reloadTODO(ctx, todo, p.workDir)
}

func (p *Provider) UpdateState(ctx context.Context, todo *types.TODO, update todos.StateUpdate) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	patch := native.IssuePatch{Actor: mutationActor}
	hasDurableChange := false
	if update.Priority != nil {
		priority, err := toNativePriority(*update.Priority)
		if err != nil {
			return err
		}
		patch.Priority = &priority
		hasDurableChange = true
	}
	if update.Status != nil {
		if status, durable, err := toDurableStatus(*update.Status); err != nil {
			return err
		} else if durable {
			patch.Status = &status
			hasDurableChange = true
		}
	}

	if hasDurableChange {
		issue, err := p.repository.UpdateIssue(ctx, id, version, patch)
		if err != nil {
			return err
		}
		if err := p.replaceTODO(ctx, todo, issue, p.workDir); err != nil {
			return err
		}
	}
	// Legacy run bookkeeping remains available to the in-process executor, but
	// is not duplicated into durable issue state. Captain links/projection own
	// planning, running, waiting, failed, and verification-failed transitions.
	applyCompatibilityState(todo, update)
	if update.Status != nil && *update.Status == types.StatusFailed {
		reason := "agent run failed before producing a durable result"
		if err := p.failPreparedRun(ctx, todo, reason); err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error {
	if result == nil {
		return nil
	}
	body, err := clicky.Format(result, clicky.FormatOptions{Markdown: true})
	if err != nil {
		return fmt.Errorf("format latest native TODO failure: %w", err)
	}
	body = strings.TrimSpace(body)
	if !strings.HasPrefix(body, "## Latest Failure") {
		body = "## Latest Failure\n\n" + body
	}
	return p.appendEvent(ctx, todo, native.EventInput{
		Kind:    "latest_failure",
		Actor:   mutationActor,
		Body:    body,
		Payload: result,
	})
}

func (p *Provider) SaveAttempt(ctx context.Context, todo *types.TODO, result *todos.ExecutionResult) error {
	if result == nil {
		return nil
	}
	if _, err := p.finishAttempt(ctx, todo, result); err != nil {
		return err
	}
	return p.appendEvent(ctx, todo, native.EventInput{
		Kind:  "attempt",
		Actor: mutationActor,
		Body:  renderAttempt(todo, result),
		Payload: map[string]any{
			"attempt":        todo.Attempts,
			"status":         attemptStatus(result),
			"driver":         result.Runtime.Driver,
			"agent":          result.Runtime.Agent,
			"provider":       result.Runtime.Provider,
			"backend":        result.Runtime.Backend,
			"model":          result.Runtime.ResolvedModel,
			"effort":         result.Runtime.Effort,
			"durationMillis": result.Duration.Milliseconds(),
			"costUsd":        result.CostUSD,
			"tokens":         result.TokensUsed,
			"turns":          result.NumTurns,
			"commit":         result.CommitSHA,
			"error":          result.ErrorMessage,
		},
	})
}

// MoveTo moves the native issue itself. It preserves UUID, aliases, links, and
// event history instead of emulating transfer as create plus delete.
func (p *Provider) MoveTo(ctx context.Context, todo *types.TODO, target todos.Provider) (*types.TODO, error) {
	targetProvider, ok := target.(*Provider)
	if !ok || targetProvider == nil {
		return nil, fmt.Errorf("native TODOs can only move to another PostgreSQL workspace")
	}
	if err := sameDatabase(p, targetProvider); err != nil {
		return nil, err
	}
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return nil, err
	}
	if p.workspace.ID == targetProvider.workspace.ID {
		return targetProvider.Get(ctx, id.String())
	}
	issue, err := p.repository.MoveIssueWorkspace(
		ctx, id, targetProvider.workspace.ID, version, mutationActor,
	)
	if err != nil {
		return nil, err
	}
	moved, err := targetProvider.todoFromIssue(ctx, issue, targetProvider.workDir, true)
	if err != nil {
		return nil, err
	}
	if todo != nil {
		*todo = *moved
	}
	return moved, nil
}

func sameDatabase(left, right *Provider) error {
	leftSQL, leftErr := left.db.DB()
	rightSQL, rightErr := right.db.DB()
	if leftErr != nil || rightErr != nil || leftSQL != rightSQL {
		return native.ErrDatabasePoolMismatch
	}
	return nil
}

func (p *Provider) appendEvent(ctx context.Context, todo *types.TODO, input native.EventInput) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	if _, err := p.repository.AppendEvent(ctx, id, version, input); err != nil {
		return err
	}
	return p.reloadTODO(ctx, todo, p.workDir)
}

func (p *Provider) mutationIdentity(todo *types.TODO) (uuid.UUID, int64, error) {
	if todo == nil {
		return uuid.Nil, 0, fmt.Errorf("native TODO is nil")
	}
	id, err := uuid.Parse(strings.TrimSpace(todo.ID))
	if err != nil {
		return uuid.Nil, 0, fmt.Errorf("invalid native TODO ID %q: %w", todo.ID, err)
	}
	if todo.WorkspaceID != "" {
		workspaceID, err := uuid.Parse(todo.WorkspaceID)
		if err != nil {
			return uuid.Nil, 0, fmt.Errorf("invalid native TODO workspace ID %q: %w", todo.WorkspaceID, err)
		}
		if workspaceID != p.workspace.ID {
			return uuid.Nil, 0, fmt.Errorf("%w: TODO %s belongs to workspace %s, provider owns %s", native.ErrCrossWorkspace, id, workspaceID, p.workspace.ID)
		}
	}
	if todo.Version <= 0 {
		return uuid.Nil, 0, fmt.Errorf("native TODO %s has no optimistic-lock version; reload it before updating", id)
	}
	return id, todo.Version, nil
}

func (p *Provider) reloadTODO(ctx context.Context, todo *types.TODO, sourceDir string) error {
	id, err := uuid.Parse(todo.ID)
	if err != nil {
		return err
	}
	issue, err := p.repository.GetIssue(ctx, id)
	if err != nil {
		return err
	}
	return p.replaceTODO(ctx, todo, issue, sourceDir)
}

func (p *Provider) replaceTODO(ctx context.Context, target *types.TODO, issue *native.Issue, sourceDir string) error {
	mapped, err := p.todoFromIssue(ctx, issue, sourceDir, true)
	if err != nil {
		return err
	}
	*target = *mapped
	return nil
}

func applyCompatibilityState(todo *types.TODO, update todos.StateUpdate) {
	if todo == nil {
		return
	}
	if update.Status != nil {
		todo.Status = *update.Status
	}
	if update.Priority != nil {
		todo.Priority = *update.Priority
	}
	if update.Attempts != nil {
		todo.Attempts = *update.Attempts
	}
	if update.LastRun != nil {
		todo.LastRun = update.LastRun
	}
	if update.SessionID != nil {
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		todo.LLM.SessionId = *update.SessionID
	}
	if update.PlanPath != nil {
		todo.PlanPath = *update.PlanPath
	}
	if update.PlanStatus != nil {
		todo.PlanStatus = *update.PlanStatus
	}
	if update.RunMode != nil {
		todo.RunMode = *update.RunMode
	}
	if update.LastRunSummary != nil {
		todo.LastRunSummary = *update.LastRunSummary
	}
	if update.Questions != nil {
		todo.Questions = append([]types.AgentQuestion(nil), (*update.Questions)...)
	}
}

func renderAttempt(todo *types.TODO, result *todos.ExecutionResult) string {
	var body strings.Builder
	fmt.Fprintf(&body, "## Attempt %d\n\n", todo.Attempts)
	fmt.Fprintf(&body, "- **Status:** %s\n", attemptStatus(result))
	fmt.Fprintf(&body, "- **Date:** %s\n", time.Now().Format("2006-01-02 15:04"))
	for _, field := range []struct{ label, value string }{
		{"Driver", result.Runtime.Driver},
		{"Agent", result.Runtime.Agent},
		{"Provider", result.Runtime.Provider},
		{"Backend", result.Runtime.Backend},
		{"Model", result.Runtime.ResolvedModel},
		{"Effort", result.Runtime.Effort},
	} {
		if strings.TrimSpace(field.value) != "" {
			fmt.Fprintf(&body, "- **%s:** %s\n", field.label, field.value)
		}
	}
	if result.Duration > 0 {
		fmt.Fprintf(&body, "- **Duration:** %s\n", result.Duration.Round(time.Second))
	}
	if result.CostUSD > 0 {
		fmt.Fprintf(&body, "- **Cost:** $%.4f\n", result.CostUSD)
	}
	if result.TokensUsed > 0 {
		fmt.Fprintf(&body, "- **Tokens:** %d\n", result.TokensUsed)
	}
	if result.CommitSHA != "" {
		fmt.Fprintf(&body, "- **Commit:** `%s`\n", result.CommitSHA)
	}
	if result.ErrorMessage != "" {
		fmt.Fprintf(&body, "- **Error:**\n\n```text\n%s\n```\n", strings.TrimSpace(result.ErrorMessage))
	}
	if todo.LLM != nil && strings.TrimSpace(todo.LLM.SessionId) != "" {
		fmt.Fprintf(&body, "- **Session:** `%s`\n", todo.LLM.SessionId)
	}
	return body.String()
}

func attemptStatus(result *todos.ExecutionResult) string {
	if result.Success {
		return "completed"
	}
	if result.Skipped {
		return "skipped"
	}
	return "failed"
}
