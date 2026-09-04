// Package runtime provides the PostgreSQL-only TODO runtime.
//
// The package adapts the native repository to the shared TODO lifecycle
// interface used by CLI, API, and UI callers. It never falls back to filesystem
// providers; configured projects initialize their native workspace on first use.
package runtime

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/labels"
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
	prepared    map[uuid.UUID]map[uuid.UUID]struct{}
	// ownership drives the heartbeats that prove this process is still running
	// the prompt runs it dispatched. See ownership.go.
	ownership runOwnership
	// labelsCache memoizes one resolver per workspace for the provider's
	// lifetime. See labelResolver — it is what keeps label rendering off the
	// per-row query path.
	labelsMu    sync.RWMutex
	labelsCache map[uuid.UUID]*labels.Resolver
	// phasesCache memoizes one workspace's per-phase run index for the
	// provider's lifetime. See phaseRuns — it is what keeps the four phase
	// columns off the per-row query path.
	phasesMu    sync.RWMutex
	phasesCache map[uuid.UUID]map[uuid.UUID]types.PhaseRuns
	// execCache memoizes one workspace's prompt-run overviews and links, primed
	// by List. See execution_index.go — it is what keeps the active run, its
	// provider session and the attempt count off the per-row query path.
	execMu    sync.RWMutex
	execCache map[uuid.UUID]*executionIndex
}

var _ todos.Provider = (*Provider)(nil)

// WorkspaceOptions identifies one project from the canonical Gavel project
// catalog. RootPath is required; the first repository is the primary durable
// repository identity.
type WorkspaceOptions struct {
	Name         string
	RootPath     string
	Repositories []string
}

// Open opens the process-owned PostgreSQL pool and initializes the configured
// project's native workspace when it does not already exist.
func Open(ctx context.Context, options WorkspaceOptions) (*Provider, error) {
	db, err := database.Require(ctx, "native TODO storage")
	if err != nil {
		return nil, err
	}
	// Todo commands open the database without migrating it — only `gavel serve`
	// applies migrations — so this is where a binary newer than its database is
	// caught, once, rather than as an obscure failure at whichever read first
	// touches a column that is not there yet.
	if err := requireVerificationColumn(db); err != nil {
		return nil, err
	}
	return New(ctx, db, options)
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
	if err := requireVerificationColumn(db); err != nil {
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
		prepared: map[uuid.UUID]map[uuid.UUID]struct{}{},
	}, nil
}

// New constructs a provider over an already migrated GORM pool. It is useful
// to hosts and tests that own the database lifecycle.
func New(ctx context.Context, db *gorm.DB, options WorkspaceOptions) (*Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	options, err := normalizeWorkspaceOptions(options)
	if err != nil {
		return nil, err
	}
	repository, err := native.NewRepository(db)
	if err != nil {
		return nil, err
	}
	workspace, err := initializeWorkspace(ctx, repository, options)
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
		workDir: options.RootPath, db: db, repository: repository,
		captain: captain, coordinator: coordinator, workspace: workspace,
		prepared: map[uuid.UUID]map[uuid.UUID]struct{}{},
	}, nil
}

func normalizeWorkspaceOptions(options WorkspaceOptions) (WorkspaceOptions, error) {
	options.Name = strings.TrimSpace(options.Name)
	if options.Name == "" {
		return WorkspaceOptions{}, fmt.Errorf("%w: project name is required", native.ErrInvalidInput)
	}
	rootPath := strings.TrimSpace(options.RootPath)
	if rootPath == "" {
		return WorkspaceOptions{}, fmt.Errorf("%w: project %q root path is required", native.ErrInvalidInput, options.Name)
	}
	absolute, err := filepath.Abs(rootPath)
	if err != nil {
		return WorkspaceOptions{}, fmt.Errorf("resolve project %q root path %q: %w", options.Name, rootPath, err)
	}
	options.RootPath = filepath.Clean(absolute)
	repositories := make([]string, 0, len(options.Repositories))
	seen := map[string]bool{}
	for _, repository := range options.Repositories {
		repository = strings.ToLower(strings.Trim(strings.TrimSpace(repository), "/"))
		repository = strings.TrimPrefix(repository, "github.com/")
		if repository == "" {
			continue
		}
		repository = "github.com/" + repository
		if !seen[repository] {
			repositories = append(repositories, repository)
			seen[repository] = true
		}
	}
	options.Repositories = repositories
	return options, nil
}

