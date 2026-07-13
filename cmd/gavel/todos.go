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

	cmuxprov "github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	maxRetries    int
	filterStatus  string
	checkTimeout  time.Duration
	maxBudget     float64
	maxTurns      int
	interactive   bool
	groupBy       string
	dirty         bool
	dryRun        bool
	commitAfter   bool
	checkAfter    bool
	todosMode     string
	todosDriver   string
	todoModel     string
	todoEffort    string
	resumeSession bool
	// todosRunMode is the parsed --mode (run/plan); verify short-circuits to the
	// verification loop before it is set. Resolved once in runTodosRun.
	todosRunMode types.RunMode
)

var todosCmd = &cobra.Command{
	Use:          "todos",
	Aliases:      []string{"todo"},
	SilenceUsage: true,
	Short:        "Automated TODO execution + AI verification with coding agents",
	Long: `Run, verify, and manage TODOs — units of work an AI coding agent implements
and gavel then scores against their acceptance criteria.

A TODO is a PostgreSQL-backed issue with a title, body, status, priority,
acceptance criteria, verification fixtures, and execution history. Todos come
from source TODO/FIXME comments ('todos sync'), explicit portable/Grite imports,
or by hand ('todos create').

Subcommands:
  list      List TODOs (filter by --status, group with --group-by)
  get       Show one TODO in detail (accepts a short id, full id, title, or alias)
  create    Create a TODO
  run       Have a coding agent implement TODOs (--mode run|plan|verify)
  verify    AI-score whether a TODO's commits implement its criteria
  check     Run a TODO's verification tests
  edit / comment / reopen / criteria / sync / plan / transfer

Examples:
  gavel todos list
  gavel todos list --all            # list every registered project
  gavel todos list --all --done     # include verified/completed items
  gavel todos get <id>
  gavel todos run                  # implement all pending todos
  gavel todos run --mode plan      # propose a reviewable plan first
  gavel todos verify --strict      # score committed work, non-zero if unmet`,
}

var todosRunCmd = &cobra.Command{
	Use:          "run [todo-titles...]",
	SilenceUsage: true,
	Short:        "Have a coding agent implement TODOs (run/plan/verify modes)",
	Long: `Drive an AI coding agent (Claude or Codex, via cmux or headless) to work TODOs.

With no arguments it runs every pending TODO; pass titles, ids, or aliases to select a
subset, or -i to pick interactively. --mode chooses the operation:
  run     implement the TODO (default)
  plan    propose a reviewable plan instead of editing (see 'todos plan')
  verify  score committed work (delegates to 'todos verify')

After each TODO's agent finishes, gavel commits its changes (--commit, on by
default; disable with --commit=false). --check additionally runs the configured
checks suite and feeds failures back to the agent until they pass. Use --dry-run
to print the prompts and commands without executing.

Examples:
  gavel todos run                          # implement all pending todos
  gavel todos run "Fix flaky parser test"  # one todo by title
  gavel todos run -i                       # interactively select
  gavel todos run --mode plan              # propose plans for review
  gavel todos run --check                  # run checks + fix failures in-loop
  gavel todos run --driver codex-headless --model o3 # codex-agent app-server
  gavel todos run --dry-run                # preview prompts, no changes`,
	RunE: runTodosRun,
}

type TodosListOptions struct {
	Status  string `json:"status" flag:"status" help:"Filter TODOs by status"`
	All     bool   `json:"all" flag:"all" help:"List PostgreSQL-backed TODOs from all registered projects"`
	Done    bool   `json:"done" flag:"done" help:"Include verified and completed TODOs"`
	Since   string `json:"since" flag:"since" help:"Show TODOs created or updated since (e.g. 7d, now-30d, 2024-01-01)"`
	GroupBy string `json:"group-by" flag:"group-by" help:"Group TODOs by: file, directory, repo, all, or none"`
}

func (opts TodosListOptions) GetName() string { return "list" }

var todosGetCmd = &cobra.Command{
	Use:          "get <id-or-alias>",
	SilenceUsage: true,
	Short:        "Display detailed information about a PostgreSQL-backed TODO",
	Long: `Show one TODO in full — metadata, body, acceptance criteria, and run history.

The argument matches a short id, a full id, the title, or an imported alias.

Examples:
  gavel todos get 3f2a1b

  gavel todos get "Fix flaky parser test"`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosGet,
}

