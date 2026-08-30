package main

import (
	"bufio"
	"context"
	"errors"
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
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/run"
	todospec "github.com/flanksource/gavel/todos/spec"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	filterStatus     string
	checkTimeout     time.Duration
	checkConcurrency int
	maxBudget        float64
	maxTurns         int
	interactive      bool
	groupBy          string
	dirty            bool
	dryRun           bool
	commitAfter      bool
	checkAfter       bool
	todosMode        string
	todosPrompt      string
	// todosModeExplicit records whether --mode was actually typed. The flag
	// defaults to "run", so without this a `--prompt triage` would be read as a
	// caller asking for a run-class triage and rejected as contradictory.
	todosModeExplicit bool
	todosDriver       string
	todoModel         string
	todoEffort        string
	resumeSession     bool
	// forceRun answers the "this todo is already running" question up front, so
	// an unattended run can dispatch alongside the live one without a prompt.
	forceRun bool
	// todosRunMode is the parsed public --mode (run/plan). Verify is an internal
	// issue lifecycle step entered by `todos check`, not an agent run mode.
	todosRunMode types.RunMode
)

var todosCmd = &cobra.Command{
	Use:          "todos",
	Aliases:      []string{"todo"},
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
	Short: "Automated TODO execution and fixture-backed verification",
	Long: `Run, check, and manage TODOs — units of work an AI coding agent implements
and gavel verifies with their persisted definition-of-done fixture.

A TODO is a PostgreSQL-backed issue with a title, body, status, priority,
acceptance criteria, verification fixtures, and execution history. Todos come
from source TODO/FIXME comments ('todos sync'), explicit portable imports,
or by hand ('todos create').

Subcommands:
  list      List TODOs (filter by --status, group with --group-by)
  get       Show one TODO in detail (accepts a short id, full id, title, or alias)
  create    Create a TODO
  run       Have a coding agent implement TODOs (--mode run|plan)
  check     Run a TODO's fixture-backed definition of done
  push      Open a GitHub issue for a TODO and link the two
  edit / comment / reopen / criteria / sync / plan / transfer

Examples:
  gavel todos list
  gavel todos list --all            # list every registered project
  gavel todos list --all --done     # include verified/completed items
  gavel todos get <id>
  gavel todos run                  # implement all pending todos
  gavel todos run --mode plan      # propose a reviewable plan first
  gavel todos check <id>           # run the issue's definition of done`,
}

