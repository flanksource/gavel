package parsers

import (
	"io"

	"github.com/flanksource/clicky/task"
)

// ResultParser defines the interface for parsing test output in different formats.
type ResultParser interface {
	// Name returns the name of this parser (e.g., "go test json").
	Name() string

	// Parse reads test output and returns test results.
	Parse(output io.Reader) ([]Test, error)

	// ParseStream reads and parses test output line-by-line, updating task progress.
	// Returns (passCount, failCount, error).
	ParseStream(output io.Reader, stdout io.Writer, t *task.Task) (int, int, error)
}
