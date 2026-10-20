package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
)

// todoKind is `kind: todo`: create a TODO (optionally starting one
// lifecycle step on it), approve or reject its reviewed plan, or read it. Every action ends with
// an observation, so its outputs are the todo as the dashboard now sees it.
type todoKind struct{}

type createConfig struct {
	Title    string   `json:"title,omitempty"`
	Body     string   `json:"body"`
	Labels   []string `json:"labels,omitempty"`
	Status   string   `json:"status,omitempty"`
	Step     string   `json:"step,omitempty"`
	AutoSave *bool    `json:"autoSave,omitempty"`
}

type todoConfig struct {
	Server  string        `json:"server,omitempty"`
	Dir     string        `json:"dir,omitempty"`
	Ref     string        `json:"ref,omitempty"`
	Create  *createConfig `json:"create,omitempty"`
	Approve bool          `json:"approve,omitempty"`
	Reject  bool          `json:"reject,omitempty"`
	Get     bool          `json:"get,omitempty"`
}

// action is the one thing the step does; a step naming none reads the todo.
func (cfg todoConfig) action() (string, error) {
	var actions []string
	for _, candidate := range []struct {
		name string
		set  bool
	}{{"create", cfg.Create != nil}, {"approve", cfg.Approve}, {"reject", cfg.Reject}, {"get", cfg.Get}} {
		if candidate.set {
			actions = append(actions, candidate.name)
		}
	}
	switch len(actions) {
	case 0:
		return "get", nil
	case 1:
		if actions[0] == "create" && strings.TrimSpace(cfg.Ref) != "" {
			return "", fmt.Errorf("create makes a new todo; it takes no ref (got %q)", cfg.Ref)
		}
		return actions[0], nil
	default:
		return "", fmt.Errorf("set exactly one of create, approve, reject, get; got %s", strings.Join(actions, ", "))
	}
}

// todoHandle is what Observe re-reads.
type todoHandle struct {
	client *apiClient
	ref    string
	dir    string
}

func (todoKind) Name() string { return "todo" }

func (todoKind) Schema() json.RawMessage { return json.RawMessage(todoSchema) }

func (todoKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	handle, exports, err := runTodoAction(ctx, step)
	if err != nil {
		return fixtures.StepOutcome{Result: stepResult(step, start, err)}, err
	}
	state, err := handle.client.observeTodo(ctx, handle.ref, handle.dir)
	if err != nil {
		return fixtures.StepOutcome{Result: stepResult(step, start, err)}, err
	}
	handle.ref = state.ID
	return fixtures.StepOutcome{
		Result:  stepResult(step, start, nil),
		Outputs: state.outputs(),
		Exports: exports,
		Handle:  handle,
	}, nil
}

// runTodoAction performs the step's action and returns the todo it addressed.
func runTodoAction(ctx context.Context, step *fixtures.StepContext) (*todoHandle, map[string]string, error) {
	cfg, err := fixtures.KindConfig[todoConfig](step)
	if err != nil {
		return nil, nil, err
	}
	action, err := cfg.action()
	if err != nil {
		return nil, nil, fmt.Errorf("todo step %q: %w", step.Spec.ID, err)
	}
	client, err := newAPIClient(step, cfg.Server)
	if err != nil {
		return nil, nil, err
	}
	dir, err := renderOptional(step, "dir", cfg.Dir)
	if err != nil {
		return nil, nil, err
	}
	handle := &todoHandle{client: client, dir: dir}
	if action == "create" {
		id, err := createTodo(ctx, client, *cfg.Create, dir)
		if err != nil {
			return nil, nil, fmt.Errorf("todo step %q: %w", step.Spec.ID, err)
		}
		handle.ref = id
		return handle, map[string]string{"TODO_ID": id, "TODO_URL": client.todoURL(id)}, nil
	}
	if handle.ref, err = resolveRef(step, cfg.Ref); err != nil {
		return nil, nil, err
	}
	if action == "approve" || action == "reject" {
		// pr/ui todoApprovePayload / todoRejectPayload: {dir?, ref}.
		payload := map[string]any{"ref": handle.ref}
		if dir != "" {
			payload["dir"] = dir
		}
		var resp map[string]any
		if err := client.postJSON(ctx, "/api/todos/plan/"+action, payload, &resp); err != nil {
			return nil, nil, fmt.Errorf("todo step %q: %s: %w", step.Spec.ID, action, err)
		}
	}
	return handle, nil, nil
}

