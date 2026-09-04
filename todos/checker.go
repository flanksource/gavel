package todos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/flanksource/gavel/todos/types"
)

// DefaultCheckConcurrency bounds how many definition-of-done checks run at once
// when nothing configures it. Each check runs the TODO's fixture — a real test
// suite — so an unbounded fan-out over a large selection thrashes the machine
// rather than finishing sooner.
const DefaultCheckConcurrency = 4

// VerifyRunner runs one todo's verify step and applies its outcome. It is the
// lifecycle host, named as an interface so this package — which the host itself
// imports — can call it without a cycle.
//
// The check is not a parallel verification path: it is the lifecycle's own
// `verify` step, dispatched by name. `gavel todos check` and the dashboard's
// verify action therefore cannot disagree with the verification an implement
// run performs, because they are the same step.
type VerifyRunner interface {
	VerifyStep(ctx context.Context, todo *types.TODO, request api.Spec) (*types.CheckResult, error)
}

// CheckOptions configures the TODO check operation.
type CheckOptions struct {
	// Runner dispatches the verify step. It is required: a check with nowhere to
	// run the definition of done is an error, never a vacuous pass.
	Runner VerifyRunner
	// Request is the caller's spec override for the verify step — `todos check`'s
	// flags, or the dashboard's payload — folded as the top layer.
	Request api.Spec
	Logger  logger.Logger
	// Concurrency caps how many checks run at once; zero uses
	// DefaultCheckConcurrency.
	Concurrency int
}

// CheckTODOs runs each issue's fixture-backed definition of done as the
// lifecycle's verify step, bounded to Concurrency at a time.
//
// The bound is an errgroup semaphore rather than a clicky task group on purpose:
// a check runs fixture steps that start tasks of their own, and nesting those
// inside a group deadlocks the drain — the outer task waits for the inner ones
// while holding the slot they need.
func CheckTODOs(ctx context.Context, todoList []*types.TODO, opts CheckOptions) ([]*types.CheckResult, error) {
	if opts.Logger == nil {
		opts.Logger = logger.StandardLogger()
	}
	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = DefaultCheckConcurrency
	}

	results := make([]*types.CheckResult, len(todoList))
	err := runBounded(ctx, len(todoList), concurrency, func(ctx context.Context, index int) {
		results[index] = CheckTODO(ctx, todoList[index], opts)
	})
	if err != nil {
		return nil, fmt.Errorf("check TODOs: %w", err)
	}
	return results, nil
}

// runBounded runs work(0..count-1) concurrently, at most concurrency at a time.
//
// It is a semaphore rather than a clicky task group because the work starts
// tasks of its own: a fixture step is itself a group, and nesting one inside an
// outer group deadlocks the drain — the outer task holds the slot the inner
// tasks are queued behind. Cancelling the context releases the waiters, which is
// the only error this can return.
func runBounded(ctx context.Context, count, concurrency int, work func(context.Context, int)) error {
	slots := semaphore.NewWeighted(int64(concurrency))
	group, groupCtx := errgroup.WithContext(ctx)
	for i := 0; i < count; i++ {
		index := i
		group.Go(func() error {
			if err := slots.Acquire(groupCtx, 1); err != nil {
				return err
			}
			defer slots.Release(1)
			work(groupCtx, index)
			return nil
		})
	}
	return group.Wait()
}

// CheckTODO runs one issue's complete definition of done — configured test/lint
// steps, its persisted Verification fixture, and its acceptance-criteria AI
// checklist — as the lifecycle's verify step. It is the shared entrypoint for
// `todos check` and the dashboard.
//
// The todo's status is written by the step's outcome, in the host, like every
// other step's: nothing here derives a verified/unverified transition of its
// own, because two places deciding what a check means is one too many.
func CheckTODO(ctx context.Context, todo *types.TODO, opts CheckOptions) *types.CheckResult {
	start := time.Now()
	if todo == nil {
		return failedCheck(start, fmt.Errorf("todo is required"))
	}
	if opts.Runner == nil {
		return failedCheck(start, fmt.Errorf("check %s: no lifecycle runner", TODOReference(todo)))
	}
	result, err := opts.Runner.VerifyStep(ctx, todo, opts.Request)
	if err != nil {
		return failedCheck(start, err)
	}
	if result == nil {
		return failedCheck(start, fmt.Errorf("check %s: verify step produced no result", TODOReference(todo)))
	}
	if result.Duration == 0 {
		result.Duration = time.Since(start)
	}
	return result
}

// failedCheck records a check that could not produce a verdict — a missing
// definition of done, a fixture that would not run, a step that could not be
// dispatched. It reports the failure rather than moving the todo: a check that
// never ran has judged nothing, and writing "unverified" from here would be a
// verdict the definition of done did not reach.
func failedCheck(start time.Time, err error) *types.CheckResult {
	return &types.CheckResult{
		AllPassed: false, Duration: time.Since(start), Error: err, ErrorText: err.Error(),
	}
}

// BuildTestResultInfo summarises a verification report as the "Latest Failure"
// record persisted onto the TODO.
func BuildTestResultInfo(report api.VerifyReport, opts types.BuildTestResultInfoOptions) *types.TestResultInfo {
	var commands []string
	var walk func(nodes []api.VerifyNode)
	walk = func(nodes []api.VerifyNode) {
		for i := range nodes {
			if nodes[i].Command != "" {
				commands = append(commands, nodes[i].Command)
			}
			walk(nodes[i].Children)
		}
	}
	walk(report.Tests)

	output := report.Feedback
	if len(output) > 2000 {
		output = output[:2000] + "\n... (output truncated)"
	}
	command := strings.Join(commands, " && ")
	if command == "" {
		command = fmt.Sprintf("gavel definition of done (checks: %d)", report.Summary.Total)
	}
	return &types.TestResultInfo{
		Command: command, CWD: opts.CWD, GitBranch: opts.GitBranch,
		GitCommit: opts.GitCommit, GitDirty: opts.GitDirty, Timestamp: opts.Timestamp,
		Passed: opts.Passed, Output: output, Duration: opts.Duration,
	}
}
