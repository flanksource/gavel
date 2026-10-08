package gitstate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/gavel/status"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Tracked returns the id of the repository containing dir when a Track of dir
// already registered it, without running git.
func (t *Tracker) Tracked(dir string) (uuid.UUID, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	repo := t.byDir[filepath.Clean(dir)]
	if repo == nil {
		return uuid.Nil, false
	}
	return repo.id, true
}

// State returns the stored git state of the repository containing dir.
func (t *Tracker) State(ctx context.Context, dir string) (State, error) {
	states, errs := t.States(ctx, []string{dir})
	if err := errs[dir]; err != nil {
		return State{}, err
	}
	return states[dir], nil
}

// States returns the stored git state of the repository containing each of
// dirs, keyed by dir. A dir whose repository cannot be tracked has an entry in
// the error map instead and never fails the others.
func (t *Tracker) States(ctx context.Context, dirs []string) (map[string]State, map[string]error) {
	ids := make(map[string]uuid.UUID, len(dirs))
	errs := map[string]error{}
	var mu sync.Mutex
	var group errgroup.Group
	group.SetLimit(cap(t.sem))
	for _, dir := range dirs {
		group.Go(func() error {
			id, err := t.Track(ctx, dir)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs[dir] = err
			} else {
				ids[dir] = id
			}
			return nil
		})
	}
	_ = group.Wait()

	unique := make([]uuid.UUID, 0, len(ids))
	seen := map[uuid.UUID]bool{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	loaded, err := t.store.Load(ctx, unique)
	states := make(map[string]State, len(ids))
	for dir, id := range ids {
		switch state, ok := loaded[id]; {
		case err != nil:
			errs[dir] = err
		case !ok:
			errs[dir] = fmt.Errorf("git state of %s: never scanned successfully", dir)
		case state.ComputedAt.IsZero():
			// Only a failed scan is recorded: there is no state to show yet.
			errs[dir] = fmt.Errorf("git state of %s: %s", dir, state.Error)
		default:
			states[dir] = state
		}
	}
	return states, errs
}

// WorktreeOf finds the live worktree of the repository containing dir whose
// resolved path is path. A path missing from the stored state may be a
// worktree created since the last ref scan, so the refs are rescanned once
// before path is reported as no worktree.
func (t *Tracker) WorktreeOf(ctx context.Context, dir, path string) (Worktree, bool, error) {
	for attempt := 0; ; attempt++ {
		state, err := t.State(ctx, dir)
		if err != nil {
			return Worktree{}, false, err
		}
		wt, found, err := state.worktreeAt(path)
		if err != nil || found || attempt > 0 {
			return wt, found, err
		}
		if err := t.rescanRefs(ctx, dir); err != nil {
			return Worktree{}, false, err
		}
	}
}

func (t *Tracker) rescanRefs(ctx context.Context, dir string) error {
	if _, err := t.Track(ctx, dir); err != nil {
		return err
	}
	t.mu.Lock()
	repo := t.byDir[filepath.Clean(dir)]
	t.mu.Unlock()
	_, err := t.runRefs(ctx, repo, true)
	return err
}

func (state State) worktreeAt(path string) (Worktree, bool, error) {
	for _, wt := range state.Worktrees {
		if wt.Prunable {
			continue
		}
		resolved, err := filepath.EvalSymlinks(wt.Path)
		if errors.Is(err, os.ErrNotExist) {
			// Removed since the state was scanned, so it cannot be path, which exists.
			continue
		}
		if err != nil {
			return Worktree{}, false, fmt.Errorf("resolve worktree %s: %w", wt.Path, err)
		}
		if resolved == path {
			return wt, true, nil
		}
	}
	return Worktree{}, false, nil
}

// Files returns the uncommitted files of the worktree at worktreePath, a
// worktree of the repository containing dir, as its last status scan
// recorded them. A worktree never scanned is scanned first.
func (t *Tracker) Files(ctx context.Context, dir, worktreePath string) ([]status.FileStatus, error) {
	id, err := t.Track(ctx, dir)
	if err != nil {
		return nil, err
	}
	worktreePath = filepath.Clean(worktreePath)
	files, scanned, err := t.store.WorktreeFiles(ctx, id, worktreePath)
	if err != nil || scanned {
		return files, err
	}
	t.mu.Lock()
	repo := t.byDir[filepath.Clean(dir)]
	t.mu.Unlock()
	if err := t.runStatus(ctx, repo, worktreePath, true); err != nil {
		return nil, err
	}
	files, scanned, err = t.store.WorktreeFiles(ctx, id, worktreePath)
	if err != nil {
		return nil, err
	}
	if !scanned {
		return nil, fmt.Errorf("worktree %s of %s has no recorded status", worktreePath, dir)
	}
	return files, nil
}
