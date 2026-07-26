package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var loadTodoProjects = ui.LoadProjects

// openRuntimeTodosProvider is the only production provider constructor used by
// TODO commands. It is a variable solely so command tests can exercise routing
// without requiring a process-owned PostgreSQL instance.
var openRuntimeTodosProvider = func(ctx context.Context, workDir string) (todos.Provider, error) {
	project, err := ui.ProjectForDir(workDir)
	if err != nil {
		return nil, err
	}
	return todoruntime.Open(ctx, project.WorkspaceOptions())
}

func runTodosList(opts TodosListOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	var since time.Time
	if opts.Since != "" {
		since, err = parseSince(opts.Since)
		if err != nil {
			return nil, err
		}
	}

	filters := todos.DiscoveryFilters{}
	if opts.Status != "" {
		filters.IncludeStatuses = []types.Status{types.Status(opts.Status)}
	} else if !opts.Done {
		filters.ExcludeStatuses = []types.Status{types.StatusVerified, types.StatusCompleted}
	}

	ctx := context.Background()
	var todoList types.TODOS
	if opts.All {
		projects, loadErr := loadTodoProjects()
		if loadErr != nil {
			return nil, loadErr
		}
		todoList, err = listAllProjectTodos(ctx, projects, filters)
		if err != nil {
			return nil, err
		}
	} else {
		provider, err := newTodosProvider(workDir)
		if err != nil {
			return nil, err
		}
		todoList, err = provider.List(ctx, filters)
		if err != nil {
			return nil, err
		}
	}
	if !since.IsZero() {
		todoList = filterTODOsSince(todoList, since)
	}

	if opts.GroupBy != "" && opts.GroupBy != todos.GroupByNone {
		groups := todos.GroupTODOsWithWorkDir(todoList, opts.GroupBy, workDir)
		return todos.FlattenGrouped(groups), nil
	}

	return todoList, nil
}

func filterTODOsSince(todoList types.TODOS, since time.Time) types.TODOS {
	filtered := make(types.TODOS, 0, len(todoList))
	for _, todo := range todoList {
		if todo == nil {
			continue
		}
		latest := todo.Created
		if todo.LastRun != nil && (latest == nil || todo.LastRun.After(*latest)) {
			latest = todo.LastRun
		}
		if latest != nil && !latest.Before(since) {
			filtered = append(filtered, todo)
		}
	}
	return filtered
}

// listAllProjectTodos aggregates TODOs from every registered workspace using
// the native PostgreSQL runtime. Duplicate project entries that resolve to the
// same directory are queried once; stored legacy provider preferences are
// intentionally ignored because they can no longer select runtime storage.
// Every database open/list failure is returned: treating an unavailable
// database as an empty workspace would be a false successful result.
func listAllProjectTodos(ctx context.Context, projects []ui.Project, filters todos.DiscoveryFilters) (types.TODOS, error) {
	seen := map[string]struct{}{}
	var todoList types.TODOS
	var failures []error
	for _, project := range projects {
		dir := project.ResolvedDir()
		if strings.TrimSpace(dir) == "" {
			logger.Warnf("list todos for project %q: workspace directory is empty", project.Name)
			continue
		}
		key := filepath.Clean(dir)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		provider, err := openRuntimeTodosProvider(ctx, dir)
		if err != nil {
			failures = append(failures, fmt.Errorf("open native TODO workspace for project %q (%s): %w", project.Name, dir, err))
			continue
		}
		items, err := provider.List(ctx, filters)
		if err != nil {
			failures = append(failures, fmt.Errorf("list native TODOs for project %q (%s): %w", project.Name, dir, err))
			continue
		}
		for _, todo := range items {
			if todo != nil {
				todo.Workspace = project.Name
				if todo.CWD == "" {
					todo.CWD = dir
				}
			}
		}
		todoList = append(todoList, items...)
	}
	todoList.Sort()
	return todoList, errors.Join(failures...)
}

func runTodosGet(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	todo, err := provider.Get(context.Background(), args[0])
	if err != nil {
		return err
	}

	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}