var todosCheckCmd = &cobra.Command{
	Use:          "check [ids-or-aliases...]",
	SilenceUsage: true,
	Short:        "Check TODOs by running their verification tests",
	Long: `Run each TODO's verification commands (the 'verify' fixtures) and report pass/fail.

This runs the deterministic checks attached to a TODO — unlike 'todos verify',
which uses AI to score acceptance criteria. Exits non-zero if any TODO fails.
With no arguments it checks every discovered TODO; pass ids or aliases to select some.

Examples:
  gavel todos check
  gavel todos check 3f2a1b
  gavel todos check --status in_progress`,
	RunE: runTodosCheck,
}

func runTodosRun(cmd *cobra.Command, args []string) error {
	if err := validateTodosRunOptions(); err != nil {
		return err
	}
	mode, err := types.ParseRunMode(todosMode)
	if err != nil {
		return err
	}
	if mode == types.ModeVerify {
		// `todos run --mode verify` is the same operation as `todos verify`.
		return runTodosVerify(cmd, args)
	}
	todosRunMode = mode

	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	logger.Infof("Discovering TODOs from PostgreSQL")

	filters := todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	}

	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}

	todoList, err := resolveRequestedTODOs(context.Background(), provider, args, filters)
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w", err)
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
	if policy, ok := provider.(todos.GroupExecutionPolicy); ok && !policy.SupportsGroupedExecution() {
		if effectiveGroupBy != "" && effectiveGroupBy != todos.GroupByNone {
			return fmt.Errorf("--group-by is not supported by the native PostgreSQL runtime; run one issue at a time")
		}
		effectiveGroupBy = todos.GroupByNone
	} else if kind, kerr := resolveDriverKind(nil); kerr == nil && kind.Mechanism() == "cmux" && effectiveGroupBy == "" {
		effectiveGroupBy = todos.GroupByRepo
	}

	logger.Infof("Found %d TODOs", len(todoList))

	groups := todos.GroupTODOsWithWorkDir(todoList, effectiveGroupBy, workDir)
	fmt.Println(clicky.MustFormat(todos.FlattenGrouped(groups)))
	fmt.Println()

	if dryRun {
		return dryRunTODOs(groups, workDir, provider)
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

func newExecutor(workDir string, todo *types.TODO, provider todos.Provider) (todos.Executor, string, error) {
	kind, err := resolveDriverKind(todo)
	if err != nil {
		return nil, "", err
	}
	cfg, err := newDriverConfig(context.Background(), kind, workDir, todo, provider)
	if err != nil {
		return nil, "", err
	}
	if todosRunMode != types.ModePlan {
		// Post-run checks are fixture-backed verify plugins inside the agent loop.
		// --check force-enables them (a bare Workflow.Verify); .gavel.yaml/frontmatter
		// `checks` enable them too.
		var checkVerify *api.Verify
		if checkAfter {
			checkVerify = &api.Verify{}
		}
		verifiers, maxIter, err := todos.BuildCheckVerifiers(workDir, []*types.TODO{todo}, checkVerify)
		if err != nil {
			return nil, "", err
		}
		cfg.Verifiers = verifiers
		cfg.MaxIterations = maxIter
	}
	return drivers.New(kind, cfg)
}

// resolveDriverKind selects the driver: the explicit --driver flag when set,
// otherwise "<agent>-cmux" for the model's agent (cmux is the default —
// drivers.Default; headless is the non-interactive opt-in). --mode selects the
// todo OPERATION (run/plan/verify), not the mechanism.
func resolveDriverKind(todo *types.TODO) (drivers.Kind, error) {
	if strings.TrimSpace(todosDriver) != "" {
		return drivers.Parse(todosDriver)
	}
	model := todoModel
	if model == "" && todo != nil && todo.LLM != nil {
		model = todo.LLM.Model
	}
	agent, _ := claude.ResolveAgent(model)
	return drivers.Parse(agent + "-cmux")
}

// newDriverConfig assembles the shared driver config from flags and the todo.
// cmux mints and manages its own --session-id (reading any prior session from
// the todo itself), so SessionID stays empty for it; the sdk/headless/api paths
// resume by carrying the prior session id explicitly.
func newDriverConfig(ctx context.Context, kind drivers.Kind, workDir string, todo *types.TODO, provider todos.Provider) (drivers.Config, error) {
	model := ""
	prior := ""
	var maxCost float64
	var turns int
	if todo != nil && todo.LLM != nil {
		model = todo.LLM.Model
		prior = todo.LLM.SessionId
		maxCost = todo.LLM.MaxCost
		turns = todo.LLM.MaxTurns
	}
	if todoModel != "" {
		model = todoModel
	}
	if maxBudget > 0 {
		maxCost = maxBudget
	}
	if maxTurns > 0 {
		turns = maxTurns
	}

	cwd := workDir
	if todo != nil && todo.CWD != "" {
		if filepath.IsAbs(todo.CWD) {
			cwd = todo.CWD
		} else {
			cwd = filepath.Join(workDir, todo.CWD)
		}
	}

	cfg := drivers.Config{
		WorkDir:      cwd,
		Model:        model,
		Effort:       todoEffort,
		Mode:         todosRunMode,
		Resume:       resumeSession,
		Timeout:      30 * time.Minute,
		MaxBudgetUsd: maxCost,
		MaxTurns:     turns,
		Tools:        drivers.DefaultTools(),
		Dirty:        dirty,
	}
	if todo != nil && (todosRunMode == types.ModePlan || todosRunMode == types.ModeRun) {
		// Plan mode: the recorded plan feeds a re-plan (updated/unchanged). Run
		// mode: the approved/edited plan steers the implementation.
		content, err := nativeExistingPlan(ctx, provider, todo, todosRunMode)
		if err != nil {
			return drivers.Config{}, err
		}
		cfg.ExistingPlan = content
	}
	if kind.Mechanism() != "cmux" {
		cfg.SessionID = prior
	}
	return cfg, nil
}

// nativeExistingPlan loads plan content through the active DB provider.
// Plan paths are source metadata only and are never read on the DB runtime.
func nativeExistingPlan(ctx context.Context, provider todos.Provider, todo *types.TODO, mode types.RunMode) (string, error) {
	if provider == nil || todo == nil {
		return "", nil
	}
	contentProvider, ok := provider.(todos.PlanContentProvider)
	if !ok {
		return "", fmt.Errorf("native PostgreSQL TODO provider does not support durable plan content")
	}
	return contentProvider.PlanMarkdown(ctx, todo, mode)
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
		executor, sessionID, err := newExecutor(workDir, group.TODOs[0], provider)
		if err != nil {
			cancel()
			return err
		}
		todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)
		todoExec.SetMode(todosRunMode)
		todoExec.SetResume(resumeSession)

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
		maybeCommitAfter(workDir, provider, group.TODOs[0], safeResult(results, 0))
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
		executor, sessionID, err := newExecutor(workDir, todo, provider)
		if err != nil {
			cancel()
			return err
		}
		todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)
		todoExec.SetMode(todosRunMode)
		todoExec.SetResume(resumeSession)

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
		maybeCommitAfter(workDir, provider, todo, result)
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
	// The executor persists lifecycle state through the native provider.
}

