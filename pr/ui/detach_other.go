//go:build !unix

package ui

import "syscall"

// detachedSysProcAttr has no session to detach into on non-unix platforms;
// the launched process simply runs unattached to the request.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
