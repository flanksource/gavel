package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/ghodss/yaml"
)

// The dashboard's run surface. Execution itself lives in todos/run and the
// resolution it performs in todos/lifecycle: running a TODO is not an HTTP
// concern, and the CLI, the clicky entity and this handler must not be able to
// disagree about which step runs or how it is configured. What stays here is the
// wire — the payload, its validation, and the response the dialog reads.
type (
	todoRunStartResult = run.StartResult
	todoRunRequest     = run.Request
	todoRunOptions     = run.Options
)

// run.Start is deliberately NOT aliased here. An alias copies the function
// value at init, so a test replacing the copy would leave every run started
// through the todos entity — which calls run.Start directly — going to the real
// driver. One seam, called by its own name.

var (
	specCommit            = run.Commit
	specDryRun            = run.DryRun
	specDirty             = run.Dirty
	todoRunRefs           = run.Refs
	todoRunStartedMessage = run.StartedMessage
)

// todoRunPayload is the run request as the dashboard sends it.
//
// The execution mechanism is not a field of its own: the lifecycle STEP decides
// which prompt renders and how the run behaves, and every model/budget/
// permission knob is the same api.Spec captain's prompt-run editor produces.
type todoRunPayload struct {
	Dir  string   `json:"dir,omitempty"`
	Ref  string   `json:"ref,omitempty"`
	Refs []string `json:"refs,omitempty"`
	// Step names the lifecycle step to run — `run`, `plan`, `verify`, `triage`,
	// or any step the project declares. Empty runs the step the lifecycle picks
	// next for this todo.
	Step string `json:"step,omitempty"`
	// Spec carries the model/mode/effort/prompt/budget/permissions/session knobs.
	//
	// It is a named field under its own `spec` key, not embedded. api.Spec
	// declares value-receiver MarshalJSON/MarshalYAML to omit its empty sections;
	// embedding promoted them onto the payload, so marshaling emitted a bare spec
	// and silently dropped every sibling field — including in the review API's
	// `options` object.
	Spec api.Spec `json:"spec,omitempty"`
	// Resume continues the todo's prior agent session (claude --resume) instead of
	// starting fresh. It stays a sibling flag rather than a Spec field because it is
	// a session-identity decision: a fresh run also carries a (minted) sessionId, so
	// resume cannot be inferred from Spec.SessionID.
	Resume bool `json:"resume,omitempty"`
	// Force dispatches even when the todo already has a live run owned by a
	// running process: the two runs proceed in parallel. Without it such a
	// dispatch is refused and the client is told which run is in the way.
	Force bool `json:"force,omitempty"`
}

// removedTodoRunFields are payload keys that named a run configuration which no
// longer exists. They are rejected by name rather than by the decoder's generic
// "unknown field" so the client is told what replaced them; dropping one
// silently would run on the lifecycle's own choice while the dialog believed it
// had chosen something.
var removedTodoRunFields = []string{"runMode", "driver", "prompt", "mode", "agent", "plan"}

// UnmarshalJSON is strict on purpose: an unknown key is a client that believes
// it configured something this server never read.
func (p *todoRunPayload) UnmarshalJSON(data []byte) error {
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(data, &keys); err != nil {
		return err
	}
	for _, field := range removedTodoRunFields {
		if _, ok := keys[field]; ok {
			return fmt.Errorf(
				"invalid run configuration: %q is not supported; name the lifecycle step with \"step\" and send every run knob in \"spec\"",
				field)
		}
	}
	type wire todoRunPayload
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wire
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	*p = todoRunPayload(decoded)
	return nil
}

type todoRunResponse struct {
	Status string   `json:"status"`
	Ref    string   `json:"ref"`
	Refs   []string `json:"refs,omitempty"`
	Count  int      `json:"count"`
	Dir    string   `json:"dir"`
	// Step is the lifecycle step that ran and Reason why it was chosen — named by
	// the client, or picked by the lifecycle's own predicates.
	Step     string `json:"step"`
	Reason   string `json:"reason,omitempty"`
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	// RuntimeMode is the resolved mechanism (cmux, agent, cli, api). It is not
	// keyed `mode`: that is a run-payload key this endpoint rejects on input,
	// and a response must not hand a client a key it cannot send back.
	RuntimeMode string  `json:"runtimeMode,omitempty"`
	Effort      string  `json:"effort,omitempty"`
	Resume      bool    `json:"resume,omitempty"`
	SessionID   string  `json:"sessionId,omitempty"`
	Timeout     string  `json:"timeout"`
	MaxBudget   float64 `json:"maxBudget,omitempty"`
	MaxTurns    int     `json:"maxTurns,omitempty"`
	Commit      bool    `json:"commit"`
	Message     string  `json:"message"`
}

