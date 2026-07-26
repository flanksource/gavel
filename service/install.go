package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// InstallOptions configures the user-level service-file installer.
type InstallOptions struct {
	// BinaryPath overrides os.Executable() when writing the service file.
	BinaryPath string
	// DryRun renders the service file to stdout without touching disk.
	DryRun bool
	// Force overwrites an existing service file.
	Force bool
}

var daemonArguments = []string{
	"pr",
	"list",
	"--all",
	"--ui",
	"--menu-bar",
	"--port=0",
	"--persist-port",
}

func userShellInvocation(shell, binary string) []string {
	return append([]string{shell, "-l", "-i", "-c", `exec "$@"`, "gavel-system", binary}, daemonArguments...)
}

func userShellPath() (string, error) {
	shell := strings.TrimSpace(os.Getenv("SHELL"))
	if shell == "" {
		return "", fmt.Errorf("resolve user shell: SHELL is not set")
	}
	resolved, err := exec.LookPath(shell)
	if err != nil {
		return "", fmt.Errorf("resolve user shell %q: %w", shell, err)
	}
	absolute, err := filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path of user shell %s: %w", resolved, err)
	}
	return absolute, nil
}

// Install writes the platform-specific user-level service file (launchd plist
// on macOS, systemd --user unit on Linux), loads it, and starts it. Exposing
// Install/Uninstall through package-level platform-dispatched symbols keeps
// the cobra wiring identical on both platforms — it just calls service.Install.
//
// On unsupported platforms the function returns an error; see the
// service_other.go build-tagged file.

// Uninstall stops and removes the service file installed by Install.
