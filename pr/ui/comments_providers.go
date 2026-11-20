package ui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/tenant"
)

func commentGitHub(ctx context.Context, p comments.Principal, t comments.Target) (github.Options, error) {
	if err := (CommentAccess{}).Authorize(ctx, p, "provider", t); err != nil {
		return github.Options{}, err
	}
	project, err := commentProject(ctx, t.ScopeURI)
	if err != nil {
		return github.Options{}, err
	}
	repo, _, err := github.ParsePRURL(t.URI)
	if err != nil {
		return github.Options{}, err
	}
	opts := github.Options{WorkDir: project.Dir, Repo: repo}
	if s := entityServer(ctx); s != nil {
		opts = s.ghOpts
		opts.WorkDir, opts.Repo = project.Dir, repo
		if s.hosted && opts.TokenSource == nil {
			return github.Options{}, fmt.Errorf("%w: hosted comments require tenant credentials", comments.ErrForbidden)
		}
	}
	return opts, nil
}

// commentConnection reaches the repository of t, an Azure DevOps or GitLab pull request: a hosted
// organization's through the git connection of its repository that holds it; a local project's,
// whose origin must be that repository, through its provider's local connection.
func commentConnection(ctx context.Context, _ comments.Principal, t comments.Target) (*git.Repository, error) {
	pr, err := forge.ParsePullRequestURL(t.URI)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", comments.ErrInvalid, err)
	}
	if s := entityServer(ctx); s != nil && s.hosted {
		return forge.PullRequestRemote(ctx, s.orgDB, tenant.Scope{OrgID: s.orgID()}, t.URI, git.ClientOptions{Defensive: true})
	}
	project, err := commentProject(ctx, t.ScopeURI)
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	command.Dir = project.Dir
	data, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read project origin: %w", err)
	}
	origin, err := git.ParseRepoURL(strings.TrimSpace(string(data)), pr.Provider)
	if err != nil {
		return nil, comments.ErrForbidden
	}
	target, err := git.ParseRepoURL(pr.Repo, pr.Provider)
	if err != nil || !origin.SameRepository(target) {
		return nil, comments.ErrForbidden
	}
	return provider.LocalRemote(ctx, pr.Provider, pr.Repo)
}
