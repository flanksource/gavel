package gitstate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/fsnotify/fsnotify"
)

// metadataWatcher watches the small, non-recursive set of .git metadata that
// moves when refs or an index change — the common dir (HEAD, packed-refs,
// index), refs/heads and each linked worktree's admin dir — never the
// working trees themselves: their content edits are found by fsmonitor-backed
// status scans.
type metadataWatcher struct {
	tracker *Tracker
	fs      *fsnotify.Watcher

	mu     sync.Mutex
	timers map[jobKey]*time.Timer
}

func newMetadataWatcher(t *Tracker) (*metadataWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("start git metadata watcher: %w", err)
	}
	return &metadataWatcher{tracker: t, fs: w, timers: map[jobKey]*time.Timer{}}, nil
}

func (w *metadataWatcher) add(repo *trackedRepo) error {
	dirs := []string{repo.commonDir, filepath.Join(repo.commonDir, "worktrees")}
	err := filepath.WalkDir(filepath.Join(repo.commonDir, "refs", "heads"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("list refs of %s: %w", repo.root, err)
	}
	if entries, err := os.ReadDir(filepath.Join(repo.commonDir, "worktrees")); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				dirs = append(dirs, filepath.Join(repo.commonDir, "worktrees", entry.Name()))
			}
		}
	}
	for _, dir := range dirs {
		if err := w.fs.Add(dir); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("watch %s: %w", dir, err)
		}
	}
	return nil
}

func (w *metadataWatcher) run(ctx context.Context) {
	defer w.fs.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-w.fs.Errors:
			if !ok {
				return
			}
			logger.Warnf("git metadata watcher: %v", err)
		case event, ok := <-w.fs.Events:
			if !ok {
				return
			}
			w.handle(event)
		}
	}
}

func (w *metadataWatcher) handle(event fsnotify.Event) {
	if strings.HasSuffix(event.Name, ".lock") {
		return
	}
	repo := w.repoOf(event.Name)
	if repo == nil {
		return
	}
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := w.fs.Add(event.Name); err != nil {
				logger.Warnf("git metadata watcher: watch %s: %v", event.Name, err)
			}
		}
	}
	rel, err := filepath.Rel(repo.commonDir, event.Name)
	if err != nil {
		return
	}
	for _, key := range metadataJobs(repo, filepath.ToSlash(rel)) {
		w.schedule(key)
	}
}

// metadataJobs maps a changed path under the common dir to the scans it
// calls for: a ref or HEAD change rescans refs (and the status of the
// worktree whose HEAD moved), an index change rescans that worktree's status.
func metadataJobs(repo *trackedRepo, rel string) []jobKey {
	refs := jobKey{root: repo.root}
	switch {
	case rel == "HEAD":
		return []jobKey{refs, {root: repo.root, worktree: repo.root}}
	case rel == "index":
		return []jobKey{{root: repo.root, worktree: repo.root}}
	case rel == "packed-refs" || strings.HasPrefix(rel, "refs/heads"):
		return []jobKey{refs}
	case strings.HasPrefix(rel, "worktrees/"):
		parts := strings.Split(rel, "/")
		if len(parts) == 2 {
			return []jobKey{refs}
		}
		worktree, err := linkedWorktreePath(repo.commonDir, parts[1])
		if err != nil {
			return []jobKey{refs}
		}
		switch parts[2] {
		case "HEAD":
			return []jobKey{refs, {root: repo.root, worktree: worktree}}
		case "index":
			return []jobKey{{root: repo.root, worktree: worktree}}
		}
	}
	return nil
}

// linkedWorktreePath reads the working tree a linked worktree's admin dir
// belongs to from its gitdir file, which names the worktree's .git file.
func linkedWorktreePath(commonDir, name string) (string, error) {
	gitdir, err := os.ReadFile(filepath.Join(commonDir, "worktrees", name, "gitdir"))
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(string(gitdir))), nil
}

func (w *metadataWatcher) repoOf(path string) *trackedRepo {
	w.tracker.mu.Lock()
	defer w.tracker.mu.Unlock()
	for _, repo := range w.tracker.repos {
		if pathWithin(path, repo.commonDir) {
			return repo
		}
	}
	return nil
}

// schedule enqueues key once events for it have been quiet for the debounce.
func (w *metadataWatcher) schedule(key jobKey) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if timer, ok := w.timers[key]; ok {
		timer.Reset(w.tracker.debounce)
		return
	}
	w.timers[key] = time.AfterFunc(w.tracker.debounce, func() {
		w.mu.Lock()
		delete(w.timers, key)
		w.mu.Unlock()
		w.tracker.enqueue(key)
	})
}
