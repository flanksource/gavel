// Package logtest captures the global commons logger in tests, so the lines a test expects are
// asserted instead of printed.
package logtest

import (
	"strings"
	"sync"
	"testing"

	"github.com/flanksource/commons/logger"
)

// Lines is what the global logger wrote during a test. A test may read it while other
// goroutines are still logging.
type Lines struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (l *Lines) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

// String is every line logged so far.
func (l *Lines) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.String()
}

// Capture sends the global logger to the returned Lines for the rest of t.
func Capture(t testing.TB) *Lines {
	t.Helper()
	lines := &Lines{}
	previous := logger.StandardLogger()
	logger.Use(lines)
	t.Cleanup(func() { logger.SetLogger(previous) })
	return lines
}
