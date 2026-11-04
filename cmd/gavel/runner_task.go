// `gavel runner run-task`, gavel's program in a substrate environment: the environment's side of a
// whole step run, with gavel's hooks rebuilt from the run's host data.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/promptrun/remoterun"
	"github.com/flanksource/captain/pkg/sandbox/substraterun"
	"github.com/flanksource/gavel/todos/lifecycle"
)

type runTaskOptions struct {
	// Dir is the run's directory, holding its task and workspace bundle.
	Dir string
	// Agent is the provider the step's turns go to; nil builds it from the job's model.
	Agent captainai.Provider
}

// runnerRunTask runs the task in opts.Dir with out and in as its channel to the caller: the
// step's run, with the hooks a local step runs, rebuilt from the run's host data, and the tools
// that run beside them.
func runnerRunTask(ctx context.Context, opts runTaskOptions, out io.Writer, in io.Reader) error {
	return substraterun.RunTask(ctx, opts.Dir, out, in, func(job remoterun.Job, src string) (remoterun.Host, error) {
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
