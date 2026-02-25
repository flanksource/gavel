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
	"github.com/flanksource/gavel/fixtures"
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
	dryRun       bool
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

	if dryRun {
		return dryRunTODOs(groups, workDir)
	}

	interaction := newInteraction()

	if groupBy != "" && groupBy != todos.GroupByNone {
		return executeGroups(workDir, groups, interaction)
	}

	// Flatten groups to ordered list for individual execution
	var orderedTodos types.TODOS
	for _, group := range groups {
		orderedTodos = append(orderedTodos, group.TODOs...)
	}
	return executeSingleTODOs(workDir, orderedTodos, interaction)
}

func newInteraction() *todos.UserInteraction {
	return &todos.UserInteraction{
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
}

func newClaudeConfig(workDir string, todo *types.TODO) claude.ClaudeExecutorConfig {
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
	return config
}

func executeGroups(workDir string, groups []todos.TODOGroup, interaction *todos.UserInteraction) error {
	for gi, group := range groups {
		if len(group.TODOs) == 0 {
			continue
		}
		fmt.Println(clicky.Text(fmt.Sprintf("=== Executing Group %d/%d: %s (%d TODOs) ===",
			gi+1, len(groups), group.Name, len(group.TODOs)), "text-blue-600 font-bold").ANSI())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
		config := newClaudeConfig(workDir, group.TODOs[0])
		claudeExec := claude.NewClaudeExecutor(config)
		todoExec := todos.NewTODOExecutor(workDir, claudeExec, config.SessionID)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		var results []*todos.ExecutionResult
		var execErr error
		executionDone := make(chan bool, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic during group execution: %v\n%s", r, debug.Stack())
				}
				executionDone <- true
			}()
			results, execErr = todoExec.ExecuteGroup(execCtx, group.TODOs)
		}()

		interrupted := false
		select {
		case <-executionDone:
		case sig := <-sigChan:
			logger.Warnf("Received signal %v, shutting down gracefully...", sig)
			cancel()
			fmt.Println(clicky.Text("Interrupted - cleaning up...", "text-yellow-600 font-bold").ANSI())
			select {
			case <-executionDone:
			case <-time.After(5 * time.Second):
				logger.Warnf("Timeout waiting for graceful shutdown")
			}
			interrupted = true
		}

		signal.Stop(sigChan)
		cancel()

		for i, todo := range group.TODOs {
			if i < len(results) && results[i] != nil {
				fmt.Println(results[i].Pretty().ANSI())
			}
			cleanupTODOStatus(todo, safeResult(results, i))
		}

		if interrupted {
			fmt.Println(clicky.Text("Execution interrupted by user", "text-red-600 font-bold").ANSI())
			return nil
		}
		if execErr != nil {
			logger.Errorf("Group execution failed: %v", execErr)
		}
	}

	fmt.Println()
	fmt.Println(clicky.MustFormat(clicky.Text("All TODOs processed", "text-blue-600 font-bold")))
	return nil
}

func executeSingleTODOs(workDir string, todoList types.TODOS, interaction *todos.UserInteraction) error {
	for i, todo := range todoList {
		fmt.Println(clicky.Text(fmt.Sprintf("=== Executing TODO %d/%d: %s ===", i+1, len(todoList), todo.Filename()), "text-blue-600 font-bold").ANSI())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
		config := newClaudeConfig(workDir, todo)
		claudeExec := claude.NewClaudeExecutor(config)
		todoExec := todos.NewTODOExecutor(workDir, claudeExec, config.SessionID)

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		var result *todos.ExecutionResult
		var execErr error
		executionDone := make(chan bool, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Errorf("Panic during TODO execution: %v\n%s", r, debug.Stack())
				}
				executionDone <- true
			}()
			result, execErr = todoExec.Execute(execCtx, todo)
		}()

		interrupted := false
		select {
		case <-executionDone:
		case sig := <-sigChan:
			logger.Warnf("Received signal %v, shutting down gracefully...", sig)
			cancel()
			fmt.Println(clicky.Text("Interrupted - cleaning up...", "text-yellow-600 font-bold").ANSI())
			select {
			case <-executionDone:
			case <-time.After(5 * time.Second):
				logger.Warnf("Timeout waiting for graceful shutdown")
			}
			interrupted = true
		}

		signal.Stop(sigChan)
		cancel()

		if result != nil {
			fmt.Println()
			fmt.Println(result.Pretty().ANSI())
		}
		cleanupTODOStatus(todo, result)

		if interrupted {
			fmt.Println(clicky.Text("Execution interrupted by user", "text-red-600 font-bold").ANSI())
			return nil
		}
		if execErr != nil {
			logger.Errorf("TODO execution failed: %v", execErr)
		}
	}

	fmt.Println()
	fmt.Println(clicky.MustFormat(clicky.Text("All TODOs processed", "text-blue-600 font-bold")))
	return nil
}

func cleanupTODOStatus(todo *types.TODO, result *todos.ExecutionResult) {
	if todo.Status != types.StatusInProgress {
		return
	}
	if result != nil {
		switch {
		case result.Success:
			todo.Status = types.StatusCompleted
		case result.Skipped:
			todo.Status = types.StatusSkipped
		default:
			todo.Status = types.StatusFailed
		}
	} else {
		todo.Status = types.StatusFailed
	}
	// FIXME: Persist frontmatter updates back to file
}

func safeResult(results []*todos.ExecutionResult, i int) *todos.ExecutionResult {
	if i < len(results) {
		return results[i]
	}
	return nil
}

func dryRunTODOs(groups []todos.TODOGroup, workDir string) error {
	isGrouped := len(groups) > 1 || (len(groups) == 1 && groups[0].Name != "")

	for _, group := range groups {
		if len(group.TODOs) == 0 {
			continue
		}

		if isGrouped {
			fmt.Printf("=== Group: %s ===\n\n", group.Name)
			printSectionCommands("Pre-check commands (steps_to_reproduce)", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.StepsToReproduce })
			fmt.Println("### Prompt")
			fmt.Println(claude.BuildGroupPrompt(group.TODOs, workDir))
			printSectionCommands("Verification commands", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.Verification })
		} else {
			for _, todo := range group.TODOs {
				fmt.Printf("=== TODO: %s ===\n\n", todo.Filename())
				printTodoCommands("Pre-check commands (steps_to_reproduce)", todo.StepsToReproduce)
				fmt.Println("### Prompt")
				fmt.Println(claude.BuildPrompt(todo, workDir))
				printTodoCommands("Verification commands", todo.Verification)
			}
		}
	}
	return nil
}

func printSectionCommands(header string, todoList []*types.TODO, getNodes func(*types.TODO) []*fixtures.FixtureNode) {
	var lines []string
	for _, todo := range todoList {
		for _, node := range getNodes(todo) {
			if node.Test != nil {
				lines = append(lines, fmt.Sprintf("  [%s] %s", todo.Filename(), node.Test.ExecBase().Pretty().String()))
			}
		}
	}
	if len(lines) == 0 {
		return
	}
	fmt.Printf("### %s\n", header)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println()
}

func printTodoCommands(header string, nodes []*fixtures.FixtureNode) {
	var lines []string
	for _, node := range nodes {
		if node.Test != nil {
			lines = append(lines, fmt.Sprintf("  %s", node.Test.ExecBase().Pretty().String()))
		}
	}
	if len(lines) == 0 {
		return
	}
	fmt.Printf("### %s\n", header)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println()
}
