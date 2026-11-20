package provider

import (
	"strings"

	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
)

func ResolveURL(opts Options, ref string) (Client, int, error) {
	if IsAzure(ref) && inJob() == nil {
		_, number, err := azuredevops.ParsePRURL(ref)
		if err != nil {
			return nil, 0, err
		}
		repoURL, _, _ := strings.Cut(ref, "/pullrequest/")
		opts.Repo = repoURL
		client, err := Resolve(opts)
		return client, number, err
	}
	repo, number, err := github.ParsePRURL(ref)
	if err != nil {
		return nil, 0, err
	}
	opts.Repo = repo
	client, err := Resolve(opts)
	return client, number, err
}
