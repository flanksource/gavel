package commit

import "github.com/flanksource/gavel/prompts"

// Prompts returns the overridable prompt templates owned by the commit package:
// the PR title/body/branch generator and the AI commit-grouping prompt. Overrides
// are the typed commit.prContentPrompt / commit.groupingPrompt fields, resolved
// against Default at their call sites. The settings UI composes this with the
// other packages' Prompts() to render an editor per prompt.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{
		{
			ID:          prompts.PRContent,
			Title:       "PR content",
			Description: "Generates the PR title, body, and branch name when opening a pull request.",
			ConfigPath:  "commit.prContentPrompt",
			Default:     prContentPromptTemplate,
		},
		{
			ID:          prompts.CommitGrouping,
			Title:       "Commit grouping",
			Description: "Groups uncommitted changes into logical commits for `gavel commit -A`. Available variables: {{table}} (the status table), {{maxCommits}}. Output schema is groups[]+ignore[]; {{maxCommits}} also caps the groups array via maxItems in the frontmatter output.schema, enforced by captain's schemaStrictness=retry policy.",
			ConfigPath:  "commit.groupingPrompt",
			Default:     groupingPromptTemplate,
		},
	}
}
