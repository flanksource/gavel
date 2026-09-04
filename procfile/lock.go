package procfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

var errLockHeld = errors.New("lock is held")

func launchLockPath(dir string) string {
	return filepath.Join(dir, "launch.lock")
}

func supervisorLockPath(dir string) string {
	return filepath.Join(dir, "supervisor.lock")
}

func acquireFileLock(path string, nonBlocking bool) (*os.File, error) {
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock %s: %w", path, err)
	}
	operation := unix.LOCK_EX
	if nonBlocking {
		operation |= unix.LOCK_NB
	}
	if err := unix.Flock(int(fd.Fd()), operation); err != nil {
		_ = fd.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, fmt.Errorf("%w: %s", errLockHeld, path)
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return fd, nil
}

func releaseFileLock(fd *os.File) error {
	if fd == nil {
		return nil
	}
	return errors.Join(unix.Flock(int(fd.Fd()), unix.LOCK_UN), fd.Close())
}
