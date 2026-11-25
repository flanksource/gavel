package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/prwatch"
)

// maxRunFetches bounds how many of a PR's workflow runs are fetched at once.
const maxRunFetches = 8

type prDetail struct {
	PR   *model.PRInfo                `json:"pr,omitempty"`
	Runs map[int64]*model.WorkflowRun `json:"runs,omitempty"`
	// RunsError says which workflow runs could not be loaded; Runs holds the rest.
	RunsError string            `json:"runsError,omitempty"`
	Comments  []model.PRComment `json:"comments,omitempty"`
	// GavelResults holds one summary per gavel sticky comment on the PR
	// (typically one per matrix shard). Order matches the order of the
	// sticky comments on the PR.
	GavelResults []*GavelResultsSummary `json:"gavelResults,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// problem says what GitHub could not confirm about the detail, or "".
func (d prDetail) problem() string {
	if d.Error != "" {
		return d.Error
	}
	if d.RunsError != "" {
		return d.RunsError
	}
	for id, run := range d.Runs {
		if run.Stale != "" {
			return fmt.Sprintf("workflow run %d is an earlier copy: %s", id, run.Stale)
		}
	}
	return ""
}

// prDetailFetch configures one fetch of a PR's detail.
type prDetailFetch struct {
	// Revalidate asks GitHub even for workflow runs cached as completed.
	Revalidate bool
	// Emit receives each section as it arrives; nil fetches silently.
	Emit detailEmitter
}

type detailEmitter func(event string, data any)

// prFrame builds the SSE `pr` payload. Comments are normalized to a non-nil
// slice: a PR with no actionable comments would otherwise marshal as
// `"comments":null`, which the UI rejects as a malformed frame.
func prFrame(pr *model.PRInfo, comments []model.PRComment) map[string]any {
	if comments == nil {
		comments = []model.PRComment{}
	}
	return map[string]any{"pr": pr, "comments": comments}
}

// emitRuns and emitGavel always send their section, empty ones included: the UI
// keeps what it last received, so a skipped frame would pin an older commit's.
func (emit detailEmitter) emitRuns(runs map[int64]*model.WorkflowRun, problem string) {
	if runs == nil {
		runs = map[int64]*model.WorkflowRun{}
	}
	emit("runs", map[string]any{"runs": runs, "error": problem})
}

func (emit detailEmitter) emitGavel(results []*GavelResultsSummary) {
	if results == nil {
		results = []*GavelResultsSummary{}
	}
	emit("gavel", map[string]any{"gavelResults": results})
}

// handleDetail streams a PR's detail over SSE so the UI can render sections as
// they arrive. The cached detail is sent first; it is followed by a fetch from
// GitHub when it is stale. refresh=1 skips it, asks GitHub for everything, and
// re-polls the list so the PR row's status follows.
func (s *Server) handleDetail(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get("repo")
	numStr := r.URL.Query().Get("number")
	if repo == "" || numStr == "" {
		http.Error(w, `{"error":"repo and number params required"}`, http.StatusBadRequest)
		return
	}
	prNumber, err := strconv.Atoi(numStr)
	if err != nil {
		http.Error(w, `{"error":"invalid number"}`, http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		// Fallback: non-streaming response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.fetchPRDetail(r.Context(), repo, prNumber, prDetailFetch{})) //nolint:errcheck
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	emit := detailEmitter(func(event string, data any) {
		b, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("marshal PR detail %s frame: %v", event, err)
			return
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, b)
		flusher.Flush()
	})
	defer emit("done", nil)

	refresh := r.URL.Query().Get("refresh") == "1"
	cacheKey := SyncStatusKey(repo, prNumber)
	if refresh {
		s.requestPoll()
	} else if entry, ok := s.detailCache.Get(cacheKey); ok {
		emit("pr", prFrame(entry.Detail.PR, entry.Detail.Comments))
		emit.emitRuns(entry.Detail.Runs, entry.Detail.RunsError)
		emit.emitGavel(entry.Detail.GavelResults)
		if entry.Detail.problem() == "" && !s.detailCache.IsStale(cacheKey, s.prUpdatedAt(repo, prNumber)) {
			return
		}
	}

	detail := s.loadPRDetail(r.Context(), repo, prNumber, prDetailFetch{Revalidate: refresh, Emit: emit})
	if detail.Error != "" {
		emit("error", map[string]string{"error": detail.Error})
	}
}

// loadPRDetail fetches a PR's detail and caches it, unless a fetch that started
// later already did or ctx was cancelled.
func (s *Server) loadPRDetail(ctx context.Context, repo string, number int, fetch prDetailFetch) prDetail {
	key := SyncStatusKey(repo, number)
	ticket := s.detailCache.Begin()
	detail := s.fetchPRDetail(ctx, repo, number, fetch)
	if ctx.Err() != nil || !s.detailCache.Put(key, ticket, detail) {
		return detail
	}
	if detail.Error == "" {
		s.setGavelSummary(repo, number, aggregateGavelSummaries(detail.GavelResults))
	}
	s.notify()
	return detail
}

// fetchPRDetail loads full PR metadata, workflow runs, and comments from
// GitHub. Returns a prDetail with Error set on failure rather than an error
// value so callers (handleDetail, export) can embed it directly.
func (s *Server) fetchPRDetail(ctx context.Context, repo string, number int, fetch prDetailFetch) prDetail {
	emit := fetch.Emit
	if emit == nil {
		emit = func(string, any) {}
	}
	opts := s.ghOpts
	opts.Repo = repo
	opts.Context = ctx

	pr, err := github.FetchPR(opts, number)
	if err != nil {
		if ctx.Err() == nil {
			logger.Warnf("failed to fetch PR %s#%d: %v", repo, number, err)
		}
		return prDetail{Error: err.Error()}
	}
	detail := prDetail{PR: pr, Comments: prwatch.MergeAndFilter(pr.Comments, pr.ReviewThreads)}
	emit("pr", prFrame(pr, detail.Comments))

	// Scan all comments (including general issue comments, not just review
	// threads) for gavel sticky comments — one per matrix shard.
	artifacts := github.FindGavelArtifacts(slices.Concat(pr.Comments, pr.ReviewThreads))
	runIDs := headRunIDs(pr)
	gavelDone := make(chan struct{})
	go func() {
		defer close(gavelDone)
		detail.GavelResults = fetchGavelArtifacts(opts, artifacts)
	}()

	runs, err := fetchHeadRuns(opts, runIDs, github.RunLogOptions{Checks: pr.StatusCheckRollup, Revalidate: fetch.Revalidate})
	if err != nil {
		if ctx.Err() == nil {
			logger.Warnf("PR %s#%d: %v", repo, number, err)
		}
		detail.RunsError = err.Error()
	}
	detail.Runs = runs
	emit.emitRuns(detail.Runs, detail.RunsError)

	<-gavelDone
	for _, summary := range detail.GavelResults {
		summary.Outdated = !slices.Contains(runIDs, summary.RunID)
	}
	emit.emitGavel(detail.GavelResults)
	return detail
}

// headRunIDs lists the workflow runs behind the checks of the PR's head commit.
func headRunIDs(pr *model.PRInfo) []int64 {
	var ids []int64
	for _, check := range pr.StatusCheckRollup {
		if id, err := github.ExtractRunID(check.DetailsURL); err == nil && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids
}

// fetchHeadRuns returns the runs that loaded, and one error naming those that
// did not.
func fetchHeadRuns(opts github.Options, ids []int64, logOpts github.RunLogOptions) (map[int64]*model.WorkflowRun, error) {
	fetched := make([]*model.WorkflowRun, len(ids))
	errs := make([]error, len(ids))
	slots := make(chan struct{}, maxRunFetches)
	var fetches sync.WaitGroup
	for i, id := range ids {
		fetches.Go(func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			run, err := github.FetchRunJobs(opts, id, logOpts)
			if err != nil {
				errs[i] = fmt.Errorf("load workflow run %d: %w", id, err)
				return
			}
			fetched[i] = run
		})
	}
	fetches.Wait()
	runs := make(map[int64]*model.WorkflowRun, len(ids))
	for i, run := range fetched {
		if run != nil {
			runs[ids[i]] = run
		}
	}
	return runs, errors.Join(errs...)
}

// populatePRDetail fills a PRViewNode's detail fields using the same fetch
// logic that powers /api/prs/detail. Returns a non-nil error only when the
// PR metadata fetch itself fails (in which case the node is left unchanged).
func (s *Server) populatePRDetail(node *PRViewNode) error {
	detail := s.fetchPRDetail(s.context(), node.Repo, node.Number, prDetailFetch{})
	if detail.Error != "" {
		return fmt.Errorf("%s", detail.Error)
	}
	node.PR = detail.PR
	node.Runs = detail.Runs
	node.Comments = detail.Comments
	node.GavelResults = detail.GavelResults
	node.Detail = detail.RunsError
	return nil
}

// prUpdatedAt returns the UpdatedAt for a PR in the current server state.
func (s *Server) prUpdatedAt(repo string, number int) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, pr := range s.prs {
		if pr.Repo == repo && pr.Number == number {
			return pr.UpdatedAt
		}
	}
	return time.Time{}
}
