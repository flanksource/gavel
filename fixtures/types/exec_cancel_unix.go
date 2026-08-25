//go:build unix

package types

import (
	"os"
	"syscall"
)

func cancelPTYProcess(process *os.Process) error {
	if process == nil {
		return os.ErrProcessDone
	}
	return syscall.Kill(-process.Pid, syscall.SIGKILL)
}
