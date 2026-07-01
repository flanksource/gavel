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
		"daemon:",
		"{{.port}}",
		"ai:",
		"verify:",
		"claude-code-sonnet",
		"Focus on the new parser path",
		"Parser errors are actionable",
		"json.score >= 80",
		"maxTokens",
		"maxConcurrent",
		"cacheTTL",
		"noCache",
		"FORMAT 4: TEST / LINT STEPS",
		"```yaml test",
		"```lint",
		"test-timeout: 2m",
		"Common test keys",
		"Common lint keys",
		"gavel test --help",
		"gavel lint --help",
		"show-passed",
		"show-failed",
		"test, lint (or yaml test/yaml lint)",
		"expected count, count",
		"expected matches, matches",
		"query",
		"expectations",
		"ansi.has_cursor_hide",
		"ansi.final_text",
		"has_color(s)",
		"OUTPUT OPTIONS",
		"--show-passed",
		"--show-stdout",
		"--show-stderr",
		"--update-golden",
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
