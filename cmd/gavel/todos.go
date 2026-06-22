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
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/cmux"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todosDir      string
	maxRetries    int
	filterStatus  string
	checkTimeout  time.Duration
	maxBudget     float64
	maxTurns      int
	interactive   bool
	groupBy       string
	dirty         bool
	dryRun        bool
	todosProvider string
	todosMode     string
	todoModel     string
	todoEffort    string
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
	GroupBy string `json:"group-by" flag:"group-by" help:"Group TODOs by: file, directory, repo, all, or none"`
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
	if err := validateTodosRunOptions(); err != nil {
		return err
	}

	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}
	logger.Infof("Discovering TODOs using provider: %s", selectedTodosProvider())

	filters := todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	}

	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}

	todoList, err := provider.List(context.Background(), filters)
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
	}

	if len(args) > 0 {
		todoList = filterTODOsByArgs(todoList, args, workDir)
	}

	if interactive && len(args) == 0 && len(todoList) > 0 {
		selected, err := selectTODOs(todoList, "Select TODOs to run:")
		if err != nil {
			return err
		}
		if selected == nil {
			logger.Infof("No TODOs selected")
			return nil
		}
		todoList = selected
	}

	if len(todoList) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}

	effectiveGroupBy := groupBy
	if todosMode == "cmux" && effectiveGroupBy == "" {
		effectiveGroupBy = todos.GroupByRepo
	}

	logger.Infof("Found %d TODOs", len(todoList))

	groups := todos.GroupTODOsWithWorkDir(todoList, effectiveGroupBy, workDir)
	fmt.Println(clicky.MustFormat(todos.FlattenGrouped(groups)))
	fmt.Println()

	if dryRun {
		return dryRunTODOs(groups, workDir)
	}

	interaction := newInteraction()

	if effectiveGroupBy != "" && effectiveGroupBy != todos.GroupByNone {
		return executeGroups(workDir, groups, interaction, provider)
	}

	// Flatten groups to ordered list for individual execution
	var orderedTodos types.TODOS
	for _, group := range groups {
		orderedTodos = append(orderedTodos, group.TODOs...)
	}
	return executeSingleTODOs(workDir, orderedTodos, interaction, provider)
}

func newInteraction() *todos.UserInteraction {
	return &todos.UserInteraction{
		AskFunc: func(question todos.Question) (string, error) {
			prompting.Prepare()
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
	if todoModel != "" {
		config.Model = todoModel
	}
	if maxBudget > 0 {
		config.MaxBudgetUsd = maxBudget
	}
	if maxTurns > 0 {
		config.MaxTurns = maxTurns
	}
	return config
}

func newExecutor(workDir string, todo *types.TODO) (todos.Executor, string) {
	if todosMode == "cmux" {
		return cmux.NewCmuxExecutor(newCmuxConfig(workDir, todo)), ""
	}
	config := newClaudeConfig(workDir, todo)
	return claude.NewClaudeExecutor(config), config.SessionID
}

func newCmuxConfig(workDir string, todo *types.TODO) cmux.CmuxExecutorConfig {
	model := ""
	if todo != nil && todo.LLM != nil {
		model = todo.LLM.Model
	}
	if todoModel != "" {
		model = todoModel
	}

	cwd := workDir
	if todo != nil && todo.CWD != "" {
		if filepath.IsAbs(todo.CWD) {
			cwd = todo.CWD
		} else {
			cwd = filepath.Join(workDir, todo.CWD)
		}
	}

	return cmux.CmuxExecutorConfig{
		WorkDir: cwd,
		Model:   model,
		Effort:  todoEffort,
		Timeout: 30 * time.Minute,
	}
}

func executeGroups(workDir string, groups []todos.TODOGroup, interaction *todos.UserInteraction, provider todos.Provider) error {
	for gi, group := range groups {
		if len(group.TODOs) == 0 {
			continue
		}
		fmt.Println(clicky.Text(fmt.Sprintf("=== Executing Group %d/%d: %s (%d TODOs) ===",
			gi+1, len(groups), group.Name, len(group.TODOs)), "text-blue-600 font-bold").ANSI())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
		executor, sessionID := newExecutor(workDir, group.TODOs[0])
		todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)

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

func executeSingleTODOs(workDir string, todoList types.TODOS, interaction *todos.UserInteraction, provider todos.Provider) error {
	for i, todo := range todoList {
		fmt.Println(clicky.Text(fmt.Sprintf("=== Executing TODO %d/%d: %s ===", i+1, len(todoList), todo.Filename()), "text-blue-600 font-bold").ANSI())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
		executor, sessionID := newExecutor(workDir, todo)
		todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)

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

		if todosMode == "cmux" {
			printCmuxDryRun(group, workDir)
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

func validateTodosRunOptions() error {
	switch todosMode {
	case "", "inline":
	case "cmux":
	default:
		return fmt.Errorf("invalid --mode %q: expected inline or cmux", todosMode)
	}

	switch todoEffort {
	case "", "low", "medium", "high":
	default:
		return fmt.Errorf("invalid --effort %q: expected low, medium, or high", todoEffort)
	}
	return nil
}

func printCmuxDryRun(group todos.TODOGroup, workDir string) {
	groupWorkDir := workDir
	if group.Name != "" && group.Name != todos.UngroupedLabel && filepath.IsAbs(group.Name) {
		groupWorkDir = group.Name
	}
	agent, model := resolveTodoAgent(todoModel)
	agentCmd := cmux.AgentCommand(agent, model)
	name := cmux.AgentWorkspaceName(groupWorkDir, agent)

	fmt.Printf("=== cmux Group: %s (%d TODOs) ===\n\n", group.Name, len(group.TODOs))
	fmt.Println("### Commands")
	fmt.Println("  cmux list-workspaces --json")
	fmt.Printf("  cmux new-workspace --cwd %q --name %q --focus true --id-format both  # if missing\n", groupWorkDir, name)
	fmt.Printf("  cmux new-surface --type terminal --workspace <workspace-ref> --working-directory %q --focus true\n", groupWorkDir)
	fmt.Println("  cmux read-screen --workspace <workspace-ref> --surface <surface-ref> --lines 120")
	fmt.Printf("  cmux send --workspace <workspace-ref> --surface <surface-ref> -- %q\n", agentCmd)
	fmt.Println("  cmux send-key --workspace <workspace-ref> --surface <surface-ref> Enter")
	fmt.Println("  cmux read-screen --workspace <workspace-ref> --surface <surface-ref> --lines 120")
	fmt.Println("  cmux send --workspace <workspace-ref> --surface <surface-ref> -- <prompt>")
	fmt.Println("  cmux send-key --workspace <workspace-ref> --surface <surface-ref> Enter")
	fmt.Println("  cmux read-screen --workspace <workspace-ref> --surface <surface-ref> --lines 120")
	fmt.Println()
	printSectionCommands("Pre-check commands (steps_to_reproduce)", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.StepsToReproduce })
	fmt.Println("### Prompt")
	fmt.Println(buildCmuxPrompt(group.TODOs, workDir))
	printSectionCommands("Verification commands", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.Verification })
}

func buildCmuxPrompt(todoList []*types.TODO, workDir string) string {
	return cmux.BuildPrompt(todoList, workDir, todoEffort)
}

func effortDirective(effort string) string {
	return cmux.EffortDirective(effort)
}

func resolveTodoAgent(model string) (agent string, modelFlag string) {
	return cmux.ResolveAgent(model)
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
