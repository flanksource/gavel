package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/commit"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/claude"
	"github.com/flanksource/gavel/todos/cmux"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

type todoCounts struct {
	Total      int `json:"total"`
	Open       int `json:"open"`
	Draft      int `json:"draft"`
	Pending    int `json:"pending"`
	InProgress int `json:"inProgress"`
	Failed     int `json:"failed"`
	Verified   int `json:"verified"`
	Completed  int `json:"completed"`
	Skipped    int `json:"skipped"`
}

type todoListResponse struct {
	Provider string        `json:"provider"`
	Dir      string        `json:"dir,omitempty"`
	Counts   todoCounts    `json:"counts"`
	Items    []todoSummary `json:"items"`
}

type todoSummary struct {
	Ref            string                `json:"ref"`
	ID             string                `json:"id,omitempty"`
	ShortID        string                `json:"shortId,omitempty"`
	Title          string                `json:"title"`
	Status         types.Status          `json:"status"`
	Priority       types.Priority        `json:"priority"`
	Provider       string                `json:"provider,omitempty"`
	ProviderState  string                `json:"providerState,omitempty"`
	FilePath       string                `json:"filePath,omitempty"`
	CWD            string                `json:"cwd,omitempty"`
	Labels         []string              `json:"labels,omitempty"`
	Attempts       int                   `json:"attempts,omitempty"`
	Created        *time.Time            `json:"created,omitempty"`
	LastRun        *time.Time            `json:"lastRun,omitempty"`
	SessionID      string                `json:"sessionId,omitempty"`
	Body           string                `json:"body,omitempty"`
	Implementation string                `json:"implementation,omitempty"`
	Events         []types.ProviderEvent `json:"events,omitempty"`
	// Criteria are the todo's acceptance criteria, parsed from its
	// "## Acceptance Criteria" section; populated on detail responses so the
	// dashboard can render and edit them structurally.
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
	// Diff is the aggregated git diff footprint of the todo's commits (those
	// carrying its Gavel-Issue-Id trailer); nil when no commits reference it.
	Diff *todoDiffStat `json:"diff,omitempty"`
}

// todoDiffStat is the JSON shape of a todo's aggregated git diff footprint,
// mirroring git.DiffStat for the dashboard.
type todoDiffStat struct {
	Commits int `json:"commits"`
	Files   int `json:"files"`
	Adds    int `json:"adds"`
	Dels    int `json:"dels"`
}

type todoSource struct {
	Provider string
	Dir      string
}

// providerAuto resolves the provider per directory the way countProjectTodos
// does: a Grite workspace if .grite exists, otherwise .todos files. It lets the
// dashboard list a workspace's todos without the caller knowing which it uses.
const providerAuto = "auto"

type todoCreatePayload struct {
	Provider string         `json:"provider,omitempty"`
	Dir      string         `json:"dir,omitempty"`
	Title    string         `json:"title"`
	Body     string         `json:"body,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
}

type todoNewPayload struct {
	todoCreatePayload
	AutoSave *bool `json:"autoSave,omitempty"`
}

type todoAttachmentSummary struct {
	Field       string `json:"field,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
	Size        int64  `json:"size"`
	ID          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	IsImage     bool   `json:"isImage,omitempty"`
}

type todoNewResponse struct {
	Todo        todoSummary             `json:"todo"`
	AutoSave    bool                    `json:"autoSave"`
	Attachments []todoAttachmentSummary `json:"attachments,omitempty"`
}

type todoUpdatePayload struct {
	Provider string         `json:"provider,omitempty"`
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	// Title/Body edit the TODO's content; a nil pointer leaves the field
	// unchanged (an explicit empty body is allowed, an empty title is not).
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	// Comment, when set, appends a comment. Combined with status it reopens (or
	// closes) the TODO with a comment in one request.
	Comment string `json:"comment,omitempty"`
}

// todoTransferPayload moves the todo at Ref from the source workspace
// (FromDir/FromProvider) to the target workspace (ToDir/ToProvider). Each
// dir/provider pair resolves the same way the list/get endpoints do.
type todoTransferPayload struct {
	Ref          string `json:"ref"`
	FromDir      string `json:"fromDir,omitempty"`
	FromProvider string `json:"fromProvider,omitempty"`
	ToDir        string `json:"toDir"`
	ToProvider   string `json:"toProvider,omitempty"`
}

type todoTransferResponse struct {
	Dir      string      `json:"dir"`
	Provider string      `json:"provider"`
	Todo     todoSummary `json:"todo"`
}

