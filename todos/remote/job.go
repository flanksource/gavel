// The remote run wire format: a Job carries one step's prompt run to a sandbox job, and Messages
// carry its start, events, verification progress and result back to serve, all as JSON.
package remote

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos/lifecycle"
)

// Job is what a sandbox job needs to run one step's prompt run: the resolved spec, the run's
// settings, and the data gavel's hooks are rebuilt from. Root is the checkout the spec's paths
// were resolved under on serve; the job rewrites them under its own checkout.
type Job struct {
	Resolved          api.ResolvedSpec   `json:"resolved"`
	Config            JobConfig          `json:"config"`
	Hooks             lifecycle.RunHooks `json:"hooks"`
	CallerOwnsCommits bool               `json:"callerOwnsCommits,omitempty"`
	VerifyTimeout     time.Duration      `json:"verifyTimeout,omitempty"`
	Timeout           time.Duration      `json:"timeout,omitempty"`
	Root              string             `json:"root"`
}

// JobConfig is the part of the provider configuration a job carries. Credentials stay out: the
// job reads its model keys from its own environment.
type JobConfig struct {
	Model     api.Model  `json:"model"`
	Budget    api.Budget `json:"budget"`
	NoCache   bool       `json:"noCache,omitempty"`
	SessionID string     `json:"sessionId,omitempty"`
	APIKey    string     `json:"-"`
}

// JobFor is the job that runs run in a sandbox.
func JobFor(run lifecycle.RemoteRun) Job {
	in := run.Input
	return Job{
		Resolved: in.Resolved,
		Config: JobConfig{
			Model: in.Config.Model, Budget: in.Config.Budget, NoCache: in.Config.NoCache, SessionID: in.Config.SessionID,
		},
		Hooks:             run.Hooks,
		CallerOwnsCommits: in.CallerOwnsCommits,
		VerifyTimeout:     in.Verify.Timeout,
		Timeout:           in.Timeout,
		Root:              in.Repo,
	}
}

// Input is the prompt run the job describes, with its paths under dir, the sandbox's checkout.
// It records nothing, and carries no hooks: the caller adds gavel's from Hooks.
func (j Job) Input(dir string) promptrun.Input {
	resolved := j.Resolved
	if cwd := resolved.Spec.Cwd(); cwd != "" {
		resolved.Spec.SetCwd(underRoot(j.Root, dir, cwd))
	}
	return promptrun.Input{
		Resolved: resolved,
		Config: captainai.Config{
			Model: j.Config.Model, Budget: j.Config.Budget, NoCache: j.Config.NoCache, SessionID: j.Config.SessionID,
		},
		CallerOwnsCommits: j.CallerOwnsCommits,
		Verify:            capverify.Options{Timeout: j.VerifyTimeout},
		Timeout:           j.Timeout,
		Repo:              dir,
	}
}

// HooksIn is the job's hook data with its paths under dir, the sandbox's checkout.
func (j Job) HooksIn(dir string) lifecycle.RunHooks {
	hooks := j.Hooks
	if hooks.PR != nil {
		pr := *hooks.PR
		pr.WorkDir = underRoot(j.Root, dir, pr.WorkDir)
		hooks.PR = &pr
	}
	return hooks
}

// underRoot moves path from under root to under dir. A path outside root moves to dir itself.
func underRoot(root, dir, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		return dir
	}
	return filepath.Join(dir, rel)
}

// Message kinds, in the order a run sends them: started once its setup is done, then events and
// verification progress as they happen, and its result last. A tool call and an approval
// request are calls: the run waits for serve's reply.
const (
	KindStarted  = "started"
	KindEvent    = "event"
	KindProgress = "progress"
	KindResult   = "result"
	KindTool     = "tool"
	KindApproval = "approval"
)

// Message is one thing a run reports to serve.
type Message struct {
	Kind      string            `json:"kind"`
	Iteration int               `json:"iteration,omitempty"`
	Spec      *api.Spec         `json:"spec,omitempty"`
	Event     *api.Event        `json:"event,omitempty"`
	Progress  *api.VerifyReport `json:"progress,omitempty"`
	Result    *Result           `json:"result,omitempty"`
	// Error is the run's error, with its result; empty when the run completed.
	Error string `json:"error,omitempty"`
	// Tool and Input are a call of one of the run's tools, which serve runs.
	Tool  string         `json:"tool,omitempty"`
	Input map[string]any `json:"input,omitempty"`
	// Approval is a tool use the run asks a person to approve.
	Approval *api.ApprovalRequest `json:"approval,omitempty"`

	// reply is where serve answers a call; it never leaves serve.
	reply chan Reply
}

