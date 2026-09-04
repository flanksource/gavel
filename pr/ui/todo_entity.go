package ui

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/rpc"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	todoentity "github.com/flanksource/gavel/todos/entity"
	"github.com/flanksource/gavel/todos/run"
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
//
// The default workspace — the one a request naming none acts on — is the
// working directory, resolved here rather than per request. The dashboard's own
// Server-scoped override (ghOpts.WorkDir) cannot be consulted, because the
// registration is process-global, and it falls back to the working directory
// anyway, which is what a CLI invocation means by "here". A working directory
// that cannot be read fails the registration: a toolbar whose actions act on
// "." would be acting on a workspace nobody chose.
func registerTodoEntity() error {
	todoEntityOnce.Do(func() {
		workDir, err := os.Getwd()
		if err != nil {
			todoEntityErr = fmt.Errorf("resolve the working directory for the todos entity: %w", err)
			return
		}
		todoEntityErr = todoentity.Register(todoentity.Deps{
			OpenProvider: func(ctx context.Context, dir string) (todos.Provider, error) {
				return openTodoProvider(ctx, dir)
			},
			OpenGlobal: func(ctx context.Context) (todos.GlobalReferenceProvider, error) {
				return openGlobalTodoProvider(ctx)
			},
			Registry:   run.Shared(),
			DefaultDir: func() string { return workDir },
			ResolveRun: resolveBulkRunOptions,
			Broker:     todoApprovalBroker,
		})
	})
	return todoEntityErr
}

// resolveBulkRunOptions resolves a bulk run the way the dashboard's own single
// run does: the batch's flags validated at the wire boundary and folded as the
// request layer, dispatched as the dashboard host so a bulk run brokers
// approvals exactly like the single run started from the same page.
func resolveBulkRunOptions(_ context.Context, req bulk.RunRequest) (run.Options, error) {
	return buildTodoRunOptions(todoRunPayload{
		Dir:    req.Dir,
		Ref:    todos.TODOReference(req.Todo),
		Step:   req.Step,
		Spec:   req.Flags.Spec(),
		Resume: req.Flags.Resume,
	}, nil)
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
