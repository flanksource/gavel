package agents

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/gavel/comments"
)

// toSession hands the thread to a live session as its agent (the user assigns,
// the agent acknowledges) and delivers the thread into it.
func (d delivery) toSession(ctx context.Context, comment comments.Comment, sessionID string) (AgentRun, error) {
	run := AgentRun{SessionID: sessionID, RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: sessionID}}
	session, err := d.live(ctx, sessionID)
	if err != nil {
		return run, err
	}
	if session == nil {
		return run, fmt.Errorf("agent session %s is no longer live; choose another", sessionID)
	}
	run.Target.Label = session.Title
	agent := comments.Actor{Kind: "agent", ID: session.Source + ":" + session.ID, Label: session.Model}
	if agent.Label == "" {
		agent.Label = session.Source
	}
	if _, err = claimForAgent(ctx, d.service, comment.ID, agent); err != nil {
		return run, err
	}
	if comment, err = setAgentRun(ctx, d.service, comment.ID, run); err != nil {
		return run, err
	}
	message, err := threadMessage(comment, agent, d.BaseURL)
	if err != nil {
		return run, err
	}
	result, err := d.Bridge.Deliver(ctx, session.ID, message)
	if err != nil {
		return run, fmt.Errorf("delivering thread %s to session %s: %w", comment.ID, session.ID, err)
	}
	run.State, run.Target.Channel = string(result.Status), string(result.Channel)
	_, err = setAgentRun(ctx, d.service, comment.ID, run)
	return run, err
}
