//go:build !darwin && !linux

package service

import "errors"

// Install returns an error on unsupported platforms. macOS uses launchd and
// Linux uses systemd user units; no other platform is wired up.
func Install(_ InstallOptions) error {
	return errors.New("gavel system install is only supported on macOS (launchd) and Linux (systemd --user)")
}

// Uninstall is the symmetric stub for unsupported platforms.
func Uninstall() error {
	return errors.New("gavel system uninstall is only supported on macOS (launchd) and Linux (systemd --user)")
}

// IsInstalled always returns false on unsupported platforms so the Start/
// Stop/ReadStatus dispatch falls through to the pidfile-based path.
func IsInstalled() (bool, error) { return false, nil }

func serviceStart() error {
	return errors.New("service management is only supported on macOS (launchd) and Linux (systemd --user)")
}

func serviceStop() error {
	return errors.New("service management is only supported on macOS (launchd) and Linux (systemd --user)")
}

func serviceStatus() (Status, error) {
	return Status{}, errors.New("service management is only supported on macOS (launchd) and Linux (systemd --user)")
}
