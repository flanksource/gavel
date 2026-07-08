package prwatch

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/commons/collections"
	"github.com/flanksource/gavel/github"
)

type resultFilters struct {
	comments []string
	actions  []string
}

func newResultFilters(comments, actions []string) resultFilters {
	return resultFilters{
		comments: normalizeMatchFilterPatterns(comments),
		actions:  normalizeMatchFilterPatterns(actions),
	}
}

func normalizeMatchFilterPatterns(patterns []string) []string {
	joined := strings.TrimSpace(strings.Join(patterns, ","))
	if joined == "" {
		return nil
	}
	if strings.HasPrefix(joined, "[") && strings.HasSuffix(joined, "]") && len(joined) >= 2 {
		joined = strings.TrimSpace(joined[1 : len(joined)-1])
	}

	var out []string
	for _, part := range strings.Split(joined, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func (f resultFilters) hasActionFilters() bool {
	return len(f.actions) > 0
}

func (f resultFilters) actionFilteredNoChecks(result *PRWatchResult) bool {
	return f.hasActionFilters() && result != nil && result.PR != nil && len(result.PR.StatusCheckRollup) == 0
}

func (f resultFilters) apply(result *PRWatchResult) {
	if result == nil {
		return
	}
	if len(f.comments) > 0 {
		result.Comments = f.filterComments(result.Comments)
	}
	if len(f.actions) > 0 {
		f.filterActions(result)
	}
}

func (f resultFilters) filterComments(comments []github.PRComment) []github.PRComment {
	filtered := make([]github.PRComment, 0, len(comments))
	for _, comment := range comments {
		if matchAnyTarget(commentMatchTargets(comment), f.comments) {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

func commentMatchTargets(comment github.PRComment) []string {
	var targets []string
	if comment.ID != 0 {
		targets = append(targets, strconv.FormatInt(comment.ID, 10))
	}
	if comment.Author != "" {
		targets = appendAuthorTargets(targets, comment.Author)
	}
	if comment.BotType != "" {
		targets = appendAuthorTargets(targets, comment.BotType)
	}
	return targets
}

func appendAuthorTargets(targets []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return targets
	}
	targets = append(targets, value, "@"+value)
	if alias := strings.TrimSuffix(value, "[bot]"); alias != value && alias != "" {
		targets = append(targets, alias, "@"+alias)
	}
	return targets
}

func (f resultFilters) filterActions(result *PRWatchResult) {
	matchedRunIDs := map[int64]bool{}
	seenRunIDs := map[int64]bool{}
	filteredRuns := make(map[int64]*github.WorkflowRun, len(result.Runs))

	for key, run := range result.Runs {
		if run == nil {
			continue
		}
		seenRunIDs[key] = true
		seenRunIDs[run.DatabaseID] = true
		if matchAnyTarget(workflowRunMatchTargets(run), f.actions) {
			filteredRuns[key] = run
			matchedRunIDs[key] = true
			matchedRunIDs[run.DatabaseID] = true
		}
	}
	result.Runs = filteredRuns

	if result.PR == nil {
		return
	}
	checks := make(github.StatusChecks, 0, len(result.PR.StatusCheckRollup))
	for _, check := range result.PR.StatusCheckRollup {
		if f.matchStatusCheck(check, matchedRunIDs, seenRunIDs) {
			checks = append(checks, check)
		}
	}
	result.PR.StatusCheckRollup = checks
}

func workflowRunMatchTargets(run *github.WorkflowRun) []string {
	var targets []string
	if run.DatabaseID != 0 {
		targets = append(targets, strconv.FormatInt(run.DatabaseID, 10))
	}
	if run.WorkflowID != 0 {
		targets = append(targets, strconv.FormatInt(run.WorkflowID, 10))
	}
	if run.WorkflowPath != "" {
		targets = append(targets, run.WorkflowPath, filepath.Base(run.WorkflowPath))
	}
	if run.Name != "" {
		targets = append(targets, run.Name)
	}
	return targets
}

func (f resultFilters) matchStatusCheck(check github.StatusCheck, matchedRunIDs, seenRunIDs map[int64]bool) bool {
	if runID, err := github.ExtractRunID(check.DetailsURL); err == nil {
		if matchedRunIDs[runID] {
			return true
		}
		if seenRunIDs[runID] {
			return false
		}

		values := []string{strconv.FormatInt(runID, 10)}
		if check.WorkflowName != "" {
			values = append(values, check.WorkflowName)
		}
		return matchAnyTarget(values, f.actions)
	}

	var values []string
	if check.WorkflowName != "" {
		values = append(values, check.WorkflowName)
	}
	if check.Name != "" {
		values = append(values, check.Name)
	}
	return matchAnyTarget(values, f.actions)
}

func matchAnyTarget(targets []string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	if len(targets) == 0 {
		return false
	}
	matched, negated := collections.MatchAny(targets, patterns...)
	return matched && !negated
}
