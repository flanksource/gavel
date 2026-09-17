package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	clickytask "github.com/flanksource/clicky/task"
)

var errCommitRunNotFound = errors.New("not found")

// retryableCommitRun is a finished commit queue run reduced to the requests
// that failed or were canceled, in the order they were queued.
type retryableCommitRun struct {
	project  Project
	requests []projectActionRequest
}

// retryCommitRun re-queues the failed and canceled commits of a finished run,
// live or archived, as a new generation and returns its run id.
func (s *Server) retryCommitRun(ctx context.Context, runID string) (projectCommitRun, error) {
	run, err := s.loadRetryableCommitRun(ctx, runID)
	if err != nil {
		return projectCommitRun{}, err
	}
	return s.enqueueCommitRetry(run)
}

// retryCommitRunControl adapts retryCommitRun to the control hooks, which only
// report whether the retry was accepted.
func (s *Server) retryCommitRunControl(ctx context.Context, runID string) error {
	_, err := s.retryCommitRun(ctx, runID)
	return err
}

func (s *Server) loadRetryableCommitRun(ctx context.Context, runID string) (retryableCommitRun, error) {
	snapshots, err := s.commitRunSnapshots(ctx, runID)
	if err != nil {
		return retryableCommitRun{}, err
	}
	var group *clickytask.TaskSnapshot
	statuses := map[string]clickytask.Status{}
	for i := range snapshots {
		switch snapshots[i].Type {
		case "group":
			group = &snapshots[i]
		case "task":
			statuses[snapshots[i].ID] = clickytask.Status(snapshots[i].Status)
		}
	}
	if group == nil {
		return retryableCommitRun{}, fmt.Errorf("run %q has no group snapshot", runID)
	}
	if group.Kind != commitQueueKind {
		return retryableCommitRun{}, fmt.Errorf("run %q is a %q run, not a commit queue run", runID, group.Kind)
	}
	if !commitQueueTerminal(clickytask.Status(group.Status)) {
		return retryableCommitRun{}, fmt.Errorf("run %q is still %s; retry it once it finishes", runID, group.Status)
	}
	name := group.Labels["project"]
	if name == "" {
		return retryableCommitRun{}, fmt.Errorf("run %q has no project label", runID)
	}
	project, err := GetProject(name)
	if err != nil {
		return retryableCommitRun{}, fmt.Errorf("run %q: %w", runID, err)
	}
	details, err := decodeCommitGroupDetails(runID, group.Details)
	if err != nil {
		return retryableCommitRun{}, err
	}
	requests, err := retryCommitRequests(runID, details, statuses)
	if err != nil {
		return retryableCommitRun{}, err
	}
	return retryableCommitRun{project: project, requests: requests}, nil
}

// commitRunSnapshots prefers the live task registry and falls back to the
// archive, which is the only record once clicky evicts a finished run.
func (s *Server) commitRunSnapshots(ctx context.Context, runID string) ([]clickytask.TaskSnapshot, error) {
	if snapshots := clickytask.SnapshotByID(runID); len(snapshots) > 0 {
		return snapshots, nil
	}
	if s.taskSource == nil {
		return nil, fmt.Errorf("load archived run %q: task source is not configured", runID)
	}
	snapshots, err := s.taskSource.Snapshot(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("load archived run %q: %w", runID, err)
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("run %q %w", runID, errCommitRunNotFound)
	}
	return snapshots, nil
}

// decodeCommitGroupDetails accepts both a live projectCommitGroupDetails and
// the generic map an archived snapshot decodes to.
func decodeCommitGroupDetails(runID string, raw any) (projectCommitGroupDetails, error) {
	var details projectCommitGroupDetails
	encoded, err := json.Marshal(raw)
	if err != nil {
		return details, fmt.Errorf("encode commit details of run %q: %w", runID, err)
	}
	if err := json.Unmarshal(encoded, &details); err != nil {
		return details, fmt.Errorf("decode commit details of run %q from %s: %w", runID, encoded, err)
	}
	if len(details.Entries) == 0 {
		return details, fmt.Errorf("run %q has no commit entries", runID)
	}
	return details, nil
}

