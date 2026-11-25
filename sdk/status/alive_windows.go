//go:build windows

package status

import "syscall"

// Alive reports whether a process with pid exists and has not exited.
func Alive(pid int) bool {
	const queryLimitedInformation, stillActive = 0x1000, 259
	handle, err := syscall.OpenProcess(queryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle) //nolint:errcheck // a read-only query handle
	var code uint32
	return syscall.GetExitCodeProcess(handle, &code) == nil && code == stillActive
}