func runTodosCheck(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}

	logger.Infof("Discovering TODOs from PostgreSQL")
	filters := todos.DiscoveryFilters{}
	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}
	todoList, err := resolveRequestedTODOs(context.Background(), provider, args, filters)
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
	}

	if len(todoList) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}

	logger.Infof("Found %d TODOs to check", len(todoList))

	checkOpts := todos.CheckOptions{
		WorkDir:  workDir,
		Timeout:  checkTimeout,
		Logger:   logger.StandardLogger(),
		Provider: provider,
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

	todosRunCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status (pending, in_progress, failed)")
	todosRunCmd.Flags().Float64Var(&maxBudget, "max-budget", 0, "Maximum budget in USD")
	todosRunCmd.Flags().IntVar(&maxTurns, "max-turns", 0, "Maximum conversation turns")
	todosRunCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactively select TODOs to run")
	todosRunCmd.Flags().StringVar(&groupBy, "group-by", "", "Group TODOs by: file, directory, repo, all, or none")
	todosRunCmd.Flags().StringVar(&todosMode, "mode", "run", "Todo operation: run (implement) or plan (propose a reviewable plan)")
	todosRunCmd.Flags().StringVar(&todosDriver, "driver", "", "Execution mechanism: cmux, cli, sdk, or api (default: cmux). The coding agent is derived from --model")
	todosRunCmd.Flags().StringVar(&todoModel, "model", "", "LLM model override for TODO execution (empty: the mode's .prompt frontmatter default)")
	todosRunCmd.Flags().StringVar(&todoEffort, "effort", "medium", "Reasoning effort directive: low, medium, high, or xhigh")
	todosRunCmd.Flags().BoolVar(&resumeSession, "resume", false, "Resume the TODO's prior agent session instead of starting a fresh one")
	todosRunCmd.Flags().BoolVar(&dirty, "dirty", false, "Skip git stash/checkout, run on dirty working tree")
	todosRunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands and prompts without executing")
	todosRunCmd.Flags().BoolVar(&commitAfter, "commit", true, "Run the equivalent of `gavel commit` after each TODO's agent completes (use --commit=false to disable)")
	todosRunCmd.Flags().BoolVar(&checkAfter, "check", false, "After each TODO's agent completes, run the configured `checks` test/lint suite and feed failures back to the agent until they pass (see .gavel.yaml checks / frontmatter)")

	todosCheckCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosCheckCmd.Flags().DurationVar(&checkTimeout, "timeout", 2*time.Minute, "Test execution timeout")
}

func newTodosProvider(workDir string) (todos.Provider, error) {
	return openRuntimeTodosProvider(context.Background(), workDir)
}

// resolveRequestedTODOs preserves direct native reference semantics for UUIDs,
// safe short IDs, and imported aliases. Listing first would discard legacy
// aliases because list rows expose the canonical native UUID.
func resolveRequestedTODOs(ctx context.Context, provider todos.Provider, args []string, filters todos.DiscoveryFilters) (types.TODOS, error) {
	if len(args) == 0 {
		return provider.List(ctx, filters)
	}

	resolved := make(types.TODOS, 0, len(args))
	seen := map[string]struct{}{}
	var listed types.TODOS
	for _, ref := range args {
		todo, err := provider.Get(ctx, ref)
		if err != nil {
			// Preserve exact-title CLI compatibility without replacing the native
			// repository's prefix length and ambiguity checks.
			if listed == nil {
				listed, _ = provider.List(ctx, todos.DiscoveryFilters{})
			}
			var titleMatches types.TODOS
			for _, candidate := range listed {
				if candidate != nil && strings.EqualFold(candidate.Title, ref) {
					titleMatches = append(titleMatches, candidate)
				}
			}
			if len(titleMatches) != 1 {
				return nil, err
			}
			todo = titleMatches[0]
		}
		if !filters.Matches(todo) {
			continue
		}
		key := todo.ID
		if key == "" {
			key = todos.TODOReference(todo)
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		resolved = append(resolved, todo)
	}
	resolved.Sort()
	return resolved, nil
}
