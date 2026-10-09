package ui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/commons/logger"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// gitFocusLease is how long one POST /api/git/focus keeps a worktree at
	// the tracker's hot rescan cadence; the UI renews it while the view is
	// mounted.
	gitFocusLease = 30 * time.Second
	// gitStreamInterval re-reads the generations even without a NOTIFY, a
	// safety net for a change made while the LISTEN was reconnecting.
	gitStreamInterval = 30 * time.Second
)

// requestGitTracker is the process's git state tracker, answering 503 when the
// process runs without the database the tracker keeps its rows in.
func (s *Server) requestGitTracker(w http.ResponseWriter) (*gitstate.Tracker, bool) {
	tracker, err := s.context().GitTracker()
	if err != nil {
		respondError(w, http.StatusServiceUnavailable, err.Error())
		return nil, false
	}
	return tracker, true
}

// gitErrorStatus answers 503 for a process without a git state tracker and
// 500 for any other failure to read the git state.
func gitErrorStatus(err error) int {
	if errors.Is(err, gavelctx.ErrNoGitTracker) {
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}

// touchGit rescans the git rows of the repositories owning paths after a
// server-side git mutation changed them. A process without a tracker keeps no
// rows, so there is nothing to refresh.
func (s *Server) touchGit(paths ...string) {
	tracker, err := s.context().GitTracker()
	if err != nil {
		return
	}
	for _, path := range paths {
		if path != "" {
			tracker.Touch(path)
		}
	}
}

// NotifyGitChange wakes every /api/git/stream subscriber to re-read the
// generations.
func (s *Server) NotifyGitChange() {
	s.gitChanges.Notify()
}

// gitGenerations is the /api/git/stream payload: each configured project whose
// repository the tracker has registered, mapped to that repository's
// generation, which advances on every change to its rows.
func (s *Server) gitGenerations(ctx context.Context) (map[string]int64, error) {
	tracker, err := s.context().GitTracker()
	if err != nil {
		return nil, err
	}
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	generations, err := tracker.Store().Generations(ctx)
	if err != nil {
		return nil, err
	}
	byProject := make(map[string]int64, len(projects))
	for _, project := range projects {
		id, tracked := tracker.Tracked(project.ResolvedDir())
		if !tracked {
			continue
		}
		generation, ok := generations[id]
		if !ok {
			return nil, fmt.Errorf("git repo %s of project %s has no row", id, project.Name)
		}
		byProject[project.Name] = generation
	}
	return byProject, nil
}

// handleGitStream pushes {project: generation} whenever a repository's git
// rows change, so the UI refetches only the projects whose generation moved.
func (s *Server) handleGitStream(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requestGitTracker(w); !ok {
		return
	}
	wake, unsubscribe := s.gitChanges.Subscribe()
	defer unsubscribe()
	err := sse.ServeSnapshot(w, r, sse.SnapshotOptions{
		Load:     func(ctx context.Context) (any, error) { return s.gitGenerations(ctx) },
		Interval: gitStreamInterval,
		Wake:     wake,
	})
	if err != nil {
		logger.Warnf("git stream: %v", err)
	}
}

// gitMetricsResponse is GET /api/git/metrics: whether this process tracks git
// state, and the tracker's metrics so far.
type gitMetricsResponse struct {
	Tracking bool `json:"tracking"`
	gitstate.MetricsSnapshot
}

// handleGitMetrics summarises the gavel_git_* metrics /metrics exposes for the
// activity page.
func (s *Server) handleGitMetrics(w http.ResponseWriter, r *http.Request) {
	snapshot, err := gitstate.Metrics(prometheus.DefaultGatherer)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, trackerErr := s.context().GitTracker()
	respondJSON(w, http.StatusOK, gitMetricsResponse{Tracking: trackerErr == nil, MetricsSnapshot: snapshot})
}

type gitFocusRequest struct {
	Project  string `json:"project"`
	Worktree string `json:"worktree,omitempty"`
}

// handleGitFocus leases the hot rescan cadence to the worktree a view is
// showing: the project's directory, or the linked worktree named like
// ?worktree= on the project endpoints.
func (s *Server) handleGitFocus(w http.ResponseWriter, r *http.Request) {
	var request gitFocusRequest
	if err := decodeStrict(r, &request); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	project, err := GetProject(request.Project)
	if err != nil {
		respondError(w, statusForProjectErr(err), err.Error())
		return
	}
	tracker, ok := s.requestGitTracker(w)
	if !ok {
		return
	}
	ctx := s.requestContext(r)
	if _, err := tracker.Track(ctx, project.ResolvedDir()); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	workDir, err := projectWorkDir(ctx, project, request.Worktree)
	if err != nil {
		respondError(w, workDirErrorStatus(err), err.Error())
		return
	}
	tracker.Focus(workDir, gitFocusLease)
	w.WriteHeader(http.StatusNoContent)
}

// GitChangeListener LISTENs on gitstate.ChangeChannel, which every gavel
// process sharing the database NOTIFYs when it changes a repository's git
// rows, and wakes the server's /api/git/stream subscribers.
type GitChangeListener struct {
	pool      *sql.DB
	notify    func()
	ready     chan struct{}
	readyOnce sync.Once
}

// GitChangeListener builds the listener over pool, which must use the pgx
// stdlib driver.
func (s *Server) GitChangeListener(pool *sql.DB) *GitChangeListener {
	return &GitChangeListener{pool: pool, notify: s.NotifyGitChange, ready: make(chan struct{})}
}

// Run blocks until ctx ends (returning ctx.Err()) or the LISTEN fails beyond
// captain's reconnect backoff. Every (re)established LISTEN wakes the
// subscribers, since nothing is delivered while nobody listens.
func (l *GitChangeListener) Run(ctx context.Context) error {
	listener, err := captaindb.Listen(ctx, l.pool, gitstate.ChangeChannel)
	if err != nil {
		return err
	}
	l.notify()
	l.readyOnce.Do(func() { close(l.ready) })
	wake := func() error {
		l.notify()
		return nil
	}
	return listener.Run(ctx, func(string) error { return wake() }, wake)
}

// Ready is closed once the first LISTEN is established.
func (l *GitChangeListener) Ready() <-chan struct{} { return l.ready }
