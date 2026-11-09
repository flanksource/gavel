package ui

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workspaceEntry struct {
	Options   todoruntime.WorkspaceOptions
	Workspace *native.Workspace
	Err       error
}
type workspaceRevision struct {
	ID          uuid.UUID `json:"id"`
	Dir         string    `json:"dir"`
	Revision    uint64    `json:"revision"`
	Fingerprint string    `json:"-"`
}
type workspaceSnapshot struct {
	Epoch      string                       `json:"epoch"`
	Catalog    uint64                       `json:"catalog"`
	Workspaces map[string]workspaceRevision `json:"workspaces"`
	Error      string                       `json:"error,omitempty"`
}
type workspaceData struct {
	mu       sync.RWMutex
	reloadMu sync.Mutex
	db       *gorm.DB
	entries  map[string]workspaceEntry
	snapshot workspaceSnapshot
	changed  sse.Notifier
}

// StartWorkspaceData prepares read dependencies before the HTTP listener starts.
func (s *Server) StartWorkspaceData(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("workspace data requires PostgreSQL")
	}
	data := &workspaceData{db: db, snapshot: workspaceSnapshot{Epoch: uuid.NewString(), Workspaces: map[string]workspaceRevision{}}}
	s.workspaceData = data
	if err := s.reloadWorkspaceData(ctx); err != nil {
		return err
	}
	if err := s.startWorkspaceChanges(ctx); err != nil {
		return err
	}
	if !s.hosted {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		if err := watcher.Add(filepath.Dir(projectsPath)); err != nil {
			watcher.Close()
			return err
		}
		go func() {
			defer watcher.Close()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					if filepath.Clean(event.Name) == filepath.Clean(projectsPath) {
						if err := s.reloadWorkspaceData(ctx); err != nil {
							data.fail(err)
						}
					}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					data.fail(err)
				}
			}
		}()
	}
	return nil
}

func (s *Server) reloadWorkspaceData(ctx context.Context) error {
	data := s.workspaceData
	data.reloadMu.Lock()
	defer data.reloadMu.Unlock()
	var projects []projectInfo
	if s.workspaces != nil {
		var err error
		projects, err = s.workspaces.Projects(ctx, s.orgID())
		if err != nil {
			return err
		}
	} else {
		rows, err := LoadProjects(LoadProjectsOptions{IncludeHidden: true})
		if err != nil {
			return err
		}
		for _, p := range rows {
			projects = append(projects, projectInfo{Name: p.Name, Dir: p.ResolvedDir(), Repos: p.Repos})
		}
	}
	entries := map[string]workspaceEntry{}
	for _, project := range projects {
		options := todoruntime.WorkspaceOptions{Name: project.Name, RootPath: project.Dir, Repositories: project.Repos, SandboxedRuns: s.hosted}
		if s.workspaces != nil {
			id, ok := s.workspaces.repoID(s.orgID(), project.Dir)
			if !ok {
				return fmt.Errorf("invalid hosted workspace %s", project.Dir)
			}
			options.RootPath = filepath.Join(s.workspaces.Sources.DataDir, "workspaces", s.orgID().String(), id.String())
		}
		entry := workspaceEntry{Options: options}
		provider, err := todoruntime.NewForOrg(ctx, data.db, options, tenant.Scope{OrgID: s.orgID()})
		entry.Err = err
		if err == nil {
			entry.Workspace = provider.Workspace()
		}
		entries[project.Dir] = entry
	}
	data.mu.Lock()
	data.entries = entries
	data.snapshot.Catalog++
	data.snapshot.Error = ""
	data.mu.Unlock()
	if err := data.refresh(ctx); err != nil {
		return err
	}
	data.changed.Notify()
	return nil
}

func (d *workspaceData) fail(err error) {
	d.mu.Lock()
	d.snapshot.Error = err.Error()
	d.mu.Unlock()
	d.changed.Notify()
}

func (s *Server) workspaceReadProvider(ctx context.Context, dir string) (*todoruntime.Provider, error) {
	data := s.workspaceData
	data.mu.RLock()
	entry, ok := data.entries[dir]
	failure := data.snapshot.Error
	data.mu.RUnlock()
	if failure != "" {
		return nil, fmt.Errorf("workspace catalog: %s", failure)
	}
	if !ok {
		return nil, fmt.Errorf("workspace %q is not configured", dir)
	}
	if entry.Err != nil {
		return nil, entry.Err
	}
	return todoruntime.Resolved(ctx, data.db, entry.Options, entry.Workspace)
}

func (s *Server) handleWorkspaceChanges(w http.ResponseWriter, r *http.Request) {
	if s.workspaceData == nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("workspace data has not started"))
		return
	}
	d := s.workspaceData
	wake, unsubscribe := d.changed.Subscribe()
	defer unsubscribe()
	err := sse.ServeSnapshot(w, r, sse.SnapshotOptions{Interval: 30 * time.Second, Wake: wake, Load: func(context.Context) (any, error) {
		d.mu.RLock()
		defer d.mu.RUnlock()
		snapshot := d.snapshot
		snapshot.Workspaces = make(map[string]workspaceRevision, len(d.snapshot.Workspaces))
		for key, value := range d.snapshot.Workspaces {
			snapshot.Workspaces[key] = value
		}
		return snapshot, nil
	}})
	if err != nil && r.Context().Err() == nil {
		d.fail(err)
	}
}
