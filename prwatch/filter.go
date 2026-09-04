package prwatch

import (
	"path/filepath"
	"sort"
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
	// matchedRunIDs holds only whole-run matches (a workflow-level selector).
	// Job-level matches keep the run pruned to the matching jobs, but their
	// checks are matched individually by name in matchStatusCheck so a sibling
	// job's check does not leak in.
	matchedRunIDs := map[int64]bool{}
	filteredRuns := make(map[int64]*github.WorkflowRun, len(result.Runs))

	for key, run := range result.Runs {
		if run == nil {
			continue
		}
		matched, negated := matchTargets(workflowRunMatchTargets(run), f.actions)
		if negated {
			continue
		}
		if matched {
			filteredRuns[key] = run
			matchedRunIDs[key] = true
			matchedRunIDs[run.DatabaseID] = true
			continue
		}
		if jobs := matchingJobs(run.Jobs, f.actions); len(jobs) > 0 {
			pruned := *run
			pruned.Jobs = jobs
			filteredRuns[key] = &pruned
		}
	}
	result.Runs = filteredRuns
	result.GavelResults = filterGavelResultsByRun(result.GavelResults, filteredRuns)

	if result.PR == nil {
		return
	}
	checks := make(github.StatusChecks, 0, len(result.PR.StatusCheckRollup))
	for _, check := range result.PR.StatusCheckRollup {
		if f.matchStatusCheck(check, matchedRunIDs) {
			checks = append(checks, check)
		}
	}
	result.PR.StatusCheckRollup = checks
}

func filterGavelResultsByRun(results []*GavelResultsSummary, runs map[int64]*github.WorkflowRun) []*GavelResultsSummary {
	filtered := make([]*GavelResultsSummary, 0, len(results))
	for _, result := range results {
		if result == nil {
			continue
		}
		if _, ok := runs[result.RunID]; ok {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// matchingJobs returns the jobs whose name matches the action patterns.
func matchingJobs(jobs []github.Job, patterns []string) []github.Job {
	var out []github.Job
	for _, job := range jobs {
		if job.Name != "" && matchAnyTarget([]string{job.Name}, patterns) {
			out = append(out, job)
		}
	}
	return out
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

// matchStatusCheck reports whether a check survives the action filter. A check
// whose run matched a workflow-level selector is always kept; otherwise the
// check matches on its own run ID, workflow name, or job/check name — so a
// job/check name shown in the tree (e.g. "lint") is a valid selector, not just
// the workflow it belongs to.
func (f resultFilters) matchStatusCheck(check github.StatusCheck, matchedRunIDs map[int64]bool) bool {
	var values []string
	if runID, err := github.ExtractRunID(check.DetailsURL); err == nil {
		if matchedRunIDs[runID] {
			return true
		}
		values = append(values, strconv.FormatInt(runID, 10))
	}
	if check.WorkflowName != "" {
		values = append(values, check.WorkflowName)
	}
	if check.Name != "" {
		values = append(values, check.Name)
	}
	return matchAnyTarget(values, f.actions)
}

// matchTargets reports whether any target matches the patterns and whether a
// negation pattern excluded a target (which takes precedence over a match).
func matchTargets(targets, patterns []string) (matched, negated bool) {
	if len(patterns) == 0 {
		return true, false
	}
	if len(targets) == 0 {
		return false, false
	}
	return collections.MatchAny(targets, patterns...)
}

func matchAnyTarget(targets, patterns []string) bool {
	matched, negated := matchTargets(targets, patterns)
	return matched && !negated
}

// noActionMatch reports whether action filters pruned a non-empty set of
// checks/runs down to nothing — a mismatched selector rather than a PR that
// simply has no checks yet.
func (f resultFilters) noActionMatch(preChecks, preRuns int, result *PRWatchResult) bool {
	if !f.hasActionFilters() || (preChecks == 0 && preRuns == 0) {
		return false
	}
	if result == nil {
		return true
	}
	checksLeft := result.PR != nil && len(result.PR.StatusCheckRollup) > 0
	return len(result.Runs) == 0 && !checksLeft
}

// actionSelectorOptions lists the selectors the user could have passed to
// --actions for this PR: workflow names, workflow YAML basenames, and job/check
// names. Used to make a no-match failure actionable.
func actionSelectorOptions(pr *github.PRInfo, runs map[int64]*github.WorkflowRun) []string {
	seen := map[string]bool{}
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, run := range runs {
		if run == nil {
			continue
		}
		add(run.Name)
		if run.WorkflowPath != "" {
			add(filepath.Base(run.WorkflowPath))
		}
		for _, job := range run.Jobs {
			add(job.Name)
		}
	}
	if pr != nil {
		for _, check := range pr.StatusCheckRollup {
			add(check.WorkflowName)
			add(check.Name)
		}
	}
	sort.Strings(out)
	return out
}
