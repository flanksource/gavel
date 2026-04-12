package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	"github.com/spf13/cobra"
)

var (
	statusFollow     bool
	statusInterval   time.Duration
	statusFetchLogs  bool
	statusTailLogs   int
	statusRepo       string
	statusSyncTodos  string
	statusOpts       clicky.FormatOptions
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Pull request commands",
}

var prStatusCmd = &cobra.Command{
	Use:          "status [repo] [pr-number]",
	Short:        "Show GitHub Actions status for a PR",
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(2),
	RunE:         runPRStatus,
}

func runPRStatus(cmd *cobra.Command, args []string) error {
	repo, prNumber, err := parseStatusArgs(args)
	if err != nil {
		return err
	}

	var ghOpts github.Options
	if repo != "" {
		ghOpts.Repo = repo
	} else if statusRepo != "" {
		ghOpts.Repo = statusRepo
	} else {
		workDir, err := getWorkingDir()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	if prNumber == 0 {
		prNumber = resolveOrFallbackPR(ghOpts)
	}

	opts := prwatch.WatchOptions{
		Options:  ghOpts,
		PRNumber: prNumber,
		Interval: statusInterval,
		Follow:   statusFollow,
		Logs:     statusFetchLogs,
		TailLogs: statusTailLogs,
	}

	result, code := prwatch.Run(opts)
	if result != nil {
		clicky.MustPrint(result, statusOpts)
		if statusSyncTodos != "" {
			if err := prwatch.SyncTodos(result, statusSyncTodos); err != nil {
				logger.Warnf("failed to sync todos: %v", err)
			}
			if err := prwatch.SyncCommentTodos(result.Comments, result.PR, statusSyncTodos); err != nil {
				logger.Warnf("failed to sync comment todos: %v", err)
			}
		}
	}
	exitCode = code
	return nil
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
	prCmd.AddCommand(prStatusCmd)
	prStatusCmd.Flags().StringVarP(&statusRepo, "repo", "R", "", "GitHub repository (owner/repo)")
	prStatusCmd.Flags().BoolVar(&statusFollow, "follow", false, "Keep watching until all checks complete")
	prStatusCmd.Flags().DurationVar(&statusInterval, "interval", 30*time.Second, "Poll interval")
	prStatusCmd.Flags().BoolVar(&statusFetchLogs, "logs", false, "Fetch and include failed job logs (uses extra GitHub API quota)")
	prStatusCmd.Flags().IntVar(&statusTailLogs, "tail-logs", 100, "Number of failed log lines to show per step (only applies with --logs)")
	prStatusCmd.Flags().StringVar(&statusSyncTodos, "sync-todos", "", "Sync TODO files for failed jobs to directory")
	prStatusCmd.Flag("sync-todos").NoOptDefVal = ".todos"
	formatters.BindPFlags(prStatusCmd.Flags(), &statusOpts)
}
