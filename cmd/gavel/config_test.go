package main

import (
	"strings"
	"testing"

	"github.com/flanksource/clicky"
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

	helpText := configHelp(cmd)
	help := helpText.String()
	for _, want := range []string{
		"UBER EXAMPLE",
		"gavel.yaml.example",
		"ai:",
		"lint:",
		"fix:",
		"lint.fix.model",
		"commit:",
		"message:",
		"commit.message.model",
		"grouping:",
		"summary:",
		"todos:",
		"run:",
		"plan:",
		"status:",
		"test:",
		"outlineSummary:",
		"pr:",
		"content:",
		"precommit:",
		"fixtures:",
		"checks:",
		"ssh:",
		"pre:",
		"post:",
		"secrets:",
		"procfile:",
		"--resolve",
		"{config, prompts}",
	} {
		if !strings.Contains(help, want) {
			t.Fatalf("config help missing %q:\n%s", want, help)
		}
	}
	for _, removed := range []string{"verify:", "compatibility:", "aiFix:"} {
		if strings.Contains(help, removed) {
			t.Fatalf("config help contains removed config %q:\n%s", removed, help)
		}
	}

	coloredExample := clicky.CodeBlock("yaml", gaveldocs.GavelConfigExample).ANSI()
	if !strings.Contains(helpText.ANSI(), coloredExample) {
		t.Fatalf("config help should render the uber example as colorized YAML:\n%s", helpText.ANSI())
	}
	if !strings.Contains(helpText.Markdown(), "```yaml\n"+gaveldocs.GavelConfigExample) {
		t.Fatalf("config help should preserve the uber example as a YAML code block in markdown")
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
	if len(resolved.Prompts) != 9 {
		t.Fatalf("resolved prompts = %d, want 9", len(resolved.Prompts))
	}
}
