package git

import "github.com/flanksource/gavel/prompts"

// Prompts returns the overridable prompt templates owned by the git package: the
// three commit-analysis prompts and the commit-grouping summary prompt. The
// overrides are the typed commit.*Prompt fields, resolved against these defaults
// at the call sites (AnalyzeOptions for the commit prompts, the summary path for
// the grouping prompt). The settings UI composes this with the other packages'
// Prompts() to render an editor per prompt.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{
		{
			ID:          prompts.CommitMessage,
			Title:       "Commit message",
			Description: "Generates the conventional-commit message for `gavel commit`.",
			ConfigPath:  "commit.messagePrompt",
			Default:     commitMessagePrompt,
			ModelPolicy: prompts.ModelFromCommitConfig,
		},
		{
			ID:          prompts.CommitFuncRemoved,
			Title:       "Functionality removed",
			Description: "Detects user-visible functionality a diff removes (pre-commit warning).",
			ConfigPath:  "commit.functionalityRemovedPrompt",
			Default:     functionalityRemovedPrompt,
			ModelPolicy: prompts.ModelFromCommitConfig,
		},
		{
			ID:          prompts.CommitCompatibility,
			Title:       "Compatibility issues",
			Description: "Detects backward-compatibility / breaking changes in a diff.",
			ConfigPath:  "commit.compatibilityPrompt",
			Default:     compatibilityIssuesPrompt,
			ModelPolicy: prompts.ModelFromCommitConfig,
		},
		{
			ID:          prompts.CommitSummary,
			Title:       "Commit grouping summary",
			Description: "Names and summarises a group of commits (`gavel git analyze --summary`).",
			ConfigPath:  "commit.summaryPrompt",
			Default:     summaryGroupPrompt,
			ModelPolicy: prompts.ModelFromDefault,
		},
	}
}
