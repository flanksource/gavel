package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

type todoSource struct {
	Dir string
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
	provider, source, err := s.todoProviderContext(r.Context(), source)
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
		Dir:    source.Dir,
		Counts: summarizeTodos(items),
		Items:  make([]todoSummary, 0, len(items)),
	}
	stats := commitDiffStats(r.Context(), source.Dir)
	for _, item := range items {
		sum := summarizeTodo(item, false)
		sum.Diff = diffStatFor(stats, item.ID)
		resp.Items = append(resp.Items, sum)
	}
	json.NewEncoder(w).Encode(resp) //nolint:errcheck
}

func (s *Server) handleTodoGet(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	requestedSource := todoSourceFromRequest(r)
	provider, source, todo, lookupSessionID, err := s.resolveTodoGetReference(r.Context(), requestedSource, ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	// The detail view's step strip and run picker read the lifecycle from here:
	// a todo whose lifecycle cannot be evaluated is a misconfigured workspace,
	// reported rather than rendered without its steps.
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("evaluate lifecycle for %s: %w", ref, err))
		return
	}
	sum.LookupSessionID = lookupSessionID
	sum.Diff = diffStatFor(commitDiffStats(r.Context(), source.Dir), todo.ID)
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

// resolveTodoGetReference extends the ordinary global issue lookup with an
// exact UUID-only session fallback. Issue UUIDs and UUID-shaped aliases remain
// authoritative; only a genuine issue miss is allowed to resolve through
// Captain's prompt-run links. The canonical issue is then re-read through its
// owning workspace provider so all detail fields retain their normal boundary.
func (s *Server) resolveTodoGetReference(
	ctx context.Context,
	requested todoSource,
	ref string,
) (todos.Provider, todoSource, *types.TODO, string, error) {
	provider, source, todo, err := s.resolveTodoReference(ctx, requested, ref)
	if err == nil {
		return provider, source, todo, "", nil
	}
	if strings.TrimSpace(requested.Dir) != "" || !errors.Is(err, native.ErrNotFound) {
		return nil, source, nil, "", err
	}
	if _, parseErr := uuid.Parse(strings.TrimSpace(ref)); parseErr != nil {
		return nil, source, nil, "", err
	}

	global, globalErr := openGlobalTodoProvider(ctx)
	if globalErr != nil {
		return nil, source, nil, "", err
	}
	sessions, ok := global.(todos.GlobalSessionReferenceProvider)
	if !ok {
		return nil, source, nil, "", err
	}
	sessionTodo, sessionID, sessionErr := sessions.GetGlobalBySession(ctx, ref)
	if sessionErr != nil {
		if errors.Is(sessionErr, native.ErrNotFound) {
			return nil, source, nil, "", fmt.Errorf("%w: TODO or session UUID %q", native.ErrNotFound, ref)
		}
		return nil, source, nil, "", sessionErr
	}
	ownerDir := strings.TrimSpace(sessionTodo.CWD)
	if ownerDir == "" {
		return nil, source, nil, "", fmt.Errorf("resolved session %q has no owning workspace path", ref)
	}
	provider, source, err = s.todoProviderContext(ctx, todoSource{Dir: ownerDir})
	if err != nil {
		return nil, source, nil, "", err
	}
	todo, err = provider.Get(ctx, sessionTodo.ID)
	if err != nil {
		return nil, source, nil, "", err
	}
	return provider, source, todo, sessionID, nil
}

// resolveTodoReference loads one issue and returns a provider scoped to its
// owning workspace. With an explicit dir the lookup is workspace-local. With
// no dir it first resolves a UUID/short UUID/imported alias globally, then
// reopens the provider for the authoritative CWD so later mutations and run
// lifecycle writes cannot accidentally target the server's default workspace.
func (s *Server) resolveTodoReference(ctx context.Context, requested todoSource, ref string) (todos.Provider, todoSource, *types.TODO, error) {
	globalLookup := strings.TrimSpace(requested.Dir) == ""
	if !globalLookup {
		provider, source, err := s.todoProviderContext(ctx, requested)
		if err != nil {
			return nil, source, nil, err
		}
		todo, err := provider.Get(ctx, ref)
		return provider, source, todo, err
	}

	global, err := openGlobalTodoProvider(ctx)
	if err != nil {
		return nil, requested, nil, err
	}
	todo, err := global.GetGlobal(ctx, ref)
	if err != nil {
		return nil, requested, nil, err
	}
	ownerDir := strings.TrimSpace(todo.CWD)
	if ownerDir == "" {
		return nil, requested, nil, fmt.Errorf("resolved TODO %q has no owning workspace path", ref)
	}
	provider, source, err := s.todoProviderContext(ctx, todoSource{Dir: ownerDir})
	if err != nil {
		return nil, source, nil, err
	}
	// Re-read through the owning repository so the returned object and provider
	// share the same optimistic-version and workspace boundary.
	todo, err = provider.Get(ctx, todo.ID)
	if err != nil {
		return nil, source, nil, err
	}
	return provider, source, todo, nil
}

// resolveTodoDir turns a request's dir param into an absolute workspace path,
// defaulting to the server's work dir and joining relative dirs onto it.
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

func (s *Server) todoProviderContext(ctx context.Context, source todoSource) (todos.Provider, todoSource, error) {
	source.Dir = s.resolveTodoDir(source.Dir)
	provider, err := openTodoProvider(ctx, source.Dir)
	if err != nil {
		return nil, source, err
	}
	return provider, source, nil
}

// ProviderForProject resolves a stored project to the PostgreSQL runtime.
func ProviderForProject(ctx context.Context, p Project) (todos.Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return todoruntime.Open(ctx, p.WorkspaceOptions())
}

// todoRuns is the in-flight run registry the dashboard starts and stops runs
// through. It is the process-wide one rather than Server state because the CLI
// and the todos entity start runs through the same registry, and a run the
// dashboard cannot see is a run it cannot stop.
func todoRuns() *run.Registry { return run.Shared() }

// openTodoProvider is the single API/UI native runtime seam. It is a variable so
// package tests can inject an in-memory implementation without opening PostgreSQL.
var openTodoProvider = func(ctx context.Context, dir string) (todos.Provider, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	project, err := ProjectForDir(dir)
	if err != nil {
		return nil, err
	}
	return todoruntime.Open(ctx, project.WorkspaceOptions())
}

var openGlobalTodoProvider = func(ctx context.Context) (todos.GlobalReferenceProvider, error) {
	return todoruntime.OpenGlobal(ctx)
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
	return todoSource{
		Dir: strings.TrimSpace(r.URL.Query().Get("dir")),
	}
}
