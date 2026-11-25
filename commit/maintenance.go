package commit

import (
	"context"

	"github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
)

func runMaintenanceTask(ctx context.Context, name string, run func(context.Context) error) error {
	work := task.StartTask(name, func(ctx commonscontext.Context, _ *task.Task) (struct{}, error) {
		return struct{}{}, run(ctx)
	}, task.WithContext(ctx), task.WithCancellationDrain(), task.WithRetryConfig(task.RetryConfig{}))
	_, err := work.GetResult()
	return err
}
