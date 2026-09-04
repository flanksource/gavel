package status

import (
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// Prompts returns the overridable prompt templates owned by the status package:
// the per-file AI summary used by `gavel status --ai`. The override is the typed
// status.summaryPrompt field, resolved against Default at the call site.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.StatusSummary,
		Title:       "Status file summary",
		Description: "One-line AI summary of each changed file for `gavel status --ai`. Variable: {{details}} (the staged/unstaged diff or file contents). Output schema is fixed ({summary}).",
		ConfigPath:  "status.summary",
		Default:     fileSummaryPromptTemplate,
		UsedBy:      []string{"gavel status --ai"},
	}}
}

// ResolveSummaryPrompt returns the resolved status.summary override for workDir,
// or the embedded default when unset. A configured-but-missing file is a hard
// error.
func ResolveSummaryPrompt(workDir string) (string, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml for status summary prompt: %w", err)
	}
	return cfg.Status.Summary.TemplateSource(workDir, fileSummaryPromptTemplate)
}

// ResolveSummaryModel returns the model `gavel status --ai` runs on: the
// status.summary spec over the ai: base, with override (--ai-model) on top.
//
// status.summary is a PromptSpec, which carries a model as well as a prompt.
// Only the prompt half was ever read, so the model came from a hardcoded
// ai.DefaultConfig() and neither .gavel.yaml nor --ai-model could move it.
func ResolveSummaryModel(workDir string, override api.Model) (api.Model, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return api.Model{}, fmt.Errorf("load .gavel.yaml for status summary model: %w", err)
	}
	return cfg.ModelFor(cfg.Status.Summary, override), nil
}
