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
	watchFollow    bool
	watchInterval  time.Duration
	watchTailLogs  int
	watchRepo      string
	watchSyncTodos string
	watchOpts      clicky.FormatOptions
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Pull request commands",
}

var prWatchCmd = &cobra.Command{
	Use:          "watch [pr-number]",
	Short:        "Watch GitHub Actions status for a PR",
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runPRWatch,
}

func runPRWatch(cmd *cobra.Command, args []string) error {
	var ghOpts github.Options
	if watchRepo != "" {
		ghOpts.Repo = watchRepo
	} else {
		workDir, err := getWorkingDir()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		ghOpts.WorkDir = workDir
	}

	var prNumber int
	if len(args) > 0 {
		var err error
		prNumber, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PR number: %w", err)
		}
	}

	opts := prwatch.WatchOptions{
		Options:  ghOpts,
		PRNumber: prNumber,
		Interval: watchInterval,
		Follow:   watchFollow,
		TailLogs: watchTailLogs,
	}

	result, code := prwatch.Run(opts)
	if result != nil {
		clicky.MustPrint(result, watchOpts)
		if watchSyncTodos != "" {
			if err := prwatch.SyncTodos(result, watchSyncTodos); err != nil {
				logger.Warnf("failed to sync todos: %v", err)
			}
			if err := prwatch.SyncCommentTodos(result.Comments, result.PR, watchSyncTodos); err != nil {
				logger.Warnf("failed to sync comment todos: %v", err)
			}
		}
	}
	exitCode = code
	return nil
}

func init() {
	rootCmd.AddCommand(prCmd)
	prCmd.AddCommand(prWatchCmd)
	prWatchCmd.Flags().StringVarP(&watchRepo, "repo", "R", "", "GitHub repository (owner/repo)")
	prWatchCmd.Flags().BoolVar(&watchFollow, "follow", false, "Keep watching until all checks complete")
	prWatchCmd.Flags().DurationVar(&watchInterval, "interval", 30*time.Second, "Poll interval")
	prWatchCmd.Flags().IntVar(&watchTailLogs, "tail-logs", 100, "Number of failed log lines to show per step")
	prWatchCmd.Flags().StringVar(&watchSyncTodos, "sync-todos", "", "Sync TODO files for failed jobs to directory")
	prWatchCmd.Flag("sync-todos").NoOptDefVal = ".todos"
	formatters.BindPFlags(prWatchCmd.Flags(), &watchOpts)
}
