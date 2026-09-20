package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

type TodosPlanApproveOptions struct {
	TodoTargetOptions
	Run bool `flag:"run" help:"Implement the approved plan immediately"`
}

type TodosPlanReviseOptions struct {
	TodoTargetOptions
	Feedback string `flag:"feedback" required:"true" help:"Change request for the agent to fold into the plan"`
}

var todosPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Act on or recover a todo plan",
	Long: `Act on a plan produced by 'gavel todos run --step plan' that is awaiting review.
Subcommands: approve (the plan becomes the one a run implements, optionally
chaining that run with --run), reject (the plan remains in history but is no
longer active, so the pending todo can be planned again) and revise (agent folds
your --feedback into the plan and returns it to review). Recover backfills the plan file from the todo's latest terminal
plan-step run; to write one yourself use 'gavel todos edit --plan'.`,
	Example: `  gavel todos plan approve 3f2a1b --run
  gavel todos plan reject 3f2a1b
  gavel todos plan revise 3f2a1b --feedback "split the migration into two steps"
  gavel todos plan recover 3f2a1b`,
}

var todosPlanApproveCmd *cobra.Command
var todosPlanRejectCmd *cobra.Command
var todosPlanReviseCmd *cobra.Command

func init() {
	todosCmd.AddCommand(todosPlanCmd)
	todosPlanApproveCmd = clicky.AddNamedCommand("approve", todosPlanCmd, TodosPlanApproveOptions{}, func(opts TodosPlanApproveOptions) (any, error) { return nil, runTodosPlanApprove(opts) })
	todosPlanRejectCmd = clicky.AddNamedCommand("reject", todosPlanCmd, TodosGetOptions{}, func(opts TodosGetOptions) (any, error) { return nil, runTodosPlanReject(opts) })
	todosPlanReviseCmd = clicky.AddNamedCommand("revise", todosPlanCmd, TodosPlanReviseOptions{}, func(opts TodosPlanReviseOptions) (any, error) { return nil, runTodosPlanRevise(opts) })
	clicky.AddNamedCommand("recover", todosPlanCmd, TodosGetOptions{}, func(opts TodosGetOptions) (any, error) { return nil, runTodosPlanRecover(opts) })
}

func resolvePlanTODO(ctx context.Context, args []string) (string, todos.Provider, *types.TODO, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return "", nil, nil, err
	}
	todoList, err := resolveRequestedTODOs(ctx, provider, args, todos.DiscoveryFilters{})
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to discover TODOs: %w", err)
	}
	if len(todoList) != 1 {
		return "", nil, nil, fmt.Errorf("expected exactly one TODO matching %q, found %d", args[0], len(todoList))
	}
	return workDir, provider, todoList[0], nil
}

// resolveReviewTODO loads the single todo named by args and confirms it is in
// review — the precondition for both plan actions.
func resolveReviewTODO(ctx context.Context, args []string) (string, todos.Provider, *types.TODO, error) {
	workDir, provider, todo, err := resolvePlanTODO(ctx, args)
	if err != nil {
		return "", nil, nil, err
	}
	if todo.Status != types.StatusReview {
		return "", nil, nil, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status)
	}
	return workDir, provider, todo, nil
}

func runTodosPlanRecover(target TodosGetOptions) error {
	ref, err := target.One()
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, provider, todo, err := resolvePlanTODO(ctx, []string{ref})
	if err != nil {
		return err
	}
	recovery, ok := provider.(todos.PlanRecoveryProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan recovery")
	}
	todo, err = recovery.RecoverPlan(ctx, todo)
	if err != nil {
		return err
	}
	logger.Infof("Recovered plan for %s — %s", todo.Filename(), todo.Status.Pretty().ANSI())
	return nil
}

// runApprovedTODO dispatches the implement run that --run chains; a var so the
// approval transition can be tested without spawning an agent.
var runApprovedTODO = func(workDir string, todo *types.TODO, provider todos.Provider, opts run.Options) error {
	return runTodoStep(context.Background(), workDir, provider, todo, opts, runStepPolicy{AllowCommit: true})
}