type todoRunPreviewResponse struct {
	Prompt      string `json:"prompt"`
	SpecYAML    string `json:"specYaml"`
	Step        string `json:"step"`
	Reason      string `json:"reason,omitempty"`
	Provider    string `json:"provider,omitempty"`
	RuntimeMode string `json:"runtimeMode,omitempty"`
	Effort      string `json:"effort,omitempty"`
	Count       int    `json:"count"`
}

// resolveTodoRunRequest decodes a run/preview payload and resolves its options,
// provider, and todos. handleTodoRun and handleTodoRunPreview share it so both
// interpret the same request identically; the returned status is the HTTP code
// to report when err is non-nil.
func (s *Server) resolveTodoRunRequest(r *http.Request) (todos.Provider, todoSource, []*types.TODO, todoRunOptions, int, error) {
	var payload todoRunPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err)
	}
	refs := normalizeTodoRunRefs(payload, r)
	return s.resolveTodoRunPayload(r.Context(), payload, refs, todoSourceFromRequest(r), requestOrigin(r))
}

// resolveTodoRunPayload turns an already-decoded payload into everything a run
// needs. It is separate from resolveTodoRunRequest so another entrypoint can
// resolve one synthetic payload through exactly the same path a single run
// takes — a second resolution path would be one more place for the dashboard's
// runs to diverge from each other.
func (s *Server) resolveTodoRunPayload(
	ctx context.Context,
	payload todoRunPayload,
	refs []string,
	source todoSource,
	origin string,
) (todos.Provider, todoSource, []*types.TODO, todoRunOptions, int, error) {
	if len(refs) == 0 {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("ref is required")
	}
	// Wire validation runs before anything is looked up, so a malformed payload
	// still fails as a 400 without opening a workspace. The resolved spec — which
	// also carries the .gavel.yaml and step layers — is validated again by the
	// lifecycle host when the run resolves.
	opts, err := buildTodoRunOptions(payload, nil)
	if err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, err
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	if err := validateTodoRunCardinality(len(refs)); err != nil {
		return nil, source, nil, todoRunOptions{}, http.StatusBadRequest, err
	}

	provider, source, todoList, status, err := s.loadTodoRunTargets(ctx, source, refs)
	if err != nil {
		return provider, source, nil, todoRunOptions{}, status, err
	}
	for _, todo := range todoList {
		todo.MarkdownBody = todos.AbsolutizeAttachmentURLs(todo.MarkdownBody, origin)
	}
	return provider, source, todoList, opts, http.StatusOK, nil
}

// loadTodoRunTargets resolves the refs a run covers through the workspace that
// owns them: a request naming no dir resolves the issue globally and then reopens
// its owning workspace, so a run can never be dispatched against the server's
// default workspace by accident.
func (s *Server) loadTodoRunTargets(
	ctx context.Context,
	source todoSource,
	refs []string,
) (todos.Provider, todoSource, []*types.TODO, int, error) {
	if strings.TrimSpace(source.Dir) == "" {
		provider, owningSource, todo, err := s.resolveTodoReference(ctx, source, refs[0])
		if err != nil {
			return provider, owningSource, nil, http.StatusNotFound, err
		}
		return provider, owningSource, []*types.TODO{todo}, http.StatusOK, nil
	}
	provider, resolvedSource, err := s.todoProviderContext(ctx, source)
	if err != nil {
		return nil, resolvedSource, nil, http.StatusBadRequest, err
	}
	todoList := make([]*types.TODO, 0, len(refs))
	for _, ref := range refs {
		todo, err := provider.Get(ctx, ref)
		if err != nil {
			return provider, resolvedSource, nil, http.StatusNotFound, err
		}
		todoList = append(todoList, todo)
	}
	return provider, resolvedSource, todoList, http.StatusOK, nil
}

func validateTodoRunCardinality(count int) error {
	if count <= 1 {
		return nil
	}
	return fmt.Errorf("grouped TODO execution is not supported by the native PostgreSQL runtime; run one issue at a time")
}

