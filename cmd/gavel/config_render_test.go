package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	promptregistry "github.com/flanksource/gavel/prompts/registry"
	"github.com/flanksource/gavel/verify"
	goccyyaml "github.com/goccy/go-yaml"
)

func TestConfigPrettyUsesMergedYAMLWhenStdoutNotTTY(t *testing.T) {
	previous := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return false }
	t.Cleanup(func() {
		stdoutIsTerminal = previous
	})

	result := ConfigResult{
		Merged: verify.GavelConfig{
			Verify:   verify.VerifyConfig{Model: "claude"},
			Fixtures: verify.FixturesConfig{Enabled: true},
		},
	}

	output := result.Pretty().String()
	if !strings.Contains(output, "verify:\n  model: claude\n") {
		t.Fatalf("expected merged verify config in plain output, got:\n%s", output)
	}
	if !strings.Contains(output, "fixtures:\n  enabled: true\n") {
		t.Fatalf("expected merged fixtures config in plain output, got:\n%s", output)
	}
	if strings.Contains(output, "# from ") {
		t.Fatalf("expected redirected output without source comments, got:\n%s", output)
	}
}

func TestResolvedConfigResultStructuredAndPrettyOutput(t *testing.T) {
	result := ResolvedConfigResult{
		Config: verify.GavelConfig{Verify: verify.VerifyConfig{
			Model:          "claude-code-sonnet",
			PromptTemplate: verify.InlinePrompt("Review {{patch}}"),
		}},
		Trace: verify.GavelConfigTrace{
			Merged: verify.GavelConfig{Verify: verify.VerifyConfig{Model: "claude-code-sonnet"}},
		},
		Prompts: []promptregistry.ResolvedPrompt{{
			ID: "verify", Title: "Verify reviewer", ConfigPath: "verify.promptTemplate",
			Source: "inline", Raw: "Review {{patch}}", Body: "Review {{patch}}",
			EffectiveModel: api.Model{Name: "claude-code-sonnet", Backend: api.BackendClaudeCLI},
			ModelSource:    "verify.model",
		}},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal resolved result: %v", err)
	}
	if !strings.Contains(string(data), `"config"`) || !strings.Contains(string(data), `"prompts"`) {
		t.Fatalf("expected config + prompts wrapper, got: %s", data)
	}
	yamlData, err := goccyyaml.Marshal(result)
	if err != nil {
		t.Fatalf("marshal resolved yaml: %v", err)
	}
	if !strings.Contains(string(yamlData), "config:\n") || !strings.Contains(string(yamlData), "prompts:\n") {
		t.Fatalf("expected resolved yaml wrapper, got:\n%s", yamlData)
	}
	if !strings.Contains(string(yamlData), "inline: Review {{patch}}") {
		t.Fatalf("expected raw inline prompt as YAML text, got:\n%s", yamlData)
	}

	pretty := result.Pretty()
	for format, output := range map[string]string{"ansi": pretty.ANSI(), "markdown": pretty.Markdown()} {
		for _, want := range []string{"Resolved Gavel config", "AI prompts", "Verify reviewer", "claude-code-sonnet", "Review {{patch}}"} {
			if !strings.Contains(output, want) {
				t.Fatalf("resolved %s output missing %q:\n%s", format, want, output)
			}
		}
	}
}

func TestConfigYAMLRendererDecodesRawInlinePrompt(t *testing.T) {
	result := ConfigResult{Merged: verify.GavelConfig{
		Verify: verify.VerifyConfig{
			Model:          "claude",
			PromptTemplate: verify.InlinePrompt("Review {{patch}}"),
		},
	}}
	output := result.Pretty().String()
	if !strings.Contains(output, "inline: Review {{patch}}") {
		t.Fatalf("expected raw inline prompt to render as YAML text, got:\n%s", output)
	}
}

