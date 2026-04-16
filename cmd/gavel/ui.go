package main

import "github.com/spf13/cobra"

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Standalone gavel UI servers",
	Long: `Manage standalone gavel UI servers.

Subcommands:
  serve    Run a test UI server that replays a previously-captured run and
           auto-stops after an idle or hard deadline. Primarily used as the
           detached child process forked by ` + "`gavel test --ui --detach`" + `,
           but also useful for manually replaying a JSON snapshot.`,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
