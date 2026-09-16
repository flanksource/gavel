//go:build unix

package fixtures

import (
	"errors"
	"os"
	"syscall"
)

func cancelCaptureProcess(process *os.Process) error {
	if process == nil {
		return os.ErrProcessDone
	}
	stopErr := syscall.Kill(-process.Pid, syscall.SIGSTOP)
	if errors.Is(stopErr, syscall.ESRCH) {
		stopErr = nil
	}
	descendantsErr := killCaptureDescendants(process.Pid)
	groupErr := syscall.Kill(-process.Pid, syscall.SIGKILL)
	if errors.Is(groupErr, syscall.ESRCH) {
		groupErr = nil
	}
	return errors.Join(stopErr, descendantsErr, groupErr)
}
