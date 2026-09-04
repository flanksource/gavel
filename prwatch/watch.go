package prwatch

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/ttyrender"
)

type WatchOptions struct {
	github.Options
	PRNumber int
	Interval time.Duration
	Follow   bool
	Logs     bool // fetch failing job log tails (extra API quota)
	TailLogs int
	Comments []string
	Actions  []string
}

func Run(opts WatchOptions) (*PRWatchResult, int) {
	logger.Debugf("starting watch (pr=%d, interval=%s, follow=%t)", opts.PRNumber, opts.Interval, opts.Follow)

	var (
		render ttyrender.State
		isTTY  = ttyrender.IsTerminal(os.Stderr)
	)

	for {
		pr, err := github.FetchPR(opts.Options, opts.PRNumber)
		if err != nil {
			if !opts.Follow {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return nil, 1
			}
			logger.Errorf("fetch failed: %v, retrying in %s", err, opts.Interval)
			time.Sleep(opts.Interval)
			continue
		}

		allComments := append(append([]github.PRComment{}, pr.Comments...), pr.ReviewThreads...)
		artifacts := github.FindGavelArtifacts(allComments)
		gavelResultsCh := make(chan []*GavelResultsSummary, 1)
		go func() {
			gavelResultsCh <- FetchGavelArtifacts(opts.Options, artifacts)
		}()

		// The persistent github cache short-circuits already-completed runs.
		runs := fetchRuns(opts, pr)
		gavelResults := <-gavelResultsCh
		annotateReproduceCommands(gavelResults, opts.Repo, pr.Number)
		comments := MergeAndFilter(pr.Comments, pr.ReviewThreads)
		comments = removeRenderedArtifactComments(comments, gavelResults)

		result := &PRWatchResult{PR: pr, Runs: runs, GavelResults: gavelResults, Comments: comments}
		filters := newResultFilters(opts.Comments, opts.Actions)

		preChecks := len(pr.StatusCheckRollup)
		preRuns := len(runs)
		var selectorOptions []string
		if filters.hasActionFilters() {
			selectorOptions = actionSelectorOptions(pr, runs)
		}

		filters.apply(result)

		if !opts.Follow {
			// Fail loudly when --actions was given but matched nothing, rather
			// than printing "No checks found" and exiting 0 — a silent empty
			// masks a mistyped selector (and would false-green a verification
			// fixture built from it).
			if filters.noActionMatch(preChecks, preRuns, result) {
				fmt.Fprintf(os.Stderr, "Error: --actions %s matched no checks or workflows on PR #%d.\nAvailable selectors: %s\n",
					strings.Join(opts.Actions, ","), opts.PRNumber, strings.Join(selectorOptions, ", "))
				return nil, 1
			}
			return result, statusExitCode(result)
		}

		done := filters.actionFilteredNoChecks(result) || result.PR.StatusCheckRollup.AllComplete()
		if done {
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

		frame := result.Pretty().ANSI()
		if !strings.HasSuffix(frame, "\n") {
			frame += "\n"
		}
		frame += fmt.Sprintf("Polling in %s...\n\n", opts.Interval)

		if isTTY {
			if err := render.Write(os.Stderr, frame); err != nil {
				logger.Warnf("render: %v", err)
			}
		} else {
			fmt.Fprint(os.Stderr, frame)
		}

		time.Sleep(opts.Interval)
	}
}

// statusExitCode weighs every failure signal the status view renders, not just
// the head commit's rollup. A repo that reports gavel results through artifact
// comments rather than a required check has no failing rollup context at all,
// so a rollup-only exit code false-greens the whole run.
//
// Called after filters.apply, so --actions scopes the exit code to the checks,
// runs, and gavel artifacts the user asked to see.
func statusExitCode(result *PRWatchResult) int {
	if result == nil {
		return 0
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

func fetchRuns(opts WatchOptions, pr *github.PRInfo) map[int64]*github.WorkflowRun {
	runs := make(map[int64]*github.WorkflowRun)
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

		if github.RunHasFailedJob(run) || newResultFilters(nil, opts.Actions).hasActionFilters() {
			if _, err := github.FetchWorkflowDefinition(opts.Options, run); err != nil {
				logger.Warnf("failed to fetch workflow definition for run %d: %v", runID, err)
			}
		}
		runs[runID] = run
	}
	return runs
}

// MergeAndFilter combines comments with thread state, extracts nitpick sub-comments, and filters noise.
func MergeAndFilter(comments []github.PRComment, threads []github.PRComment) []github.PRComment {
	comments = mergeThreadState(comments, threads)
	comments = annotateBots(comments)
	comments = extractNitpicks(comments)
	return filterActionableComments(comments)
}

func mergeThreadState(comments []github.PRComment, threads []github.PRComment) []github.PRComment {
	threadByID := make(map[int64]github.PRComment, len(threads))
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

func extractNitpicks(comments []github.PRComment) []github.PRComment {
	var result []github.PRComment
	for _, c := range comments {
		result = append(result, c)
		if c.BotType == "coderabbit" {
			result = append(result, parseNitpickComments(c)...)
		}
	}
	return result
}

func filterActionableComments(comments []github.PRComment) []github.PRComment {
	var result []github.PRComment
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
