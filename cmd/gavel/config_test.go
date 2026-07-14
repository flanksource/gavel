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
		"--resolve",
		"{config, prompts}",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("config help missing %q:\n%s", want, help)
		}
	}
}

func TestRunConfigResolveReturnsResolvedResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	target := t.TempDir()
	result, err := runConfig(ConfigOptions{Args: []string{target}, Resolve: true})
	if err != nil {
		t.Fatalf("run config --resolve: %v", err)
	}
	resolved, ok := result.(ResolvedConfigResult)
	if !ok {
		t.Fatalf("result type = %T, want ResolvedConfigResult", result)
	}
	if len(resolved.Prompts) != 12 {
		t.Fatalf("resolved prompts = %d, want 12", len(resolved.Prompts))
	}
}
