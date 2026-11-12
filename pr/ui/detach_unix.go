//go:build unix

package ui

import "syscall"

// detachedSysProcAttr starts a launched process in its own session, so a
// signal to the gavel server's process group never reaches it.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
