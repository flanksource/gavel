package gitstate

import "path/filepath"

// Untrack forgets dir. When no other tracked directory names the same
// repository, the repository itself is dropped: the cadence stops scheduling
// it, its focus leases and job records go, and its .git metadata is no longer
// watched. Its stored rows stay; they are simply not refreshed or read.
func (t *Tracker) Untrack(dir string) error {
	dir = filepath.Clean(dir)
	t.mu.Lock()
	repo := t.byDir[dir]
	if repo == nil {
		t.mu.Unlock()
		return nil
	}
	delete(t.byDir, dir)
	for _, other := range t.byDir {
		if other == repo {
			t.mu.Unlock()
			return nil
		}
	}
	delete(t.repos, repo.root)
	trackedRepos.Set(float64(len(t.repos)))
	for key := range t.jobs {
		if key.root == repo.root {
			delete(t.jobs, key)
		}
	}
	for _, wt := range repo.worktrees {
		delete(t.focus, wt)
	}
	watcher := t.watcher
	t.mu.Unlock()
	if watcher == nil {
		return nil
	}
	return watcher.remove(repo)
}
