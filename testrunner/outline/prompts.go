package outline

import (
	"fmt"

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