type todoRunPayload struct {
	Provider string   `json:"provider,omitempty"`
	Dir      string   `json:"dir,omitempty"`
	Ref      string   `json:"ref,omitempty"`
	Refs     []string `json:"refs,omitempty"`
	Agent    string   `json:"agent,omitempty"`
	Mode     string   `json:"mode,omitempty"`
	// Driver selects the agent driver (claude-cmux, claude-headless, claude-sdk,
	// claude-api, codex-cmux, codex-headless). When empty it is derived from the
	// legacy agent+mode pair for backward compatibility.
	Driver    string  `json:"driver,omitempty"`
	Model     string  `json:"model,omitempty"`
	Effort    string  `json:"effort,omitempty"`
	Plan      bool    `json:"plan,omitempty"`
	Resume    bool    `json:"resume,omitempty"`
	Timeout   string  `json:"timeout,omitempty"`
	MaxBudget float64 `json:"maxBudget,omitempty"`
	MaxCost   float64 `json:"maxCost,omitempty"`
	MaxTurns  int     `json:"maxTurns,omitempty"`
	Dirty     bool    `json:"dirty,omitempty"`
	DryRun    bool    `json:"dryRun,omitempty"`
	// Commit controls whether `gavel commit` runs over the agent's changes once
	// the run finishes. A nil pointer defaults to true (the dashboard auto-commits
	// like the CLI's `todos run --commit`); send false to disable it.
	Commit *bool `json:"commit,omitempty"`
	// Check, when true, runs the configured `checks` test/lint suite after the
	// agent completes and feeds failures back to it. Opt-in (defaults off),
	// mirroring the CLI's `todos run --check`.
	Check *bool `json:"check,omitempty"`
	// Verify, when true, runs an AI issue verification over the committed work
	// after the run, scoring it against the issue's acceptance criteria. Opt-in,
	// mirroring the CLI's `todos run --verify`.
	Verify *bool `json:"verify,omitempty"`
}

type todoRunResponse struct {
	Status    string   `json:"status"`
	Ref       string   `json:"ref"`
	Refs      []string `json:"refs,omitempty"`
	Count     int      `json:"count"`
	Dir       string   `json:"dir"`
	Provider  string   `json:"provider"`
	Agent     string   `json:"agent"`
	Mode      string   `json:"mode"`
	Model     string   `json:"model,omitempty"`
	Effort    string   `json:"effort,omitempty"`
	Driver    string   `json:"driver,omitempty"`
	Plan      bool     `json:"plan,omitempty"`
	Resume    bool     `json:"resume,omitempty"`
	SessionID string   `json:"sessionId,omitempty"`
	Timeout   string   `json:"timeout"`
	MaxBudget float64  `json:"maxBudget,omitempty"`
	MaxTurns  int      `json:"maxTurns,omitempty"`
	Commit    bool     `json:"commit"`
	Message   string   `json:"message"`
}

type todoRunPreviewResponse struct {
	Prompt string `json:"prompt"`
	Mode   string `json:"mode"`
	Agent  string `json:"agent"`
	Effort string `json:"effort,omitempty"`
	Plan   bool   `json:"plan,omitempty"`
	Count  int    `json:"count"`
}

type todoRunRequest struct {
	Provider todos.Provider
	// Todos are executed together in a single agent session (multi-select run);
	// a single-element slice is the ordinary one-todo run.
	Todos   []*types.TODO
	Source  todoSource
	Backend string
	Options todoRunOptions
}

type todoRunOptions struct {
	Agent           string
	Mode            string
	Driver          string
	Model           string
	Effort          string
	Plan            bool
	Resume          bool
	SessionID       string
	Timeout         time.Duration
	MaxBudget       float64
	MaxTurns        int
	Dirty           bool
	DryRun          bool
	Commit          bool
	Check           bool
	Verify          bool
	TimeoutOriginal string
}

