package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

// retiredTodoRunFlags maps each flag `todos run` no longer accepts to what
// replaced it, so a stale invocation is answered with the replacement rather
// than cobra's bare "unknown flag". The MANUAL's retired-flags table is the
// prose form of this map.
var retiredTodoRunFlags = map[string]string{
	"driver":   "the compact model form, e.g. --model cli:opus:high",
	"mode":     "--step (a run is one lifecycle step for one todo)",
	"prompt":   "--step",
	"group-by": "--step (grouping is gone; runs dispatch per todo)",
	"check":    ".gavel.yaml checks.enabled (loop budget: todos.run.workflow.verify.maxIterations)",
}

// retiredFlagError rewrites cobra's unknown-flag error for a retired flag into
// one that names its replacement; any other flag error passes through.
func retiredFlagError(cmd *cobra.Command, err error) error {
	var name string
	if _, scanErr := fmt.Sscanf(err.Error(), "unknown flag: --%s", &name); scanErr != nil {
		return err
	}
	replacement, retired := retiredTodoRunFlags[name]
	if !retired {
		return err
	}
	return fmt.Errorf("--%s was retired from %s; use %s", name, cmd.CommandPath(), replacement)
}

func init() {
	todosRunCmd.SetFlagErrorFunc(retiredFlagError)
	todosRunCmd.Flags().StringVar(&filterStatus, "status", "", "Filter TODOs by status (pending, in_progress, failed)")
	todosRunCmd.Flags().StringVar(&todosStep, "step", "",
		"Lifecycle step to run: run, plan, verify, triage, or any step the project's lifecycle declares "+
			"(empty: the step the lifecycle picks next for each todo); 'gavel todos steps' lists them")
	todosRunCmd.Flags().Float64Var(&maxBudget, "max-budget", 0, "Maximum budget in USD")
	todosRunCmd.Flags().IntVar(&maxTurns, "max-turns", 0, "Maximum conversation turns")
	todosRunCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Interactively select TODOs to run")
	todosRunCmd.Flags().StringVar(&todoModel, "model", "",
		"LLM model override for TODO execution, as the compact mode:model:effort form (e.g. cli:opus:high); "+
			"empty uses the step's .prompt frontmatter default")
	// Empty, not "medium": the flag is the highest resolution layer, so a non-zero
	// default would beat the .prompt frontmatter it claims to defer to. The
	// lifecycle host applies medium once nothing else has spoken.
	todosRunCmd.Flags().StringVar(&todoEffort, "effort", "", "Reasoning effort directive: low, medium, high, or xhigh (empty: the step's .prompt frontmatter, else medium)")
	todosRunCmd.Flags().BoolVar(&resumeSession, "resume", false, "Resume the TODO's prior agent session instead of starting a fresh one")
	todosRunCmd.Flags().BoolVar(&forceRun, "force", false, "Dispatch even when the TODO already has a live run on another process: the two runs proceed in parallel (without it you are asked)")
	todosRunCmd.Flags().BoolVar(&dirty, "dirty", false, "Run in a new worktree that carries the working tree's uncommitted and gitignored content across, declaring one when the project configures no checkout")
	todosRunCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the rendered prompt, the spec layer stack and the resolved spec without dispatching")
	todosRunCmd.Flags().BoolVar(&commitAfter, "commit", true, "Let the run commit its work as the lifecycle step declares (use --commit=false to refuse a committing run)")
}

// errRunInterrupted is a run the operator interrupted; the loop stops there.
var errRunInterrupted = errors.New("todo run interrupted")

// todoRunFailure is one todo's run failing after it resolved. It is a failure
// of that run, not of the command: the loop reports it and moves on, where a
// resolution error — configuration the whole command shares — aborts.
type todoRunFailure struct {
	todo *types.TODO
	err  error
}

func (f *todoRunFailure) Error() string { return fmt.Sprintf("todo %s: %v", run.Label(f.todo), f.err) }
func (f *todoRunFailure) Unwrap() error { return f.err }