func safeResult(results []*todos.ExecutionResult, i int) *todos.ExecutionResult {
	if i < len(results) {
		return results[i]
	}
	return nil
}

func dryRunTODOs(groups []todos.TODOGroup, workDir string, provider todos.Provider) error {
	isGrouped := len(groups) > 1 || (len(groups) == 1 && groups[0].Name != "")

	for _, group := range groups {
		if len(group.TODOs) == 0 {
			continue
		}

		if kind, kerr := resolveDriverKind(nil); kerr == nil && kind.Mechanism() == "cmux" {
			if err := printCmuxDryRun(group, workDir, provider); err != nil {
				return err
			}
			continue
		}

		if isGrouped {
			fmt.Printf("=== Group: %s ===\n\n", group.Name)
			printSectionCommands("Pre-check commands (steps_to_reproduce)", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.StepsToReproduce })
			fmt.Println("### Prompt")
			groupPrompt, err := buildTodoRunPrompt(group.TODOs, workDir, provider)
			if err != nil {
				return err
			}
			fmt.Println(groupPrompt)
			printSectionCommands("Verification commands", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.Verification })
		} else {
			for _, todo := range group.TODOs {
				fmt.Printf("=== TODO: %s ===\n\n", todo.Filename())
				printTodoCommands("Pre-check commands (steps_to_reproduce)", todo.StepsToReproduce)
				fmt.Println("### Prompt")
				todoPrompt, err := buildTodoRunPrompt([]*types.TODO{todo}, workDir, provider)
				if err != nil {
					return err
				}
				fmt.Println(todoPrompt)
				printTodoCommands("Verification commands", todo.Verification)
			}
		}
	}
	return nil
}

