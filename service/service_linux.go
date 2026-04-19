//go:build linux

package service

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/flanksource/commons/logger"
)

// unitPath returns ~/.config/systemd/user/<ServiceName>.service — a user-level
// systemd unit so install doesn't require root. Users need linger enabled
// (`loginctl enable-linger`) to keep it running after logout; noted in the
// install output.
func unitPath() (string, error) {
	return userUnitPathFor(ServiceName)
}

// legacyUnitPath returns the unit path for the pre-rename service name,
// used during upgrade cleanup.
func legacyUnitPath() (string, error) {
	return userUnitPathFor(legacyServiceName)
}

func userUnitPathFor(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "systemd", "user", name+".service"), nil
}

// cleanupLegacyService disables and removes the pre-rename systemd unit if
// present. Swallows errors — an upgraded machine with a half-removed legacy
// unit is still better off than one where install hard-fails.
func cleanupLegacyService() {
	path, err := legacyUnitPath()
	if err != nil {
		return
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return
	}
	logger.Infof("Removing legacy systemd unit %s", path)
	if err := systemctlUser("disable", "--now", legacyServiceName+".service"); err != nil {
		logger.Warnf("systemctl --user disable legacy unit: %v", err)
	}
	if err := os.Remove(path); err != nil {
		logger.Warnf("remove legacy unit %s: %v", path, err)
	}
	_ = systemctlUser("daemon-reload")
}

type unitData struct {
	BinaryPath string
}

const userUnitTemplate = `[Unit]
Description=Gavel PR UI (pr list --all --ui)
After=default.target

[Service]
Type=simple
ExecStart={{.BinaryPath}} pr list --all --ui --menu-bar --port=0 --persist-port
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`

func renderUserUnit(bin string) (string, error) {
	t, err := template.New("unit").Parse(userUnitTemplate)
	if err != nil {
		return "", fmt.Errorf("parse unit template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, unitData{BinaryPath: bin}); err != nil {
		return "", fmt.Errorf("render unit template: %w", err)
	}
	return buf.String(), nil
}

// Install writes the user-level systemd unit, reloads the user daemon, and
// enables + starts the service.
func Install(opts InstallOptions) error {
	bin := opts.BinaryPath
	if bin == "" {
		b, err := BinaryPath()
		if err != nil {
			return err
		}
		bin = b
	}
	unit, err := renderUserUnit(bin)
	if err != nil {
		return err
	}
	path, err := unitPath()
	if err != nil {
		return err
	}

	if opts.DryRun {
		logger.Infof("[dry-run] would remove legacy unit %s.service if present", legacyServiceName)
		logger.Infof("[dry-run] would write user unit %s", path)
		logger.Infof("[dry-run] would run: systemctl --user daemon-reload && systemctl --user enable --now %s.service", ServiceName)
		fmt.Println("---")
		fmt.Println(unit)
		fmt.Println("---")
		return nil
	}

	// Clean up the pre-rename unit before writing the new one — otherwise
	// an upgraded machine ends up with two units both trying to bind
	// localhost:9092.
	cleanupLegacyService()

	if _, err := os.Stat(path); err == nil && !opts.Force {
		return fmt.Errorf("unit %s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create unit dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(unit), 0o644); err != nil {
		return fmt.Errorf("write unit %s: %w", path, err)
	}
	logger.Infof("Wrote systemd user unit to %s", path)

	if err := systemctlUser("daemon-reload"); err != nil {
		return err
	}
	if err := systemctlUser("enable", "--now", ServiceName+".service"); err != nil {
		return err
	}

	logger.Infof("Enabled and started %s.service (user scope)", ServiceName)
	logger.Infof("To keep it running after logout, enable linger: loginctl enable-linger $USER")
	logger.Infof("Check status with: systemctl --user status %s.service", ServiceName)
	return nil
}

// Uninstall stops + disables the unit and removes the unit file. Also
// cleans up the legacy (pre-rename) unit so `gavel system uninstall` is a
// reliable "remove everything" button after upgrades.
func Uninstall() error {
	cleanupLegacyService()

	path, err := unitPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		logger.Infof("No unit at %s; nothing to uninstall", path)
		return nil
	}
	if err := systemctlUser("disable", "--now", ServiceName+".service"); err != nil {
		logger.Warnf("systemctl --user disable: %v", err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	if err := systemctlUser("daemon-reload"); err != nil {
		logger.Warnf("systemctl --user daemon-reload: %v", err)
	}
	logger.Infof("Removed systemd user unit %s", path)
	return nil
}

func systemctlUser(args ...string) error {
	full := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", full...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	logger.V(1).Infof("running: systemctl %v", full)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %v: %w", full, err)
	}
	return nil
}

// systemctlUserOutput runs systemctl --user and returns captured stdout. Used
// by the status parser which needs to inspect `systemctl show` output.
func systemctlUserOutput(args ...string) (string, error) {
	full := append([]string{"--user"}, args...)
	cmd := exec.Command("systemctl", full...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	logger.V(1).Infof("running: systemctl %v", full)
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("systemctl %v: %w", full, err)
	}
	return stdout.String(), nil
}

// IsInstalled reports whether the user-level systemd unit file exists on disk.
// A unit file on disk is sufficient evidence that the service dispatch should
// be used; whether the unit is currently enabled/active is a separate query.
func IsInstalled() (bool, error) {
	path, err := unitPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", path, err)
}

// serviceStart asks systemd to (re)start the unit. `restart` stops any running
// copy first, matching Start()'s "tear down then spawn" contract.
func serviceStart() error {
	return systemctlUser("restart", ServiceName+".service")
}

// serviceStop stops the unit without disabling it (uninstall is what removes
// the unit file and disables autostart).
func serviceStop() error {
	return systemctlUser("stop", ServiceName+".service")
}

// serviceStatus queries systemd for the unit's state. `systemctl show` with
// explicit properties emits deterministic KEY=VALUE lines, so we don't need
// to parse the verbose `status` output.
func serviceStatus() (Status, error) {
	st := Status{}
	pf, err := PidFile()
	if err != nil {
		return st, err
	}
	lf, err := LogFile()
	if err != nil {
		return st, err
	}
	st.PidFile = pf
	st.LogFile = lf

	out, err := systemctlUserOutput("show", ServiceName+".service", "--property=ActiveState,MainPID")
	if err != nil {
		logger.V(1).Infof("systemctl show %s: %v", ServiceName, err)
		return st, nil
	}
	running, pid := parseSystemctlShow(out)
	st.Running = running
	st.PID = pid
	return st, nil
}

// parseSystemctlShow extracts (running, pid) from `systemctl show
// --property=ActiveState,MainPID` output. ActiveState=active with a non-zero
// MainPID is the only state we call "running" — `activating` and
// `deactivating` are treated as not-running because the pid may be 0 or
// about to change. Exposed unexported for testing.
func parseSystemctlShow(out string) (bool, int) {
	var activeState string
	var mainPID int
	for line := range strings.SplitSeq(out, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "ActiveState":
			activeState = strings.TrimSpace(v)
		case "MainPID":
			if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				mainPID = n
			}
		}
	}
	return activeState == "active" && mainPID > 0, mainPID
}
