package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/pr/blame"
	"github.com/flanksource/gavel/pr/provider"
)

type PRBlameOptions struct {
	Repo         string          `flag:"repo" short:"R" help:"GitHub owner/repo or hosted Azure DevOps repository URL; requires a matching local checkout"`
	Limit        int             `flag:"limit" default:"5" help:"Maximum recommended reviewers to return"`
	HistoryLimit int             `flag:"history-limit" default:"100" help:"Maximum historical commits to inspect per changed file"`
	Since        string          `flag:"since" help:"Only count historical contributions since this date or duration (e.g. 30d, now-90d, 2026-01-01); ownership rules still apply"`
	Exclude      []string        `flag:"exclude" help:"Additional logins, email addresses or ADO identity IDs to exclude"`
	Args         []string        `args:"true"`
	Context      context.Context `json:"-"`
}

func (o PRBlameOptions) Help() api.Textable {
	return clicky.Text(`Recommend reviewers for GitHub and Azure DevOps pull requests using ownership rules, changed-line blame, insertion context and historical file contributions.

Accepts the current branch's PR, a PR number, owner/repo + number,
owner/repo#number, or a full GitHub or hosted Azure DevOps PR URL. Run inside a checkout with a remote
pointing at the PR's base repository and full history. Missing PR endpoints are
fetched without moving branches, remote-tracking refs, the index or working files.

CODEOWNERS is read from the base branch tip (.github/, root, then docs/), with
the last matching rule winning. Historical blame and file commits use the PR's
merge base, excluding changes introduced by the PR itself. Pure insertions use
adjacent existing lines, reported separately. New files use ownership rules.

--since limits file history, changed-line blame and insertion context by commit date (inclusive). Accepts 30d, now-90d or 2026-01-01. CODEOWNERS and ADO reviewer policies still apply. --history-limit caps the matching commits per file after the date filter.

ADO uses native automatically included reviewer policies for the selected PR, preserving required versus optional rules, evaluation status and groups. A stale merge snapshot must settle before recommendations are returned. Git authors retain their name and email; policy identities merge with matching Git email addresses. CODEOWNERS files are supplementary conventions on ADO, rather than native branch policies. Authentication uses AZURE_DEVOPS_EXT_PAT or Azure CLI login.

Candidates rank by required policies, applicable policies, owned files, changed lines, adjacent lines, file commits,
then file coverage. The PR author and bot accounts are excluded. Teams remain
team recommendations; membership and reviewer permissions are not checked.
Unlinked Git authors retain their name and email. No review requests are sent.

Examples:
  gavel pr blame
  gavel pr blame 123
  gavel pr blame 123 --since 90d
  gavel pr blame acme/widgets#123 --json
  gavel pr blame https://github.com/acme/widgets/pull/123 --limit 10
  gavel pr blame https://dev.azure.com/acme/product/_git/service/pullrequest/27
  gavel pr blame 123 --history-limit 50 --exclude former-maintainer`)
}

func runPRBlame(opts PRBlameOptions) (any, error) {
	if opts.Limit <= 0 || opts.HistoryLimit <= 0 {
		return nil, fmt.Errorf("--limit and --history-limit must be positive")
	}
	var since time.Time
	if opts.Since != "" {
		parsed, err := parseSince(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("--since: %w", err)
		}
		since = parsed.UTC().Truncate(time.Second)
	}
	target, err := parseStatusTarget(opts.Args)
	if err != nil {
		return nil, err
	}
	if target.RunID != 0 {
		return nil, fmt.Errorf("pr blame requires a pull request, not an Actions run or job")
	}
	dir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	if provider.InJob() {
		return nil, fmt.Errorf("pr blame requires a local checkout")
	}
	repo := target.Repo
	if repo == "" {
		repo = opts.Repo
	}
	client, err := provider.Resolve(provider.Options{Repo: repo, WorkDir: dir})
	if err != nil {
		return nil, err
	}
	return client.RecommendReviewers(blame.Request{Context: opts.Context, Dir: dir, Number: target.PR,
		Limit: opts.Limit, HistoryLimit: opts.HistoryLimit, Exclude: opts.Exclude, Since: since})
}

func init() {
	clicky.AddNamedCommand("blame", prCmd, PRBlameOptions{}, runPRBlame)
}
