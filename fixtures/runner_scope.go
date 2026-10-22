package fixtures

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
)

// defaultFixtureTimeout caps a markdown fixture that declares no timeout.
const defaultFixtureTimeout = 2 * time.Minute

// fileScope is the StepScope one fixture file's steps share.
type fileScope struct {
	file  *FixtureNode
	scope *StepScope
}

// addFixtureTasks queues every test node, one StepScope per file. A file with
// typed steps runs them in order, each task depending on the one before:
// exports, `steps.<id>` references and background steps only mean anything in
// sequence, so MaxWorkers parallelises across files and markdown rows only.
func (r *Runner) addFixtureTasks(group task.TypedGroup[FixtureResult], parent context.Context) ([]fileScope, []queuedFixture) {
	var scopes []fileScope
	var queued []queuedFixture
	for _, file := range r.tree.Children {
		var setup *PreparedSetup
		if file.Origin != nil {
			setup = r.setups[file.Origin.File]
		}
		scope := NewStepScope(scopeSeed(setup, firstFrontMatter([]*FixtureNode{file}), r.fileTraceEnv(file)))
		scopes = append(scopes, fileScope{file: file, scope: scope})
		sequential := hasTypedSteps(file)
		var previous *task.Task
		file.Walk(func(node *FixtureNode) {
			if node.Test == nil {
				return
			}
			env := r.envForNode(node)
			env.scope = scope
			opts := []task.Option{task.WithContext(parent), clicky.WithTaskTimeout(fixtureTimeout(*node.Test, r.backstopGrace))}
			if sequential && previous != nil {
				opts = append(opts, task.WithDependencies(previous))
			}
			queuedTask := group.Add(node.Test.String(), r.fixtureTask(node, env), opts...)
			queued = append(queued, queuedFixture{node: node, task: queuedTask, scope: scope})
			previous = queuedTask.Task
		})
	}
	return scopes, queued
}

// queuedFixture pairs a fixture node with the task that runs it and the scope
// its file's steps share.
type queuedFixture struct {
	node  *FixtureNode
	task  task.TypedTask[FixtureResult]
	scope *StepScope
}

func (r *Runner) fixtureTask(node *FixtureNode, env fixtureEnv) func(flanksourceContext.Context, *task.Task) (FixtureResult, error) {
	return func(ctx flanksourceContext.Context, t *task.Task) (FixtureResult, error) {
		if err := r.progress.Start(ctx, node); err != nil {
			return FixtureResult{}, fmt.Errorf("start fixture progress: %w", err)
		}
		runContext := flanksourceContext.NewContext(withFixtureProcess(withFixtureProgress(ctx, func(done, total int) error {
			return r.progress.Update(ctx, node, done, total)
		}), func(pid int) error {
			return r.progress.AddProcess(ctx, node, pid)
		}))
		result, err := r.executeFixture(runContext, *node.Test, env)
		return result, errors.Join(err, r.completeFixture(ctx, node, result))
	}
}

// completeFixture lands a run function's result on its node and progress,
// unless the fixture was already cut off: a run function that ignored its
// deadline returns after the verdict was written.
func (r *Runner) completeFixture(ctx context.Context, node *FixtureNode, result FixtureResult) error {
	r.resultMu.Lock()
	defer r.resultMu.Unlock()
	if r.cutOff[node] {
		return nil
	}
	r.attachResult(node, result)
	return r.progress.Complete(ctx, node, result)
}

// collectFixtureResults reads every fixture task on its own, in tree order, so
// one task's error becomes that fixture's verdict instead of discarding the
// rest. It returns the result nodes and any error recording a verdict.
func (r *Runner) collectFixtureResults(ctx context.Context, queued []queuedFixture) ([]FixtureNode, error) {
	var tests []FixtureNode
	var errs []error
	for _, q := range queued {
		returned, taskErr := q.task.GetResult()
		landed, err := r.settleFixture(ctx, q, returned, taskErr)
		errs = append(errs, err)
		tests = append(tests, FixtureNode{Name: landed.Name, Type: TestNode, Results: landed})
	}
	return tests, errors.Join(errs...)
}

