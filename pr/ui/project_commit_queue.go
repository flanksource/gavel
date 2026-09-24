package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/google/uuid"
)

const (
	projectActionOpenPR projectAction = "open-pr"
	commitQueueKind                   = "gavel-" + string(projectActionCommit)
)

type projectCommitRun struct {
	RunID string `json:"runId"`
}

// projectCommitTaskDetails records what one queued commit was asked to do, so a
// finished run — live or archived — can be replayed by retryCommitRun.
type projectCommitTaskDetails struct {
	TaskID  string         `json:"taskId"`
	Action  projectAction  `json:"action"`
	Files   []string       `json:"files"`
	Options map[string]any `json:"options,omitempty"`
}

type projectCommitGroupDetails struct {
	Entries []projectCommitTaskDetails `json:"entries"`
}

// commitQueueRequest is a validated commit ready to enqueue: the gavel args to
// run plus the request inputs (files, options) needed to replay it.
type commitQueueRequest struct {
	action  projectAction
	files   []string
	args    []string
	options map[string]any
}

type commitQueueEntry struct {
	commitQueueRequest
	task clickytask.TypedTask[cexec.ExecResult]
}

type commitQueueGeneration struct {
	runID     string
	entries   []*commitQueueEntry
	group     *clickytask.TypedGroup[cexec.ExecResult]
	archived  bool
	archiving bool
}

type commitQueue struct {
	mu      sync.Mutex
	current *commitQueueGeneration
}

type commitQueueRegistry struct {
	queues map[string]*commitQueue
}

type commitQueueConflictError struct {
	files []string
}

