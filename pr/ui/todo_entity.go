package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	todoentity "github.com/flanksource/gavel/todos/entity"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"gorm.io/gorm"
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
// The default workspace — the one an action naming none acts on — and the
// registered projects an unscoped list reads are resolved per invocation rather
// than captured here. The registration is process-global
// and sync.Once'd, so anything read at registration time is frozen for the life
// of the process and for every surface at once: a CLI invocation that later
// `cd`s elsewhere, or a dashboard started before the directory existed, would
// act on a workspace nobody chose. The dashboard's own Server-scoped override
// (ghOpts.WorkDir) still cannot be consulted for the same reason, and it falls
// back to the working directory anyway.
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
			Workspaces: todoEntityWorkspaces,
			DefaultDir: todoEntityDefaultDir,
			ResolveRun: resolveBulkRunOptions,
			Broker: func(_ context.Context, dir string) todos.ApprovalBroker {
				return todoApprovalBroker(dir)
			},
			PushBaseURL: func(dir, requested string) (string, error) {
				return resolveTodoPushBaseURL(requested, dir, "")
			},
			Workspace:  todoEntityWorkspace,
			ProjectDir: todoEntityProjectDir,
			DB:         func(ctx context.Context) (*gorm.DB, error) { return database.Require(ctx, "gavel todos") },
		})
	})
	return todoEntityErr
}

// todoEntityWorkspaces lists the registered projects an unscoped TODO list
// reads, read per call so a project registered after startup is listed.
func todoEntityWorkspaces(context.Context) ([]query.Workspace, error) {
	projects, err := LoadProjects()
	if err != nil {
		return nil, err
	}
	return TodoWorkspaces(projects), nil
}

// todoEntityDefaultDir is the workspace an action naming none acts on.
func todoEntityDefaultDir(context.Context) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve the working directory for the todos entity: %w", err)
	}
	return dir, nil
}

// todoEntityWorkspace resolves a directory to the registered project's
// workspace options, which is what the portable import/export address rows by.
func todoEntityWorkspace(_ context.Context, dir string) (todoruntime.WorkspaceOptions, error) {
	project, err := ProjectForDir(dir)
	if err != nil {
		return todoruntime.WorkspaceOptions{}, err
	}
	return project.WorkspaceOptions(), nil
}

// todoEntityProjectDir resolves a registered project name to its absolute
// directory, so `transfer` can name a destination the way the projects list
// displays it rather than by path.
func todoEntityProjectDir(_ context.Context, name string) (string, error) {
	project, err := GetProject(name)
	if err != nil {
		return "", err
	}
	dir, err := filepath.Abs(project.ResolvedDir())
	if err != nil {
		return "", fmt.Errorf("resolve project %q directory: %w", name, err)
	}
	return dir, nil
}

// resolveBulkRunOptions resolves a bulk run the way the dashboard's own single
// run does: the batch's flags validated at the wire boundary and folded as the
// request layer, dispatched as the dashboard host so a bulk run brokers
// approvals exactly like the single run started from the same page.
func resolveBulkRunOptions(_ context.Context, req bulk.RunRequest) (run.Options, error) {
	presets := append([]string(nil), req.Flags.Presets...)
	if req.Flags.NoPresets {
		presets = []string{}
	}
	return buildTodoRunOptions(todoRunPayload{
		Presets: presets,
		Dir:     req.Dir,
		Ref:     todos.TODOReference(req.Todo),
		Step:    req.Step,
		Spec:    req.Flags.Spec(),
		Resume:  req.Flags.Resume,
	}, nil)
}
