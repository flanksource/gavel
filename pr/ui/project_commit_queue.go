package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/google/uuid"
)

const projectActionOpenPR projectAction = "open-pr"

// The commit queue lets the dashboard build up several commits without waiting:
// each selection is posted as its own group and becomes one task inside a
// per-project clicky group started WithConcurrency(1). That group *is* the
// queue — it owns the single-runner guarantee, cancellation and history, so
// nothing here re-implements a scheduler.
//
// Order comes from declaring every group already queued as a dependency with
// WithDependencies. The group's permit alone only guarantees one commit at a
// time, not which one: tasks waiting on the permit are re-queued with a short
// delay and equally urgent tasks have no defined order between them, so a plain
// fan-in commits the groups shuffled. The dependencies state the ordering the
// user built by hand, and they also stop the queue at the first failing group,
// which is where clicky cancels a task whose dependency failed. Cancelling a
// group is not a failure and takes nothing else with it.
//
// Serialization is not cosmetic: `gavel commit <files>` stages explicitly and
// resets the git index, so two commit processes must never overlap on one repo.
// This mirrors the CLI's sequential loop in commit/interactive_batch.go.

// Times are pointers so a group that has not started — clicky stamps a task's
// start time when it is created, not when it is dequeued — reports no elapsed
// time at all instead of a clock ticking since it was queued.
type commitQueueEntryView struct {
	ID        string        `json:"id"`
	Action    projectAction `json:"action"`
	Files     []string      `json:"files"`
	Status    string        `json:"status"`
	StartedAt *time.Time    `json:"startedAt,omitempty"`
	EndedAt   *time.Time    `json:"endedAt,omitempty"`
	ExitCode  int           `json:"exitCode,omitempty"`
	Output    string        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type commitQueueStatus struct {
	RunID   string                 `json:"runId,omitempty"`
	Href    string                 `json:"href,omitempty"`
	Running bool                   `json:"running"`
	Entries []commitQueueEntryView `json:"entries,omitempty"`
}

type commitQueueEntry struct {
	action projectAction
	files  []string
	task   clickytask.TypedTask[cexec.ExecResult]
	output *commitQueueOutput
}

type commitQueue struct {
	mu       sync.Mutex
	entries  []*commitQueueEntry
	group    *clickytask.TypedGroup[cexec.ExecResult]
	runID    string
	archived bool
}

type commitQueueRegistry struct {
	mu     sync.Mutex
	queues map[string]*commitQueue
}

// commitQueueOutput collects one group's streamed command output; the dashboard
// polls it while the commit is still running.
type commitQueueOutput struct {
	mu   sync.Mutex
	text strings.Builder
}

func (o *commitQueueOutput) Write(data []byte) (int, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.Write(data)
}

func (o *commitQueueOutput) String() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.text.String()
}

func (s *Server) projectCommitQueue(project string) *commitQueue {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.commitQueues == nil {
		s.commitQueues = &commitQueueRegistry{queues: map[string]*commitQueue{}}
	}
	queue, ok := s.commitQueues.queues[project]
	if !ok {
		queue = &commitQueue{}
		s.commitQueues.queues[project] = queue
	}
	return queue
}

func (s *Server) commitQueueFor(project string) commitQueueStatus {
	queue := s.projectCommitQueue(project)
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return queue.statusLocked()
}

// enqueueCommitGroup validates one selection through the same argv builder the
// other project actions use and hands it to the project's commit queue. It
// returns as soon as the task is enqueued — the commit runs asynchronously,
// behind whatever is already queued.
func (s *Server) enqueueCommitGroup(project Project, request projectActionRequest) (commitQueueStatus, error) {
	action, files, args, err := s.commitQueueActionArgs(project, request)
	if err != nil {
		return commitQueueStatus{}, err
	}
	return s.projectCommitQueue(project.Name).enqueue(s, project, action, files, args), nil
}

func (s *Server) commitQueueActionArgs(project Project, request projectActionRequest) (projectAction, []string, []string, error) {
	action := request.Action
	if action != projectActionCommit && action != projectActionOpenPR {
		return "", nil, nil, fmt.Errorf("unknown commit queue action %q", action)
	}
	if action == projectActionOpenPR {
		if request.Options != nil {
			return "", nil, nil, errors.New("advanced options are not supported for open-pr")
		}
		request.Action = projectActionCommit
	}
	args, err := s.projectActionArgs(project, request)
	if err != nil {
		return "", nil, nil, err
	}
	files, err := commitGroupFiles(request)
	if err != nil {
		return "", nil, nil, err
	}
	if action == projectActionOpenPR {
		position := len(args) - len(files)
		withPush := make([]string, 0, len(args)+1)
		withPush = append(withPush, args[:position]...)
		withPush = append(withPush, "--push")
		args = append(withPush, args[position:]...)
	}
	return action, files, args, nil
}

// commitGroupFiles reports the paths a queued group owns so the dashboard can
// lock them out of the next selection. projectActionArgs has already validated
// them against the current status.
func commitGroupFiles(request projectActionRequest) ([]string, error) {
	if request.Options != nil {
		raw, present := request.Options["files"]
		if !present {
			return nil, fmt.Errorf("commit requires at least one selected file")
		}
		return projectActionOptionPaths(raw)
	}
	return request.Files, nil
}

