package ui

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func (s *Server) startWorkspaceFiles(ctx context.Context) error {
	d := s.workspaceData
	if err := os.MkdirAll(filepath.Dir(projectsPath), 0o755); err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		watcher.Close()
		return err
	}
	for _, dir := range []string{filepath.Dir(projectsPath), home} {
		if err := watcher.Add(dir); err != nil {
			watcher.Close()
			return fmt.Errorf("watch workspace catalog/configuration %s: %w", dir, err)
		}
	}
	d.watcher = watcher
	d.workers.Add(1)
	go func() {
		defer d.workers.Done()
		defer watcher.Close()
		select {
		case <-ctx.Done():
			return
		case <-d.ready:
		}
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if err := s.workspaceFileChanged(ctx, event.Name, home); err != nil {
					d.fail(err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				d.fail(err)
			}
		}
	}()
	return nil
}

func (s *Server) workspaceFileChanged(ctx context.Context, path, home string) error {
	d := s.workspaceData
	if filepath.Clean(path) == filepath.Clean(projectsPath) {
		if err := s.reloadWorkspaceData(ctx); err != nil {
			return err
		}
		if err := s.prepareWorkspaceDiffs(ctx); err != nil {
			return err
		}
		return d.refresh(ctx)
	}
	if filepath.Base(path) != ".gavel.yaml" {
		return nil
	}
	d.mu.RLock()
	entries := d.entries
	d.mu.RUnlock()
	for dir, entry := range entries {
		if filepath.Dir(path) != entry.Options.RootPath && filepath.Dir(path) != home {
			continue
		}
		var content []byte
		for _, root := range []string{home, entry.Options.RootPath} {
			encoded, err := os.ReadFile(filepath.Join(root, ".gavel.yaml"))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
			content = append(content, encoded...)
			content = append(content, 0)
		}
		d.mu.Lock()
		d.configChanges[dir] = fmt.Sprintf("%x", sha256.Sum256(content))
		d.mu.Unlock()
	}
	return d.refresh(ctx)
}

func (d *workspaceData) watchWorkspaceRoots(entries map[string]workspaceEntry) error {
	if d.watcher == nil {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	keep := map[string]bool{home: true, filepath.Dir(projectsPath): true}
	for dir, entry := range entries {
		info, err := os.Stat(entry.Options.RootPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			entry.Err = fmt.Errorf("inspect workspace %s: %w", entry.Options.RootPath, err)
			entries[dir] = entry
			continue
		}
		if info.IsDir() {
			keep[entry.Options.RootPath] = true
			if err := d.watcher.Add(entry.Options.RootPath); err != nil {
				return err
			}
		}
	}
	for _, path := range d.watcher.WatchList() {
		if !keep[path] {
			if err := d.watcher.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}