// settleFixture gives every scheduled fixture exactly one result. A run
// function's own result stands when its task neither erred nor was cancelled;
// otherwise the task's outcome becomes the final verdict, built on whatever
// result already landed, and a failed typed step fails its file's scope so the
// steps after it skip.
func (r *Runner) settleFixture(ctx context.Context, q queuedFixture, returned FixtureResult, taskErr error) (*FixtureResult, error) {
	r.resultMu.Lock()
	defer r.resultMu.Unlock()
	node := q.node
	if taskErr == nil && node.Results != nil {
		return node.Results, nil
	}
	base := returned
	if node.Results != nil {
		base = *node.Results
	}
	base.Name, base.Test = node.Name, *node.Test
	result := cutOffVerdict(base, taskErr, q.failedStep())
	r.cutOff[node] = true
	r.attachResult(node, result)
	if node.Test.Step != nil && isFailureStatus(result.Status) && !node.Test.Step.ContinueOnError {
		q.scope.MarkFailed(stepID(*node.Test))
	}
	return node.Results, r.progress.Complete(ctx, node, result)
}

// cutOffVerdict is the result of a fixture task that ended without its own
// result standing: cut off by its backstop deadline it fails as timed out;
// cancelled behind a failed typed step it skips like step_exec's skip; any
// other task error errors it; cancelled for any other reason it is cancelled.
func cutOffVerdict(base FixtureResult, taskErr error, failedStep string) FixtureResult {
	switch {
	case errors.Is(taskErr, context.DeadlineExceeded):
		result := base.Failf("%s exceeded its backstop deadline: %v", base.Name, taskErr)
		result.TimedOut = true
		return result
	case failedStep != "":
		base.Status, base.Error = task.StatusSKIP, fmt.Sprintf("skipped: step %s failed", failedStep)
		return base
	case taskErr != nil:
		return base.Errorf(taskErr, "fixture task failed")
	default:
		base.Status, base.Error = task.StatusCancelled, "cancelled before the fixture produced a result"
		return base
	}
}

// failedStep is the typed step that failed before this one in its file, when
// the task was cancelled because of it; empty otherwise.
func (q queuedFixture) failedStep() string {
	if q.node.Test.Step == nil || q.task.Status() != task.StatusCancelled {
		return ""
	}
	return q.scope.FailedStep()
}

func stepID(test FixtureTest) string {
	if test.Step.ID != "" {
		return test.Step.ID
	}
	return test.Name
}

// attachTeardownResults tears every file's scope down once its steps are done,
// last file first, adding background verdicts and teardown failures to the
// file they belong to.
func (r *Runner) attachTeardownResults(ctx context.Context, scopes []fileScope) {
	for i := len(scopes) - 1; i >= 0; i-- {
		for _, result := range teardownResults(ctx, scopes[i].scope) {
			r.resultMu.Lock()
			scopes[i].file.AddChild(&FixtureNode{Name: result.Name, Type: TestNode, Results: &result})
			r.resultMu.Unlock()
		}
	}
}

func hasTypedSteps(file *FixtureNode) bool {
	typed := false
	file.Walk(func(node *FixtureNode) {
		typed = typed || (node.Test != nil && node.Test.Step != nil)
	})
	return typed
}

// declaredTimeoutGrace is how long a fixture's task deadline outlasts its
// declared timeout, so the declared one fires first and reads as a timeout.
const declaredTimeoutGrace = 10 * time.Second

// stepTaskTimeout is the task deadline of a typed step that declares no
// timeout: room for its wait (declared, document default, or built-in) on top
// of the default fixture timeout.
func stepTaskTimeout(step StepSpec) time.Duration {
	wait := BuiltinWaitTimeout
	switch {
	case step.Wait != nil && step.Wait.Timeout != nil:
		wait = step.Wait.Timeout.Std()
	case step.Defaults != nil && step.Defaults.Wait != nil && step.Defaults.Wait.Timeout != nil:
		wait = step.Defaults.Wait.Timeout.Std()
	}
	return defaultFixtureTimeout + wait
}
