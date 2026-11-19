package commit

import (
	"context"

	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
)

func validatePushProvider(ctx context.Context, opts Options) error {
	client, err := provider.Resolve(provider.Options{WorkDir: opts.WorkDir, Repo: opts.target.repo})
	if err != nil {
		return err
	}
	if client.Kind() != "azuredevops" {
		return nil
	}
	return client.Preflight(ctx)
}

func providerPushDeps(ctx context.Context, opts Options) (pushDeps, error) {
	deps := defaultPushDeps(ctx, opts)
	if opts.PushBranch != "" && !opts.AutoMerge {
		return deps, nil
	}
	client, err := provider.Resolve(provider.Options{WorkDir: opts.WorkDir, Repo: opts.target.repo})
	if err != nil {
		return deps, err
	}
	azure, ok := client.(*azuredevops.Client)
	if !ok {
		return deps, nil
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
	deps.enableAutoMerge = func(_ github.Options, pr *model.CreatePRResult, method string) error {
		return azure.EnableAutoMerge(ctx, pr.Number, method)
	}
	return deps, nil
}
