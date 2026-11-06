// `gavel runner run-task`, gavel's program in a substrate environment: the environment's side of a
// whole run, with gavel's hooks rebuilt from the run's host data and the run API served to it.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/promptrun/remoterun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox/runclient"
	"github.com/flanksource/gavel/todos/lifecycle"
)

type runTaskOptions struct {
	// Dir is the run's directory, holding its task and workspace bundle.
	Dir string
	// Agent is the provider the step's turns go to; nil builds it from the job's model.
	Agent captainai.Provider
}

// runAPISocket is the socket, in the run's directory, the run API is served to the run on.
const runAPISocket = "run-api.sock"

// runnerRunTask runs the task in opts.Dir with out and in as its channel to the caller: the
// step's run, with the hooks a local step runs, rebuilt from the run's host data, and the tools
// that run beside them. The run reaches the caller's run API over a host call: the step's PR this
// process reaches itself, and what the run starts, such as a job's `gavel runner exec`, reaches
// it through runAPISocket, which the run's environment names.
func runnerRunTask(ctx context.Context, opts runTaskOptions, out io.Writer, in io.Reader) (err error) {
	// The step's PR, if it has one, is reached through the run's caller, which has the credentials
	// the environment never holds.
	defer provider.UseRunAPI(nil)
	stopServing := func() error { return nil }
	defer func() { err = errors.Join(err, stopServing()) }()
	return substraterun.RunTask(ctx, opts.Dir, out, in, func(job remoterun.Job, src string, call remoterun.HostCall) (remoterun.Host, error) {
		provider.UseRunAPI(runclient.OverHostCall(call))
		dir, err := filepath.Abs(opts.Dir)
		if err != nil {
			return remoterun.Host{}, err
		}
		socket := filepath.Join(dir, runAPISocket)
		stop, err := runclient.ServeOverHostCall(socket, call)
		if err != nil {
			return remoterun.Host{}, err
		}
		stopServing = func() error { return errors.Join(stop(), os.Unsetenv(runclient.SocketEnv)) }
		if err := os.Setenv(runclient.SocketEnv, socket); err != nil {
			return remoterun.Host{}, fmt.Errorf("name the run API's socket for the run: %w", err)
		}
		var hooks lifecycle.RunHooks
		if len(job.Host) > 0 {
			if err := json.Unmarshal(job.Host, &hooks); err != nil {
				return remoterun.Host{}, fmt.Errorf("decode gavel's hooks from the run's host data: %w", err)
			}
		}
		built, tools, err := hooks.In(job.Root, src).Build()
		if err != nil {
			return remoterun.Host{}, err
		}
		return remoterun.Host{Provider: opts.Agent, Hooks: built, Tools: tools}, nil
	})
}
