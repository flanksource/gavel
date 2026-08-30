package run

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	todospec "github.com/flanksource/gavel/todos/spec"
	"github.com/flanksource/gavel/todos/types"
)

// Start admits a run and returns as soon as the agent session is prepared; the
// run itself continues in the background. It is a var so tests can substitute a
// stub without standing up an agent.
// Start reports todos.ErrRunOwnedElsewhere rather than resolving it: whether to
// run alongside a live run is the caller's question to ask, and its callers are
// HTTP handlers that must not block on a dialog. Terminal callers ask with
// ConfirmConcurrent and retry with Options.Concurrent set.
var Start = defaultStart

func defaultStart(req Request) (StartResult, error) {
	if req.Registry == nil {
		return StartResult{}, errors.New("todo run registry is required")
	}
	executor, sessionID, err := NewExecutor(req)
	if err != nil {
		return StartResult{}, err
	}
	// The run outlives the call, so it gets its own context rather than the
	// caller's — a request finishing must not cancel the agent it started.
	ctx, timeoutCancel := context.WithTimeout(context.Background(), req.Options.Timeout)
	runCtx, stop := context.WithCancelCause(ctx)
	handle, err := req.Registry.Register(RegisterOptions{
		IssueIDs:   IssueIDs(req.Todos),
		Stoppable:  IsStoppable(req.Options),
		Concurrent: req.Options.Concurrent,
		Cancel:     stop,
	})
	if err != nil {
		timeoutCancel()
		return StartResult{}, err
	}
	type startOutcome struct {
		result StartResult
		err    error
	}
	started := make(chan startOutcome, 1)
	var notifyOnce sync.Once
	notify := func(result StartResult, err error) {
		notifyOnce.Do(func() { started <- startOutcome{result: result, err: err} })
	}
	go func() {
		defer timeoutCancel()
		defer handle.Release()

		execCtx := todos.NewExecutorContext(runCtx, logger.StandardLogger(), nil)
		execCtx.SetRunPreparedHook(func(preparation todos.RunPreparationResult) {
			// The run is stoppable by the identity Captain admitted for it, which
			// only exists from here on.
			handle.BindPromptRun(preparation.PromptRunID)
			notify(StartResult{Status: "started", SessionID: preparation.SessionID}, nil)
		})
		runner := todos.NewTODOExecutor(req.Dir, executor, sessionID, req.Provider)
		runner.SetMode(req.Options.RunMode)
		runner.SetResume(req.Options.Resume)
		runner.SetConcurrent(req.Options.Concurrent)
		var runErr error
		var result *todos.ExecutionResult
		// A single selection runs through Execute; a multi-select runs every todo
		// in one combined agent session via ExecuteGroup.
		if len(req.Todos) == 1 {
			result, runErr = runner.Execute(execCtx, req.Todos[0])
		} else {
			var results []*todos.ExecutionResult
			results, runErr = runner.ExecuteGroup(execCtx, req.Todos)
			if len(results) > 0 {
				result = results[0]
			}
		}
		if runErr != nil && (result == nil || !result.Cancelled) {
			logger.Warnf("todo run %s failed: %v", Label(req.Todos), runErr)
		}
		switch {
		case runErr != nil:
			notify(StartResult{}, runErr)
		case result != nil && result.Skipped:
			notify(StartResult{Status: "skipped", Message: "TODO already passes; run skipped"}, nil)
		default:
			notify(StartResult{}, errors.New("todo run completed before Captain admission"))
		}
	}()
	outcome := <-started
	return outcome.result, outcome.err
}

func NewExecutor(req Request) (todos.Executor, string, error) {
	return NewExecutorContext(context.Background(), req)
}

func NewExecutorContext(ctx context.Context, req Request) (todos.Executor, string, error) {
	kind, err := drivers.Parse(req.Options.Driver)
	if err != nil {
		return nil, "", err
	}
	// cmux returns "" as the orchestrator session id (it manages its own
	// --session-id, passed via SessionID) so TODOExecutor does not overwrite the
	// todo's recorded prior session.
	mode := req.Options.RunMode
	// Post-run checks run inside the agent loop as fixture-backed verify
	// plugins; a failing round's feedback re-runs the same session.
	var verifiers []agent.Verify
	var maxIterations int
	if mode == types.ModeRun {
		// The grader has its own chain (.gavel.yaml todos.verify > ai:); the run
		// spec decides only whether to verify and for how many rounds, so the
		// implementer's model, backend and session never mark their own work.
		// CanApprove mirrors the run's own resolve: an entrypoint that drains the
		// approval queue must not fail here while its run resolves fine.
		grader, err := todospec.Resolve(todospec.Input{
			WorkDir:    req.Dir,
			Mode:       types.ModeVerify,
			CanApprove: true,
		})
		if err != nil {
			return nil, "", err
		}
		verifiers, maxIterations, err = todos.BuildCheckVerifiers(todos.CheckVerifierOptions{
			WorkDir: req.Dir,
			Todos:   req.Todos,
			Run:     &req.Options.Spec,
			Grader:  grader.Spec,
		})
		if err != nil {
			return nil, "", err
		}
	}
	// The todo's recorded plan feeds both flows: a plan re-run reports
	// updated/unchanged, and an implement run follows the approved/edited plan.
	// Single-todo only — a group run has no single plan to attribute.
	existingPlan, err := PlanMarkdown(ctx, req.Provider, req.Todos, mode)
	if err != nil {
		return nil, "", err
	}
	// Triage compares against the rest of the backlog to spot duplicates; every
	// other prompt sees only the todos it was given.
	backlog, err := TriageBacklog(ctx, req)
	if err != nil {
		return nil, "", err
	}
	return drivers.New(kind, todos.AgentRunConfig{
		Spec:          req.Options.Spec,
		WorkDir:       req.Dir,
		Mode:          mode,
		Prompt:        req.Options.Prompt,
		Envelope:      req.Options.Envelope,
		Template:      req.Options.Template,
		ExistingPlan:  existingPlan,
		Backlog:       backlog,
		Verifiers:     verifiers,
		MaxIterations: maxIterations,
		Resume:        req.Options.Resume,
		Approvals:     req.Options.Approvals,
	})
}

// TriageBacklog loads the duplicate-detection index for a triage run. A backlog
// that cannot be listed degrades duplicate detection but not the other four
// verdicts, so it is logged rather than fatal.
func TriageBacklog(ctx context.Context, req Request) (string, error) {
	if req.Options.Envelope != todoprompt.EnvelopeTriage || req.Provider == nil {
		return "", nil
	}
	candidates, err := req.Provider.List(ctx, todos.DiscoveryFilters{
		ExcludeStatuses: []types.Status{types.StatusCompleted},
	})
	if err != nil {
		logger.Warnf("triage duplicate detection is degraded: could not list the backlog: %v", err)
		return "", nil
	}
	return todos.BuildBacklogIndex(candidates, req.Todos), nil
}

func PlanMarkdown(ctx context.Context, provider todos.Provider, todoList []*types.TODO, mode types.RunMode) (string, error) {
	if len(todoList) != 1 || (mode != types.ModePlan && mode != types.ModeRun) {
		return "", nil
	}
	content, ok := provider.(todos.PlanContentProvider)
	if !ok {
		return "", fmt.Errorf("PostgreSQL TODO runtime does not support durable plan content")
	}
	return content.PlanMarkdown(ctx, todoList[0], mode)
}
