package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/flanksource/gavel/status"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/google/uuid"
)

type projectAction string

const (
	projectActionCommit projectAction = "commit"
	projectActionLint   projectAction = "lint"
	projectActionTest   projectAction = "test"
)

type projectActionRequest struct {
	Action  projectAction  `json:"action"`
	Files   []string       `json:"files,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

type ProjectActionSchema struct {
	SchemaVersion int            `json:"schemaVersion"`
	Action        string         `json:"action"`
	Schema        map[string]any `json:"schema"`
	Defaults      map[string]any `json:"defaults"`
}

type ProjectActionOptionsProvider interface {
	Schema(action string) (ProjectActionSchema, error)
	Args(action string, options map[string]any) ([]string, error)
}

type projectActionStatus struct {
	Action    projectAction `json:"action,omitempty"`
	RunID     string        `json:"runId,omitempty"`
	Href      string        `json:"href,omitempty"`
	Running   bool          `json:"running"`
	StartedAt time.Time     `json:"startedAt,omitempty"`
	EndedAt   time.Time     `json:"endedAt,omitempty"`
	ExitCode  int           `json:"exitCode,omitempty"`
	Output    string        `json:"output,omitempty"`
	Error     string        `json:"error,omitempty"`
}

type projectFileProblem struct {
	Kind     status.ProblemKind `json:"kind"`
	Severity string             `json:"severity"`
	Label    string             `json:"label"`
	Line     int                `json:"line,omitempty"`
	Message  string             `json:"message,omitempty"`
}

type projectTestStatus struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

type projectLintStatus struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

type projectFileStatus struct {
	Path           string                `json:"path"`
	PreviousPath   string                `json:"previousPath,omitempty"`
	State          status.FileState      `json:"state"`
	StagedKind     status.ChangeKind     `json:"stagedKind,omitempty"`
	WorkKind       status.ChangeKind     `json:"workKind,omitempty"`
	Adds           int                   `json:"adds"`
	Dels           int                   `json:"dels"`
	ModifiedAt     time.Time             `json:"modifiedAt,omitempty"`
	Language       string                `json:"language,omitempty"`
	Scopes         []string              `json:"scopes,omitempty"`
	TestStatus     projectTestStatus     `json:"testStatus"`
	LintStatus     projectLintStatus     `json:"lintStatus"`
	Problems       []projectFileProblem  `json:"problems,omitempty"`
	ResultsStale   bool                  `json:"resultsStale"`
	RepomapError   string                `json:"repomapError,omitempty"`
	ConflictReason status.ConflictReason `json:"conflictReason,omitempty"`
}

type projectStatusResponse struct {
	Project      Project             `json:"project"`
	WorkDir      string              `json:"workDir"`
	Branch       string              `json:"branch"`
	Files        []projectFileStatus `json:"files"`
	ResultsStale bool                `json:"resultsStale"`
	Action       projectActionStatus `json:"action"`
	CommitQueue  commitQueueStatus   `json:"commitQueue"`
}

type projectActionRegistry struct {
	mu      sync.RWMutex
	actions map[string]projectActionStatus
}

var gatherProjectStatus = status.GatherBase

// projectActionExecutor enqueues a gavel command into group and returns its task
// handle without waiting for it: the commit queue keeps the handle so it can
// report and cancel a queued group, while callers that only want the outcome
// call GetResult on the returned task.
type projectActionExecutor func(ctx context.Context, workDir string, args []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult]

var executeProjectAction projectActionExecutor = runProjectCommand

func runProjectCommand(_ context.Context, workDir string, args []string, output io.Writer, group *clickytask.Group, opts ...clickytask.Option) clickytask.TypedTask[cexec.ExecResult] {
	name := "Run gavel"
	if len(args) > 0 {
		name += " " + args[0]
	}
	taskOpts := append([]clickytask.Option{clickytask.WithGroup(group)}, opts...)
	executable, err := os.Executable()
	if err != nil {
		return clickytask.StartTask(name, func(commonscontext.Context, *clickytask.Task) (cexec.ExecResult, error) {
			return cexec.ExecResult{ExitCode: -1}, fmt.Errorf("resolve gavel executable: %w", err)
		}, taskOpts...)
	}
	process := cexec.NewExec(executable, args...).WithCwd(workDir).Stream(output, output)
	return process.RunAsTask(name, taskOpts...)
}

type projectActionController struct {
	mu    sync.RWMutex
	group *clickytask.Group
}

func (c *projectActionController) Actions() []clickytask.ControlAction {
	c.mu.RLock()
	group := c.group
	c.mu.RUnlock()
	if group == nil || (group.Status() != clickytask.StatusRunning && group.Status() != clickytask.StatusPending) {
		return nil
	}
	return []clickytask.ControlAction{clickytask.ControlStop}
}

func (c *projectActionController) Control(_ context.Context, action clickytask.ControlAction) error {
	if action != clickytask.ControlStop {
		return fmt.Errorf("project action does not support %q", action)
	}
	c.mu.RLock()
	group := c.group
	c.mu.RUnlock()
	if group == nil {
		return errors.New("project action task group is not ready")
	}
	group.Cancel()
	return nil
}

func (c *projectActionController) setGroup(group *clickytask.Group) {
	c.mu.Lock()
	c.group = group
	c.mu.Unlock()
}

func (s *Server) projectActionRegistry() *projectActionRegistry {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.projectActions == nil {
		s.projectActions = &projectActionRegistry{actions: map[string]projectActionStatus{}}
	}
	return s.projectActions
}

func (s *Server) projectActionFor(name string) projectActionStatus {
	registry := s.projectActionRegistry()
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	return registry.actions[name]
}

func (s *Server) startProjectAction(project Project, action projectAction, args []string) (projectActionStatus, error) {
	registry := s.projectActionRegistry()
	registry.mu.Lock()
	if current := registry.actions[project.Name]; current.Running {
		registry.mu.Unlock()
		return projectActionStatus{}, fmt.Errorf("%s is already running for project %s", current.Action, project.Name)
	}
	current := projectActionStatus{Action: action, RunID: uuid.NewString(), Running: true, StartedAt: time.Now().UTC()}
	current.Href = "/tasks/" + current.RunID
	if action == projectActionTest || action == projectActionLint {
		current.Href = "/projects/" + url.PathEscape(project.Name) + "?action=" + string(action) + "&run=" + current.RunID
	}
	controller := &projectActionController{}
	group := clickytask.StartGroup[cexec.ExecResult](
		projectActionTaskName(project.Name, action),
		clickytask.WithGroupID(current.RunID),
		clickytask.WithKind("gavel-"+string(action)),
		clickytask.WithLabels(map[string]string{"project": project.Name, "action": string(action)}),
		clickytask.WithHref(current.Href),
		clickytask.WithController(controller),
	)
	controller.setGroup(group.Group)
	var run *testui.Server
	if action == projectActionTest || action == projectActionLint {
		run = s.projectRunServer().BeginRun(current.RunID, "initial")
		run.SetGitRoot(project.ResolvedDir())
		run.SetRunArgs(map[string]any{"action": action, "args": args})
		run.SetStopFunc(group.Cancel)
		if action == projectActionLint {
			run.SetLintResults(nil)
		}
	}
	registry.actions[project.Name] = current
	registry.mu.Unlock()

	go func() {
		stopSnapshots := make(chan struct{})
		snapshotsStopped := make(chan struct{})
		if run != nil {
			go func() {
				defer close(snapshotsStopped)
				streamProjectActionSnapshots(project.ResolvedDir(), current.StartedAt, run, stopSnapshots)
			}()
		}
		output := &projectActionOutputWriter{registry: registry, project: project.Name}
		result, runErr := executeProjectAction(group.Context(), project.ResolvedDir(), args, output, group.Group).GetResult()
		exitCode := result.ExitCode
		if run != nil {
			close(stopSnapshots)
			<-snapshotsStopped
			loadCompletedProjectActionSnapshot(project.ResolvedDir(), current.StartedAt, run)
			run.SetRunExitCode(exitCode, errors.Is(group.Group.Context().Err(), context.DeadlineExceeded))
			run.SetStopFunc(nil)
			run.MarkDone()
		}
		completed := current
		completed.Running = false
		completed.EndedAt = time.Now().UTC()
		completed.ExitCode = exitCode
		if runErr != nil {
			completed.Error = runErr.Error()
		}

		registry.mu.Lock()
		completed.Output = registry.actions[project.Name].Output
		registry.actions[project.Name] = completed
		registry.mu.Unlock()
		s.notifyTestRunSyncer()
		s.notify()
		if archiveErr := taskhistory.Archive(project.ResolvedDir(), current.RunID); archiveErr != nil {
			logger.Errorf("archive project task %s: %v", current.RunID, archiveErr)
		}
	}()

	return current, nil
}

func projectActionTaskName(project string, action projectAction) string {
	return strings.ToUpper(string(action[:1])) + string(action[1:]) + " " + project
}

type projectActionOutputWriter struct {
	registry *projectActionRegistry
	project  string
}

func (w *projectActionOutputWriter) Write(data []byte) (int, error) {
	w.registry.mu.Lock()
	current := w.registry.actions[w.project]
	current.Output += string(data)
	w.registry.actions[w.project] = current
	w.registry.mu.Unlock()
	return len(data), nil
}

func (s *Server) handleProjectStatus(w http.ResponseWriter, r *http.Request) {
	project, err := GetProject(r.PathValue("name"))
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	result, err := gatherProjectStatus(project.ResolvedDir(), status.Options{NoRepomap: true})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, projectStatusResponse{
		Project:      project,
		WorkDir:      result.WorkDir,
		Branch:       result.Branch,
		Files:        projectStatusFiles(result.Files),
		ResultsStale: result.ResultsStale,
		Action:       s.projectActionFor(project.Name),
		CommitQueue:  s.commitQueueFor(project.Name),
	})
}

func projectStatusFiles(files []status.FileStatus) []projectFileStatus {
	result := make([]projectFileStatus, 0, len(files))
	for _, file := range files {
		view := projectFileStatus{
			Path:         file.Path,
			PreviousPath: file.PreviousPath,
			State:        file.State,
			StagedKind:   file.StagedKind,
			WorkKind:     file.WorkKind,
			Adds:         file.Adds,
			Dels:         file.Dels,
			ModifiedAt:   file.ModifiedAt,
			TestStatus: projectTestStatus{
				Passed: file.TestStatus.Passed, Failed: file.TestStatus.Failed, Skipped: file.TestStatus.Skipped,
			},
			LintStatus: projectLintStatus{
				Errors: file.LintStatus.Errors, Warnings: file.LintStatus.Warnings, Infos: file.LintStatus.Infos,
			},
			ResultsStale:   file.ResultsStale,
			ConflictReason: file.ConflictReason,
		}
		if file.FileMap != nil {
			view.Language = file.FileMap.Language
			for _, scope := range file.FileMap.Scopes {
				view.Scopes = append(view.Scopes, string(scope))
			}
		}
		if file.RepomapError != nil {
			view.RepomapError = file.RepomapError.Error()
		}
		for _, problem := range file.Problems {
			view.Problems = append(view.Problems, projectFileProblem{
				Kind: problem.Kind, Severity: problem.Severity, Label: problem.Label,
				Line: problem.Line, Message: problem.Message,
			})
		}
		result = append(result, view)
	}
	return result
}

func (s *Server) handleProjectAction(w http.ResponseWriter, r *http.Request) {
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
	if request.Action == projectActionCommit {
		respondError(w, http.StatusBadRequest, "commits are queued: POST /api/projects/{name}/commit-queue")
		return
	}
	args, err := s.projectActionArgs(project, request)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	action, err := s.startProjectAction(project, request.Action, args)
	if err != nil {
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, action)
}

func (s *Server) handleProjectActionSchema(w http.ResponseWriter, r *http.Request) {
	if _, err := GetProject(r.PathValue("name")); err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	action := r.URL.Query().Get("action")
	if action != string(projectActionCommit) && action != string(projectActionLint) && action != string(projectActionTest) {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("unsupported project action %q", action))
		return
	}
	if s.projectActionOptionsProvider == nil {
		respondError(w, http.StatusServiceUnavailable, "project action schema provider is not configured")
		return
	}
	definition, err := s.projectActionOptionsProvider.Schema(action)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, definition)
}

func (s *Server) projectActionArgs(project Project, request projectActionRequest) ([]string, error) {
	dir := project.ResolvedDir()
	if request.Options != nil {
		if len(request.Files) > 0 {
			return nil, errors.New("advanced project actions cannot include top-level files")
		}
		if request.Action != projectActionCommit && request.Action != projectActionLint && request.Action != projectActionTest {
			return nil, fmt.Errorf("advanced options are not supported for %s", request.Action)
		}
		if s.projectActionOptionsProvider == nil {
			return nil, errors.New("project action options provider is not configured")
		}
		positional := "paths"
		if request.Action != projectActionTest {
			positional = "files"
		}
		if raw, present := request.Options[positional]; present {
			paths, err := projectActionOptionPaths(raw)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", positional, err)
			}
			if request.Action == projectActionCommit {
				if len(paths) == 0 {
					return nil, errors.New("commit requires at least one selected file")
				}
				if err := validateProjectActionFiles(dir, paths); err != nil {
					return nil, err
				}
			} else if err := validateProjectOptionPaths(dir, paths); err != nil {
				return nil, err
			}
		} else if request.Action == projectActionCommit {
			return nil, errors.New("commit requires at least one selected file")
		}
		advanced, err := s.projectActionOptionsProvider.Args(string(request.Action), request.Options)
		if err != nil {
			return nil, err
		}
		base := []string{string(request.Action), "--work-dir", dir}
		if request.Action == projectActionCommit {
			base = append(base, "--precommit=fail")
		}
		return append(base, advanced...), nil
	}
	switch request.Action {
	case projectActionCommit:
		if len(request.Files) == 0 {
			return nil, errors.New("commit requires at least one selected file")
		}
		if err := validateProjectActionFiles(dir, request.Files); err != nil {
			return nil, err
		}
		return append([]string{"commit", "--work-dir", dir, "--precommit=fail"}, request.Files...), nil
	case projectActionLint:
		if len(request.Files) > 0 {
			if err := validateProjectActionFiles(dir, request.Files); err != nil {
				return nil, err
			}
		}
		return append([]string{"lint", "--work-dir", dir}, request.Files...), nil
	case projectActionTest:
		if len(request.Files) > 0 {
			return nil, errors.New("test action derives its scope from current changes; files must be empty")
		}
		return []string{"test", "--work-dir", dir, "--changed"}, nil
	default:
		return nil, fmt.Errorf("unknown project action %q", request.Action)
	}
}

func projectActionOptionPaths(value any) ([]string, error) {
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("expected an array, got %T", value)
	}
	paths := make([]string, 0, len(values))
	for _, value := range values {
		path, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string path, got %T", value)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func validateProjectOptionPaths(dir string, paths []string) error {
	for _, path := range paths {
		clean := filepath.Clean(path)
		if path == "" || filepath.IsAbs(path) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("project action path %q escapes the project", path)
		}
		absolute, err := filepath.Abs(filepath.Join(dir, clean))
		if err != nil {
			return fmt.Errorf("resolve project action path %q: %w", path, err)
		}
		relative, err := filepath.Rel(dir, absolute)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("project action path %q escapes the project", path)
		}
	}
	return nil
}

func validateProjectActionFiles(dir string, requested []string) error {
	result, err := gatherProjectStatus(dir, status.Options{NoRepomap: true})
	if err != nil {
		return fmt.Errorf("gather project status: %w", err)
	}
	available := make(map[string]status.FileState, len(result.Files))
	for _, file := range result.Files {
		available[file.Path] = file.State
	}
	seen := make(map[string]struct{}, len(requested))
	for _, path := range requested {
		state, ok := available[path]
		if !ok {
			return fmt.Errorf("%q is not a current project change", path)
		}
		if state == status.StateConflict {
			return fmt.Errorf("%q has unresolved conflicts", path)
		}
		if _, duplicate := seen[path]; duplicate {
			return fmt.Errorf("duplicate project file %q", path)
		}
		seen[path] = struct{}{}
	}
	return nil
}