func (s *Server) handleTodoRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider, source, todoList, opts, status, err := s.resolveTodoRunRequest(r)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	req := todoRunRequest{
		Provider: provider,
		Registry: todoRuns(),
		Todo:     todoList[0],
		Dir:      source.Dir,
		Options:  opts,
		// The dashboard serves the approval endpoints, so a tool call the run cannot
		// pre-approve becomes a durable request a person can answer.
		Broker: todoApprovalBroker(source.Dir),
	}
	// Pre-flight: the one fold the run performs, so a request that cannot
	// resolve is a 400 the dialog renders instead of a background failure. It is
	// where the response's step, reason and runtime come from, and it is the
	// fold the dispatch below runs — not a second one that could disagree.
	prepared, err := run.Resolve(r.Context(), req)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateTodoRunRuntime(prepared.Resolution.Spec); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	req.Prepared = prepared
	resp := todoRunResponseFor(source, todoList, opts, prepared)
	// A dry run still executes the agent; Captain's commit hook reports rather
	// than cuts the declared commit. Prompt-only inspection uses the preview API.
	started, err := run.Start(req)
	if err != nil {
		// A todo that is already running is a question for the user, not a bad
		// request: answer it by retrying with force, and the two runs proceed in
		// parallel. The dialog needs the incumbent's identity to say so.
		var owned *todos.ErrRunOwnedElsewhere
		if errors.As(err, &owned) {
			writeTodoRunConflict(w, owned)
			return
		}
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if started.Status == "started" && strings.TrimSpace(started.SessionID) == "" {
		writeTodoError(w, http.StatusInternalServerError, errors.New("todo run was admitted without a Captain session id"))
		return
	}
	resp.Status = started.Status
	resp.SessionID = started.SessionID
	resp.Message = started.Message
	if resp.Message == "" && started.Status == "started" {
		resp.Message = todoRunStartedMessage(len(todoList))
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// todoRunResponseFor describes the run the dashboard just asked for, read from
// the fold rather than from the payload: a knob the client left unset was
// supplied by the lifecycle, and the dialog has to show what will actually run.
func todoRunResponseFor(source todoSource, todoList []*types.TODO, opts todoRunOptions, prepared *run.Prepared) todoRunResponse {
	spec := prepared.Resolution.Spec
	return todoRunResponse{
		Ref:         todos.TODOReference(todoList[0]),
		Refs:        todoRunRefs(todoList),
		Count:       len(todoList),
		Dir:         source.Dir,
		Step:        prepared.Step.Name,
		Reason:      prepared.Reason,
		Provider:    providerKey(spec.Model),
		RuntimeMode: string(spec.Mode),
		Model:       spec.Name,
		Effort:      string(spec.Effort),
		Resume:      opts.Resume,
		Timeout:     prepared.Resolution.Timeout.String(),
		MaxBudget:   spec.Budget.Cost,
		MaxTurns:    spec.Budget.MaxTurns,
		Commit:      specCommit(spec) && !specDryRun(spec),
	}
}

// handleTodoRunPreview renders the exact request a run would dispatch, without
// starting it, so the advanced run dialog can show the prompt that will be sent
// before the user commits to a run. It accepts the same payload as handleTodoRun
// and resolves through the same seam, so a preview can never describe a
// different run from the one that follows it.
func (s *Server) handleTodoRunPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider, source, todoList, opts, status, err := s.resolveTodoRunRequest(r)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	prepared, specYAML, err := buildTodoRunSpecPreview(r.Context(), provider, source, todoList[0], opts)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	spec := prepared.Resolution.Spec
	json.NewEncoder(w).Encode(todoRunPreviewResponse{ //nolint:errcheck
		Prompt:      prepared.Resolution.Prompt,
		SpecYAML:    specYAML,
		Step:        prepared.Step.Name,
		Reason:      prepared.Reason,
		Provider:    providerKey(spec.Model),
		RuntimeMode: string(spec.Mode),
		Effort:      string(spec.Effort),
		Count:       len(todoList),
	})
}

// buildTodoRunSpecPreview folds the run without dispatching it and renders the
// resolved spec as YAML.
func buildTodoRunSpecPreview(
	ctx context.Context,
	provider todos.Provider,
	source todoSource,
	todo *types.TODO,
	opts todoRunOptions,
) (*run.Prepared, string, error) {
	prepared, err := run.Resolve(ctx, todoRunRequest{
		Provider: provider,
		Registry: todoRuns(),
		Todo:     todo,
		Dir:      source.Dir,
		Options:  opts,
	})
	if err != nil {
		return nil, "", err
	}
	encoded, err := yaml.Marshal(prepared.Resolution.Spec)
	if err != nil {
		return nil, "", fmt.Errorf("marshal resolved Captain spec as YAML: %w", err)
	}
	return prepared, string(encoded), nil
}

// normalizeTodoRunRefs collects the todo refs to run, de-duplicated and in order:
// the explicit refs[] (multi-select), then the single ref, then the ?ref query
// param.
func normalizeTodoRunRefs(payload todoRunPayload, r *http.Request) []string {
	seen := make(map[string]bool)
	var refs []string
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref == "" || seen[ref] {
			return
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	for _, ref := range payload.Refs {
		add(ref)
	}
	add(payload.Ref)
	if len(refs) == 0 {
		add(r.URL.Query().Get("ref"))
	}
	return refs
}

// buildTodoRunOptions validates the payload at the wire boundary and turns it
// into the decisions a run is made of. Nothing is defaulted here: defaults
// belong to the lifecycle's fold, and one applied at this layer would outrank
// the configuration it claims to defer to.
//
// prior are the layers a continuation inherits from the run it continues; a
// fresh run has none.
func buildTodoRunOptions(payload todoRunPayload, prior []api.SpecLayer) (todoRunOptions, error) {
	spec, err := validateTodoRunSpec(payload.Spec)
	if err != nil {
		return todoRunOptions{}, err
	}
	return todoRunOptions{
		Step:       strings.TrimSpace(payload.Step),
		Request:    spec,
		Prior:      prior,
		Resume:     payload.Resume,
		Concurrent: payload.Force,
		Host:       lifecycle.HostDashboard,
	}, nil
}

// validateTodoRunSpec checks the request spec's own sections and normalises the
// two the wire can express loosely (effort casing, timeout formatting).
func validateTodoRunSpec(spec api.Spec) (api.Spec, error) {
	if effort := strings.ToLower(strings.TrimSpace(string(spec.Effort))); effort != "" {
		if !validTodoRunEffort(effort) {
			return api.Spec{}, fmt.Errorf("invalid effort %q", spec.Effort)
		}
		spec.Effort = api.Effort(effort)
	}
	if raw := strings.TrimSpace(spec.Budget.Timeout); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return api.Spec{}, fmt.Errorf("invalid timeout %q", spec.Budget.Timeout)
		}
		if parsed <= 0 {
			return api.Spec{}, fmt.Errorf("timeout must be greater than zero")
		}
		spec.Budget.Timeout = parsed.String()
	}
	if err := spec.Budget.Validate(); err != nil {
		return api.Spec{}, err
	}
	if err := spec.Permissions.Validate(); err != nil {
		return api.Spec{}, err
	}
	if err := spec.Workflow.Validate(); err != nil {
		return api.Spec{}, fmt.Errorf("workflow: %w", err)
	}
	return spec, nil
}