func initializeWorkspace(ctx context.Context, repository *native.Repository, options WorkspaceOptions) (*native.Workspace, error) {
	workspace, err := resolveWorkspace(ctx, repository, options)
	if err == nil {
		return reconcileWorkspace(ctx, repository, workspace, options)
	}
	if !errors.Is(err, native.ErrNotFound) {
		return nil, err
	}

	input := native.CreateWorkspaceInput{RootPath: options.RootPath, DisplayName: options.Name}
	if len(options.Repositories) > 0 {
		input.RepoKey = options.Repositories[0]
	}
	workspace, err = repository.CreateWorkspace(ctx, input)
	if err == nil {
		return workspace, nil
	}
	if !errors.Is(err, native.ErrWorkspaceConflict) {
		return nil, fmt.Errorf("initialize native TODO workspace for project %q: %w", options.Name, err)
	}
	workspace, resolveErr := resolveWorkspace(ctx, repository, options)
	if resolveErr != nil {
		return nil, fmt.Errorf("resolve concurrently initialized native TODO workspace for project %q after %v: %w", options.Name, err, resolveErr)
	}
	return reconcileWorkspace(ctx, repository, workspace, options)
}

func resolveWorkspace(ctx context.Context, repository *native.Repository, options WorkspaceOptions) (*native.Workspace, error) {
	var matched *native.Workspace
	byPath, err := repository.GetWorkspaceByPath(ctx, options.RootPath)
	switch {
	case err == nil:
		matched = byPath
	case !errors.Is(err, native.ErrNotFound):
		return nil, fmt.Errorf("resolve native TODO workspace by project %q path %q: %w", options.Name, options.RootPath, err)
	}
	for _, repoKey := range options.Repositories {
		byRepo, err := repository.GetWorkspaceByRepoKey(ctx, repoKey)
		switch {
		case err == nil:
			if matched != nil && matched.ID != byRepo.ID {
				return nil, fmt.Errorf(
					"%w: project %q path %q resolves to workspace %s while repository %q resolves to workspace %s",
					native.ErrWorkspaceConflict, options.Name, options.RootPath, matched.ID, repoKey, byRepo.ID,
				)
			}
			matched = byRepo
		case !errors.Is(err, native.ErrNotFound):
			return nil, fmt.Errorf("resolve native TODO workspace by project %q repository %q: %w", options.Name, repoKey, err)
		}
	}
	if matched == nil {
		return nil, fmt.Errorf("%w: native TODO workspace for configured project %q", native.ErrNotFound, options.Name)
	}
	return matched, nil
}

func reconcileWorkspace(ctx context.Context, repository *native.Repository, workspace *native.Workspace, options WorkspaceOptions) (*native.Workspace, error) {
	update := native.UpdateWorkspaceInput{}
	if workspace.RootPath != options.RootPath {
		update.RootPath = &options.RootPath
	}
	if workspace.DisplayName != options.Name {
		update.DisplayName = &options.Name
	}
	if len(options.Repositories) > 0 && workspace.RepoKey != options.Repositories[0] {
		update.RepoKey = &options.Repositories[0]
	}
	if update.RootPath == nil && update.DisplayName == nil && update.RepoKey == nil {
		return workspace, nil
	}
	updated, err := repository.UpdateWorkspace(ctx, workspace.ID, update)
	if err != nil {
		return nil, fmt.Errorf("reconcile native TODO workspace for project %q: %w", options.Name, err)
	}
	return updated, nil
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

func (p *Provider) List(ctx context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	issues, err := p.repository.ListIssues(ctx, p.workspace.ID)
	if err != nil {
		return nil, err
	}
	// One query for the whole workspace's GitHub links, not one per issue.
	links, err := p.repository.ListAliasesByKind(ctx, p.workspace.ID, githubpush.AliasKind)
	if err != nil {
		return nil, err
	}
	// Two more for every issue's active run and run links, for the same reason.
	if err := p.primeExecutionIndex(ctx, p.workspace.ID); err != nil {
		return nil, err
	}
	result := make(types.TODOS, 0, len(issues))
	for index := range issues {
		todo, err := p.todoFromIssue(ctx, &issues[index], p.workDir, false)
		if err != nil {
			return nil, fmt.Errorf("decode native TODO %s: %w", issues[index].ID, err)
		}
		todo.ExternalIssue = externalIssueFrom(links[issues[index].ID])
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
	todo, err := p.todoFromIssue(ctx, issue, p.workDir, true)
	if err != nil {
		return nil, err
	}
	aliases, err := p.repository.ListAliases(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	todo.ExternalIssue = externalIssueFrom(aliases)
	return todo, nil
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
	body, bodyVerification, _ := todos.SplitVerificationFixture(body)
	verification := todos.CombineVerificationFixtures(request.Verification, bodyVerification)
	issueInput := native.CreateIssueInput{
		ID:           uuid.New(),
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
		created, createErr := p.coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
			Issue: issueInput, RootSession: p.todoRootSessionInput(issueInput),
		})
		if createErr != nil {
			err = createErr
		} else {
			issue = created.Issue
		}
	} else {
		issue, err = p.createIssueWithPlan(ctx, issueInput, *request.Plan)
	}
	if err != nil {
		return nil, err
	}
	return p.todoFromIssue(ctx, issue, p.workDir, true)
}
