package fixtures

import "context"

type progressSinkContextKey struct{}

func WithProgressSink(ctx context.Context, sink ProgressSink) context.Context {
	return context.WithValue(ctx, progressSinkContextKey{}, sink)
}

func ProgressSinkFromContext(ctx context.Context) ProgressSink {
	if ctx == nil {
		return nil
	}
	sink, _ := ctx.Value(progressSinkContextKey{}).(ProgressSink)
	return sink
}

type ExecutionReporter struct {
	tracker *executionTracker
}

func NewExecutionReporter(nodes []*FixtureNode, workDir string, steps []ExecutionStep, sink ProgressSink) *ExecutionReporter {
	root := &FixtureNode{Name: "Fixtures", Type: SectionNode, Children: nodes}
	return &ExecutionReporter{tracker: newExecutionTracker(root, workDir, steps, sink)}
}

// Snapshot returns the reporter's current execution tree, so a caller that
// registered no sink can still read the state the run ended in.
func (r *ExecutionReporter) Snapshot() ExecutionSnapshot {
	if r == nil {
		return ExecutionSnapshot{}
	}
	return r.tracker.Snapshot()
}

func (r *ExecutionReporter) Publish(ctx context.Context) error {
	if r == nil {
		return nil
	}
	return r.tracker.Publish(ctx)
}

func (r *ExecutionReporter) StartFixture(ctx context.Context, node *FixtureNode) error {
	if r == nil {
		return nil
	}
	return r.tracker.Start(ctx, node)
}

func (r *ExecutionReporter) UpdateFixture(ctx context.Context, node *FixtureNode, done, total int) error {
	if r == nil {
		return nil
	}
	return r.tracker.Update(ctx, node, done, total)
}

func (r *ExecutionReporter) CompleteFixture(ctx context.Context, node *FixtureNode, result FixtureResult) error {
	if r == nil {
		return nil
	}
	return r.tracker.Complete(ctx, node, result)
}

func (r *ExecutionReporter) StartStep(ctx context.Context, key string) error {
	if r == nil {
		return nil
	}
	return r.tracker.StartStep(ctx, key)
}

func (r *ExecutionReporter) UpdateStep(ctx context.Context, key string, done, total int) error {
	if r == nil {
		return nil
	}
	return r.tracker.UpdateStep(ctx, key, done, total)
}

func (r *ExecutionReporter) CompleteStep(ctx context.Context, key string, result FixtureResult) error {
	if r == nil {
		return nil
	}
	return r.tracker.CompleteStep(ctx, key, executionState(result.Status), fixtureResultError(result))
}

func fixtureResultError(result FixtureResult) error {
	if result.Error == "" {
		return nil
	}
	return executionError(result.Error)
}

type executionError string

func (e executionError) Error() string { return string(e) }
