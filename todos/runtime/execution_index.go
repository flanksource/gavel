package runtime

import (
	"context"

	"github.com/google/uuid"

	"github.com/flanksource/gavel/todos/native"
)

// executionIndex is one workspace's prompt-run links, read in bulk.
//
// decorateExecution needs an issue's run links for its attempt count and active
// step. Resolving those as each row is rendered is an N+1 across the whole
// backlog — the same one ListIssuePhaseRuns was written to remove for the phase
// columns.
//
// The active run itself is deliberately NOT batched here. Captain's only
// multi-id accessor is ListPromptRunOverviews, whose query left-joins
// captain_session_overview; that view materializes every session in the
// database and then nested-loop joins it, so it costs ~2s no matter how few ids
// are requested (measured: 17,024 sessions materialized, 1.34M rows discarded by
// the join filter, 10.5M buffer hits, for 79 wanted rows). Batching through it
// is far slower than the per-issue GetPromptRun/GetSession point reads it would
// replace. Batching the run needs that view fixed in captain first.
//
// Only List primes this. A detail read resolves a single issue, where the point
// lookup is already the cheapest answer; both paths go through the same accessor
// below, so there is one read path with a cache in front.
type executionIndex struct {
	links map[uuid.UUID][]native.PromptRunLink
}

// primeExecutionIndex loads a whole workspace's prompt-run links in one query,
// so the subsequent decorateExecution calls resolve them from memory.
func (p *Provider) primeExecutionIndex(ctx context.Context, workspaceID uuid.UUID) error {
	if p.repository == nil {
		return nil
	}

	links, err := p.repository.ListPromptRunLinks(ctx, workspaceID)
	if err != nil {
		return err
	}

	p.execMu.Lock()
	if p.execCache == nil {
		p.execCache = map[uuid.UUID]*executionIndex{}
	}
	p.execCache[workspaceID] = &executionIndex{links: links}
	p.execMu.Unlock()
	return nil
}

// dropExecutionIndex invalidates the memo so a read after a dispatch through
// this provider sees the run it just started.
func (p *Provider) dropExecutionIndex(workspaceID uuid.UUID) {
	p.execMu.Lock()
	delete(p.execCache, workspaceID)
	p.execMu.Unlock()
}

// promptRunLinks resolves one issue's run links from the primed workspace index,
// falling back to a point read. A primed index with no entry for the issue means
// the issue genuinely has no links, not a miss.
func (p *Provider) promptRunLinks(
	ctx context.Context,
	workspaceID, issueID uuid.UUID,
) ([]native.PromptRunLink, error) {
	p.execMu.RLock()
	index := p.execCache[workspaceID]
	p.execMu.RUnlock()
	if index != nil {
		return index.links[issueID], nil
	}
	return p.repository.ListPromptRuns(ctx, issueID)
}
