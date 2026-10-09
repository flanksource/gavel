package prwatch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/ttyrender"
	"github.com/flanksource/gavel/pr/model"
)

type WatchOptions struct {
	Context       context.Context
	FetchSnapshot func(context.Context, WatchOptions) (*PRWatchResult, error)
	github.Options
	PRNumber int
	Interval time.Duration
	Follow   bool
	// FailFast stops --follow at the first definitive failure instead of waiting
	// for every remaining check. Off by default: the default contract of --follow
	// is a complete picture of the run.
	FailFast bool
	Logs     bool // fetch failing job log tails (extra API quota)
	TailLogs int
	Comments []string
	Actions  []string
}

func Run(opts WatchOptions) (*PRWatchResult, int) {
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	fetch := opts.FetchSnapshot
	if fetch == nil {
		fetch = fetchGitHubSnapshot
	}
	logger.Debugf("starting watch (pr=%d, interval=%s, follow=%t)", opts.PRNumber, opts.Interval, opts.Follow)

	var (
		render ttyrender.State
		isTTY  = ttyrender.IsTerminal(os.Stderr)
	)

	for {
		if opts.Context.Err() != nil {
			return nil, 1
		}
		result, err := fetch(opts.Context, opts)
		if err != nil {
			if !opts.Follow || opts.FetchSnapshot != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return nil, 1
			}
			logger.Errorf("fetch failed: %v, retrying in %s", err, opts.Interval)
			if !waitForPoll(opts) {
				return nil, 1
			}
			continue
		}

		if result == nil || result.PR == nil {
			fmt.Fprintln(os.Stderr, "Error: PR provider returned no PR snapshot")
			return nil, 1
		}
		pr, runs := result.PR, result.Runs
		filters := newResultFilters(opts.Comments, opts.Actions)

		preChecks := len(pr.StatusCheckRollup)
		preRuns := len(runs)
		preComments := len(result.Comments)
		var selectorOptions []string
		if filters.hasActionFilters() {
			selectorOptions = actionSelectorOptions(pr, runs)
		}

		filters.apply(result)

		// Fail loudly when a selector was given but matched nothing, rather than
		// printing "No checks found" and exiting 0 — a silent empty masks a
		// mistyped selector (and would false-green a verification fixture built
		// from it). Under --follow it was worse than silent: an empty filtered
		// set satisfied the completion gate on the very first poll, so a typo —
		// or simply a push whose checks GitHub had not registered yet — returned
		// 0 before CI had run a single step. Both no-match predicates are scoped
		// to a selector that pruned a non-empty set down to nothing, so a PR
		// whose checks or comments have not appeared yet still keeps polling.
		if filters.noActionMatch(preChecks, preRuns, result) {
			fmt.Fprintf(os.Stderr, "Error: --actions %s matched no checks or workflows on PR #%d.\nAvailable selectors: %s\n",
				strings.Join(opts.Actions, ","), opts.PRNumber, strings.Join(selectorOptions, ", "))
			return nil, 1
		}
		if filters.noCommentMatch(preComments, result) {
			fmt.Fprintf(os.Stderr, "Error: --comments %s matched no comments on PR #%d.\n",
				strings.Join(opts.Comments, ","), opts.PRNumber)
			return nil, 1
		}

		if !opts.Follow {
			return result, statusExitCode(result)
		}

		if followDone(filters, result, opts.FailFast) {
			// The caller prints the completed report to stdout. Painting the
			// final frame to stderr as well would show it twice to anyone
			// merging the two streams; on a TTY it also has to be erased so
			// the stdout copy does not stack under the last live frame.
			if isTTY {
				if err := render.Clear(os.Stderr); err != nil {
					logger.Warnf("render: %v", err)
				}
			}
			return result, statusExitCode(result)
		}

		if isTTY {
			frame := result.Pretty().ANSI()
			if !strings.HasSuffix(frame, "\n") {
				frame += "\n"
			}
			frame += fmt.Sprintf("Polling in %s...\n\n", opts.Interval)
			if err := render.Write(os.Stderr, frame); err != nil {
				logger.Warnf("render: %v", err)
			}
		} else {
			// A reader that cannot redraw gets a heartbeat, not the frame again.
			// The frame is printed once at the end by the caller.
			fmt.Fprintln(os.Stderr, followProgressLine(result, opts.Interval))
		}

		if !waitForPoll(opts) {
			return result, 1
		}
	}
}

