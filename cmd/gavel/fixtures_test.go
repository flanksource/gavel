package main

import (
	"strings"
	"testing"
)

func TestFixturesHelpIncludesArgumentAndFlagReference(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"fixtures"})
	if err != nil {
		t.Fatalf("find fixtures command: %v", err)
	}

	help := fixturesHelp(cmd).ANSI()
	for _, want := range []string{
		"USAGE",
		"ARGUMENTS",
		"fixture-files",
		"FLAGS",
		"--update-golden",
		"--show-passed",
		"--show-stdout",
		"--show-stderr",
		"GLOBAL FLAGS",
		"--cwd",
		"-v, --loglevel",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("fixtures help missing %q:\n%s", want, help)
		}
	}
}
