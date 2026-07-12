// Package runtime provides the transitional PostgreSQL-only TODO provider.
//
// The package adapts the native repository to the legacy todos.Provider
// interface while CLI, API, and UI callers are cut over. It deliberately does
// not fall back to Grite or .todos files and never creates workspaces as a side
// effect of opening a provider.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const mutationActor = "gavel"

// Provider adapts one existing native workspace to todos.Provider.
type Provider struct {
	workDir     string
	db          *gorm.DB
	repository  *native.Repository
	captain     *captaindb.DB
	coordinator *native.LaunchCoordinator
	workspace   *native.Workspace
	preparedMu  sync.RWMutex
	prepared    map[uuid.UUID]uuid.UUID
}

var _ todos.Provider = (*Provider)(nil)

// Open opens the process-owned PostgreSQL pool and resolves workDir to an
// existing native workspace. It never creates a workspace or selects a legacy
// provider when PostgreSQL is unavailable.
func Open(ctx context.Context, workDir string) (*Provider, error) {
	db, err := database.Require(ctx, "native TODO storage")
	if err != nil {
		return nil, err
	}
	return New(ctx, workDir, db)
}

// OpenGlobal opens the native repository without requiring a caller workspace.
// It is intentionally limited to global-reference resolution; once an issue is
// found its returned TODO carries the owning CWD and callers should open the
// workspace-scoped provider for subsequent list/write operations.
func OpenGlobal(ctx context.Context) (*Provider, error) {
	db, err := database.Require(ctx, "native TODO global reference resolution")
	if err != nil {
		return nil, err
	}
	repository, err := native.NewRepository(db)
	if err != nil {
		return nil, err
	}
	captain, err := captaindb.Use(db)
	if err != nil {
		return nil, err
	}
	coordinator, err := native.NewLaunchCoordinator(captain, repository)
	if err != nil {
		return nil, err
	}
	return &Provider{
		db: db, repository: repository, captain: captain, coordinator: coordinator,
		prepared: map[uuid.UUID]uuid.UUID{},
	}, nil
}

// New constructs a provider over an already migrated GORM pool. It is useful
// to hosts and tests that own the database lifecycle.
func New(ctx context.Context, workDir string, db *gorm.DB) (*Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	normalizedDir, err := normalizeWorkDir(workDir)
	if err != nil {
		return nil, err
	}
	repository, err := native.NewRepository(db)
	if err != nil {
		return nil, err
	}
	workspace, err := resolveExistingWorkspace(ctx, repository, normalizedDir)
	if err != nil {
		return nil, err
	}
	captain, err := captaindb.Use(db)
	if err != nil {
		return nil, err
	}
	coordinator, err := native.NewLaunchCoordinator(captain, repository)
	if err != nil {
		return nil, err
	}
	return &Provider{
		workDir: normalizedDir, db: db, repository: repository,
		captain: captain, coordinator: coordinator, workspace: workspace,
		prepared: map[uuid.UUID]uuid.UUID{},
	}, nil
}

func normalizeWorkDir(workDir string) (string, error) {
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve native TODO working directory: %w", err)
		}
	}
	absolute, err := filepath.Abs(workDir)
	if err != nil {
		return "", fmt.Errorf("resolve native TODO working directory %q: %w", workDir, err)
	}
	return native.NormalizeImportWorkspace(native.ImportWorkspace{RootPath: absolute}).RootPath, nil
}

func resolveExistingWorkspace(ctx context.Context, repository *native.Repository, workDir string) (*native.Workspace, error) {
	workspace, pathErr := repository.GetWorkspaceByPath(ctx, workDir)
	if pathErr == nil {
		return workspace, nil
	}
	if !errors.Is(pathErr, native.ErrNotFound) {
		return nil, fmt.Errorf("resolve native TODO workspace by path %q: %w", workDir, pathErr)
	}

	repoKey := ""
	if repo, err := github.ResolveRepoFromDir(workDir); err == nil && strings.TrimSpace(repo) != "" {
		repoKey = native.NormalizeImportWorkspace(native.ImportWorkspace{
			RepoKey: "github.com/" + strings.TrimSpace(repo),
		}).RepoKey
		workspace, repoErr := repository.GetWorkspaceByRepoKey(ctx, repoKey)
		if repoErr == nil {
			return workspace, nil
		}
		if !errors.Is(repoErr, native.ErrNotFound) {
			return nil, fmt.Errorf("resolve native TODO workspace by repository %q: %w", repoKey, repoErr)
		}
	}

	identity := fmt.Sprintf("path %q", workDir)
	if repoKey != "" {
		identity += fmt.Sprintf(" or repository %q", repoKey)
	}
	return nil, fmt.Errorf(
		"native TODO workspace not found for %s; run `gavel todos import-grite` for an existing Grite workspace or complete native TODO workspace setup before retrying (no Grite or .todos fallback is available)",
		identity,
	)
}

