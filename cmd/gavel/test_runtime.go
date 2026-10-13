package main

import "github.com/flanksource/gavel/testrunner"

func configuredLintTimeout(opts testrunner.RunOptions) string {
	if opts.LintTimeout > 0 {
		return opts.LintTimeout.String()
	}
	return ""
}
