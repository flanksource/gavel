package commit

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
)

func validatePushProvider(ctx context.Context, opts Options) error {
	client, err := provider.Resolve(provider.Options{WorkDir: opts.WorkDir})
	if err != nil {
		return err
	}
	if client.Kind() != "azuredevops" {
		return nil
	}
	if opts.AutoMerge {
		return fmt.Errorf("--auto-merge is not supported for Azure DevOps")
	}
	return client.Preflight(ctx)
}

func providerPushDeps(ctx context.Context, opts Options) (pushDeps, error) {
	deps := defaultPushDeps()
	if opts.PushBranch != "" && !opts.AutoMerge {
		return deps, nil
	}
	client, err := provider.Resolve(provider.Options{WorkDir: opts.WorkDir})
	if err != nil {
		return deps, err
	}
	if client.Kind() != "azuredevops" {
		return deps, nil
	}
	if opts.AutoMerge {
		return deps, fmt.Errorf("--auto-merge is not supported for Azure DevOps")
	}
	if err := client.Preflight(ctx); err != nil {
		return deps, err
	}
	deps.searchPRs = func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
		prs, err := client.OpenPRs(ctx)
		if err != nil {
			return nil, nil, err
		}
		out := make(github.PRSearchResults, 0, len(prs))
		for _, pr := range prs {
			out = append(out, github.PRListItem{Number: pr.Number, Title: pr.Title, Source: pr.HeadRefName, Target: pr.BaseRefName, URL: pr.URL, State: pr.State})
		}
		return out, nil, nil
	}
	deps.defaultBranch = func(github.Options) (string, error) { return client.DefaultBranch(ctx) }
	deps.createPR = func(_ github.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
		return client.CreatePR(ctx, in)
	}
	return deps, nil
}
