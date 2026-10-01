package native

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// ErrNoRunWorkspace reports an issue none of whose run steps recorded a
// worktree: there is no run commit to land or to verify.
var ErrNoRunWorkspace = errors.New("no run step with a recorded worktree")

// RunWorktree is the worktree the issue's newest run step recorded, and how its
// commits were landed.
type RunWorktree struct {
	PromptRunID uuid.UUID
	Worktree    api.WorktreeState
	// Landing is nil until the run's commits were landed.
	Landing *RunLanding
}

// LatestRunWorktree is the newest run step linked to the issue whose Captain
// record carries a worktree, with its landing. An issue with no such run is
// ErrNoRunWorkspace.
func (r *Repository) LatestRunWorktree(ctx context.Context, captain *captaindb.DB, issueID uuid.UUID) (*RunWorktree, error) {
	if captain == nil {
		return nil, fmt.Errorf("resolve the run worktree of issue %s: no Captain database", issueID)
	}
	ids, err := r.runStepsNewestFirst(ctx, issueID)
	if err != nil {
		return nil, err
	}
	overviews, err := captain.ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return nil, fmt.Errorf("read run steps of issue %s: %w", issueID, err)
	}
	byID := make(map[uuid.UUID]captaindb.PromptRun, len(overviews))
	for _, overview := range overviews {
		byID[overview.ID] = overview.PromptRun
	}
	for _, id := range ids {
		run, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: linked run %s", captaindb.ErrPromptRunNotFound, id)
		}
		if run.Workspace == nil || run.Workspace.Worktree == nil {
			continue
		}
		landing, err := r.landingOf(ctx, issueID, run.ID)
		if err != nil {
			return nil, err
		}
		return &RunWorktree{PromptRunID: run.ID, Worktree: *run.Workspace.Worktree, Landing: landing}, nil
	}
	return nil, fmt.Errorf("%w: none of issue %s's %d run steps recorded one", ErrNoRunWorkspace, issueID, len(ids))
}

func (r *Repository) runStepsNewestFirst(ctx context.Context, issueID uuid.UUID) ([]uuid.UUID, error) {
	links, err := r.ListPromptRuns(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list runs of issue %s: %w", issueID, err)
	}
	var runs []PromptRunLink
	for _, link := range links {
		if link.StepKind == StepRun {
			runs = append(runs, link)
		}
	}
	if len(runs) == 0 {
		return nil, fmt.Errorf("%w: issue %s has no run step", ErrNoRunWorkspace, issueID)
	}
	sort.SliceStable(runs, func(i, j int) bool { return runs[i].CreatedAt.After(runs[j].CreatedAt) })
	ids := make([]uuid.UUID, len(runs))
	for i := range runs {
		ids[i] = runs[i].PromptRunID
	}
	return ids, nil
}

func (r *Repository) landingOf(ctx context.Context, issueID, promptRunID uuid.UUID) (*RunLanding, error) {
	landings, err := r.ListLandings(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("list landings of issue %s: %w", issueID, err)
	}
	for i := range landings {
		if landings[i].PromptRunID == promptRunID {
			return &landings[i], nil
		}
	}
	return nil, nil
}
