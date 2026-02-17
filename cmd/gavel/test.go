package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner"
)

func init() {
	testCmd := clicky.AddNamedCommand("test", rootCmd, testrunner.RunOptions{}, testrunner.Run)
	testCmd.Flags().SetInterspersed(false)
}
