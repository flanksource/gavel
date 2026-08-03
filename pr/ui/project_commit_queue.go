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

const projectActionOpenPR projectAction = "open-pr"

type projectCommitRun struct {
	RunID string `json:"runId"`
}

type projectCommitTaskDetails struct {
	TaskID string        `json:"taskId"`
	Action projectAction `json:"action"`
	Files  []string      `json:"files"`
}

type projectCommitGroupDetails struct {
	Entries []projectCommitTaskDetails `json:"entries"`
}

type commitQueueEntry struct {
	action projectAction
	files  []string
	task   clickytask.TypedTask[cexec.ExecResult]
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
	action, files, args, err := s.commitQueueActionArgs(project, request)
	if err != nil {
		return projectCommitRun{}, err
	}
	return s.projectCommitQueue(project.Name).enqueue(s, project, action, files, args)
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

func (q *commitQueue) enqueue(s *Server, project Project, action projectAction, files, args []string) (projectCommitRun, error) {
	q.mu.Lock()
	if duplicates := q.duplicateFilesLocked(files); len(duplicates) > 0 {
		q.mu.Unlock()
		return projectCommitRun{}, &commitQueueConflictError{files: duplicates}
	}
	generation := q.ensureGenerationLocked(project)
	entry := &commitQueueEntry{action: action, files: append([]string(nil), files...)}
	var opts []clickytask.Option
	if predecessors := predecessorsLocked(generation); len(predecessors) > 0 {
		opts = append(opts, clickytask.WithDependencies(predecessors...))
	}
	entry.task = executeProjectAction(generation.group.Context(), project.ResolvedDir(), args, io.Discard, generation.group.Group, opts...)
	entry.task.SetName(projectCommitTaskName(action, files))
	entry.task.SetDescription(strings.Join(files, ", "))
	generation.entries = append(generation.entries, entry)
	q.mu.Unlock()

	go q.watch(s, project, generation, entry)
	return projectCommitRun{RunID: generation.runID}, nil
}

func (q *commitQueue) duplicateFilesLocked(files []string) []string {
	if q.current == nil || commitQueueTerminal(q.current.group.Status()) {
		return nil
	}
	claimed := map[string]struct{}{}
	for _, entry := range q.current.entries {
		if commitQueueTerminal(entry.task.Status()) {
			continue
		}
		for _, file := range entry.files {
			claimed[file] = struct{}{}
		}
	}
	seen := map[string]struct{}{}
	var duplicates []string
	for _, file := range files {
		if _, exists := claimed[file]; !exists {
			continue
		}
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}
		duplicates = append(duplicates, file)
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

func (q *commitQueue) ensureGenerationLocked(project Project) *commitQueueGeneration {
	if q.current != nil && !commitQueueTerminal(q.current.group.Status()) {
		return q.current
	}
	generation := &commitQueueGeneration{runID: uuid.NewString()}
	controller := &projectActionController{}
	group := clickytask.StartGroup[cexec.ExecResult](
		"Commit "+project.Name,
		clickytask.WithGroupID(generation.runID),
		clickytask.WithKind("gavel-"+string(projectActionCommit)),
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
			TaskID: entry.task.ID(),
			Action: entry.action,
			Files:  append([]string(nil), entry.files...),
		})
	}
	return details
}

func projectCommitTaskName(action projectAction, files []string) string {
	verb := "Commit"
	if action == projectActionOpenPR {
		verb = "Open PR"
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