func TestConfigPrettyAnnotatesNonGitRootSources(t *testing.T) {
	previous := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return true }
	t.Cleanup(func() {
		stdoutIsTerminal = previous
	})

	gitRoot := t.TempDir()
	gitRootConfig := verify.GavelConfig{
		Commit: verify.CommitConfig{Model: "repo"},
	}
	targetConfig := verify.GavelConfig{
		Commit: verify.CommitConfig{
			Hooks: []verify.CommitHook{{
				Name: "svc",
				Run:  "echo service",
			}},
		},
	}

	merged := verify.GavelConfig{Verify: verify.DefaultVerifyConfig()}
	merged = verify.MergeGavelConfig(merged, gitRootConfig)
	merged = verify.MergeGavelConfig(merged, targetConfig)

	targetPath := filepath.Join(gitRoot, "service", ".gavel.yaml")
	result := ConfigResult{
		GitRoot: gitRoot,
		Sources: []verify.GavelConfigSource{
			{
				Origin: "git-root",
				Path:   filepath.Join(gitRoot, ".gavel.yaml"),
				Config: gitRootConfig,
			},
			{
				Origin: "target-directory",
				Path:   targetPath,
				Config: targetConfig,
			},
		},
		Merged: merged,
	}

	output := result.Pretty().String()
	if !strings.Contains(output, "# from "+targetPath) {
		t.Fatalf("expected target-directory comment in pretty output, got:\n%s", output)
	}
	if strings.Contains(output, filepath.Join(gitRoot, ".gavel.yaml")) {
		t.Fatalf("did not expect git-root config path comment in pretty output, got:\n%s", output)
	}
	if !strings.Contains(output, "hooks:\n    - name: svc\n      run: echo service\n") {
		t.Fatalf("expected merged hooks block in pretty output, got:\n%s", output)
	}
}

func TestConfigResultMarshalJSONOnlyMergedConfig(t *testing.T) {
	result := ConfigResult{
		TargetPath: "/repo/service",
		Sources: []verify.GavelConfigSource{{
			Origin: "target-directory",
			Path:   "/repo/service/.gavel.yaml",
		}},
		Merged: verify.GavelConfig{
			Verify: verify.VerifyConfig{Model: "claude"},
			Commit: verify.CommitConfig{Model: "opus"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}

	output := string(data)
	if strings.Contains(output, "targetPath") || strings.Contains(output, "sources") || strings.Contains(output, "merged") {
		t.Fatalf("expected merged config only in json output, got: %s", output)
	}
	if !strings.Contains(output, `"verify":{"model":"claude"}`) {
		t.Fatalf("expected merged verify config in json output, got: %s", output)
	}
	if !strings.Contains(output, `"commit":{"model":"opus"}`) {
		t.Fatalf("expected merged commit config in json output, got: %s", output)
	}
}

func TestConfigResultMarshalYAMLOnlyMergedConfig(t *testing.T) {
	result := ConfigResult{
		TargetPath: "/repo/service",
		Sources: []verify.GavelConfigSource{{
			Origin: "target-directory",
			Path:   "/repo/service/.gavel.yaml",
		}},
		Merged: verify.GavelConfig{
			Verify: verify.VerifyConfig{Model: "claude"},
			Commit: verify.CommitConfig{Model: "opus"},
		},
	}

	data, err := goccyyaml.Marshal(result)
	if err != nil {
		t.Fatalf("marshal yaml: %v", err)
	}

	output := string(data)
	if strings.Contains(output, "targetPath:") || strings.Contains(output, "sources:") || strings.Contains(output, "merged:") {
		t.Fatalf("expected merged config only in yaml output, got:\n%s", output)
	}
	if !strings.Contains(output, "verify:\n  model: claude\n") {
		t.Fatalf("expected merged verify config in yaml output, got:\n%s", output)
	}
	if !strings.Contains(output, "commit:\n  model: opus\n") {
		t.Fatalf("expected merged commit config in yaml output, got:\n%s", output)
	}
}
