package main

import (
	"strings"
	"testing"

	gaveldocs "github.com/flanksource/gavel"
)

func TestConfigHelpIncludesExample(t *testing.T) {
	if !strings.Contains(gaveldocs.GavelConfigExample, "precommit:") {
		t.Fatalf("embedded config example is missing precommit section:\n%s", gaveldocs.GavelConfigExample)
	}

	cmd, _, err := rootCmd.Find([]string{"config"})
	if err != nil {
		t.Fatalf("find config command: %v", err)
	}

	help := configHelp(cmd).ANSI()
	for _, want := range []string{
		"UBER EXAMPLE",
		"gavel.yaml.example",
		"verify:",
		"precommit:",
		"fixtures:",
		"secrets:",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("config help missing %q:\n%s", want, help)
		}
	}
}
