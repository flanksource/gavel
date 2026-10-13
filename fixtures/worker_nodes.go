package fixtures

import (
	"context"
	"fmt"
	"github.com/flanksource/captain/pkg/api"
)

func runNodesIsolated(ctx context.Context, nodes []*FixtureNode, opts RunOptions) ([]FixtureResult, *ExecutionSnapshot, error) {
	if opts.Evaluator != nil || opts.Progress != nil || opts.Recorder != nil {
		return nil, nil, fmt.Errorf("fixture isolation cannot serialize an evaluator, progress callback, or active recorder")
	}
	if opts.Runtime != nil {
		runtime := *opts.Runtime
		opts.ResolveSpec = func(context.Context) (api.ResolveSpecOptions, error) { return runtime, nil }
	}
	carrier := &Runner{options: RunnerOptions{Context: ctx, WorkDir: opts.WorkDir, ExecutablePath: opts.ExecutablePath, ProcessRuntime: opts.ProcessRuntime, Spec: opts.Spec, ResolveSpec: opts.ResolveSpec, ProgressSink: ProgressSinkFromContext(ctx)}}
	result, err := carrier.runWorkerRequest(workerRequest{Options: carrier.options, Nodes: nodes, NodeOptions: &opts})
	if result == nil {
		return nil, carrier.workerSnapshot, err
	}
	var tests []FixtureTest
	for _, node := range nodes {
		if node == nil {
			continue
		}
		node.Walk(func(n *FixtureNode) {
			if n.Test != nil {
				tests = append(tests, *n.Test)
			}
		})
	}
	if len(result.Results) > len(tests) {
		return nil, carrier.workerSnapshot, fmt.Errorf("fixture worker returned %d results for %d tests", len(result.Results), len(tests))
	}
	for i := range result.Results {
		result.Results[i].Test = tests[i]
	}
	return result.Results, result.Snapshot, err
}
