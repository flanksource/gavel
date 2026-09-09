package main

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/github"
)

type PRCloseOptions struct {
	Repo    string   `flag:"repo" short:"R" help:"GitHub repository (owner/repo)"`
	Comment string   `flag:"comment" short:"c" help:"Post this comment on the PR before closing it"`
	Args    []string `args:"true"`
}

func (o PRCloseOptions) Help() api.Textable {
	return clicky.Text(`Close a pull request without merging it — the gavel alternative to gh pr close.

With no argument it closes the current branch's PR. Accepts a PR number,
owner/repo + number, owner/repo#number, or a full PR URL, the same references
'gavel pr status' takes.

The PR is fetched first so an already-merged or already-closed PR fails before
any mutation is sent. Closing is reversible — reopen the PR on github.com.

--comment posts a comment before closing, so the note lands while the PR is
still open. If the comment cannot be posted the PR is left open rather than
closed without it.

Examples:
  gavel pr close                                # current branch's PR
  gavel pr close 123                            # PR #123 in this repo
  gavel pr close owner/repo 123                 # PR #123 in another repo
  gavel pr close https://github.com/o/r/pull/1  # by URL
  gavel pr close 123 -c "superseded by #130"    # comment, then close`)
}

// PRCloseResult is the closed PR, rendered as the command's output. Comment is
// set only when --comment posted one.
type PRCloseResult struct {
	Repo    string `json:"repo,omitempty"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	State   string `json:"state"`
	URL     string `json:"url,omitempty"`
	Comment string `json:"comment,omitempty"`
}

func runPRClose(opts PRCloseOptions) (any, error) {
	// A --comment that is present but blank is a mistake worth catching before
	// any PR is touched, rather than silently closing without the note.
	if opts.Comment != "" && strings.TrimSpace(opts.Comment) == "" {
		return nil, fmt.Errorf("--comment is blank; pass a non-empty comment or omit the flag")
	}

	repo, prNumber, err := parseStatusArgs(opts.Args)
	if err != nil {
		return nil, err
	}

	ghOpts, err := prGitHubOptions(repo, opts.Repo)
	if err != nil {
		return nil, err
	}

	// Resolving the PR first turns "already merged" and "already closed" into
	// a clear local error instead of an opaque GraphQL rejection, and yields
	// the node ID the close mutation needs.
	pr, err := github.FetchPR(ghOpts, prNumber)
	if err != nil {
		return nil, err
	}

	switch strings.ToUpper(pr.State) {
	case "MERGED":
		return nil, fmt.Errorf("PR #%d (%s) is already merged and cannot be closed", pr.Number, pr.Title)
	case "CLOSED":
		return nil, fmt.Errorf("PR #%d (%s) is already closed", pr.Number, pr.Title)
	}

	// Comment first so the note lands while the PR is still open, and let a
	// failure here abort: closing anyway would silently drop the explanation
	// the user asked to leave behind.
	if opts.Comment != "" {
		if err := github.CommentOnPullRequest(ghOpts, pr.NodeID, opts.Comment); err != nil {
			return nil, err
		}
	}

	if err := github.ClosePullRequest(ghOpts, pr.NodeID); err != nil {
		return nil, err
	}

	return &PRCloseResult{
		Repo:    ghOpts.Repo,
		Number:  pr.Number,
		Title:   pr.Title,
		State:   "CLOSED",
		URL:     pr.URL,
		Comment: opts.Comment,
	}, nil
}

func init() {
	clicky.AddNamedCommand("close", prCmd, PRCloseOptions{}, runPRClose)
}
