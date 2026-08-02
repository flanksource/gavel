//go:build windows

package main

import "fmt"

func requestProcessStack(pid int) error {
	return fmt.Errorf("stack signal is unsupported for process %d on Windows", pid)
}
