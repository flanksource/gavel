package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
	tracedprocess "github.com/flanksource/commons-db/tracing/process"
	"github.com/flanksource/commons-db/tracing/stacks"
	"github.com/flanksource/commons/logger"
)

const defaultDiagnosticsProbeTimeout = 2 * time.Second

type runDiagnosticsOptions struct {
	Output       io.Writer
	CaptureStack func(io.Writer) error
	CaptureTasks func() []task.TaskSnapshot
	ProbeTimeout time.Duration
}

// diagnosticsTrace is the run's process trace (traces.ProcessTrace): each
// capture dumps gavel's goroutines into its stacks and re-observes its tree,
// whose records the capture's process tree is printed from.
type diagnosticsTrace interface {
	CheckpointStacks(ctx context.Context, params stacks.CheckpointParams) error
	CheckpointTree(ctx context.Context) error
	Nodes(ctx context.Context) ([]tracedprocess.Node, error)
}

const diagnosticsCheckpointTimeout = 10 * time.Second

type runDiagnosticsReporter struct {
	options runDiagnosticsOptions
	mu      sync.Mutex
	trace   diagnosticsTrace
}

// recordOn checkpoints every later capture into trace, the run's process
// capture, tagged with the capture's reason; the capture is still logged.
func (r *runDiagnosticsReporter) recordOn(trace diagnosticsTrace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trace = trace
}

func newRunDiagnosticsReporter(options runDiagnosticsOptions) *runDiagnosticsReporter {
	if options.Output == nil {
		options.Output = logger.GetOutput()
	}
	if options.CaptureStack == nil {
		options.CaptureStack = func(w io.Writer) error {
			return pprof.Lookup("goroutine").WriteTo(w, 2)
		}
	}
	if options.CaptureTasks == nil {
		options.CaptureTasks = func() []task.TaskSnapshot { return task.SnapshotAll() }
	}
	if options.ProbeTimeout <= 0 {
		options.ProbeTimeout = defaultDiagnosticsProbeTimeout
	}
	return &runDiagnosticsReporter{options: options}
}

func (r *runDiagnosticsReporter) Capture(reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Fprintf(r.options.Output, "\n=== Gavel diagnostics: %s (%s) ===\n", reason, time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintln(r.options.Output, "-- goroutine stack --")
	if err := r.options.CaptureStack(r.options.Output); err != nil {
		fmt.Fprintf(r.options.Output, "stack capture failed: %v\n", err)
	}

	tasks, tasksReady := boundedProbe(r.options.ProbeTimeout, r.options.CaptureTasks)
	printActiveTasks(r.options.Output, tasks, tasksReady, r.options.ProbeTimeout)
	r.record(reason)
	fmt.Fprintln(r.options.Output, "=== end Gavel diagnostics ===")
}

// record checkpoints the capture into the run's process trace — gavel's
// goroutines into its stacks, a re-observed tree — and prints the tree its
// records hold. A run without one, or a checkpoint that fails, is reported in
// the log, never failing the run the watchdog is watching.
func (r *runDiagnosticsReporter) record(reason string) {
	fmt.Fprintln(r.options.Output, "-- process trace --")
	if r.trace == nil {
		fmt.Fprintln(r.options.Output, "not checkpointed: the run has no process trace (it needs --ui or --diagnostics)")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), diagnosticsCheckpointTimeout)
	defer cancel()
	err := errors.Join(
		r.trace.CheckpointStacks(ctx, stacks.CheckpointParams{PID: os.Getpid(), Transport: stacks.TransportInProcess, Reason: reason}),
		r.trace.CheckpointTree(ctx),
	)
	if err != nil {
		fmt.Fprintf(r.options.Output, "checkpoint failed: %v\n", err)
	} else {
		fmt.Fprintf(r.options.Output, "checkpointed gavel's goroutines into the run's stacks (reason %q)\n", reason)
	}
	nodes, err := r.trace.Nodes(ctx)
	printDiagnosticsProcesses(r.options.Output, nodes, err)
}

