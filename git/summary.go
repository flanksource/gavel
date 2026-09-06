package git

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"

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
	MaxWorkers    int                                    `json:"-"`
	Prompt        verify.PromptSpec                      `json:"-"`
	PromptOptions verify.PromptResolveOptions            `json:"-"`
	Saved         captainconfig.Config                   `json:"-"`
	AgentFactory  func(ai.AgentConfig) (ai.Agent, error) `json:"-"`
}

type windowScopeKey struct {
	windowStart time.Time
	scope       models.ScopeType
}

type summaryGroup struct {
	windowStart  time.Time
	window       *TimeWindow
	scope        models.ScopeType
	commits      models.CommitAnalyses
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

	Scopes       []models.ScopeType       `json:"scopes,omitempty"`
	Tech         []models.ScopeTechnology `json:"tech,omitempty"`
	Repositories []string                 `json:"repositories,omitempty"`
	Commits      Count                    `json:"commits,omitempty"`
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

func GetWindowForCommit(commit models.CommitAnalysis, windows []TimeWindow) *TimeWindow {
	commitTime := commit.Author.Date
	for i := range windows {
		if (commitTime.Equal(windows[i].Start) || commitTime.After(windows[i].Start)) &&
			(commitTime.Equal(windows[i].End) || commitTime.Before(windows[i].End)) {
			return &windows[i]
		}
	}
	return nil
}

func SelectTopScopes(commits models.CommitAnalyses, maxCategories int) []models.ScopeType {
	scopeCounts := make(map[models.ScopeType]int)
	for _, commit := range commits {
		if commit.Scope != models.ScopeTypeUnknown {
			scopeCounts[commit.Scope]++
		}
	}

	type scopeCount struct {
		scope models.ScopeType
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

	result := make([]models.ScopeType, limit)
	for i := 0; i < limit; i++ {
		result[i] = scopes[i].scope
	}

	return result
}

// SelectTopScopesPerWindow selects the top N-1 scopes per time window to leave room for "Other"
// Returns a map of window start times to the top scope types for that window
func SelectTopScopesPerWindow(grouped map[windowScopeKey]models.CommitAnalyses, maxCategories int) map[time.Time][]models.ScopeType {
	if maxCategories <= 0 {
		return make(map[time.Time][]models.ScopeType)
	}

	// Group commits by window and count per scope
	windowScopeCounts := make(map[time.Time]map[models.ScopeType]int)
	for key, commits := range grouped {
		if _, exists := windowScopeCounts[key.windowStart]; !exists {
			windowScopeCounts[key.windowStart] = make(map[models.ScopeType]int)
		}
		windowScopeCounts[key.windowStart][key.scope] += len(commits)
	}

	// For each window, select top (maxCategories - 1) scopes to leave room for "Other"
	result := make(map[time.Time][]models.ScopeType)
	for windowStart, scopeCounts := range windowScopeCounts {
		type scopeCount struct {
			scope models.ScopeType
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

		topScopes := make([]models.ScopeType, limit)
		for i := 0; i < limit; i++ {
			topScopes[i] = scopes[i].scope
		}
		result[windowStart] = topScopes
	}

	return result
}

func AggregateCommitGroup(commits models.CommitAnalyses) Count {
	count := Count{
		Scopes:      make(map[models.ScopeType]int),
		CommitTypes: make(map[models.CommitType]int),
		Tech:        make(map[models.ScopeTechnology]int),
	}

	uniqueFiles := make(map[string]struct{})

	for _, commit := range commits {
		count.Commits++

		if commit.Scope != models.ScopeTypeUnknown {
			count.Scopes[commit.Scope]++
		}

		if commit.CommitType != models.CommitTypeUnknown {
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

func GenerateFallbackDescription(scope models.ScopeType, commits models.CommitAnalyses) (string, string) {
	name := fmt.Sprintf("%s changes", scope)

	typeCounts := make(map[models.CommitType]int)
	for _, commit := range commits {
		if commit.CommitType != models.CommitTypeUnknown {
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
