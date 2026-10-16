//go:build !windows

package testrunner

import (
	"fmt"
	"os"
	"syscall"
)

// signalQuit asks a Go process for its goroutines: on SIGQUIT the runtime
// prints every goroutine's stack and exits.
func signalQuit(pid int) error {
	if err := syscall.Kill(pid, syscall.SIGQUIT); err != nil {
		return fmt.Errorf("SIGQUIT go test binary %d: %w", pid, err)
	}
	return nil
}

func signalAlive(process *os.Process) bool {
	return process.Signal(syscall.Signal(0)) == nil
}