// validateTodoRunRuntime checks the RESOLVED runtime against the catalog the run
// dialog offers. The catalog decides nothing — presentation defaults folded as
// the top layer would outrank the `.gavel.yaml` and prompt frontmatter they
// claim to defer to, which is what todoRunPromptDefaults exists to prevent — it
// only refuses a pair this dashboard cannot run, before a background run fails
// on it. A spec naming no model is not its business: whether a step needs one is
// the lifecycle's question, asked by lifecycle.RequireModel.
func validateTodoRunRuntime(spec api.Spec) error {
	if mode := strings.TrimSpace(string(spec.Mode)); mode != "" {
		if _, ok := registry.ParseRuntimeMode(mode); !ok {
			return fmt.Errorf("the resolved run mode %q is not one this dashboard can run", mode)
		}
	}
	if strings.TrimSpace(spec.Name) == "" {
		return nil
	}
	if _, err := registry.ResolveModel(spec.Model); err != nil {
		return fmt.Errorf("the resolved run model %q is not one this dashboard can run: %w", spec.Name, err)
	}
	return nil
}

// writeTodoRunConflict answers a dispatch that lost to a live run with what the
// client needs to decide: which run is in the way, who is driving it, and that
// retrying with force runs the two in parallel.
func writeTodoRunConflict(w http.ResponseWriter, owned *todos.ErrRunOwnedElsewhere) {
	w.WriteHeader(http.StatusConflict)
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"error":       owned.Error(),
		"reason":      "run_owned_elsewhere",
		"promptRunId": owned.PromptRunID,
		"stepKind":    owned.StepKind,
		"owner":       owned.Owner,
		"runningFor":  owned.Since.Round(time.Second).String(),
		"retryWith":   map[string]bool{"force": true},
	})
}
