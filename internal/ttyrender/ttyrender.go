// Package ttyrender provides helpers for redrawing output in place on a TTY.
//
// Use State for a single output stream that you want to re-render repeatedly
// (e.g. a status block refreshed on a polling interval). Each call to Write
// clears the previously written block before emitting the new one, so the
// terminal shows a stable in-place update instead of stacking blocks.
package ttyrender

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// State tracks how many lines were last written so the next write can clear
// them before redrawing. A zero State is ready to use.
type State struct {
	lines int
}

// Write clears the previously written block (if any) and writes rendered to w.
// rendered is the complete next frame; callers do not need to pre-strip the
// old output.
func (s *State) Write(w io.Writer, rendered string) error {
	if s.lines > 0 {
		if _, err := fmt.Fprintf(w, "\x1b[%dA\x1b[J", s.lines); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, rendered); err != nil {
		return err
	}
	s.lines = CountLines(rendered)
	return nil
}

// CountLines returns the number of terminal lines rendered will occupy.
// A trailing newline is treated as a line terminator, not as a separate
// empty line.
func CountLines(rendered string) int {
	if rendered == "" {
		return 0
	}
	lines := strings.Count(rendered, "\n")
	if strings.HasSuffix(rendered, "\n") {
		return lines
	}
	return lines + 1
}

// IsTerminal reports whether w is an *os.File attached to a terminal.
// Non-file writers (bytes.Buffer, piped stdout, etc.) return false.
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}
