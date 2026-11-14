package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/fsnotify/fsnotify"
)

// ActiveRun is a gavel process running against an edited file's workspace,
// as its published status describes it.
type ActiveRun struct {
	PID     int            `json:"pid"`
	Command string         `json:"command"`
	Started time.Time      `json:"started"`
	Groups  []status.Group `json:"groups"`
}

// WatchRuns follows every gavel process that publishes its status under cache
// (sdk/status) from inside one of the edited files' workspaces: its progress
// shows as an ActiveRun, and once it exits the problems are reloaded from the
// runs it recorded. Call it after Reload; it stops with s.Context.
func (s *Server) WatchRuns(cache string) error {
	dir := status.Dir(cache)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("watch gavel runs: %w", err)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch gavel runs: %w", err)
	}
	if err := watcher.Add(dir); err != nil {
		_ = watcher.Close()
		return fmt.Errorf("watch gavel runs in %s: %w", dir, err)
	}
	running, err := status.List(cache)
	if err != nil {
		_ = watcher.Close()
		return err
	}
	followed := map[int]bool{}
	for _, run := range running {
		s.followRun(cache, run.PID, followed)
	}
	go func() {
		defer watcher.Close()
		for {
			select {
			case <-s.Context.Done():
				return
			case err := <-watcher.Errors:
				logger.Errorf("watch gavel runs in %s: %v", dir, err)
			case event := <-watcher.Events:
				name := filepath.Base(event.Name)
				pid, err := strconv.Atoi(strings.TrimSuffix(name, ".json"))
				if err != nil || !strings.HasSuffix(name, ".json") || !event.Has(fsnotify.Create|fsnotify.Write) {
					continue
				}
				s.followRun(cache, pid, followed)
			}
		}
	}()
	return nil
}

// followRun starts following pid unless it is this process, already followed,
// or running outside the edited files' workspaces. followed is only touched
// from the watch loop (and WatchRuns before it starts).
func (s *Server) followRun(cache string, pid int, followed map[int]bool) {
	if pid == os.Getpid() || followed[pid] {
		return
	}
	current, err := status.Read(status.Path(cache, pid))
	if err != nil {
		// A status file is published once written in full; a partial one is
		// read again on its next write.
		logger.V(1).Infof("read gavel run %d: %v", pid, err)
		return
	}
	followed[pid] = true
	if !s.inWorkspace(current.WorkDir) {
		return
	}
	watcher, err := status.Watch(s.Context, status.WatchOptions{Cache: cache, PID: pid})
	if err != nil {
		logger.V(1).Infof("follow gavel run %d: %v", pid, err)
		return
	}
	go func() {
		for update := range watcher.Updates() {
			s.setActive(ActiveRun{PID: pid, Command: update.Command, Started: update.Started, Groups: update.Groups})
		}
		s.mu.Lock()
		s.run.Active = slices.DeleteFunc(s.run.Active, func(run ActiveRun) bool { return run.PID == pid })
		s.mu.Unlock()
		if err := watcher.Err(); err != nil {
			logger.Errorf("follow gavel run %d: %v", pid, err)
		}
		if s.Context.Err() != nil {
			return
		}
		if err := s.Reload(); err != nil {
			s.mu.Lock()
			s.run.Error = err.Error()
			s.mu.Unlock()
			s.updated.Notify()
		}
	}()
}

func (s *Server) setActive(run ActiveRun) {
	s.mu.Lock()
	if i := slices.IndexFunc(s.run.Active, func(active ActiveRun) bool { return active.PID == run.PID }); i >= 0 {
		s.run.Active[i] = run
	} else {
		s.run.Active = append(s.run.Active, run)
	}
	s.mu.Unlock()
	s.updated.Notify()
}

// inWorkspace reports whether a run invoked in workDir records into, or runs
// above, one of the edited files' state roots.
func (s *Server) inWorkspace(workDir string) bool {
	if workDir == "" {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(workDir); err == nil {
		workDir = resolved
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, root := range s.stateRoots {
		if linters.PathMatchesAnyScope(workDir, []string{root}) || linters.PathMatchesAnyScope(root, []string{workDir}) {
			return true
		}
	}
	return false
}
