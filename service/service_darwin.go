//go:build darwin

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

// launchdLabel is the reverse-DNS identifier used for the LaunchAgent. It
// must match the plist Label key and the filename stem (minus .plist).
const launchdLabel = "com.flanksource." + ServiceName

// legacyLaunchdLabel is the pre-rename identifier. Install/Uninstall clean
// up plists written under this name so upgraders don't end up with two
// LaunchAgents running the same binary side by side.
const legacyLaunchdLabel = "com.flanksource." + legacyServiceName

// plistPath returns ~/Library/LaunchAgents/<label>.plist — the per-user
// LaunchAgents directory, which doesn't require root and loads at login.
func plistPath() (string, error) {
	return launchAgentPath(launchdLabel)
}

// legacyPlistPath returns the plist path for the pre-rename service name,
// used during upgrade cleanup.
func legacyPlistPath() (string, error) {
	return launchAgentPath(legacyLaunchdLabel)
}

func launchAgentPath(label string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

// cleanupLegacyService removes the pre-rename LaunchAgent if present.
// Safe to call when nothing legacy exists — returns nil in that case.
// Failures are logged but not returned: a half-removed legacy agent is
// better than a failed install.
func cleanupLegacyService() {
	path, err := legacyPlistPath()
	if err != nil {
		return
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return
	}
	logger.Infof("Removing legacy LaunchAgent %s", path)
	// bootout is more reliable than unload on modern macOS but either way
	// we swallow errors — the file removal below is what actually matters.
	_ = launchctl("unload", path)
	if err := os.Remove(path); err != nil {
		logger.Warnf("remove legacy plist %s: %v", path, err)
	}
}

type plistData struct {
	Label      string
	BinaryPath string
	LogFile    string
}

// launchdTemplate is kept minimal on purpose. KeepAlive w/ SuccessfulExit=false
// restarts the service only when it crashes, not when it exits cleanly from a
// subsequent `system stop`. RunAtLoad makes it start on launchctl load and at
// login.
const launchdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>{{.Label}}</string>
  <key>ProgramArguments</key>
  <array>
    <string>{{.BinaryPath}}</string>
    <string>pr</string>
    <string>list</string>
    <string>--all</string>
    <string>--ui</string>
    <string>--menu-bar</string>
    <string>--port=0</string>
    <string>--persist-port</string>
  </array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key><false/>
  </dict>
  <key>StandardOutPath</key><string>{{.LogFile}}</string>
  <key>StandardErrorPath</key><string>{{.LogFile}}</string>
  <key>ProcessType</key><string>Background</string>
</dict>
</plist>
`

func renderPlist(bin, logPath string) (string, error) {
	t, err := template.New("plist").Parse(launchdTemplate)
	if err != nil {
		return "", fmt.Errorf("parse plist template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, plistData{Label: launchdLabel, BinaryPath: bin, LogFile: logPath}); err != nil {
		return "", fmt.Errorf("render plist template: %w", err)
	}
	return buf.String(), nil
}

// Install writes the LaunchAgent plist and loads it via launchctl.
func Install(opts InstallOptions) error {
	bin := opts.BinaryPath
	if bin == "" {
		b, err := BinaryPath()
		if err != nil {
			return err
		}
		bin = b
	}
	logPath, err := LogFile()
	if err != nil {
		return err
	}
	plist, err := renderPlist(bin, logPath)
	if err != nil {
		return err
	}
	path, err := plistPath()
	if err != nil {
		return err
	}

	if opts.DryRun {
		logger.Infof("[dry-run] would remove legacy LaunchAgent %s if present", legacyLaunchdLabel)
		logger.Infof("[dry-run] would write LaunchAgent %s", path)
		logger.Infof("[dry-run] would run: launchctl unload %s (ignored on first install)", path)
		logger.Infof("[dry-run] would run: launchctl load -w %s", path)
		fmt.Println("---")
		fmt.Println(plist)
		fmt.Println("---")
		return nil
	}

	// Clean up the pre-rename LaunchAgent before writing the new one —
	// otherwise an upgraded machine ends up with two competing agents both
	// trying to bind localhost:9092.
	cleanupLegacyService()

	if _, err := os.Stat(path); err == nil && !opts.Force {
		return fmt.Errorf("LaunchAgent %s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}

	// If a previous plist exists, unload it first so the new one takes effect.
	_ = launchctl("unload", path)

	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return fmt.Errorf("write plist %s: %w", path, err)
	}
	logger.Infof("Wrote LaunchAgent to %s", path)

	if err := launchctl("load", "-w", path); err != nil {
		return err
	}
	logger.Infof("Loaded LaunchAgent %s; it will start now and at each login", launchdLabel)
	logger.Infof("Check status with: launchctl list | grep %s", launchdLabel)
	return nil
}

// Uninstall stops the LaunchAgent and removes the plist file. Also cleans
// up the legacy (pre-rename) plist so `gavel system uninstall` is a reliable
// "remove everything" button after upgrades.
func Uninstall() error {
	cleanupLegacyService()

	path, err := plistPath()
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		logger.Infof("No LaunchAgent at %s; nothing to uninstall", path)
		return nil
	}
	if err := launchctl("unload", path); err != nil {
		logger.Warnf("launchctl unload %s: %v", path, err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	logger.Infof("Removed LaunchAgent %s", path)
	return nil
}

func launchctl(args ...string) error {
	cmd := exec.Command("launchctl", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	logger.V(1).Infof("running: launchctl %v", args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launchctl %v: %w", args, err)
	}
	return nil
}

// launchctlOutput runs launchctl and returns captured stdout. Used by the
// status parser which needs to inspect `launchctl print` output.
func launchctlOutput(args ...string) (string, error) {
	cmd := exec.Command("launchctl", args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout
	logger.V(1).Infof("running: launchctl %v", args)
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("launchctl %v: %w", args, err)
	}
	return stdout.String(), nil
}

// serviceTarget is the launchctl domain target for the user-level agent. The
// gui/<uid>/<label> form is the modern launchctl addressing scheme — the older
// "<label>" alone works with list/load/unload but not with kickstart/bootout.
func serviceTarget() string {
	return fmt.Sprintf("gui/%d/%s", os.Getuid(), launchdLabel)
}

// IsInstalled reports whether the user-level LaunchAgent plist exists on disk.
// It does not check whether launchd has the agent currently loaded — a plist
// on disk is sufficient evidence that the service dispatch should be used.
func IsInstalled() (bool, error) {
	path, err := plistPath()
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

// serviceStart asks launchd to (re)start the agent. `kickstart -k` terminates
// any running copy first, matching Start()'s "tear down then spawn" contract.
// If the agent isn't bootstrapped yet we fall back to `launchctl load -w`.
func serviceStart() error {
	if err := launchctl("kickstart", "-k", serviceTarget()); err != nil {
		path, perr := plistPath()
		if perr != nil {
			return err
		}
		// Agent wasn't bootstrapped — load the plist, which also starts it
		// thanks to RunAtLoad=true in the template.
		if loadErr := launchctl("load", "-w", path); loadErr != nil {
			return fmt.Errorf("kickstart and load both failed: %v; %w", err, loadErr)
		}
	}
	return nil
}

// serviceStop removes the agent from launchd. `bootout` is the modern verb
// and stops the running process; treat a "not bootstrapped" exit as success.
func serviceStop() error {
	if err := launchctl("bootout", serviceTarget()); err != nil {
		// bootout exits non-zero when the service is not bootstrapped. We
		// can't reliably distinguish that from other failures by exit code
		// alone, so log and swallow — the effect we want (no running agent)
		// is already the state.
		logger.V(1).Infof("launchctl bootout %s: %v (treating as already-stopped)", serviceTarget(), err)
	}
	return nil
}

// serviceStatus queries launchd for the agent's current state. We parse
// `launchctl print <target>` which emits lines like:
//
//	state = running
//	pid = 12345
//
// Missing `pid = N` (or pid = 0) means launchd knows about the service but it
// isn't currently running.
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

	out, err := launchctlOutput("print", serviceTarget())
	if err != nil {
		// print exits non-zero when the service isn't bootstrapped; that's a
		// valid "not running" answer, not an error to propagate.
		logger.V(1).Infof("launchctl print %s: %v", serviceTarget(), err)
		return st, nil
	}
	running, pid := parseLaunchctlPrint(out)
	st.Running = running
	st.PID = pid
	return st, nil
}

// parseLaunchctlPrint extracts (running, pid) from `launchctl print` output.
// Running is true when the `state = running` line is present AND a non-zero
// pid is set. Exposed unexported for testing.
func parseLaunchctlPrint(out string) (bool, int) {
	var state string
	var pid int
	for line := range strings.SplitSeq(out, "\n") {
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		val := strings.TrimSpace(v)
		switch key {
		case "state":
			state = val
		case "pid":
			if n, err := strconv.Atoi(val); err == nil {
				pid = n
			}
		}
	}
	return state == "running" && pid > 0, pid
}
