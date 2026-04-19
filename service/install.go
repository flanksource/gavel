package service

// InstallOptions configures the user-level service-file installer.
type InstallOptions struct {
	// BinaryPath overrides os.Executable() when writing the service file.
	BinaryPath string
	// DryRun renders the service file to stdout without touching disk.
	DryRun bool
	// Force overwrites an existing service file.
	Force bool
}

// Install writes the platform-specific user-level service file (launchd plist
// on macOS, systemd --user unit on Linux), loads it, and starts it. Exposing
// Install/Uninstall through package-level platform-dispatched symbols keeps
// the cobra wiring identical on both platforms — it just calls service.Install.
//
// On unsupported platforms the function returns an error; see the
// service_other.go build-tagged file.

// Uninstall stops and removes the service file installed by Install.
