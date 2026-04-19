package main

import (
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/service"
)

type SystemStopOptions struct{}

func (SystemStopOptions) Help() string {
	return `Stop the detached gavel pr UI daemon if running.

Sends SIGTERM to the pid in ~/.config/gavel/pr-ui.pid and waits up to 5s for
graceful exit before escalating to SIGKILL. Does nothing if no daemon is
recorded or the pid is already dead (stale pidfile is cleaned up).`
}

func init() {
	clicky.AddNamedCommand("stop", systemCmd, SystemStopOptions{}, runSystemStop)
}

func runSystemStop(_ SystemStopOptions) (any, error) {
	before, err := service.ReadStatus()
	if err != nil {
		return nil, err
	}
	if err := service.Stop(5 * time.Second); err != nil {
		return nil, err
	}
	if before.Running {
		fmt.Printf("Stopped gavel pr UI (pid %d)\n", before.PID)
		return nil, nil
	}
	fmt.Println("No running gavel pr UI daemon")
	return nil, nil
}
