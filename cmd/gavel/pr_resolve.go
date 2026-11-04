package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/pr/provider"
)

type PRResolveOptions struct {
	Repo   string   `flag:"repo" short:"R" help:"Repository owning the pull request"`
	Thread string   `flag:"thread" help:"Native review thread ID from gavel pr status --json"`
	Args   []string `args:"true"`
}

func (o PRResolveOptions) Help() api.Textable {
	return clicky.Text(`Resolve an addressed review thread on an explicitly selected pull request.

The thread must belong to the PR and be resolvable by the current viewer.
Find its native threadId with gavel pr status <PR> --json.

Example:
  gavel pr resolve 123 --thread PRRT_example --repo acme/widgets`)
}

func runPRResolve(opts PRResolveOptions) (any, error) {
	if strings.TrimSpace(opts.Thread) == "" {
		return nil, fmt.Errorf("--thread is required")
	}
	target, err := parseStatusTarget(opts.Args)
	if err != nil {
		return nil, err
	}
	if target.PR <= 0 || target.RunID != 0 {
		return nil, fmt.Errorf("an explicit PR number or URL is required")
	}
	optsGH, err := prGitHubOptions(target.Repo, opts.Repo)
	if err != nil {
		return nil, err
	}
	client, err := provider.Resolve(provider.Options{WorkDir: optsGH.WorkDir, Repo: optsGH.Repo})
	if err != nil {
		return nil, err
	}
	if err := client.ResolveThread(context.Background(), target.PR, opts.Thread); err != nil {
		return nil, err
	}
	return clicky.Text(fmt.Sprintf("Resolved review thread %s on PR #%d", opts.Thread, target.PR)), nil
}

func init() {
	clicky.AddNamedCommand("resolve", prCmd, PRResolveOptions{}, runPRResolve)
}
