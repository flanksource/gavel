package main

import (
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner"
)

var fixturesUI fixtureUIOptions

type fixtureUIOptions struct {
	UI          bool
	Addr        string
	Detach      bool
	AutoStop    time.Duration
	IdleTimeout time.Duration
}

type fixtureUIRunRequest struct {
	Runner *fixtures.RunnerOptions
	UI     fixtureUIOptions
}

func fixtureUIRunOptions(request fixtureUIRunRequest) (testrunner.RunOptions, bool) {
	opts := testrunner.RunOptions{
		WorkDir:       request.Runner.WorkDir,
		UI:            request.UI.UI,
		Addr:          request.UI.Addr,
		Fixtures:      true,
		FixtureFiles:  append([]string(nil), request.Runner.Paths...),
		FixturesOnly:  true,
		FixtureRunner: request.Runner,
		SkipHooks:     true,
		AutoStop:      request.UI.AutoStop,
		IdleTimeout:   request.UI.IdleTimeout,
	}
	if request.Runner.Display != nil {
		opts.ShowPassed = request.Runner.Display.ShowPassed
		opts.ShowStdout = testrunner.OutputMode(request.Runner.Display.ShowStdout)
		opts.ShowStderr = testrunner.OutputMode(request.Runner.Display.ShowStderr)
	}
	return opts, request.UI.Detach
}

func init() {
	fixturesCmd.Flags().BoolVar(&fixturesUI.UI, "ui", false, "Launch a browser with the live fixture execution tree")
	fixturesCmd.Flags().StringVar(&fixturesUI.Addr, "addr", "0.0.0.0", "Interface to bind the fixture UI HTTP server")
	fixturesCmd.Flags().BoolVar(&fixturesUI.Detach, "detach", false, "With --ui, keep serving the completed fixture run in a detached process")
	fixturesCmd.Flags().DurationVar(&fixturesUI.AutoStop, "auto-stop", 0, "With --ui --detach, hard deadline for the detached UI server")
	fixturesCmd.Flags().DurationVar(&fixturesUI.IdleTimeout, "idle-timeout", 0, "With --ui --detach, stop serving after this long without HTTP requests")
}
