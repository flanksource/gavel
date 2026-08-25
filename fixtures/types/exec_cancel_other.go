//go:build !unix

package types

import "os"

func cancelPTYProcess(process *os.Process) error {
	if process == nil {
		return os.ErrProcessDone
	}
	return process.Kill()
}
