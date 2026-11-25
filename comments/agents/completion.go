package agents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/comments"
)

type runRecorder struct {
	mu                 sync.Mutex
	delivery           delivery
	ctx                context.Context
	tenant             string
	threads            []comments.Comment
	run                AgentRun
	session            comments.ManagedSession
	reported, finished bool
	// started, when set, receives the reported session once it is recorded, or
	// why the run ended without reporting one; it is buffered for that one outcome.
	started chan<- startOutcome
}

// startOutcome is the session a started run reported, or why it reported none.
type startOutcome struct {
	session comments.ManagedSession
	err     error
}

func (r *runRecorder) signal(outcome startOutcome) {
	if r.started != nil {
		r.started <- outcome
		r.started = nil
	}
}

func (r *runRecorder) event(_ int, event api.Event) {
	if event.Kind != api.EventSystem || event.SessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reported || r.finished {
		return
	}
	r.reported = true
	r.run.SessionID, r.run.State = event.SessionID, "sent"
	now := time.Now().UTC()
	r.session.ID, r.session.Model, r.session.LastDeliveredAt = event.SessionID, event.Model, now
	if r.session.StartedAt.IsZero() {
		r.session.StartedAt = r.run.RequestedAt
	}
	if err := r.delivery.setRuns(r.ctx, r.threads, r.run); err != nil {
		logger.Errorf("recording session %s on %s: %v", event.SessionID, threadIDs(r.threads), err)
		r.signal(startOutcome{err: fmt.Errorf("recording session %s: %w", event.SessionID, err)})
		return
	}
	if err := r.delivery.Groups.RecordSession(r.ctx, r.tenant, r.session); err != nil {
		logger.Errorf("recording managed session %s: %v", event.SessionID, err)
		r.signal(startOutcome{err: fmt.Errorf("recording managed session %s: %w", event.SessionID, err)})
		return
	}
	r.signal(startOutcome{session: r.session})
}

func (r *runRecorder) complete(summary captaincli.PromptRunSummary) {
	if summary.Success && summary.SessionID != "" {
		r.event(0, api.Event{Kind: api.EventSystem, SessionID: summary.SessionID, Model: summary.Model})
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.finished {
		return
	}
	r.finished = true
	r.run.State, r.run.Error = "done", summary.Error
	if !summary.Success {
		r.run.State = "error"
		if r.run.Error == "" {
			r.run.Error = "agent run ended without a successful outcome"
		}
	}
	if r.run.SessionID == "" {
		r.run.SessionID = summary.SessionID
	}
	if !r.reported {
		r.signal(startOutcome{err: fmt.Errorf("agent run ended before reporting a session id: %s", r.run.Error)})
	}
	for _, thread := range r.threads {
		if _, err := setAgentRun(r.ctx, r.delivery.service, thread.ID, r.run); err != nil {
			logger.Errorf("recording completion of agent run on thread %s: %v", thread.ID, err)
		}
	}
}
