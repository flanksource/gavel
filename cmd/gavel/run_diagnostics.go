package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/process"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
)

const defaultDiagnosticsProbeTimeout = 2 * time.Second

type runDiagnosticsOptions struct {
	Output           io.Writer
	CaptureStack     func(io.Writer) error
	CaptureTasks     func() []task.TaskSnapshot
	CaptureProcesses func() ([]process.Process, error)
	ProbeTimeout     time.Duration
}

type runDiagnosticsReporter struct {
	options runDiagnosticsOptions
	mu      sync.Mutex
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
	if options.CaptureProcesses == nil {
		options.CaptureProcesses = func() ([]process.Process, error) {
			snapshot, err := process.Discover(context.Background(), process.SnapshotOptions{})
			if err != nil {
				return nil, err
			}
			return snapshot.Subtree(os.Getpid()), nil
		}
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

	processes, processesReady := boundedProbe(r.options.ProbeTimeout, r.options.captureProcessesResult)
	printDiagnosticsProcesses(r.options.Output, processes, processesReady, r.options.ProbeTimeout)
	fmt.Fprintln(r.options.Output, "=== end Gavel diagnostics ===")
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

type processProbeResult struct {
	processes []process.Process
	err       error
}

func (o runDiagnosticsOptions) captureProcessesResult() processProbeResult {
	processes, err := o.CaptureProcesses()
	return processProbeResult{processes: processes, err: err}
}

func printDiagnosticsProcesses(w io.Writer, result processProbeResult, ready bool, timeout time.Duration) {
	fmt.Fprintln(w, "-- process tree --")
	if !ready {
		fmt.Fprintf(w, "process snapshot timed out after %s\n", timeout)
		return
	}
	if result.err != nil {
		fmt.Fprintf(w, "process snapshot failed: %v\n", result.err)
		return
	}
	if len(result.processes) == 0 {
		fmt.Fprintln(w, "unavailable")
		return
	}
	children := map[int][]process.Process{}
	for _, p := range result.processes[1:] {
		children[p.PPID] = append(children[p.PPID], p)
	}
	printDiagnosticsProcess(w, children, result.processes[0], 0)
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

// printDiagnosticsProcess prints p and, indented beneath it, its children.
func printDiagnosticsProcess(w io.Writer, children map[int][]process.Process, p process.Process, depth int) {
	fmt.Fprintf(w, "%s- pid=%d ppid=%d status=%s command=%s\n", strings.Repeat("  ", depth), p.PID, p.PPID, p.Status, p.Command)
	for _, child := range children[p.PID] {
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