func runTodosRun(_ *cobra.Command, args []string) error {
	if err := validateTodosRunOptions(); err != nil {
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
	ctx := context.Background()
	logger.Infof("Discovering TODOs from PostgreSQL")
	filters := todos.DiscoveryFilters{ExcludeStatuses: []types.Status{types.StatusCompleted}}
	if filterStatus != "" {
		filters.IncludeStatuses = []types.Status{types.Status(filterStatus)}
	}
	todoList, err := resolveRequestedTODOs(ctx, provider, args, filters)
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
	logger.Infof("Found %d TODOs", len(todoList))

	for i, todo := range todoList {
		fmt.Println(clicky.Text(fmt.Sprintf("=== TODO %d/%d: %s ===", i+1, len(todoList), todo.Filename()), "text-blue-600 font-bold").ANSI())
		err := runTodoStep(ctx, workDir, provider, todo, todosRunOptions())
		var failure *todoRunFailure
		switch {
		case err == nil:
		case errors.Is(err, errRunInterrupted):
			fmt.Println(clicky.Text("Execution interrupted by user", "text-red-600 font-bold").ANSI())
			return nil
		case errors.As(err, &failure):
			logger.Errorf("TODO execution failed: %v", failure.err)
		default:
			return err
		}
	}
	fmt.Println()
	fmt.Println(clicky.MustFormat(clicky.Text("All TODOs processed", "text-blue-600 font-bold")))
	return nil
}

// todosRunOptions is what the run flags decide; the lifecycle decides the rest.
func todosRunOptions() run.Options {
	return run.Options{
		Step:       todosStep,
		Request:    todosRequestSpec(),
		Resume:     resumeSession,
		Concurrent: forceRun,
		// The CLI drains no approval queue: a run that asked for one would block
		// until its timeout, so it contributes no approval-brokering posture.
		Host: lifecycle.HostCLI,
	}
}

// todosRequestSpec is the request layer: only the flags that were actually
// set. It is the highest layer of the fold, so an unset flag must contribute
// nothing — a non-zero default here would silently beat the .prompt
// frontmatter and the lifecycle step it claims to defer to.
//
// The configured `checks` suite is not a request-layer toggle: the definition
// of done is rendered from the todo and `.gavel.yaml checks.enabled` (or the
// todo's `checks:` front matter), and no run spec is consulted for it.
func todosRequestSpec() api.Spec {
	var request api.Spec
	request.Name = todoModel
	request.Effort = api.Effort(todoEffort)
	request.Budget.Cost = maxBudget
	request.Budget.MaxTurns = maxTurns
	// --dirty is a request-layer shorthand for the worktree posture, not a patch
	// applied to an already-resolved checkout: a project that declares no
	// checkout block still gets one, so the flag can never be a silent no-op.
	if dirty {
		request.Setup = &shell.Setup{Checkout: &shell.Checkout{Worktree: &shell.Worktree{
			Mode:        shell.WorktreeNew,
			Uncommitted: shell.CloneClone,
			Ignored:     shell.CloneClone,
		}}}
	}
	return request
}

func validateTodosRunOptions() error {
	switch todoEffort {
	case "", "low", "medium", "high", "xhigh":
	default:
		return fmt.Errorf("invalid --effort %q: expected low, medium, high, or xhigh", todoEffort)
	}
	return nil
}

// assertRunCommitPolicy refuses `--commit=false` on a run whose resolved spec
// declares commits. Captain's spec merge reads an empty slice as "not stated",
// so a request layer cannot clear a lifecycle step's `commits:` — dispatching
// anyway would commit against an explicit instruction not to.
func assertRunCommitPolicy(prepared *run.Prepared) error {
	if commitAfter || !run.Commit(prepared.Resolution.Spec) {
		return nil
	}
	return fmt.Errorf(
		"--commit=false: the %s step declares commits and a request layer cannot clear them; drop the flag, or configure todos.%s.workflow without commits",
		prepared.Step.Name, prepared.Step.Name)
}

// runTodoStep runs one lifecycle step for one todo, end to end: the step
// chosen and reported, the run admitted, and — unless previewing — awaited to
// its outcome. It is the one path `todos run` and the plan actions dispatch
// through, so the two cannot disagree about what a run is.
func runTodoStep(ctx context.Context, workDir string, provider todos.Provider, todo *types.TODO, opts run.Options) error {
	req := run.Request{Provider: provider, Registry: run.Shared(), Todo: todo, Dir: workDir, Options: opts}
	prepared, err := run.Resolve(ctx, req)
	if err != nil {
		return err
	}
	req.Prepared = prepared
	fmt.Println(clicky.Text("step: "+prepared.Step.Name, "text-green-600 font-bold").
		Append("  "+prepared.Reason, "text-gray-500").ANSI())
	if dryRun {
		return printDryRun(prepared)
	}
	if err := assertRunCommitPolicy(prepared); err != nil {
		return err
	}
	started, err := run.Start(req)
	// A live run on another process is a question, not a failure: ask, and
	// dispatch alongside it when that is the answer.
	var owned *todos.ErrRunOwnedElsewhere
	if errors.As(err, &owned) && !opts.Concurrent && run.ConfirmConcurrent(ctx, owned.Error()) {
		req.Options.Concurrent = true
		started, err = run.Start(req)
	}
	if err != nil {
		return &todoRunFailure{todo: todo, err: err}
	}
	if err := awaitRun(started); err != nil {
		if errors.Is(err, errRunInterrupted) {
			return err
		}
		return &todoRunFailure{todo: todo, err: err}
	}
	fmt.Println(clicky.Text(fmt.Sprintf("%s finished — %s", prepared.Step.Name, todo.Status), "text-green-600").ANSI())
	return nil
}

// awaitRun blocks until a started run reports its outcome, honouring an
// interrupt by stopping the run and giving it a moment to drain. A result with
// nothing to wait for — a run that already finished — returns at once.
func awaitRun(started run.StartResult) error {
	if started.Done == nil {
		return nil
	}
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(interrupts)
	select {
	case err := <-started.Done:
		return err
	case sig := <-interrupts:
		logger.Warnf("Received signal %v, shutting down gracefully...", sig)
		if started.Stop != nil {
			started.Stop(todos.ErrExecutionCancelled)
		}
		fmt.Println(clicky.Text("Interrupted - cleaning up...", "text-yellow-600 font-bold").ANSI())
		select {
		case <-started.Done:
		case <-time.After(5 * time.Second):
			logger.Warnf("Timeout waiting for graceful shutdown")
		}
		return errRunInterrupted
	}
}

// printDryRun shows exactly the run that would follow: the rendered prompt,
// the layer stack that produced the spec (lowest precedence first) and the
// resolved spec itself.
func printDryRun(prepared *run.Prepared) error {
	resolution := prepared.Resolution
	fmt.Println("### Prompt")
	fmt.Println(resolution.Prompt)
	fmt.Println("### Layers (lowest precedence first)")
	for _, layer := range resolution.Trace {
		fmt.Printf("  %-40s %s\n", layer.Name, layer.Scope)
	}
	encoded, err := yaml.Marshal(resolution.Spec)
	if err != nil {
		return fmt.Errorf("marshal resolved Captain spec as YAML: %w", err)
	}
	fmt.Println("### Spec")
	fmt.Println(string(encoded))
	return nil
}
