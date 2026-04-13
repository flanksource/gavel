package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/menubar"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/timberio/go-datemath"
)

type PRListOptions struct {
	Author   string   `flag:"author" help:"GitHub username (default: @me)"`
	Since    string   `flag:"since" help:"Show PRs updated since (e.g. 7d, now-30d, 2024-01-01)" default:"7d"`
	State    string   `flag:"state" help:"PR state: open, closed, merged, all" default:"open"`
	All      bool     `flag:"all" help:"List PRs across all repos in the org"`
	Any      bool     `flag:"any" help:"Show PRs from all authors (remove @me filter)"`
	Bots     bool     `flag:"bots" help:"Include bot-authored PRs" default:"false"`
	Org      string   `flag:"org" help:"GitHub org for --all (auto-detected from git remote)"`
	Limit    int      `flag:"limit" help:"Maximum PRs to return" default:"50"`
	Status   bool     `flag:"status" help:"Show GitHub Actions check status counts"`
	Logs     bool     `flag:"logs" help:"Fetch failed job log tails (requires --status -v, uses extra API quota)"`
	URL      bool     `flag:"url" help:"Show PR URL instead of number"`
	UI       bool     `flag:"ui" help:"Open PR dashboard in browser with live updates"`
	MenuBar  bool     `flag:"menu-bar" help:"Show macOS menu bar status indicator"`
	Interval string   `flag:"interval" help:"Poll interval for --ui/--menu-bar (e.g. 30s, 1m, 5m)" default:"60s"`
	Repos    []string `args:"true"`
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

func buildPRSearchOpts(opts PRListOptions) (github.Options, github.PRSearchOptions, error) {
	author := opts.Author
	if author == "" && !opts.Any {
		author = "@me"
	}

	var since time.Time
	if opts.Since != "" {
		var err error
		since, err = parseSince(opts.Since)
		if err != nil {
			return github.Options{}, github.PRSearchOptions{}, err
		}
	}

	searchOpts := github.PRSearchOptions{
		Author:     author,
		Since:      since,
		State:      opts.State,
		All:        opts.All,
		Org:        opts.Org,
		Limit:      opts.Limit,
		Status:     opts.Status,
		Verbose:    clicky.Flags.LevelCount > 0,
		FetchLogs:  opts.Logs,
		ShowURL:    opts.URL,
		ShowAuthor: author != "@me",
	}

	if len(opts.Repos) > 0 {
		for _, arg := range opts.Repos {
			repo, err := resolveRepoArg(arg)
			if err != nil {
				return github.Options{}, github.PRSearchOptions{}, fmt.Errorf("cannot resolve %q: %w", arg, err)
			}
			searchOpts.Repos = append(searchOpts.Repos, repo)
		}
	}

	var ghOpts github.Options
	if len(searchOpts.Repos) == 0 {
		workDir, err := getWorkingDir()
		if err != nil {
			return github.Options{}, github.PRSearchOptions{}, fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	return ghOpts, searchOpts, nil
}

func runPRList(opts PRListOptions) (any, error) {
	if opts.UI || opts.MenuBar {
		return nil, runPRUI(opts)
	}

	ghOpts, searchOpts, err := buildPRSearchOpts(opts)
	if err != nil {
		return nil, err
	}

	results, _, err := github.SearchPRs(ghOpts, searchOpts)
	if err != nil {
		return nil, err
	}

	if !opts.Bots {
		results = filterByBot(results, false)
	}

	return results, nil
}

func runPRUI(opts PRListOptions) error {
	ghOpts, searchOpts, err := buildPRSearchOpts(opts)
	if err != nil {
		return err
	}

	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval %q: %w", opts.Interval, err)
	}

	searchOpts.Status = true

	author := searchOpts.Author
	isAny := opts.Any
	isBots := opts.Bots

	if saved := ui.LoadSettings(); saved.Repos != nil || saved.Author != "" || saved.Any || saved.Bots {
		if len(searchOpts.Repos) == 0 && len(saved.Repos) > 0 {
			searchOpts.Repos = saved.Repos
		}
		if saved.Author != "" {
			author = saved.Author
		}
		if saved.Any {
			isAny = true
			author = ""
		}
		if saved.Bots {
			isBots = true
		}
	}

	srv := ui.NewServer(interval, ghOpts, ui.SearchConfig{
		Repos:  searchOpts.Repos,
		All:    searchOpts.All,
		Org:    searchOpts.Org,
		Author: author,
		Any:    isAny,
		Bots:   isBots,
	})

	srv.RepoSearchFn = func() (github.PRSearchResults, error) {
		since, _ := parseSince("30d")
		results, _, err := github.SearchPRs(ghOpts, github.PRSearchOptions{
			All:   true,
			Org:   searchOpts.Org,
			State: "open",
			Since: since,
			Limit: 100,
		})
		return results, err
	}

	searchFn := func(since time.Time) (github.PRSearchResults, *github.RateLimit, error) {
		cfg := srv.GetConfig()
		so := searchOpts
		if !since.IsZero() {
			so.Since = since
		}
		if len(cfg.Repos) > 0 {
			so.Repos = cfg.Repos
			so.All = false
		}
		if cfg.Any {
			so.Author = ""
		} else if cfg.Author != "" {
			so.Author = cfg.Author
		}
		so.ShowAuthor = so.Author != "@me"
		results, rl, err := github.SearchPRs(ghOpts, so)
		if err != nil {
			return nil, rl, err
		}
		if !cfg.Bots {
			results = filterByBot(results, false)
		}
		return results, rl, nil
	}

	poller := ui.NewPoller(srv, searchFn, interval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller.Start(ctx)

	var dashboardURL string
	if opts.UI {
		addr := "localhost:9092"
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to start PR UI server on %s: %w", addr, err)
		}
		dashboardURL = fmt.Sprintf("http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)

		go http.Serve(listener, srv.Handler()) //nolint:errcheck

		logger.Infof("PR Dashboard at %s", dashboardURL)
		openBrowser(dashboardURL)
	}

	if opts.MenuBar {
		mb := menubar.New(srv)
		mb.DashboardURL = dashboardURL
		mb.Run() // blocks on main thread (macOS Cocoa)
		return nil
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	return nil
}

func isBot(author string) bool {
	return strings.HasSuffix(author, "[bot]") || strings.HasSuffix(author, "bot")
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
