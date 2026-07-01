package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

// ProcessAlive returns true if signal 0 can be delivered to pid (i.e. the
// process exists and belongs to the current user). A non-positive pid is never
// alive. EPERM — "exists but not ours" — is treated as "not ours" (false).
func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// TailFile returns the last n lines of the file at path. A missing file is
// treated as "no content yet" and returns an empty slice (no error) so callers
// can tail logs that may not exist yet. Lines are returned without trailing
// newlines, in original order (oldest first).
//
// For large files the tail is read in 8KiB chunks from the end rather than
// loading the whole file — process logs can grow to many MB over time.
func TailFile(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	size := info.Size()
	if size == 0 {
		return nil, nil
	}

	const chunk = 8 * 1024
	var (
		buf  []byte
		off  = size
		seen int
	)
	for off > 0 && seen <= n {
		read := min(int64(chunk), off)
		off -= read
		tmp := make([]byte, read)
		if _, err := f.ReadAt(tmp, off); err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		buf = append(tmp, buf...)
		seen = strings.Count(string(buf), "\n")
	}

	lines := strings.Split(strings.TrimRight(string(buf), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}
