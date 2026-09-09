package git

import "github.com/flanksource/gavel/prompts"

// Prompts returns the overridable prompt templates owned by the git package: the
// commit-message prompt and the commit-grouping summary prompt. The overrides are
// the typed commit.message / commit.summary specs, resolved against these defaults
// at the call sites (AnalyzeOptions for the message prompt, the summary path for
// the grouping-summary prompt). The settings UI composes this with the other
// packages' Prompts() to render an editor per prompt. ConfigPath equals the prompt
// ID so the settings UI can key each descriptor to its schema node's x-prompt-id.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{
		{
			ID:          prompts.CommitMessage,
			Title:       "Commit message",
			Description: "Generates the conventional-commit message for `gavel commit`.",
			ConfigPath:  prompts.CommitMessage,
			Default:     commitMessagePrompt,
			UsedBy:      []string{"gavel commit"},
		},
		{
			ID:          prompts.CommitSummary,
			Title:       "Commit grouping summary",
			Description: "Names and summarises a group of commits (`gavel git analyze --summary`).",
			ConfigPath:  prompts.CommitSummary,
			Default:     summaryGroupPrompt,
			UsedBy:      []string{"gavel git analyze --summary --ai"},
		},
	}
}
