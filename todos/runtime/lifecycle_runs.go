package runtime

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

type activeRun struct {
	issue *native.Issue
	link  *native.PromptRunLink
	run   *captaindb.PromptRun
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// ActivePromptRun returns the Captain prompt run backing the todo's current
// attempt, or nil when it has none. Callers deciding how to continue a run — the
// runtime it resolved, whether an agent is still live — read it rather than
// re-deriving those facts from the projected status.
func (p *Provider) ActivePromptRun(ctx context.Context, todo *types.TODO) (*captaindb.PromptRun, error) {
	active, err := p.loadActiveRun(ctx, todo)
	if err != nil {
		if errors.Is(err, native.ErrNotFound) || errors.Is(err, captaindb.ErrPromptRunNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return active.run, nil
}

// SessionAttempt resolves the attempt a session belongs to. A resumed turn
// keeps its provider session, so several attempts can share one; the newest
// link wins, as it does in the Session tab.
func (p *Provider) SessionAttempt(ctx context.Context, todo *types.TODO, sessionID string) (run.SessionAttempt, error) {
	if strings.TrimSpace(sessionID) == "" {
		return run.SessionAttempt{}, fmt.Errorf("%w: a session ID is required to resolve an attempt", native.ErrInvalidInput)
	}
	issueID, err := p.todoID(todo)
	if err != nil {
		return run.SessionAttempt{}, err
	}
	links, err := p.repository.ListPromptRuns(ctx, issueID)
	if err != nil {
		return run.SessionAttempt{}, err
	}
	ids := make([]uuid.UUID, len(links))
	for index := range links {
		ids[index] = links[index].PromptRunID
	}
	overviews, err := p.captain.ListPromptRunOverviews(ctx, captaindb.PromptRunOverviewFilter{IDs: ids})
	if err != nil {
		return run.SessionAttempt{}, err
	}
	byID := make(map[uuid.UUID]captaindb.PromptRunOverview, len(overviews))
	for _, overview := range overviews {
		byID[overview.ID] = overview
	}
	sort.SliceStable(links, func(i, j int) bool { return links[i].CreatedAt.After(links[j].CreatedAt) })
	for _, link := range links {
		overview, ok := byID[link.PromptRunID]
		if !ok {
			return run.SessionAttempt{}, fmt.Errorf("%w: linked prompt run %s", captaindb.ErrPromptRunNotFound, link.PromptRunID)
		}
		if run.RunMatchesSession(overview, sessionID) {
			return run.SessionAttempt{PromptRunID: link.PromptRunID, Step: string(link.StepKind)}, nil
		}
	}
	return run.SessionAttempt{}, fmt.Errorf("%w: session %s belongs to no attempt of issue %s", native.ErrNotFound, sessionID, issueID)
}

// loadActiveRun resolves the run a caller is acting on. Inside an execution
// that is the run this execution was prepared for — a TODO may have several
// runs in flight, and each has to report its own outcome. Outside one (read
// surfaces, recovery commands) it is the TODO's current active run.
func (p *Provider) loadActiveRun(ctx context.Context, todo *types.TODO) (*activeRun, error) {
	issueID, err := p.todoID(todo)
	if err != nil {
		return nil, err
	}
	issue, err := p.repository.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	runID := todos.PromptRunFromContext(ctx)
	if runID == uuid.Nil {
		if issue.ActivePromptRunID == nil {
			return nil, fmt.Errorf("%w: issue %s has no active Captain prompt run", native.ErrNotFound, issue.ID)
		}
		runID = *issue.ActivePromptRunID
	}
	run, err := p.captain.GetPromptRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	links, err := p.repository.ListPromptRuns(ctx, issue.ID)
	if err != nil {
		return nil, err
	}
	for i := range links {
		if links[i].PromptRunID == run.ID {
			link := links[i]
			return &activeRun{issue: issue, link: &link, run: run}, nil
		}
	}
	return nil, fmt.Errorf("%w: prompt run %s is not linked to issue %s", native.ErrLinkConflict, run.ID, issue.ID)
}

func (p *Provider) nextPromptOrdinal(ctx context.Context, issueID uuid.UUID, step native.StepKind) (int, error) {
	links, err := p.repository.ListPromptRuns(ctx, issueID)
	if err != nil {
		return 0, err
	}
	next := 0
	for _, link := range links {
		if link.StepKind == step && link.Ordinal >= next {
			next = link.Ordinal + 1
		}
	}
	return next, nil
}

func terminalPromptRun(state captaindb.PromptRunState) bool {
	return state == captaindb.PromptRunStateSucceeded ||
		state == captaindb.PromptRunStateFailed ||
		state == captaindb.PromptRunStateCancelled
}

// markPrepared records that this process dispatched a run for an issue. It is a
// set per issue, not one entry: a TODO can have several runs in flight and each
// of them is separately this process's to finish.
func (p *Provider) markPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	defer p.preparedMu.Unlock()
	if p.prepared == nil {
		p.prepared = map[uuid.UUID]map[uuid.UUID]struct{}{}
	}
	if p.prepared[issueID] == nil {
		p.prepared[issueID] = map[uuid.UUID]struct{}{}
	}
	p.prepared[issueID][runID] = struct{}{}
}

func (p *Provider) isPrepared(issueID, runID uuid.UUID) bool {
	p.preparedMu.RLock()
	defer p.preparedMu.RUnlock()
	_, ok := p.prepared[issueID][runID]
	return ok
}

func (p *Provider) hasPrepared(issueID uuid.UUID) bool {
	p.preparedMu.RLock()
	defer p.preparedMu.RUnlock()
	return len(p.prepared[issueID]) > 0
}

// clearPrepared drops this process's binding to a finished run. It is the one
// choke point every terminal path goes through, so it is also where the durable
// ownership claim is released: a run nothing is driving must not look owned.
func (p *Provider) clearPrepared(issueID, runID uuid.UUID) {
	p.preparedMu.Lock()
	delete(p.prepared[issueID], runID)
	if len(p.prepared[issueID]) == 0 {
		delete(p.prepared, issueID)
	}
	p.preparedMu.Unlock()
	p.releaseRun(context.Background(), runID)
}
