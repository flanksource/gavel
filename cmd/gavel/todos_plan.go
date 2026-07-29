package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	planReviseFeedback string
	planApproveRun     bool
)

var todosPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Act on a reviewed plan (approve, reject or revise) — the CLI side of the dashboard plan-review actions",
	Long: `Act on a plan produced by 'gavel todos run --mode plan' that is awaiting review.
Subcommands: approve (the plan becomes the one a run implements, optionally
chaining that run with --run), reject (todo returns to pending, plan cleared) and
revise (agent folds your --feedback into the plan and returns it to review).`,
	Example: `  gavel todos plan approve 3f2a1b --run
  gavel todos plan reject 3f2a1b
  gavel todos plan revise 3f2a1b --feedback "split the migration into two steps"`,
}

var todosPlanApproveCmd = &cobra.Command{
	Use:   "approve <todo>",
	Short: "Approve a reviewed plan so a run can implement it; --run starts that run immediately",
	Args:  cobra.ExactArgs(1),
	RunE:  runTodosPlanApprove,
}

var todosPlanRejectCmd = &cobra.Command{
	Use:   "reject <todo>",
	Short: "Reject a reviewed plan: the todo returns to pending and its plan pointer is cleared",
	Args:  cobra.ExactArgs(1),
	RunE:  runTodosPlanReject,
}

var todosPlanReviseCmd = &cobra.Command{
	Use:   "revise <todo>",
	Short: "Ask the agent to revise a reviewed plan with feedback; the plan session resumes and the todo returns to review",
	Args:  cobra.ExactArgs(1),
	RunE:  runTodosPlanRevise,
}

func init() {
	todosCmd.AddCommand(todosPlanCmd)
	todosPlanCmd.AddCommand(todosPlanApproveCmd)
	todosPlanCmd.AddCommand(todosPlanRejectCmd)
	todosPlanCmd.AddCommand(todosPlanReviseCmd)
	todosPlanApproveCmd.Flags().BoolVar(&planApproveRun, "run", false, "Implement the approved plan immediately, as 'gavel todos run' would")
	todosPlanReviseCmd.Flags().StringVar(&planReviseFeedback, "feedback", "", "The change request for the agent to fold into the plan (required)")
}

// resolveReviewTODO loads the single todo named by args and confirms it is in
// review — the precondition for both plan actions.
func resolveReviewTODO(ctx context.Context, args []string) (string, todos.Provider, *types.TODO, error) {
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
	todo := todoList[0]
	if todo.Status != types.StatusReview {
		return "", nil, nil, fmt.Errorf("todo is not awaiting plan review (status: %s)", todo.Status)
	}
	return workDir, provider, todo, nil
}

// runApprovedTODO dispatches the implement run that --run chains; a var so the
// approval transition can be tested without spawning an agent.
var runApprovedTODO = executeSingleTODOs

func runTodosPlanApprove(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	workDir, provider, todo, err := resolveReviewTODO(ctx, args)
	if err != nil {
		return err
	}
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan review")
	}
	todo, err = reviewer.ApprovePlan(ctx, todo, "gavel-cli", "")
	if err != nil {
		return err
	}
	logger.Infof("Approved plan for %s", todo.Filename())
	if !planApproveRun {
		logger.Infof("Run it with: gavel todos run %s", todo.Filename())
		return nil
	}

	// An approved plan is implemented, not re-planned, and the implement run is a
	// fresh turn rather than a continuation of the planning conversation — the
	// approved plan reaches it through the provider (lifecycle loads PlanMarkdown
	// for ModeRun), not through session history. Everything else resolves from
	// .gavel.yaml exactly as 'gavel todos run' would.
	todosRunMode = types.ModeRun
	resumeSession = false
	return runApprovedTODO(workDir, types.TODOS{todo}, newInteraction(), provider)
}

func runTodosPlanReject(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	_, provider, todo, err := resolveReviewTODO(ctx, args)
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

func runTodosPlanRevise(_ *cobra.Command, args []string) error {
	feedback := strings.TrimSpace(planReviseFeedback)
	if feedback == "" {
		return fmt.Errorf("--feedback is required")
	}
	ctx := context.Background()
	workDir, provider, todo, err := resolveReviewTODO(ctx, args)
	if err != nil {
		return err
	}
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		return fmt.Errorf("todo has no recorded plan session to revise")
	}
	reviewer, ok := provider.(todos.PlanReviewProvider)
	if !ok {
		return fmt.Errorf("native PostgreSQL TODO provider does not support durable plan review")
	}
	todo, err = reviewer.RequestPlanRevision(ctx, todo, "gavel-cli", feedback)
	if err != nil {
		return err
	}

	// Resume the plan session in plan mode so the agent updates its native plan
	// file and the run lands back in review (applyPlanOutcome). Going through
	// newExecutor is what gives a revise the same .gavel.yaml resolution — model,
	// budget, timeout, driver — as the plan it revises.
	//
	// workDir is the un-joined discovery root: the todo's CWD is joined exactly
	// once, downstream in headless.groupWorkDir.
	todosRunMode = types.ModePlan
	resumeSession = true
	executor, sessionID, timeout, err := newExecutor(workDir, []*types.TODO{todo}, provider)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	runner := todos.NewTODOExecutor(workDir, executor, sessionID, provider)
	runner.SetMode(types.ModePlan)
	execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), nil)
	if _, err := runner.Resume(execCtx, []*types.TODO{todo}, feedback); err != nil {
		return fmt.Errorf("revise plan: %w", err)
	}
	logger.Infof("Revised plan for %s — %s", todo.Filename(), todo.Status.Pretty().ANSI())
	return nil
}
