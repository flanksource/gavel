package ui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/github"
)

func commentGitHub(ctx context.Context, p comments.Principal, t comments.Target) (github.Options, error) {
	if err := (commentAccess{}).Authorize(ctx, p, "provider", t); err != nil {
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
func commentAzure(ctx context.Context, _ comments.Principal, t comments.Target) (*azuredevops.Client, error) {
	if s := entityServer(ctx); s != nil && s.hosted {
		return nil, fmt.Errorf("%w: hosted Azure credentials are not configured", comments.ErrForbidden)
	}
	project, err := commentProject(ctx, t.ScopeURI)
	if err != nil {
		return nil, err
	}
	repo, _, err := azuredevops.ParsePRURL(t.URI)
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	command.Dir = project.Dir
	data, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("read project origin: %w", err)
	}
	origin, err := azuredevops.ParseRepository(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}
	if origin != repo {
		return nil, comments.ErrForbidden
	}
	return azuredevops.New(repo), nil
}
