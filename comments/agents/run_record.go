package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
)

// AgentRun is a thread's hand-off to an agent session, kept under "agentRun" in
// the comment's content in clicky-ui's CommentAgentRun shape.
type AgentRun struct {
	SessionID   string         `json:"sessionId"`
	RequestedAt time.Time      `json:"requestedAt"`
	State       string         `json:"state"`
	Error       string         `json:"error,omitempty"`
	Target      AgentRunTarget `json:"target"`
}

// AgentRunTarget is where the thread went, and the channel captain delivered it through.
type AgentRunTarget struct {
	Label   string `json:"label"`
	Channel string `json:"channel,omitempty"`
}

// claimForAgent assigns the thread to agent as the requesting user, unless it
// already is, then acknowledges it as that agent, so the agent holds it while
// it works. Each step runs on the thread's latest version, so the creator's own
// follow-up changes (attachments) do not fail the claim.
func claimForAgent(ctx context.Context, service *comments.Service, id string, agent comments.Actor) (comments.Comment, error) {
	_, err := service.WithLatest(ctx, id, func(c comments.Comment) (comments.Comment, error) {
		if c.Assignee != nil && c.Assignee.Kind == agent.Kind && c.Assignee.ID == agent.ID {
			return c, nil
		}
		return service.Act(ctx, comments.ActionRequest{ID: id, ExpectedVersion: c.Version, Action: "assign", Assignee: &agent})
	})
	if err != nil {
		return comments.Comment{}, fmt.Errorf("assigning thread %s to %s: %w", id, agent.ID, err)
	}
	asAgent := commentapi.WithActor(ctx, agent)
	comment, err := service.WithLatest(asAgent, id, func(c comments.Comment) (comments.Comment, error) {
		return service.Act(asAgent, comments.ActionRequest{ID: id, ExpectedVersion: c.Version, Action: "acknowledge"})
	})
	if err != nil {
		return comments.Comment{}, fmt.Errorf("acknowledging thread %s as %s: %w", id, agent.ID, err)
	}
	return comment, nil
}

// setAgentRun records run under "agentRun" in the latest comment's content, keeping its other keys.
func setAgentRun(ctx context.Context, service *comments.Service, id string, run AgentRun) (comments.Comment, error) {
	encoded, err := json.Marshal(run)
	if err != nil {
		return comments.Comment{}, err
	}
	comment, err := service.WithLatest(ctx, id, func(c comments.Comment) (comments.Comment, error) {
		content := map[string]json.RawMessage{}
		if len(c.Content) > 0 && string(c.Content) != "null" {
			if err := json.Unmarshal(c.Content, &content); err != nil {
				return comments.Comment{}, fmt.Errorf("comment %s content is not a JSON object: %w", id, err)
			}
		}
		content["agentRun"] = encoded
		merged, err := json.Marshal(content)
		if err != nil {
			return comments.Comment{}, err
		}
		return service.Act(ctx, comments.ActionRequest{ID: id, ExpectedVersion: c.Version, Action: "metadata", Content: merged})
	})
	if err != nil {
		return comments.Comment{}, fmt.Errorf("recording agent run on comment %s: %w", id, err)
	}
	return comment, nil
}

// targetLabel names where target sends a thread before the session is known.
func targetLabel(target comments.DeliverTarget) string {
	if target.Kind == "run" {
		return "New " + target.Prompt + " session"
	}
	return target.SessionID
}
