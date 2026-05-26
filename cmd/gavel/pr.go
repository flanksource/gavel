package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	"github.com/spf13/cobra"
)

// statusDefaultTailLogs is the default --tail-logs value, re-used by other
// pr-subcommands (e.g. `pr fix`) that piggyback on prwatch.
const statusDefaultTailLogs = 100

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Pull request commands",
}

type PRStatusOptions struct {
	Repo      string          `flag:"repo" short:"R" help:"GitHub repository (owner/repo)"`
	Follow    bool            `flag:"follow" help:"Keep watching until all checks complete"`
	Interval  string          `flag:"interval" help:"Poll interval (e.g. 30s, 1m)" default:"30s"`
	Logs      bool            `flag:"logs" help:"Fetch and include failed job logs (uses extra GitHub API quota)"`
	TailLogs  int             `flag:"tail-logs" help:"Number of failed log lines to show per step (only applies with --logs)" default:"100"`
	SyncTodos string          `flag:"sync-todos" help:"Sync TODO files for failed jobs to directory"`
	Args      []string        `args:"true"`
	Context   context.Context `json:"-"`

	AIFix         bool `flag:"ai-fix" help:"Feed the rendered PR status into the AI configured by 'captain configure' to fix failing checks/comments"`
	AIFixMaxIters int  `flag:"ai-fix-max-iterations" help:"Max AI iterations driven by the status prompt" default:"1"`

	// Embedded: contributes --model, --backend, --api-key, --no-cache,
	// --budget, --debug, --max-tokens, --temperature, --permission-mode,
	// --edit, --allowed-tools, --disallowed-tools, --mcp, --hooks,
	// --skills, --skill-dir, --user, --project, --memory, --bare.
	// Defaults overlay from ~/.captain.yaml via captain configure.
	captaincli.AIRuntimeOptions
}

func (o PRStatusOptions) Help() string {
	return `Show GitHub Actions status for a PR.

Examples:
  gavel pr status                              # current branch's PR
  gavel pr status 123                          # PR #123 in this repo
  gavel pr status owner/repo 123               # PR #123 in another repo
  gavel pr status https://github.com/o/r/pull/1
  gavel pr status --follow                     # block until checks complete
  gavel pr status --ai-fix                     # feed status into the AI to fix failures`
}

func runPRStatus(opts PRStatusOptions) (any, error) {
	repo, prNumber, err := parseStatusArgs(opts.Args)
	if err != nil {
		return nil, err
	}

	var ghOpts github.Options
	switch {
	case repo != "":
		ghOpts.Repo = repo
	case opts.Repo != "":
		ghOpts.Repo = opts.Repo
	default:
		workDir, err := getWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	if prNumber == 0 {
		prNumber = resolveOrFallbackPR(ghOpts)
	}

	interval, err := time.ParseDuration(opts.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid --interval %q: %w", opts.Interval, err)
	}
	watchOpts := prwatch.WatchOptions{
		Options:  ghOpts,
		PRNumber: prNumber,
		Interval: interval,
		Follow:   opts.Follow,
		Logs:     opts.Logs,
		TailLogs: opts.TailLogs,
	}

	result, code := prwatch.Run(watchOpts)
	exitCode = code
	if result == nil {
		return nil, nil
	}
	if opts.SyncTodos != "" {
		if err := prwatch.SyncTodos(result, opts.SyncTodos); err != nil {
			logger.Warnf("failed to sync todos: %v", err)
		}
		if err := prwatch.SyncCommentTodos(result.Comments, result.PR, opts.SyncTodos); err != nil {
			logger.Warnf("failed to sync comment todos: %v", err)
		}
	}

	if opts.AIFix {
		// Print the status block first so the user sees the input the AI is
		// about to act on before live ai-fix events start streaming to
		// stderr. Returning nil suppresses clicky's trailing print of the
		// same data — the captain history rendered after the run is the
		// meaningful tail in this mode.
		clicky.MustPrint(result, clicky.FormatOptions{})

		ctx := opts.Context
		if ctx == nil {
			ctx = commonsContext.NewContext(context.Background())
		}
		if aiErr := runPRStatusAIFix(ctx, opts, result); aiErr != nil {
			return nil, fmt.Errorf("ai-fix: %w", aiErr)
		}
		return nil, nil
	}

	return result, nil
}

func parseStatusArgs(args []string) (repo string, prNumber int, err error) {
	switch len(args) {
	case 0:
		return "", 0, nil
	case 1:
		if n, err := strconv.Atoi(args[0]); err == nil {
			return "", n, nil
		}
		repo, pr, err := github.ParsePRURL(args[0])
		if err == nil {
			return repo, pr, nil
		}
		return "", 0, fmt.Errorf("expected PR number or URL, got %q", args[0])
	case 2:
		prNumber, err = strconv.Atoi(args[1])
		if err != nil {
			return "", 0, fmt.Errorf("invalid PR number %q: %w", args[1], err)
		}
		repo, err = resolveRepoArg(args[0])
		if err != nil {
			return "", 0, err
		}
		return repo, prNumber, nil
	}
	return "", 0, fmt.Errorf("too many arguments")
}

func resolveOrFallbackPR(ghOpts github.Options) int {
	_, err := github.FetchPR(ghOpts, 0)
	if err == nil {
		return 0
	}

	logger.Infof("No PR found for current branch, checking most recent PR...")
	results, _, searchErr := github.SearchPRs(ghOpts, github.PRSearchOptions{
		Author: "@me",
		State:  "all",
		Limit:  1,
	})
	if searchErr != nil || len(results) == 0 {
		return 0
	}

	logger.Infof("Using most recent PR #%d: %s", results[0].Number, results[0].Title)
	return results[0].Number
}

func init() {
	rootCmd.AddCommand(prCmd)
	statusCmd := clicky.AddNamedCommand("status", prCmd, PRStatusOptions{}, runPRStatus)
	if f := statusCmd.Flags().Lookup("sync-todos"); f != nil {
		f.NoOptDefVal = ".todos"
	}
}
