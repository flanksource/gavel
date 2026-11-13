package fixtures

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/testrunner/parsers"
)

const ExecutionSnapshotVersion = 1

type ExecutionState string

const (
	ExecutionQueued    ExecutionState = "queued"
	ExecutionRunning   ExecutionState = "running"
	ExecutionWaiting   ExecutionState = "waiting" // running, blocked on a person answering a form
	ExecutionPassed    ExecutionState = "passed"
	ExecutionFailed    ExecutionState = "failed"
	ExecutionErrored   ExecutionState = "errored"
	ExecutionWarned    ExecutionState = "warned"
	ExecutionSkipped   ExecutionState = "skipped"
	ExecutionCancelled ExecutionState = "cancelled"
	ExecutionTimedOut  ExecutionState = "timed_out"
)

type ExecutionKind string

const (
	ExecutionKindRoot       ExecutionKind = "root"
	ExecutionKindFile       ExecutionKind = "file"
	ExecutionKindSection    ExecutionKind = "section"
	ExecutionKindTable      ExecutionKind = "table"
	ExecutionKindCommand    ExecutionKind = "command"
	ExecutionKindTest       ExecutionKind = "test"
	ExecutionKindLint       ExecutionKind = "lint"
	ExecutionKindAI         ExecutionKind = "ai"
	ExecutionKindChecklist  ExecutionKind = "checklist"
	ExecutionKindSetup      ExecutionKind = "setup"
	ExecutionKindBuild      ExecutionKind = "build"
	ExecutionKindDaemon     ExecutionKind = "daemon"
	ExecutionKindDaemonStop ExecutionKind = "daemon-stop"
	ExecutionKindCleanup    ExecutionKind = "cleanup"
)

type ExecutionStep struct {
	Key    string
	Name   string
	Kind   ExecutionKind
	Origin *FixtureOrigin
}

type ExecutionNode struct {
	Key        string         `json:"key"`
	Name       string         `json:"name"`
	Kind       ExecutionKind  `json:"kind"`
	State      ExecutionState `json:"state"`
	Origin     *FixtureOrigin `json:"origin,omitempty"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	Duration   time.Duration  `json:"duration,omitempty"`
	Done       int            `json:"done,omitempty"`
	Total      int            `json:"total,omitempty"`
	Error      string         `json:"error,omitempty"`
	// PIDs are the processes the fixture started: its services as they launch,
	// its commands once they exit.
	PIDs []int `json:"pids,omitempty"`
	// Traces are the windows of the traces the fixture armed, once it finished.
	Traces []parsers.TraceResult `json:"traces,omitempty"`
	// Result is the fixture's result once it finished, so a live view shows a
	// step's detail and children (a query step's findings) without waiting for
	// the run to end.
	Result *FixtureResult `json:"result,omitempty"`
	// Source is the step's YAML/Markdown as authored, read once when the tree is
	// built, so a queued, running or waiting step shows what it will execute.
	Source *parsers.SourceSnippet `json:"source,omitempty"`
	// Output is the tail (the last MaxExecutionOutput bytes) of what the running
	// fixture wrote to its StepContext.Stdout/Stderr; OutputTruncated says earlier bytes were
	// dropped. It is cleared when the fixture finishes: Result has the rest.
	Output          string `json:"output,omitempty"`
	OutputTruncated bool   `json:"output_truncated,omitempty"`
	// Metadata is the running fixture's structured live detail, by the same
	// keys its result's Metadata will carry (a query step's "query" record),
	// set through StepContext.SetDetail. It is cleared when the fixture finishes:
	// Result has the final one.
	Metadata map[string]any   `json:"metadata,omitempty"`
	Children []*ExecutionNode `json:"children,omitempty"`
}

type ExecutionSummary struct {
	Total     int `json:"total"`
	Queued    int `json:"queued,omitempty"`
	Running   int `json:"running,omitempty"`
	Waiting   int `json:"waiting,omitempty"`
	Passed    int `json:"passed,omitempty"`
	Failed    int `json:"failed,omitempty"`
	Errored   int `json:"errored,omitempty"`
	Warned    int `json:"warned,omitempty"`
	Skipped   int `json:"skipped,omitempty"`
	Cancelled int `json:"cancelled,omitempty"`
	TimedOut  int `json:"timed_out,omitempty"`
}

type ExecutionSnapshot struct {
	Version   int              `json:"version"`
	Iteration int64            `json:"iteration"`
	State     ExecutionState   `json:"state"`
	StartedAt *time.Time       `json:"started_at,omitempty"`
	EndedAt   *time.Time       `json:"ended_at,omitempty"`
	Summary   ExecutionSummary `json:"summary"`
	Root      *ExecutionNode   `json:"root"`
}

type ProgressSink func(context.Context, ExecutionSnapshot) error

type fixtureProgressKey struct{}

func withFixtureProgress(ctx context.Context, progress func(done, total int) error) context.Context {
	return context.WithValue(ctx, fixtureProgressKey{}, progress)
}

func fixtureProgressFromContext(ctx context.Context) func(done, total int) error {
	progress, _ := ctx.Value(fixtureProgressKey{}).(func(done, total int) error)
	return progress
}

type fixtureWaitingKey struct{}

func withFixtureWaiting(ctx context.Context, report func(waiting bool) error) context.Context {
	return context.WithValue(ctx, fixtureWaitingKey{}, report)
}

// ReportWaiting marks the progress node of the fixture ctx runs as waiting on
// a person (true) or running again (false). A fixture run with no progress
// tracker has no node to mark.
func ReportWaiting(ctx context.Context, waiting bool) error {
	report, _ := ctx.Value(fixtureWaitingKey{}).(func(waiting bool) error)
	if report == nil {
		return nil
	}
	return report(waiting)
}

type fixtureOutputKey struct{}

func withFixtureOutput(ctx context.Context, output io.Writer) context.Context {
	return context.WithValue(ctx, fixtureOutputKey{}, output)
}

// fixtureOutput is the live output stream of the fixture ctx runs (an
// *outputStream onto its progress node, published at most once per
// outputPublishInterval). A fixture run with no progress tracker, or one
// nobody watches, has nowhere to stream to, so its writes are discarded. Kinds
// write through StepContext.Stdout/Stderr/SetDetail, which also keep it.
func fixtureOutput(ctx context.Context) io.Writer {
	if output, ok := ctx.Value(fixtureOutputKey{}).(io.Writer); ok {
		return output
	}
	return io.Discard
}

type fixtureTaskKey struct{}

func withFixtureTask(ctx context.Context, t *task.Task) context.Context {
	return context.WithValue(ctx, fixtureTaskKey{}, t)
}

// fixtureTask is the clicky task the fixture ctx runs as, when it runs as one
// of its own (the Runner's per-fixture tasks); nil otherwise — RunNodes runs
// nodes inline, and a task found further up ctx belongs to its caller.
func fixtureTask(ctx context.Context) *task.Task {
	t, _ := ctx.Value(fixtureTaskKey{}).(*task.Task)
	return t
}

type fixtureProcessKey struct{}

func withFixtureProcess(ctx context.Context, report func(pid int) error) context.Context {
	return context.WithValue(ctx, fixtureProcessKey{}, report)
}

// ReportProcess records pid on the progress node of the fixture ctx runs, so
// a process trace can link the process to it. A fixture run with no progress
// tracker has no node to record it on.
func ReportProcess(ctx context.Context, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("report process: invalid pid %d", pid)
	}
	report, _ := ctx.Value(fixtureProcessKey{}).(func(pid int) error)
	if report == nil {
		return nil
	}
	return report(pid)
}