func (e *commitQueueConflictError) Error() string {
	return "project files already queued: " + strings.Join(e.files, ", ")
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

func (s *Server) enqueueCommitGroup(project Project, request projectActionRequest) (projectCommitRun, error) {
	queued, err := s.commitQueueActionArgs(project, request)
	if err != nil {
		return projectCommitRun{}, err
	}
	return s.projectCommitQueue(project.Name).enqueue(s, project, []commitQueueRequest{queued})
}

func (s *Server) commitQueueActionArgs(project Project, request projectActionRequest) (commitQueueRequest, error) {
	action := request.Action
	if action != projectActionCommit && action != projectActionOpenPR {
		return commitQueueRequest{}, fmt.Errorf("unknown commit queue action %q", action)
	}
	if action == projectActionOpenPR {
		if request.Options != nil {
			return commitQueueRequest{}, errors.New("advanced options are not supported for open-pr")
		}
		if len(request.Files) == 0 {
			// Push-only: gavel commit --push skips the commit when nothing is
			// staged. --stage=staged stops the default session staging from
			// picking up a session id inherited by the server process.
			return commitQueueRequest{
				action: action,
				files:  []string{},
				args:   []string{"commit", "--work-dir", project.ResolvedDir(), "--precommit=fail", "--stage=staged", "--push"},
			}, nil
		}
		request.Action = projectActionCommit
	}
	args, err := s.projectActionArgs(project, request)
	if err != nil {
		return commitQueueRequest{}, err
	}
	files, err := commitGroupFiles(request)
	if err != nil {
		return commitQueueRequest{}, err
	}
	if action == projectActionOpenPR {
		position := len(args) - len(files)
		withPush := make([]string, 0, len(args)+1)
		withPush = append(withPush, args[:position]...)
		withPush = append(withPush, "--push")
		args = append(withPush, args[position:]...)
	}
	return commitQueueRequest{action: action, files: files, args: args, options: request.Options}, nil
}

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

// enqueue appends requests, in order, to the project's current generation as
// one atomic batch: every file is checked for conflicts before any commit
// starts, and all of them land in the same run.
func (q *commitQueue) enqueue(s *Server, project Project, requests []commitQueueRequest) (projectCommitRun, error) {
	if len(requests) == 0 {
		return projectCommitRun{}, fmt.Errorf("no commits to queue for project %s", project.Name)
	}
	q.mu.Lock()
	if duplicates := q.duplicateFilesLocked(requests); len(duplicates) > 0 {
		q.mu.Unlock()
		return projectCommitRun{}, &commitQueueConflictError{files: duplicates}
	}
	generation := q.ensureGenerationLocked(s, project)
	entries := make([]*commitQueueEntry, 0, len(requests))
	for _, request := range requests {
		entry := &commitQueueEntry{commitQueueRequest: request}
		entry.files = append([]string{}, request.files...)
		var opts []clickytask.Option
		if predecessors := predecessorsLocked(generation); len(predecessors) > 0 {
			opts = append(opts, clickytask.WithDependencies(predecessors...))
		}
		entry.task = executeProjectAction(generation.group.Context(), project.ResolvedDir(), request.args, io.Discard, generation.group.Group, opts...)
		entry.task.SetName(projectCommitTaskName(request.action, request.files))
		entry.task.SetDescription(strings.Join(request.files, ", "))
		generation.entries = append(generation.entries, entry)
		entries = append(entries, entry)
	}
	q.mu.Unlock()

	for _, entry := range entries {
		go q.watch(s, project, generation, entry)
	}
	return projectCommitRun{RunID: generation.runID}, nil
}

// duplicateFilesLocked lists files claimed by an unfinished commit in the
// current generation or by an earlier request in the same batch.
func (q *commitQueue) duplicateFilesLocked(requests []commitQueueRequest) []string {
	claimed := map[string]struct{}{}
	if q.current != nil && !commitQueueTerminal(q.current.group.Status()) {
		for _, entry := range q.current.entries {
			if commitQueueTerminal(entry.task.Status()) {
				continue
			}
			for _, file := range entry.files {
				claimed[file] = struct{}{}
			}
		}
	}
	reported := map[string]struct{}{}
	var duplicates []string
	for _, request := range requests {
		for _, file := range request.files {
			_, conflict := claimed[file]
			claimed[file] = struct{}{}
			if _, seen := reported[file]; !conflict || seen {
				continue
			}
			reported[file] = struct{}{}
			duplicates = append(duplicates, file)
		}
	}
	return duplicates
}

func predecessorsLocked(generation *commitQueueGeneration) []*clickytask.Task {
	predecessors := make([]*clickytask.Task, 0, len(generation.entries))
	for _, entry := range generation.entries {
		predecessors = append(predecessors, entry.task.Task)
	}
	return predecessors
}

func (q *commitQueue) ensureGenerationLocked(s *Server, project Project) *commitQueueGeneration {
	if q.current != nil && !commitQueueTerminal(q.current.group.Status()) {
		return q.current
	}
	generation := &commitQueueGeneration{runID: uuid.NewString()}
	controller := &commitQueueController{server: s, runID: generation.runID}
	group := clickytask.StartGroup[cexec.ExecResult](
		"Commit "+project.Name,
		clickytask.WithGroupID(generation.runID),
		clickytask.WithKind(commitQueueKind),
		clickytask.WithLabels(map[string]string{"project": project.Name, "action": string(projectActionCommit)}),
		clickytask.WithHref("/tasks/"+generation.runID),
		clickytask.WithConcurrency(1),
		clickytask.WithController(controller),
		clickytask.WithDetailsProvider(func() any { return q.details(generation) }),
	)
	controller.setGroup(group.Group)
	generation.group = &group
	q.current = generation
	return generation
}

func (q *commitQueue) details(generation *commitQueueGeneration) projectCommitGroupDetails {
	q.mu.Lock()
	defer q.mu.Unlock()
	details := projectCommitGroupDetails{Entries: make([]projectCommitTaskDetails, 0, len(generation.entries))}
	for _, entry := range generation.entries {
		details.Entries = append(details.Entries, projectCommitTaskDetails{
			TaskID:  entry.task.ID(),
			Action:  entry.action,
			Files:   append([]string{}, entry.files...),
			Options: entry.options,
		})
	}
	return details
}

func projectCommitTaskName(action projectAction, files []string) string {
	verb := "Commit"
	if action == projectActionOpenPR {
		verb = "Open PR"
	}
	if len(files) == 0 {
		return verb
	}
	if len(files) == 1 {
		return verb + " " + files[0]
	}
	return fmt.Sprintf("%s %d files", verb, len(files))
}

func (q *commitQueue) watch(s *Server, project Project, generation *commitQueueGeneration, entry *commitQueueEntry) {
	_, _ = entry.task.GetResult()
	q.settle(s, project, generation)
}

func (q *commitQueue) settle(s *Server, project Project, generation *commitQueueGeneration) {
	q.mu.Lock()
	if !generation.readyToArchive() {
		q.mu.Unlock()
		s.notify()
		return
	}
	generation.archiving = true
	q.mu.Unlock()

	err := taskhistory.Archive(project.ResolvedDir(), generation.runID)
	q.mu.Lock()
	generation.archiving = false
	generation.archived = err == nil
	q.mu.Unlock()
	if err != nil {
		logger.Errorf("archive commit task group %s: %v", generation.runID, err)
	}
	s.nudgeTaskHistoryImport()
	s.notify()
}

func (generation *commitQueueGeneration) readyToArchive() bool {
	if generation.archived || generation.archiving || len(generation.entries) == 0 {
		return false
	}
	for _, entry := range generation.entries {
		if !commitQueueTerminal(entry.task.Status()) {
			return false
		}
	}
	return true
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
	run, err := s.enqueueCommitGroup(project, request)
	if err != nil {
		var conflict *commitQueueConflictError
		if errors.As(err, &conflict) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, run)
}
