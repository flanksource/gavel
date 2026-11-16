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
// propose is true when the delivery's edits policy is comments.EditsPropose.
var HostVariables = []string{"dir", "scope", "target", "actor", "gavelExecutable", "message", "propose"}

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

// sessionRequest is a managed session for threads (one scope's, the first
// leading): what it runs, as which agent, where.
type sessionRequest struct {
	Threads []comments.Comment
	Dir     string
	Agent   comments.Actor
	Prompt  string
	Options RunOptions
	// Edits is the delivery's edits policy: comments.EditsDirect or EditsPropose.
	Edits string
}

// launch starts a new managed session for the threads.
func (d delivery) launch(ctx context.Context, threads []comments.Comment, target comments.DeliverTarget) (AgentRun, error) {
	run := AgentRun{RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: targetLabel(target)}, Edits: target.Edits}
	session := comments.ManagedSession{ScopeURI: threads[0].Target.ScopeURI, GroupKey: target.GroupKey, Prompt: target.Prompt, RunOptions: target.RunOptions, Label: firstLine(threads[0].Body)}
	return d.start(ctx, threads, session, run)
}

// resume restarts an ended managed session with its recorded prompt and run
// options, resuming its conversation with the threads under this delivery's edits policy.
func (d delivery) resume(ctx context.Context, threads []comments.Comment, session comments.ManagedSession, edits string) (AgentRun, error) {
	run := AgentRun{SessionID: session.ID, RequestedAt: time.Now().UTC(), State: "queued", Target: AgentRunTarget{Label: session.Label}, Edits: edits}
	return d.start(ctx, threads, session, run)
}

// start claims the threads for a new gavel agent and launches session's prompt
// for them, resuming session.ID when set, recording the session it reports.
func (d delivery) start(ctx context.Context, threads []comments.Comment, session comments.ManagedSession, run AgentRun) (AgentRun, error) {
	options, err := decodeRunOptions(session.RunOptions)
	if err != nil {
		return run, err
	}
	dir, err := d.Workdir(ctx, threads[0].Target)
	if err != nil {
		return run, err
	}
	agent := comments.Actor{Kind: "agent", ID: "gavel:" + uuid.NewString(), Label: "Gavel agent"}
	if threads, err = d.claim(ctx, threads, agent, run); err != nil {
		return run, err
	}
	invocation, err := d.invocation(sessionRequest{Threads: threads, Dir: dir, Agent: agent, Prompt: session.Prompt, Options: options, Edits: run.Edits})
	if err != nil {
		return run, err
	}
	invocation.Options.Resume = session.ID
	// The session outlives this request; it keeps the request's values (the
	// host's database, principal and prompt sources) but not its cancellation.
	background, err := promptRunContext(context.WithoutCancel(ctx), dir)
	if err != nil {
		return run, err
	}
	session.CWD = dir
	invocation.OnEvent = d.recordSession(background, threads, run, session)
	if _, err = d.Bridge.Launch(background, invocation); err != nil {
		return run, fmt.Errorf("starting an agent session for %s: %w", threadIDs(threads), err)
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
	message, err := sessionMessage(session.Threads, messageOptions{BaseURL: d.BaseURL, Edits: session.Edits, Dir: session.Dir, Instructions: session.Options.Instructions})
	if err != nil {
		return captaincli.PromptInvocation{}, err
	}
	vars := maps.Clone(session.Options.Variables)
	if vars == nil {
		vars = map[string]any{}
	}
	lead := session.Threads[0].Target
	maps.Copy(vars, map[string]any{
		"dir": session.Dir, "scope": lead.ScopeURI, "target": lead.URI,
		"actor": commentapi.EncodeActor(session.Agent), "gavelExecutable": gavel, "message": message,
		"propose": session.Edits == comments.EditsPropose,
	})
	encoded, err := json.Marshal(vars)
	if err != nil {
		return captaincli.PromptInvocation{}, fmt.Errorf("encode the session's variables: %w", err)
	}
	options := captaincli.AIPromptOptions{Vars: string(encoded), Spec: runSpec(session.Options.Spec, session.Edits)}
	if session.Options.Presets != nil {
		options.Presets, options.PresetsSet = session.Options.Presets, true
	}
	prompt := strings.TrimSpace(session.Prompt)
	return captaincli.PromptInvocation{
		Reference: prompt, Name: prompt, Options: options,
		// The session's input is its variables; the host's own stdin is never the prompt's.
		Stdin: strings.NewReader(""),
		Tools: threadTools(d.service, session.Threads, session.Agent, session.Edits),
	}, nil
}

// proposeDeniedTools are the file-writing tools a propose-mode run never gets:
// gavel commits its edits from propose_change itself.
var proposeDeniedTools = []string{"Edit", "Write", "MultiEdit", "NotebookEdit"}

// runSpec is the request layer a session runs with over its prompt's
// frontmatter: the run's own spec, with proposeDeniedTools denied in propose
// mode (a request layer's tool keys win over the frontmatter's).
func runSpec(spec *api.Spec, edits string) *api.Spec {
	if edits != comments.EditsPropose {
		return spec
	}
	deny := api.Spec{Permissions: api.Permissions{Tools: api.Tools{}}}
	for _, tool := range proposeDeniedTools {
		deny.Permissions.Tools[tool] = api.ToolPolicyDeny
	}
	if spec == nil {
		return &deny
	}
	merged := spec.Merge(deny)
	return &merged
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

// recordSession stores the session the run reports first as every thread's
// agentRun (sent), and as a managed session of their scope (its group's when
// session has a GroupKey), so it is listed after it ends and later threads
// reuse or resume it.
func (d delivery) recordSession(ctx context.Context, threads []comments.Comment, run AgentRun, session comments.ManagedSession) func(int, api.Event) {
	var once sync.Once
	return func(_ int, event api.Event) {
		if event.Kind != api.EventSystem || event.SessionID == "" {
			return
		}
		once.Do(func() {
			run.SessionID, run.State = event.SessionID, "sent"
			now := time.Now().UTC()
			session.ID, session.Model, session.StartedAt, session.LastDeliveredAt = event.SessionID, event.Model, now, now
			err := d.setRuns(ctx, threads, run)
			if err == nil {
				err = d.Groups.RecordSession(ctx, threads[0].TenantID, session)
			}
			if err != nil {
				// No request is left to answer: the run reports its session after the launch returned.
				logger.Errorf("recording session %s on %s: %v", event.SessionID, threadIDs(threads), err)
			}
		})
	}
}

// threadIDs names the threads in a message: "thread <id>" or "threads <id>, <id>".
func threadIDs(threads []comments.Comment) string {
	ids := make([]string, 0, len(threads))
	for _, thread := range threads {
		ids = append(ids, thread.ID)
	}
	if len(ids) == 1 {
		return "thread " + ids[0]
	}
	return "threads " + strings.Join(ids, ", ")
}

func firstLine(text string) string {
	return strings.TrimSpace(strings.SplitN(strings.TrimSpace(text), "\n", 2)[0])
}
