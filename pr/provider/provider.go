package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
)

type Options struct {
	WorkDir, Repo, Token string
	Context              context.Context
}

type Client interface {
	Kind() string
	Preflight(context.Context) error
	DefaultBranch(context.Context) (string, error)
	OpenPRs(context.Context) ([]model.PRInfo, error)
	CreatePR(context.Context, model.CreatePRInput) (*model.CreatePRResult, error)
	Status(context.Context, int, model.StatusOptions) (*model.StatusSnapshot, error)
}

func Resolve(opts Options) (Client, error) {
	repo := opts.Repo
	if repo == "" {
		remote, err := Origin(opts.WorkDir)
		if err != nil {
			return nil, err
		}
		repo = remote
	}
	if IsAzure(repo) {
		parsed, err := azuredevops.ParseRepository(repo)
		if err != nil {
			return nil, err
		}
		return azuredevops.New(parsed), nil
	}
	repo = strings.TrimPrefix(repo, "git@github.com:")
	repo = strings.TrimSuffix(repo, ".git")
	if strings.Contains(repo, ":") || strings.Contains(repo, "github.com/") {
		parsed, err := github.ParseRepoURL(repo)
		if err != nil {
			return nil, err
		}
		repo = parsed
	}
	if len(strings.Split(repo, "/")) != 2 {
		return nil, fmt.Errorf("unsupported PR repository %q", repo)
	}
	return gitHubClient{opts: github.Options{WorkDir: opts.WorkDir, Repo: repo, Token: opts.Token}}, nil
}

func IsAzure(ref string) bool {
	return strings.Contains(ref, "dev.azure.com") || strings.Contains(ref, "visualstudio.com")
}

func Origin(dir string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url origin: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func PreflightCreate(ctx context.Context, opts Options) error {
	client, err := Resolve(opts)
	if err != nil {
		return err
	}
	if opts.Repo != "" {
		origin, err := Origin(opts.WorkDir)
		if err != nil {
			return err
		}
		originClient, err := Resolve(Options{Repo: origin, WorkDir: opts.WorkDir, Token: opts.Token})
		if err != nil {
			return err
		}
		if client.Kind() != originClient.Kind() {
			return fmt.Errorf("PR repository does not match origin; pushes target origin")
		}
		if client.Kind() == "azuredevops" {
			requested, err := azuredevops.ParseRepository(opts.Repo)
			if err != nil {
				return err
			}
			remote, err := azuredevops.ParseRepository(origin)
			if err != nil {
				return err
			}
			if !strings.EqualFold(requested.Organization, remote.Organization) || !strings.EqualFold(requested.Name, remote.Name) || (requested.Project != "" && remote.Project != "" && !strings.EqualFold(requested.Project, remote.Project)) {
				return fmt.Errorf("azure PR repository does not match origin; pushes target origin")
			}
			if requested.Project == "" || remote.Project == "" {
				requested, err = client.(*azuredevops.Client).Repository(ctx)
				if err != nil {
					return err
				}
				remote, err = originClient.(*azuredevops.Client).Repository(ctx)
				if err != nil {
					return err
				}
				if !strings.EqualFold(requested.URL(), remote.URL()) {
					return fmt.Errorf("azure PR repository does not match origin; pushes target origin")
				}
			}
		} else if client.(gitHubClient).opts.Repo != originClient.(gitHubClient).opts.Repo {
			return fmt.Errorf("GitHub PR repository does not match origin; pushes target origin")
		}
	}
	return client.Preflight(ctx)
}

func CreatePR(ctx context.Context, opts Options, input model.CreatePRInput) (*model.CreatePRResult, error) {
	client, err := Resolve(opts)
	if err != nil {
		return nil, err
	}
	return client.CreatePR(ctx, input)
}
