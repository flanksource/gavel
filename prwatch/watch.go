package prwatch

import (
	"fmt"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
)

type WatchOptions struct {
	github.Options
	PRNumber int
	Interval time.Duration
	Follow   bool
	TailLogs int
}

func Run(opts WatchOptions) (*PRWatchResult, int) {
	logger.Debugf("starting watch (pr=%d, interval=%s, follow=%t)", opts.PRNumber, opts.Interval, opts.Follow)
	cachedRuns := make(map[int64]*github.WorkflowRun)

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

		runs := fetchRuns(opts, pr, cachedRuns)
		maps.Copy(cachedRuns, runs)

		comments, err := github.FetchPRComments(opts.Options, pr.Number)
		if err != nil {
			logger.Warnf("failed to fetch PR comments: %v", err)
		}

		threads, err := github.FetchReviewThreads(opts.Options, pr.Number)
		if err != nil {
			logger.Warnf("failed to fetch review threads: %v", err)
		}
		comments = mergeThreadState(comments, threads)
		comments = extractNitpicks(comments)
		comments = filterActionableComments(comments)

		result := &PRWatchResult{PR: pr, Runs: runs, Comments: comments}

		if !opts.Follow {
			if pr.StatusCheckRollup.HasFailure() {
				return result, 1
			}
			return result, 0
		}

		fmt.Fprint(os.Stderr, result.Pretty().ANSI()+"\n")
		if pr.StatusCheckRollup.AllComplete() {
			if pr.StatusCheckRollup.HasFailure() {
				return result, 1
			}
			return result, 0
		}

		fmt.Fprintf(os.Stderr, "Polling in %s...\n\n", opts.Interval)
		time.Sleep(opts.Interval)
	}
}

func fetchRuns(opts WatchOptions, pr *github.PRInfo, cached map[int64]*github.WorkflowRun) map[int64]*github.WorkflowRun {
	runs := make(map[int64]*github.WorkflowRun)
	seen := make(map[int64]bool)

	for _, check := range pr.StatusCheckRollup {
		runID, err := github.ExtractRunID(check.DetailsURL)
		if err != nil || seen[runID] {
			continue
		}
		seen[runID] = true

		if existing, ok := cached[runID]; ok && existing.Status == "completed" {
			logger.Tracef("run %d already completed, using cache", runID)
			runs[runID] = existing
			continue
		}

		run, err := github.FetchRunJobs(opts.Options, runID)
		if err != nil {
			logger.Warnf("failed to fetch run %d: %v", runID, err)
			continue
		}

		if run.Conclusion == "failure" {
			github.FetchAndAttachLogs(opts.Options, run, opts.TailLogs)
			if _, err := github.FetchWorkflowDefinition(opts.Options, run); err != nil {
				logger.Warnf("failed to fetch workflow definition for run %d: %v", runID, err)
			}
		}
		runs[runID] = run
	}
	return runs
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
		if c.Author == "coderabbitai[bot]" {
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
