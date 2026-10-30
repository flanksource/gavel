package branchrebase

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

func lockStash(ctx context.Context, dir string) (func(), error) {
	common, err := git(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	l := flock.New(filepath.Join(common, "gavel-rebase-stash.lock"))
	ok, err := l.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("could not acquire the repository stash lock")
	}
	return func() { mustUnlock(l) }, nil
}