func printActiveTasks(w io.Writer, snapshots []task.TaskSnapshot, ready bool, timeout time.Duration) {
	fmt.Fprintln(w, "-- active tasks --")
	if !ready {
		fmt.Fprintf(w, "task snapshot timed out after %s\n", timeout)
		return
	}
	active := 0
	for _, snapshot := range snapshots {
		if snapshot.Status != string(task.StatusPending) && snapshot.Status != string(task.StatusRunning) {
			continue
		}
		active++
		fmt.Fprintf(w, "[%s] %s status=%s", snapshot.Type, snapshot.Name, snapshot.Status)
		if snapshot.Group != "" {
			fmt.Fprintf(w, " group=%q", snapshot.Group)
		}
		if snapshot.StartedAt != "" {
			fmt.Fprintf(w, " started=%s", snapshot.StartedAt)
		}
		fmt.Fprintln(w)
	}
	if active == 0 {
		fmt.Fprintln(w, "none")
	}
}

// printDiagnosticsProcesses prints the run's process tree as its
// process_tree records hold it, each process indented under its parent.
func printDiagnosticsProcesses(w io.Writer, nodes []tracedprocess.Node, err error) {
	fmt.Fprintln(w, "-- process tree --")
	if err != nil {
		fmt.Fprintf(w, "process tree unavailable: %v\n", err)
		return
	}
	children := map[string][]tracedprocess.Node{}
	var roots []tracedprocess.Node
	for _, node := range nodes {
		if node.IsRoot {
			roots = append(roots, node)
		} else {
			children[node.ParentKey] = append(children[node.ParentKey], node)
		}
	}
	if len(roots) == 0 {
		fmt.Fprintln(w, "unavailable: the trace has recorded no root yet")
		return
	}
	for _, root := range roots {
		printDiagnosticsProcess(w, children, root, 0)
	}
}

func boundedProbe[T any](timeout time.Duration, probe func() T) (T, bool) {
	result := make(chan T, 1)
	go func() {
		result <- probe()
	}()
	select {
	case value := <-result:
		return value, true
	case <-time.After(timeout):
		var zero T
		return zero, false
	}
}

// printDiagnosticsProcess prints node and, indented beneath it, its children.
func printDiagnosticsProcess(w io.Writer, children map[string][]tracedprocess.Node, node tracedprocess.Node, depth int) {
	fmt.Fprintf(w, "%s- pid=%d ppid=%d status=%s command=%s\n", strings.Repeat("  ", depth), node.PID, node.PPID, node.Status, node.Command)
	for _, child := range children[node.Key] {
		printDiagnosticsProcess(w, children, child, depth+1)
	}
}

type monitoredStopOptions struct {
	Timeout  time.Duration
	Interval time.Duration
	Report   func(string)
}

type monitoredContext struct {
	context.Context
	deadline time.Time
	hasLimit bool
	done     chan struct{}
	once     sync.Once
	mu       sync.RWMutex
	err      error
}

func (c *monitoredContext) Deadline() (time.Time, bool) {
	return c.deadline, c.hasLimit
}

func (c *monitoredContext) Done() <-chan struct{} {
	return c.done
}

func (c *monitoredContext) Err() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.err
}

func (c *monitoredContext) finish(err error) {
	c.once.Do(func() {
		c.mu.Lock()
		c.err = err
		c.mu.Unlock()
		close(c.done)
	})
}

func newMonitoredStopContext(parent context.Context, options monitoredStopOptions) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	ctx := &monitoredContext{Context: parent, done: make(chan struct{})}
	if options.Timeout > 0 {
		ctx.deadline = time.Now().Add(options.Timeout)
		ctx.hasLimit = true
	}
	if parentDeadline, ok := parent.Deadline(); ok && (!ctx.hasLimit || parentDeadline.Before(ctx.deadline)) {
		ctx.deadline = parentDeadline
		ctx.hasLimit = true
	}

	go monitorRunContext(ctx, options)
	return ctx, func() { ctx.finish(context.Canceled) }
}

func monitorRunContext(ctx *monitoredContext, options monitoredStopOptions) {
	var timeout <-chan time.Time
	if options.Timeout > 0 {
		timer := time.NewTimer(options.Timeout)
		defer timer.Stop()
		timeout = timer.C
	}
	var periodic <-chan time.Time
	if options.Interval > 0 && (options.Timeout <= 0 || options.Interval < options.Timeout) {
		ticker := time.NewTicker(options.Interval)
		defer ticker.Stop()
		periodic = ticker.C
	}

	for {
		select {
		case <-ctx.Context.Done():
			ctx.finish(ctx.Context.Err())
			return
		case <-ctx.Done():
			return
		case <-periodic:
			if options.Report != nil {
				options.Report("periodic")
			}
		case <-timeout:
			if options.Report != nil {
				options.Report("timeout")
			}
			ctx.finish(context.DeadlineExceeded)
			return
		}
	}
}
