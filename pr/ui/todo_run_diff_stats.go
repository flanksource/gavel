package ui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// runDiffStatProvider is a todo provider backed by native storage, where each
// issue's prompt runs are linked and captain holds the workspace each recorded.
type runDiffStatProvider interface {
	sessionDetailProvider
	Workspace() *native.Workspace
}

// runDiffStats returns each todo's diff footprint, keyed by issue id: the
// Setup..Head range of its latest `run` step that recorded a worktree. A
// provider without native run storage has recorded no runs, so it has no
// stats.
func (s *Server) runDiffStats(ctx context.Context, provider todos.Provider) (map[string]gavelgit.DiffStat, error) {
	source, ok := provider.(runDiffStatProvider)
	if !ok || source.Captain() == nil || source.Repository() == nil || source.Workspace() == nil {
		return map[string]gavelgit.DiffStat{}, nil
	}
	links, err := source.Repository().ListPromptRunLinks(ctx, source.Workspace().ID)
	if err != nil {
		return nil, fmt.Errorf("list prompt run links for run diff stats: %w", err)
	}
	var ids []uuid.UUID
	for _, issueLinks := range links {
		for _, link := range issueLinks {
			if link.StepKind == native.StepRun {
				ids = append(ids, link.PromptRunID)
			}
		}
	}
	if len(ids) == 0 {
		return map[string]gavelgit.DiffStat{}, nil
	}
	overviews, err := source.Captain().ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return nil, fmt.Errorf("load run workspaces for diff stats: %w", err)
	}
	landings, err := source.Repository().ListLandingsForRuns(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load run landings for diff stats: %w", err)
	}
	tracker, err := s.context().GitTracker()
	if err != nil {
		return nil, fmt.Errorf("run diff stats: %w", err)
	}
	return computeRunDiffStats(ctx, tracker, links, overviews, landings)
}

// issueRange is the range an issue's diff footprint is measured over.
type issueRange struct {
	issue  uuid.UUID
	branch string
	key    gitstate.RangeKey
}

// computeRunDiffStats measures each issue's latest recorded run range. Only
// `run` steps count, newest first, and the first one whose worktree recorded
// both Setup and Head wins: Setup..Head is the agent's own work, never the
// setup snapshot of copied work-in-progress.
//
// A landed run is measured from its landing instead, landedSha~N..landedSha
// with N the commits the landing carried, since landing deletes the run branch
// and its Head may be garbage-collected. A landed sha missing from the local
// repository (a PR topic head that only lives on the remote) gives that issue
// no stats. A linked run captain does not know, or any other range git cannot
// compare, is an error.
//
// The footprints are read from git_range_stats; a range never compared before
// is compared once and stored.
func computeRunDiffStats(
	ctx context.Context,
	tracker *gitstate.Tracker,
	links map[uuid.UUID][]native.PromptRunLink,
	overviews []captaindb.PromptRunOverview,
	landings map[uuid.UUID]native.RunLanding,
) (map[string]gavelgit.DiffStat, error) {
	byRepo, err := runRanges(links, overviews, landings)
	if err != nil {
		return nil, err
	}
	stats := map[string]gavelgit.DiffStat{}
	for repo, ranges := range byRepo {
		if err := storedRunDiffStats(ctx, tracker, repo, ranges, stats); err != nil {
			return nil, err
		}
	}
	return stats, nil
}

// runRanges resolves every issue's run range, grouped by repository.
func runRanges(
	links map[uuid.UUID][]native.PromptRunLink,
	overviews []captaindb.PromptRunOverview,
	landings map[uuid.UUID]native.RunLanding,
) (map[string][]issueRange, error) {
	workspaces := make(map[uuid.UUID]*api.WorkspaceRecord, len(overviews))
	for index := range overviews {
		workspaces[overviews[index].ID] = overviews[index].Workspace
	}
	byRepo := map[string][]issueRange{}
	for issueID, issueLinks := range links {
		runID, worktree, err := latestRunWorktree(issueLinks, workspaces)
		if err != nil {
			return nil, err
		}
		if worktree == nil {
			continue
		}
		landing, landed := landings[runID]
		if !landed && worktree.Setup == worktree.Head {
			continue
		}
		key := gitstate.RangeKey{Base: worktree.Setup, Head: worktree.Head}
		if landed {
			base, err := landedBase(worktree.Repo, landing.LandedSHA, landing.CommitCount)
			if errors.Is(err, gavelgit.ErrCommitNotFound) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf("diff stats for todo %s (%s): %w", issueID, worktree.Branch, err)
			}
			key = gitstate.RangeKey{Base: base, Head: landing.LandedSHA}
		}
		byRepo[worktree.Repo] = append(byRepo[worktree.Repo], issueRange{issue: issueID, branch: worktree.Branch, key: key})
	}
	return byRepo, nil
}

// storedRunDiffStats adds the footprint of each of ranges, ranges of the
// repository at repo, to stats.
func storedRunDiffStats(ctx context.Context, tracker *gitstate.Tracker, repo string, ranges []issueRange, stats map[string]gavelgit.DiffStat) error {
	repoID, err := tracker.Track(ctx, repo)
	if err != nil {
		return fmt.Errorf("diff stats for todos in %s: %w", repo, err)
	}
	keys := make([]gitstate.RangeKey, 0, len(ranges))
	for _, r := range ranges {
		keys = append(keys, r.key)
	}
	cached, err := tracker.Store().Ranges(ctx, repoID, keys)
	if err != nil {
		return err
	}
	for _, r := range ranges {
		compared, ok := cached[r.key]
		if !ok {
			if compared, err = compareAndStoreRange(ctx, tracker.Store(), repoID, repo, r.key); err != nil {
				return fmt.Errorf("diff stats for todo %s (%s): %w", r.issue, r.branch, err)
			}
			cached[r.key] = compared
		}
		stats[r.issue.String()] = compared.Diff
	}
	return nil
}

// landedBases memoizes tip~n per repository: both commits are immutable, so
// once resolved the answer never changes, and the todo list resolves every
// landed run's base on each load.
var landedBases sync.Map

type landedBaseKey struct {
	repo, tip string
	n         int
}

func landedBase(repo, tip string, n int) (string, error) {
	key := landedBaseKey{repo: repo, tip: tip, n: n}
	if base, ok := landedBases.Load(key); ok {
		return base.(string), nil
	}
	base, err := gavelgit.LandedBase(repo, tip, n)
	if err != nil {
		return "", err
	}
	landedBases.Store(key, base)
	return base, nil
}

// latestRunWorktree is the newest `run` step whose worktree recorded both
// Setup and Head, with its prompt run id; nil when none did.
func latestRunWorktree(links []native.PromptRunLink, workspaces map[uuid.UUID]*api.WorkspaceRecord) (uuid.UUID, *api.WorktreeState, error) {
	runs := make([]native.PromptRunLink, 0, len(links))
	for _, link := range links {
		if link.StepKind == native.StepRun {
			runs = append(runs, link)
		}
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].CreatedAt.After(runs[j].CreatedAt) })
	for _, link := range runs {
		workspace, ok := workspaces[link.PromptRunID]
		if !ok {
			return uuid.Nil, nil, fmt.Errorf("%w: linked prompt run %s", captaindb.ErrPromptRunNotFound, link.PromptRunID)
		}
		if workspace != nil && workspace.Worktree != nil && workspace.Worktree.Setup != "" && workspace.Worktree.Head != "" {
			return link.PromptRunID, workspace.Worktree, nil
		}
	}
	return uuid.Nil, nil, nil
}
