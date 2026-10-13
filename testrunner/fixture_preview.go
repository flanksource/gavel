package testrunner

import (
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
)

func printFixtureDryRun(tree *fixtures.FixtureNode, opts fixtures.RunnerOptions) {
	for _, path := range opts.Paths {
		PrintDryRunCommand("fixture", path, opts.ExecutablePath, []string{"fixtures", path}, opts.WorkDir)
	}
	if summary := opts.ProcessRuntime.Summary().String(); summary != "" {
		logger.Infof("  runtime: %s", summary)
	}
	tree.Walk(func(node *fixtures.FixtureNode) {
		if node.Test != nil {
			logger.Infof("  fixture: %s", node.Test.Name)
		}
	})
}
