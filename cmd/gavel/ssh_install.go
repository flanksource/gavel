package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/serve"
)

type SSHInstallOptions struct {
	Port       int    `flag:"port" help:"SSH server port" default:"2222"`
	Host       string `flag:"host" help:"Listen address" default:"0.0.0.0"`
	User       string `flag:"user" help:"System user to run the service as" default:"gavel"`
	UnitPath   string `flag:"unit-path" help:"Path to write the systemd unit" default:"/etc/systemd/system/gavel-ssh.service"`
	DataDir    string `flag:"data-dir" help:"Directory for host key and cached repos" default:"/var/lib/gavel"`
	BinaryPath string `flag:"binary" help:"Path to the gavel binary (defaults to the current executable)"`
	DryRun     bool   `flag:"dry-run" help:"Print actions and rendered unit without writing anything"`
	Force      bool   `flag:"force" help:"Overwrite an existing unit file"`
}

func (o SSHInstallOptions) Help() string {
	return `Install and enable a systemd unit that runs the gavel SSH git-push backend.

Creates a dedicated system user, writes /etc/systemd/system/gavel-ssh.service,
runs systemctl daemon-reload, and enables the service. Requires root.

Linux only. Use --dry-run to preview without making changes.`
}

func init() {
	clicky.AddNamedCommand("install", sshCmd, SSHInstallOptions{}, runSSHInstall)
}

func runSSHInstall(opts SSHInstallOptions) (any, error) {
	return nil, serve.Install(serve.InstallOptions{
		Port:       opts.Port,
		Host:       opts.Host,
		User:       opts.User,
		UnitPath:   opts.UnitPath,
		DataDir:    opts.DataDir,
		BinaryPath: opts.BinaryPath,
		DryRun:     opts.DryRun,
		Force:      opts.Force,
	})
}
