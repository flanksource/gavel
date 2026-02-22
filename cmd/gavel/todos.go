package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/cmd/gavel/choose"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todosDir     string
	maxRetries   int
	filterStatus string
	checkTimeout time.Duration
	maxBudget    float64
	maxTurns     int
	interactive  bool
	groupBy      string
	dirty        bool
)

var todosCmd = &cobra.Command{
	Use:          "todos",
	SilenceUsage: true,
	Short:        "Automated TODO execution system with Claude Code integration",
}

var todosRunCmd = &cobra.Command{
	Use:          "run [todo-titles...]",
	SilenceUsage: true,
	Short:        "Execute TODOs using Claude Code with automated verification",
	RunE:         runTodosRun,
}

type TodosListOptions struct {
	Dir     string `json:"dir" flag:"dir" help:"TODOs directory (default: .todos)"`
	Status  string `json:"status" flag:"status" help:"Filter TODOs by status"`
	GroupBy string `json:"group-by" flag:"group-by" help:"Group TODOs by: file, directory, or none"`
}

func (opts TodosListOptions) GetName() string { return "list" }

var todosGetCmd = &cobra.Command{
	Use:          "get <todo-file>",
	SilenceUsage: true,
	Short:        "Display detailed information about a TODO",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosGet,
}

var todosCheckCmd = &cobra.Command{
	Use:          "check [todo-files...]",
	SilenceUsage: true,
	Short:        "Check TODOs by running their verification tests",
	RunE:         runTodosCheck,
}

