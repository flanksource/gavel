package jira

import (
	"context"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
)

type jiraTaskGroupKey struct{}

// Command rendering owns Pretty values; task rendering keeps the status line.
type jiraTaskResult[T any] struct {
	Value T
}

func runJiraTask[T any](ctx context.Context, name string, run func(context.Context, *task.Task) (T, error)) (result T, returnErr error) {
	group, nested := ctx.Value(jiraTaskGroupKey{}).(*task.Group)
	if !nested {
		surface := entity.OperationSurfaceFromContext(ctx)
		options := []task.TaskGroupOption{task.WithKind("external-issues"), task.WithLabels(map[string]string{"provider": "jira"}), task.WithConcurrency(1)}
		if surface != "" && surface != "cli" {
			options = append(options, task.WithNoRender())
		}
		group := clicky.StartGroup[jiraTaskResult[T]](name, options...)
		if surface == "cli" {
			defer clicky.WaitForGlobalCompletion()
		}
		value, err := group.Add(name, func(ctx commonscontext.Context, progress *task.Task) (jiraTaskResult[T], error) {
			result, err := run(context.WithValue(ctx, jiraTaskGroupKey{}, group.Group), progress)
			return jiraTaskResult[T]{Value: result}, err
		}, task.WithContext(ctx), task.WithCancellationDrain(), task.WithRetryConfig(task.RetryConfig{})).GetResult()
		return value.Value, err
	}
	step := group.StartStep(name)
	defer func() { step.Finish(returnErr) }()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	stop := context.AfterFunc(step.Task().Context(), func() { cancel(context.Cause(step.Task().Context())) })
	defer stop()
	if err := ctx.Err(); err != nil {
		return result, context.Cause(ctx)
	}
	return run(commonscontext.NewContext(ctx, commonscontext.WithLogger(step.Task())), step.Task())
}
