package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/gavel/comments"
)

// toSession delivers threads into session id: live, or an ended managed
// session of their scope, which is resumed. Any other session is an error.
func (d delivery) toSession(ctx context.Context, threads []comments.Comment, id string) (AgentRun, error) {
	run := AgentRun{SessionID: id, RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: id}}
	live, err := d.live(ctx, id)
	if err != nil {
		return run, err
	}
	if live != nil {
		return d.toLive(ctx, threads, *live)
	}
	managed, err := d.managed(ctx, threads[0], id)
	if err != nil {
		return run, err
	}
	if managed == nil {
		return run, fmt.Errorf("agent session %s is no longer live; choose another", id)
	}
	return d.resume(ctx, threads, *managed)
}

// toLive hands threads to a live session as its agent (the user assigns, the
// agent acknowledges) and delivers them into it in one message.
func (d delivery) toLive(ctx context.Context, threads []comments.Comment, session comments.AgentSession) (AgentRun, error) {
	run := AgentRun{SessionID: session.ID, RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: session.Title}}
	agent := comments.Actor{Kind: "agent", ID: session.Source + ":" + session.ID, Label: session.Model}
	if agent.Label == "" {
		agent.Label = session.Source
	}
	threads, err := d.claim(ctx, threads, agent, run)
	if err != nil {
		return run, err
	}
	message, err := threadMessage(threads, agent, d.BaseURL)
	if err != nil {
		return run, err
	}
	result, err := d.Bridge.Deliver(ctx, session.ID, message)
	if err != nil {
		return run, fmt.Errorf("delivering %s to session %s: %w", threadIDs(threads), session.ID, err)
	}
	run.State, run.Target.Channel = string(result.Status), string(result.Channel)
	lead := threads[0]
	if err = d.Groups.TouchSession(ctx, lead.TenantID, lead.Target.ScopeURI, session.ID, time.Now().UTC()); err != nil {
		return run, fmt.Errorf("recording the delivery into session %s: %w", session.ID, err)
	}
	return run, d.setRuns(ctx, threads, run)
}

// claim assigns every thread to agent and records run as its agentRun,
// answering the threads as they now are.
func (d delivery) claim(ctx context.Context, threads []comments.Comment, agent comments.Actor, run AgentRun) ([]comments.Comment, error) {
	claimed := make([]comments.Comment, 0, len(threads))
	for _, thread := range threads {
		if _, err := claimForAgent(ctx, d.service, thread.ID, agent); err != nil {
			return nil, err
		}
		updated, err := setAgentRun(ctx, d.service, thread.ID, run)
		if err != nil {
			return nil, err
		}
		claimed = append(claimed, updated)
	}
	return claimed, nil
}

// setRuns records run as every thread's agentRun.
func (d delivery) setRuns(ctx context.Context, threads []comments.Comment, run AgentRun) error {
	for _, thread := range threads {
		if _, err := setAgentRun(ctx, d.service, thread.ID, run); err != nil {
			return err
		}
	}
	return nil
}
