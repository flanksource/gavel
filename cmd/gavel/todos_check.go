package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var loadTodoProjects = ui.LoadProjects

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
		if opts.Dir != "" {
			return nil, fmt.Errorf("--dir cannot be used with --all")
		}
		todoList = listAllProjectTodos(ctx, loadTodoProjects(), filters)
	} else {
		provider, err := newTodosProvider(workDir, opts.Dir)
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
// the same per-project provider resolution as the dashboard. Duplicate project
// entries that resolve to the same directory and backend are queried once.
// One stale or unavailable workspace should not hide results from the others,
// so project-local failures are reported as warnings and skipped.
func listAllProjectTodos(ctx context.Context, projects []ui.Project, filters todos.DiscoveryFilters) types.TODOS {
	seen := map[string]struct{}{}
	var todoList types.TODOS
	for _, project := range projects {
		dir := project.ResolvedDir()
		if strings.TrimSpace(dir) == "" {
			logger.Warnf("list todos for project %q: workspace directory is empty", project.Name)
			continue
		}
		backend := strings.TrimSuffix(ui.TodoBackendLabel(dir, project.TodoProvider), " (auto)")
		key := filepath.Clean(dir) + "\x00" + backend
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		items, err := ui.ProviderForProject(project).List(ctx, filters)
		if err != nil {
			logger.Warnf("list todos for project %q (%s): %v", project.Name, dir, err)
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
	return todoList
}

func runTodosGet(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir, todosDir)
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

	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}

	var todoList []*types.TODO

	hasFilePaths := selectedTodosProvider() == todos.ProviderFiles && len(args) > 0 && (filepath.IsAbs(args[0]) || strings.Contains(args[0], string(filepath.Separator)))

	if hasFilePaths {
		logger.Infof("Checking specific TODO files...")
		for _, arg := range args {
			todo, err := provider.Get(context.Background(), arg)
			if err != nil {
				return err
			}
			todoList = append(todoList, todo)
		}
	} else {
		logger.Infof("Discovering TODOs using provider: %s", selectedTodosProvider())

		filters := todos.DiscoveryFilters{}
		if filterStatus != "" {
			filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
		}

		todoList, err = provider.List(context.Background(), filters)
		if err != nil {
			return fmt.Errorf("failed to discover TODOs: %w", err)
		}

		if len(args) > 0 {
			todoList = filterTODOsByArgs(todoList, args, workDir)
		}
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
	todosCmd.PersistentFlags().StringVar(&todosProvider, "provider", todos.ProviderGrite, "TODO provider: grite or todos")
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
	todosRunCmd.Flags().StringVar(&groupBy, "group-by", "", "Group TODOs by: file, directory, repo, all, or none")
	todosRunCmd.Flags().StringVar(&todosMode, "mode", "run", "Todo operation: run (implement), plan (propose a reviewable plan), or verify (score committed work)")
	todosRunCmd.Flags().StringVar(&todosDriver, "driver", "", "Agent driver: claude-cmux, claude-headless, codex-cmux, codex-headless (default: <agent>-cmux)")
	todosRunCmd.Flags().StringVar(&todoModel, "model", "", "LLM model override for TODO execution (empty: the mode's .prompt frontmatter default)")
	todosRunCmd.Flags().StringVar(&todoEffort, "effort", "medium", "Reasoning effort directive: low, medium, high, or xhigh")
	todosRunCmd.Flags().BoolVar(&resumeSession, "resume", false, "Resume the TODO's prior agent session instead of starting a fresh one")
	todosRunCmd.Flags().BoolVar(&dirty, "dirty", false, "Skip git stash/checkout, run on dirty working tree")
	todosRunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print commands and prompts without executing")
	todosRunCmd.Flags().BoolVar(&commitAfter, "commit", true, "Run the equivalent of `gavel commit` after each TODO's agent completes (use --commit=false to disable)")
	todosRunCmd.Flags().BoolVar(&checkAfter, "check", false, "After each TODO's agent completes, run the configured `checks` test/lint suite and feed failures back to the agent until they pass (see .gavel.yaml checks / frontmatter)")

	todosGetCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")

	todosCheckCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosCheckCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosCheckCmd.Flags().DurationVar(&checkTimeout, "timeout", 2*time.Minute, "Test execution timeout")
}

func selectedTodosProvider() string {
	if todosProvider == "" {
		return todos.ProviderGrite
	}
	return todosProvider
}

func newTodosProvider(workDir, dir string) (todos.Provider, error) {
	switch selectedTodosProvider() {
	case todos.ProviderGrite:
		if dir != "" {
			return nil, fmt.Errorf("--dir is only supported with --provider=todos")
		}
		return todos.ResolveGriteProvider(workDir, cache.Shared(), 0), nil
	case todos.ProviderFiles:
		return todos.NewFileProvider(workDir, dir), nil
	default:
		return nil, fmt.Errorf("unknown todos provider %q (expected grite or todos)", selectedTodosProvider())
	}
}

func filterTODOsByArgs(todoList types.TODOS, args []string, workDir string) types.TODOS {
	var filtered types.TODOS
	for _, todo := range todoList {
		for _, arg := range args {
			if todoMatchesArg(todo, arg, workDir) {
				filtered = append(filtered, todo)
				break
			}
		}
	}
	return filtered
}

func todoMatchesArg(todo *types.TODO, arg, workDir string) bool {
	if todo == nil {
		return false
	}
	if todo.ID != "" && (strings.EqualFold(todo.ID, arg) || strings.HasPrefix(todo.ID, arg)) {
		return true
	}
	if todo.Title != "" && strings.EqualFold(todo.Title, arg) {
		return true
	}
	if strings.EqualFold(todo.Filename(), arg) {
		return true
	}
	if todo.FilePath == "" || (!strings.Contains(arg, string(filepath.Separator)) && !strings.HasSuffix(arg, ".md")) {
		return false
	}
	absArg := arg
	if !filepath.IsAbs(arg) {
		absArg = filepath.Join(workDir, arg)
	}
	return filepath.Clean(absArg) == filepath.Clean(todo.FilePath)
}