func runTodosPlanApprove(target TodosPlanApproveOptions) error {
	ref, err := target.One()
	if err != nil {
		return err
	}
	ctx := context.Background()
	workDir, provider, todo, err := resolveReviewTODO(ctx, []string{ref})
	if err != nil {
		return err
	}
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan review")
	}
	// Read the plan's prompt run before approving: the approval retires it, and it
	// is the record of the runtime the implement run continues.
	planRun, err := run.PriorRun(ctx, provider, todo)
	if err != nil {
		return err
	}
	// The implement run is placed in the lifecycle before the plan is approved,
	// so a continuation the lifecycle refuses changes nothing. An approved plan
	// is implemented, not re-planned, and the implement run is a fresh turn
	// rather than a continuation of the planning conversation — the approved
	// plan reaches it through the provider, not through session history. Only
	// the runtime the plan run resolved carries: a codex plan is implemented by
	// codex. Everything else resolves from .gavel.yaml exactly as 'gavel todos
	// run' would.
	var opts run.Options
	if target.Run {
		opts, err = run.Continue(run.Continuation{
			Dir: workDir, Provider: provider, Todo: todo, Prior: planRun,
			Override: run.Options{Host: lifecycle.HostCLI}, Step: string(types.RunPhase),
		})
		if err != nil {
			return err
		}
	}
	todo, err = reviewer.ApprovePlan(ctx, todo, "gavel-cli", "")
	if err != nil {
		return err
	}
	logger.Infof("Approved plan for %s", todo.Filename())
	if !target.Run {
		logger.Infof("Run it with: gavel todos run %s", todo.Filename())
		return nil
	}
	return runApprovedTODO(workDir, todo, provider, opts)
}

func runTodosPlanReject(target TodosGetOptions) error {
	ref, err := target.One()
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, provider, todo, err := resolveReviewTODO(ctx, []string{ref})
	if err != nil {
		return err
	}
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan review")
	}
	todo, err = reviewer.RejectPlan(ctx, todo, "gavel-cli", "")
	if err != nil {
		return err
	}
	logger.Infof("Rejected plan for %s — back to pending", todo.Filename())
	return nil
}

func runTodosPlanRevise(target TodosPlanReviseOptions) error {
	ref, err := target.One()
	if err != nil {
		return err
	}
	feedback := strings.TrimSpace(target.Feedback)
	if feedback == "" {
		return fmt.Errorf("--feedback is required")
	}
	ctx := context.Background()
	workDir, provider, todo, err := resolveReviewTODO(ctx, []string{ref})
	if err != nil {
		return err
	}
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan review")
	}
	// Read the plan's prompt run before the revision transition retires it.
	planRun, err := run.PriorRun(ctx, provider, todo)
	if err != nil {
		return err
	}
	// Revise re-enters the plan step, continuing the plan run's own
	// configuration and the runtime it resolved. With a recorded session it
	// resumes that session with the feedback as the next turn, so the agent
	// updates its plan and the step's outcome lands the todo back in review;
	// without one it plans afresh with the feedback in the todo's own prompt.
	c := run.Continuation{
		Dir: workDir, Provider: provider, Todo: todo, Prior: planRun,
		Override: run.Options{Host: lifecycle.HostCLI}, Step: string(types.PlanPhase),
	}
	resume := run.PriorSessionID(todo) != ""
	if resume {
		c.Resume, c.Message = true, feedback
	}
	opts, err := run.Continue(c)
	if err != nil {
		return err
	}
	todo, err = reviewer.RequestPlanRevision(ctx, todo, "gavel-cli", feedback)
	if err != nil {
		return err
	}
	if !resume {
		todo.Prompt = "Revise the existing plan using this reviewer feedback:\n\n" + feedback
	}
	if err := runTodoStep(ctx, workDir, provider, todo, opts, runStepPolicy{AllowCommit: true}); err != nil {
		return fmt.Errorf("revise plan: %w", err)
	}
	logger.Infof("Revised plan for %s — %s", todo.Filename(), todo.Status.Pretty().ANSI())
	return nil
}
