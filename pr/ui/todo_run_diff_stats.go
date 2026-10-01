package ui

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// commitStatTTL bounds how often the todo list recomputes a workspace's run diff
// stats. Within the TTL the per-issue counts are served straight from the cache
// DB; past it, every issue's latest recorded run range is diffed again.
const commitStatTTL = 60 * time.Second

// runDiffStatProvider is a todo provider backed by native storage, where each
// issue's prompt runs are linked and captain holds the workspace each recorded.
type runDiffStatProvider interface {
	sessionDetailProvider
	Workspace() *native.Workspace
}

// runDiffStats returns each todo's diff footprint, keyed by issue id: the
// Setup..Head range of its latest `run` step that recorded a worktree. It
// serves the cache within commitStatTTL. A provider without native run storage
// has recorded no runs, so it has no stats.
func runDiffStats(ctx context.Context, provider todos.Provider, dir string) (map[string]gavelgit.DiffStat, error) {
	source, ok := provider.(runDiffStatProvider)
	if !ok || source.Captain() == nil || source.Repository() == nil || source.Workspace() == nil {
		return map[string]gavelgit.DiffStat{}, nil
	}
	repo := repoStatKey(dir)
	store := cache.Shared()
	if cached, syncedAt, err := store.CommitStats(ctx, repo); err != nil {
		logger.Debugf("read run diff stats for %s: %v", repo, err)
	} else if !syncedAt.IsZero() && time.Since(syncedAt) < commitStatTTL {
		return cached, nil
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
	stats, err := computeRunDiffStats(links, overviews, landings)
	if err != nil {
		return nil, err
	}
	if err := store.SaveCommitStats(ctx, repo, stats); err != nil {
		logger.Debugf("save run diff stats for %s: %v", repo, err)
	}
	return stats, nil
}

// computeRunDiffStats diffs each issue's latest recorded run range. Only `run`
// steps count, newest first, and the first one whose worktree recorded both
// Setup and Head wins: Setup..Head is the agent's own work, never the setup
// snapshot of copied work-in-progress.
//
// A landed run is diffed from its landing instead, landedSha~N..landedSha with
// N the commits the landing carried, since landing deletes the run branch and
// its Head may be garbage-collected. A landed sha missing from the local
// repository (a PR topic head that only lives on the remote) gives that issue
// no stats. A linked run captain does not know, or any other range git cannot
// diff, is an error.
func computeRunDiffStats(
	links map[uuid.UUID][]native.PromptRunLink,
	overviews []captaindb.PromptRunOverview,
	landings map[uuid.UUID]native.RunLanding,
) (map[string]gavelgit.DiffStat, error) {
	workspaces := make(map[uuid.UUID]*api.WorkspaceRecord, len(overviews))
	for index := range overviews {
		workspaces[overviews[index].ID] = overviews[index].Workspace
	}
	stats := map[string]gavelgit.DiffStat{}
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
		stat, err := runDiffStat(worktree, landing, landed)
		if errors.Is(err, gavelgit.ErrCommitNotFound) && landed {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("diff stats for todo %s (%s): %w", issueID, worktree.Branch, err)
		}
		stats[issueID.String()] = stat
	}
	return stats, nil
}

func runDiffStat(worktree *api.WorktreeState, landing native.RunLanding, landed bool) (gavelgit.DiffStat, error) {
	if landed {
		return gavelgit.LandedDiffStat(worktree.Repo, landing.LandedSHA, landing.CommitCount)
	}
	return gavelgit.RangeDiffStat(worktree.Repo, worktree.Setup, worktree.Head)
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

// repoStatKey normalizes a workspace directory into the stable key used by the
// commit-stat cache tables.
func repoStatKey(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return filepath.Clean(abs)
	}
	return dir
}
