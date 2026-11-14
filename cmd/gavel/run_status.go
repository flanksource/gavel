package main

import (
	"context"
	"os"
	"strings"

	"github.com/flanksource/gavel/runstatus"
	"github.com/flanksource/gavel/workspace"
)

// startRunStatus publishes the run's progress and slots, which `gavel slots`
// and telltale top read, until the reporter stops.
func startRunStatus(dirs workspace.Dirs) (*runstatus.Reporter, error) {
	return runstatus.Start(context.Background(), runstatus.Options{Cache: dirs.User.Cache, WorkDir: dirs.Invocation, Command: strings.Join(os.Args, " ")})
}
