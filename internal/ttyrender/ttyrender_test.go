package ttyrender

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/flanksource/clicky/api"
)

func TestStateWriteSecondCallClearsPrevious(t *testing.T) {
	t.Cleanup(api.InvalidateTerminalSize)
	api.SetTerminalWidth(200)

	var buf bytes.Buffer
	state := State{}

	first := "line one\nline two\nline three\n"
	if err := state.Write(&buf, first); err != nil {
		t.Fatalf("first write: %v", err)
	}
	if got := buf.String(); got != first {
		t.Fatalf("first write should pass through unchanged.\n got: %q\nwant: %q", got, first)
	}

	buf.Reset()
	second := "alpha\nbeta\n"
	if err := state.Write(&buf, second); err != nil {
		t.Fatalf("second write: %v", err)
	}
	wantPrefix := fmt.Sprintf("\x1b[%dA\x1b[J", 3)
	got := buf.String()
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("second write missing clear prefix.\n got: %q\nwant prefix: %q", got, wantPrefix)
	}
	if !strings.HasSuffix(got, second) {
		t.Fatalf("second write missing rendered payload.\n got: %q\nwant suffix: %q", got, second)
	}
}

// A frame line wider than the terminal wraps onto extra physical rows. Clearing
// only the logical line count leaves the wrapped remainder on screen.
func TestStateWriteClearsWrappedRows(t *testing.T) {
	t.Cleanup(api.InvalidateTerminalSize)
	const width = 20
	api.SetTerminalWidth(width)

	var buf bytes.Buffer
	state := State{}

	// Two logical lines: one occupying 3 rows at width 20, one occupying 1.
	frame := strings.Repeat("x", width*2+1) + "\nshort\n"
	if err := state.Write(&buf, frame); err != nil {
		t.Fatalf("first write: %v", err)
	}

	buf.Reset()
	if err := state.Write(&buf, "next\n"); err != nil {
		t.Fatalf("second write: %v", err)
	}
	wantPrefix := fmt.Sprintf("\x1b[%dA\x1b[J", 4)
	if got := buf.String(); !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("clear did not account for wrapped rows.\n got: %q\nwant prefix: %q", got, wantPrefix)
	}
}

func TestStateWriteEmptyRenderDoesNotEmitClear(t *testing.T) {
	var buf bytes.Buffer
	state := State{}
	if err := state.Write(&buf, ""); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := state.Write(&buf, "hello\n"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Fatalf("empty first render should not arm clear sequence.\n got: %q", got)
	}
}

func TestCountRows(t *testing.T) {
	const width = 10
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"single line no newline", "a", 1},
		{"trailing newline terminates", "a\n", 1},
		{"two lines", "a\nb", 2},
		{"two lines trailing newline", "a\nb\n", 2},
		{"three lines", "a\nb\nc\n", 3},
		{"exactly one row", strings.Repeat("a", width), 1},
		{"one column over wraps", strings.Repeat("a", width+1), 2},
		{"two rows exactly", strings.Repeat("a", width*2), 2},
		{"ansi escapes are not printable columns", "\x1b[38;2;1;2;3m" + strings.Repeat("a", width) + "\x1b[0m", 1},
		{"wrapped line plus short line", strings.Repeat("a", width+1) + "\nb\n", 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CountRows(c.in, width); got != c.want {
				t.Errorf("CountRows(%q, %d) = %d, want %d", c.in, width, got, c.want)
			}
		})
	}
}

func TestCountRowsFallsBackOnUnmeasurableWidth(t *testing.T) {
	line := strings.Repeat("a", 100)
	if got := CountRows(line, 0); got != 1 {
		t.Errorf("CountRows at width 0 = %d, want 1 row via the %d-column fallback", got, defaultWidth)
	}
}
