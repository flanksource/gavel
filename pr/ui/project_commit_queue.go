package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/flanksource/gavel/projects"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/commit"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/flanksource/gavel/status"
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
	// Worktree is the request's worktree, so a retry commits in the same one.
	Worktree string `json:"worktree,omitempty"`
	// IndexPatch is the staged snapshot the commit lands, kept until it is
	// committed so a retry commits the same staging.
	IndexPatch string `json:"indexPatch,omitempty"`
}

type projectCommitGroupDetails struct {
	Entries []projectCommitTaskDetails `json:"entries"`
}

// commitQueueRequest is a validated commit ready to enqueue: the gavel args to
// run in workDir plus the request inputs (files, options, worktree) needed to
// replay it.
type commitQueueRequest struct {
	action   projectAction
	files    []string
	args     []string
	options  map[string]any
	worktree string
	workDir  string
	// indexPatch is the staged snapshot the commit lands; snapshot lists the
	// staged paths enqueue still has to write into it (empty on a retry,
	// whose snapshot already exists).
	indexPatch string
	snapshot   []string
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

// commitQueue serializes the commits of one project working directory: the
// project's own directory or one of its linked worktrees.
type commitQueue struct {
	workDir string
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

// projectCommitQueue is the queue of one project working directory, so each
// worktree serializes its own commits and file claims.
func (s *Server) projectCommitQueue(project projects.Project, workDir string) *commitQueue {
	key := project.Name + "\x00" + workDir
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.commitQueues == nil {
		s.commitQueues = &commitQueueRegistry{queues: map[string]*commitQueue{}}
	}
	queue, ok := s.commitQueues.queues[key]
	if !ok {
		queue = &commitQueue{workDir: workDir}
		s.commitQueues.queues[key] = queue
	}
	return queue
}

// labels tag a generation with its project and, for a linked worktree, the
// worktree path, so the dashboard can list each worktree's commit runs.
func (q *commitQueue) labels(project projects.Project) map[string]string {
	labels := map[string]string{"project": project.Name, "action": string(projectActionCommit)}
	if q.workDir != project.ResolvedDir() {
		labels["worktree"] = q.workDir
	}
	return labels
}

func (s *Server) enqueueCommitGroup(ctx context.Context, project projects.Project, request projectActionRequest) (projectCommitRun, error) {
	queued, err := s.commitQueueActionArgs(project, request)
	if err != nil {
		return projectCommitRun{}, err
	}
	return s.projectCommitQueue(project, queued.workDir).enqueue(ctx, s, project, []commitQueueRequest{queued})
}

func (s *Server) commitQueueActionArgs(project projects.Project, request projectActionRequest) (commitQueueRequest, error) {
	action := request.Action
	if action != projectActionCommit && action != projectActionOpenPR {
		return commitQueueRequest{}, fmt.Errorf("unknown commit queue action %q", action)
	}
	workDir, err := projectWorkDir(s.context(), project, request.Worktree)
	if err != nil {
		return commitQueueRequest{}, err
	}
	queued, err := s.commitQueueArgs(workDir, request)
	if err != nil {
		return commitQueueRequest{}, err
	}
	queued.worktree, queued.workDir = request.Worktree, workDir
	return queued, nil
}

func (s *Server) commitQueueArgs(workDir string, request projectActionRequest) (commitQueueRequest, error) {
	action := request.Action
	if action == projectActionOpenPR {
		if request.Options != nil {
			return commitQueueRequest{}, errors.New("advanced options are not supported for open-pr")
		}
		if len(request.Files) == 0 {
			return pushOnlyCommitRequest(workDir), nil
		}
		request.Action = projectActionCommit
	}
	args, err := s.projectActionArgs(workDir, request)
	if err != nil {
		return commitQueueRequest{}, err
	}
	files, err := commitGroupFiles(request)
	if err != nil {
		return commitQueueRequest{}, err
	}
	queued := commitQueueRequest{action: action, files: files, options: request.Options, indexPatch: request.indexPatch}
	if queued.indexPatch == "" {
		if queued.snapshot, err = stagedCommitPaths(workDir, files); err != nil {
			return commitQueueRequest{}, err
		}
		if queued.indexPatch, err = commit.IndexPatchFile(workDir, uuid.NewString()); err != nil {
			return commitQueueRequest{}, err
		}
	} else if _, err := os.Stat(queued.indexPatch); errors.Is(err, os.ErrNotExist) {
		// A run deletes its snapshot only once the commit lands, so a failed
		// open-pr without one still owes its push.
		if action == projectActionOpenPR {
			return pushOnlyCommitRequest(workDir), nil
		}
		return commitQueueRequest{}, fmt.Errorf("staged snapshot of %s was already committed; nothing to retry", strings.Join(files, ", "))
	} else if err != nil {
		return commitQueueRequest{}, fmt.Errorf("read staged snapshot of %s: %w", strings.Join(files, ", "), err)
	}
	// The positional files trail args; the snapshot replaces them, so a
	// commit lands exactly what was staged, line selections included.
	queued.args = slices.Concat(args[:len(args)-len(files)], []string{"--index-patch", queued.indexPatch})
	if action == projectActionOpenPR {
		queued.args = append(queued.args, "--push")
	}
	return queued, nil
}

// pushOnlyCommitRequest pushes what HEAD already has: gavel commit --push skips
// the commit when nothing is staged, and --stage=staged stops the default
// session staging from picking up a session id inherited by the server process.
func pushOnlyCommitRequest(workDir string) commitQueueRequest {
	return commitQueueRequest{
		action: projectActionOpenPR,
		files:  []string{},
		args:   []string{"commit", "--work-dir", workDir, "--stage=staged", "--push"},
	}
}

// stagedCommitPaths checks every file is a staged change and returns the
// paths its snapshot covers, with each rename's previous path.
func stagedCommitPaths(workDir string, files []string) ([]string, error) {
	result, err := gatherProjectStatus(workDir, status.Options{NoRepomap: true, NoResults: true})
	if err != nil {
		return nil, fmt.Errorf("gather project status: %w", err)
	}
	entries := make(map[string]status.FileStatus, len(result.Files))
	for _, file := range result.Files {
		entries[file.Path] = file
	}
	var paths []string
	for _, file := range files {
		entry := entries[file]
		if entry.State != status.StateStaged && entry.State != status.StateBoth {
			return nil, fmt.Errorf("%q has no staged changes; stage it before committing", file)
		}
		paths = append(paths, file)
		if entry.PreviousPath != "" {
			paths = append(paths, entry.PreviousPath)
		}
	}
	return paths, nil
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
func (q *commitQueue) enqueue(ctx context.Context, s *Server, project projects.Project, requests []commitQueueRequest) (projectCommitRun, error) {
	if len(requests) == 0 {
		return projectCommitRun{}, fmt.Errorf("no commits to queue for project %s", project.Name)
	}
	q.mu.Lock()
	if duplicates := q.duplicateFilesLocked(requests); len(duplicates) > 0 {
		q.mu.Unlock()
		return projectCommitRun{}, &commitQueueConflictError{files: duplicates}
	}
	if err := q.snapshotLocked(ctx, requests); err != nil {
		q.mu.Unlock()
		return projectCommitRun{}, err
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
		entry.task = executeProjectCommit(s, generation.group.Context(), q.workDir, request.args, generation.group.Group, opts...)
		entry.task.SetName(projectCommitTaskName(request.action, request.files))
		entry.task.SetDescription(strings.Join(request.files, ", "))
		generation.entries = append(generation.entries, entry)
		entries = append(entries, entry)
	}
	q.mu.Unlock()

	for _, entry := range entries {
		go q.watch(s, generation, entry)
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

// snapshotLocked moves each new request's staged changes out of the index into
// its patch, so staging done after it is queued cannot leak into the commit.
// A failure restages the snapshots already taken: the batch queues whole or
// not at all.
func (q *commitQueue) snapshotLocked(ctx context.Context, requests []commitQueueRequest) error {
	var taken []commitQueueRequest
	for _, request := range requests {
		if len(request.snapshot) == 0 {
			continue
		}
		err := commit.SnapshotStagedPatch(ctx, commit.SnapshotOptions{WorkDir: q.workDir, Paths: request.snapshot, PatchFile: request.indexPatch})
		if err != nil {
			for _, done := range taken {
				err = errors.Join(err, commit.RestoreStagedPatch(context.WithoutCancel(ctx), q.workDir, done.indexPatch))
			}
			return err
		}
		taken = append(taken, request)
	}
	return nil
}

func predecessorsLocked(generation *commitQueueGeneration) []*clickytask.Task {
	predecessors := make([]*clickytask.Task, 0, len(generation.entries))
	for _, entry := range generation.entries {
		predecessors = append(predecessors, entry.task.Task)
	}
	return predecessors
}

func (q *commitQueue) ensureGenerationLocked(s *Server, project projects.Project) *commitQueueGeneration {
	if q.current != nil && !commitQueueTerminal(q.current.group.Status()) {
		return q.current
	}
	generation := &commitQueueGeneration{runID: uuid.NewString()}
	controller := &commitQueueController{server: s, runID: generation.runID}
	group := clickytask.StartGroup[cexec.ExecResult](
		"Commit "+project.Name,
		clickytask.WithGroupID(generation.runID),
		clickytask.WithKind(commitQueueKind),
		clickytask.WithLabels(q.labels(project)),
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
			TaskID:     entry.task.ID(),
			Action:     entry.action,
			Files:      append([]string{}, entry.files...),
			Options:    entry.options,
			Worktree:   entry.worktree,
			IndexPatch: entry.indexPatch,
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

func (q *commitQueue) watch(s *Server, generation *commitQueueGeneration, entry *commitQueueEntry) {
	_, _ = entry.task.GetResult()
	q.settle(s, generation)
}

func (q *commitQueue) settle(s *Server, generation *commitQueueGeneration) {
	s.touchGit(q.workDir)
	q.mu.Lock()
	if !generation.readyToArchive() {
		q.mu.Unlock()
		s.notify()
		return
	}
	generation.archiving = true
	q.mu.Unlock()

	err := taskhistory.Archive(q.workDir, generation.runID)
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
	project, err := getProject(r.Context(), r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	var request projectActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	run, err := s.enqueueCommitGroup(projectGitLockContext(r), project, request)
	if err != nil {
		var conflict *commitQueueConflictError
		if errors.As(err, &conflict) || errors.Is(err, commit.ErrGitLockHeld) {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		if errors.Is(err, gavelctx.ErrNoGitTracker) {
			respondError(w, http.StatusServiceUnavailable, err.Error())
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, run)
}
