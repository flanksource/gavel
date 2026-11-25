package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofrs/flock"
)

// StashLock coordinates Gavel operations that modify the repository-wide stash.
func StashLock(ctx context.Context, dir string) (*flock.Flock, error) {
	common, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err != nil {
		return nil, fmt.Errorf("locate the shared Git directory for %s: %w", dir, err)
	}
	return flock.New(filepath.Join(strings.TrimSpace(string(common)), "gavel-rebase-stash.lock")), nil
}
