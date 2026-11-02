package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func runPromptLint(ctx context.Context, options lintprompt.Options) (*testui.Snapshot, error) {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	reporter, err := newPromptLintReporter(promptLintReportOptions{
		Options: LintOptions{
			WorkDir: options.Dirs.Invocation, Dirs: options.Dirs, Exclude: options.Exclude,
			Summary: options.Summary, SummaryLimit: options.SummaryLimit, Selection: options.Selection,
		},
		Writer: os.Stderr, Publisher: options.Publisher, Cancel: cancel,
	})
	if err != nil {
		return nil, err
	}
	options.Publisher = reporter
	renderer := captaincli.NewEventRenderer(os.Stderr)
	options.OnEvent = func(iteration int, event api.Event) {
		if event.Rejected || event.Kind == api.EventError || (event.Kind == api.EventSystem && strings.HasPrefix(event.Text, "schema validation failed")) {
			renderer.Handle(iteration, event)
		}
	}
	results, runErr := lintprompt.Run(ctx, options)
	runErr = errors.Join(runErr, renderer.Flush())
	snapshot, reportErr := reporter.Finish(results, errors.Join(runErr, context.Cause(ctx)))
	return snapshot, errors.Join(runErr, reportErr)
}

func promptLintReturnValue(snapshot *testui.Snapshot) any {
	exitCode = 0
	if snapshot.Error != "" || linters.HasFailures(snapshot.Lint) {
		exitCode = 1
	}
	if isPrettyFormat() {
		return nil
	}
	return snapshot
}
