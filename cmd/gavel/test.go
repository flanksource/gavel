package main

import (
	"github.com/flanksource/clicky"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

func runTests(opts testrunner.RunOptions) (any, error) {
	result, err := testrunner.Run(opts)
	if err != nil {
		return result, err
	}
	if tests, ok := result.([]parsers.Test); ok {
		if parsers.Tests(tests).Sum().Failed > 0 {
			exitCode = 1
		}
	}
	return result, nil
}

func init() {
	testCmd := clicky.AddNamedCommand("test", rootCmd, testrunner.RunOptions{}, runTests)
	testCmd.Flags().SetInterspersed(false)
}
