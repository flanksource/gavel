package aifix

import (
	"context"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
)

type repairTaskProvider struct{ captainai.StreamingProvider }

// Each repair turn holds a worker until its stream ends. Re-linting between
// turns runs its own tasks, even when Gavel has only one worker.
func (p repairTaskProvider) ExecuteStream(ctx context.Context, req captainai.Request) (<-chan captainai.Event, error) {
	started := make(chan struct{})
	events := make(chan captainai.Event)
	work := task.StartTask("AI lint repair", func(ctx commonscontext.Context, _ *task.Task) (struct{}, error) {
		input, err := p.StreamingProvider.ExecuteStream(ctx, req)
		if err != nil {
			return struct{}{}, err
		}
		close(started)
		for {
			select {
			case event, ok := <-input:
				if !ok {
					return struct{}{}, nil
				}
				select {
				case events <- event:
				case <-ctx.Done():
					return struct{}{}, ctx.Err()
				}
			case <-ctx.Done():
				return struct{}{}, ctx.Err()
			}
		}
	}, task.WithContext(ctx), task.WithCancellationDrain(), task.WithRetryConfig(task.RetryConfig{}))

	finished := make(chan error, 1)
	go func() {
		defer close(events)
		_, err := work.GetResult()
		select {
		case <-started:
		default:
			finished <- err
			return
		}
		if err != nil {
			select {
			case events <- captainai.Event{Kind: captainai.EventError, Error: err.Error()}:
			case <-ctx.Done():
			}
		}
	}()
	select {
	case <-started:
		return events, nil
	case err := <-finished:
		return nil, err
	}
}