func waitForPoll(opts WatchOptions) bool {
	timer := time.NewTimer(opts.Interval)
	defer timer.Stop()
	select {
	case <-opts.Context.Done():
		return false
	case <-timer.C:
		return true
	}
}

func fetchGitHubSnapshot(ctx context.Context, opts WatchOptions) (*PRWatchResult, error) {
	pr, err := github.FetchPR(opts.Options, opts.PRNumber)
	if err != nil {
		return nil, err
	}
	allComments := append(append([]model.PRComment{}, pr.Comments...), pr.ReviewThreads...)
	artifacts := github.FindGavelArtifacts(allComments)
	gavelResultsCh := make(chan []*GavelResultsSummary, 1)
	go func() { gavelResultsCh <- FetchGavelArtifacts(opts.Options, artifacts) }()
	runs := fetchRuns(opts, pr)
	gavelResults := <-gavelResultsCh
	annotateReproduceCommands(gavelResults, opts.Repo, pr.Number)
	comments := removeRenderedArtifactComments(MergeAndFilter(pr.Comments, pr.ReviewThreads), gavelResults)
	return &PRWatchResult{PR: pr, Runs: runs, Conflicts: github.DetectMergeConflicts(opts.Options, pr), GavelResults: gavelResults, Comments: comments}, nil
}

// followDone reports whether --follow has seen everything it was asked to wait
// for. failFast short-circuits the completion gate once a failure can no longer
// change, so a test job that goes red at two minutes is not held behind a scan
// that runs for seven.
func followDone(filters resultFilters, result *PRWatchResult, failFast bool) bool {
	if failFast && result.HasTerminalFailure() {
		return true
	}
	if result != nil && result.PR != nil && result.PR.Provider == "azuredevops" {
		if result.PR.State != "OPEN" {
			return true
		}
		r := result.PR.MergeReadiness
		if r == nil {
			return true
		}
		if r.WaitingForMerge {
			return false
		}
		if !filters.hasActionFilters() && r.WaitingForBuilds {
			return false
		}
		if len(result.PR.StatusCheckRollup) == 0 {
			return !filters.hasActionFilters()
		}
		return result.PR.StatusCheckRollup.AllComplete()
	}
	return filters.isComplete(result)
}

// statusExitCode weighs every failure signal the status view renders, not just
// the head commit's rollup. A repo that reports gavel results through artifact
// comments rather than a required check has no failing rollup context at all,
// so a rollup-only exit code false-greens the whole run.
//
// Called after filters.apply, so --actions scopes the exit code to the checks,
// runs, and gavel artifacts the user asked to see. A merge conflict is the one
// signal filters never scope away: it blocks the merge no matter which checks
// the caller asked about.
func statusExitCode(result *PRWatchResult) int {
	if result == nil {
		return 0
	}
	if result.PR != nil && result.PR.Provider == "azuredevops" {
		if result.PR.State == "OPEN" && (result.PR.MergeReadiness == nil || result.PR.MergeReadiness.State != "ready") {
			return 1
		}
		for _, check := range result.PR.StatusCheckRollup {
			if check.Conclusion == "CANCELLED" {
				return 1
			}
		}
		for _, run := range result.Runs {
			if run == nil {
				continue
			}
			if len(run.Jobs) == 0 && (model.IsFailureConclusion(run.Conclusion) || run.Conclusion == "CANCELLED") {
				return 1
			}
			for _, job := range run.Jobs {
				if model.IsFailureConclusion(job.Conclusion) || job.Conclusion == "CANCELLED" {
					return 1
				}
			}
		}
	}
	if result.HasMergeConflict() {
		return 1
	}
	if result.PR != nil && result.PR.StatusCheckRollup.HasFailure() {
		return 1
	}
	for _, summary := range result.GavelResults {
		if summary != nil && summary.HasFailure() {
			return 1
		}
	}
	if result.HasFailedRun() {
		return 1
	}
	return 0
}

