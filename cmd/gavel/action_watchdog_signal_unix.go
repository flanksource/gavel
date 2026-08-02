//go:build !windows

package main

import (
	"fmt"
	"os"
	"syscall"
)

func requestProcessStack(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := process.Signal(syscall.SIGQUIT); err != nil {
		return fmt.Errorf("signal process %d: %w", pid, err)
	}
	return nil
}
