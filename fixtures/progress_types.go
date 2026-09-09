package fixtures

import (
	"context"
	"time"
)

const ExecutionSnapshotVersion = 1

type ExecutionState string

const (
	ExecutionQueued    ExecutionState = "queued"
	ExecutionRunning   ExecutionState = "running"
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
	ExecutionKindRoot      ExecutionKind = "root"
	ExecutionKindFile      ExecutionKind = "file"
	ExecutionKindSection   ExecutionKind = "section"
	ExecutionKindTable     ExecutionKind = "table"
	ExecutionKindCommand   ExecutionKind = "command"
	ExecutionKindTest      ExecutionKind = "test"
	ExecutionKindLint      ExecutionKind = "lint"
	ExecutionKindAI        ExecutionKind = "ai"
	ExecutionKindChecklist ExecutionKind = "checklist"
	ExecutionKindSetup     ExecutionKind = "setup"
	ExecutionKindBuild     ExecutionKind = "build"
	ExecutionKindDaemon    ExecutionKind = "daemon"
)

type ExecutionStep struct {
	Key    string
	Name   string
	Kind   ExecutionKind
	Origin *FixtureOrigin
}

type ExecutionNode struct {
	Key        string           `json:"key"`
	Name       string           `json:"name"`
	Kind       ExecutionKind    `json:"kind"`
	State      ExecutionState   `json:"state"`
	Origin     *FixtureOrigin   `json:"origin,omitempty"`
	StartedAt  *time.Time       `json:"started_at,omitempty"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	Duration   time.Duration    `json:"duration,omitempty"`
	Done       int              `json:"done,omitempty"`
	Total      int              `json:"total,omitempty"`
	Error      string           `json:"error,omitempty"`
	Children   []*ExecutionNode `json:"children,omitempty"`
}

type ExecutionSummary struct {
	Total     int `json:"total"`
	Queued    int `json:"queued,omitempty"`
	Running   int `json:"running,omitempty"`
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
