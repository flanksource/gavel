//go:build unix

package service

import "syscall"

// detachedSysProcAttr detaches the spawned process from the current
// controlling terminal and session so a SIGHUP from the parent's terminal
// close won't reach it.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
