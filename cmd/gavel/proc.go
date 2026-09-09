package main

import "github.com/spf13/cobra"

var procCmd = &cobra.Command{
	Use:     "proc",
	Aliases: []string{"procfile"},
	Short:   "Run and supervise the processes declared in a Procfile",
	Long: `Run the processes declared in a Heroku/foreman-style Procfile
(one "name: command" per line).

Subcommands:
  run        Run all processes in the foreground, multiplexing their output
  start      Start the processes as a detached background daemon
  stop       Stop the daemon, or named processes
  restart    Restart the daemon, or named processes
  status     Show the status of each process
  list       List the configured processes and their commands
  logs       Tail (and optionally follow) per-process logs

Restart behaviour and per-process environment are configured under the
` + "`procfile`" + ` key in .gavel.yaml.`,
}

func init() {
	rootCmd.AddCommand(procCmd)
}