var todosRunCmd = &cobra.Command{
	Use:          "run [todo-titles...]",
	SilenceUsage: true,
	Short:        "Have a coding agent implement TODOs (run/plan modes)",
	Long: `Drive an AI coding agent (Claude or Codex, via cmux or headless) to work TODOs.

With no arguments it runs every pending TODO; pass titles, ids, or aliases to select a
subset, or -i to pick interactively. --mode chooses the operation:
  run     implement the TODO (default)
  plan    propose a reviewable plan instead of editing (see 'todos plan')

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
  gavel todos run --driver cli --model o3            # headless CLI; o3 selects codex
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
	Short:        "Run TODOs' fixture-backed definitions of done",
	Long: `Run each TODO's complete definition of done and report pass/fail.

The check is a verify-only issue lifecycle run: configured test/lint steps, the
persisted Verification fixture, and acceptance-criteria AI checklist all use the
same gavel fixture/CEL pipeline as verification after implementation. Exits
non-zero if any TODO fails.
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
	todosModeExplicit = cmd.Flags().Changed("mode")
	todosRunMode = mode

	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Resolved without todos: only the config-level answers (driver, groupBy) are
	// needed to decide how to group, and grouping decides which todos each run
	// resolves with.
	resolved, err := todosSpec(workDir, nil)
	if err != nil {
		return err
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
	if effectiveGroupBy == "" {
		effectiveGroupBy = resolved.GroupBy
	}
	if policy, ok := provider.(todos.GroupExecutionPolicy); ok && !policy.SupportsGroupedExecution() {
		if effectiveGroupBy != "" && effectiveGroupBy != todos.GroupByNone {
			return fmt.Errorf("--group-by is not supported by the native PostgreSQL runtime; run one issue at a time")
		}
		effectiveGroupBy = todos.GroupByNone
	} else if resolved.Driver.Mechanism() == "cmux" && effectiveGroupBy == "" {
		effectiveGroupBy = todos.GroupByRepo
	}

	logger.Infof("Found %d TODOs", len(todoList))

	groups := todos.GroupTODOsWithWorkDir(todoList, effectiveGroupBy, workDir)
	fmt.Println(clicky.MustFormat(todos.FlattenGrouped(groups)))
	fmt.Println()

	if dryRun {
		return dryRunTODOs(groups, workDir, resolved, provider)
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

func newExecutor(workDir string, todoList []*types.TODO, provider todos.Provider) (todos.Executor, string, time.Duration, error) {
	cfg, resolved, err := newAgentRunConfig(context.Background(), workDir, todoList, provider)
	if err != nil {
		return nil, "", 0, err
	}
	if cfg.Mode != types.ModePlan {
		// Post-run checks are fixture-backed verify plugins inside the agent loop.
		// --check force-enables them (a bare Workflow.Verify); .gavel.yaml/frontmatter
		// `checks` enable them too.
		if checkAfter {
			if cfg.Spec.Workflow == nil {
				cfg.Spec.Workflow = &api.Workflow{}
			}
			if cfg.Spec.Workflow.Verify == nil {
				cfg.Spec.Workflow.Verify = &api.Verify{}
			}
		}
		// The grader resolves through its own chain — .gavel.yaml todos.verify >
		// ai: — so a definition of done is never marked by the model, backend and
		// session that implemented it. The run's flags and the todos' `llm:`
		// frontmatter are deliberately not fed in: they state how to implement,
		// and a todo that pins a cmux model must not thereby pin its own grader.
		// `gavel todos check` is the entrypoint whose flags are a verification
		// request, and it resolves them as the override layer itself.
		grader, err := todospec.Resolve(todospec.Input{WorkDir: workDir, Mode: types.ModeVerify})
		if err != nil {
			return nil, "", 0, err
		}
		verifiers, maxIter, err := todos.BuildCheckVerifiers(todos.CheckVerifierOptions{
			WorkDir: workDir,
			Todos:   todoList,
			Run:     &cfg.Spec,
			Grader:  grader.Spec,
		})
		if err != nil {
			return nil, "", 0, err
		}
		cfg.Verifiers = verifiers
		cfg.MaxIterations = maxIter
	}
	executor, sessionID, err := drivers.New(resolved.Driver, cfg)
	return executor, sessionID, resolved.Timeout, err
}

// todosSpec resolves the run configuration for todoList through the shared
// todos/spec seam, so the CLI and the dashboard layer .gavel.yaml, the mode's
// .prompt frontmatter, per-todo frontmatter and the request identically.
//
// The flags are the request layer, so an unset flag must stay zero — that is why
// --model and --effort have empty defaults rather than "medium": a non-zero
// default here would silently beat the .prompt frontmatter it claims to defer to.
func todosSpec(workDir string, todoList []*types.TODO) (todospec.Resolved, error) {
	return todosSpecForMode(workDir, todosRunMode, todoList, 0)
}

// todosSpecForMode is todosSpec for a command that owns its mode and timeout
// flag rather than reading the shared --mode/--timeout run flags.
func todosSpecForMode(workDir string, mode types.RunMode, todoList []*types.TODO, timeout time.Duration) (todospec.Resolved, error) {
	var override api.Spec
	override.Name = todoModel
	override.Effort = api.Effort(todoEffort)
	override.Budget.Cost = maxBudget
	override.Budget.MaxTurns = maxTurns
	if timeout > 0 {
		override.Budget.Timeout = timeout.String()
	}
	// A named prompt declares its own behaviour class, so the defaulted --mode
	// must not be passed alongside it — only a mode the caller actually typed,
	// which Resolve then checks for agreement.
	if strings.TrimSpace(todosPrompt) != "" && !todosModeExplicit {
		mode = ""
	}
	return todospec.Resolve(todospec.Input{
		WorkDir:  workDir,
		Mode:     mode,
		Prompt:   todosPrompt,
		Todos:    todoList,
		Override: override,
		Driver:   todosDriver,
		// The CLI drains no approval queue: a run that asked for one would block
		// forever, so an approval-gated posture must fail loud in Resolve.
		CanApprove: false,
	})
}

// newAgentRunConfig assembles Captain's canonical Spec plus Gavel-only
// orchestration state from the resolved seam and the todos being run.
// cmux mints and manages its own --session-id (reading any prior session from
// the todo itself), so SessionID stays empty for it; the agent/cli/api paths
// resume by carrying the prior session id explicitly.
func newAgentRunConfig(ctx context.Context, workDir string, todoList []*types.TODO, provider todos.Provider) (todos.AgentRunConfig, todospec.Resolved, error) {
	resolved, err := todosSpec(workDir, todoList)
	if err != nil {
		return todos.AgentRunConfig{}, todospec.Resolved{}, err
	}
	spec := resolved.Spec
	if spec.Setup == nil {
		spec.Setup = &shell.Setup{}
	}
	// Where the work happens and how the checkout is materialised both belong to
	// the executor: it derives the group's working directory from WorkDir plus the
	// todo's CWD and hands it to the setup hook, which anchors the rest of the
	// setup against it once, inside the run. Resolving or pre-joining here as well
	// would anchor Cwd and BaseDir to the invocation directory and double-join a
	// relative todo CWD.
	// One representation of dirty, shared with the dashboard: --dirty makes a
	// checkout's worktree carry the working tree's uncommitted changes across
	// rather than starting from a pristine tree. With no checkout, or a checkout
	// with no worktree, the run already happens in the dirty tree, so there is
	// nothing to carry.
	if dirty && spec.Setup.Checkout != nil && spec.Setup.Checkout.Worktree != nil {
		spec.Setup.Checkout.Worktree.Uncommitted = shell.CloneClone
	}
	// Resolve already cleared commits for the modes that must never commit; this
	// is only the --commit flag, which is the CLI's alone.
	if !commitAfter {
		if spec.Workflow != nil {
			spec.Workflow.Commits = nil
		}
	} else if resolved.Mode == types.ModeRun {
		if spec.Workflow == nil {
			spec.Workflow = &api.Workflow{}
		}
		if len(spec.Workflow.Commits) == 0 {
			spec.Workflow.Commits = []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesFull}}
		}
	}
	cfg := todos.AgentRunConfig{
		Spec:      spec,
		WorkDir:   workDir,
		Mode:      resolved.Mode,
		Prompt:    resolved.Prompt,
		Envelope:  resolved.Envelope,
		Template:  resolved.Template,
		Resume:    resumeSession,
		Approvals: resolved.Approvals,
	}
	if resolved.Envelope == todoprompt.EnvelopeTriage {
		backlog, err := triageBacklog(ctx, provider, todoList)
		if err != nil {
			return todos.AgentRunConfig{}, todospec.Resolved{}, err
		}
		cfg.Backlog = backlog
	}
	if len(todoList) == 1 && (resolved.Mode == types.ModePlan || resolved.Mode == types.ModeRun) {
		// Plan mode: the recorded plan feeds a re-plan (updated/unchanged). Run
		// mode: the approved/edited plan steers the implementation. Single-todo
		// only — a group run has no single plan to attribute.
		content, err := nativeExistingPlan(ctx, provider, todoList[0], resolved.Mode)
		if err != nil {
			return todos.AgentRunConfig{}, todospec.Resolved{}, err
		}
		cfg.ExistingPlan = content
	}
	if resolved.Driver.Mechanism() != "cmux" {
		cfg.Spec.SessionID = priorSessionID(todoList)
	}
	return cfg, resolved, nil
}

// triageBacklog loads the other open TODOs a triage run compares against for
// duplicates. A backlog that cannot be listed is not fatal: triage's other four
// verdicts do not depend on it, so the run proceeds without the section rather
// than failing outright.
func triageBacklog(ctx context.Context, provider todos.Provider, todoList []*types.TODO) (string, error) {
	if provider == nil {
		return "", nil
	}
	candidates, err := provider.List(ctx, todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		logger.Warnf("triage duplicate detection is degraded: could not list the backlog: %v", err)
		return "", nil
	}
	return todos.BuildBacklogIndex(candidates, todoList), nil
}

// priorSessionID returns the first recorded agent session among todoList, which
// the non-cmux drivers resume into.
func priorSessionID(todoList []*types.TODO) string {
	for _, todo := range todoList {
		if todo != nil && todo.LLM != nil && todo.LLM.SessionId != "" {
			return todo.LLM.SessionId
		}
	}
	return ""
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

		// The deadline comes from the same resolution that built the spec, so the
		// run's budget and the context that cancels it cannot disagree.
		executor, sessionID, timeout, err := newExecutor(workDir, group.TODOs, provider)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
		todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)
		todoExec.SetMode(todosRunMode)
		todoExec.SetResume(resumeSession)
		todoExec.SetConcurrent(forceRun)

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

		result, execErr, interrupted, err := executeTODO(workDir, todo, interaction, provider, forceRun)
		if err != nil {
			return err
		}
		// A live run on another process is a question, not a failure: ask, and
		// dispatch alongside it when that is the answer.
		var owned *todos.ErrRunOwnedElsewhere
		if errors.As(execErr, &owned) && !forceRun && run.ConfirmConcurrent(context.Background(), owned.Error()) {
			if result, execErr, interrupted, err = executeTODO(workDir, todo, interaction, provider, true); err != nil {
				return err
			}
		}

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

// executeTODO runs one TODO to completion under the interrupt handler. The
// separate `err` is a setup failure that aborts the whole command, where
// execErr is this TODO's own outcome.
func executeTODO(
	workDir string,
	todo *types.TODO,
	interaction *todos.UserInteraction,
	provider todos.Provider,
	concurrent bool,
) (result *todos.ExecutionResult, execErr error, interrupted bool, err error) {
	executor, sessionID, timeout, err := newExecutor(workDir, []*types.TODO{todo}, provider)
	if err != nil {
		return nil, nil, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), interaction)
	todoExec := todos.NewTODOExecutor(workDir, executor, sessionID, provider)
	todoExec.SetMode(todosRunMode)
	todoExec.SetResume(resumeSession)
	todoExec.SetConcurrent(concurrent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

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
	return result, execErr, interrupted, nil
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

func dryRunTODOs(groups []todos.TODOGroup, workDir string, resolved todospec.Resolved, provider todos.Provider) error {
	isGrouped := len(groups) > 1 || (len(groups) == 1 && groups[0].Name != "")

	for _, group := range groups {
		if len(group.TODOs) == 0 {
			continue
		}

		if resolved.Driver.Mechanism() == "cmux" {
			if err := printCmuxDryRun(group, workDir, resolved.Spec.Model, provider); err != nil {
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

func printCmuxDryRun(group todos.TODOGroup, workDir string, model api.Model, provider todos.Provider) error {
	groupWorkDir := workDir
	if group.Name != "" && group.Name != todos.UngroupedLabel && filepath.IsAbs(group.Name) {
		groupWorkDir = group.Name
	}
	if model.Provider == nil {
		return fmt.Errorf("resolved model %q is missing its Captain provider", model.Name)
	}
	agent := model.Provider.AgentName
	sessionID := ""
	if agent == "claude" {
		sessionID = "<session-id>"
	}
	agentCmd := cmuxprov.AgentCommand(cmuxprov.AgentCommandOpts{Agent: agent, Model: model.Name, SessionID: sessionID})
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
	cfg, _, err := newAgentRunConfig(context.Background(), workDir, todoList, provider)
	if err != nil {
		return "", err
	}
	req, _, err := todoprompt.Render(todoList, cfg.PromptOptions(workDir))
	if err != nil {
		return "", err
	}
	return req.Prompt.User, nil
}

func effortDirective(effort string) string {
	return todoprompt.EffortDirective(effort)
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
