package git

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	. "github.com/flanksource/gavel/models"

	"github.com/samber/lo"
)

type GroupByWindow string

const (
	GroupByDay   GroupByWindow = "day"
	GroupByWeek  GroupByWindow = "week"
	GroupByMonth GroupByWindow = "month"
)

type SummaryOptions struct {
	Window        GroupByWindow `json:"window,omitempty" tag:"window" default:"month"`
	MaxCategories int           `json:"maxGroups,omitempty" tag:"maxGroups" default:"6"`
	// Agent for AI-powered summary generation (optional, uses fallback if nil)
	Agent ai.Agent `json:"-"`
	// Context for AI operations
	Context context.Context `json:"-"`
	// MaxWorkers for parallel AI summary generation (default: 3)
	MaxWorkers int `json:"-"`
}

type windowScopeKey struct {
	windowStart time.Time
	scope       ScopeType
}

type summaryGroup struct {
	windowStart  time.Time
	window       *TimeWindow
	scope        ScopeType
	commits      CommitAnalyses
	repositories map[string]struct{}
	isOther      bool
}

type GitSummary struct {

	// Group Name e.g. Terraform migration
	Group string `json:"group,omitempty" validate:"maxLen=30"`
	// A summary of all the commits in this group
	Description string `json:"description,omitempty" validate:"maxLen=200"`

	TimeWindow string     `json:"time_window,omitempty"`
	From       *time.Time `json:"from,omitempty" tag:"from"`
	Until      *time.Time `json:"until,omitempty" tag:"until"`

	Scopes       []ScopeType       `json:"scopes,omitempty"`
	Tech         []ScopeTechnology `json:"tech,omitempty"`
	Repositories []string          `json:"repositories,omitempty"`
	Commits      Count             `json:"commits,omitempty"`
}

type TimeWindow struct {
	Start time.Time
	End   time.Time
}

const Day = 24 * time.Hour
const Week = 7 * Day
const Month = 30 * Day

func CalculateTimeWindows(from, until time.Time, window GroupByWindow) []TimeWindow {
	var windows []TimeWindow

	if window == "" {
		if until.Sub(from) < Week {
			window = GroupByDay
		} else if until.Sub(from) < Week*8 {
			window = GroupByWeek
		} else {
			window = GroupByMonth
		}
	}

	switch window {
	case GroupByDay:
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
		for current.Before(until) || current.Equal(until) {
			end := time.Date(current.Year(), current.Month(), current.Day(), 23, 59, 59, 999999999, current.Location())
			windows = append(windows, TimeWindow{Start: current, End: end})
			current = current.AddDate(0, 0, 1)
		}

	case GroupByWeek:
		// Find Monday of the week containing 'from'
		weekday := int(from.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		daysToMonday := weekday - 1
		current := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location()).AddDate(0, 0, -daysToMonday)

		for current.Before(until) || current.Equal(until) {
			end := current.AddDate(0, 0, 7).Add(-time.Nanosecond)
			windows = append(windows, TimeWindow{Start: current, End: end})
			current = current.AddDate(0, 0, 7)
		}

	case GroupByMonth:
		current := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, from.Location())
		for current.Before(until) || current.Equal(until) {
			// Last day of month
			nextMonth := current.AddDate(0, 1, 0)
			end := nextMonth.Add(-time.Nanosecond)
			windows = append(windows, TimeWindow{Start: current, End: end})
			current = nextMonth
		}
	}

	return windows
}

func GetWindowForCommit(commit CommitAnalysis, windows []TimeWindow) *TimeWindow {
	commitTime := commit.Author.Date
	for i := range windows {
		if (commitTime.Equal(windows[i].Start) || commitTime.After(windows[i].Start)) &&
			(commitTime.Equal(windows[i].End) || commitTime.Before(windows[i].End)) {
			return &windows[i]
		}
	}
	return nil
}

func SelectTopScopes(commits CommitAnalyses, maxCategories int) []ScopeType {
	scopeCounts := make(map[ScopeType]int)
	for _, commit := range commits {
		if commit.Scope != ScopeTypeUnknown {
			scopeCounts[commit.Scope]++
		}
	}

	type scopeCount struct {
		scope ScopeType
		count int
	}

	var scopes []scopeCount
	for scope, count := range scopeCounts {
		scopes = append(scopes, scopeCount{scope, count})
	}

	sort.Slice(scopes, func(i, j int) bool {
		return scopes[i].count > scopes[j].count
	})

	limit := len(scopes)
	if maxCategories < limit {
		limit = maxCategories
	}

	result := make([]ScopeType, limit)
	for i := 0; i < limit; i++ {
		result[i] = scopes[i].scope
	}

	return result
}

