package changeset

import (
	"context"
	"fmt"
	"strconv"
)

type runnerProvider struct {
	name   string
	marker string
	read   func(Environment) (Comparison, error)
}

func (p runnerProvider) Name() string { return p.name }
func (p runnerProvider) Detect(env Environment) bool {
	value, _ := strconv.ParseBool(env.Lookup(p.marker))
	return value
}
func (p runnerProvider) Comparison(_ context.Context, env Environment) (Comparison, error) {
	comparison, err := p.read(env)
	comparison.Provider = p.name
	return comparison, err
}

func detectComparison(ctx context.Context, env Environment, providers []Provider) (*Comparison, error) {
	for _, marker := range []string{"CI", "GITHUB_ACTIONS", "TF_BUILD", "GITLAB_CI"} {
		if value := env.Lookup(marker); value != "" {
			if _, err := strconv.ParseBool(value); err != nil {
				return nil, fmt.Errorf("invalid CI marker %s: %w", marker, err)
			}
		}
	}
	if providers == nil {
		providers = []Provider{
			runnerProvider{"github", "GITHUB_ACTIONS", githubComparison},
			runnerProvider{"azure", "TF_BUILD", azureComparison},
			runnerProvider{"gitlab", "GITLAB_CI", gitlabComparison},
		}
	}
	var selected Provider
	for _, provider := range providers {
		if provider.Detect(env) {
			if selected != nil {
				return nil, fmt.Errorf("conflicting CI providers %s and %s", selected.Name(), provider.Name())
			}
			selected = provider
		}
	}
	if selected == nil {
		if ci, _ := strconv.ParseBool(env.Lookup("CI")); ci {
			return nil, fmt.Errorf("unknown CI provider; specify --since or supply an explicit comparison")
		}
		return nil, nil
	}
	comparison, err := selected.Comparison(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("%s changeset: %w", selected.Name(), err)
	}
	if comparison.Provider == "" {
		comparison.Provider = selected.Name()
	}
	return &comparison, nil
}

func previousComparison(base, head string) Comparison {
	if base == "" || zeroSHA(base) {
		return Comparison{Kind: "commit", Head: head}
	}
	return Comparison{Kind: "range", Base: base, Head: head}
}

func zeroSHA(value string) bool {
	for _, c := range value {
		if c != '0' {
			return false
		}
	}
	return value != ""
}
