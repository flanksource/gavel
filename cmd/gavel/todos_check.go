package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

func runTodosList(opts TodosListOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := opts.Dir
	if dir == "" {
		dir = filepath.Join(workDir, ".todos")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf(".todos directory not found: %s", dir)
	}

	filters := todos.DiscoveryFilters{}
	if opts.Status != "" {
		filters.IncludeStatuses = []types.Status{types.Status(opts.Status)}
	}

	todoList, err := todos.DiscoverTODOs(dir, filters)
	if err != nil {
		return nil, err
	}

	if opts.GroupBy != "" && opts.GroupBy != todos.GroupByNone {
		groups := todos.GroupTODOs(todoList, opts.GroupBy)
		return todos.FlattenGrouped(groups), nil
	}

	return todoList, nil
}

func runTodosGet(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if todosDir == "" {
		todosDir = filepath.Join(workDir, ".todos")
	}

	todoPath := args[0]
	if !filepath.IsAbs(todoPath) && !strings.Contains(todoPath, string(filepath.Separator)) {
		todoPath = filepath.Join(todosDir, todoPath)
	}

	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		return fmt.Errorf("TODO file not found: %s", todoPath)
	}

	todo, err := todos.ParseTODO(todoPath)
	if err != nil {
		return fmt.Errorf("failed to parse TODO: %w", err)
	}

	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}

func runTodosCheck(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if todosDir == "" {
		todosDir = filepath.Join(workDir, ".todos")
	}

	var todoList []*types.TODO

	hasFilePaths := len(args) > 0 && (filepath.IsAbs(args[0]) || strings.Contains(args[0], string(filepath.Separator)))

	if hasFilePaths {
		logger.Infof("Checking specific TODO files...")
		for _, arg := range args {
			todoPath := arg
			if !filepath.IsAbs(todoPath) && !strings.Contains(todoPath, string(filepath.Separator)) {
				todoPath = filepath.Join(todosDir, todoPath)
			}
			if _, err := os.Stat(todoPath); os.IsNotExist(err) {
				return fmt.Errorf("TODO file not found: %s", todoPath)
			}
			todo, err := todos.ParseTODO(todoPath)
			if err != nil {
				return fmt.Errorf("failed to parse TODO %s: %w", todoPath, err)
			}
			todoList = append(todoList, todo)
		}
	} else {
		if _, err := os.Stat(todosDir); os.IsNotExist(err) {
			return fmt.Errorf(".todos directory not found: %s", todosDir)
		}
		logger.Infof("Discovering TODOs in: %s", todosDir)

		filters := todos.DiscoveryFilters{}
		if filterStatus != "" {
			filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
		}

		todoList, err = todos.DiscoverTODOs(todosDir, filters)
		if err != nil {
			return fmt.Errorf("failed to discover TODOs: %w", err)
		}

		if len(args) > 0 {
			var filtered []*types.TODO
			for _, todo := range todoList {
				for _, arg := range args {
					if strings.EqualFold(todo.Filename(), arg) {
						filtered = append(filtered, todo)
						break
					}
				}
			}
			todoList = filtered
		}
	}

	if len(todoList) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}

	logger.Infof("Found %d TODOs to check", len(todoList))

	checkOpts := todos.CheckOptions{
		WorkDir: workDir,
		Timeout: checkTimeout,
		Logger:  logger.StandardLogger(),
	}

	ctx := context.Background()
	results, err := todos.CheckTODOs(ctx, todoList, checkOpts)
	if err != nil {
		return fmt.Errorf("failed to check TODOs: %w", err)
	}

	fmt.Println()
	fmt.Println(clicky.Text("Check Results:", "text-blue-600 font-bold").ANSI())
	for _, result := range results {
		fmt.Println(result.Pretty().ANSI())
	}

	passed := 0
	failed := 0
	for _, result := range results {
		if result.AllPassed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Println()
	if failed == 0 {
		fmt.Println(clicky.Text(fmt.Sprintf("Summary: %d passed, %d failed", passed, failed), "font-bold text-green-600").ANSI())
	} else {
		fmt.Println(clicky.Text(fmt.Sprintf("Summary: %d passed, %d failed", passed, failed), "font-bold text-red-600").ANSI())
	}

	if failed > 0 {
		return fmt.Errorf("%d TODOs failed verification", failed)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(todosCmd)
	todosCmd.AddCommand(todosRunCmd)
	clicky.AddCommand(todosCmd, TodosListOptions{}, runTodosList)
	todosCmd.AddCommand(todosGetCmd)
	todosCmd.AddCommand(todosCheckCmd)

	todosRunCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosRunCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Maximum retry attempts for failed TODOs")
	todosRunCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status (pending, in_progress, failed)")
	todosRunCmd.Flags().Float64Var(&maxBudget, "max-budget", 0, "Maximum budget in USD")
	todosRunCmd.Flags().IntVar(&maxTurns, "max-turns", 0, "Maximum conversation turns")
	todosRunCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactively select TODOs to run")
	todosRunCmd.Flags().StringVar(&groupBy, "group-by", "", "Group TODOs by: file, directory, or none")
	todosRunCmd.Flags().BoolVar(&dirty, "dirty", false, "Skip git stash/checkout, run on dirty working tree")
	todosRunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands and prompts without executing")

	todosGetCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")

	todosCheckCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosCheckCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosCheckCmd.Flags().DurationVar(&checkTimeout, "timeout", 2*time.Minute, "Test execution timeout")
}
