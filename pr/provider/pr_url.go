package provider

import (
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
)

func ResolveURL(opts Options, ref string) (Client, int, error) {
	if IsAzure(ref) {
		repo, number, err := azuredevops.ParsePRURL(ref)
		if err != nil {
			return nil, 0, err
		}
		return azuredevops.New(repo), number, nil
	}
	repo, number, err := github.ParsePRURL(ref)
	if err != nil {
		return nil, 0, err
	}
	opts.Repo = repo
	client, err := Resolve(opts)
	return client, number, err
}