func validateTodosRunOptions() error {
	if strings.TrimSpace(todosDriver) != "" {
		if _, err := drivers.Parse(todosDriver); err != nil {
			return err
		}
	}

	if _, err := types.ParseRunMode(todosMode); err != nil {
		return err
	}

	switch todoEffort {
	case "", "low", "medium", "high", "xhigh":
	default:
		return fmt.Errorf("invalid --effort %q: expected low, medium, high, or xhigh", todoEffort)
	}
	return nil
}

func printCmuxDryRun(group todos.TODOGroup, workDir string, provider todos.Provider) error {
	groupWorkDir := workDir
	if group.Name != "" && group.Name != todos.UngroupedLabel && filepath.IsAbs(group.Name) {
		groupWorkDir = group.Name
	}
	agent, model := resolveTodoAgent(todoModel)
	sessionID := ""
	if agent == "claude" {
		sessionID = "<session-id>"
	}
	agentCmd := cmuxprov.AgentCommand(cmuxprov.AgentCommandOpts{Agent: agent, Model: model, SessionID: sessionID})
	name := cmuxprov.AgentWorkspaceName(groupWorkDir, agent)

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
	if agent == "claude" {
		fmt.Println("  # then tail ~/.claude/projects/<cwd>/<session-id>.jsonl for progress until the turn ends")
	} else {
		fmt.Println("  cmux read-screen --workspace <workspace-ref> --surface <surface-ref> --lines 120")
	}
	fmt.Println()
	printSectionCommands("Pre-check commands (steps_to_reproduce)", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.StepsToReproduce })
	fmt.Println("### Prompt")
	cmuxPrompt, err := buildTodoRunPrompt(group.TODOs, workDir, provider)
	if err != nil {
		return err
	}
	fmt.Println(cmuxPrompt)
	printSectionCommands("Verification commands", group.TODOs, func(t *types.TODO) []*fixtures.FixtureNode { return t.Verification })
	return nil
}

// buildTodoRunPrompt renders the mode's prompt exactly as a dispatch would:
// framing, sections, effort directive, and the envelope schema instruction.
func buildTodoRunPrompt(todoList []*types.TODO, workDir string, provider todos.Provider) (string, error) {
	mode := todosRunMode
	if mode == "" {
		mode = types.ModeRun
	}
	tmpl, err := todoprompt.ResolveTemplate(workDir, mode)
	if err != nil {
		return "", err
	}
	opts := todoprompt.Options{WorkDir: workDir, Mode: mode, Effort: todoEffort, Template: tmpl}
	// Mirror the executor: a single todo's recorded plan feeds the prompt, so the
	// dry-run preview matches the run (an approved plan for run mode, the prior
	// plan for a re-plan).
	if len(todoList) == 1 && (mode == types.ModePlan || mode == types.ModeRun) {
		content, err := nativeExistingPlan(context.Background(), provider, todoList[0], mode)
		if err != nil {
			return "", err
		}
		opts.ExistingPlan = content
	}
	req, _, err := todoprompt.Render(todoList, opts)
	if err != nil {
		return "", err
	}
	return req.Prompt.User, nil
}

func effortDirective(effort string) string {
	return todoprompt.EffortDirective(effort)
}

func resolveTodoAgent(model string) (agent string, modelFlag string) {
	return claude.ResolveAgent(model)
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
