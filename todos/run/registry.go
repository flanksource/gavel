// Package run owns TODO run execution: the in-flight run registry, the
// resolved run options, and the seam that starts an agent session for one or
// more TODOs.
//
// It lives here rather than in the dashboard because executing a TODO is not an
// HTTP concern. The dashboard, the CLI and the clicky entity all need to start
// the same run the same way, and the package that owns the behaviour cannot be
// the one that happens to have been the first caller — a handler package is not
// importable by the other two.
//
// It is a sibling of `todos` rather than part of it because starting a run
// needs `todos/drivers`, and `todos/drivers` imports `todos`.
package run

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

var (
	// ErrStopped is the cause a stopped run's context is cancelled with. It is
	// the executor's own cancellation error so a stop reads as a cancellation
	// everywhere downstream rather than as a distinct failure mode.
	ErrStopped      = todos.ErrExecutionCancelled
	ErrAlreadyOwned = errors.New("todo already has a dashboard-owned run")
	ErrNotOwned     = errors.New("todo run is not owned by this dashboard process")
	ErrNotStoppable = errors.New("todo run driver cannot be stopped safely")
	ErrStopping     = errors.New("todo run is already stopping")
)

// ControlStatus reports whether a TODO's in-flight run can be stopped, and
// whether a stop is already under way.
type ControlStatus struct {
	CanStop  bool
	Stopping bool
}

type activeRun struct {
	token       uuid.UUID
	issueIDs    []uuid.UUID
	promptRunID uuid.UUID
	cancel      context.CancelCauseFunc
	stoppable   bool
	stopping    bool
}

// Registry tracks the runs this process owns. It is keyed by run rather than by
// issue because a TODO can have more than one run in flight — a second run is
// started deliberately, after the caller confirmed it — and stopping one of
// them must not stop the other.
type Registry struct {
	mu    sync.Mutex
	byRun map[uuid.UUID]*activeRun
}

func NewRegistry() *Registry {
	return &Registry{byRun: map[uuid.UUID]*activeRun{}}
}

// shared is the process's run registry. "Already running" and "stop this run"
// are properties of the process, not of whichever entrypoint asked: a run
// started from the CLI or the clicky entity has to be visible to — and
// stoppable from — the dashboard, and a TODO must not be startable twice
// because two entrypoints kept separate maps.
var shared = NewRegistry()

// Shared returns the process-wide registry. Tests wanting isolation construct
// their own with NewRegistry.
func Shared() *Registry { return shared }

// RegisterOptions describes the run being claimed.
type RegisterOptions struct {
	IssueIDs  []uuid.UUID
	Stoppable bool
	// Concurrent admits a run against a TODO this process is already running.
	// It is set only when the caller has confirmed the second run.
	Concurrent bool
	Cancel     context.CancelCauseFunc
}

// Handle is one registered run. It is returned instead of a bare release
// function because a run's durable identity is not known until Captain admits
// it, and stopping the run needs that identity.
type Handle struct {
	registry *Registry
	token    uuid.UUID
}

// Register claims a run over the given issues. Claiming is all-or-nothing: a
// group run that overlaps an existing run takes none of its issues, unless the
// caller confirmed running alongside it.
func (r *Registry) Register(opts RegisterOptions) (*Handle, error) {
	if opts.Cancel == nil {
		return nil, errors.New("todo run cancel function is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byRun == nil {
		r.byRun = map[uuid.UUID]*activeRun{}
	}
	if !opts.Concurrent {
		for _, issueID := range opts.IssueIDs {
			if r.runsForIssueLocked(issueID) > 0 {
				return nil, fmt.Errorf("%w: %s", ErrAlreadyOwned, issueID)
			}
		}
	}
	active := &activeRun{
		token: uuid.New(), issueIDs: opts.IssueIDs,
		cancel: opts.Cancel, stoppable: opts.Stoppable,
	}
	r.byRun[active.token] = active
	return &Handle{registry: r, token: active.token}, nil
}

// BindPromptRun records the durable identity Captain admitted for this run, so
// a stop request naming that run reaches this goroutine.
func (h *Handle) BindPromptRun(promptRunID uuid.UUID) {
	if h == nil || promptRunID == uuid.Nil {
		return
	}
	h.registry.mu.Lock()
	defer h.registry.mu.Unlock()
	if active := h.registry.byRun[h.token]; active != nil {
		active.promptRunID = promptRunID
	}
}

// Release gives the claim back when the run finishes.
func (h *Handle) Release() {
	if h == nil {
		return
	}
	h.registry.mu.Lock()
	defer h.registry.mu.Unlock()
	delete(h.registry.byRun, h.token)
}

func (r *Registry) runsForIssueLocked(issueID uuid.UUID) int {
	count := 0
	for _, active := range r.byRun {
		for _, id := range active.issueIDs {
			if id == issueID {
				count++
				break
			}
		}
	}
	return count
}

// Status reports whether any of a TODO's in-flight runs can be stopped.
func (r *Registry) Status(issueID uuid.UUID) ControlStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	status := ControlStatus{}
	for _, active := range r.byRun {
		if !active.ownsIssue(issueID) || !active.stoppable {
			continue
		}
		if active.stopping {
			status.Stopping = true
			continue
		}
		status.CanStop = true
	}
	return status
}

// Stop cancels the run Captain admitted under promptRunID.
func (r *Registry) Stop(promptRunID uuid.UUID) error {
	r.mu.Lock()
	var active *activeRun
	for _, candidate := range r.byRun {
		if candidate.promptRunID == promptRunID && promptRunID != uuid.Nil {
			active = candidate
			break
		}
	}
	switch {
	case active == nil:
		r.mu.Unlock()
		return ErrNotOwned
	case !active.stoppable:
		r.mu.Unlock()
		return ErrNotStoppable
	case active.stopping:
		r.mu.Unlock()
		return ErrStopping
	}
	active.stopping = true
	cancel := active.cancel
	r.mu.Unlock()
	cancel(ErrStopped)
	return nil
}

func (a *activeRun) ownsIssue(issueID uuid.UUID) bool {
	for _, id := range a.issueIDs {
		if id == issueID {
			return true
		}
	}
	return false
}

// IssueIDs are the parseable issue ids of a run's TODOs. A TODO whose ID is not
// a UUID is skipped rather than failing the run: the registry exists to prevent
// overlapping runs, and an unidentifiable TODO cannot overlap anything.
func IssueIDs(todoList []*types.TODO) []uuid.UUID {
	issueIDs := make([]uuid.UUID, 0, len(todoList))
	for _, todo := range todoList {
		if todo == nil {
			continue
		}
		if issueID, err := uuid.Parse(todo.ID); err == nil {
			issueIDs = append(issueIDs, issueID)
		}
	}
	return issueIDs
}
