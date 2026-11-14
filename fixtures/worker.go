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
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/gavel/slots"
)

const WorkerCommand = "__fixture-worker"

type workerSpec struct {
	Layers       []api.SpecLayer
	Saved        *captainconfig.AIDefaults
	RequireModel bool
}

type workerRequest struct {
	SlotLimits  slots.Limits
	Options     RunnerOptions
	Spec        *workerSpec
	Nodes       []*FixtureNode
	NodeOptions *RunOptions
	// Prompts reports that the parent has a Prompter: the worker's form steps
	// relay their prompts to it. NodeParams are NodeOptions.Params, which
	// RunOptions does not serialize.
	Prompts    bool
	NodeParams map[string]any
}

// WorkerIO are the descriptors of the worker protocol: the request on Input,
// events on Output, and the parent's prompt answers on Answers.
type WorkerIO struct {
	Input   io.Reader
	Output  io.Writer
	Answers io.Reader
}

// WorkerProcessIO is the worker protocol of a process started as
// WorkerCommand: the request on stdin, events on fd 3, answers on fd 4.
func WorkerProcessIO() (WorkerIO, error) {
	output := os.NewFile(3, "fixture-worker-protocol")
	answers := os.NewFile(4, "fixture-worker-answers")
	if output == nil || answers == nil {
		return WorkerIO{}, fmt.Errorf("fixture worker protocol descriptors 3 and 4 are missing")
	}
	return WorkerIO{Input: os.Stdin, Output: output, Answers: answers}, nil
}

type workerEvent struct {
	Version       int                `json:"version"`
	Kind          string             `json:"kind"`
	Snapshot      *ExecutionSnapshot `json:"snapshot,omitempty"`
	Tree          *FixtureNode       `json:"tree,omitempty"`
	Benchmark     *BenchmarkReport   `json:"benchmark,omitempty"`
	BenchmarkPath string             `json:"benchmark_path,omitempty"`
	Results       []FixtureResult    `json:"results,omitempty"`
	PromptID      string             `json:"prompt_id,omitempty"`
	Prompt        *prompt.Prompt     `json:"prompt,omitempty"`
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

func readWorkerRequest(input io.Reader) (workerRequest, error) {
	var request workerRequest
	decoder := json.NewDecoder(input)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, fmt.Errorf("fixture worker input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return request, fmt.Errorf("fixture worker input must contain exactly one request")
	}
	return request, nil
}

func RunWorker(wio WorkerIO) error {
	request, err := readWorkerRequest(wio.Input)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	request.Options.Context = ctx
	request.Options.worker = true
	request.Options.ProcessRuntime.Slots = request.SlotLimits
	request.Options.ProcessRuntime.CPU = nil
	request.Options.ProcessRuntime.Mem = nil
	if !request.SlotLimits.Enabled() {
		request.Options.ProcessRuntime.Timeout = nil
	}
	zeroNice := 0
	request.Options.ProcessRuntime.Nice = &zeroNice
	if request.Spec != nil {
		request.Options.ResolveSpec = func(context.Context) (api.ResolveSpecOptions, error) {
			return api.ResolveSpecOptions{Layers: request.Spec.Layers, Saved: request.Spec.Saved, RequireModel: request.Spec.RequireModel}, nil
		}
	}
	var lock sync.Mutex
	encoder := json.NewEncoder(wio.Output)
	send := func(event workerEvent) error {
		lock.Lock()
		defer lock.Unlock()
		return encoder.Encode(event)
	}
	request.Options.ProgressSink = func(_ context.Context, snapshot ExecutionSnapshot) error {
		return send(workerEvent{Version: 1, Kind: "progress", Snapshot: &snapshot})
	}
	if request.Prompts {
		if request.Options.Prompter, err = startWorkerPrompter(send, wio.Answers); err != nil {
			return err
		}
	}
	if request.NodeOptions != nil {
		return send(runWorkerNodes(ctx, request))
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
	return send(event)
}

// runWorkerNodes runs a RunNodes request with the worker's runtime, spec and
// prompt relay over the node options the parent sent.
func runWorkerNodes(ctx context.Context, request workerRequest) workerEvent {
	opts := *request.NodeOptions
	opts.ProcessRuntime = request.Options.ProcessRuntime
	opts.Context = ctx
	opts.ResolveSpec = request.Options.ResolveSpec
	opts.Prompter = request.Options.Prompter
	opts.Params = request.NodeParams
	results, snapshot, runErr := RunNodes(WithProgressSink(ctx, request.Options.ProgressSink), request.Nodes, opts)
	event := workerEvent{Version: 1, Kind: "result", Results: results, Snapshot: snapshot}
	event.setError(runErr)
	return event
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
	input, err := r.workerInput(ctx, request)
	if err != nil {
		return nil, err
	}
	if r.options.ExecutablePath == "" {
		return nil, fmt.Errorf("fixture isolation requires a worker-capable executable")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("fixture worker protocol pipe: %w", err)
	}
	defer reader.Close()
	answerReader, answerWriter, err := os.Pipe()
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("fixture worker answer pipe: %w", err)
	}
	defer answerWriter.Close()
	p := cexec.NewExec(r.options.ExecutablePath, WorkerCommand).WithCwd(r.options.WorkDir).WithContext(ctx).WithProcessGroup().WithStdin(bytes.NewReader(input)).WithExtraFiles([]*os.File{writer, answerReader})
	if err := r.options.ProcessRuntime.Apply(p); err != nil {
		_ = writer.Close()
		_ = answerReader.Close()
		return nil, err
	}
	if request.SlotLimits.Enabled() {
		p.WithTimeout(0)
	}
	limits := p.ResourceLimits()
	limits.StopGrace = 5 * time.Second
	p.WithResourceLimits(limits)
	done := make(chan error, 1)
	var result *workerEvent
	relay := newPromptRelay(ctx, r.options.Prompter, answerWriter)
	go func() { done <- r.readWorkerEvents(ctx, reader, &result, p, relay) }()
	execution := p.Run().Result()
	closeErr := errors.Join(writer.Close(), answerReader.Close())
	protocolErr := <-done
	relayErr := relay.close()
	if execution.Error != nil || protocolErr != nil || closeErr != nil || relayErr != nil {
		return result, fmt.Errorf("fixture worker: %w", errors.Join(execution.Error, protocolErr, closeErr, relayErr))
	}
	if result == nil {
		return nil, fmt.Errorf("fixture worker exited without a result")
	}
	return result, result.err()
}

// workerInput encodes the request with what the worker cannot resolve itself:
// the resolved spec, the slot limits and whether a prompter answers it.
func (r *Runner) workerInput(ctx context.Context, request workerRequest) ([]byte, error) {
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
	request.SlotLimits = r.options.ProcessRuntime.Slots
	request.Prompts = r.options.Prompter != nil
	input, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("fixture worker options: %w", err)
	}
	return input, nil
}

func (r *Runner) readWorkerEvents(ctx context.Context, reader io.Reader, result **workerEvent, process *cexec.Process, relay *promptRelay) error {
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
		if event.Version != 1 || !slices.Contains(workerEventKinds, event.Kind) || *result != nil {
			_ = process.KillTree()
			return fmt.Errorf("invalid fixture worker event version %d kind %q", event.Version, event.Kind)
		}
		if err := relay.handle(event); err != nil {
			_ = process.KillTree()
			return err
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
