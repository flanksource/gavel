// Resolves a pull request URL to the client that reads it and its number: an Azure DevOps or
// GitLab one through its provider's local connection, a GitHub one through GitHub.
package provider

import (
	"context"

	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
)

func ResolveURL(opts Options, ref string) (Client, int, error) {
	if pr, err := forge.ParsePullRequestURL(ref); err == nil && inJob() == nil {
		ctx := opts.Context
		if ctx == nil {
			ctx = context.Background()
		}
		remote, err := LocalRemote(ctx, pr.Provider, pr.Repo)
		if err != nil {
			return nil, 0, err
		}
		return NewConnectionClient(remote, ""), pr.Number, nil
	}
	repo, number, err := github.ParsePRURL(ref)
	if err != nil {
		return nil, 0, err
	}
	opts.Repo = repo
	client, err := Resolve(opts)
	return client, number, err
}
