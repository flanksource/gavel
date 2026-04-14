package main

import "github.com/spf13/cobra"

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH git-push backend commands",
	Long: `Manage the gavel SSH git-push backend.

Subcommands:
  serve    Run the SSH server in the foreground
  install  Install and enable a systemd unit for the server (Linux only)`,
}

func init() {
	rootCmd.AddCommand(sshCmd)
}
