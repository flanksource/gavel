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

		// The persistent github cache short-circuits already-completed runs,
		// so a per-iteration in-memory map is no longer needed.
		runs := fetchRuns(opts, pr)

		// Comments and review threads arrive with the PR in a single GraphQL request.
		comments := MergeAndFilter(pr.Comments, pr.ReviewThreads)

		result := &PRWatchResult{PR: pr, Runs: runs, Comments: comments}

		if !opts.Follow {
			if pr.StatusCheckRollup.HasFailure() {
				return result, 1
			}
			return result, 0
		}

		done := pr.StatusCheckRollup.AllComplete()
		frame := result.Pretty().ANSI()
		if !strings.HasSuffix(frame, "\n") {
			frame += "\n"
		}
		if !done {
			frame += fmt.Sprintf("Polling in %s...\n\n", opts.Interval)
		}

		if isTTY {
			if err := render.Write(os.Stderr, frame); err != nil {
				logger.Warnf("render: %v", err)
			}
		} else {
			fmt.Fprint(os.Stderr, frame)
		}

		if done {
			if pr.StatusCheckRollup.HasFailure() {
				return result, 1
			}
			return result, 0
		}

		time.Sleep(opts.Interval)
	}
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

		if github.RunHasFailedJob(run) {
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
	return false
}
