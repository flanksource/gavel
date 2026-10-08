package gitstate

import (
	"context"
	"path/filepath"
)

// Rescan runs the scans the background would, synchronously: the refs of the
// repository containing dir and then the status of each of its worktrees.
func (t *Tracker) Rescan(ctx context.Context, dir string) error {
	if _, err := t.Track(ctx, dir); err != nil {
		return err
	}
	t.mu.Lock()
	repo := t.byDir[filepath.Clean(dir)]
	t.mu.Unlock()
	return t.scanRepo(ctx, repo)
}

var PorcelainPaths = porcelainPaths
