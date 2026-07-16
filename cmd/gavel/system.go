package main

import "github.com/spf13/cobra"

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "System integration commands (background daemon, service install)",
	Long: `Manage the background gavel PR UI daemon and the launchd / systemd
service files that keep it running.

Subcommands:
  start      Start gavel pr list --all --ui --menu-bar detached in the background
  stop       Stop the running background instance
  status     Report whether the background instance is running
  install    Install a user-level launchd (macOS) or systemd (Linux) service
  uninstall  Remove the installed user-level service`,
}

func init() {
	rootCmd.AddCommand(systemCmd)
}
