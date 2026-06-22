package procfile

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/flanksource/gavel/utils"
)

const followPollInterval = 250 * time.Millisecond

// Logs writes the last `lines` lines of each selected process log to w. With
// follow=true it then streams new output until the process is interrupted.
// names selects a subset; empty means every process in the Procfile.
func Logs(rootOverride, pfOverride string, names []string, lines int, follow bool, w io.Writer) error {
	root, pf, err := resolveTarget(rootOverride, pfOverride)
	if err != nil {
		return err
	}
	dir, err := StateDir(root)
	if err != nil {
		return err
	}
	selected, err := selectEntries(pf, names)
	if err != nil {
		return err
	}

	width := prefixWidth(selected)
	for _, e := range selected {
		tail, err := utils.TailFile(LogPath(dir, e.Name), lines)
		if err != nil {
			return err
		}
		for _, line := range tail {
			fmt.Fprintf(w, "%-*s | %s\n", width, e.Name, line)
		}
	}
	if !follow {
		return nil
	}
	writers, offsets := newTailState(dir, selected, w, width, false)
	return followLoop(context.Background(), dir, selected, writers, offsets)
}

// Stream tails the last `tailLines` of each selected process log to w, then
// follows new output — each line prefixed with a coloured "name |" — until ctx
// is cancelled. It is the backend for `proc restart`/`proc stop`, which tail the
// affected processes' output live while the operation settles; colour and
// per-process prefixes mirror `proc run`'s foreground multiplexing.
func Stream(ctx context.Context, rootOverride, pfOverride string, names []string, tailLines int, w io.Writer) error {
	root, pf, err := resolveTarget(rootOverride, pfOverride)
	if err != nil {
		return err
	}
	dir, err := StateDir(root)
	if err != nil {
		return err
	}
	selected, err := selectEntries(pf, names)
	if err != nil {
		return err
	}

	width := prefixWidth(selected)
	writers, offsets := newTailState(dir, selected, w, width, true)
	// Emit the trailing context once. offsets were seeded to the current size, so
	// the follow loop forwards only output appended after this point.
	for _, e := range selected {
		tail, err := utils.TailFile(LogPath(dir, e.Name), tailLines)
		if err != nil {
			return err
		}
		for _, line := range tail {
			if _, err := writers[e.Name].Write([]byte(line + "\n")); err != nil {
				return err
			}
		}
	}
	return followLoop(ctx, dir, selected, writers, offsets)
}

// selectEntries loads the Procfile and returns the named subset (empty = all).
func selectEntries(pf string, names []string) ([]Entry, error) {
	entries, err := Load(pf)
	if err != nil {
		return nil, err
	}
	return Select(entries, names)
}

// prefixWidth is the column width that pads process-name prefixes so output from
// multiple processes lines up.
func prefixWidth(entries []Entry) int {
	width := 0
	for _, e := range entries {
		if len(e.Name) > width {
			width = len(e.Name)
		}
	}
	return width
}

// newTailState builds one prefix writer per process (sharing a mutex so writes
// stay line-atomic across processes) and seeds each offset to the current log
// size, so a following loop forwards only subsequent growth. colour selects the
// per-process palette (matching `proc run`); when false the prefixes are
// uncoloured (matching `proc logs`).
func newTailState(dir string, entries []Entry, w io.Writer, width int, colour bool) (map[string]*prefixWriter, map[string]int64) {
	mu := &sync.Mutex{}
	writers := make(map[string]*prefixWriter, len(entries))
	offsets := make(map[string]int64, len(entries))
	for i, e := range entries {
		colorIdx := -1
		if colour {
			colorIdx = i
		}
		writers[e.Name] = newPrefixWriter(w, mu, e.Name, width, colorIdx)
		if info, err := os.Stat(LogPath(dir, e.Name)); err == nil {
			offsets[e.Name] = info.Size()
		}
	}
	return writers, offsets
}

// followLoop forwards bytes appended to each entry's log to its prefix writer
// until ctx is cancelled, polling every followPollInterval. A truncated/rotated
// file (size < offset) is re-read from the start. A final drain runs on
// cancellation so the last lines are never dropped.
func followLoop(ctx context.Context, dir string, entries []Entry, writers map[string]*prefixWriter, offsets map[string]int64) error {
	drain := func() error {
		for _, e := range entries {
			path := LogPath(dir, e.Name)
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			size := info.Size()
			prev := offsets[e.Name]
			if size < prev {
				prev = 0 // file was truncated/rotated
			}
			if size > prev {
				chunk, err := readRange(path, prev, size)
				if err != nil {
					return err
				}
				if _, err := writers[e.Name].Write(chunk); err != nil {
					return err
				}
				offsets[e.Name] = size
			}
		}
		return nil
	}
	for {
		if err := drain(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return drain() // catch output appended since the last poll
		case <-time.After(followPollInterval):
		}
	}
}

func readRange(path string, from, to int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, to-from)
	n, err := f.ReadAt(buf, from)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return buf[:n], nil
}
