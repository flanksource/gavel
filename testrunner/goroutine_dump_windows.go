//go:build windows

package testrunner

import (
	"fmt"
	"os"
)

func signalQuit(pid int) error {
	return fmt.Errorf("a goroutine dump (SIGQUIT) of go test binary %d is unsupported on Windows", pid)
}

// signalAlive is never waited on: no binary was signalled on Windows.
func signalAlive(*os.Process) bool { return false }
