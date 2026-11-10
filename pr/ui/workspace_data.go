package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/commons/logger"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/fsnotify/fsnotify"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workspaceEntry struct {
	Project   projectInfo
	Options   todoruntime.WorkspaceOptions
	Workspace *native.Workspace
	Err       error
}
type workspaceRevision struct {
	ID                uuid.UUID `json:"id"`
	Dir               string    `json:"dir"`
	Revision          uint64    `json:"revision"`
	Fingerprint       string    `json:"-"`
	Labels            uint64    `json:"labels"`
	LabelFingerprint  string    `json:"-"`
	Config            uint64    `json:"config"`
	ConfigFingerprint string    `json:"-"`
}
type workspaceSnapshot struct {
	Epoch      string                       `json:"epoch"`
	Catalog    uint64                       `json:"catalog"`
	Workspaces map[string]workspaceRevision `json:"workspaces"`
	Error      string                       `json:"error,omitempty"`
}
type workspaceData struct {
	mu                 sync.RWMutex
	reloadMu           sync.Mutex
	refreshMu          sync.Mutex
	db                 *gorm.DB
	entries            map[string]workspaceEntry
	snapshot           workspaceSnapshot
	changed            sse.Notifier
	org                uuid.UUID
	workers            sync.WaitGroup
	ready              chan struct{}
	watcher            *fsnotify.Watcher
	catalogFingerprint string
	listenerError      string
	configChanges      map[string]string
	diffs              map[uuid.UUID]map[string]gavelgit.DiffStat
	diffErrors         map[uuid.UUID]error
	diffFingerprints   map[uuid.UUID]string
}

// StartWorkspaceData prepares read dependencies before the HTTP listener starts.
func (s *Server) StartWorkspaceData(ctx context.Context, db *gorm.DB) (err error) {
	if db == nil {
		return fmt.Errorf("workspace data requires PostgreSQL")
	}
	ctx, cancel := context.WithCancel(ctx)
	data := &workspaceData{db: db, org: s.orgID(), ready: make(chan struct{}), configChanges: map[string]string{}, diffs: map[uuid.UUID]map[string]gavelgit.DiffStat{}, diffErrors: map[uuid.UUID]error{}, diffFingerprints: map[uuid.UUID]string{}, snapshot: workspaceSnapshot{Epoch: uuid.NewString(), Workspaces: map[string]workspaceRevision{}}}
	s.workspaceData = data
	defer func() {
		if err != nil {
			cancel()
			data.workers.Wait()
		}
	}()
	if err := s.startWorkspaceChanges(ctx); err != nil {
		return err
	}
	if !s.hosted {
		if err := s.startWorkspaceFiles(ctx); err != nil {
			return err
		}
	}
	if err := s.reloadWorkspaceData(ctx); err != nil {
		return err
	}
	if err := s.reconcileWorkspaceData(ctx); err != nil {
		return err
	}
	if err := s.prepareWorkspaceDiffs(ctx); err != nil {
		return err
	}
	if err := data.refresh(ctx); err != nil {
		return err
	}
	close(data.ready)
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
			info, err := newProjectInfo(ctx, p, todoCountsResult{})
			if err != nil {
				return err
			}
			info.TodoCounts = nil
			projects = append(projects, info)
		}
	}
	encoded, err := json.Marshal(projects)
	if err != nil {
		return err
	}
	data.mu.RLock()
	previous := data.entries
	unchanged := data.catalogFingerprint == string(encoded)
	data.mu.RUnlock()
	if unchanged {
		data.mu.Lock()
		data.snapshot.Error = ""
		data.mu.Unlock()
		return data.refresh(ctx)
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
		entry, exists := previous[project.Dir]
		if exists && reflect.DeepEqual(entry.Options, options) && entry.Err == nil {
			entry.Project = project
			entries[project.Dir] = entry
			continue
		}
		entry = workspaceEntry{Options: options, Project: project}
		provider, err := todoruntime.NewForOrg(ctx, data.db, options, tenant.Scope{OrgID: s.orgID()})
		entry.Err = err
		if err == nil {
			entry.Workspace = provider.Workspace()
		}
		entries[project.Dir] = entry
	}
	if err := data.watchWorkspaceRoots(entries); err != nil {
		return err
	}
	data.mu.Lock()
	data.entries = entries
	data.catalogFingerprint = string(encoded)
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

func (d *workspaceData) snapshotCopy() workspaceSnapshot {
	d.mu.RLock()
	defer d.mu.RUnlock()
	snapshot := d.snapshot
	snapshot.Workspaces = make(map[string]workspaceRevision, len(d.snapshot.Workspaces))
	if snapshot.Error == "" {
		snapshot.Error = d.listenerError
	}
	for key, value := range d.snapshot.Workspaces {
		snapshot.Workspaces[key] = value
	}
	return snapshot
}

func (s *Server) catalogProjects() ([]projectInfo, error) {
	d := s.workspaceData
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.snapshot.Error != "" {
		return nil, fmt.Errorf("workspace catalog: %s", d.snapshot.Error)
	}
	rows := make([]projectInfo, 0, len(d.entries))
	for _, entry := range d.entries {
		info := entry.Project
		if entry.Err != nil {
			info.Error = entry.Err.Error()
		}
		rows = append(rows, info)
	}
	slices.SortFunc(rows, func(a, b projectInfo) int { return strings.Compare(a.Name, b.Name) })
	return rows, nil
}

func (s *Server) workspaceReadProvider(ctx context.Context, dir string) (*todoruntime.Provider, error) {
	data := s.workspaceData
	data.mu.RLock()
	entry, ok := data.entries[dir]
	if !ok {
		for _, candidate := range data.entries {
			if candidate.Options.RootPath == dir {
				entry, ok = candidate, true
				break
			}
		}
	}
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
		return d.snapshotCopy(), nil
	}})
	if err != nil && r.Context().Err() == nil {
		logger.Warnf("workspace change stream: %v", err)
	}
}
