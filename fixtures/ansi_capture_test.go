package fixtures

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCaptureANSI_RecordsTimedEventsAndSnapshots(t *testing.T) {
	cap, err := CaptureANSI(CaptureOptions{
		Width:            40,
		Height:           10,
		SnapshotInterval: 50 * time.Millisecond,
		Command:          []string{"/bin/sh", "-c", "printf 'hello\\n'; sleep 0.2; printf 'world\\n'"},
		Snapshots:        true,
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	if cap.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", cap.ExitCode)
	}
	if len(cap.Events) == 0 {
		t.Fatalf("expected at least one output event")
	}
	prev := -1.0
	for i, e := range cap.Events {
		if e.Time < 0 {
			t.Fatalf("event %d has negative time %f", i, e.Time)
		}
		if e.Time < prev {
			t.Fatalf("event %d time %f went backwards from %f", i, e.Time, prev)
		}
		prev = e.Time
	}
	if len(cap.Snapshots) == 0 {
		t.Fatalf("expected at least the final snapshot")
	}
	last := cap.Snapshots[len(cap.Snapshots)-1]
	if last.Screen != cap.Final.Screen {
		t.Fatalf("last snapshot screen != final screen:\n last=%q\nfinal=%q", last.Screen, cap.Final.Screen)
	}
	if !strings.Contains(cap.Final.Screen, "hello") || !strings.Contains(cap.Final.Screen, "world") {
		t.Fatalf("final screen missing output: %q", cap.Final.Screen)
	}
}

func TestCaptureANSI_WrapInducedSmearIsCaptured(t *testing.T) {
	// Same shape as the settle unit test, but driven through a real PTY: a row
	// wraps at width 8, then a cursor-up sized by logical-line count (1) lands
	// on the wrapped tail, so the redraw leaves the row's first physical line
	// "ROWADDXX" stale — the status-screen smear.
	cap, err := CaptureANSI(CaptureOptions{
		Width:     8,
		Height:    10,
		Command:   []string{"/bin/sh", "-c", "printf 'HEADER\\nROWADDXXXXXX\\n\\033[1A\\033[2KNEW'"},
		Snapshots: true,
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	if !strings.Contains(cap.Final.Screen, "ROWADDXX") {
		t.Fatalf("expected stale wrapped row to survive in settled screen, got %q", cap.Final.Screen)
	}
	if !strings.Contains(cap.Final.Screen, "NEW") {
		t.Fatalf("expected redraw text 'NEW', got %q", cap.Final.Screen)
	}
}

func TestCaptureANSI_PropagatesExitCode(t *testing.T) {
	cap, err := CaptureANSI(CaptureOptions{
		Width:   40,
		Height:  10,
		Command: []string{"/bin/sh", "-c", "exit 3"},
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	if cap.ExitCode != 3 {
		t.Fatalf("exit code = %d, want 3", cap.ExitCode)
	}
}

// Dir is what the previous PTY implementation in fixtures/types had and this
// one lacked; merging them the other way round would have silently run every
// pty fixture in gavel's own working directory.
func TestCaptureANSI_RunsInDir(t *testing.T) {
	dir := t.TempDir()
	cap, err := CaptureANSI(CaptureOptions{
		Width:   80,
		Height:  10,
		Dir:     dir,
		Command: []string{"/bin/sh", "-c", "pwd"},
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	// macOS resolves TMPDIR through /private, so compare the resolved paths.
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if !strings.Contains(cap.Raw(), want) {
		t.Fatalf("expected the command to run in %q, got %q", want, cap.Raw())
	}
}

// Without Snapshots a capture is only the output stream: no timeline, no
// settled screen, no duplicate analysis — the cheap path a plain `terminal: pty`
// fixture takes.
func TestCaptureANSI_WithoutSnapshotsSkipsScreenTracking(t *testing.T) {
	cap, err := CaptureANSI(CaptureOptions{
		Width:            40,
		Height:           10,
		SnapshotInterval: time.Millisecond,
		Command:          []string{"/bin/sh", "-c", "printf 'streamed\\n'; sleep 0.05"},
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	if len(cap.Snapshots) != 0 {
		t.Fatalf("expected no snapshots, got %d", len(cap.Snapshots))
	}
	if cap.Final.Screen != "" || cap.MaxLineWidth() != 0 {
		t.Fatalf("expected no settled screen, got %q / width %d", cap.Final.Screen, cap.MaxLineWidth())
	}
	if !strings.Contains(cap.Raw(), "streamed") {
		t.Fatalf("raw stream missing output: %q", cap.Raw())
	}
}

// The event stream is capped, the stream a fixture asserts on is not: dropping
// stdout bytes to bound an artifact would fail tests for a diagnostic.
func TestCaptureANSI_TruncatesEventsButNotRaw(t *testing.T) {
	const lines = 200
	cap, err := CaptureANSI(CaptureOptions{
		Width:    40,
		Height:   10,
		MaxBytes: 64,
		Command:  []string{"/bin/sh", "-c", fmt.Sprintf("i=0; while [ $i -lt %d ]; do echo padding-line; i=$((i+1)); done", lines)},
	})
	if err != nil {
		t.Fatalf("CaptureANSI: %v", err)
	}
	if !cap.Truncated {
		t.Fatalf("expected the cast to report truncation")
	}
	if got := strings.Count(cap.Raw(), "padding-line"); got != lines {
		t.Fatalf("raw stream lost output: %d of %d lines", got, lines)
	}
}

func TestCaptureANSI_RejectsInvalidOptions(t *testing.T) {
	if _, err := CaptureANSI(CaptureOptions{Width: 80, Height: 24}); err == nil {
		t.Fatalf("expected error for missing command")
	}
	if _, err := CaptureANSI(CaptureOptions{Width: 0, Height: 24, Command: []string{"/bin/sh", "-c", "true"}}); err == nil {
		t.Fatalf("expected error for non-positive width")
	}
}
