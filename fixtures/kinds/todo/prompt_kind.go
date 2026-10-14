package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
)

// promptKind is `kind: todo_prompt`: launch one lifecycle step of a TODO (plan,
// run, verify, triage, triage.new, or any step the project's lifecycle
// declares) through POST /api/todos/run, then follow the attempt it admitted.
type promptKind struct{}

type promptConfig struct {
	Server      string          `json:"server,omitempty"`
	Dir         string          `json:"dir,omitempty"`
	Step        string          `json:"step"`
	Todo        string          `json:"todo,omitempty"`
	Spec        json.RawMessage `json:"spec,omitempty"`
	Force       bool            `json:"force,omitempty"`
	Resume      bool            `json:"resume,omitempty"`
	ReuseBranch bool            `json:"reuseBranch,omitempty"`
}

// promptHandle is the launched attempt Observe follows.
type promptHandle struct {
	client *apiClient
	ref    string
	dir    string
	launch launchResult
}

func (promptKind) Name() string { return "todo_prompt" }

func (promptKind) Schema() json.RawMessage { return json.RawMessage(promptSchema) }

func (promptKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	handle, err := launchPrompt(ctx, step)
	if err != nil {
		// `launch` is always present so retry.when evaluates: a step that failed
		// before sending anything (no server, no ref) reports failed=false — the
		// dashboard never refused it, so retrying cannot help.
		launch := launchResult{Error: err.Error()}
		if handle != nil {
			launch = handle.launch
		}
		return fixtures.StepOutcome{Result: stepResult(step, start, err), Outputs: map[string]any{"launch": launch.outputs()}}, err
	}
	outputs, err := handle.observe(ctx)
	if err != nil {
		return fixtures.StepOutcome{Result: stepResult(step, start, err), Outputs: map[string]any{"launch": handle.launch.outputs()}}, err
	}
	return fixtures.StepOutcome{Result: stepResult(step, start, nil), Outputs: outputs, Handle: handle}, nil
}

// launchPrompt returns a handle carrying the launch whenever the request was
// sent, so a failed launch still reports `launch` to retry.when.
func launchPrompt(ctx context.Context, step *fixtures.StepContext) (*promptHandle, error) {
	cfg, err := fixtures.KindConfig[promptConfig](step)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.Step) == "" {
		return nil, fmt.Errorf("todo_prompt step %q: step is required (plan, run, verify, triage, triage.new, …)", step.Spec.ID)
	}
	client, err := newAPIClient(step, cfg.Server)
	if err != nil {
		return nil, err
	}
	ref, err := resolveRef(step, cfg.Todo)
	if err != nil {
		return nil, err
	}
	dir, err := renderOptional(step, "dir", cfg.Dir)
	if err != nil {
		return nil, err
	}
	handle := &promptHandle{client: client, ref: ref, dir: dir}
	handle.launch, err = client.launch(ctx, runPayload{
		Dir: dir, Ref: ref, Step: strings.TrimSpace(cfg.Step), Spec: cfg.Spec,
		Resume: cfg.Resume, Force: cfg.Force, ReuseBranch: cfg.ReuseBranch,
	})
	if err != nil {
		return handle, fmt.Errorf("todo_prompt step %q: %w", step.Spec.ID, err)
	}
	return handle, nil
}

func (promptKind) Observe(ctx context.Context, _ *fixtures.StepContext, handle any) (map[string]any, error) {
	h, ok := handle.(*promptHandle)
	if !ok {
		return nil, fmt.Errorf("todo_prompt kind: observe got handle %T, want *promptHandle", handle)
	}
	return h.observe(ctx)
}

// observe reads the todo, the attempt this launch admitted, that session's
// stats and the todo's selected plan.
func (h *promptHandle) observe(ctx context.Context) (map[string]any, error) {
	state, err := h.client.observeTodo(ctx, h.ref, h.dir)
	if err != nil {
		return nil, err
	}
	launch := h.launch.outputs()
	attempt, err := state.attemptByRun(launch["promptRunId"].(string))
	if err != nil {
		return nil, err
	}
	sessionID := launch["sessionId"].(string)
	var stats map[string]any
	if err := h.client.getJSON(ctx, "/api/todos/session/stats", url.Values{"dir": {state.Dir}, "sessionId": {sessionID}}, &stats); err != nil {
		return nil, err
	}
	var plan map[string]any
	if err := h.client.getJSON(ctx, "/api/todos/session/plan", url.Values{"dir": {state.Dir}, "ref": {state.ID}}, &plan); err != nil {
		return nil, err
	}
	outputs := state.outputs()
	outputs["launch"] = launch
	outputs["session"] = sessionOutputs(sessionID, attempt, stats, h.launch, state.Attempts)
	outputs["plan"] = map[string]any{
		"found":   plan["found"] == true,
		"content": stringField(plan, "content"),
		"status":  stringField(state.Item, "planStatus"),
	}
	outputs["verify"] = verifyOutputs(attempt, state.Item)
	outputs["envelope"] = mapField(attempt, "resultJson")
	return outputs, nil
}

// sessionOutputs is the `session` root. errors gathers what fails a run
// loudly: the launch's own error and the session stats' error (the API or
// process failure Captain recorded on the transcript).
func sessionOutputs(sessionID string, attempt, stats map[string]any, launch launchResult, attempts []any) map[string]any {
	errs := []any{}
	if launch.Failed {
		errs = append(errs, launch.Error)
	}
	if message := stringField(stats, "error"); message != "" {
		errs = append(errs, message)
	}
	summary := attemptSummary(attempt)
	return map[string]any{
		"id":          sessionID,
		"promptRunId": summary["promptRunId"],
		"state":       summary["state"],
		"runState":    summary["runState"],
		"error":       summary["error"],
		"errors":      errs,
		"stats":       stats,
		"attempts":    attempts,
	}
}

// verifyOutputs is the `verify` root: the definition-of-done report of the
// launched attempt (captain's api.VerifyReport, null until it verified), so a
// report left by an earlier run never satisfies this launch's wait. phase is
// the todo's latest verify phase run (todoSummary.phases.verify).
func verifyOutputs(attempt, item map[string]any) map[string]any {
	report := mapField(attempt, "verification")
	return map[string]any{
		"ran":     report["ran"] == true,
		"passed":  report["passed"] == true,
		"state":   stringField(report, "state"),
		"reason":  stringField(report, "reason"),
		"summary": mapField(report, "summary"),
		"report":  report,
		"phase":   mapField(mapField(item, "phases"), "verify"),
	}
}

// DefaultWait follows the launched attempt until it is terminal, failing early
// when the launch failed or the session recorded an error.
func (promptKind) DefaultWait(fixtures.StepSpec) *fixtures.WaitPolicy {
	return &fixtures.WaitPolicy{
		Until:    "session.state in ['completed', 'failed', 'cancelled']",
		FailWhen: "launch.failed || size(session.errors) > 0",
	}
}
