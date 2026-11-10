package editor

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strings"
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

func (b AgentBridge) render(ctx context.Context, invocation captaincli.PromptInvocation) (captaincli.PromptRenderResult, error) {
	if b.Render != nil {
		return b.Render(ctx, invocation)
	}
	if err := openCaptain(ctx); err != nil {
		return captaincli.PromptRenderResult{}, sendError{http.StatusServiceUnavailable, err}
	}
	return captaincli.RenderPrompt(ctx, invocation)
}

// permissionCatalog serves GET /api/file/{id}/agent-run/permissions?provider&mode:
// captain's permission catalog (tools, MCP servers, plugins, skills) for that
// runtime, rooted at the project holding the file, wherever the editor was started.
func (s *Server) permissionCatalog(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if err := openCaptain(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	captaincli.PermissionCatalogHandler(target.Dir).ServeHTTP(w, r)
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
	options, ok := decodeRunOptions(w, r)
	if !ok {
		return
	}
	comment, err := s.runThread(r.Context(), runRequest{FileID: r.PathValue("id"), CommentID: r.PathValue("cid"), Options: options})
	var failed sendError
	if errors.As(err, &failed) {
		http.Error(w, failed.Error(), failed.status)
		return
	}
	respond(w, comment, err)
}

// RunOptions are the new-session dialog's choices, in captain's run-request
// shape: Prompt is the discovered prompt to run (default editor-comments),
// Presets replace the frontmatter's when set, Spec is layered over the
// frontmatter, Variables feed the template below the host's own, and
// Instructions are appended to the thread the session is given.
type RunOptions struct {
	Prompt       string         `json:"prompt,omitempty"`
	Presets      []string       `json:"presets,omitempty"`
	Spec         *api.Spec      `json:"spec,omitempty"`
	Variables    map[string]any `json:"variables,omitempty"`
	Instructions string         `json:"instructions,omitempty"`
}

// decodeRunOptions reads a run request body; an empty body is the defaults.
func decodeRunOptions(w http.ResponseWriter, r *http.Request) (RunOptions, bool) {
	var options RunOptions
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&options); err != nil && !errors.Is(err, io.EOF) {
		http.Error(w, fmt.Sprintf("invalid run request: %v", err), http.StatusBadRequest)
		return RunOptions{}, false
	}
	return options, true
}

type runRequest struct {
	FileID, CommentID string
	Options           RunOptions
}

func (s *Server) runThread(ctx context.Context, req runRequest) (comments.Comment, error) {
	doc, service, comment, err := s.fileComment(ctx, req.FileID, req.CommentID)
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
	session := sessionRequest{Service: service, ThreadID: comment.ID, Target: target, Agent: agent, Message: sessionMessage(comment, doc.Path, req.Options.Instructions), Options: req.Options}
	launchErr := s.startSession(ctx, session, run)
	if launchErr == nil {
		return comment, nil
	}
	run.State, run.Error = "error", launchErr.Error()
	if _, err := setAgentRun(ctx, service, comment, run); err != nil {
		return comments.Comment{}, fmt.Errorf("record failed launch of %s: %w (launch: %v)", comment.ID, err, launchErr)
	}
	return comments.Comment{}, sendError{http.StatusBadGateway, fmt.Errorf("starting an agent session for thread %s: %w", comment.ID, launchErr)}
}

// startSession launches the session's prompt, recording the session it reports on run.
func (s *Server) startSession(ctx context.Context, session sessionRequest, run AgentRun) error {
	invocation, err := sessionInvocation(session)
	if err != nil {
		return err
	}
	// The session outlives this request; it keeps the request's values (the editor's
	// database and prompt sources) but not its cancellation.
	background, err := promptRunContext(context.WithoutCancel(ctx), session.Target.Dir)
	if err != nil {
		return err
	}
	invocation.OnEvent = recordSession(background, session.Service, session.ThreadID, run)
	_, err = s.Agents.launch(background, invocation)
	return err
}

// hostVariables are the template variables the editor sets on every run; a
// request's own variables of the same name are overridden.
var hostVariables = []string{"dir", "scope", "target", "actor", "gavelExecutable", "message"}

// sessionRequest is one thread's session: what it runs, as which agent, on which file.
type sessionRequest struct {
	Service  *comments.Service
	ThreadID string
	Target   fileTarget
	Agent    comments.Actor
	Message  string
	Options  RunOptions
}

func sessionMessage(comment comments.Comment, path, instructions string) string {
	message := threadMarkdown(comment, path)
	if instructions = strings.TrimSpace(instructions); instructions != "" {
		message += "\n\n## Additional instructions\n\n" + instructions + "\n"
	}
	return message
}

// sessionInvocation is the prompt run a session starts (or previews): the
// chosen prompt with the host's variables over the request's, and the thread
// tools attached whatever the prompt declares.
func sessionInvocation(session sessionRequest) (captaincli.PromptInvocation, error) {
	gavel, err := os.Executable()
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("locate the gavel executable for the session: %w", err)
	}
	vars := maps.Clone(session.Options.Variables)
	if vars == nil {
		vars = map[string]any{}
	}
	maps.Copy(vars, map[string]any{
		"dir": session.Target.Dir, "scope": session.Target.Scope, "target": session.Target.URI,
		"actor": commentapi.EncodeActor(session.Agent), "gavelExecutable": gavel, "message": session.Message,
	})
	encoded, err := json.Marshal(vars)
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("encode the session's variables: %w", err)
	}
	options := captaincli.AIPromptOptions{Vars: string(encoded), Spec: session.Options.Spec}
	if session.Options.Presets != nil {
		options.Presets, options.PresetsSet = session.Options.Presets, true
	}
	prompt := cmp.Or(strings.TrimSpace(session.Options.Prompt), managedPrompt)
	return captaincli.PromptInvocation{
		Reference: prompt, Name: prompt, Options: options,
		// The session's input is its variables; the editor's own stdin is never the prompt's.
		Stdin: strings.NewReader(""),
		Tools: threadTools(session.Service, session.Target, session.Agent),
	}, nil
}

// promptRunContext resolves prompts as captain does in dir (the project's
// .captain/prompts, ~/.captain/prompts and configured dirs) plus gavel's library.
func promptRunContext(ctx context.Context, dir string) (context.Context, error) {
	ctx, err := captaincli.PromptContext(captaincli.ContextWithPromptSources(ctx, librarySources), dir)
	if err != nil {
		return nil, fmt.Errorf("resolve prompts in project %s: %w", dir, err)
	}
	return ctx, nil
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
