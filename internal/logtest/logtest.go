// Package logtest captures the global commons logger in tests, so the lines a test expects are
// asserted instead of printed.
package logtest

import (
	"strings"
	"testing"

	"github.com/flanksource/commons/logger"
)

// Capture sends the global logger to the returned buffer for the rest of t.
func Capture(t testing.TB) *strings.Builder {
	t.Helper()
	var buf strings.Builder
	previous := logger.StandardLogger()
	logger.Use(&buf)
	t.Cleanup(func() { logger.SetLogger(previous) })
	return &buf
}
