package editor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	promptlibrary "github.com/flanksource/gavel/prompts/library"
	"github.com/google/uuid"
)

// managedPrompt is the library prompt a new editor agent session runs.
const managedPrompt = "editor-comments"

// librarySources scopes captain's prompt resolution to gavel's prompt library.
var librarySources = captaincli.PromptSources{FS: []captaincli.PromptFS{{Label: promptlibrary.Label, FS: promptlibrary.FS}}}

func (b AgentBridge) launch(ctx context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRunResult, error) {
	if b.Launch != nil {
		return b.Launch(ctx, invocation)
	}
	if err := openCaptain(ctx); err != nil {
		return captaincli.PromptRunResult{}, err
	}
	return captaincli.LaunchPrompt(ctx, invocation)
}

// captainPromptRuns serves captain's prompt-run routes (snapshot, stream, stop)
// for the runs the editor launched, once gavel's pool is captain's.
func captainPromptRuns() http.Handler {
	runs := http.StripPrefix("/api/captain/prompt/runs", captaincli.PromptRunHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := openCaptain(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		runs.ServeHTTP(w, r)
	})
}

// runComment serves POST /api/file/{id}/comments/{cid}/run: it starts a new
// managed agent session in the file's project, whose tools reply to and resolve
// the thread through gavel comments, and records the hand-off as agentRun.
func (s *Server) runComment(w http.ResponseWriter, r *http.Request) {
	comment, err := s.runThread(r.Context(), r.PathValue("id"), r.PathValue("cid"))
	var failed sendError
	if errors.As(err, &failed) {
		http.Error(w, failed.Error(), failed.status)
		return
	}
	respond(w, comment, err)
}

func (s *Server) runThread(ctx context.Context, fileID, commentID string) (comments.Comment, error) {
	doc, service, comment, err := s.fileComment(ctx, fileID, commentID)
	if err != nil {
		return comments.Comment{}, err
	}
	target, err := targetOf(ctx, doc.Path)
	if err != nil {
		return comments.Comment{}, err
	}
	agent := comments.Actor{Kind: "agent", ID: "gavel-edit:" + uuid.NewString(), Label: "Editor agent"}
	if comment, err = claimForAgent(ctx, service, comment, agent); err != nil {
		return comments.Comment{}, err
	}
	run := AgentRun{State: "queued", RequestedAt: time.Now().UTC(), Target: AgentRunTarget{Label: "New session in " + target.Project}}
	if comment, err = setAgentRun(ctx, service, comment, run); err != nil {
		return comments.Comment{}, err
	}
	invocation, err := managedInvocation(target, agent, threadMarkdown(comment, doc.Path))
	if err != nil {
		return comments.Comment{}, err
	}
	// The session outlives this request; it keeps the request's values (the editor's
	// database and prompt sources) but not its cancellation.
	background := captaincli.ContextWithPromptSources(context.WithoutCancel(ctx), librarySources)
	invocation.OnEvent = recordSession(background, service, comment.ID, run)
	_, launchErr := s.Agents.launch(background, invocation)
	if launchErr == nil {
		return comment, nil
	}
	run.State, run.Error = "error", launchErr.Error()
	if _, err := setAgentRun(ctx, service, comment, run); err != nil {
		return comments.Comment{}, fmt.Errorf("record failed launch of %s: %w (launch: %v)", comment.ID, err, launchErr)
	}
	return comments.Comment{}, sendError{http.StatusBadGateway, fmt.Errorf("starting an agent session for thread %s: %w", comment.ID, launchErr)}
}

func managedInvocation(target fileTarget, agent comments.Actor, thread string) (captaincli.PromptInvocation, error) {
	gavel, err := os.Executable()
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("locate the gavel executable for the session's tools: %w", err)
	}
	vars, err := json.Marshal(map[string]string{
		"dir": target.Dir, "scope": target.Scope, "target": target.URI,
		"actor": commentapi.EncodeActor(agent), "gavelExecutable": gavel, "message": thread,
	})
	if err != nil {
		return captaincli.PromptInvocation{}, err
	}
	return captaincli.PromptInvocation{Reference: managedPrompt, Name: managedPrompt, Options: captaincli.AIPromptOptions{Vars: string(vars)}}, nil
}

// recordSession stores the session the run reports first as the thread's
// agentRun (sent), so the thread can show it and later threads can be sent to it.
func recordSession(ctx context.Context, service *comments.Service, commentID string, run AgentRun) func(int, api.Event) {
	var once sync.Once
	return func(_ int, event api.Event) {
		if event.Kind != api.EventSystem || event.SessionID == "" {
			return
		}
		once.Do(func() {
			comment, err := service.Get(ctx, commentID)
			if err == nil {
				run.SessionID, run.State = event.SessionID, "sent"
				_, err = setAgentRun(ctx, service, comment, run)
			}
			if err != nil {
				// No request is left to answer: the run reports its session after the launch returned.
				logger.Errorf("recording session %s on comment thread %s: %v", event.SessionID, commentID, err)
			}
		})
	}
}