// createTodo is POST /api/todos/new (pr/ui todoNewPayload) and returns the new
// todo's id. A requested step that the server could not start fails the step:
// the todo exists, but the run the fixture is about to wait on never will.
func createTodo(ctx context.Context, client *apiClient, cfg createConfig, dir string) (string, error) {
	if strings.TrimSpace(cfg.Body) == "" && strings.TrimSpace(cfg.Title) == "" {
		return "", fmt.Errorf("create needs a body or a title")
	}
	payload := map[string]any{"body": cfg.Body}
	for key, value := range map[string]string{"title": cfg.Title, "status": cfg.Status, "dir": dir} {
		if value != "" {
			payload[key] = value
		}
	}
	if len(cfg.Labels) > 0 {
		payload["labels"] = cfg.Labels
	}
	if cfg.Step != "" {
		payload["step"] = cfg.Step
	}
	if cfg.AutoSave != nil {
		payload["autoSave"] = *cfg.AutoSave
	}
	var resp struct {
		Todo map[string]any `json:"todo"`
		Run  *struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"run"`
	}
	if err := client.postJSON(ctx, "/api/todos/new", payload, &resp); err != nil {
		return "", err
	}
	id := stringField(resp.Todo, "id")
	if id == "" {
		return "", fmt.Errorf("POST /api/todos/new: response todo has no id")
	}
	if cfg.Step == "" {
		return id, nil
	}
	if resp.Run == nil {
		return "", fmt.Errorf("POST /api/todos/new: step %s was requested for %s but the response carries no run", cfg.Step, id)
	}
	if resp.Run.Status == "failed" {
		return "", fmt.Errorf("created %s but its %s run failed to start: %s", id, cfg.Step, resp.Run.Error)
	}
	return id, nil
}

func (todoKind) Observe(ctx context.Context, _ *fixtures.StepContext, handle any) (map[string]any, error) {
	h, ok := handle.(*todoHandle)
	if !ok {
		return nil, fmt.Errorf("todo kind: observe got handle %T, want *todoHandle", handle)
	}
	state, err := h.client.observeTodo(ctx, h.ref, h.dir)
	if err != nil {
		return nil, err
	}
	return state.outputs(), nil
}

// DefaultWait: a create that starts a step waits for that attempt to settle,
// failing loudly when it fails; every other action is done when it returns.
// A config that does not decode is reported by Execute, which decodes it again.
func (todoKind) DefaultWait(spec fixtures.StepSpec) *fixtures.WaitPolicy {
	var cfg todoConfig
	if len(spec.Config) == 0 || json.Unmarshal(spec.Config, &cfg) != nil || cfg.Create == nil || cfg.Create.Step == "" {
		return nil
	}
	return &fixtures.WaitPolicy{
		Until:    "run.state in ['completed', 'failed', 'cancelled']",
		FailWhen: "run.state in ['failed', 'cancelled']",
	}
}

func (todoKind) FailureCause(outputs map[string]any) string { return runFailureCause(outputs) }

// runFailureCause is why a todo's run did not succeed, from an observation's
// outputs: the latest run's recorded error, else the newest attempt (attempts
// are newest first) that recorded one.
func runFailureCause(outputs map[string]any) string {
	run, _ := outputs["run"].(map[string]any)
	if cause := stringField(run, "error"); cause != "" {
		return cause
	}
	attempts, _ := outputs["attempts"].([]any)
	for _, raw := range attempts {
		attempt, _ := raw.(map[string]any)
		if cause := stringField(attempt, "error"); cause != "" {
			return cause
		}
	}
	return ""
}

// stepResult is the step's own verdict: whether its action ran. A `cel:` or
// `wait:` on the step judges the outputs afterwards.
func stepResult(step *fixtures.StepContext, start time.Time, err error) fixtures.FixtureResult {
	name := step.Spec.Name
	if name == "" {
		name = step.Spec.ID
	}
	result := fixtures.FixtureResult{Name: name, Type: step.Spec.Kind, Start: &start, Duration: time.Since(start), Status: task.StatusPASS}
	if err != nil {
		result.Status, result.Error = task.StatusERR, err.Error()
	}
	return result
}