// Reply is serve's answer to a call: a tool's output, or the decision on an approval.
type Reply struct {
	Output   json.RawMessage       `json:"output,omitempty"`
	Decision *api.ApprovalDecision `json:"decision,omitempty"`
	Error    string                `json:"error,omitempty"`
}

// Answer replies to a call.
func (m Message) Answer(reply Reply) {
	m.reply <- reply
}

// ToolCallMessage calls the run's tool with input.
func ToolCallMessage(tool string, input map[string]any) Message {
	return Message{Kind: KindTool, Tool: tool, Input: input}
}

// ApprovalMessage asks for a person's approval of request.
func ApprovalMessage(request api.ApprovalRequest) Message {
	return Message{Kind: KindApproval, Approval: &request}
}

// DecodedEvent is the event an event message carries.
func (m Message) DecodedEvent() api.Event {
	if m.Event == nil {
		return api.Event{}
	}
	return decodedEvent(*m.Event)
}

// decodedEvent undoes what JSON does to an event: an absent structured output decodes as the
// literal null, which reads as output that is there.
func decodedEvent(event api.Event) api.Event {
	if string(event.StructuredData) == "null" {
		event.StructuredData = nil
	}
	return event
}

// StartedMessage reports that the run's setup is done and its first turn is next.
func StartedMessage(spec api.Spec) Message {
	return Message{Kind: KindStarted, Spec: &spec}
}

// EventMessage reports one of the run's events, without its backend-native value.
func EventMessage(iteration int, event api.Event) Message {
	event.Raw = nil
	return Message{Kind: KindEvent, Iteration: iteration, Event: &event}
}

// ProgressMessage reports a verification snapshot.
func ProgressMessage(report api.VerifyReport) Message {
	return Message{Kind: KindProgress, Progress: &report}
}

// ResultMessage reports how the run ended.
func ResultMessage(result promptrun.Result, err error) Message {
	message := Message{Kind: KindResult, Result: wireResult(result)}
	if err != nil {
		message.Error = err.Error()
	}
	return message
}

// Result is a promptrun.Result as it crosses the wire: backend-native values dropped, and each
// loop iteration's error carried as its message.
type Result struct {
	promptrun.Result
	Loop *Loop `json:"Loop,omitempty"`
}

// Loop is a captainai.LoopResult as it crosses the wire.
type Loop struct {
	Iterations []LoopIteration `json:"iterations,omitempty"`
	TotalCost  float64         `json:"totalCost,omitempty"`
	StopReason string          `json:"stopReason,omitempty"`
}

// LoopIteration is a captainai.LoopIteration as it crosses the wire.
type LoopIteration struct {
	captainai.LoopIteration
	Err string `json:"Err,omitempty"`
}

func wireResult(result promptrun.Result) *Result {
	if result.Response != nil {
		response := *result.Response
		response.Raw = nil
		result.Response = &response
	}
	out := &Result{Result: result}
	out.Result.Loop = nil
	if result.Loop == nil {
		return out
	}
	out.Loop = &Loop{TotalCost: result.Loop.TotalCost, StopReason: result.Loop.StopReason}
	for _, iteration := range result.Loop.Iterations {
		if iteration == nil {
			continue
		}
		wire := LoopIteration{LoopIteration: *iteration}
		wire.LoopIteration.Err = nil
		if iteration.Err != nil {
			wire.Err = iteration.Err.Error()
		}
		if iteration.Events != nil {
			wire.Events = make([]api.Event, len(iteration.Events))
			for i, event := range iteration.Events {
				event.Raw = nil
				wire.Events[i] = event
			}
		}
		out.Loop.Iterations = append(out.Loop.Iterations, wire)
	}
	return out
}

// Unwrap is the promptrun.Result the wire result carries.
func (r *Result) Unwrap() promptrun.Result {
	if r == nil {
		return promptrun.Result{}
	}
	result := r.Result
	if r.Loop == nil {
		return result
	}
	result.Loop = &captainai.LoopResult{TotalCost: r.Loop.TotalCost, StopReason: r.Loop.StopReason}
	for _, wire := range r.Loop.Iterations {
		iteration := wire.LoopIteration
		for i := range iteration.Events {
			iteration.Events[i] = decodedEvent(iteration.Events[i])
		}
		if wire.Err != "" {
			iteration.Err = errors.New(wire.Err)
		}
		result.Loop.Iterations = append(result.Loop.Iterations, &iteration)
	}
	return result
}
