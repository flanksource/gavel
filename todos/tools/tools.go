// Package tools is the caller-tool surface a todo prompt changes TODOs
// through. Every tool writes when it is called — there is no envelope of edits
// for gavel to apply afterwards — and a lifecycle step declares which sets it
// runs with (`tools:` in todos.yaml), so nothing selects a tool by prompt name.
//
// Reads and reversible edits are allow; merging and closing, which a later run
// cannot undo, are ask, so a person approves each one through the run's broker.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// Group is the tool-preference key every todo tool shares.
const Group = "todos"

// The tool sets a lifecycle step may declare.
const (
	SetTodos = "todos"
	SetAsk   = "ask"
	SetPR    = "pr"
)

// Options is what one step run's tools act on.
type Options struct {
	Provider todos.Provider
	// Subject is the TODO the run is for: the ref every tool defaults to. Writes
	// to it land on this value, so the host's copy keeps the version the store
	// moved it to.
	Subject *types.TODO
	// WorkDir resolves a rewritten fixture's relative references.
	WorkDir string
	// Labels is the closed taxonomy todo_labels may add from.
	Labels []string
	// Preview makes todo_merge and todo_close report what they would do and
	// write nothing.
	Preview bool
	// Run collects what the run's tools did, for the host to read afterwards.
	Run *Run
	// PlanStep is true for a plan-mode step, the only kind whose attempt
	// persists the plan todo_plan records.
	PlanStep bool
	// PRWrite is true for a run that pushes each completed turn to the todo's
	// linked PR; only such a run offers pr_thread_addressed.
	PRWrite bool
}

// Sets is the registry a lifecycle step's `tools:` names are validated
// against and built from.
func Sets() map[string]func(Options) []api.ToolDefinition {
	return map[string]func(Options) []api.ToolDefinition{
		SetTodos: func(opts Options) []api.ToolDefinition { return (&toolset{opts: opts}).todoTools() },
		SetAsk:   func(Options) []api.ToolDefinition { return []api.ToolDefinition{askTool()} },
		SetPR:    func(opts Options) []api.ToolDefinition { return (&toolset{opts: opts}).prTools() },
	}
}

// SetNames are the declarable set names, sorted.
func SetNames() []string { return slices.Sorted(maps.Keys(Sets())) }

// For builds the tools of the named sets. An unknown set, or options a tool
// could not act on, is an error rather than a run with fewer tools.
func For(names []string, opts Options) ([]api.ToolDefinition, error) {
	if len(names) == 0 {
		return nil, nil
	}
	if opts.Provider == nil || opts.Subject == nil || opts.Run == nil {
		return nil, fmt.Errorf("todo tools need a provider, a subject TODO and run state")
	}
	sets := Sets()
	var out []api.ToolDefinition
	for _, name := range names {
		build, ok := sets[name]
		if !ok {
			return nil, fmt.Errorf("unknown todo tool set %q; available: %s", name, strings.Join(SetNames(), ", "))
		}
		out = append(out, build(opts)...)
	}
	return out, nil
}

// NeedsApprovals reports whether any tool raises an approval: an ask-policy
// tool, or todo_ask, which puts its own question to the run's broker.
func NeedsApprovals(defs []api.ToolDefinition) bool {
	return slices.ContainsFunc(defs, func(def api.ToolDefinition) bool {
		return def.DefaultPermission == api.ToolPolicyAsk || def.Name == AskToolName
	})
}

// Run is the state one step run's tools share with the host: the writes they
// made, the plan todo_plan saved, and the PR threads a turn reported addressed.
type Run struct {
	mu      sync.Mutex
	writes  []todos.ToolWrite
	plan    string
	threads []string
}

func (r *Run) record(name string, todo *types.TODO) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.writes = append(r.writes, todos.ToolWrite{Name: name, Ref: todos.RefOf(todo)})
}

// Writes are the writes the run's tools made, in call order.
func (r *Run) Writes() []todos.ToolWrite {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.writes)
}

// Plan is the plan markdown todo_plan last recorded, empty when it was not called.
func (r *Run) Plan() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.plan
}

// TakeThreads returns the PR thread ids reported addressed since the last call
// and forgets them: each completed turn resolves its own.
func (r *Run) TakeThreads() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	threads := r.threads
	r.threads = nil
	return threads
}

// toolset is the handlers' shared access to the run's options.
type toolset struct {
	opts Options
}

func (t *toolset) def(name, description string, policy api.ToolPolicy, schema map[string]any, handler api.ToolHandler) api.ToolDefinition {
	return api.ToolDefinition{
		Name: name, Description: description, Group: Group, InputSchema: schema,
		DefaultPermission: policy, Handler: handler,
	}
}

// resolve is the open TODO a ref names; an empty ref is the subject.
func (t *toolset) resolve(ctx context.Context, ref string) (*types.TODO, error) {
	todo, err := t.lookup(ctx, ref)
	if err != nil {
		return nil, err
	}
	if closed(todo) {
		return nil, fmt.Errorf("TODO %s is %s; closed TODOs take no further changes", todos.RefOf(todo), todo.Status)
	}
	return todo, nil
}

// lookup is the TODO a ref names, open or not; an empty ref is the subject.
func (t *toolset) lookup(ctx context.Context, ref string) (*types.TODO, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" || todos.SelfRefs(t.opts.Subject)[strings.ToLower(ref)] {
		return t.opts.Subject, nil
	}
	todo, err := t.opts.Provider.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("TODO %q: %w", ref, err)
	}
	if todo == nil {
		return nil, fmt.Errorf("TODO %q matched nothing", ref)
	}
	if todo.ID == t.opts.Subject.ID {
		return t.opts.Subject, nil
	}
	return todo, nil
}

func closed(todo *types.TODO) bool {
	return todo.Status == types.StatusCompleted || todo.Status == types.StatusVerified || todo.Status == types.StatusSkipped
}

// decode reads a tool call's arguments into target, refusing an argument the
// tool does not declare: a misspelt field silently ignored is an edit the agent
// believes it made.
func decode(input map[string]any, target any) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid arguments: %w", err)
	}
	return nil
}

func object(required []string, properties map[string]any) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func str(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func strs(description string) map[string]any {
	return map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": description}
}

func enum(description string, values ...string) map[string]any {
	return map[string]any{"type": "string", "enum": values, "description": description}
}

func refProp() map[string]any {
	return str("Short id of the TODO to act on; omit for the TODO this run is for")
}
