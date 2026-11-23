// What a hosted organization's member hides of its repositories: those they hid one by one
// (hidden-projects, repo:<id>) and every repository of an owner they hid (hidden-repo-owners), in
// their PR list, projects and todos.
package ui

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/filters"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// repoVisibility is what one member hides, as lowercased host/full name and host/owner keys. The
// zero value hides nothing.
type repoVisibility struct {
	repos  map[string]bool
	owners map[string]bool
}

// hides reports whether the repository fullName on host is hidden, by itself or by its owner.
func (v repoVisibility) hides(host, fullName string) bool {
	if len(v.repos) == 0 && len(v.owners) == 0 {
		return false
	}
	key := strings.ToLower(host + "/" + fullName)
	owner, _, _ := strings.Cut(strings.ToLower(fullName), "/")
	return v.repos[key] || v.owners[strings.ToLower(host)+"/"+owner]
}

// hidesPR reports whether pr's repository is hidden: its host is its URL's.
func (v repoVisibility) hidesPR(pr github.PRListItem) bool {
	link, err := url.Parse(pr.URL)
	return err == nil && link.Host != "" && v.hides(link.Host, pr.Repo)
}

// visible is snap without the PRs v hides.
func (v repoVisibility) visible(snap snapshot) snapshot {
	if len(v.repos) == 0 && len(v.owners) == 0 {
		return snap
	}
	shown := make(github.PRSearchResults, 0, len(snap.PRs))
	for _, pr := range snap.PRs {
		if v.hidesPR(pr) {
			delete(snap.GavelResults, prKey(pr))
			continue
		}
		shown = append(shown, pr)
	}
	snap.PRs = shown
	return snap
}

// repoPlace is the host and full name a repository's PRs carry: its remote's, parsed as the PR list
// parses it (owner, Azure DevOps project, name), else github.com's for a GitHub App repository. A
// pushed repository has none.
func repoPlace(source, name, remoteURL string) (host, fullName string, ok bool) {
	if remoteURL != "" {
		if remote, err := git.ParseRepoURL(remoteURL, ""); err == nil && remote.Host != "" {
			return remote.Host, strings.Join(nonEmpty(remote.Owner, remote.Project, remote.Name), "/"), true
		}
	}
	if source == "github" {
		return github.WebHost(), name, true
	}
	return "", "", false
}

// repoVisibility is what scope's member hides of the hosted organization's repositories; a local
// dashboard's PR list hides nothing.
func (s *Server) repoVisibility(ctx context.Context, scope tenant.Scope) (repoVisibility, error) {
	if !s.hosted || scope.IsLocal() || s.orgDB == nil {
		return repoVisibility{}, nil
	}
	store, err := s.filterStore(ctx)
	if err != nil {
		return repoVisibility{}, err
	}
	hiddenRepos, err := store.Get(ctx, scope, filters.HiddenProjects)
	if err != nil {
		return repoVisibility{}, err
	}
	owners, err := s.hiddenOwners(ctx, scope)
	if err != nil {
		return repoVisibility{}, err
	}
	visibility := repoVisibility{repos: map[string]bool{}, owners: owners}
	var ids []uuid.UUID
	for _, value := range hiddenRepos.Values {
		if id, err := uuid.Parse(strings.TrimPrefix(value, "repo:")); err == nil && strings.HasPrefix(value, "repo:") {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return visibility, nil
	}
	var rows []struct {
		Source, Name string
		RemoteURL    *string
	}
	if err := s.orgDB.WithContext(ctx).Raw(`SELECT source, name, remote_url FROM repos WHERE org_id = ? AND id IN ?`, scope.OrgID, ids).
		Scan(&rows).Error; err != nil {
		return repoVisibility{}, fmt.Errorf("read the repositories you hid: %w", err)
	}
	for _, row := range rows {
		remote := ""
		if row.RemoteURL != nil {
			remote = *row.RemoteURL
		}
		if host, fullName, ok := repoPlace(row.Source, row.Name, remote); ok {
			visibility.repos[strings.ToLower(host+"/"+fullName)] = true
		}
	}
	return visibility, nil
}

// hiddenOwners is the repository owners scope's member hides, as lowercased host/owner; none
// outside a hosted dashboard.
func (s *Server) hiddenOwners(ctx context.Context, scope tenant.Scope) (map[string]bool, error) {
	owners := map[string]bool{}
	if s == nil || !s.hosted || scope.IsLocal() {
		return owners, nil
	}
	store, err := s.filterStore(ctx)
	if err != nil {
		return nil, err
	}
	hidden, err := store.Get(ctx, scope, filters.HiddenRepoOwners)
	if err != nil {
		return nil, err
	}
	for _, owner := range hidden.Values {
		owners[owner] = true
	}
	return owners, nil
}

// hostedRepoOwners is the owner of each of the hosted organization's repositories, as lowercased
// host/owner, by its workspace (repo:<id>); a pushed repository has none.
func (s *Server) hostedRepoOwners(ctx context.Context) (map[string]string, error) {
	var rows []struct {
		ID           uuid.UUID
		Source, Name string
		RemoteURL    *string
	}
	if err := s.orgDB.WithContext(ctx).Raw(`SELECT id, source, name, remote_url FROM repos WHERE org_id = ?`, s.orgID()).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("read the organization's repositories: %w", err)
	}
	owners := make(map[string]string, len(rows))
	for _, row := range rows {
		remote := ""
		if row.RemoteURL != nil {
			remote = *row.RemoteURL
		}
		if host, fullName, ok := repoPlace(row.Source, row.Name, remote); ok {
			owner, _, _ := strings.Cut(fullName, "/")
			owners[RepoWorkspace(row.ID)] = strings.ToLower(host + "/" + owner)
		}
	}
	return owners, nil
}

// ownerHiddenWorkspaces is the workspaces (repo:<id>) of the repositories whose owner scope's
// member hides.
func (s *Server) ownerHiddenWorkspaces(ctx context.Context, scope tenant.Scope) ([]string, error) {
	hidden, err := s.hiddenOwners(ctx, scope)
	if err != nil || len(hidden) == 0 || s.orgDB == nil {
		return nil, err
	}
	owners, err := s.hostedRepoOwners(ctx)
	if err != nil {
		return nil, err
	}
	var dirs []string
	for dir, owner := range owners {
		if hidden[owner] {
			dirs = append(dirs, dir)
		}
	}
	return dirs, nil
}

// withOwners marks each hosted project with its repository's owner, and whether the caller hides
// that owner.
func (s *Server) withOwners(ctx context.Context, projects []projectInfo) error {
	if !s.hosted || s.orgDB == nil {
		return nil
	}
	owners, err := s.hostedRepoOwners(ctx)
	if err != nil {
		return err
	}
	hidden, err := s.hiddenOwners(ctx, tenant.FromContext(ctx))
	if err != nil {
		return err
	}
	for i := range projects {
		projects[i].Owner = owners[projects[i].Dir]
		projects[i].OwnerHidden = hidden[projects[i].Owner]
	}
	return nil
}
