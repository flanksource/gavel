package agents

import (
	"context"
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
	threads            []comments.Comment
	run                AgentRun
	session            comments.ManagedSession
	reported, finished bool
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
		return
	}
	if err := r.delivery.Groups.RecordSession(r.ctx, r.threads[0].TenantID, r.session); err != nil {
		logger.Errorf("recording managed session %s: %v", event.SessionID, err)
	}
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
	for _, thread := range r.threads {
		if _, err := setAgentRun(r.ctx, r.delivery.service, thread.ID, r.run); err != nil {
			logger.Errorf("recording completion of agent run on thread %s: %v", thread.ID, err)
		}
	}
}
