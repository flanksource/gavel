package git

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	"github.com/samber/lo"
)

func Summarize(commits models.CommitAnalyses, options SummaryOptions) (GitSummaries, error) {
	clicky.Infof("Generating git summary with window=%s, maxCategories=%d", options.Window, options.MaxCategories)
	if len(commits) == 0 {
		return GitSummaries{}, nil
	}

	windows := CalculateTimeWindows(commits.From(), commits.To(), options.Window)

	logger.Debugf("Using time windows: %v", windows)

	// Group commits by (window, scope)
	grouped := make(map[windowScopeKey]models.CommitAnalyses)

	for _, commit := range commits {
		window := GetWindowForCommit(commit, windows)
		if window == nil {
			continue
		}

		// Treat unknown scopes as "Other" to ensure all commits are included
		scope := commit.Scope
		if scope == models.ScopeTypeUnknown {
			scope = models.ScopeTypeOther
		}

		key := windowScopeKey{
			windowStart: window.Start,
			scope:       scope,
		}
		grouped[key] = append(grouped[key], commit)
	}

	// Select top scopes per window (leaving room for "Other")
	topScopesPerWindow := SelectTopScopesPerWindow(grouped, options.MaxCategories)

	// Collect all groups for batch processing
	var groups []summaryGroup
	otherCommits := make(map[time.Time]models.CommitAnalyses)

	// Collect top scopes and track "Other" commits
	for key, commits := range grouped {
		windowStart := key.windowStart
		scope := key.scope

		// Check if this scope is in the top scopes for this window
		topScopes := topScopesPerWindow[windowStart]
		isTopScope := false
		for _, topScope := range topScopes {
			if topScope == scope {
				isTopScope = true
				break
			}
		}

		// Find the window object
		window := &TimeWindow{Start: windowStart}
		for _, w := range windows {
			if w.Start.Equal(windowStart) {
				window = &w
				break
			}
		}

		if isTopScope {
			repositories := make(map[string]struct{})
			for _, commit := range commits {
				if commit.Repository != "" {
					repositories[commit.Repository] = struct{}{}
				}
			}
			groups = append(groups, summaryGroup{
				windowStart:  windowStart,
				window:       window,
				scope:        scope,
				commits:      commits,
				repositories: repositories,
				isOther:      false,
			})
		} else {
			// Add to "Other" bucket for this window
			otherCommits[windowStart] = append(otherCommits[windowStart], commits...)
		}
	}

	// Add "Other" groups
	for windowStart, commits := range otherCommits {
		if len(commits) == 0 {
			continue
		}

		window := &TimeWindow{Start: windowStart}
		for _, w := range windows {
			if w.Start.Equal(windowStart) {
				window = &w
				break
			}
		}

		repositories := make(map[string]struct{})
		for _, commit := range commits {
			if commit.Repository != "" {
				repositories[commit.Repository] = struct{}{}
			}
		}
		groups = append(groups, summaryGroup{
			windowStart:  windowStart,
			window:       window,
			scope:        models.ScopeTypeOther,
			commits:      commits,
			repositories: repositories,
			isOther:      true,
		})
	}

	// Process groups in parallel if AI is enabled, otherwise sequentially
	var summaries []GitSummary

	if (options.Agent != nil || options.AgentFactory != nil) && options.Context != nil {
		// Use batch processing for AI-powered summaries
		maxWorkers := options.MaxWorkers
		if maxWorkers <= 0 {
			maxWorkers = 3 // Default concurrency
		}

		batch := task.Batch[GitSummary]{
			Name:        "Generate AI Summaries",
			MaxWorkers:  maxWorkers,
			ItemTimeout: time.Minute * 5,
		}

		for _, group := range groups {
			group := group // Capture for closure
			batch.Items = append(batch.Items, func(logger logger.Logger) (GitSummary, error) {
				count := AggregateCommitGroup(group.commits)
				windowLabel := formatTimeWindow(group.window, options.Window)

				logger.Infof("generating summary for %s in %s", group.scope, windowLabel)

				name, desc, err := GenerateGroupSummary(options.Context, group.scope, windowLabel, group.commits, options)
				if err != nil {
					return GitSummary{}, fmt.Errorf("generate %s summary: %w", group.scope, err)
				}

				scopes := lo.Keys(count.Scopes)
				tech := lo.Keys(count.Tech)
				repositories := lo.Keys(group.repositories)
				sort.Strings(repositories)

				return GitSummary{
					Group:        name,
					Description:  desc,
					TimeWindow:   windowLabel,
					From:         &group.window.Start,
					Until:        &group.window.End,
					Scopes:       scopes,
					Tech:         tech,
					Repositories: repositories,
					Commits:      count,
				}, nil
			})
		}

		// Execute batch and collect results
		var summaryErr error
		for item := range batch.Run() {
			if item.Error != nil {
				summaryErr = errors.Join(summaryErr, item.Error)
				continue
			}
			summaries = append(summaries, item.Value)
		}
		if summaryErr != nil {
			return nil, summaryErr
		}
	} else {
		// Non-AI path: process sequentially without batch
		for _, group := range groups {
			count := AggregateCommitGroup(group.commits)
			name, desc := GenerateFallbackDescription(group.scope, group.commits)

			scopes := lo.Keys(count.Scopes)
			tech := lo.Keys(count.Tech)
			repositories := lo.Keys(group.repositories)
			sort.Strings(repositories)

			summary := GitSummary{
				Group:        name,
				Description:  desc,
				TimeWindow:   formatTimeWindow(group.window, options.Window),
				From:         &group.window.Start,
				Until:        &group.window.End,
				Scopes:       scopes,
				Tech:         tech,
				Repositories: repositories,
				Commits:      count,
			}

			summaries = append(summaries, summary)
		}
	}

	// Sort newest first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].From.After(*summaries[j].From)
	})

	clicky.Infof("Generated %d summary items", len(summaries))

	return summaries, nil
}
