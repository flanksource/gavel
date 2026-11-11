package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/google/uuid"
)

// RunOptions are a run target's runOptions in captain's run-request shape:
// Presets replace the prompt frontmatter's when set, Spec is layered over the
// frontmatter, Variables feed the template below the host's own, and
// Instructions are appended to the thread the session is given.
type RunOptions struct {
	Presets      []string       `json:"presets,omitempty"`
	Spec         *api.Spec      `json:"spec,omitempty"`
	Variables    map[string]any `json:"variables,omitempty"`
	Instructions string         `json:"instructions,omitempty"`
}

// HostVariables are the template variables every run sets; a run's own
// variables of the same name are overridden.
var HostVariables = []string{"dir", "scope", "target", "actor", "gavelExecutable", "message"}

func decodeRunOptions(raw json.RawMessage) (RunOptions, error) {
	var options RunOptions
	if len(raw) == 0 {
		return options, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&options); err != nil && !errors.Is(err, io.EOF) {
		return RunOptions{}, fmt.Errorf("%w: runOptions: %v", comments.ErrInvalid, err)
	}
	return options, nil
}

// sessionRequest is one thread's managed session: what it runs, as which agent, where.
type sessionRequest struct {
	Comment comments.Comment
	Dir     string
	Agent   comments.Actor
	Prompt  string
	Options RunOptions
}

// launch starts a managed session for the thread, recording the session it reports.
func (d delivery) launch(ctx context.Context, comment comments.Comment, target comments.DeliverTarget) (AgentRun, error) {
	run := AgentRun{RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: targetLabel(target)}}
	options, err := decodeRunOptions(target.RunOptions)
	if err != nil {
		return run, err
	}
	dir, err := d.Workdir(ctx, comment.Target)
	if err != nil {
		return run, err
	}
	agent := comments.Actor{Kind: "agent", ID: "gavel:" + uuid.NewString(), Label: "Gavel agent"}
	if _, err = claimForAgent(ctx, d.service, comment.ID, agent); err != nil {
		return run, err
	}
	if comment, err = setAgentRun(ctx, d.service, comment.ID, run); err != nil {
		return run, err
	}
	invocation, err := d.invocation(sessionRequest{Comment: comment, Dir: dir, Agent: agent, Prompt: target.Prompt, Options: options})
	if err != nil {
		return run, err
	}
	// The session outlives this request; it keeps the request's values (the
	// host's database, principal and prompt sources) but not its cancellation.
	background, err := promptRunContext(context.WithoutCancel(ctx), dir)
	if err != nil {
		return run, err
	}
	invocation.OnEvent = d.recordSession(background, comment, run, target.GroupKey)
	if _, err = d.Bridge.Launch(background, invocation); err != nil {
		return run, fmt.Errorf("starting an agent session for thread %s: %w", comment.ID, err)
	}
	return run, nil
}

// invocation is the prompt run a session starts (or previews): the chosen
// prompt with the host's variables over the request's, and the thread tools
// attached whatever the prompt declares.
func (d delivery) invocation(session sessionRequest) (captaincli.PromptInvocation, error) {
	gavel, err := os.Executable()
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("locate the gavel executable for the session: %w", err)
	}
	message, err := sessionMessage(session.Comment, session.Options.Instructions, d.BaseURL)
	if err != nil {
		return captaincli.PromptInvocation{}, err
	}
	vars := maps.Clone(session.Options.Variables)
	if vars == nil {
		vars = map[string]any{}
	}
	maps.Copy(vars, map[string]any{
		"dir": session.Dir, "scope": session.Comment.Target.ScopeURI, "target": session.Comment.Target.URI,
		"actor": commentapi.EncodeActor(session.Agent), "gavelExecutable": gavel, "message": message,
	})
	encoded, err := json.Marshal(vars)
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("encode the session's variables: %w", err)
	}
	options := captaincli.AIPromptOptions{Vars: string(encoded), Spec: session.Options.Spec}
	if session.Options.Presets != nil {
		options.Presets, options.PresetsSet = session.Options.Presets, true
	}
	prompt := strings.TrimSpace(session.Prompt)
	return captaincli.PromptInvocation{
		Reference: prompt, Name: prompt, Options: options,
		// The session's input is its variables; the host's own stdin is never the prompt's.
		Stdin: strings.NewReader(""),
		Tools: threadTools(d.service, session.Comment.Target, session.Agent),
	}, nil
}

// promptRunContext resolves prompts as captain does in dir (the project's
// .captain/prompts, ~/.captain/prompts and configured dirs) plus gavel's library.
func promptRunContext(ctx context.Context, dir string) (context.Context, error) {
	ctx, err := captaincli.PromptContext(captaincli.ContextWithPromptSources(ctx, PromptSources), dir)
	if err != nil {
		return nil, fmt.Errorf("resolve prompts in %s: %w", dir, err)
	}
	return ctx, nil
}

// recordSession stores the session the run reports first as the thread's
// agentRun (sent), and as its group's session, so later threads reuse it.
func (d delivery) recordSession(ctx context.Context, comment comments.Comment, run AgentRun, group string) func(int, api.Event) {
	var once sync.Once
	return func(_ int, event api.Event) {
		if event.Kind != api.EventSystem || event.SessionID == "" {
			return
		}
		once.Do(func() {
			run.SessionID, run.State = event.SessionID, "sent"
			_, err := setAgentRun(ctx, d.service, comment.ID, run)
			if err == nil && group != "" {
				err = d.Groups.RecordGroupSession(ctx, comment.TenantID, comment.Target.ScopeURI, group, event.SessionID)
			}
			if err != nil {
				// No request is left to answer: the run reports its session after the launch returned.
				logger.Errorf("recording session %s on comment thread %s: %v", event.SessionID, comment.ID, err)
			}
		})
	}
}
