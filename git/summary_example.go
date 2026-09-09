package git

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

// Example usage of Summarize with AI integration
func ExampleSummarize() error {
	// Get commits from analyzer
	commits := models.CommitAnalyses{} // populated from git analyzer

	// Option 1: Use fallback descriptions (no AI)
	summariesBasic, err := Summarize(commits, SummaryOptions{
		Window:        GroupByMonth,
		MaxCategories: 6,
	})
	if err != nil {
		return err
	}

	// Option 2: Resolve each summary before constructing its provider.
	summariesWithAI, err := Summarize(commits, SummaryOptions{
		Window:        GroupByMonth,
		MaxCategories: 6,
		AgentFactory:  ai.NewAgent,
		Prompt:        verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "api:haiku"}}},
		Context:       context.Background(),
	})
	if err != nil {
		return err
	}

	_ = summariesBasic
	_ = summariesWithAI
	return nil
}

// ExampleSummarizeWithScope shows how to use repomap for scoping
func ExampleSummarizeWithScope() {
	// Commits should already have Scope field populated from repomap analysis
	// The Summarize function will:
	// 1. Group by time window
	// 2. Within each window, select top N scopes by commit count
	// 3. Create summaries for each (window, scope) combination
	// 4. Use AI to generate names/descriptions if agent provided
	//
	// Example commit with scope from repomap:
	// commit := CommitAnalysis{
	//   Commit: Commit{
	//     Scope: ScopeTypeApp, // from repomap.GetScopeByPath()
	//   },
	//   Changes: []CommitChange{...},
	// }
}
