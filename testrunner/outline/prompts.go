package outline

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// Prompts returns the overridable prompt templates owned by the outline package:
// the per-test AI summary used by `gavel test outline --ai-summary`. The override
// is the typed test.outlineSummaryPrompt field, resolved against Default at the
// call site.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.TestOutlineSummary,
		Title:       "Test outline summary",
		Description: "One-line AI summary of what each test verifies for `gavel test outline --ai-summary`. Variables: {{ids}} (the test ids), {{file}}, {{source}}. Output schema is fixed (tests[]).",
		ConfigPath:  "test.outlineSummary",
		Default:     testSummaryPromptTemplate,
		UsedBy:      []string{"gavel test outline --ai-summary"},
	}}
}

// resolveSummaryPrompt returns the resolved test.outlineSummaryPrompt override
// for workDir, or the embedded default when unset. A configured-but-missing file
// is a hard error.
func resolveSummaryPrompt(workDir string) (string, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml for test outline summary prompt: %w", err)
	}
	return cfg.Test.OutlineSummary.TemplateSource(workDir, testSummaryPromptTemplate)
}

// resolveSummaryModel returns the model `gavel test outline --ai-summary` runs
// on: the test.outlineSummary spec over the ai: base.
//
// The config slot always carried a model — test.outlineSummary is a PromptSpec —
// but only its prompt half was read, so the model was pinned to a hardcoded
// ai.DefaultConfig() and the command had no way to select one at all.
func resolveSummaryModel(workDir string) (api.Model, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return api.Model{}, fmt.Errorf("load .gavel.yaml for test outline summary model: %w", err)
	}
	return cfg.ModelFor(cfg.Test.OutlineSummary, api.Model{}), nil
}
