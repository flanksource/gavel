package testrunner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/traces"
)

// attemptStamp is what one subprocess run records on every test it produced.
type attemptStamp struct {
	RunKind  string
	Started  time.Time
	Ended    time.Time
	PID      int
	ExitCode int
	// Usage is the process's key and peaks over the run, read from the run's
	// process_samples; zero without a process trace.
	Usage traces.ProcessUsage
	// SharedProcess: the PID and metrics are a whole go test batch's.
	SharedProcess bool
}

// stampAttempts appends one TestAttempt per test (leaves and groups) capturing
// the command, PID, framework and final state of the subprocess run. Existing
// attempts (earlier runs, reruns merged via the streamer) are preserved.
func stampAttempts(results parsers.TestSuiteResults, stamp attemptStamp) {
	if stamp.RunKind == "" {
		stamp.RunKind = "initial"
	}
	for ri := range results {
		for ti := range results[ri].Tests {
			stampAttemptSubtree(&results[ri].Tests[ti], stamp, results[ri].Command, results[ri].Framework)
		}
	}
}

func stampAttemptSubtree(t *parsers.Test, stamp attemptStamp, cmd string, framework parsers.Framework) {
	for i := range t.Children {
		stampAttemptSubtree(&t.Children[i], stamp, cmd, framework)
	}
	code := stamp.ExitCode
	attempt := parsers.TestAttempt{
		Sequence: len(t.Attempts) + 1, RunKind: stamp.RunKind,
		Started: stamp.Started, Ended: stamp.Ended, Duration: t.Duration,
		PID: stamp.PID, ProcessKey: stamp.Usage.Key, Command: cmd, Framework: framework, ExitCode: &code,
		Passed: t.Passed, Failed: t.Failed, Skipped: t.Skipped, Pending: t.Pending, TimedOut: t.TimedOut,
		Message: t.Message, Stdout: t.Stdout, Stderr: t.Stderr,
		CPUPercent: stamp.Usage.PeakCPUPercent, RSS: stamp.Usage.PeakRSSBytes,
		SharedProcess: stamp.SharedProcess,
	}
	// Go's -timeout, a SIGQUIT and ginkgo's --poll-progress-after all print
	// the goroutine dump to the test's output before it is killed.
	if t.TimedOut {
		attempt.StackTrace = extractStackTrace(t.Stdout, t.Stderr)
		attempt.GoroutineCt = goroutineCount(attempt.StackTrace)
	}
	t.Attempts = append(t.Attempts, attempt)
}

// processSpan is where a subprocess's run begins in the run's process
// samples; usage reads what pid used since. Without a process trace both are
// no-ops.
type processSpan struct {
	trace *traces.ProcessTrace
	mark  traces.ProcessMark
}

func (o *TestOrchestrator) beginProcessSpan() (processSpan, error) {
	span := processSpan{trace: o.ProcessTrace}
	if span.trace == nil {
		return span, nil
	}
	var err error
	if span.mark, err = span.trace.Mark(context.Background()); err != nil {
		return span, fmt.Errorf("mark the process samples: %w", err)
	}
	return span, nil
}

func (s processSpan) usage(pid int) (traces.ProcessUsage, error) {
	if s.trace == nil || pid <= 0 {
		return traces.ProcessUsage{}, nil
	}
	readings, err := s.trace.Readings(context.Background(), s.mark)
	if err != nil {
		return traces.ProcessUsage{}, fmt.Errorf("read the process samples of pid %d: %w", pid, err)
	}
	return readings.Usage(pid), nil
}

// extractStackTrace scans captured streams for the start of a Go panic /
// goroutine-dump block and returns from that marker to EOF. Returns "" when
// no recognizable dump is present — better empty than invented noise.
func extractStackTrace(stdout, stderr string) string {
	for _, src := range []string{stderr, stdout} {
		if src == "" {
			continue
		}
		markers := []string{
			"panic: test timed out",
			"SIGQUIT: quit",
			"\ngoroutine 1 [",
			"\nruntime: goroutine stack exceeds",
		}
		for _, m := range markers {
			if idx := strings.Index(src, m); idx >= 0 {
				return strings.TrimSpace(src[idx:])
			}
		}
	}
	return ""
}