var startTodoRun = defaultStartTodoRun

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodosList(w, r)
	case http.MethodPost:
		s.handleTodoCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodoItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodoGet(w, r)
	case http.MethodPatch:
		s.handleTodoPatch(w, r)
	case http.MethodDelete:
		s.handleTodoDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodosList(w http.ResponseWriter, r *http.Request) {
	source := todoSourceFromRequest(r)
	provider, source, err := s.todoProvider(source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	filters, err := todoFiltersFromRequest(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	items, err := provider.List(r.Context(), filters)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	for _, item := range items {
		reconcileSessionStatus(item, source.Dir)
	}
	resp := todoListResponse{
		Provider: source.Provider,
		Dir:      source.Dir,
		Counts:   summarizeTodos(items),
		Items:    make([]todoSummary, 0, len(items)),
	}
	stats := commitDiffStats(r.Context(), source.Dir)
	for _, item := range items {
		sum := summarizeTodo(item, false)
		sum.Diff = diffStatFor(stats, item.ID)
		resp.Items = append(resp.Items, sum)
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// diffStatFor builds the JSON diff stat for a todo id from the workspace's
// computed map, returning nil when the todo has no linked commits so the field
// is omitted rather than serialized as an all-zero object.
func diffStatFor(stats map[string]gavelgit.DiffStat, id string) *todoDiffStat {
	d, ok := stats[strings.TrimSpace(id)]
	if !ok || (d.Commits == 0 && d.Files == 0) {
		return nil
	}
	return &todoDiffStat{Commits: d.Commits, Files: d.Files, Adds: d.Adds, Dels: d.Dels}
}

func (s *Server) handleTodoGet(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, source, err := s.todoProvider(todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	reconcileSessionStatus(todo, source.Dir)
	sum := summarizeTodo(todo, true)
	sum.Diff = diffStatFor(commitDiffStats(r.Context(), source.Dir), todo.ID)
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

func (s *Server) handleTodoCreate(w http.ResponseWriter, r *http.Request) {
	var payload todoCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	source := todoSourceFromRequest(r)
	if payload.Provider != "" {
		source.Provider = payload.Provider
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, err := s.todoProvider(source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     payload.Body,
		Priority: payload.Priority,
		Status:   payload.Status,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
}

func (s *Server) handleTodoNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, attachments, err := parseTodoNewPayload(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	autoSave := false
	if payload.AutoSave != nil {
		autoSave = *payload.AutoSave
	}
	if payload.Status == "" {
		if autoSave {
			payload.Status = types.StatusPending
		} else {
			payload.Status = types.StatusDraft
		}
	}
	if payload.Status != "" && !validTodoStatus(payload.Status) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid status %q", payload.Status))
		return
	}
	if payload.Priority != "" && !validTodoPriority(payload.Priority) {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid priority %q", payload.Priority))
		return
	}

	source := todoSourceFromRequest(r)
	if payload.Provider != "" {
		source.Provider = payload.Provider
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, err := s.todoProvider(source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     todoBodyWithAttachments(payload.Body, attachments),
		Priority: payload.Priority,
		Status:   payload.Status,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todoNewResponse{ //nolint:errcheck
		Todo:        summarizeTodo(todo, true),
		AutoSave:    autoSave,
		Attachments: attachments,
	})
}

func (s *Server) handleTodoPatch(w http.ResponseWriter, r *http.Request) {
	var payload todoUpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	ref := strings.TrimSpace(payload.Ref)
	if ref == "" {
		ref = strings.TrimSpace(r.URL.Query().Get("ref"))
	}
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	// A PATCH may edit content (title/body), change state (status/priority),
	// add a comment, or any combination; at least one operation is required.
	var update todos.StateUpdate
	if payload.Status != "" {
		if !validTodoStatus(payload.Status) {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid status %q", payload.Status))
			return
		}
		update.Status = &payload.Status
	}
	if payload.Priority != "" {
		if !validTodoPriority(payload.Priority) {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid priority %q", payload.Priority))
			return
		}
		update.Priority = &payload.Priority
	}

	var edit todos.EditRequest
	if payload.Title != nil {
		title := strings.TrimSpace(*payload.Title)
		if title == "" {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("title cannot be empty"))
			return
		}
		edit.Title = &title
	}
	if payload.Body != nil {
		edit.Body = payload.Body
	}
	comment := strings.TrimSpace(payload.Comment)

	hasState := update.Status != nil || update.Priority != nil
	if !hasState && edit.IsEmpty() && comment == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("status, priority, title, body, or comment is required"))
		return
	}

	source := todoSourceFromRequest(r)
	if payload.Provider != "" {
		source.Provider = payload.Provider
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, _, err := s.todoProvider(source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	// Order: edit content, then reopen/close, then comment, so a reopen-with-comment
	// posts the comment against the now-open TODO and it lands last in the timeline.
	if !edit.IsEmpty() {
		if err := provider.Edit(r.Context(), todo, edit); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if hasState {
		if err := provider.UpdateState(r.Context(), todo, update); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if comment != "" {
		if err := provider.Comment(r.Context(), todo, comment); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	// Edits and comments mutate the body/event history; re-read so the response
	// reflects the provider's authoritative state (new event, rewritten body).
	if !edit.IsEmpty() || comment != "" {
		if refreshed, gerr := provider.Get(r.Context(), ref); gerr == nil {
			todo = refreshed
		}
	}
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, _, err := s.todoProvider(todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Get(r.Context(), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if err := provider.Delete(r.Context(), todo); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, `{"status":"ok"}`)
}

func (s *Server) handleTodoTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload todoTransferPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if strings.TrimSpace(payload.ToDir) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("toDir is required"))
		return
	}
	source, src, err := s.todoProvider(todoSource{Provider: payload.FromProvider, Dir: payload.FromDir})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	target, dst, err := s.todoProvider(todoSource{Provider: payload.ToProvider, Dir: payload.ToDir})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Refuse a no-op self-transfer (same dir resolving to the same backend),
	// which would create a duplicate and then delete the original. Different
	// backends in one dir is a legitimate migration, so only guard same+same.
	srcBackend, _ := resolveTodoBackend(src.Dir, payload.FromProvider)
	dstBackend, _ := resolveTodoBackend(dst.Dir, payload.ToProvider)
	if src.Dir == dst.Dir && srcBackend == dstBackend {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("source and target are the same workspace"))
		return
	}
	created, err := todos.Transfer(r.Context(), source, target, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoTransferResponse{ //nolint:errcheck
		Dir:      dst.Dir,
		Provider: dstBackend,
		Todo:     summarizeTodo(created, true),
	})
}

// resolveTodoRunRequest decodes a run/preview payload and resolves its options,
// provider, and todos. handleTodoRun and handleTodoRunPreview share it so both
// interpret the same request identically; the returned status is the HTTP code
// to report when err is non-nil.
func (s *Server) resolveTodoRunRequest(r *http.Request) (todos.Provider, todoSource, []*types.TODO, todoRunOptions, int, error) {
	var payload todoRunPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("invalid json")
	}
	refs := normalizeTodoRunRefs(payload, r)
	if len(refs) == 0 {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, fmt.Errorf("ref is required")
	}
	opts, err := normalizeTodoRunOptions(payload)
	if err != nil {
		return nil, todoSource{}, nil, todoRunOptions{}, http.StatusBadRequest, err
	}
	source := todoSourceFromRequest(r)
	if payload.Provider != "" {
		source.Provider = payload.Provider
	}
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, source, err := s.todoProvider(source)
	if err != nil {
		return nil, source, nil, opts, http.StatusBadRequest, err
	}
	origin := requestOrigin(r)
	todoList := make([]*types.TODO, 0, len(refs))
	for _, ref := range refs {
		todo, err := provider.Get(r.Context(), ref)
		if err != nil {
			return provider, source, nil, opts, http.StatusNotFound, err
		}
		todo.MarkdownBody = absolutizeAttachmentURLs(todo.MarkdownBody, origin)
		todoList = append(todoList, todo)
	}
	return provider, source, todoList, opts, http.StatusOK, nil
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
	backend, _ := resolveTodoBackend(source.Dir, source.Provider)
	// Resolve the run's session id once, up front, so it is stable across the
	// validation and start calls below and can be returned to the client to
	// follow the session log live (see handleTodoSessionStream).
	opts.SessionID = resolveRunSessionID(opts, todoList)
	req := todoRunRequest{
		Provider: provider,
		Todos:    todoList,
		Source:   source,
		Backend:  backend,
		Options:  opts,
	}
	if _, _, err := newTodoRunExecutor(req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	resp := todoRunResponse{
		Status:    "started",
		Ref:       todos.TODOReference(todoList[0]),
		Refs:      todoRunRefs(todoList),
		Count:     len(todoList),
		Dir:       source.Dir,
		Provider:  backend,
		Agent:     opts.Agent,
		Mode:      opts.Mode,
		Driver:    opts.Driver,
		Model:     opts.Model,
		Effort:    opts.Effort,
		Plan:      opts.Plan,
		Resume:    opts.Resume,
		SessionID: opts.SessionID,
		Timeout:   opts.Timeout.String(),
		MaxBudget: opts.MaxBudget,
		MaxTurns:  opts.MaxTurns,
		Commit:    opts.Commit,
		Message:   todoRunStartedMessage(len(todoList)),
	}
	if opts.DryRun {
		resp.Status = "dry_run"
		resp.Message = "Todo run validated"
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
		return
	}
	if err := startTodoRun(req); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// handleTodoRunPreview renders the exact prompt a run would dispatch, without
// starting it, so the advanced run dialog can show the prompt that will be sent
// before the user commits to a run. It accepts the same payload as handleTodoRun.
func (s *Server) handleTodoRunPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_, source, todoList, opts, status, err := s.resolveTodoRunRequest(r)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	resp := todoRunPreviewResponse{
		Prompt: buildTodoRunPromptPreview(source.Dir, todoList, opts),
		Mode:   opts.Mode,
		Agent:  opts.Agent,
		Effort: opts.Effort,
		Plan:   opts.Plan,
		Count:  len(todoList),
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

// buildTodoRunPromptPreview renders the prompt a run would dispatch, mirroring
// newTodoRunExecutor: cmux wraps the group prompt with the run's title and the
// implement/plan suffix, while inline sends the bare claude prompt.
func buildTodoRunPromptPreview(dir string, todoList []*types.TODO, opts todoRunOptions) string {
	if opts.Mode == "inline" {
		if len(todoList) == 1 {
			return claude.BuildPrompt(todoList[0], dir)
		}
		return claude.BuildGroupPrompt(todoList, dir)
	}
	return cmux.PreviewInstruction(todoList, dir, opts.Effort, opts.Plan, opts.Resume, opts.Agent)
}

// resolveTodoDir turns a request's dir param into an absolute workspace path,
// defaulting to the server's work dir and joining relative dirs onto it. Shared
// by todoProvider and the session stream so both resolve dirs identically.
func (s *Server) resolveTodoDir(dir string) string {
	workDir := s.todoWorkDir()
	if dir == "" {
		return workDir
	}
	if !filepath.IsAbs(dir) {
		return filepath.Join(workDir, dir)
	}
	return dir
}

func (s *Server) todoProvider(source todoSource) (todos.Provider, todoSource, error) {
	source.Dir = s.resolveTodoDir(source.Dir)
	switch source.Provider {
	case "", providerAuto, todos.ProviderGrite, todos.ProviderFiles:
		return providerForDir(source.Dir, source.Provider), source, nil
	default:
		return nil, source, fmt.Errorf("unknown todo provider %q", source.Provider)
	}
}

// ProviderForProject resolves the todo provider for a stored project, honoring
// its pinned TodoProvider (or auto-detecting from the resolved directory). The
// `gavel todos transfer` command uses it to build a transfer target from a
// named project the same way the dashboard resolves a workspace.
func ProviderForProject(p Project) todos.Provider {
	return providerForDir(p.ResolvedDir(), p.TodoProvider)
}

// providerForDir builds the todo provider for a workspace directory. An explicit
// "grite" or "todos" pins the backend (Grite is scoped to the workspace's git
// repo, "todos" to its .todos files); "" or "auto" auto-detects. dir is always
// the workspace directory, never a .todos path.
func providerForDir(dir, provider string) todos.Provider {
	switch provider {
	case todos.ProviderGrite:
		return resolveGrite(dir)
	case todos.ProviderFiles:
		return todos.NewFileProvider(dir, "")
	default:
		return autoTodoProvider(dir)
	}
}

// autoTodoProvider resolves a directory's todo provider: a local .todos file
// store if present, otherwise Grite (which tracks issues globally per repo and
// needs no per-directory marker, so it must not be gated on a .grite dir).
func autoTodoProvider(dir string) todos.Provider {
	if _, err := os.Stat(filepath.Join(dir, ".todos")); err == nil {
		return todos.NewFileProvider(dir, "")
	}
	return resolveGrite(dir)
}

// resolveTodoBackend reports which todo backend dir would use for the configured
// value, and whether that result was auto-detected (configured empty / "auto").
// It mirrors providerForDir's choice without building a provider.
func resolveTodoBackend(dir, configured string) (name string, auto bool) {
	switch configured {
	case todos.ProviderGrite, todos.ProviderFiles:
		return configured, false
	}
	if dir != "" {
		if _, err := os.Stat(filepath.Join(dir, ".todos")); err == nil {
			return todos.ProviderFiles, true
		}
	}
	return todos.ProviderGrite, true
}

// TodoBackendLabel renders the resolved todo backend for display, suffixing
// " (auto)" when it was auto-detected rather than pinned in the project config.
func TodoBackendLabel(dir, configured string) string {
	name, auto := resolveTodoBackend(dir, configured)
	if auto {
		return name + " (auto)"
	}
	return name
}

// resolveGrite returns a DB-backed Grite provider when the gavel cache is
// configured (reads served from the DB, kept fresh by `grite export --since`),
// falling back to direct grite CLI calls when no DB is available.
func resolveGrite(dir string) todos.Provider {
	return todos.ResolveGriteProvider(dir, cache.Shared(), 0)
}

func (s *Server) todoWorkDir() string {
	if s != nil && s.ghOpts.WorkDir != "" {
		return s.ghOpts.WorkDir
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func todoSourceFromRequest(r *http.Request) todoSource {
	provider := strings.TrimSpace(r.URL.Query().Get("provider"))
	if provider == "" {
		provider = strings.TrimSpace(r.URL.Query().Get("todoProvider"))
	}
	return todoSource{
		Provider: provider,
		Dir:      strings.TrimSpace(r.URL.Query().Get("dir")),
	}
}

func todoFiltersFromRequest(r *http.Request) (todos.DiscoveryFilters, error) {
	status := types.Status(strings.TrimSpace(r.URL.Query().Get("status")))
	if status == "" {
		return todos.DiscoveryFilters{}, nil
	}
	if !validTodoStatus(status) {
		return todos.DiscoveryFilters{}, fmt.Errorf("invalid status %q", status)
	}
	return todos.DiscoveryFilters{IncludeStatuses: []types.Status{status}}, nil
}

// reconcileSessionStatus surfaces a failed todo as in-progress while its recorded
// agent session is executing again — a resumed run (or a continued cmux REPL)
// reflects the live work instead of a stale "failed" until the provider's own
// status write catches up. Only failed todos that carry a session are inspected,
// so idle/finished rows never read a session log. A missing or unreadable log is
// the normal "no live session" case, so it leaves the persisted status untouched.
func reconcileSessionStatus(todo *types.TODO, dir string) {
	if todo == nil || todo.Status != types.StatusFailed || todo.LLM == nil || todo.LLM.SessionId == "" {
		return
	}
	path, err := cmux.SessionLogPath(dir, todo.LLM.SessionId)
	if err != nil {
		return
	}
	stats, err := cmux.GlobalSessionStats().Get(todo.LLM.SessionId, path)
	if err != nil {
		return
	}
	if stats.Executing(time.Now()) {
		todo.Status = types.StatusInProgress
	}
}

func summarizeTodo(todo *types.TODO, detail bool) todoSummary {
	if todo == nil {
		return todoSummary{}
	}
	title := todo.Title
	if title == "" {
		title = todo.Filename()
	}
	out := todoSummary{
		Ref:           todos.TODOReference(todo),
		ID:            todo.ID,
		ShortID:       todo.DisplayID(),
		Title:         title,
		Status:        todo.Status,
		Priority:      todo.Priority,
		Provider:      todo.Provider,
		ProviderState: todo.ProviderState,
		FilePath:      todo.FilePath,
		CWD:           todo.CWD,
		Labels:        todo.Labels,
		Attempts:      todo.Attempts,
		Created:       todo.Created,
		LastRun:       todo.LastRun,
	}
	if todo.LLM != nil {
		out.SessionID = todo.LLM.SessionId
	}
	if out.Ref == "" {
		out.Ref = todo.FilePath
	}
	if detail {
		out.Body = strings.TrimSpace(todo.MarkdownBody)
		out.Implementation = strings.TrimSpace(todo.Implementation)
		out.Events = todo.ProviderEvents
		out.Criteria = todo.AcceptanceCriteria
		if out.Body == "" {
			out.Body = out.Implementation
		}
	}
	return out
}

func summarizeTodos(items types.TODOS) todoCounts {
	var counts todoCounts
	for _, item := range items {
		counts.Total++
		switch item.Status {
		case types.StatusCompleted:
			counts.Completed++
		case types.StatusDraft:
			counts.Open++
			counts.Draft++
		case types.StatusInProgress:
			counts.Open++
			counts.InProgress++
		case types.StatusFailed:
			counts.Open++
			counts.Failed++
		case types.StatusVerified:
			counts.Open++
			counts.Verified++
		case types.StatusSkipped:
			counts.Open++
			counts.Skipped++
		default:
			counts.Open++
			counts.Pending++
		}
	}
	return counts
}

func validTodoStatus(status types.Status) bool {
	return types.IsKnownStatus(status)
}

func validTodoPriority(priority types.Priority) bool {
	switch priority {
	case types.PriorityHigh, types.PriorityMedium, types.PriorityLow:
		return true
	default:
		return false
	}
}

func writeTodoError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	json.NewEncoder(w).Encode(map[string]string{"error": msg}) //nolint:errcheck
}

func parseTodoNewPayload(r *http.Request) (todoNewPayload, []todoAttachmentSummary, error) {
	var payload todoNewPayload
	var attachments []todoAttachmentSummary
	contentType := strings.ToLower(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "multipart/form-data"):
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return payload, nil, fmt.Errorf("invalid multipart form: %w", err)
		}
		if r.MultipartForm != nil {
			if err := applyTodoNewValues(&payload, r.MultipartForm.Value, true); err != nil {
				return payload, nil, err
			}
			stored, err := persistMultipartAttachments(r.MultipartForm)
			if err != nil {
				return payload, nil, err
			}
			attachments = stored
		}
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			return payload, nil, fmt.Errorf("invalid form: %w", err)
		}
		if err := applyTodoNewValues(&payload, r.PostForm, true); err != nil {
			return payload, nil, err
		}
	case strings.HasPrefix(contentType, "application/json"):
		if r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				return payload, nil, fmt.Errorf("invalid json")
			}
		}
	case contentType == "":
		// Query-only create requests are valid.
	default:
		return payload, nil, fmt.Errorf("unsupported content type %q", r.Header.Get("Content-Type"))
	}

	if err := applyTodoNewValues(&payload, r.URL.Query(), false); err != nil {
		return payload, nil, err
	}
	return payload, attachments, nil
}

func applyTodoNewValues(payload *todoNewPayload, values map[string][]string, overwrite bool) error {
	assignString := func(target *string, keys ...string) {
		if !overwrite && strings.TrimSpace(*target) != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = value
		}
	}
	assignPriority := func(target *types.Priority, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Priority(value)
		}
	}
	assignStatus := func(target *types.Status, keys ...string) {
		if !overwrite && *target != "" {
			return
		}
		if value := firstTodoNewValue(values, keys...); value != "" {
			*target = types.Status(value)
		}
	}

	assignString(&payload.Provider, "provider", "todoProvider")
	assignString(&payload.Dir, "dir")
	assignString(&payload.Title, "title", "name")
	assignString(&payload.Body, "body", "description", "text")
	assignPriority(&payload.Priority, "priority", "severity")
	assignStatus(&payload.Status, "status")
	if !overwrite && payload.AutoSave != nil {
		return nil
	}
	if raw := firstTodoNewValue(values, "autoSave", "autosave", "auto_save"); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid autoSave %q", raw)
		}
		payload.AutoSave = &parsed
	}
	return nil
}

func firstTodoNewValue(values map[string][]string, keys ...string) string {
	for _, key := range keys {
		for _, value := range values[key] {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

// normalizeTodoRunRefs collects the todo refs to run, de-duplicated and in order:
// the explicit refs[] (multi-select), then the single ref, then the ?ref query
// param. Multiple refs run together in a single agent session.
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

func todoRunRefs(todoList []*types.TODO) []string {
	refs := make([]string, len(todoList))
	for i, todo := range todoList {
		refs[i] = todos.TODOReference(todo)
	}
	return refs
}

func todoRunStartedMessage(count int) string {
	if count > 1 {
		return fmt.Sprintf("Started run for %d todos", count)
	}
	return "Todo run started"
}

func todoRunLabel(todoList []*types.TODO) string {
	if len(todoList) == 1 {
		return todos.TODOReference(todoList[0])
	}
	return fmt.Sprintf("%d todos", len(todoList))
}

func normalizeTodoRunOptions(payload todoRunPayload) (todoRunOptions, error) {
	kind, err := resolveDriverFromPayload(payload)
	if err != nil {
		return todoRunOptions{}, err
	}
	agent := kind.Agent()
	model := strings.TrimSpace(payload.Model)
	if model == "" {
		model = agent
	}
	if got, _ := cmux.ResolveAgent(model); got != agent {
		return todoRunOptions{}, fmt.Errorf("driver %s expects a %s model but %q resolves to %s", kind, agent, model, got)
	}

	// Mode is the legacy mechanism label (cmux vs inline) some response/preview
	// paths still read; the authoritative selection is Driver.
	mode := "inline"
	if kind.Mechanism() == "cmux" {
		mode = "cmux"
	}
	if payload.Plan && kind.Mechanism() != "cmux" {
		return todoRunOptions{}, fmt.Errorf("plan mode requires a cmux driver")
	}

	effort := strings.ToLower(strings.TrimSpace(payload.Effort))
	if effort == "" {
		effort = "medium"
	}
	switch effort {
	case "low", "medium", "high":
	default:
		return todoRunOptions{}, fmt.Errorf("invalid effort %q", payload.Effort)
	}

	timeout := 30 * time.Minute
	if strings.TrimSpace(payload.Timeout) != "" {
		parsed, err := time.ParseDuration(strings.TrimSpace(payload.Timeout))
		if err != nil {
			return todoRunOptions{}, fmt.Errorf("invalid timeout %q", payload.Timeout)
		}
		if parsed <= 0 {
			return todoRunOptions{}, fmt.Errorf("timeout must be greater than zero")
		}
		timeout = parsed
	}

	maxBudget := payload.MaxBudget
	if maxBudget == 0 {
		maxBudget = payload.MaxCost
	}
	if maxBudget < 0 {
		return todoRunOptions{}, fmt.Errorf("max cost must be greater than or equal to zero")
	}
	if payload.MaxTurns < 0 {
		return todoRunOptions{}, fmt.Errorf("max turns must be greater than or equal to zero")
	}

	// Auto-commit defaults on (matching the CLI's `todos run --commit`); a nil
	// pointer means the client did not opt out.
	commit := true
	if payload.Commit != nil {
		commit = *payload.Commit
	}

	// The post-completion check loop is opt-in (matching `todos run --check`).
	check := false
	if payload.Check != nil {
		check = *payload.Check
	}

	// Post-commit issue verification is opt-in (matching `todos run --verify`).
	verify := false
	if payload.Verify != nil {
		verify = *payload.Verify
	}

	return todoRunOptions{
		Agent:           agent,
		Mode:            mode,
		Driver:          string(kind),
		Model:           model,
		Effort:          effort,
		Plan:            payload.Plan,
		Resume:          payload.Resume,
		Timeout:         timeout,
		MaxBudget:       maxBudget,
		MaxTurns:        payload.MaxTurns,
		Dirty:           payload.Dirty,
		DryRun:          payload.DryRun,
		Commit:          commit,
		Check:           check,
		Verify:          verify,
		TimeoutOriginal: payload.Timeout,
	}, nil
}

func defaultStartTodoRun(req todoRunRequest) error {
	executor, sessionID, err := newTodoRunExecutor(req)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), req.Options.Timeout)
		defer cancel()

		execCtx := todos.NewExecutorContext(ctx, logger.StandardLogger(), nil)
		runner := todos.NewTODOExecutor(req.Source.Dir, executor, sessionID, req.Provider)
		runner.EnableChecks(req.Options.Check)
		var runErr error
		var result *todos.ExecutionResult
		// A single selection runs through Execute; a multi-select runs every todo
		// in one combined agent session via ExecuteGroup.
		if len(req.Todos) == 1 {
			result, runErr = runner.Execute(execCtx, req.Todos[0])
		} else {
			var results []*todos.ExecutionResult
			results, runErr = runner.ExecuteGroup(execCtx, req.Todos)
			if len(results) > 0 {
				result = results[0]
			}
		}
		if runErr != nil {
			logger.Warnf("todo run %s failed: %v", todoRunLabel(req.Todos), runErr)
		}
		maybeCommitAfterRun(req, result)
	}()
	return nil
}

// maybeCommitAfterRun runs the `gavel commit` pipeline over the agent's changes
// once a dashboard run finishes, mirroring the CLI's `todos run --commit`, then
// runs issue verification over the resulting commits when enabled. Auto-commit
// is skipped for plan-only runs (which make no changes) and whenever the
// executor already committed its own changes (see todos.ShouldCommitAfter).
func maybeCommitAfterRun(req todoRunRequest, result *todos.ExecutionResult) {
	if len(req.Todos) == 0 || req.Todos[0] == nil {
		return
	}
	todo := req.Todos[0]

	var hashes []string
	if enabled := req.Options.Commit && !req.Options.Plan; todos.ShouldCommitAfter(result, enabled) {
		meta := commit.AgentRunMetadata{IssueID: todo.ID}
		if todo.LLM != nil {
			meta.SessionID = todo.LLM.SessionId
		}
		commitResult, err := commit.RunAfterAgent(context.Background(), req.Source.Dir, todo.CWD, meta)
		if err != nil {
			logger.Errorf("commit after todo run failed: %v", err)
		}
		hashes = commitHashes(commitResult)
	}

	if req.Options.Verify && result != nil && result.Success {
		verifyAfterRun(req, todo, hashes)
	}
}

// verifyAfterRun scores the committed work against the issue's acceptance
// criteria and records the verdict. It is advisory — failures are logged.
func verifyAfterRun(req todoRunRequest, todo *types.TODO, hashes []string) {
	if _, err := todos.RunIssueVerification(context.Background(), req.Provider, todo, todos.VerifyOptions{
		WorkDir: todoVerifyWorkDir(req.Source.Dir, todo),
		Commits: hashes,
	}); err != nil {
		logger.Warnf("issue verification after todo run skipped: %v", err)
	}
}

// commitHashes extracts the commit hashes from a commit result.
func commitHashes(result *commit.Result) []string {
	if result == nil {
		return nil
	}
	var out []string
	for _, c := range result.Commits {
		if c.Hash != "" {
			out = append(out, c.Hash)
		}
	}
	return out
}

// resolveDriverFromPayload selects the driver kind: the explicit Driver field
// when set, otherwise the legacy agent+mode pair (mode cmux → "<agent>-cmux",
// mode inline → claude-sdk; codex was never an inline agent).
func resolveDriverFromPayload(p todoRunPayload) (drivers.Kind, error) {
	if s := strings.TrimSpace(p.Driver); s != "" {
		return drivers.Parse(s)
	}
	agent := strings.ToLower(strings.TrimSpace(p.Agent))
	if agent == "" {
		agent, _ = cmux.ResolveAgent(strings.TrimSpace(p.Model))
	}
	mode := strings.ToLower(strings.TrimSpace(p.Mode))
	if mode == "" {
		mode = "cmux"
	}
	switch mode {
	case "cmux":
		return drivers.Parse(agent + "-cmux")
	case "inline":
		if agent == "codex" {
			return "", fmt.Errorf("codex runs require a cmux driver")
		}
		return drivers.ClaudeSDK, nil
	default:
		return "", fmt.Errorf("invalid mode %q", p.Mode)
	}
}

func newTodoRunExecutor(req todoRunRequest) (todos.Executor, string, error) {
	kind, err := drivers.Parse(req.Options.Driver)
	if err != nil {
		return nil, "", err
	}
	// cmux returns "" as the orchestrator session id (it manages its own
	// --session-id, passed via SessionID) so TODOExecutor does not overwrite the
	// todo's recorded prior session.
	return drivers.New(kind, drivers.Config{
		WorkDir:      req.Source.Dir,
		Model:        req.Options.Model,
		Effort:       req.Options.Effort,
		Plan:         req.Options.Plan,
		Resume:       req.Options.Resume,
		SessionID:    req.Options.SessionID,
		Timeout:      req.Options.Timeout,
		MaxBudgetUsd: req.Options.MaxBudget,
		MaxTurns:     req.Options.MaxTurns,
		Dirty:        req.Options.Dirty,
	})
}

// resolveRunSessionID determines the claude session id a run will use, so the
// caller knows it up front. A resume run reuses the todo's prior session; a
// fresh cmux run mints a new id (claude is launched with it, so the dashboard
// can follow the log immediately); inline resumes a single todo's session if it
// has one and otherwise lets claude manage its own id.
func resolveRunSessionID(opts todoRunOptions, todoList []*types.TODO) string {
	if opts.Resume {
		if sid := firstTodoSessionID(todoList); sid != "" {
			return sid
		}
	}
	switch opts.Mode {
	case "cmux":
		return uuid.NewString()
	case "inline":
		if len(todoList) == 1 {
			return firstTodoSessionID(todoList)
		}
	}
	return ""
}

func firstTodoSessionID(todoList []*types.TODO) string {
	for _, todo := range todoList {
		if todo != nil && todo.LLM != nil && todo.LLM.SessionId != "" {
			return todo.LLM.SessionId
		}
	}
	return ""
}

func countProjectTodos(ctx context.Context, dir, provider string) todoCounts {
	if dir == "" {
		return todoCounts{}
	}
	items, err := providerForDir(dir, provider).List(ctx, todos.DiscoveryFilters{})
	if err != nil {
		logger.Debugf("count todos in %s: %v", dir, err)
		return todoCounts{}
	}
	return summarizeTodos(items)
}
