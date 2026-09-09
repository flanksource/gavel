package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
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

	// The check is the lifecycle's verify step, dispatched by the host that owns
	// the workspace's lifecycle: `.gavel.yaml` ai:/todos.verify reach `todos
	// check` exactly as they reach the dashboard's verify action, and the flag is
	// the request layer on top.
	host, err := lifecycle.NewHost(provider, workDir, lifecycle.HostCLI)
	if err != nil {
		return err
	}
	var request api.Spec
	if checkTimeout > 0 {
		request.Budget.Timeout = checkTimeout.String()
	}
	concurrency, err := resolveCheckConcurrency(workDir)
	if err != nil {
		return err
	}
	checkOpts := todos.CheckOptions{
		Runner:      host,
		Request:     request,
		Logger:      logger.StandardLogger(),
		Concurrency: concurrency,
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

	todosCheckCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosCheckCmd.Flags().DurationVar(&checkTimeout, "timeout", 0, "Test execution timeout (empty: .gavel.yaml todos.timeout, else 30m)")
	todosCheckCmd.Flags().IntVar(&checkConcurrency, "concurrency", 0,
		"How many TODOs to check at once (0: .gavel.yaml todos.checkConcurrency, else 4). Each check runs that TODO's fixture")
}

func newTodosProvider(workDir string) (todos.Provider, error) {
	return openRuntimeTodosProvider(context.Background(), workDir)
}

// resolveCheckConcurrency reads the configured definition-of-done concurrency,
// with the --concurrency flag on top. An unreadable .gavel.yaml is the check's
// own error: every check resolves its verify chain from the same file, so
// running on a default here would only defer the same failure to each todo.
func resolveCheckConcurrency(workDir string) (int, error) {
	if checkConcurrency > 0 {
		return checkConcurrency, nil
	}
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return 0, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	return cfg.Todos.CheckConcurrency, nil
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
				var listErr error
				if listed, listErr = provider.List(ctx, todos.DiscoveryFilters{}); listErr != nil {
					return nil, fmt.Errorf("%w (and listing todos to match %q by title failed: %v)", err, ref, listErr)
				}
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
