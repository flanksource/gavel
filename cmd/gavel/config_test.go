package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/clicky"
	gaveldocs "github.com/flanksource/gavel"
	"github.com/flanksource/gavel/prompts/registry"
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
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai:\n  defaultModel: agent:claude-sonnet-5\n  timeout: 17m\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	result, err := runConfig(ConfigOptions{Args: []string{target}, Resolve: true})
	if err != nil {
		t.Fatalf("run config --resolve: %v", err)
	}
	resolved, ok := result.(ResolvedConfigResult)
	if !ok {
		t.Fatalf("result type = %T, want ResolvedConfigResult", result)
	}
	if want := len(registry.All()); len(resolved.Prompts) != want {
		t.Fatalf("resolved prompts = %d, want every registered prompt (%d)", len(resolved.Prompts), want)
	}
	for _, item := range resolved.Prompts {
		switch item.ID {
		case "commit.message":
			if item.Effective.Budget.Timeout != "17m" || item.Provenance["/budget/timeout"].Source.Key != "ai.timeout" {
				t.Fatalf("%s timeout = %q, origin = %#v; want saved timeout 17m", item.ID, item.Effective.Budget.Timeout, item.Provenance["/budget/timeout"])
			}
		case "pr.fix":
			if item.Effective.Budget.Timeout != "45m" || item.Provenance["/budget/timeout"].Source.LayerID != "default" {
				t.Fatalf("%s timeout = %q, origin = %#v; want authored timeout 45m", item.ID, item.Effective.Budget.Timeout, item.Provenance["/budget/timeout"])
			}
		}
	}
}
