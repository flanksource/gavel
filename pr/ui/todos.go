package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type todoCounts struct {
	Total      int `json:"total"`
	Open       int `json:"open"`
	Pending    int `json:"pending"`
	InProgress int `json:"inProgress"`
	Failed     int `json:"failed"`
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
	LastRun        *time.Time            `json:"lastRun,omitempty"`
	Body           string                `json:"body,omitempty"`
	Implementation string                `json:"implementation,omitempty"`
	Events         []types.ProviderEvent `json:"events,omitempty"`
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

type todoUpdatePayload struct {
	Provider string         `json:"provider,omitempty"`
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
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
	resp := todoListResponse{
		Provider: source.Provider,
		Dir:      source.Dir,
		Counts:   summarizeTodos(items),
		Items:    make([]todoSummary, 0, len(items)),
	}
	for _, item := range items {
		resp.Items = append(resp.Items, summarizeTodo(item, false))
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (s *Server) handleTodoGet(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
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
	// A PATCH may set status, priority, or both; at least one is required.
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
	if update.Status == nil && update.Priority == nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("status or priority is required"))
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
	if err := provider.UpdateState(r.Context(), todo, update); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
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

func (s *Server) todoProvider(source todoSource) (todos.Provider, todoSource, error) {
	workDir := s.todoWorkDir()
	dir := source.Dir
	if dir == "" {
		dir = workDir
	} else if !filepath.IsAbs(dir) {
		dir = filepath.Join(workDir, dir)
	}
	source.Dir = dir
	switch source.Provider {
	case "", providerAuto, todos.ProviderGrite, todos.ProviderFiles:
		return providerForDir(dir, source.Provider), source, nil
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
		LastRun:       todo.LastRun,
	}
	if out.Ref == "" {
		out.Ref = todo.FilePath
	}
	if detail {
		out.Body = strings.TrimSpace(todo.MarkdownBody)
		out.Implementation = strings.TrimSpace(todo.Implementation)
		out.Events = todo.ProviderEvents
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
		case types.StatusInProgress:
			counts.Open++
			counts.InProgress++
		case types.StatusFailed:
			counts.Open++
			counts.Failed++
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
	switch status {
	case types.StatusPending, types.StatusInProgress, types.StatusCompleted, types.StatusFailed, types.StatusSkipped:
		return true
	default:
		return false
	}
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
