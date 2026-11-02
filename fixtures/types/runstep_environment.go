package types

import (
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/processruntime"
)

func fixtureStepRuntime(opts fixtures.RunOptions) (processruntime.Options, error) {
	if opts.Scope == nil {
		return opts.ProcessRuntime, nil
	}
	env, err := opts.ChildEnv()
	if err != nil {
		return processruntime.Options{}, err
	}
	return opts.ProcessRuntime.Merge(processruntime.Options{Env: env}), nil
}
