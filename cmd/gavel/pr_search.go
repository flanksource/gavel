package main

import (
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/github/cache"
)

type PRSearchOptions struct {
	Repo  string   `flag:"repo" short:"R" help:"Only PRs in this repository (owner/name, case-insensitive)"`
	State string   `flag:"state" help:"open, closed, merged or all" default:"all"`
	Limit int      `flag:"limit" short:"L" help:"Maximum number of PRs to list (1-200)" default:"20"`
	Args  []string `args:"true"`
}

func (PRSearchOptions) Help() api.Textable {
	return clicky.Text(`Search the PR index the PR dashboard keeps in the Gavel database.

Read-only and offline: it never calls GitHub. The index holds the PRs the
dashboard poller (gavel pr list --ui / gavel serve) has seen, so a PR it has
not synced yet is not found.

Words match PR titles as prefixes, all of them required. A number ("123" or
"#123") matches that PR number. With no words it lists the most recently
updated PRs.

Examples:
  gavel pr search parser panic           # titles with words starting parser and panic
  gavel pr search 123 --repo acme/widgets
  gavel pr search --state open --limit 50`)
}

// prSearchRow is one PR as the search table shows it.
type prSearchRow struct {
	Number int    `json:"number"`
	Repo   string `json:"repo"`
	State  string `json:"state"`
	Title  string `json:"title"`
	URL    string `json:"url"`
}

func runPRSearch(opts PRSearchOptions) ([]prSearchRow, error) {
	store, err := cache.Open()
	if err != nil {
		return nil, err
	}
	defer store.Close() //nolint:errcheck
	prs, err := store.SearchPRs(cache.PRQuery{
		Text: strings.Join(opts.Args, " "), Repo: opts.Repo, State: opts.State, Limit: opts.Limit,
	})
	if err != nil {
		return nil, err
	}
	rows := make([]prSearchRow, len(prs))
	for i, pr := range prs {
		rows[i] = prSearchRow{Number: pr.Number, Repo: pr.Repo, State: pr.State, Title: pr.Title, URL: pr.URL}
	}
	return rows, nil
}

func init() {
	clicky.AddNamedCommand("search", prCmd, PRSearchOptions{}, runPRSearch)
}
