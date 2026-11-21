package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	cexec "github.com/flanksource/clicky/exec"
)

const WorkerCommand = "__fixture-worker"

type workerSpec struct {
	Layers       []api.SpecLayer
	Saved        *captainconfig.AIDefaults
	RequireModel bool
}

type workerRequest struct {
	Options     RunnerOptions
	Spec        *workerSpec
	Nodes       []*FixtureNode
	NodeOptions *RunOptions
}

type workerEvent struct {
	Version       int                `json:"version"`
	Kind          string             `json:"kind"`
	Snapshot      *ExecutionSnapshot `json:"snapshot,omitempty"`
	Tree          *FixtureNode       `json:"tree,omitempty"`
	Benchmark     *BenchmarkReport   `json:"benchmark,omitempty"`
	BenchmarkPath string             `json:"benchmark_path,omitempty"`
	Results       []FixtureResult    `json:"results,omitempty"`
	Error         string             `json:"error,omitempty"`
	// ErrorCode carries the identity of a sentinel error across the process
	// boundary, which Error's text alone loses.
	ErrorCode string `json:"error_code,omitempty"`
}

const workerErrorNoFixtures = "no_fixtures"

func (e *workerEvent) setError(err error) {
	if err == nil {
		return
	}
	e.Error = err.Error()
	if errors.Is(err, ErrNoFixtures) {
		e.ErrorCode = workerErrorNoFixtures
	}
}

func (e *workerEvent) err() error {
	switch {
	case e.ErrorCode == workerErrorNoFixtures:
		return ErrNoFixtures
	case e.Error != "":
		return errors.New(e.Error)
	}
	return nil
}

func RunWorker(input io.Reader, output io.Writer) error {
	var request workerRequest
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return fmt.Errorf("fixture worker input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("fixture worker input must contain exactly one request")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	request.Options.Context = ctx
	request.Options.worker = true
	request.Options.ProcessRuntime.CPU = nil
	request.Options.ProcessRuntime.Mem = nil
	request.Options.ProcessRuntime.Timeout = nil
	zeroNice := 0
	request.Options.ProcessRuntime.Nice = &zeroNice
	if request.Spec != nil {
		request.Options.ResolveSpec = func(context.Context) (api.ResolveSpecOptions, error) {
			return api.ResolveSpecOptions{Layers: request.Spec.Layers, Saved: request.Spec.Saved, RequireModel: request.Spec.RequireModel}, nil
		}
	}
	var lock sync.Mutex
	encoder := json.NewEncoder(output)
	request.Options.ProgressSink = func(_ context.Context, snapshot ExecutionSnapshot) error {
		lock.Lock()
		defer lock.Unlock()
		return encoder.Encode(workerEvent{Version: 1, Kind: "progress", Snapshot: &snapshot})
	}
	if request.NodeOptions != nil {
		opts := *request.NodeOptions
		opts.ProcessRuntime = request.Options.ProcessRuntime
		opts.Context = ctx
		opts.ResolveSpec = request.Options.ResolveSpec
		ctx = WithProgressSink(ctx, request.Options.ProgressSink)
		results, snapshot, runErr := RunNodes(ctx, request.Nodes, opts)
		event := workerEvent{Version: 1, Kind: "result", Results: results, Snapshot: snapshot}
		event.setError(runErr)
		return encoder.Encode(event)
	}
	runner, err := NewRunner(request.Options)
	if err != nil {
		return err
	}
	tree, runErr := runner.Run()
	event := workerEvent{Version: 1, Kind: "result", Tree: tree, Benchmark: runner.BenchmarkReport()}
	if event.Benchmark != nil {
		event.BenchmarkPath = event.Benchmark.Path
	}
	if runner.progress != nil {
		snapshot := runner.ExecutionSnapshot()
		event.Snapshot = &snapshot
	}
	event.setError(runErr)
	return encoder.Encode(event)
}

func (r *Runner) runIsolated() (*FixtureNode, error) {
	result, err := r.runWorkerRequest(workerRequest{Options: r.options})
	if result == nil {
		return nil, err
	}
	r.tree, r.benchmark = result.Tree, result.Benchmark
	if r.benchmark != nil {
		r.benchmark.Path = result.BenchmarkPath
	}
	if r.tree != nil {
		restoreWorkerTree(r.tree, nil, r.options.Display)
	}
	return r.tree, err
}

func (r *Runner) runWorkerRequest(request workerRequest) (*workerEvent, error) {
	ctx := r.options.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if r.options.Spec == nil && r.options.ResolveSpec != nil {
		spec, err := r.options.ResolveSpec(ctx)
		if err != nil {
			return nil, fmt.Errorf("fixture worker spec: %w", err)
		}
		if spec.Normalize != nil {
			return nil, fmt.Errorf("fixture worker cannot serialize a spec normalization callback")
		}
		request.Spec = &workerSpec{Layers: spec.Layers, Saved: spec.Saved, RequireModel: spec.RequireModel}
	}
	input, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("fixture worker options: %w", err)
	}
	if r.options.ExecutablePath == "" {
		return nil, fmt.Errorf("fixture isolation requires a worker-capable executable")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("fixture worker protocol pipe: %w", err)
	}
	defer reader.Close()
	p := cexec.NewExec(r.options.ExecutablePath, WorkerCommand).WithCwd(r.options.WorkDir).WithContext(ctx).WithProcessGroup().WithStdin(bytes.NewReader(input)).WithExtraFiles([]*os.File{writer})
	if err := r.options.ProcessRuntime.Apply(p); err != nil {
		_ = writer.Close()
		return nil, err
	}
	limits := p.ResourceLimits()
	limits.StopGrace = 5 * time.Second
	p.WithResourceLimits(limits)
	done := make(chan error, 1)
	var result *workerEvent
	go func() { done <- r.readWorkerEvents(ctx, reader, &result, p) }()
	execution := p.Run().Result()
	closeErr := writer.Close()
	protocolErr := <-done
	if execution.Error != nil || protocolErr != nil || closeErr != nil {
		return result, fmt.Errorf("fixture worker: %w", errors.Join(execution.Error, protocolErr, closeErr))
	}
	if result == nil {
		return nil, fmt.Errorf("fixture worker exited without a result")
	}
	return result, result.err()
}

func (r *Runner) readWorkerEvents(ctx context.Context, reader io.Reader, result **workerEvent, process *cexec.Process) error {
	decoder := json.NewDecoder(reader)
	for {
		var event workerEvent
		if err := decoder.Decode(&event); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			_ = process.KillTree()
			return fmt.Errorf("malformed fixture worker response: %w", err)
		}
		if event.Version != 1 || (event.Kind != "progress" && event.Kind != "result") || *result != nil {
			_ = process.KillTree()
			return fmt.Errorf("invalid fixture worker event version %d kind %q", event.Version, event.Kind)
		}
		if event.Snapshot != nil {
			r.resultMu.Lock()
			r.workerSnapshot = event.Snapshot
			r.resultMu.Unlock()
			if r.options.ProgressSink != nil {
				if err := r.options.ProgressSink(ctx, *event.Snapshot); err != nil {
					_ = process.KillTree()
					return err
				}
			}
		}
		if event.Kind == "result" {
			*result = &event
		}
	}
}

func restoreWorkerTree(node, parent *FixtureNode, display *DisplayOptions) {
	node.Parent = parent
	if node.Results != nil {
		if node.Test != nil {
			node.Results.Test = *node.Test
		}
		node.Results.Display = display
	}
	for _, child := range node.Children {
		restoreWorkerTree(child, node, display)
	}
}
