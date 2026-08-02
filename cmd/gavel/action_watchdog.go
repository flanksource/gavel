package main

import (
	"fmt"
	"io"
	"os"
	"time"

	clickyexec "github.com/flanksource/clicky/exec"
	"github.com/spf13/cobra"
)

const actionTimeoutExitCode = 124

type actionWatchdogOptions struct {
	Executable   string
	Args         []string
	Timeout      time.Duration
	StackGrace   time.Duration
	Stdout       io.Writer
	Stderr       io.Writer
	RequestStack func(int) error
}

func runActionWatchdog(options actionWatchdogOptions) int {
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}
	if options.StackGrace <= 0 {
		options.StackGrace = 5 * time.Second
	}
	if options.RequestStack == nil {
		options.RequestStack = requestProcessStack
	}
	if options.Executable == "" || options.Timeout <= 0 {
		fmt.Fprintln(options.Stderr, "action watchdog requires an executable and a positive timeout")
		return 1
	}

	process := clickyexec.NewExec(options.Executable, options.Args...).WithProcessGroup().Stream(options.Stdout, options.Stderr)
	finished := make(chan *clickyexec.ExecResult, 1)
	go func() {
		finished <- process.Run().Result()
	}()

	timer := time.NewTimer(options.Timeout)
	defer timer.Stop()
	select {
	case result := <-finished:
		return actionResultExitCode(result)
	case <-timer.C:
		fmt.Fprintf(options.Stderr, "gavel invocation exceeded action timeout %s; requesting stack trace before terminating\n", options.Timeout)
	}

	if pid := process.Pid(); pid > 0 {
		if err := options.RequestStack(pid); err != nil {
			fmt.Fprintf(options.Stderr, "request stack trace from pid %d: %v\n", pid, err)
		}
	}

	grace := time.NewTimer(options.StackGrace)
	select {
	case <-finished:
		grace.Stop()
	case <-grace.C:
	}
	if err := process.KillTree(); err != nil {
		fmt.Fprintf(options.Stderr, "kill timed-out gavel process tree: %v\n", err)
	}
	return actionTimeoutExitCode
}

func actionResultExitCode(result *clickyexec.ExecResult) int {
	if result == nil {
		return 1
	}
	return result.ExitCode
}

func newActionWatchdogCommand() *cobra.Command {
	var timeout time.Duration
	var stackGrace time.Duration
	cmd := &cobra.Command{
		Use:    "action-watchdog --timeout duration -- <gavel args...>",
		Short:  "Run gavel under the GitHub Action wall-clock watchdog",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			executable, err := os.Executable()
			if err != nil {
				return fmt.Errorf("resolve gavel executable: %w", err)
			}
			if timeout <= 0 {
				return fmt.Errorf("--timeout must be positive")
			}
			exitCode = runActionWatchdog(actionWatchdogOptions{
				Executable: executable,
				Args:       args,
				Timeout:    timeout,
				StackGrace: stackGrace,
			})
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "timeout", 20*time.Minute, "hard wall-clock limit for the gavel invocation")
	cmd.Flags().DurationVar(&stackGrace, "stack-grace", 5*time.Second, "time to allow stack output before killing the process tree")
	return cmd
}

func init() {
	rootCmd.AddCommand(newActionWatchdogCommand())
}
