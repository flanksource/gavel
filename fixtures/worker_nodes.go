package fixtures

import (
	"context"
	"fmt"
	"github.com/flanksource/captain/pkg/api"
)

func runNodesIsolated(ctx context.Context, nodes []*FixtureNode, opts RunOptions) ([]FixtureResult, *ExecutionSnapshot, error) {
	if opts.Evaluator != nil || opts.Progress != nil || opts.Traces != nil || opts.Scope != nil {
		return nil, nil, fmt.Errorf("fixture isolation cannot serialize an evaluator, progress callback, armed traces, or fixture scope")
	}
	// The worker prepares the document's setup and resolves step connections
	// itself (RunNodes), so a host resolver it never receives would leave both
	// resolving against gavel's own store instead.
	if opts.Connections != nil {
		return nil, nil, fmt.Errorf("fixture isolation cannot serialize a host connection resolver (RunOptions.Connections); run without process isolation")
	}
	if opts.Runtime != nil {
		runtime := *opts.Runtime
		opts.ResolveSpec = func(context.Context) (api.ResolveSpecOptions, error) { return runtime, nil }
	}
	carrier := &Runner{options: RunnerOptions{Context: ctx, WorkDir: opts.WorkDir, ExecutablePath: opts.ExecutablePath, ProcessRuntime: opts.ProcessRuntime, Spec: opts.Spec, ResolveSpec: opts.ResolveSpec, ProgressSink: ProgressSinkFromContext(ctx), Prompter: opts.Prompter}}
	result, err := carrier.runWorkerRequest(workerRequest{Options: carrier.options, Nodes: nodes, NodeOptions: &opts, NodeParams: opts.Params})
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
	next := 0
	for i := range result.Results {
		if isTeardownResult(result.Results[i]) {
			continue
		}
		if next >= len(tests) {
			return nil, carrier.workerSnapshot, fmt.Errorf("fixture worker returned more results than the %d tests sent", len(tests))
		}
		result.Results[i].Test = tests[next]
		next++
	}
	return result.Results, result.Snapshot, err
}