func (q *commitQueue) enqueue(s *Server, project Project, action projectAction, files, args []string) commitQueueStatus {
	q.mu.Lock()
	group := q.ensureGroupLocked(project)
	entry := &commitQueueEntry{action: action, files: files, output: &commitQueueOutput{}}
	var opts []clickytask.Option
	if predecessors := q.predecessorsLocked(); len(predecessors) > 0 {
		opts = append(opts, clickytask.WithDependencies(predecessors...))
	}
	entry.task = executeProjectAction(group.Context(), project.ResolvedDir(), args, entry.output, group.Group, opts...)
	q.entries = append(q.entries, entry)
	current := q.statusLocked()
	q.mu.Unlock()

	go q.watch(s, project, entry)
	return current
}

// predecessorsLocked returns the groups a newly queued one has to wait for:
// every group still in the queue, not just the one in front of it. Depending on
// all of them is what makes a failure stop the queue rather than only the next
// group — a group cancelled because an earlier one failed counts as finished to
// the group behind it, so a pairwise chain lets the third group commit against
// the tree the first one never produced.
func (q *commitQueue) predecessorsLocked() []*clickytask.Task {
	predecessors := make([]*clickytask.Task, 0, len(q.entries))
	for _, entry := range q.entries {
		predecessors = append(predecessors, entry.task.Task)
	}
	return predecessors
}

// ensureGroupLocked returns the group new tasks join. A finished group is
// replaced rather than reused: clicky evicts terminal groups from its registry
// after its run-retention window, so a stale handle would silently lose the
// /tasks view. Entries belong to a group, so a fresh group starts a fresh list.
func (q *commitQueue) ensureGroupLocked(project Project) *clickytask.TypedGroup[cexec.ExecResult] {
	if q.group != nil && q.group.FinishedAt().IsZero() {
		return q.group
	}
	q.runID = uuid.NewString()
	q.entries = nil
	q.archived = false
	controller := &projectActionController{}
	group := clickytask.StartGroup[cexec.ExecResult](
		"Commit "+project.Name,
		clickytask.WithGroupID(q.runID),
		clickytask.WithKind("gavel-"+string(projectActionCommit)),
		clickytask.WithLabels(map[string]string{"project": project.Name, "action": string(projectActionCommit)}),
		clickytask.WithHref(commitQueueHref(q.runID)),
		clickytask.WithConcurrency(1),
		clickytask.WithController(controller),
	)
	controller.setGroup(group.Group)
	q.group = &group
	return q.group
}

func commitQueueHref(runID string) string {
	return "/tasks/" + runID
}

// watch waits for one group's commit so the run is archived once the last of
// them is done. Stopping the queue after a failure is not its job: a commit that
// exits non-zero fails its task, and the groups behind it depend on that task.
func (q *commitQueue) watch(s *Server, project Project, entry *commitQueueEntry) {
	_, _ = entry.task.GetResult()
	q.settle(s, project)
}

// settle archives the run once every group in it has finished, mirroring what
// startProjectAction does for lint and test runs. Archiving happens under the
// lock so `archived` means "the run is on disk" rather than "someone is about
// to write it" — the last group's watcher is the only one that gets here.
func (q *commitQueue) settle(s *Server, project Project) {
	q.mu.Lock()
	if q.readyToArchiveLocked() {
		q.archived = true
		if err := taskhistory.Archive(project.ResolvedDir(), q.runID); err != nil {
			logger.Errorf("archive commit queue run %s: %v", q.runID, err)
		}
	}
	q.mu.Unlock()

	s.notify()
}

func (q *commitQueue) readyToArchiveLocked() bool {
	if q.archived || q.runID == "" {
		return false
	}
	for _, entry := range q.entries {
		if !commitQueueTerminal(entry.task.Status()) {
			return false
		}
	}
	return true
}

// cancel stops a queued or running group — clicky's Task.Cancel covers both,
// skipping it at dequeue when pending and killing the process when running —
// and drops it from the queue.
func (q *commitQueue) cancel(id string) (commitQueueStatus, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for index, entry := range q.entries {
		if entry.task.ID() != id {
			continue
		}
		entry.task.Cancel()
		q.entries = append(q.entries[:index], q.entries[index+1:]...)
		return q.statusLocked(), nil
	}
	return commitQueueStatus{}, fmt.Errorf("unknown commit group %q", id)
}

func (q *commitQueue) statusLocked() commitQueueStatus {
	if q.group == nil {
		return commitQueueStatus{}
	}
	current := commitQueueStatus{RunID: q.runID, Href: commitQueueHref(q.runID)}
	for _, entry := range q.entries {
		status := entry.task.Status()
		if !commitQueueTerminal(status) {
			current.Running = true
		}
		view := commitQueueEntryView{
			ID:      entry.task.ID(),
			Action:  entry.action,
			Files:   entry.files,
			Status:  string(status),
			EndedAt: optionalTime(entry.task.EndTime()),
			Output:  entry.output.String(),
		}
		if status != clickytask.StatusPending {
			view.StartedAt = optionalTime(entry.task.StartTime())
		}
		if result, _ := entry.task.Task.GetResult(); result != nil {
			if exec, ok := result.(cexec.ExecResult); ok {
				view.ExitCode = exec.ExitCode
			}
		}
		if err := entry.task.Error(); err != nil {
			view.Error = err.Error()
		}
		current.Entries = append(current.Entries, view)
	}
	return current
}

func optionalTime(at time.Time) *time.Time {
	if at.IsZero() {
		return nil
	}
	return &at
}

func commitQueueTerminal(status clickytask.Status) bool {
	return status != clickytask.StatusPending && status != clickytask.StatusRunning
}

func (s *Server) handleCommitQueue(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	var request projectActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	current, err := s.enqueueCommitGroup(project, request)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, current)
}

func (s *Server) handleCommitQueueEntry(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	current, err := s.projectCommitQueue(project.Name).cancel(r.PathValue("id"))
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, current)
}
