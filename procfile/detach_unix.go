//go:build unix

package procfile

import "syscall"

// detachedSysProcAttr detaches the spawned supervisor from the current
// controlling terminal and session (Setsid) so a SIGHUP from the parent shell
// closing won't reach it — the background daemon outlives the CLI invocation.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

// terminateGroup sends SIGTERM to the whole process group led by pid. clicky's
// WithProcessGroup() puts each child in its own group whose id equals the
// leader's pid, so signalling -pid reaches the child and all its descendants,
// giving servers a chance to shut down gracefully before KillTree escalates to
// SIGKILL.
func terminateGroup(pid int) error {
	if pid <= 0 {
		return nil
	}
	return syscall.Kill(-pid, syscall.SIGTERM)
}
