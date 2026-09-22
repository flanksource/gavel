package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
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

type TodosRunOptions struct {
	TodoTargetOptions
	Step           string   `flag:"step" help:"The lifecycle step to run; gavel todos steps lists available steps"`
	RuntimeProfile string   `flag:"runtime-profile" help:"Retired; use --preset"`
	Presets        []string `flag:"preset" help:"Runtime preset name or ID; repeat to layer presets"`
	NoPresets      bool     `flag:"no-presets" help:"Clear prompt and project runtime preset defaults"`
	MaxBudget      float64  `flag:"max-budget" help:"Maximum budget in USD"`
	MaxTurns       int      `flag:"max-turns" help:"Maximum conversation turns"`
	Model          string   `flag:"model" help:"LLM model override as mode:model:effort"`
	Effort         string   `flag:"effort" help:"Reasoning effort: low, medium, high, or xhigh"`
	Resume         bool     `flag:"resume" help:"Resume the TODO's prior agent session"`
	Force          bool     `flag:"force" help:"Dispatch despite another live run"`
	Dirty          bool     `flag:"dirty" help:"Carry uncommitted and ignored files into a new worktree"`
	DryRun         bool     `flag:"dry-run" help:"Print the prompt and resolved spec without dispatching"`
	Preview        bool     `flag:"preview" help:"Run the agent, but report a triage verdict that would close a TODO instead of applying it (unlike --dry-run, which never runs the agent)"`
	Commit         bool     `flag:"commit" default:"true" help:"Allow committing lifecycle steps"`
}

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
	todosRunCmd = clicky.AddNamedCommand("run", todosCmd, TodosRunOptions{}, func(opts TodosRunOptions) (any, error) { return nil, runTodosRun(opts) })
	todosRunCmd.Use = "run <id>..."
	todosRunCmd.Short = "Run a lifecycle step for explicit TODO IDs"
	todosRunCmd.SetFlagErrorFunc(retiredFlagError)
	_ = todosRunCmd.Flags().MarkDeprecated("runtime-profile", "runtime profiles are ignored; use --preset")
}

// lifecycleStepNames is the project's declared step vocabulary, loaded without a
// provider because naming a step needs the lifecycle, not the database.
func lifecycleStepNames(workDir string) ([]string, error) {
	host, err := lifecycle.NewHost(nil, workDir, lifecycle.HostCLI)
	if err != nil {
		return nil, err
	}
	return host.Def.Definition().StepNames(), nil
}

// splitLeadingStep accepts `todos run <step> <id>...` alongside
// `todos run <id>... --step <step>`.
//
// `todos run triage <id>` is what the command's own help invites — it documents
// the lifecycle by step name — and without this the step name is resolved as a
// TODO reference, failing with "short issue reference must contain at least 8
// characters", which names neither the argument nor the real problem.
//
// The two forms cannot collide: a leading word is taken as the step only when it
// matches a declared step exactly, --step was not given, and at least one
// argument remains to be a TODO. A short id is 8 hex characters, so it never
// matches a step name, and an explicit --step always wins.
func splitLeadingStep(step string, args, steps []string) (string, []string) {
	if strings.TrimSpace(step) != "" || len(args) < 2 {
		return step, args
	}
	for _, name := range steps {
		if args[0] == name {
			return name, args[1:]
		}
	}
	return step, args
}

