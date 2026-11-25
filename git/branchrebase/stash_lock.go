package branchrebase

import (
	"context"
	"errors"
	"time"

	gavelgit "github.com/flanksource/gavel/git"
)

func lockStash(ctx context.Context, dir string) (func(), error) {
	l, err := gavelgit.StashLock(ctx, dir)
	if err != nil {
		return nil, err
	}
	ok, err := l.TryLockContext(ctx, 50*time.Millisecond)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("could not acquire the repository stash lock")
	}
	return func() { mustUnlock(l) }, nil
}
