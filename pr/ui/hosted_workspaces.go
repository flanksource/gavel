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

// HostedWorkspaces resolves a hosted organization's repositories to todo workspaces.
type HostedWorkspaces struct {
	DB      *gorm.DB
	Sources *sandbox.Sources
}

// repos reads org's repositories, or the one id names, shaped as the sandbox reads them.
func (w HostedWorkspaces) repos(ctx context.Context, org uuid.UUID, id *uuid.UUID) ([]sandbox.Repo, error) {
	query := w.DB.WithContext(ctx).Table("repos").
		Select("id, org_id AS org, source, name, COALESCE(github_repo_id, 0) AS git_hub_repo_id, COALESCE(installation_id, 0) AS installation_id").
		Where("org_id = ?", org)
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

// Resolve checks out the repository ref names and returns its workspace. ref is repo:<id>, or
// the repository's checkout path as an earlier response gave it.
func (w HostedWorkspaces) Resolve(ctx context.Context, org uuid.UUID, ref string) (todoruntime.WorkspaceOptions, error) {
	ref = strings.TrimSpace(ref)
	id, ok := w.repoID(org, ref)
	if !ok {
		return todoruntime.WorkspaceOptions{}, fmt.Errorf("todo workspace %q %w: name it repo:<id>", ref, errHostedWorkspace)
	}
	repos, err := w.repos(ctx, org, &id)
	if err != nil {
		return todoruntime.WorkspaceOptions{}, err
	}
	if len(repos) == 0 {
		return todoruntime.WorkspaceOptions{}, fmt.Errorf("%w: todo workspace %q %w", ErrProjectNotFound, ref, errHostedWorkspace)
	}
	checkout, err := w.Sources.Checkout(ctx, repos[0])
	if err != nil {
		return todoruntime.WorkspaceOptions{}, err
	}
	options := todoruntime.WorkspaceOptions{Name: repos[0].Name, RootPath: checkout.Dir}
	if repos[0].Source == "github" {
		options.Repositories = []string{repos[0].Name}
	}
	return options, nil
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
	return listProjects(ctx)
}