// Repository returns the native Gavel repository used by this provider.
func (p *Provider) Repository() *native.Repository { return p.repository }

// Captain returns the Captain database handle sharing the provider's pool.
func (p *Provider) Captain() *captaindb.DB { return p.captain }

// Coordinator returns the cross-owner launch/plan coordinator. Lifecycle
// orchestration is intentionally implemented by the caller-facing cutover.
func (p *Provider) Coordinator() *native.LaunchCoordinator { return p.coordinator }

// Workspace returns a copy of the provider's resolved native workspace.
func (p *Provider) Workspace() *native.Workspace {
	if p == nil || p.workspace == nil {
		return nil
	}
	workspace := *p.workspace
	return &workspace
}

// SupportsGroupedExecution reports the native invariant that one Captain
// prompt run is attached to exactly one issue.
func (p *Provider) SupportsGroupedExecution() bool { return false }

func (p *Provider) List(ctx context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	issues, err := p.repository.ListIssues(ctx, p.workspace.ID)
	if err != nil {
		return nil, err
	}
	result := make(types.TODOS, 0, len(issues))
	for index := range issues {
		todo, err := p.todoFromIssue(ctx, &issues[index], p.workDir, false)
		if err != nil {
			return nil, fmt.Errorf("decode native TODO %s: %w", issues[index].ID, err)
		}
		if filters.Matches(todo) {
			result = append(result, todo)
		}
	}
	result.Sort()
	return result, nil
}

func (p *Provider) Get(ctx context.Context, ref string) (*types.TODO, error) {
	issue, err := p.repository.GetIssueByRef(ctx, p.workspace.ID, ref)
	if err != nil {
		return nil, err
	}
	return p.todoFromIssue(ctx, issue, p.workDir, true)
}

// GlobalGet resolves a UUID, short UUID, or legacy alias without a caller
// workspace and returns both the issue and its owning workspace CWD. It is used
// for compatibility deep links before a workspace-specific provider exists.
func (p *Provider) GlobalGet(ctx context.Context, ref string) (*types.TODO, string, error) {
	issue, err := p.repository.GetIssueByGlobalRef(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	workspace, err := p.repository.GetWorkspace(ctx, issue.WorkspaceID)
	if err != nil {
		return nil, "", err
	}
	cwd := workspace.RootPath
	todo, err := p.todoFromIssue(ctx, issue, cwd, true)
	if err != nil {
		return nil, "", err
	}
	return todo, cwd, nil
}

// GetGlobal implements todos.GlobalReferenceProvider. The returned TODO carries
// its owning workspace CWD, so callers can immediately adopt the canonical
// workspace without a list/sync fan-out.
func (p *Provider) GetGlobal(ctx context.Context, ref string) (*types.TODO, error) {
	todo, _, err := p.GlobalGet(ctx, ref)
	return todo, err
}

func (p *Provider) Create(ctx context.Context, request todos.CreateRequest) (*types.TODO, error) {
	title := strings.TrimSpace(request.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	priority, err := toNativePriority(request.Priority)
	if err != nil {
		return nil, err
	}
	status, err := durableStatusForCreate(request.Status)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(request.Body)
	issue, err := p.repository.CreateIssue(ctx, native.CreateIssueInput{
		WorkspaceID:  p.workspace.ID,
		Title:        title,
		Body:         body,
		Verification: todos.ExtractVerificationFixture(body),
		Labels:       request.Labels,
		Priority:     priority,
		Status:       status,
		// Execution is projected from an attached Captain prompt run. A create
		// request never manufactures transient execution state.
		ExecutionState: native.ExecutionIdle,
		Actor:          mutationActor,
	})
	if err != nil {
		return nil, err
	}
	return p.todoFromIssue(ctx, issue, p.workDir, true)
}

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
			"model":          result.ExecutorName,
			"durationMillis": result.Duration.Milliseconds(),
			"costUsd":        result.CostUSD,
			"tokens":         result.TokensUsed,
			"turns":          result.NumTurns,
			"commit":         result.CommitSHA,
			"error":          result.ErrorMessage,
		},
	})
}

func (p *Provider) SaveVerification(ctx context.Context, todo *types.TODO, result *verify.VerifyResult) error {
	if result == nil {
		return nil
	}
	return p.appendEvent(ctx, todo, native.EventInput{
		Kind:    "verification_result",
		Actor:   mutationActor,
		Body:    todos.RenderVerificationSection(result),
		Payload: result,
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
	if result.ExecutorName != "" {
		fmt.Fprintf(&body, "- **Model:** %s\n", result.ExecutorName)
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
		fmt.Fprintf(&body, "- **Error:** %s\n", result.ErrorMessage)
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
