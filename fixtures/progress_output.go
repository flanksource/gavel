package fixtures

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
	"unicode/utf8"
)

// MaxExecutionOutput is how much of a running fixture's live output its
// progress node keeps: the tail, so the latest lines are always there.
const MaxExecutionOutput = 64 << 10

// outputPublishInterval bounds how often one node's output publishes a
// snapshot, so a chatty process cannot flood the sink.
const outputPublishInterval = 250 * time.Millisecond

// outputStream is the io.Writer OutputWriter hands one fixture: each write
// lands on the fixture's node, the first of a burst publishes at once and the
// rest publish together once the interval has passed.
//
// A write never fails: it is teed from a running command's pipe, and failing
// it would cut the command's own capture short. A failed publish is kept on
// the tracker instead and fails the fixture's Complete.
type outputStream struct {
	tracker   *executionTracker
	ctx       context.Context
	fixture   *FixtureNode
	mu        sync.Mutex
	published time.Time
	timer     *time.Timer
}

// OutputWriter is the stream fixture writes its live output to. A tracker
// nobody watches has no one to publish to, so its fixtures' output is
// discarded.
func (t *executionTracker) OutputWriter(ctx context.Context, fixture *FixtureNode) io.Writer {
	if t == nil || t.sink == nil {
		return io.Discard
	}
	return &outputStream{tracker: t, ctx: ctx, fixture: fixture}
}

func (s *outputStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.tracker.appendOutput(s.fixture, p) {
		return len(p), nil
	}
	wait := outputPublishInterval - time.Since(s.published)
	if wait <= 0 {
		s.publishLocked()
	} else if s.timer == nil {
		s.timer = time.AfterFunc(wait, s.flush)
	}
	return len(p), nil
}

// flush publishes what the writes since the last publish appended, unless the
// fixture finished meanwhile: Complete published its final state.
func (s *outputStream) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timer = nil
	if s.tracker.running(s.fixture) {
		s.publishLocked()
	}
}

func (s *outputStream) publishLocked() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.published = time.Now()
	if err := s.tracker.Publish(s.ctx); err != nil {
		s.tracker.failOutput(s.fixture, fmt.Errorf("publish %s output: %w", s.fixture.Name, err))
	}
}

func (t *executionTracker) running(fixture *FixtureNode) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	node, ok := t.fixture[fixture]
	return ok && node.FinishedAt == nil && (node.State == ExecutionRunning || node.State == ExecutionWaiting)
}

// appendOutput adds p to fixture's node output, keeping its tail, and reports
// whether it did. Output a process writes after its fixture finished (a
// background service still logging) is not the node's: a finished node shows
// its Result. A fixture the tracker does not hold fails its Complete.
func (t *executionTracker) appendOutput(fixture *FixtureNode, p []byte) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	node, ok := t.fixture[fixture]
	if !ok {
		t.failOutputLocked(fixture, fmt.Errorf("fixture progress node not found: %q", fixture.Name))
		return false
	}
	if node.FinishedAt != nil {
		return false
	}
	output := node.Output + string(p)
	if len(output) > MaxExecutionOutput {
		cut := len(output) - MaxExecutionOutput
		for cut < len(output) && !utf8.RuneStart(output[cut]) {
			cut++
		}
		output, node.OutputTruncated = output[cut:], true
	}
	node.Output = output
	return true
}

func (t *executionTracker) failOutput(fixture *FixtureNode, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.failOutputLocked(fixture, err)
}

func (t *executionTracker) failOutputLocked(fixture *FixtureNode, err error) {
	if t.outputErr == nil {
		t.outputErr = map[*FixtureNode]error{}
	}
	if t.outputErr[fixture] == nil {
		t.outputErr[fixture] = err
	}
}

// takeOutputError is the first failure streaming fixture's output met.
func (t *executionTracker) takeOutputError(fixture *FixtureNode) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	err := t.outputErr[fixture]
	delete(t.outputErr, fixture)
	return err
}