func runTodosRun(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if todosDir == "" {
		todosDir = filepath.Join(workDir, ".todos")
	}

	if _, err := os.Stat(todosDir); os.IsNotExist(err) {
		return fmt.Errorf(".todos directory not found: %s", todosDir)
	}

	logger.Infof("Discovering TODOs in: %s", todosDir)

	filters := todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	}

	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}

	todoList, err := todos.DiscoverTODOs(todosDir, filters)
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
	}

	if len(args) > 0 {
		var filtered []*types.TODO
		for _, todo := range todoList {
			for _, arg := range args {
				matched := false
				if strings.Contains(arg, string(filepath.Separator)) || strings.HasSuffix(arg, ".md") {
					absArg := arg
					if !filepath.IsAbs(arg) {
						absArg = filepath.Join(workDir, arg)
					}
					matched = filepath.Clean(absArg) == filepath.Clean(todo.FilePath)
				} else {
					matched = strings.EqualFold(todo.Filename(), arg)
				}
				if matched {
					filtered = append(filtered, todo)
					break
				}
			}
		}
		todoList = filtered
	}

	if interactive && len(args) == 0 && len(todoList) > 0 {
		items := make([]string, len(todoList))
		for i, todo := range todoList {
			title := todo.Title
			if title == "" {
				title = todo.Filename()
			}
			items[i] = fmt.Sprintf("%s (%s)", title, todo.Status)
		}
		selected, err := choose.Run(items,
			choose.WithHeader("Select TODOs to run:"),
			choose.WithLimit(0),
		)
		if err != nil {
			return fmt.Errorf("interactive selection failed: %w", err)
		}
		if len(selected) == 0 {
			logger.Infof("No TODOs selected")
			return nil
		}
		filtered := make([]*types.TODO, len(selected))
		for i, idx := range selected {
			filtered[i] = todoList[idx]
		}
		todoList = filtered
	}

	if len(todoList) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}

	logger.Infof("Found %d TODOs", len(todoList))

	groups := todos.GroupTODOs(todoList, groupBy)
	fmt.Println(clicky.MustFormat(todos.FlattenGrouped(groups)))
	fmt.Println()

	// Flatten groups to ordered list for execution
	var orderedTodos types.TODOS
	for _, group := range groups {
		orderedTodos = append(orderedTodos, group.TODOs...)
	}
	todoList = orderedTodos

	interaction := &todos.UserInteraction{
		AskFunc: func(question todos.Question) (string, error) {
			fmt.Println(question.Pretty().ANSI())
			fmt.Print(clicky.Text("Your response: ", "text-green-600").ANSI())
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return "", fmt.Errorf("failed to read user input: %w", err)
			}
			return strings.TrimSpace(response), nil
		},
		NotifyFunc: func(notification todos.Notification) {
			fmt.Println(notification.Pretty().ANSI())
		},
	}

	for i, todo := range todoList {
		fmt.Println(clicky.Text(fmt.Sprintf("=== Executing TODO %d/%d: %s ===", i+1, len(todoList), todo.Filename()), "text-blue-600 font-bold").ANSI())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)

		var sessionID string
		if todo.LLM != nil && todo.LLM.SessionId != "" {
			sessionID = todo.LLM.SessionId
			logger.Infof("Resuming session: %s", sessionID)
		}

		config := claude.ClaudeExecutorConfig{
			WorkDir:   workDir,
			SessionID: sessionID,
			Timeout:   30 * time.Minute,
			Tools:     []string{"Read", "Edit", "Write", "Bash", "Glob", "Grep"},
			Dirty:     dirty,
		}
		if todo.LLM != nil {
			if todo.LLM.MaxCost > 0 {
				config.MaxBudgetUsd = todo.LLM.MaxCost
			}
			if todo.LLM.MaxTurns > 0 {
				config.MaxTurns = todo.LLM.MaxTurns
			}
			if todo.LLM.Model != "" {
				config.Model = todo.LLM.Model
			}
		}
		if maxBudget > 0 {
			config.MaxBudgetUsd = maxBudget
		}
		if maxTurns > 0 {
			config.MaxTurns = maxTurns
		}

		claudeExec := claude.NewClaudeExecutor(config)
		todoExec := todos.NewTODOExecutor(workDir, claudeExec, sessionID)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		var result *todos.ExecutionResult
		executionDone := make(chan bool, 1)

		cleanup := func() {
			if result != nil {
				fmt.Println()
				fmt.Println(result.Pretty().ANSI())
			}
			if todo.Status == types.StatusInProgress {
				if result != nil {
					if result.Success {
						todo.Status = types.StatusCompleted
					} else if result.Skipped {
						todo.Status = types.StatusSkipped
					} else {
						todo.Status = types.StatusFailed
					}
				} else {
					todo.Status = types.StatusFailed
				}
			}
			// FIXME: Persist frontmatter updates back to file
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic during TODO execution: %v\n%s", r, debug.Stack())
				}
				executionDone <- true
			}()
			result, err = todoExec.Execute(execCtx, todo)
		}()

		select {
		case <-executionDone:
			cleanup()
			if err != nil {
				logger.Errorf("TODO execution failed: %v", err)
			}
		case sig := <-sigChan:
			logger.Warnf("Received signal %v, shutting down gracefully...", sig)
			cancel()
			fmt.Println()
			fmt.Println(clicky.Text("Interrupted - cleaning up...", "text-yellow-600 font-bold").ANSI())
			select {
			case <-executionDone:
				cleanup()
			case <-time.After(5 * time.Second):
				logger.Warnf("Timeout waiting for graceful shutdown")
				cleanup()
			}
			signal.Stop(sigChan)
			fmt.Println()
			fmt.Println(clicky.Text("Execution interrupted by user", "text-red-600 font-bold").ANSI())
			return nil
		}

		signal.Stop(sigChan)
	}

	fmt.Println()
	fmt.Println(clicky.MustFormat(clicky.Text("All TODOs processed", "text-blue-600 font-bold")))
	return nil
}

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

	todosGetCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")

	todosCheckCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosCheckCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status")
	todosCheckCmd.Flags().DurationVar(&checkTimeout, "timeout", 2*time.Minute, "Test execution timeout")
}
