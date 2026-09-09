package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
)

// runOwnership is the process's heartbeat pump. One goroutine per claimed run
// refreshes the claim so another process can tell a live dispatcher from one
// that exited without finalizing its run.
type runOwnership struct {
	mu     sync.Mutex
	claims map[uuid.UUID]context.CancelFunc
}

func (o *runOwnership) start(repository *native.Repository, promptRunID uuid.UUID, owner native.RunOwner) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.claims == nil {
		o.claims = map[uuid.UUID]context.CancelFunc{}
	}
	if _, running := o.claims[promptRunID]; running {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	o.claims[promptRunID] = cancel
	go func() {
		ticker := time.NewTicker(native.OwnerHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				held, err := repository.TouchPromptRunOwner(ctx, promptRunID, owner.Token)
				if err != nil {
					if ctx.Err() == nil {
						logger.Warnf("could not refresh ownership of prompt run %s: %v", promptRunID, err)
					}
					continue
				}
				if !held {
					// Another process reclaimed the run. Stop asserting a claim
					// this process no longer holds; the run itself continues and
					// reports its own outcome.
					logger.Warnf("ownership of prompt run %s was reclaimed elsewhere", promptRunID)
					return
				}
			}
		}
	}()
}

func (o *runOwnership) stop(promptRunID uuid.UUID) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if cancel, ok := o.claims[promptRunID]; ok {
		cancel()
		delete(o.claims, promptRunID)
	}
}

// claimRun records this process as the run's driver and starts its heartbeat.
func (p *Provider) claimRun(ctx context.Context, promptRunID uuid.UUID, owner native.RunOwner) error {
	if err := p.repository.ClaimPromptRunOwner(ctx, promptRunID, owner); err != nil {
		return err
	}
	p.ownership.start(p.repository, promptRunID, owner)
	return nil
}

// releaseRun stops the heartbeat and drops the claim once a run is finished.
func (p *Provider) releaseRun(ctx context.Context, promptRunID uuid.UUID) {
	p.ownership.stop(promptRunID)
	owner, err := native.LocalOwner()
	if err != nil {
		return
	}
	if err := p.repository.ReleasePromptRunOwner(ctx, promptRunID, owner.Token); err != nil {
		logger.Warnf("could not release ownership of prompt run %s: %v", promptRunID, err)
	}
}

// resolveActiveRunConflict decides what a new dispatch does about a TODO whose
// active prompt run has not finished.
//
// A run is finalized by the process driving it, so a dispatcher that exits
// first leaves the run non-terminal forever and Captain's active-run index
// blocks the TODO for good. Reclaiming that run is the only way back, and the
// owner claim on the link row is what makes "the dispatcher is gone" provable
// rather than a guess about elapsed time.
//
// It reports whether the new run needs an identity of its own instead of the
// deterministic one its seed produces, which is true exactly when it is being
// dispatched alongside a live incumbent.
//
// An error means the dispatch must not proceed.
func (p *Provider) resolveActiveRunConflict(
	ctx context.Context,
	issue *native.Issue,
	preparation todos.RunPreparation,
	dispatchID uuid.UUID,
) (bool, error) {
	if issue.ActivePromptRunID == nil {
		return false, nil
	}
	run, err := p.captain.GetPromptRun(ctx, *issue.ActivePromptRunID)
	if err != nil {
		return false, err
	}
	if terminalPromptRun(run.State) {
		return false, nil
	}
	links, err := p.repository.ListPromptRuns(ctx, issue.ID)
	if err != nil {
		return false, err
	}
	var link *native.PromptRunLink
	for i := range links {
		if links[i].PromptRunID == run.ID {
			link = &links[i]
			break
		}
	}
	if link == nil {
		return false, fmt.Errorf("%w: active prompt run %s is not linked to issue %s", native.ErrLinkConflict, run.ID, issue.ID)
	}
	owner := link.Owner()
	alive, reason := owner.Alive(time.Now())
	if !alive {
		// Nothing is driving this run to completion, so it will never reach a
		// terminal state on its own. Reclaim it rather than leaving the TODO
		// permanently undispatchable.
		logger.Warnf("reclaiming orphaned %s-step run %s on todo %s: %s", link.StepKind, run.ID, issue.ID, reason)
		// The redispatch keeps its deterministic identity: dispatching advanced the
		// issue version, so its seed already differs from the reclaimed run's.
		return false, p.reclaimRun(ctx, run, reason)
	}
	if *issue.ActivePromptRunID == dispatchID {
		// The live run is the one this dispatch would itself create: a replay, or
		// a contender that lost the race to admit it. Not a second run —
		// deterministic admission decides who owns it.
		return false, nil
	}
	if preparation.Resume {
		// Resuming a live run is exactly what PrepareRun's resume branch does
		// next: one interactive operation continues, no second run.
		return false, nil
	}
	if preparation.Concurrent {
		return true, nil
	}
	since := time.Since(run.QueuedAt)
	if run.StartedAt != nil {
		since = time.Since(*run.StartedAt)
	}
	return false, &todos.ErrRunOwnedElsewhere{
		IssueID: issue.ID, PromptRunID: run.ID, StepKind: string(link.StepKind),
		Owner: owner.Describe(), Since: since,
	}
}

// reclaimRun marks an abandoned run cancelled so the TODO can be dispatched
// again. The reason is recorded on the run itself: a reclaimed run must not
// look like one that simply stopped.
func (p *Provider) reclaimRun(ctx context.Context, run *captaindb.PromptRun, reason string) error {
	state := captaindb.PromptRunStateCancelled
	phase := run.Phase
	if phase == captaindb.PromptRunPhaseQueued || phase == captaindb.PromptRunPhasePreRun {
		phase = captaindb.PromptRunPhaseGenerate
	}
	errorText := "reclaimed: " + reason
	if _, err := p.captain.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{
		ID: run.ID, ExpectedVersion: run.Version,
		State: &state, Phase: &phase, Error: &errorText,
	}); err != nil {
		return fmt.Errorf("reclaim orphaned prompt run %s: %w", run.ID, err)
	}
	p.ownership.stop(run.ID)
	return nil
}

// ReclaimRun reclaims one abandoned run by ID, for the stop surfaces that find
// a run no process owns. It reports whether the run was actually reclaimed.
func (p *Provider) ReclaimRun(ctx context.Context, promptRunID uuid.UUID) (bool, string, error) {
	run, err := p.captain.GetPromptRun(ctx, promptRunID)
	if err != nil {
		return false, "", err
	}
	if terminalPromptRun(run.State) {
		return false, fmt.Sprintf("run is already %s", run.State), nil
	}
	owner, err := p.repository.PromptRunOwner(ctx, promptRunID)
	if err != nil && !errors.Is(err, native.ErrNotFound) {
		return false, "", err
	}
	alive, reason := owner.Alive(time.Now())
	if alive {
		return false, reason, nil
	}
	if err := p.reclaimRun(ctx, run, reason); err != nil {
		return false, "", err
	}
	return true, reason, nil
}
