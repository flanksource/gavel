package verify

import "github.com/flanksource/gavel/prompts"

// Prompts returns the overridable prompt templates owned by the verify package.
// The settings UI composes this with the other packages' Prompts() to render an
// editor per prompt; the override itself is the typed verify.promptTemplate
// field, resolved against Default at the call site (see resolveVerifyPrompt).
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.Verify,
		Title:       "Verify reviewer",
		Description: "The reviewer prompt for `gavel verify` and fixture AI checks.",
		ConfigPath:  "verify.promptTemplate",
		Default:     verifyPromptTemplate,
	}}
}
