package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/service"
)

type SystemUninstallOptions struct{}

func (SystemUninstallOptions) Help() api.Textable {
	return clicky.Text("Remove the user-level launchd (macOS) or systemd (Linux) service installed by `gavel system install`.")
}

func init() {
	clicky.AddNamedCommand("uninstall", systemCmd, SystemUninstallOptions{}, runSystemUninstall)
}

func runSystemUninstall(_ SystemUninstallOptions) (any, error) {
	return nil, service.Uninstall()
}
