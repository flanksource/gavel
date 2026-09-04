package git

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
)

// Example usage of Summarize with AI integration
func ExampleSummarize() {
	// Get commits from analyzer
	commits := models.CommitAnalyses{} // populated from git analyzer

	// Option 1: Use fallback descriptions (no AI)
	summariesBasic, _ := Summarize(commits, SummaryOptions{
		Window:        GroupByMonth,
		MaxCategories: 6,
	})

	// Option 2: Use AI-powered descriptions. The model is resolved from
	// .gavel.yaml by the caller (see verify.GavelConfig.ModelFor); there is no
	// default agent, because nothing may choose a model on the user's behalf.
	cfg := ai.DefaultConfig()
	cfg.Model = api.Model{Name: "api:haiku"}
	agent, _ := ai.NewAgent(cfg)
	defer agent.Close() //nolint:errcheck // process-backed backends only release their child on Close
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
