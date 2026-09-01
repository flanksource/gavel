package ui

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	todoentity "github.com/flanksource/gavel/todos/entity"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

// The todos entity is what replaced the hand-written bulk and triage handlers.
// Declaring an action once now gives it a CLI command, a REST route, an OpenAPI
// operation and a catalog entry that the dashboard renders its selection
// toolbar from — so a new bulk action reaches the UI without any React change.
//
// Registration is process-global because the clicky entity registry is, and the
// closures it stores must therefore not capture a *Server. They do not need to:
// openTodoProvider and openGlobalTodoProvider are already package-level seams
// (and package tests swap them), and the run registry is deliberately
// process-wide so a run started here is stoppable from the dashboard.

var (
	todoEntityOnce sync.Once
	todoEntityErr  error
)

// registerTodoEntity declares the entity exactly once. Registering twice would
// duplicate every generated command and route.
func registerTodoEntity() error {
	todoEntityOnce.Do(func() {
		todoEntityErr = todoentity.Register(todoentity.Deps{
			OpenProvider: func(ctx context.Context, dir string) (todos.Provider, error) {
				return openTodoProvider(ctx, dir)
			},
			OpenGlobal: func(ctx context.Context) (todos.GlobalReferenceProvider, error) {
				return openGlobalTodoProvider(ctx)
			},
			Registry:   run.Shared(),
			DefaultDir: todoEntityDefaultDir,
			ResolveRun: resolveBulkRunOptions,
		})
	})
	return todoEntityErr
}

// resolveBulkRunOptions resolves a bulk run the way the dashboard's own single
// run resolves: through normalizeTodoRunOptions, which applies the (driver,
// mode) catalog and admits the run to ask for a Bash approval because the
// dashboard serves /api/todos/session/approve. Without this a bulk run would
// quietly differ from the single run started from the same page.
func resolveBulkRunOptions(_ context.Context, req bulk.RunRequest) (run.Options, error) {
	payload := todoRunPayload{
		Dir:    req.Dir,
		Ref:    todos.TODOReference(req.Todo),
		Prompt: req.Prompt,
		Spec:   req.Flags.Spec(),
		Driver: req.Flags.Driver,
		Resume: req.Flags.Resume,
	}
	return normalizeTodoRunOptions(req.Dir, []*types.TODO{req.Todo}, payload)
}

// todoEntityDefaultDir is the workspace a request that names none acts on. The
// dashboard's own Server-scoped override (ghOpts.WorkDir) cannot be consulted
// here — the registration is process-global — and it falls back to the working
// directory anyway, which is what a CLI invocation means by "here".
func todoEntityDefaultDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// todoEntityRoutes builds the generated REST surface: POST
// /api/v1/todo/{id}/{action} for each bulk action, plus GET /api/entities,
// which is the catalog the dashboard's selection toolbar is derived from.
//
// The OpenAPI handlers are deliberately not mounted here — pr/ui already serves
// /api/openapi.json from its own merged document.
func (s *Server) registerTodoEntityRoutes(mux *http.ServeMux) {
	if err := registerTodoEntity(); err != nil {
		// A failed registration means the dashboard would silently serve a
		// toolbar with no actions behind it.
		panic("ui: registering the todos entity: " + err.Error())
	}
	root := &cobra.Command{Use: "gavel"}
	clicky.GenerateCLI(root)

	server := rpc.NewSwaggerServer(&rpc.ServeConfig{
		Title:      "gavel",
		SkipHealth: true,
		Executor:   &rpc.ExecutorConfig{Enabled: true, PathPrefix: "/api/v1"},
	}, root, nil)
	server.RegisterExecutionRoutes(mux)
	mux.HandleFunc("GET /api/entities", server.HandleEntities)
}
