package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	fixSyncDir string
	fixRepo    string
)

var prFixCmd = &cobra.Command{
	Use:          "fix [pr-number]",
	Short:        "Sync TODOs from PR failures and interactively select which to fix",
	SilenceUsage: true,
	Args:         cobra.MaximumNArgs(1),
	RunE:         runPRFix,
}

func runPRFix(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	var ghOpts github.Options
	if fixRepo != "" {
		ghOpts.Repo = fixRepo
	} else {
		ghOpts.WorkDir = workDir
	}

	var prNumber int
	if len(args) > 0 {
		prNumber, err = strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PR number: %w", err)
		}
	}

	syncDir := fixSyncDir
	if syncDir == "" {
		syncDir = filepath.Join(workDir, ".todos")
	}

	// Step 1: Watch PR and sync TODOs
	logger.Infof("Fetching PR status and syncing TODOs...")
	result, _ := prwatch.Run(prwatch.WatchOptions{
		Options:  ghOpts,
		PRNumber: prNumber,
		TailLogs: watchTailLogs,
	})

	if result == nil {
		return fmt.Errorf("failed to fetch PR information")
	}

	clicky.MustPrint(result, clicky.FormatOptions{})

	if err := prwatch.SyncTodos(result, syncDir); err != nil {
		logger.Warnf("failed to sync todos: %v", err)
	}
	if err := prwatch.SyncCommentTodos(result.Comments, result.PR, syncDir); err != nil {
		logger.Warnf("failed to sync comment todos: %v", err)
	}

	// Step 2: Discover and present TODOs
	if _, err := os.Stat(syncDir); os.IsNotExist(err) {
		logger.Infof("No TODOs synced")
		return nil
	}

	todoList, err := todos.DiscoverTODOs(syncDir, todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
	}

	if len(todoList) == 0 {
		logger.Infof("No TODOs to fix")
		return nil
	}

	// Step 3: Interactive selection
	selected, err := selectTODOs(todoList, "Select TODOs to fix:")
	if err != nil {
		return err
	}
	if selected == nil {
		logger.Infof("No TODOs selected")
		return nil
	}

	// Step 4: Execute selected TODOs
	logger.Infof("Executing %d TODOs...", len(selected))
	interaction := newInteraction()

	groups := todos.GroupTODOs(selected, groupBy)
	fmt.Println(clicky.MustFormat(todos.FlattenGrouped(groups)))
	fmt.Println()

	if groupBy != "" && groupBy != todos.GroupByNone {
		return executeGroups(workDir, groups, interaction)
	}
	return executeSingleTODOs(workDir, selected, interaction)
}

func init() {
	prCmd.AddCommand(prFixCmd)
	prFixCmd.Flags().StringVarP(&fixRepo, "repo", "R", "", "GitHub repository (owner/repo)")
	prFixCmd.Flags().StringVar(&fixSyncDir, "dir", "", "TODOs directory (default: .todos)")
	prFixCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum retry attempts")
	prFixCmd.Flags().Float64Var(&maxBudget, "max-budget", 0, "Maximum budget in USD")
	prFixCmd.Flags().IntVar(&maxTurns, "max-turns", 0, "Maximum conversation turns")
	prFixCmd.Flags().StringVar(&groupBy, "group-by", "", "Group TODOs by: file, directory, all, or none")
	prFixCmd.Flags().BoolVar(&dirty, "dirty", false, "Skip git stash/checkout")
	prFixCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands without executing")
}
