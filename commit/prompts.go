package commit

import "github.com/flanksource/gavel/prompts"

// Prompts returns the overridable prompt templates owned by the commit package:
// the PR title/body/branch generator. The override is the typed
// commit.prContentPrompt field, resolved against Default at the call sites that
// build a PRContentInput. The settings UI composes this with the other packages'
// Prompts() to render an editor per prompt.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.PRContent,
		Title:       "PR content",
		Description: "Generates the PR title, body, and branch name when opening a pull request.",
		ConfigPath:  "commit.prContentPrompt",
		Default:     prContentPromptTemplate,
	}}
}
