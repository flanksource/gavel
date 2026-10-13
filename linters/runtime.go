package linters

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/verify"
	"path/filepath"
	"time"
)

func (o RunOptions) resolveRuntime(name string) (RunOptions, error) {
	if o.runtimeResolved {
		return o, nil
	}
	workDir := o.WorkDir
	if workDir == "" {
		workDir = "."
	}
	if o.ConfigDir != "" {
		executionDir, err := filepath.Abs(workDir)
		if err != nil {
			return o, err
		}
		invocationDir, err := filepath.Abs(o.ConfigDir)
		if err != nil {
			return o, err
		}
		workDir = verify.RuntimeConfigDirectory(executionDir, invocationDir)
	}
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return o, err
	}
	caller := o.Runtime
	if o.Timeout > 0 {
		value := o.Timeout.String()
		caller.Timeout = &value
	}
	runtime, args, err := config.RuntimeFor("lint", name, caller)
	if err != nil {
		return o, err
	}
	o.Runtime = runtime
	if runtime.Timeout != nil && *runtime.Timeout != "" {
		o.Timeout, err = time.ParseDuration(*runtime.Timeout)
		if err != nil {
			return o, err
		}
	}
	o.ExtraArgs = append(args, o.ExtraArgs...)
	o.runtimeResolved = true
	return o, nil
}

func (o RunOptions) CommandContext(ctx context.Context, name, binary string, args ...string) *processruntime.Command {
	resolved, err := o.resolveRuntime(name)
	if err != nil {
		return &processruntime.Command{ConfigError: fmt.Errorf("%s runtime: %w", name, err)}
	}
	if !o.runtimeResolved {
		args = append(args, resolved.ExtraArgs[:len(resolved.ExtraArgs)-len(o.ExtraArgs)]...)
	}
	if o.Executable != "" && name != "tsc" {
		binary = o.Executable
	}
	command := processruntime.NewCommand(ctx, resolved.Runtime, binary, args...)
	command.Dir = o.WorkDir
	return command
}
