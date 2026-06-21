package fixtures

import (
	"encoding/json"
	"fmt"
	"os"
	osExec "os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
)

// CaptureOptions configures an ANSI/PTY capture. Width and Height are the
// pseudo-terminal dimensions the command sees (and the width the settled
// snapshots wrap at). SnapshotInterval controls how often the live screen is
// settled into the snapshot timeline.
type CaptureOptions struct {
	Width, Height    int
	SnapshotInterval time.Duration
	Command          []string
	Env              []string // extra env appended to os.Environ()
}

// Event is a single asciinema-v2 output event. It marshals to the canonical
// [time_seconds, "o", data] triple so captures replay in asciinema players.
type Event struct {
	Time float64
	Data string
}

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{e.Time, "o", e.Data})
}

// Snapshot is the width-aware settled screen at a point in time.
type Snapshot struct {
	TMs    int64  `json:"t_ms"`
	Screen string `json:"screen"`
}

// FinalState is the settled screen at end of stream plus the duplicate-line
// report — duplicates are the tell-tale of a redraw that left stale content.
type FinalState struct {
	Screen     string    `json:"screen"`
	Duplicates []DupLine `json:"duplicates"`
}

// Capture is the full record of a PTY run: asciinema-compatible timed events
// plus a timeline of settled-screen snapshots.
type Capture struct {
	Version    int        `json:"version"`
	Width      int        `json:"width"`
	Height     int        `json:"height"`
	Command    []string   `json:"command"`
	ExitCode   int        `json:"exit_code"`
	DurationMs int64      `json:"duration_ms"`
	Events     []Event    `json:"events"`
	Snapshots  []Snapshot `json:"snapshots"`
	Final      FinalState `json:"final"`
}

// CaptureANSI runs opts.Command under a PTY of the given size, recording every
// output chunk as a timed asciinema event and settling the screen on a fixed
// interval. Stdout and stderr are merged onto the single PTY exactly as a real
// terminal presents them.
func CaptureANSI(opts CaptureOptions) (*Capture, error) {
	if len(opts.Command) == 0 {
		return nil, fmt.Errorf("ansi capture: command is required")
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		return nil, fmt.Errorf("ansi capture: width and height must be > 0 (got %dx%d)", opts.Width, opts.Height)
	}
	interval := opts.SnapshotInterval
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}

	cmd := osExec.Command(opts.Command[0], opts.Command[1:]...)
	cmd.Env = append(os.Environ(), opts.Env...)
	cmd.Env = ensureEnv(cmd.Env, "CLICKY_FORCE_INTERACTIVE", "1")
	cmd.Env = ensureEnv(cmd.Env, "TERM", "xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(opts.Height), Cols: uint16(opts.Width)})
	if err != nil {
		return nil, fmt.Errorf("ansi capture: start pty: %w", err)
	}
	defer ptmx.Close()

	var (
		mu     sync.Mutex
		raw    strings.Builder
		events []Event
		snaps  []Snapshot
	)
	start := time.Now()

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				mu.Lock()
				snaps = append(snaps, Snapshot{
					TMs:    time.Since(start).Milliseconds(),
					Screen: settleANSI(raw.String(), opts.Width),
				})
				mu.Unlock()
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, rerr := ptmx.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			mu.Lock()
			raw.WriteString(chunk)
			events = append(events, Event{Time: time.Since(start).Seconds(), Data: chunk})
			mu.Unlock()
		}
		if rerr != nil {
			// EOF/EIO on the master once the child exits is expected.
			break
		}
	}
	close(done)
	wg.Wait()

	exitCode := 0
	if werr := cmd.Wait(); werr != nil {
		ee, ok := werr.(*osExec.ExitError)
		if !ok {
			return nil, fmt.Errorf("ansi capture: wait for %q: %w", opts.Command[0], werr)
		}
		exitCode = ee.ExitCode()
	}

	final := raw.String()
	elapsedMs := time.Since(start).Milliseconds()
	snaps = append(snaps, Snapshot{TMs: elapsedMs, Screen: settleANSI(final, opts.Width)})

	return &Capture{
		Version:    2,
		Width:      opts.Width,
		Height:     opts.Height,
		Command:    opts.Command,
		ExitCode:   exitCode,
		DurationMs: elapsedMs,
		Events:     events,
		Snapshots:  snaps,
		Final: FinalState{
			Screen:     settleANSI(final, opts.Width),
			Duplicates: duplicateLines(final, opts.Width),
		},
	}, nil
}

func ensureEnv(env []string, key, val string) []string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return env
		}
	}
	return append(env, prefix+val)
}
