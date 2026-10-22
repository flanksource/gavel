package testrunner

import (
	"fmt"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/verify"
	"time"
)

func (o *TestOrchestrator) buildCommand(runner runners.Runner, pkg string, extraArgs []string) (*runners.TestRun, error) {
	runtime, args, err := o.runtimeFor(runner.Name())
	if err != nil {
		return nil, err
	}
	timeout := 10 * time.Minute
	if runtime.Timeout != nil {
		if *runtime.Timeout == "" {
			timeout = 0
		} else {
			timeout, err = time.ParseDuration(*runtime.Timeout)
			if err != nil {
				return nil, err
			}
		}
	}
	args = append(subprocessTimeoutArgs(runner.Name(), subprocessTimeoutFor(timeout)), args...)
	args = append(args, extraArgs...)
	run, err := runner.BuildCommand(pkg, args...)
	if err != nil {
		return nil, err
	}
	for key, value := range run.Process.Env {
		if configured, exists := runtime.Env[key]; exists && configured != value {
			return nil, fmt.Errorf("runtime env %s conflicts with %s report output", key, runner.Name())
		}
	}
	if err := runtime.Apply(run.Process); err != nil {
		return nil, err
	}
	run.Process.WithTimeout(timeout)
	return run, nil
}

func (o *TestOrchestrator) runtimeFor(framework parsers.Framework) (processruntime.Options, []string, error) {
	workDir := o.runtimeWorkDir()
	if workDir == "" {
		workDir = "."
	}
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return processruntime.Options{}, nil, err
	}
	caller := o.Runtime
	if o.TestTimeout > 0 {
		timeout := o.TestTimeout.String()
		caller.Timeout = &timeout
	}
	return config.RuntimeFor("test", string(framework), caller)
}

func (o RunOptions) fixtureRuntime() processruntime.Options {
	runtime := o.Runtime
	if o.TestTimeout > 0 {
		timeout := o.TestTimeout.String()
		runtime.Timeout = &timeout
	}
	return runtime
}

func (o RunOptions) runtimeWorkDir() string {
	return verify.RuntimeConfigDirectory(o.WorkDir, o.runtimeConfigDir)
}
