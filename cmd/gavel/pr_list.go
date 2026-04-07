package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/github"
	"github.com/timberio/go-datemath"
)

type PRListOptions struct {
	Author  string   `flag:"author" help:"GitHub username (default: @me)"`
	Since   string   `flag:"since" help:"Show PRs updated since (e.g. 7d, now-30d, 2024-01-01)" default:"7d"`
	State   string   `flag:"state" help:"PR state: open, closed, merged, all" default:"open"`
	All     bool     `flag:"all" help:"List PRs across all repos in the org"`
	Any     bool     `flag:"any" help:"Show PRs from all authors (remove @me filter)"`
	Bots    bool     `flag:"bots" help:"Include bot-authored PRs" default:"false"`
	Org     string   `flag:"org" help:"GitHub org for --all (auto-detected from git remote)"`
	Limit   int      `flag:"limit" help:"Maximum PRs to return" default:"50"`
	Status  bool     `flag:"status" help:"Show GitHub Actions check status counts"`
	Verbose bool     `flag:"verbose" help:"With --status, show failed steps and log tails"`
	URL     bool     `flag:"url" help:"Show PR URL instead of number"`
	Repos   []string `args:"true"`
}

var bareDurationRe = regexp.MustCompile(`^\d+[dhms]$`)

func parseSince(s string) (time.Time, error) {
	if bareDurationRe.MatchString(s) {
		s = "now-" + s
	}
	expr, err := datemath.Parse(s)
	if err != nil {
		t, err2 := time.Parse("2006-01-02", s)
		if err2 != nil {
			return time.Time{}, fmt.Errorf("unable to parse since: %s", s)
		}
		return t, nil
	}
	return expr.Time(datemath.WithNow(time.Now())), nil
}

func resolveRepoArg(arg string) (string, error) {
	if strings.HasPrefix(arg, "https://") || strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "github.com/") {
		return github.ParseRepoURL(arg)
	}

	if strings.Contains(arg, "/") {
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			return github.ResolveRepoFromDir(arg)
		}
		return arg, nil
	}

	abs, err := filepath.Abs(arg)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", arg, err)
	}
	return github.ResolveRepoFromDir(abs)
}

func runPRList(opts PRListOptions) (any, error) {
	author := opts.Author
	if author == "" && !opts.Any {
		author = "@me"
	}

	var since time.Time
	if opts.Since != "" {
		var err error
		since, err = parseSince(opts.Since)
		if err != nil {
			return nil, err
		}
	}

	searchOpts := github.PRSearchOptions{
		Author:  author,
		Since:   since,
		State:   opts.State,
		All:     opts.All,
		Org:     opts.Org,
		Limit:   opts.Limit,
		Status:  opts.Status,
		Verbose: opts.Verbose,
		ShowURL: opts.URL,
	}

	if len(opts.Repos) > 0 {
		for _, arg := range opts.Repos {
			repo, err := resolveRepoArg(arg)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve %q: %w", arg, err)
			}
			searchOpts.Repos = append(searchOpts.Repos, repo)
		}
	}

	var ghOpts github.Options
	if len(searchOpts.Repos) == 0 {
		workDir, err := getWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	results, err := github.SearchPRs(ghOpts, searchOpts)
	if err != nil {
		return nil, err
	}

	if !opts.Bots {
		results = filterByBot(results, false)
	}

	return results, nil
}

var knownBots = map[string]bool{
	"dependabot": true, "renovate": true, "greenkeeper": true,
	"snyk-bot": true, "imgbot": true, "allcontributors": true,
}

func isBot(author string) bool {
	return strings.HasSuffix(author, "[bot]") || knownBots[author]
}

func filterByBot(items github.PRSearchResults, botsOnly bool) github.PRSearchResults {
	var filtered github.PRSearchResults
	for _, item := range items {
		if botsOnly && isBot(item.Author) {
			filtered = append(filtered, item)
		} else if !botsOnly && !isBot(item.Author) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func init() {
	clicky.AddNamedCommand("list", prCmd, PRListOptions{}, runPRList)
}