// SelectTopScopesPerWindow selects the top N-1 scopes per time window to leave room for "Other"
// Returns a map of window start times to the top scope types for that window
func SelectTopScopesPerWindow(grouped map[windowScopeKey]CommitAnalyses, maxCategories int) map[time.Time][]ScopeType {
	if maxCategories <= 0 {
		return make(map[time.Time][]ScopeType)
	}

	// Group commits by window and count per scope
	windowScopeCounts := make(map[time.Time]map[ScopeType]int)
	for key, commits := range grouped {
		if _, exists := windowScopeCounts[key.windowStart]; !exists {
			windowScopeCounts[key.windowStart] = make(map[ScopeType]int)
		}
		windowScopeCounts[key.windowStart][key.scope] += len(commits)
	}

	// For each window, select top (maxCategories - 1) scopes to leave room for "Other"
	result := make(map[time.Time][]ScopeType)
	for windowStart, scopeCounts := range windowScopeCounts {
		type scopeCount struct {
			scope ScopeType
			count int
		}

		var scopes []scopeCount
		for scope, count := range scopeCounts {
			scopes = append(scopes, scopeCount{scope, count})
		}

		sort.Slice(scopes, func(i, j int) bool {
			return scopes[i].count > scopes[j].count
		})

		// Leave room for "Other" by selecting maxCategories - 1
		limit := len(scopes)
		if maxCategories > 1 && limit >= maxCategories {
			limit = maxCategories - 1
		}

		topScopes := make([]ScopeType, limit)
		for i := 0; i < limit; i++ {
			topScopes[i] = scopes[i].scope
		}
		result[windowStart] = topScopes
	}

	return result
}

func AggregateCommitGroup(commits CommitAnalyses) Count {
	count := Count{
		Scopes:      make(map[ScopeType]int),
		CommitTypes: make(map[CommitType]int),
		Tech:        make(map[ScopeTechnology]int),
	}

	uniqueFiles := make(map[string]struct{})

	for _, commit := range commits {
		count.Commits++

		if commit.Scope != ScopeTypeUnknown {
			count.Scopes[commit.Scope]++
		}

		if commit.CommitType != CommitTypeUnknown {
			count.CommitTypes[commit.CommitType]++
		}

		for _, tech := range commit.Tech {
			count.Tech[tech]++
		}

		for _, change := range commit.Changes {
			count.Adds += change.Adds
			count.Dels += change.Dels
			uniqueFiles[change.File] = struct{}{}
		}
	}

	count.Files = len(uniqueFiles)

	return count
}

func formatTimeWindow(window *TimeWindow, windowType GroupByWindow) string {
	start := window.Start
	switch windowType {
	case GroupByDay:
		return start.Format("Jan 2, 2006")
	case GroupByWeek:
		return fmt.Sprintf("Week of %s", start.Format("Jan 2, 2006"))
	case GroupByMonth:
		return start.Format("January 2006")
	default:
		return start.Format("Jan 2, 2006")
	}
}

func GenerateFallbackDescription(scope ScopeType, commits CommitAnalyses) (string, string) {
	name := fmt.Sprintf("%s changes", scope)

	typeCounts := make(map[CommitType]int)
	for _, commit := range commits {
		if commit.CommitType != CommitTypeUnknown {
			typeCounts[commit.CommitType]++
		}
	}

	typeList := lo.Entries(typeCounts)
	sort.Slice(typeList, func(i, j int) bool {
		return typeList[i].Value > typeList[j].Value
	})

	desc := fmt.Sprintf("%d commits", len(commits))
	if len(typeList) > 0 {
		desc += ": "
		for i, entry := range typeList {
			if i > 0 {
				desc += ", "
			}
			desc += fmt.Sprintf("%d %s", entry.Value, entry.Key)
			if i >= 2 {
				break
			}
		}
	}

	return name, desc
}

func (gs GitSummary) Pretty() api.Text {
	t := clicky.Text("").Append(gs.Group, "font-medium").Append(" - ").Append(gs.TimeWindow).NewLine()

	if len(gs.Repositories) > 0 {
		t = t.Append("Repositories: ", "text-muted").Append(strings.Join(gs.Repositories, ", ")).NewLine()
	}

	t = t.Append(gs.Description).NewLine().
		Append(gs.Commits).HR()

	return t
}

type GitSummaries []GitSummary

func (gs GitSummaries) Pretty() api.Text {
	list := api.TextList{}
	for _, summary := range gs {
		list = append(list, summary.Pretty())
	}
	return list.Join()

}

func Summarize(commits CommitAnalyses, options SummaryOptions) (GitSummaries, error) {
	clicky.Infof("Generating git summary with window=%s, maxCategories=%d", options.Window, options.MaxCategories)
	if len(commits) == 0 {
		return GitSummaries{}, nil
	}

	windows := CalculateTimeWindows(commits.From(), commits.To(), options.Window)

	logger.Debugf("Using time windows: %v", windows)

	// Group commits by (window, scope)
	grouped := make(map[windowScopeKey]CommitAnalyses)

	for _, commit := range commits {
		window := GetWindowForCommit(commit, windows)
		if window == nil {
			continue
		}

		// Treat unknown scopes as "Other" to ensure all commits are included
		scope := commit.Scope
		if scope == ScopeTypeUnknown {
			scope = ScopeTypeOther
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
	otherCommits := make(map[time.Time]CommitAnalyses)

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
			scope:        ScopeTypeOther,
			commits:      commits,
			repositories: repositories,
			isOther:      true,
		})
	}

	// Process groups in parallel if AI is enabled, otherwise sequentially
	var summaries []GitSummary

	if options.Agent != nil && options.Context != nil {
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

				aiName, aiDesc, err := GenerateGroupSummary(options.Context, group.scope, windowLabel, group.commits, options.Agent)
				var name, desc string
				if err != nil {
					logger.Warnf("AI summary generation failed, using fallback: %v", err)
					name, desc = GenerateFallbackDescription(group.scope, group.commits)
				} else {
					name, desc = aiName, aiDesc
					logger.Debugf("AI generated summary for %s: %s", group.scope, name)
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
		for item := range batch.Run() {
			if item.Error != nil {
				logger.Warnf("Batch item failed: %v", item.Error)
				continue
			}
			summaries = append(summaries, item.Value)
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
