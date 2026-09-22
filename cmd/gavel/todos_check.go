package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
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
	"github.com/google/uuid"
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

func runTodosGet(opts TodosGetOptions) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	todo, err := provider.Get(context.Background(), ref)
	if err != nil {
		return err
	}

	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}

func runTodosCheck(opts TodosCheckOptions) error {
	ids, err := opts.Many()
	if err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}

	logger.Infof("Discovering TODOs from PostgreSQL")
	todoList, err := resolveRequestedTODOs(context.Background(), provider, workDir, ids, todos.DiscoveryFilters{})
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
	if opts.Timeout > 0 {
		request.Budget.Timeout = opts.Timeout.String()
	}
	concurrency, err := resolveCheckConcurrency(workDir, opts.Concurrency)
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
	clicky.AddCommand(todosCmd, TodosListOptions{}, runTodosList)
	todosGetCmd = clicky.AddNamedCommand("get", todosCmd, TodosGetOptions{}, func(opts TodosGetOptions) (any, error) {
		return nil, runTodosGet(opts)
	})
	todosGetCmd.Short = "Display detailed information about a PostgreSQL-backed TODO"
	todosGetCmd.Use = "get <id>"
	todosCheckCmd = clicky.AddNamedCommand("check", todosCmd, TodosCheckOptions{}, func(opts TodosCheckOptions) (any, error) {
		return nil, runTodosCheck(opts)
	})
	todosCheckCmd.Short = "Run TODOs' fixture-backed definitions of done"
	todosCheckCmd.Use = "check <id>..."
}

func newTodosProvider(workDir string) (todos.Provider, error) {
	return openRuntimeTodosProvider(context.Background(), workDir)
}

// resolveCheckConcurrency reads the configured definition-of-done concurrency,
// with the --concurrency flag on top. An unreadable .gavel.yaml is the check's
// own error: every check resolves its verify chain from the same file, so
// running on a default here would only defer the same failure to each todo.
func resolveCheckConcurrency(workDir string, override int) (int, error) {
	if override > 0 {
		return override, nil
	}
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return 0, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	return cfg.Todos.CheckConcurrency, nil
}

// todoTarget is a resolved TODO together with the workspace that owns it: the
// provider every subsequent write must go through, and the directory a run
// executes in. They differ from the caller's own only for a TODO named by a UUID
// from outside its workspace.
type todoTarget struct {
	Provider todos.Provider
	WorkDir  string
	Todo     *types.TODO
}

