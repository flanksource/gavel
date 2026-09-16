package fixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	osExec "os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/flanksource/gavel/fixtures/record"
)

// defaultCastBytes and maxCastEvents bound a capture. A command in a redraw
// loop emits megabytes a second, and the artifact is diagnostic evidence rather
// than an archive: past these the events stop accumulating and Truncated says
// so. Screen tracking continues, so the final state stays true.
const (
	defaultCastBytes = 4 << 20
	maxCastEvents    = 20000
)

// CaptureOptions configures an ANSI/PTY capture. Width and Height are both the
// pseudo-terminal dimensions the command sees and the viewport dimensions used
// by settled snapshots. SnapshotInterval controls how often the live viewport
// is appended to the snapshot timeline.
type CaptureOptions struct {
	Context          context.Context
	Width, Height    int
	SnapshotInterval time.Duration
	Command          []string
	Env              []string // extra env appended to os.Environ()
	// Dir is the working directory of the command. Empty inherits gavel's,
	// which is almost never what a fixture wants.
	Dir string
	// Snapshots turns on settled-screen tracking: the snapshot timeline, the
	// final screen, its duplicate lines and MaxLineWidth. It is off by default
	// because settling the screen on every chunk and every tick is the expensive
	// half of a capture, and a fixture that only asked for `terminal: pty` wants
	// the output stream, not an analysis of it.
	Snapshots bool
	// MaxBytes caps the recorded event stream. Zero uses defaultCastBytes.
	MaxBytes int64
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

// Snapshot is the width- and height-bounded settled screen at a point in time.
type Snapshot struct {
	TMs    int64  `json:"t_ms"`
	Screen string `json:"screen"`
}

// FinalState is the visible viewport at end of stream plus the duplicate-line
// report — duplicates are the tell-tale of a redraw that left stale content.
type FinalState struct {
	Screen     string    `json:"screen"`
	Duplicates []DupLine `json:"duplicates"`
}

// Capture is the full record of a PTY run: asciinema-compatible timed events
// plus a timeline of settled-screen snapshots.
type Capture struct {
	Version      int        `json:"version"`
	Width        int        `json:"width"`
	Height       int        `json:"height"`
	Command      []string   `json:"command"`
	ExitCode     int        `json:"exit_code"`
	DurationMs   int64      `json:"duration_ms"`
	Events       []Event    `json:"events"`
	Snapshots    []Snapshot `json:"snapshots"`
	Final        FinalState `json:"final"`
	Truncated    bool       `json:"truncated,omitempty"`
	raw          string
	maxLineWidth int
}

// MaxLineWidth returns the widest unwrapped line in the final terminal state.
func (c *Capture) MaxLineWidth() int { return c.maxLineWidth }

// Raw is the output stream as the child wrote it, escape sequences and all —
// the merged stdout+stderr a PTY presents. Unlike Events it is never truncated,
// because it is what a fixture's assertions run against.
func (c *Capture) Raw() string { return c.raw }

// Cast converts a capture into the artifact form record knows how to write.
// The conversion lives here rather than in record because record deliberately
// imports nothing from gavel.
func (c *Capture) Cast() record.Cast {
	cast := record.Cast{
		Width:     c.Width,
		Height:    c.Height,
		Command:   c.Command,
		ExitCode:  c.ExitCode,
		Duration:  time.Duration(c.DurationMs) * time.Millisecond,
		Final:     c.Final.Screen,
		Truncated: c.Truncated,
	}
	for _, event := range c.Events {
		cast.Events = append(cast.Events, record.CastEvent{
			Offset: time.Duration(event.Time * float64(time.Second)),
			Data:   event.Data,
		})
	}
	for _, snapshot := range c.Snapshots {
		cast.Snapshots = append(cast.Snapshots, record.CastSnapshot{
			Offset: time.Duration(snapshot.TMs) * time.Millisecond,
			Screen: snapshot.Screen,
		})
	}
	for _, duplicate := range c.Final.Duplicates {
		cast.Duplicates = append(cast.Duplicates, record.CastDuplicate{Text: duplicate.Text, Count: duplicate.Count})
	}
	return cast
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

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("ansi capture: %w", err)
	}
	cmd := osExec.Command(opts.Command[0], opts.Command[1:]...)
	cmd.Dir = opts.Dir
	cmd.Env = append(os.Environ(), opts.Env...)
	cmd.Env = ensureEnv(cmd.Env, "CLICKY_FORCE_INTERACTIVE", "1")
	cmd.Env = ensureEnv(cmd.Env, "TERM", "xterm-256color")

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: uint16(opts.Height), Cols: uint16(opts.Width)})
	if err != nil {
		return nil, fmt.Errorf("ansi capture: start pty: %w", err)
	}
	defer ptmx.Close()
	readDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			// Kill before closing the PTY so descendants that created their own
			// process groups remain attached long enough to be discovered.
			_ = cancelCaptureProcess(cmd.Process)
			_ = ptmx.Close()
		case <-readDone:
		}
	}()

	var (
		mu        sync.Mutex
		viewport  = newANSIScreen(opts.Width, opts.Height)
		unwrapped = newANSIScreen(0, 0)
		raw       strings.Builder
		events    []Event
		snaps     []Snapshot
	)
	start := time.Now()

	done := make(chan struct{})
	var wg sync.WaitGroup
	if opts.Snapshots {
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
						Screen: viewport.String(),
					})
					mu.Unlock()
				}
			}
		}()
	}

	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultCastBytes
	}
	var truncated bool

	buf := make([]byte, 4096)
	for {
		n, rerr := ptmx.Read(buf)
		if n > 0 {
			mu.Lock()
			// raw is never capped: it is the fixture's stdout, and silently
			// dropping bytes an assertion is written against would be worse than
			// a large artifact. The caps below bound only the timed event stream,
			// and screen tracking sees every byte either way.
			raw.Write(buf[:n])
			if opts.Snapshots {
				viewport.Write(buf[:n])
				unwrapped.Write(buf[:n])
			}
			if int64(raw.Len()) > maxBytes || len(events) >= maxCastEvents {
				truncated = true
			} else {
				events = append(events, Event{Time: time.Since(start).Seconds(), Data: string(buf[:n])})
			}
			mu.Unlock()
		}
		if rerr != nil {
			// EOF/EIO on the master once the child exits is expected.
			break
		}
	}
	close(readDone)
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

	elapsedMs := time.Since(start).Milliseconds()
	capture := &Capture{
		Version:    2,
		Width:      opts.Width,
		Height:     opts.Height,
		Command:    opts.Command,
		ExitCode:   exitCode,
		DurationMs: elapsedMs,
		Events:     events,
		Truncated:  truncated,
		raw:        raw.String(),
	}
	if opts.Snapshots {
		final := viewport.String()
		capture.Snapshots = append(snaps, Snapshot{TMs: elapsedMs, Screen: final})
		capture.maxLineWidth = unwrapped.MaxLineWidth()
		capture.Final = FinalState{Screen: final, Duplicates: duplicateSettledLines(final)}
	}
	return capture, nil
}

func captureProcessDescendants(root int32) ([]*process.Process, error) {
	all, err := process.Processes()
	if err != nil {
		return nil, err
	}
	children := make(map[int32][]*process.Process)
	for _, candidate := range all {
		parent, err := candidate.Ppid()
		if err == nil {
			children[parent] = append(children[parent], candidate)
		}
	}

	var descendants []*process.Process
	parents := []int32{root}
	seen := map[int32]bool{root: true}
	for len(parents) > 0 {
		parent := parents[0]
		parents = parents[1:]
		for _, child := range children[parent] {
			if seen[child.Pid] {
				continue
			}
			seen[child.Pid] = true
			descendants = append(descendants, child)
			parents = append(parents, child.Pid)
		}
	}
	return descendants, nil
}

func killCaptureDescendants(root int) error {
	descendants, err := captureProcessDescendants(int32(root))
	failures := []error{err}
	// Leaves first prevents a killed parent from orphaning children before they
	// receive the cancellation signal.
	for i := len(descendants) - 1; i >= 0; i-- {
		if err := descendants[i].Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
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
