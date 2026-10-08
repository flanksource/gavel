package ui

import (
	"context"
	"fmt"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/google/uuid"
)

// storedRange is the comparison of key.Head against key.Base in the repository
// at dir, read from git_range_stats. A range is a pure function of its two
// commits, so one never compared is compared once here and stored for every
// later read.
func storedRange(ctx context.Context, store *gitstate.Store, repoID uuid.UUID, dir string, key gitstate.RangeKey) (gitstate.Range, error) {
	cached, err := store.Ranges(ctx, repoID, []gitstate.RangeKey{key})
	if err != nil {
		return gitstate.Range{}, err
	}
	if r, ok := cached[key]; ok {
		return r, nil
	}
	return compareAndStoreRange(ctx, store, repoID, dir, key)
}

func compareAndStoreRange(ctx context.Context, store *gitstate.Store, repoID uuid.UUID, dir string, key gitstate.RangeKey) (gitstate.Range, error) {
	compare, err := gavelgit.CompareRange(dir, key.Base, key.Head)
	if err != nil {
		return gitstate.Range{}, err
	}
	r := gitstate.Range{RangeKey: key, RangeCompare: compare}
	if err := store.SaveRanges(ctx, repoID, []gitstate.Range{r}); err != nil {
		return gitstate.Range{}, err
	}
	return r, nil
}

// storedRangeFiles lists the files merge-base(key.Base, key.Head)..key.Head
// changed in the repository at dir, with their repomap language and scopes.
// The list is read from the range's file_list, filled by the first read; the
// classification is applied on every read because it follows repomap config,
// not the commits. It also returns the merge base the list was taken from.
func storedRangeFiles(ctx context.Context, tracker *gitstate.Tracker, dir string, key gitstate.RangeKey) ([]gavelgit.CommitFile, string, error) {
	repoID, err := tracker.Track(ctx, dir)
	if err != nil {
		return nil, "", err
	}
	store := tracker.Store()
	r, err := storedRange(ctx, store, repoID, dir, key)
	if err != nil {
		return nil, "", err
	}
	files, cached, err := store.RangeFiles(ctx, repoID, key)
	if err != nil {
		return nil, "", err
	}
	if !cached {
		files, err = gavelgit.CommitFiles(dir, gavelgit.CommitDiffOptions{Base: r.MergeBase, Head: key.Head, NoScopes: true})
		if err != nil {
			return nil, "", err
		}
		if err := store.SaveRangeFiles(ctx, repoID, key, files); err != nil {
			return nil, "", fmt.Errorf("cache files of %s..%s: %w", r.MergeBase, key.Head, err)
		}
	}
	gavelgit.EnrichCommitFileScopes(dir, files)
	return files, r.MergeBase, nil
}