func retryCommitRequests(runID string, details projectCommitGroupDetails, statuses map[string]clickytask.Status) ([]projectActionRequest, error) {
	var requests []projectActionRequest
	for _, entry := range details.Entries {
		status, ok := statuses[entry.TaskID]
		if !ok {
			return nil, fmt.Errorf("run %q has no task snapshot for commit %s", runID, entry.TaskID)
		}
		if !commitRetryable(status) {
			continue
		}
		request := projectActionRequest{Action: entry.Action, Files: entry.Files}
		if entry.Options != nil {
			request = projectActionRequest{Action: entry.Action, Options: entry.Options}
		}
		requests = append(requests, request)
	}
	if len(requests) == 0 {
		return nil, fmt.Errorf("run %q has no failed commits to retry", runID)
	}
	return requests, nil
}

// enqueueCommitRetry validates every request against the current working tree
// before queueing any, so a stale file rejects the whole retry.
func (s *Server) enqueueCommitRetry(run retryableCommitRun) (projectCommitRun, error) {
	queued := make([]commitQueueRequest, 0, len(run.requests))
	for _, request := range run.requests {
		next, err := s.commitQueueActionArgs(run.project, request)
		if err != nil {
			return projectCommitRun{}, fmt.Errorf("retry %s: %w", request.Action, err)
		}
		queued = append(queued, next)
	}
	return s.projectCommitQueue(run.project.Name).enqueue(s, run.project, queued)
}

func commitRetryable(status clickytask.Status) bool {
	return status == clickytask.StatusFailed || status == clickytask.StatusCancelled
}

func (s *Server) handleCommitQueueRetry(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	runID := r.PathValue("runId")
	run, err := s.loadRetryableCommitRun(r.Context(), runID)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, errCommitRunNotFound) {
			status = http.StatusNotFound
		}
		respondError(w, status, err.Error())
		return
	}
	if run.project.Name != project.Name {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("run %q belongs to project %q, not %q", runID, run.project.Name, project.Name))
		return
	}
	retried, err := s.enqueueCommitRetry(run)
	if err != nil {
		var conflict *commitQueueConflictError
		if errors.As(err, &conflict) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, retried)
}

// commitQueueController advertises stop while a generation runs and retry once
// it finished with failed or canceled commits.
//
// Actions reads task state from the clicky group rather than the queue's
// entries: clicky may call it while holding its manager lock (GCRuns), and
// enqueue takes that lock while holding commitQueue.mu.
type commitQueueController struct {
	server *Server
	runID  string

	mu    sync.RWMutex
	group *clickytask.Group
}

func (c *commitQueueController) setGroup(group *clickytask.Group) {
	c.mu.Lock()
	c.group = group
	c.mu.Unlock()
}

func (c *commitQueueController) currentGroup() *clickytask.Group {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.group
}

func (c *commitQueueController) Actions() []clickytask.ControlAction {
	group := c.currentGroup()
	if group == nil {
		return nil
	}
	if !commitQueueTerminal(group.Status()) {
		return []clickytask.ControlAction{clickytask.ControlStop}
	}
	for _, item := range group.GetTasks() {
		if commitRetryable(item.GetTask().Status()) {
			return []clickytask.ControlAction{clickytask.ControlRetry}
		}
	}
	return nil
}

func (c *commitQueueController) Control(ctx context.Context, action clickytask.ControlAction) error {
	switch action {
	case clickytask.ControlStop:
		group := c.currentGroup()
		if group == nil {
			return fmt.Errorf("commit run %q task group is not ready", c.runID)
		}
		group.Cancel()
		return nil
	case clickytask.ControlRetry:
		return c.server.retryCommitRunControl(ctx, c.runID)
	default:
		return fmt.Errorf("commit run %q does not support %q", c.runID, action)
	}
}
