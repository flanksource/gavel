package git

import (
	"context"

	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/clicky/ai"
)

// Example usage of Summarize with AI integration
func ExampleSummarize() {
	// Get commits from analyzer
	commits := CommitAnalyses{} // populated from git analyzer

	// Option 1: Use fallback descriptions (no AI)
	summariesBasic, _ := Summarize(commits, SummaryOptions{
		Window:        GroupByMonth,
		MaxCategories: 6,
	})

	// Option 2: Use AI-powered descriptions
	agent, _ := ai.GetDefaultAgent() // or your custom agent
	summariesWithAI, _ := Summarize(commits, SummaryOptions{
		Window:        GroupByMonth,
		MaxCategories: 6,
		Agent:         agent,
		Context:       context.Background(),
	})

	_ = summariesBasic
	_ = summariesWithAI
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