func fetchRuns(opts WatchOptions, pr *model.PRInfo) map[int64]*model.WorkflowRun {
	runs := make(map[int64]*model.WorkflowRun)
	seen := make(map[int64]bool)

	for _, check := range pr.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err != nil || seen[runID] {
			continue
		}
		seen[runID] = true

		// FetchRunJobs short-circuits via the persistent github cache when
		// the run is already completed — and atomically attaches failed-job
		// logs before caching when opts.Logs is set, so a previously
		// log-less cache entry can't suppress --logs.
		run, err := github.FetchRunJobs(opts.Options, runID, github.RunLogOptions{
			FetchLogs: opts.Logs,
			TailLines: opts.TailLogs,
		})
		if err != nil {
			logger.Warnf("failed to fetch run %d: %v", runID, err)
			continue
		}

		if model.RunHasFailedJob(run) || newResultFilters(nil, opts.Actions).hasActionFilters() {
			if _, err := github.FetchWorkflowDefinition(opts.Options, run); err != nil {
				logger.Warnf("failed to fetch workflow definition for run %d: %v", runID, err)
			}
		}
		runs[runID] = run
	}
	return runs
}

// MergeAndFilter combines comments with thread state, extracts nitpick sub-comments, and filters noise.
func MergeAndFilter(comments []model.PRComment, threads []model.PRComment) []model.PRComment {
	comments = mergeThreadState(comments, threads)
	comments = annotateBots(comments)
	comments = extractNitpicks(comments)
	return filterActionableComments(comments)
}

func mergeThreadState(comments []model.PRComment, threads []model.PRComment) []model.PRComment {
	threadByID := make(map[int64]model.PRComment, len(threads))
	for _, t := range threads {
		threadByID[t.ID] = t
	}
	for i, c := range comments {
		if t, ok := threadByID[c.ID]; ok {
			comments[i].IsResolved = t.IsResolved
			comments[i].IsOutdated = t.IsOutdated
			if comments[i].Path == "" {
				comments[i].Path = t.Path
			}
			if comments[i].Line == 0 {
				comments[i].Line = t.Line
			}
			if comments[i].Severity == "" {
				comments[i].Severity = parseSeverityFromBadge(c.Body)
			}
		}
	}
	return comments
}

func extractNitpicks(comments []model.PRComment) []model.PRComment {
	var result []model.PRComment
	for _, c := range comments {
		result = append(result, c)
		if c.BotType == "coderabbit" {
			result = append(result, parseNitpickComments(c)...)
		}
	}
	return result
}

func filterActionableComments(comments []model.PRComment) []model.PRComment {
	var result []model.PRComment
	for _, c := range comments {
		if c.Severity != "" || c.Path != "" {
			result = append(result, c)
			continue
		}
		body := strings.TrimSpace(c.Body)
		if isNoiseComment(body) {
			continue
		}
		result = append(result, c)
	}
	return result
}

func isNoiseComment(body string) bool {
	if strings.HasPrefix(body, "> [!") {
		return true
	}
	if strings.HasPrefix(body, "**Actionable comments posted:") {
		return true
	}
	if strings.HasPrefix(body, "Actionable comments posted:") {
		return true
	}
	return reportsPullRequestClosed(body)
}

// reportsPullRequestClosed matches bot comments whose entire content is "this
// PR is closed, so I did nothing". They carry no action, but they arrive as
// ordinary top-level comments and would otherwise dominate the comment count
// on any merged PR.
func reportsPullRequestClosed(body string) bool {
	// CodeRabbit posts its closed-PR skip behind a failure marker; a review
	// that failed for any other reason stays visible.
	if strings.Contains(body, "auto-generated comment: failure by coderabbit.ai") &&
		strings.Contains(body, "The pull request is closed.") {
		return true
	}
	// rossjrw/pr-preview-action tear-down notice.
	return strings.Contains(body, "Preview removed because the pull request was closed.")
}
