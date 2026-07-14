package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

func (p *Provider) todoFromIssue(
	ctx context.Context,
	issue *native.Issue,
	sourceDir string,
	includeEvents bool,
) (*types.TODO, error) {
	if issue == nil {
		return nil, fmt.Errorf("native TODO issue is nil")
	}
	workspace := p.workspace
	if workspace == nil || workspace.ID != issue.WorkspaceID {
		var err error
		workspace, err = p.repository.GetWorkspace(ctx, issue.WorkspaceID)
		if err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(sourceDir) == "" {
		sourceDir = workspace.RootPath
	}

	body := issue.Body
	if verification := strings.TrimSpace(issue.Verification); verification != "" {
		body = todos.UpsertVerificationFixture(body, verification)
	}
	status := todoStatus(issue.Status, issue.ExecutionState)
	priority := todoPriority(issue.Priority)
	created := issue.CreatedAt
	defaults := types.TODOFrontmatter{
		Title:    issue.Title,
		Priority: priority,
		Status:   status,
		Created:  &created,
		CWD:      sourceDir,
	}
	todo, err := todos.ParseTODOContent(issue.Title, body, sourceDir, defaults)
	if err != nil {
		return nil, err
	}

	// Database columns remain authoritative even when imported body markdown
	// contains stale YAML frontmatter.
	todo.ID = issue.ID.String()
	todo.ShortID = issue.ID.String()[:8]
	todo.Version = issue.Version
	todo.WorkspaceID = issue.WorkspaceID.String()
	todo.ExecutionState = string(issue.ExecutionState)
	todo.Provider = todos.ProviderDB
	todo.ProviderState = string(issue.Status)
	todo.Workspace = workspaceName(workspace)
	todo.Labels = append([]string(nil), issue.Labels...)
	todo.Title = issue.Title
	todo.Priority = priority
	todo.Status = status
	todo.Created = &created
	lastActivity := issue.UpdatedAt
	todo.LastRun = &lastActivity
	todo.CWD = sourceDir
	if includeEvents {
		events, err := p.repository.ListEvents(ctx, issue.ID)
		if err != nil {
			return nil, err
		}
		todo.ProviderEvents = providerEvents(events)
	}
	if err := p.decorateExecution(ctx, issue, todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func workspaceName(workspace *native.Workspace) string {
	if workspace == nil {
		return ""
	}
	if name := strings.TrimSpace(workspace.DisplayName); name != "" {
		return name
	}
	if repo := strings.TrimSpace(workspace.RepoKey); repo != "" {
		return repo
	}
	if root := strings.TrimSpace(workspace.RootPath); root != "" {
		return filepath.Base(root)
	}
	return workspace.ID.String()
}

func providerEvents(events []native.Event) []types.ProviderEvent {
	result := make([]types.ProviderEvent, 0, len(events))
	for _, event := range events {
		id := event.ID.String()
		result = append(result, types.ProviderEvent{
			ID: id, ShortID: id[:8], Kind: event.Kind, Actor: event.Actor,
			Timestamp: event.CreatedAt, Title: event.Kind, Body: event.Body,
		})
	}
	return result
}

func todoPriority(priority native.Priority) types.Priority {
	switch priority {
	case native.PriorityLow:
		return types.PriorityLow
	case native.PriorityHigh, native.PriorityCritical:
		return types.PriorityHigh
	default:
		return types.PriorityMedium
	}
}

func toNativePriority(priority types.Priority) (native.Priority, error) {
	if priority == "" {
		return native.PriorityMedium, nil
	}
	switch priority {
	case types.PriorityLow:
		return native.PriorityLow, nil
	case types.PriorityMedium:
		return native.PriorityMedium, nil
	case types.PriorityHigh:
		return native.PriorityHigh, nil
	default:
		return "", fmt.Errorf("unsupported TODO priority %q", priority)
	}
}

func todoStatus(status native.IssueStatus, execution native.ExecutionState) types.Status {
	switch status {
	case native.StatusDraft:
		return types.StatusDraft
	case native.StatusVerified:
		return types.StatusVerified
	case native.StatusClosed, native.StatusCancelled:
		return types.StatusCompleted
	}
	switch execution {
	case native.ExecutionPlanning, native.ExecutionRunning, native.ExecutionVerifying:
		return types.StatusInProgress
	case native.ExecutionWaiting:
		return types.StatusAsk
	case native.ExecutionStalled, native.ExecutionFailed:
		return types.StatusFailed
	case native.ExecutionVerificationFailed:
		return types.StatusUnverified
	default:
		return types.StatusPending
	}
}

func durableStatusForCreate(status types.Status) (native.IssueStatus, error) {
	if status == "" {
		status = types.StatusPending
	}
	durable, isDurable, err := toDurableStatus(status)
	if err != nil {
		return "", err
	}
	// Transient statuses are valid input for compatibility, but a newly created
	// issue has no Captain prompt run from which to derive them.
	if !isDurable {
		return native.StatusOpen, nil
	}
	return durable, nil
}

// toDurableStatus intentionally declines to map execution-only states. Those
// are projected from Captain, not written by the legacy Provider adapter.
func toDurableStatus(status types.Status) (native.IssueStatus, bool, error) {
	if !types.IsKnownStatus(status) {
		return "", false, fmt.Errorf("unsupported TODO status %q", status)
	}
	switch status {
	case types.StatusDraft:
		return native.StatusDraft, true, nil
	case types.StatusPending:
		return native.StatusOpen, true, nil
	case types.StatusVerified:
		return native.StatusVerified, true, nil
	case types.StatusCompleted:
		return native.StatusClosed, true, nil
	case types.StatusSkipped:
		// A skipped native run means its explicit precondition/fixture already
		// passed. That is verification evidence, not human closure.
		return native.StatusVerified, true, nil
	case types.StatusInProgress, types.StatusReview, types.StatusAsk,
		types.StatusFailed, types.StatusUnverified:
		return "", false, nil
	default:
		return "", false, nil
	}
}
