//go:build linux

package serve

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/flanksource/commons/logger"
)

const (
	defaultUnitName = "gavel-ssh.service"
	defaultUnitPath = "/etc/systemd/system/" + defaultUnitName
	defaultDataDir  = "/var/lib/gavel"
	defaultUser     = "gavel"
)

type InstallOptions struct {
	Port       int
	Host       string
	User       string
	UnitPath   string
	DataDir    string
	BinaryPath string
	DryRun     bool
	Force      bool
}

func (o *InstallOptions) applyDefaults() error {
	if o.Port == 0 {
		o.Port = 2222
	}
	if o.Host == "" {
		o.Host = "0.0.0.0"
	}
	if o.User == "" {
		o.User = defaultUser
	}
	if o.UnitPath == "" {
		o.UnitPath = defaultUnitPath
	}
	if o.DataDir == "" {
		o.DataDir = defaultDataDir
	}
	if o.BinaryPath == "" {
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("resolve current executable: %w", err)
		}
		abs, err := filepath.Abs(exe)
		if err != nil {
			return fmt.Errorf("resolve absolute path of %s: %w", exe, err)
		}
		o.BinaryPath = abs
	}
	return nil
}

func Install(opts InstallOptions) error {
	if err := opts.applyDefaults(); err != nil {
		return err
	}

	unit, err := renderUnit(opts)
	if err != nil {
		return err
	}

	if opts.DryRun {
		logger.Infof("[dry-run] would ensure system user %q", opts.User)
		logger.Infof("[dry-run] would ensure data dir %s (owned by %s)", opts.DataDir, opts.User)
		logger.Infof("[dry-run] would write unit file %s", opts.UnitPath)
		logger.Infof("[dry-run] would run: systemctl daemon-reload && systemctl enable --now %s", defaultUnitName)
		fmt.Println("---")
		fmt.Println(unit)
		fmt.Println("---")
		return nil
	}

	if err := ensureRoot(); err != nil {
		return err
	}
	if err := ensureSystemUser(opts.User, opts.DataDir); err != nil {
		return err
	}
	if err := ensureDataDir(opts.DataDir, opts.User); err != nil {
		return err
	}
	if err := writeUnit(opts.UnitPath, unit, opts.Force); err != nil {
		return err
	}
	if err := systemctl("daemon-reload"); err != nil {
		return err
	}
	if err := systemctl("enable", "--now", defaultUnitName); err != nil {
		return err
	}

	logger.Infof("Installed and started %s (listening on %s:%d)", defaultUnitName, opts.Host, opts.Port)
	logger.Infof("Check status with: systemctl status %s", defaultUnitName)
	return nil
}

func ensureRoot() error {
	if os.Geteuid() != 0 {
		return errors.New("gavel ssh install must be run as root (try sudo)")
	}
	return nil
}

func ensureSystemUser(name, homeDir string) error {
	if _, err := user.Lookup(name); err == nil {
		logger.V(1).Infof("system user %q already exists", name)
		return nil
	}
	logger.Infof("Creating system user %q", name)
	cmd := exec.Command("useradd",
		"--system",
		"--home-dir", homeDir,
		"--shell", "/usr/sbin/nologin",
		"--user-group",
		name,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("useradd failed: %w: %s", err, bytes.TrimSpace(out))
	}
	return nil
}

func ensureDataDir(path, owner string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("create data dir %s: %w", path, err)
	}
	u, err := user.Lookup(owner)
	if err != nil {
		return fmt.Errorf("lookup user %q after creation: %w", owner, err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("parse uid %q: %w", u.Uid, err)
	}
	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("parse gid %q: %w", u.Gid, err)
	}
	if err := os.Chown(path, uid, gid); err != nil {
		return fmt.Errorf("chown %s: %w", path, err)
	}
	return nil
}

// unitTemplate intentionally omits systemd hardening directives
// (NoNewPrivileges, ProtectSystem, ProtectHome, ReadWritePaths, PrivateTmp).
// Gavel's post-receive hook spawns go build / go test subprocesses that
// need a real /tmp and read/write access to $HOME/go/pkg/mod and
// $HOME/.cache/go-build. User-level isolation via the dedicated `gavel`
// system user is still in place.
const unitTemplate = `[Unit]
Description=Gavel SSH git-push backend
After=network.target
Wants=network.target

[Service]
Type=simple
User={{.User}}
Group={{.User}}
WorkingDirectory={{.DataDir}}
ExecStart={{.BinaryPath}} ssh serve --host {{.Host}} --port {{.Port}} --host-key {{.DataDir}}/ssh_host_key --repo-dir {{.DataDir}}/repos
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`

func renderUnit(opts InstallOptions) (string, error) {
	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return "", fmt.Errorf("parse unit template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("render unit template: %w", err)
	}
	return buf.String(), nil
}

func writeUnit(path, contents string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("unit file %s already exists (use --force to overwrite)", path)
	} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create unit dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write unit %s: %w", path, err)
	}
	logger.Infof("Wrote systemd unit to %s", path)
	return nil
}

func systemctl(args ...string) error {
	cmd := exec.Command("systemctl", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	logger.V(1).Infof("running: systemctl %v", args)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %v: %w", args, err)
	}
	return nil
}
