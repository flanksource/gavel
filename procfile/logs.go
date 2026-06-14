package procfile

import (
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
	entries, err := Load(pf)
	if err != nil {
		return err
	}
	selected, err := Select(entries, names)
	if err != nil {
		return err
	}

	width := 0
	for _, e := range selected {
		if len(e.Name) > width {
			width = len(e.Name)
		}
	}

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
	return followLogs(dir, selected, width, w)
}

// followLogs polls each process log for growth and forwards new lines (prefixed
// with the process name) to w. It runs until the process is interrupted.
func followLogs(dir string, entries []Entry, width int, w io.Writer) error {
	var mu sync.Mutex
	writers := make(map[string]*prefixWriter, len(entries))
	offsets := make(map[string]int64, len(entries))
	for _, e := range entries {
		writers[e.Name] = newPrefixWriter(w, &mu, e.Name, width, -1)
		if info, err := os.Stat(LogPath(dir, e.Name)); err == nil {
			offsets[e.Name] = info.Size()
		}
	}
	for {
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
		time.Sleep(followPollInterval)
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
