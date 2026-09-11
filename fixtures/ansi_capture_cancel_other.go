//go:build !unix

package fixtures

import (
	"errors"
	"os"
)

func cancelCaptureProcess(process *os.Process) error {
	if process == nil {
		return os.ErrProcessDone
	}
	return errors.Join(killCaptureDescendants(process.Pid), process.Kill())
}