// misplacedStepHint explains a resolution failure caused by a step name sitting
// where a TODO reference belongs — `todos run <id> triage`, or a leading step
// that --step already claimed.
//
// It is conditional because the common failure is an id that simply does not
// exist, and a lecture about step vocabulary on every one of those trains readers
// to skip the line that matters.
func misplacedStepHint(args, steps []string) string {
	for _, arg := range args {
		for _, step := range steps {
			if arg == step {
				return fmt.Sprintf("\n(%q is a lifecycle step, not a TODO: gavel todos run %s <id>)", arg, arg)
			}
		}
	}
	return ""
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

func runTodosRun(opts TodosRunOptions) error {
	args, err := opts.Many()
	if err != nil {
		return err
	}
	if err := validateTodosRunOptions(opts); err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	steps, err := lifecycleStepNames(workDir)
	if err != nil {
		return err
	}
	step, args := splitLeadingStep(opts.Step, args, steps)
	opts.Step = step
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	logger.Infof("Discovering TODOs from PostgreSQL")
	// Targets rather than bare TODOs: a UUID resolves across workspaces, and a run
	// must dispatch through the provider and working directory that own it.
	targets, err := resolveRequestedTargets(ctx, provider, workDir, args, todos.DiscoveryFilters{})
	if err != nil {
		return fmt.Errorf("failed to discover TODOs: %w%s", err, misplacedStepHint(args, steps))
	}
	if len(targets) == 0 {
		logger.Infof("No TODOs found")
		return nil
	}
	logger.Infof("Found %d TODOs", len(targets))

	// Every run in this invocation is told the whole selection, so a triage pass
	// can see which backlog entries are having their verdicts decided alongside it.
	batch := make([]string, 0, len(targets))
	for _, target := range targets {
		batch = append(batch, todos.TODOReference(target.Todo))
	}

	for i, target := range targets {
		todo := target.Todo
		fmt.Println(clicky.Text(fmt.Sprintf("=== TODO %d/%d: %s ===", i+1, len(targets), todo.Filename()), "text-blue-600 font-bold").ANSI())
		runOpts := todosRunOptions(opts)
		runOpts.Batch = batch
		err := runTodoStep(ctx, target.WorkDir, target.Provider, todo, runOpts, runStepPolicy{DryRun: opts.DryRun, AllowCommit: opts.Commit})
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
func todosRunOptions(opts TodosRunOptions) run.Options {
	presets := append([]string(nil), opts.Presets...)
	if opts.NoPresets {
		presets = []string{}
	}
	return run.Options{
		Presets:        presets,
		PresetsSet:     opts.NoPresets || opts.Presets != nil,
		RuntimeProfile: opts.RuntimeProfile,
		Step:           opts.Step,
		Request:        todosRequestSpec(opts),
		Resume:         opts.Resume,
		Concurrent:     opts.Force,
		Preview:        opts.Preview,
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
func todosRequestSpec(opts TodosRunOptions) api.Spec {
	var request api.Spec
	request.Name = opts.Model
	request.Effort = api.Effort(opts.Effort)
	request.Budget.Cost = opts.MaxBudget
	request.Budget.MaxTurns = opts.MaxTurns
	// --dirty is a request-layer shorthand for the worktree posture, not a patch
	// applied to an already-resolved checkout: a project that declares no
	// checkout block still gets one, so the flag can never be a silent no-op.
	if opts.Dirty {
		request.Setup = &shell.Setup{Checkout: &shell.Checkout{Worktree: &shell.Worktree{
			Mode:        shell.WorktreeNew,
			Uncommitted: shell.CloneClone,
			Ignored:     shell.CloneClone,
		}}}
	}
	return request
}

func validateTodosRunOptions(opts TodosRunOptions) error {
	if opts.NoPresets && len(opts.Presets) > 0 {
		return fmt.Errorf("--preset and --no-presets are mutually exclusive")
	}
	switch opts.Effort {
	case "", "low", "medium", "high", "xhigh":
	default:
		return fmt.Errorf("invalid --effort %q: expected low, medium, high, or xhigh", opts.Effort)
	}
	return nil
}

// assertRunCommitPolicy refuses `--commit=false` on a run whose resolved spec
// declares commits. Captain's spec merge reads an empty slice as "not stated",
// so a request layer cannot clear a lifecycle step's `commits:` — dispatching
// anyway would commit against an explicit instruction not to.
func assertRunCommitPolicy(prepared *run.Prepared, allowCommit bool) error {
	if allowCommit || !run.Commit(prepared.Resolution.Spec) {
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
type runStepPolicy struct {
	DryRun      bool
	AllowCommit bool
}

func runTodoStep(ctx context.Context, workDir string, provider todos.Provider, todo *types.TODO, opts run.Options, policy runStepPolicy) error {
	req := run.Request{Provider: provider, Registry: run.Shared(), Todo: todo, Dir: workDir, Options: opts}
	prepared, err := run.Resolve(ctx, req)
	if err != nil {
		return err
	}
	req.Prepared = prepared
	// The report is rendered from the run's own result, before its outcome is
	// persisted: a failed SaveAttempt must not be able to swallow what the agent
	// said, and the status printed is the one the lifecycle decided rather than one
	// read back off a TODO the write may never have reached.
	req.OnComplete = func(outcome *lifecycle.StepOutcome, status string, runErr error) {
		printRunResult(prepared.Step.Name, outcome, status, runErr)
	}
	fmt.Println(clicky.Text("step: "+prepared.Step.Name, "text-green-600 font-bold").
		Append("  "+prepared.Reason, "text-gray-500").ANSI())
	if policy.DryRun {
		return printDryRun(prepared)
	}
	if err := assertRunCommitPolicy(prepared, policy.AllowCommit); err != nil {
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
	// No closing line here: OnComplete already reported the step, its status and
	// what the agent said, and a second line reading off todo.Status would only
	// repeat it — differently, whenever the two disagree.
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
	for _, warning := range resolution.Warnings {
		fmt.Printf("Warning: %s\n", warning)
	}
	if resolution.RuntimePresets != nil {
		fmt.Println("Runtime presets:")
		for _, preset := range resolution.RuntimePresets.Presets {
			fmt.Printf("  %s (%s)\n", preset.Name, preset.ID)
		}
	}
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
