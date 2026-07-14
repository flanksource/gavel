package verify

import "github.com/flanksource/gavel/prompts"

// DefaultPromptTemplate is the embedded generic AI-review template used by AI
// fixture steps.
func DefaultPromptTemplate() string { return verifyPromptTemplate }

// Prompts returns the overridable prompt templates owned by the verify package.
// The settings UI composes this with the other packages' Prompts() to render an
// editor per prompt; the override itself is the typed verify.promptTemplate
// field, resolved against Default at the call site (see resolveVerifyPrompt).
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.Verify,
		Title:       "AI fixture reviewer",
		Description: "The reviewer prompt used by AI fixture steps.",
		ConfigPath:  "verify.promptTemplate",
		Default:     verifyPromptTemplate,
		ModelPolicy: prompts.ModelFromVerifyConfig,
	}}
}
