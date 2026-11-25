package agents

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/verify"
)

// startSessionTimeout bounds how long StartSession waits for a started run to report its session id.
const startSessionTimeout = 60 * time.Second

const sessionNameAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// StartSession launches req's prompt for its page with no threads, as a fresh
// managed session named <prefix>-<6 random characters>, and answers the
// session once the run reports its id. The session is recorded as managed in
// the page's scope, so a later delivery to its id resumes it with the threads.
func (d delivery) StartSession(ctx context.Context, req comments.StartSessionRequest) (comments.AgentSession, error) {
	if err := requireEdits(comments.DeliverTarget{Edits: req.Edits}); err != nil {
		return comments.AgentSession{}, err
	}
	name, err := sessionName(req.Prefix)
	if err != nil {
		return comments.AgentSession{}, err
	}
	prompt, group := strings.TrimSpace(req.Prompt), req.GroupKey
	if prompt == "" {
		prompt = verify.DefaultCommentAgentPrompt
	}
	if group == "" {
		group = req.Target
	}
	started := make(chan startOutcome, 1)
	_, err = d.start(ctx, sessionStart{
		Tenant: req.TenantID, Target: comments.Target{ScopeURI: req.ScopeURI, URI: req.Target}, Fresh: true, Started: started,
		Session: comments.ManagedSession{ScopeURI: req.ScopeURI, GroupKey: group, Prompt: prompt, RunOptions: req.RunOptions, Label: name},
		Run:     AgentRun{RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: name}, Edits: req.Edits},
	})
	if err != nil {
		return comments.AgentSession{}, err
	}
	wait, cancel := context.WithTimeout(ctx, startSessionTimeout)
	defer cancel()
	select {
	case outcome := <-started:
		if outcome.err != nil {
			return comments.AgentSession{}, fmt.Errorf("%w: session %s for %s: %v", comments.ErrDelivery, name, req.Target, outcome.err)
		}
		return d.startedSession(ctx, outcome.session)
	case <-wait.Done():
		return comments.AgentSession{}, fmt.Errorf("%w: session %s for %s reported no session id: %v", comments.ErrDelivery, name, req.Target, wait.Err())
	}
}

// startedSession lists a session StartSession recorded as the sessions listing
// does: live and managed, or ended and resumable once its first run is over.
func (d delivery) startedSession(ctx context.Context, managed comments.ManagedSession) (comments.AgentSession, error) {
	live, err := d.live(ctx, managed.ID)
	if err != nil {
		return comments.AgentSession{}, err
	}
	if live == nil {
		return endedSession(managed), nil
	}
	live.Managed, live.Title = true, managed.Label
	return *live, nil
}

// sessionName is prefix and a 6-character random lowercase alphanumeric suffix.
func sessionName(prefix string) (string, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "", fmt.Errorf("%w: a session name needs a non-blank prefix", comments.ErrInvalid)
	}
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return "", fmt.Errorf("generate a session name suffix: %w", err)
	}
	for i, b := range suffix {
		suffix[i] = sessionNameAlphabet[int(b)%len(sessionNameAlphabet)]
	}
	return prefix + "-" + string(suffix), nil
}
