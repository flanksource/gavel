package agents

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/session/amp"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/comments"
)

// AmpThreads is the outbound HTTP channel for existing Amp Orb threads.
type AmpThreads interface {
	List(context.Context) ([]amp.Thread, error)
	OwnedOrb(context.Context, string) (amp.Thread, error)
	Send(ctx context.Context, id, message string) (string, error)
}

func AmpFromEnvironment() AmpThreads {
	if token := strings.TrimSpace(os.Getenv("AMP_API_KEY")); token != "" {
		return amp.New(token)
	}
	return nil
}

func (d delivery) resolveAmp(ctx context.Context, id string) (comments.AgentSession, error) {
	if d.Amp == nil {
		return comments.AgentSession{}, fmt.Errorf("set AMP_API_KEY on the Gavel host to use Amp thread delivery")
	}
	thread, err := d.Amp.OwnedOrb(ctx, id)
	if err != nil {
		return comments.AgentSession{}, err
	}
	return ampSession(thread), nil
}

func (d delivery) ampSessions(ctx context.Context, query comments.SessionQuery) []comments.AgentSession {
	if d.Amp == nil || query.GroupKey != "" || query.SessionID != "" {
		return nil
	}
	var repos []string
	if query.ScopeURI != "" && !query.All {
		var err error
		repos, err = d.Repositories(ctx, query.ScopeURI)
		if err != nil {
			logger.Warnf("resolving repositories for Amp thread discovery: %v", err)
			return nil
		}
		if len(repos) == 0 {
			return nil
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	threads, err := d.Amp.List(ctx)
	if err != nil {
		logger.Warnf("listing Amp threads: %v", err)
		return nil
	}
	return ampSessionsFor(threads, repos)
}

func ampSessionsFor(threads []amp.Thread, repos []string) []comments.AgentSession {
	allowed := map[string]bool{}
	for _, repo := range repos {
		allowed[strings.ToLower(repo)] = true
	}
	seen := map[string]bool{}
	var sessions []comments.AgentSession
	for _, thread := range threads {
		if !thread.CanReceive() || seen[thread.ID] || !matchesAmpRepo(thread.Repositories, allowed) {
			continue
		}
		if id, err := amp.ParseThreadRef(thread.ID); err != nil || id != thread.ID {
			continue
		}
		seen[thread.ID] = true
		sessions = append(sessions, ampSession(thread))
	}
	return sessions
}

func matchesAmpRepo(repositories []amp.Repository, allowed map[string]bool) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, repository := range repositories {
		uri, err := url.Parse(repository.URL)
		if err == nil && strings.EqualFold(uri.Hostname(), "github.com") && allowed[strings.ToLower(strings.TrimSuffix(strings.Trim(uri.Path, "/"), ".git"))] {
			return true
		}
	}
	return false
}

func ampSession(thread amp.Thread) comments.AgentSession {
	title := thread.Title
	if title == "" {
		title = thread.ID
	}
	return comments.AgentSession{ID: thread.ID, Title: title, Source: "amp", Status: "available"}
}

func (d delivery) toAmp(ctx context.Context, threads []comments.Comment, id, edits string) (AgentRun, error) {
	run := AgentRun{SessionID: id, RequestedAt: time.Now().UTC(), Target: AgentRunTarget{Label: id, Channel: "amp-http"}, Edits: edits}
	session, err := d.resolveAmp(ctx, id)
	if err != nil {
		return run, err
	}
	run.Target.Label = session.Title
	message, err := threadsMarkdown(threads, d.BaseURL)
	if err != nil {
		return run, err
	}
	if edits == comments.EditsPropose {
		message += "\n\nPropose a patch in this Amp thread. Leave the repository files unchanged."
	}
	messageID, err := d.Amp.Send(ctx, id, message)
	if err != nil {
		return run, fmt.Errorf("delivering %s to Amp: %w", threadIDs(threads), err)
	}
	run.State, run.MessageID = "sent", messageID
	return run, d.setRuns(context.WithoutCancel(ctx), threads, run)
}
