package testrunner

import (
	"fmt"
	"strings"
	"time"

	"os"
	"path/filepath"

	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/flanksource/gavel/verify"
)

func (o *TestOrchestrator) buildCommand(runner runners.Runner, pkg string, opts runners.TestOptions) (*runners.TestRun, error) {
	runtime, configured, err := o.runtimeFor(runner.Name(), pkg)
	if err != nil {
		return nil, err
	}
	timeout, err := runtimeTimeout(runtime)
	if err != nil {
		return nil, err
	}
	opts, err = configured.Merge(opts)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", runner.Name(), pkg, err)
	}
	opts.Timeout = subprocessTimeoutFor(timeout)
	run, err := runner.BuildCommand(opts.ForPackages(pkg))
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
	run.Slots = runtime.Slots
	return run, nil
}

// runtimeTimeout is the per-process deadline a runtime configures: 10m by
// default, none when set to "".
func runtimeTimeout(runtime processruntime.Options) (time.Duration, error) {
	if runtime.Timeout == nil {
		return 10 * time.Minute, nil
	}
	if *runtime.Timeout == "" {
		return 0, nil
	}
	timeout, err := time.ParseDuration(*runtime.Timeout)
	if err != nil {
		return 0, fmt.Errorf("invalid test runtime timeout %q: %w", *runtime.Timeout, err)
	}
	return timeout, nil
}

// runtimeFor resolves the process runtime and the .gavel.yaml test options
// for framework's package pkg.
func (o *TestOrchestrator) runtimeFor(framework parsers.Framework, pkg string) (processruntime.Options, runners.TestOptions, error) {
	workDir := o.runtimeWorkDir()
	if pkg != "" {
		workDir = pkg
		if !filepath.IsAbs(workDir) {
			workDir = filepath.Join(o.WorkDir, pkg)
		}
		if info, err := os.Stat(workDir); err != nil {
			return processruntime.Options{}, runners.TestOptions{}, err
		} else if !info.IsDir() {
			workDir = filepath.Dir(workDir)
		}
		workDir = verify.RuntimeConfigDirectory(workDir, o.runtimeConfigDir)
	}
	if workDir == "" {
		workDir = "."
	}
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return processruntime.Options{}, runners.TestOptions{}, err
	}
	runtime, args, err := o.configuredRuntime(config, framework)
	if err != nil {
		return runtime, runners.TestOptions{}, err
	}
	opts, err := o.configuredTestOptions(framework, args)
	return runtime, opts, err
}

// configuredTestOptions parses a package's .gavel.yaml runner arguments with
// the framework's own flags.
func (o *TestOrchestrator) configuredTestOptions(framework parsers.Framework, args []string) (runners.TestOptions, error) {
	cwd, err := filepath.Abs(o.WorkDir)
	if err != nil {
		return runners.TestOptions{}, err
	}
	if len(args) == 0 {
		return runners.TestOptions{Cwd: cwd}, nil
	}
	runner, ok := o.registryRunner(framework)
	if !ok {
		return runners.TestOptions{}, fmt.Errorf("no runner registered for framework %s", framework)
	}
	opts, err := runners.ParseTestArgs(cwd, append([]string{"--"}, args...), runner.Flags()...)
	if err != nil {
		return runners.TestOptions{}, fmt.Errorf(".gavel.yaml %s arguments %q: %w", framework, args, err)
	}
	return *opts, nil
}

// configuredRuntime resolves framework's process runtime and the runner
// arguments its .gavel.yaml configures.
func (o *TestOrchestrator) configuredRuntime(config verify.GavelConfig, framework parsers.Framework) (processruntime.Options, []string, error) {
	options := o.RunOptions
	fields := options.OptionFields
	if fields == nil {
		fields = runconfig.Present(options)
	}
	configured := config.TestOptionsFor(string(framework))
	if err := runconfig.Apply(&options, configured, fields); err != nil {
		return processruntime.Options{}, nil, err
	}
	caller := options.Runtime
	if fields["test-timeout"] || configured.TestTimeout != nil {
		timeout := options.TestTimeout.String()
		if options.TestTimeout == 0 {
			timeout = ""
		}
		caller.Timeout = &timeout
	}
	runtime, args, err := config.RuntimeFor("test", string(framework), caller)
	if err != nil {
		return runtime, nil, err
	}
	if !fields["extra-args"] {
		args = append(args, configured.ExtraArgs...)
	}
	if !fields["tags"] && configured.Tags != nil && len(options.Tags) > 0 {
		if !goFamily(framework) {
			return runtime, nil, fmt.Errorf("tags unsupported by %s", framework)
		}
		args = append(args, "-tags="+strings.Join(options.Tags, ","))
	}
	if framework == parsers.Ginkgo && configured.Nodes != nil && !fields["nodes"] && options.Nodes != 0 {
		args = append(args, fmt.Sprintf("--nodes=%d", options.Nodes))
	}
	if configured.Bench != nil && !fields["bench"] && options.Bench != "" && goFamily(framework) {
		args = append(args, "-bench="+options.Bench)
	}
	return runtime, args, nil
}

func (o *TestOrchestrator) registryRunner(framework parsers.Framework) (runners.Runner, bool) {
	registry := o.registry
	if registry == nil {
		registry = DefaultRegistry(o.WorkDir)
	}
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
