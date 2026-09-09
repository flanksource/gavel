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

	"charm.land/lipgloss/v2"
	"github.com/flanksource/clicky/api"
	"golang.org/x/term"
)

// State tracks how many rows were last written so the next write can clear
// them before redrawing. A zero State is ready to use.
type State struct {
	rows int
}

// Write clears the previously written block (if any) and writes rendered to w.
// rendered is the complete next frame; callers do not need to pre-strip the
// old output.
func (s *State) Write(w io.Writer, rendered string) error {
	if s.rows > 0 {
		if _, err := fmt.Fprintf(w, "\x1b[%dA\x1b[J", s.rows); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, rendered); err != nil {
		return err
	}
	s.rows = CountRows(rendered, api.GetTerminalWidth())
	return nil
}

// Clear erases the previously written block and leaves the cursor where it
// began, so a caller that finishes by printing elsewhere does not stack its
// final output under the last live frame.
func (s *State) Clear(w io.Writer) error {
	if s.rows == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "\x1b[%dA\x1b[J", s.rows); err != nil {
		return err
	}
	s.rows = 0
	return nil
}

// CountRows returns the number of physical terminal rows rendered occupies at
// the given width.
//
// Counting newlines is not enough: a logical line wider than the terminal
// soft-wraps onto extra rows, so a newline count under-reports how far the
// cursor has advanced. The next frame's cursor-up would then move too few rows
// and leave the tail of the previous frame on screen, smearing frames together.
// A trailing newline terminates the last line rather than starting an empty one.
func CountRows(rendered string, width int) int {
	if rendered == "" {
		return 0
	}
	if width <= 0 {
		width = defaultWidth
	}

	rows := 0
	for _, line := range strings.Split(strings.TrimSuffix(rendered, "\n"), "\n") {
		// lipgloss.Width measures printable columns, excluding ANSI escapes.
		visible := lipgloss.Width(line)
		if visible <= width {
			rows++
			continue
		}
		rows += (visible-1)/width + 1
	}
	return rows
}

// defaultWidth mirrors clicky's fallback for an unmeasurable terminal.
const defaultWidth = 120

// IsTerminal reports whether w is an *os.File attached to a terminal.
// Non-file writers (bytes.Buffer, piped stdout, etc.) return false.
func IsTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}
