//go:build !windows

package status

import (
	"errors"
	"syscall"
)

// Alive reports whether a process with pid exists, including one owned by
// another user.
func Alive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
