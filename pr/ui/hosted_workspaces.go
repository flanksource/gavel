// Hosted todo workspaces: a hosted organization's repositories are its todo workspaces, named
// repo:<id>, each backed by serve's read-only checkout of the repository's default branch.
package ui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/sandbox"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// errHostedWorkspace is a hosted todo request whose workspace isn't named as one of the
// organization's repositories.
var errHostedWorkspace = errors.New("is not one of this organization's repositories")

// errNoHostedWorkspaces is a todo request to a hosted dashboard that serves no organization's
// repositories, such as the one answering before an organization is chosen.
var errNoHostedWorkspaces = errors.New("this hosted dashboard has no todo workspaces")

// HostedWorkspaces resolves a hosted organization's repositories to todo workspaces.
type HostedWorkspaces struct {
	DB      *gorm.DB
	Sources *sandbox.Sources
}

// repoQuery selects org's repositories, shaped as the sandbox reads them.
func (w HostedWorkspaces) repoQuery(ctx context.Context, org uuid.UUID) *gorm.DB {
	return w.DB.WithContext(ctx).Table("repos").
		Select("id, org_id AS org, source, name, COALESCE(github_repo_id, 0) AS git_hub_repo_id, COALESCE(installation_id, 0) AS installation_id").
		Where("org_id = ?", org)
}

// repos reads org's repositories, or the one id names.
func (w HostedWorkspaces) repos(ctx context.Context, org uuid.UUID, id *uuid.UUID) ([]sandbox.Repo, error) {
	query := w.repoQuery(ctx, org)
	if id != nil {
		query = query.Where("id = ?", *id)
	}
	var repos []sandbox.Repo
	if err := query.Order("name").Scan(&repos).Error; err != nil {
		return nil, fmt.Errorf("read the repositories of organization %s: %w", org, err)
	}
	return repos, nil
}

// Projects lists org's repositories as the dashboard's projects.
func (w HostedWorkspaces) Projects(ctx context.Context, org uuid.UUID) ([]projectInfo, error) {
	repos, err := w.repos(ctx, org, nil)
	if err != nil {
		return nil, err
	}
	out := make([]projectInfo, 0, len(repos))
	for _, repo := range repos {
		info := projectInfo{Name: repo.Name, Short: repo.Name, Dir: RepoWorkspace(repo.ID), Repos: []string{}, TodoBackend: "native"}
		if repo.Source == "github" {
			info.Repos = []string{repo.Name}
		}
		out = append(out, info)
	}
	return out, nil
}

// HostedWorkspace is one of the organization's repositories as a todo workspace, checked out at
// SHA, its default branch's commit.
type HostedWorkspace struct {
	Options todoruntime.WorkspaceOptions
	Repo    sandbox.Repo
	SHA     string
}

// Resolve checks out the repository ref names and returns its workspace. ref is repo:<id>, or
// the repository's checkout path as an earlier response gave it.
func (w HostedWorkspaces) Resolve(ctx context.Context, org uuid.UUID, ref string) (HostedWorkspace, error) {
	repo, err := w.repo(ctx, org, ref)
	if err != nil {
		return HostedWorkspace{}, err
	}
	checkout, err := w.Sources.Checkout(ctx, repo)
	if err != nil {
		return HostedWorkspace{}, err
	}
	workspace := HostedWorkspace{
		Options: todoruntime.WorkspaceOptions{Name: repo.Name, RootPath: checkout.Dir, SandboxedRuns: true},
		Repo:    repo,
		SHA:     checkout.SHA,
	}
	if repo.Source == "github" {
		workspace.Options.Repositories = []string{repo.Name}
	}
	return workspace, nil
}

// repo is the organization's repository ref names: repo:<id>, or its checkout path.
func (w HostedWorkspaces) repo(ctx context.Context, org uuid.UUID, ref string) (sandbox.Repo, error) {
	ref = strings.TrimSpace(ref)
	id, ok := w.repoID(org, ref)
	if !ok {
		return sandbox.Repo{}, fmt.Errorf("todo workspace %q %w: name it repo:<id>", ref, errHostedWorkspace)
	}
	repos, err := w.repos(ctx, org, &id)
	if err != nil {
		return sandbox.Repo{}, err
	}
	if len(repos) == 0 {
		return sandbox.Repo{}, fmt.Errorf("%w: todo workspace %q %w", ErrProjectNotFound, ref, errHostedWorkspace)
	}
	return repos[0], nil
}

// GitDir is the git directory of the repository ref names, as Resolve reads ref, without
// checking it out: a run's commits are read from it.
func (w HostedWorkspaces) GitDir(ctx context.Context, org uuid.UUID, ref string) (string, error) {
	repo, err := w.repo(ctx, org, ref)
	if err != nil {
		return "", err
	}
	return w.Sources.GitDir(ctx, repo)
}

func (w HostedWorkspaces) repoID(org uuid.UUID, ref string) (uuid.UUID, bool) {
	if raw, ok := strings.CutPrefix(ref, "repo:"); ok {
		id, err := uuid.Parse(raw)
		return id, err == nil
	}
	root := filepath.Join(w.Sources.DataDir, "workspaces", org.String()) + string(filepath.Separator)
	raw, ok := strings.CutPrefix(filepath.Clean(ref), root)
	if !ok || strings.ContainsRune(raw, filepath.Separator) {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	return id, err == nil
}

// projectList is the dashboard's projects: the organization's repositories in hosted mode,
// the host's registered projects otherwise.
func (s *Server) projectList(ctx context.Context) ([]projectInfo, error) {
	if s.workspaces != nil {
		return s.workspaces.Projects(ctx, s.orgID())
	}
	if s.hosted {
		return nil, errNoHostedWorkspaces
	}
	return listProjects(ctx)
}
