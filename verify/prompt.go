package verify

import (
	_ "embed"

	"github.com/flanksource/gomplate/v3"
)

//go:embed verify-prompt.md
var verifyPromptTemplate string

func renderPrompt(scope ReviewScope, cfg VerifyConfig) (string, error) {
	checks := EnabledChecks(cfg.Checks)
	byCategory := ChecksByCategory(checks)

	data := map[string]any{
		"scope":        scope,
		"extra_prompt": cfg.Prompt,
		"categories":   byCategory,
		"catOrder":     AllCategories,
		"ratings":      RatingDimensions,
	}
	return gomplate.RunTemplate(data, gomplate.Template{
		Template: verifyPromptTemplate,
	})
}
