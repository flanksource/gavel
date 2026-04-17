//go:build !unix

package service

import "syscall"

// detachedSysProcAttr is a stub for non-unix platforms. The package targets
// darwin + linux primarily; this placeholder lets Windows builds compile.
func detachedSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
