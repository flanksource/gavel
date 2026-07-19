// Package runtime provides the PostgreSQL-only TODO runtime.
//
// The package adapts the native repository to the shared TODO lifecycle
// interface used by CLI, API, and UI callers. It never falls back to Grite or
// .todos files and never creates workspaces as a side effect of opening it.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
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

// GetGlobalBySession implements todos.GlobalSessionReferenceProvider. Session
// ownership is resolved through Gavel's durable prompt-run links, then decoded
// through the owning workspace exactly like an ordinary global issue lookup.
func (p *Provider) GetGlobalBySession(ctx context.Context, ref string) (*types.TODO, string, error) {
	issue, sessionID, err := p.repository.GetIssueBySessionRef(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	workspace, err := p.repository.GetWorkspace(ctx, issue.WorkspaceID)
	if err != nil {
		return nil, "", err
	}
	todo, err := p.todoFromIssue(ctx, issue, workspace.RootPath, true)
	if err != nil {
		return nil, "", err
	}
	return todo, sessionID, nil
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
	verification := strings.TrimSpace(request.Verification)
	if verification == "" {
		verification = todos.ExtractVerificationFixture(body)
	}
	issueInput := native.CreateIssueInput{
		WorkspaceID:  p.workspace.ID,
		Title:        title,
		Body:         body,
		Verification: verification,
		Labels:       request.Labels,
		Priority:     priority,
		Status:       status,
		Actor:        mutationActor,
	}
	var issue *native.Issue
	if request.Plan == nil {
		issue, err = p.repository.CreateIssue(ctx, issueInput)
	} else {
		issue, err = p.createIssueWithPlan(ctx, issueInput, *request.Plan)
	}
	if err != nil {
		return nil, err
	}
	return p.todoFromIssue(ctx, issue, p.workDir, true)
}
