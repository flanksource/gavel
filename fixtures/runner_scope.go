package fixtures

import (
	"context"
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
func (r *Runner) addFixtureTasks(group task.TypedGroup[FixtureResult], parent context.Context) []fileScope {
	var scopes []fileScope
	for _, file := range r.tree.Children {
		var setup *PreparedSetup
		if file.Origin != nil {
			setup = r.setups[file.Origin.File]
		}
		scope := NewStepScope(scopeSeed(setup, firstFrontMatter([]*FixtureNode{file})))
		scopes = append(scopes, fileScope{file: file, scope: scope})
		sequential := hasTypedSteps(file)
		var previous *task.Task
		file.Walk(func(node *FixtureNode) {
			if node.Test == nil {
				return
			}
			env := r.envForNode(node)
			env.scope = scope
			opts := []task.Option{task.WithContext(parent), clicky.WithTaskTimeout(fixtureTimeout(*node.Test))}
			if sequential && previous != nil {
				opts = append(opts, task.WithDependencies(previous))
			}
			previous = group.Add(node.Test.String(), r.fixtureTask(node, env), opts...).Task
		})
	}
	return scopes
}

func (r *Runner) fixtureTask(node *FixtureNode, env fixtureEnv) func(flanksourceContext.Context, *task.Task) (FixtureResult, error) {
	return func(ctx flanksourceContext.Context, t *task.Task) (FixtureResult, error) {
		if err := r.progress.Start(ctx, node); err != nil {
			return FixtureResult{Name: node.Name, Status: task.StatusERR, Error: err.Error()}, err
		}
		runContext := flanksourceContext.NewContext(withFixtureProgress(ctx, func(done, total int) error {
			return r.progress.Update(ctx, node, done, total)
		}))
		result, err := r.executeFixture(runContext, *node.Test, env)
		r.attachResult(node, result)
		if progressErr := r.progress.Complete(ctx, node, result); progressErr != nil {
			return result, progressErr
		}
		return result, err
	}
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

// stepTaskTimeout is the task deadline of a typed step: its declared timeout,
// else room for its wait (declared, document default, or built-in) on top of
// the default fixture timeout.
func stepTaskTimeout(step StepSpec) time.Duration {
	if timeout := step.EffectiveTimeout(); timeout != nil {
		return timeout.Std()
	}
	wait := BuiltinWaitTimeout
	switch {
	case step.Wait != nil && step.Wait.Timeout != nil:
		wait = step.Wait.Timeout.Std()
	case step.Defaults != nil && step.Defaults.Wait != nil && step.Defaults.Wait.Timeout != nil:
		wait = step.Defaults.Wait.Timeout.Std()
	}
	return defaultFixtureTimeout + wait
}
