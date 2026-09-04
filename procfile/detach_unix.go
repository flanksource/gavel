//go:build unix

package procfile

import "syscall"

// detachedSysProcAttr detaches the spawned supervisor from the current
// controlling terminal and session (Setsid) so a SIGHUP from the parent shell
// closing won't reach it — the background daemon outlives the CLI invocation.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
