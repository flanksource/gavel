//go:build !linux

package serve

import "errors"

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

func Install(opts InstallOptions) error {
	return errors.New("gavel ssh install is only supported on Linux (systemd)")
}