// resolveRequestedTargets resolves references to TODOs and the workspace each one
// belongs to.
//
// A UUID names exactly one issue in the database, so it is resolved globally when
// the caller's workspace does not hold it: requiring the right working directory
// for an id that is already unique is a filter with nothing to disambiguate, and
// it is how `todos run <uuid>` failed with "native todo record not found" for an
// issue that existed the whole time. Short ids and titles stay workspace-scoped —
// those genuinely can collide between projects.
//
// A TODO found this way is re-read through its owning workspace's provider, as
// runtime.GlobalGet's contract requires: the global read reconciles nothing, and
// acting through the caller's provider would write against the wrong working
// directory and the wrong .gavel.yaml.
func resolveRequestedTargets(ctx context.Context, provider todos.Provider, workDir string, args []string, filters todos.DiscoveryFilters) ([]todoTarget, error) {
	local := func(todo *types.TODO) todoTarget {
		return todoTarget{Provider: provider, WorkDir: workDir, Todo: todo}
	}
	if len(args) == 0 {
		listed, err := provider.List(ctx, filters)
		if err != nil {
			return nil, err
		}
		targets := make([]todoTarget, 0, len(listed))
		for _, todo := range listed {
			targets = append(targets, local(todo))
		}
		return targets, nil
	}

	resolved := make([]todoTarget, 0, len(args))
	seen := map[string]struct{}{}
	var listed types.TODOS

	// recover handles a reference this workspace's Get could not resolve: a UUID
	// owned elsewhere, else an exact title. getErr is preserved as the reported
	// failure, because it describes the reference the caller actually typed.
	recover := func(ref string, getErr error) (todoTarget, error) {
		adopted, found, err := adoptGlobalUUID(ctx, provider, workDir, ref)
		if err != nil {
			return todoTarget{}, err
		}
		if found {
			return adopted, nil
		}
		// Preserve exact-title CLI compatibility without replacing the native
		// repository's prefix length and ambiguity checks.
		if listed == nil {
			var listErr error
			if listed, listErr = provider.List(ctx, todos.DiscoveryFilters{}); listErr != nil {
				return todoTarget{}, fmt.Errorf("%w (and listing todos to match %q by title failed: %v)", getErr, ref, listErr)
			}
		}
		var titleMatches types.TODOS
		for _, candidate := range listed {
			if candidate != nil && strings.EqualFold(candidate.Title, ref) {
				titleMatches = append(titleMatches, candidate)
			}
		}
		if len(titleMatches) != 1 {
			// The provider's error describes the reference but never quotes it, so a
			// batch of refs reported "short issue reference must contain at least 8
			// characters" without saying which argument it meant.
			return todoTarget{}, fmt.Errorf("resolve todo %q: %w", ref, getErr)
		}
		return local(titleMatches[0]), nil
	}

	for _, ref := range args {
		target := local(nil)
		todo, err := provider.Get(ctx, ref)
		if err != nil {
			recovered, recoverErr := recover(ref, err)
			if recoverErr != nil {
				return nil, recoverErr
			}
			target, todo = recovered, recovered.Todo
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
		target.Todo = todo
		resolved = append(resolved, target)
	}
	sort.SliceStable(resolved, func(i, j int) bool {
		return todoSortKey(resolved[i].Todo) < todoSortKey(resolved[j].Todo)
	})
	return resolved, nil
}

// adoptGlobalUUID resolves a UUID that the caller's workspace does not hold, and
// returns it bound to the workspace that does.
//
// Only a UUID qualifies: it identifies one issue in the whole database, so there
// is nothing for a workspace filter to disambiguate. It reports found=false for
// anything else, and for a UUID nothing owns, so the caller falls through to its
// own error — a global miss must not mask the workspace-scoped message.
func adoptGlobalUUID(ctx context.Context, provider todos.Provider, workDir, ref string) (todoTarget, bool, error) {
	ref = strings.TrimSpace(ref)
	if _, err := uuid.Parse(ref); err != nil {
		return todoTarget{}, false, nil
	}
	global, ok := provider.(todos.GlobalReferenceProvider)
	if !ok {
		return todoTarget{}, false, nil
	}
	todo, err := global.GetGlobal(ctx, ref)
	if err != nil || todo == nil {
		return todoTarget{}, false, nil
	}
	owner := strings.TrimSpace(todo.CWD)
	if owner == "" || !filepath.IsAbs(owner) || filepath.Clean(owner) == filepath.Clean(workDir) {
		return todoTarget{Provider: provider, WorkDir: workDir, Todo: todo}, true, nil
	}
	// The global read reconciles nothing and carries the caller's working
	// directory nowhere, so the owning workspace re-reads its own issue.
	owned, err := openRuntimeTodosProvider(ctx, owner)
	if err != nil {
		return todoTarget{}, false, fmt.Errorf("todo %s belongs to workspace %s, which could not be opened: %w",
			ref, owner, err)
	}
	reread, err := owned.Get(ctx, todo.ID)
	if err != nil {
		return todoTarget{}, false, fmt.Errorf("todo %s belongs to workspace %s, which could not read it: %w",
			ref, owner, err)
	}
	logger.Infof("TODO %s belongs to %s; running there", reread.ShortID, owner)
	return todoTarget{Provider: owned, WorkDir: owner, Todo: reread}, true, nil
}

// todoSortKey mirrors types.TODOS.Sort's ordering — priority, then name — for a
// slice that carries a provider alongside each TODO.
func todoSortKey(todo *types.TODO) string {
	if todo == nil {
		return ""
	}
	order := map[types.Priority]int{types.PriorityHigh: 0, types.PriorityMedium: 1, types.PriorityLow: 2}
	rank, ok := order[todo.Priority]
	if !ok {
		rank = 9
	}
	return fmt.Sprintf("%d:%s", rank, strings.ToLower(todos.TODOReference(todo)))
}

// resolveRequestedTODOs resolves references for a command that acts through the
// single provider it opened for the current workspace.
//
// A UUID naming a TODO in another workspace resolves — that is the point — but it
// is refused here rather than acted on through the wrong provider, which would
// write against this directory's .gavel.yaml and working tree. Commands that
// support running elsewhere use resolveRequestedTargets and honour each target's
// own workspace.
func resolveRequestedTODOs(ctx context.Context, provider todos.Provider, workDir string, args []string, filters todos.DiscoveryFilters) (types.TODOS, error) {
	targets, err := resolveRequestedTargets(ctx, provider, workDir, args, filters)
	if err != nil {
		return nil, err
	}
	resolved := make(types.TODOS, 0, len(targets))
	for _, target := range targets {
		if target.WorkDir != workDir {
			return nil, fmt.Errorf("todo %s belongs to workspace %s; run this command from there (--cwd %s)",
				todos.TODOReference(target.Todo), target.WorkDir, target.WorkDir)
		}
		resolved = append(resolved, target.Todo)
	}
	resolved.Sort()
	return resolved, nil
}
