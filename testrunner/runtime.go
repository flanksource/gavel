package testrunner

import (
	"fmt"
	"time"

	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/verify"
	"os"
	"path/filepath"
)

func (o *TestOrchestrator) buildCommand(runner runners.Runner, pkg string, extraArgs []string) (*runners.TestRun, error) {
	runtime, args, err := o.runtimeFor(runner.Name(), pkg)
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

func (o *TestOrchestrator) runtimeFor(framework parsers.Framework, pkg string) (processruntime.Options, []string, error) {
	workDir := o.runtimeWorkDir()
	if pkg != "" {
		workDir = pkg
		if !filepath.IsAbs(workDir) { workDir = filepath.Join(o.WorkDir, pkg) }
		if info, err := os.Stat(workDir); err != nil { return processruntime.Options{}, nil, err } else if !info.IsDir() { workDir = filepath.Dir(workDir) }
		workDir = verify.RuntimeConfigDirectory(workDir, o.runtimeConfigDir)
	}
	if workDir == "" {
		workDir = "."
	}
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return processruntime.Options{}, nil, err
	}
	options := o.RunOptions
	fields := options.OptionFields
	if fields == nil { fields = runconfig.Present(options) }
	configured := config.TestOptionsFor(string(framework))
	if err := runconfig.Apply(&options, configured, fields); err != nil { return processruntime.Options{}, nil, err }
	caller := options.Runtime
	if fields["test-timeout"] || configured.TestTimeout != nil {
		timeout := options.TestTimeout.String()
		caller.Timeout = &timeout
	}
	runtime, args, err := config.RuntimeFor("test", string(framework), caller)
	if err != nil { return runtime, nil, err }
	if !fields["extra-args"] { args = append(args, configured.ExtraArgs...) }
	if !fields["tags"] && configured.Tags != nil {
		if runner, ok := o.registryRunner(framework); ok {
			if mapper, ok := runner.(runners.BuildTagsMapper); ok { args = append(args, mapper.BuildTagsArgs(options.Tags)...) } else if len(options.Tags) > 0 { return runtime, nil, fmt.Errorf("tags unsupported by %s", framework) }
		}
	}
	if framework == parsers.Ginkgo && configured.Nodes != nil && !fields["nodes"] && options.Nodes != 0 { args = append(args, fmt.Sprintf("--nodes=%d", options.Nodes)) }
	if configured.Bench != nil && !fields["bench"] && options.Bench != "" && (framework == parsers.GoTest || framework == parsers.Ginkgo) { args = append(args, "-bench="+options.Bench) }
	return runtime, args, nil
}

func (o *TestOrchestrator) registryRunner(framework parsers.Framework) (runners.Runner, bool) {
	registry := o.registry
	if registry == nil { registry = DefaultRegistry(o.WorkDir) }
	return registry.Get(framework)
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
