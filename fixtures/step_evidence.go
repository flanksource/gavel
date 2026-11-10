package fixtures

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"sync"
	"time"

	"github.com/flanksource/clicky/task"
)

// stepEvidence is what one attempt of a step recorded as it ran: its stdout
// and stderr, the structured details a UI renders (a query step's "query"
// record), and the command it ran. Each record is kept here first, then
// published to the live view (the fixture's progress node) and, when the step
// runs as a clicky task of its own, to that task's snapshot — so what a running
// step showed is what its result keeps, even when Execute failed.
type stepEvidence struct {
	started  time.Time
	live     io.Writer
	task     *task.Task
	progress func(done, total int) error

	mu             sync.Mutex
	stdout, stderr bytes.Buffer
	details        map[string]any
	command, cwd   string
}

// newStepEvidence starts an attempt's record. ctx carries the live output
// stream and the step's own task, when the runner gave it either.
func newStepEvidence(ctx context.Context, progress func(done, total int) error) *stepEvidence {
	e := &stepEvidence{started: time.Now(), live: fixtureOutput(ctx), task: fixtureTask(ctx), progress: progress, details: map[string]any{}}
	if e.task != nil {
		e.task.SetOutputProvider(e.output)
		e.task.SetDetailsProvider(e.snapshot)
	}
	return e
}

// recorder is the attempt's record. A StepContext the runner did not start —
// a kind's unit test — records into one of its own with no live view.
func (s *StepContext) recorder() *stepEvidence {
	if s.evidence == nil {
		s.evidence = newStepEvidence(context.Background(), s.Opts.Progress)
	}
	return s.evidence
}

// Stdout is where the step writes its standard output (a process's stdout, a
// SQL program's PRINT messages): kept as the result's Stdout and streamed to
// the live view as it is written.
func (s *StepContext) Stdout() io.Writer { return evidenceStream{s.recorder(), false} }

// Stderr is Stdout for the step's error output.
func (s *StepContext) Stderr() io.Writer { return evidenceStream{s.recorder(), true} }

// SetDetail records key in the result's Metadata (copied to plain values, so
// the caller may go on mutating value) and publishes it to the live view.
func (s *StepContext) SetDetail(key string, value any) error {
	plain, err := PlainValue(value)
	if err != nil {
		return fmt.Errorf("step detail %q: %w", key, err)
	}
	e := s.recorder()
	e.mu.Lock()
	e.details[key] = plain
	e.mu.Unlock()
	if stream, ok := e.live.(*outputStream); ok {
		stream.setMetadata(key, plain)
	}
	return nil
}

// SetCommand records the command the step runs and where, as the result's
// Command and CWD.
func (s *StepContext) SetCommand(command, cwd string) {
	e := s.recorder()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.command, e.cwd = command, cwd
}

// Progress reports done of total units of the step's work (tests, linters) to
// the live view and the step's task.
func (s *StepContext) Progress(done, total int) error {
	e := s.recorder()
	if e.task != nil {
		e.task.SetProgress(done, total)
	}
	if e.progress == nil {
		return nil
	}
	return e.progress(done, total)
}

// fill folds the record into result wherever the kind left the field empty:
// what a kind set itself wins.
func (e *stepEvidence) fill(result FixtureResult) FixtureResult {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, field := range []struct {
		target *string
		value  string
	}{{&result.Stdout, e.stdout.String()}, {&result.Stderr, e.stderr.String()}, {&result.Command, e.command}, {&result.CWD, e.cwd}} {
		if *field.target == "" {
			*field.target = field.value
		}
	}
	if len(e.details) > 0 && result.Metadata == nil {
		result.Metadata = map[string]any{}
	}
	for key, value := range e.details {
		if _, ok := result.Metadata[key]; !ok {
			result.Metadata[key] = value
		}
	}
	if result.Duration == 0 {
		result.Duration = time.Since(e.started)
	}
	return result
}

func (e *stepEvidence) output() task.OutputSnapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	return task.OutputSnapshot{Stdout: e.stdout.String(), Stderr: e.stderr.String()}
}

func (e *stepEvidence) snapshot() any {
	e.mu.Lock()
	defer e.mu.Unlock()
	details := maps.Clone(e.details)
	if e.command != "" {
		details["command"] = e.command
	}
	if e.cwd != "" {
		details["cwd"] = e.cwd
	}
	return details
}

// evidenceStream is Stdout or Stderr: it keeps each write, then streams it to
// the live view. A write never fails — it may be teed from a running process's
// pipe, and failing it would cut the process's own capture short.
type evidenceStream struct {
	e      *stepEvidence
	stderr bool
}

func (w evidenceStream) Write(p []byte) (int, error) {
	w.e.mu.Lock()
	if w.stderr {
		w.e.stderr.Write(p)
	} else {
		w.e.stdout.Write(p)
	}
	w.e.mu.Unlock()
	_, _ = w.e.live.Write(p)
	return len(p), nil
}
