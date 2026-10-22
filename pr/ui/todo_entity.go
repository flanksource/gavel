package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	todoentity "github.com/flanksource/gavel/todos/entity"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
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
				if s := entityServer(ctx); s != nil {
					provider, _, err := s.todoProviderContext(ctx, todoSource{Dir: dir})
					return provider, err
				}
				return openTodoProvider(ctx, dir)
			},
			OpenGlobal: func(ctx context.Context) (todos.GlobalReferenceProvider, error) {
				if s := entityServer(ctx); s != nil {
					return s.globalTodoProvider(ctx)
				}
				return openGlobalTodoProvider(ctx)
			},
			Registry:   run.Shared(),
			Workspaces: todoEntityWorkspaces,
			DefaultDir: todoEntityDefaultDir,
			ResolveRun: resolveBulkRunOptions,
			Approvals:  true,
			PushBaseURL: func(dir, requested string) (string, error) {
				return resolveTodoPushBaseURL(requested, dir, "")
			},
			Workspace:  todoEntityWorkspace,
			ProjectDir: todoEntityProjectDir,
			DB: func(ctx context.Context) (*gorm.DB, error) {
				if entityServer(ctx) != nil {
					return nil, errEntityHostFiles
				}
				return database.Require(ctx, "gavel todos")
			},
			AttachRun:      todoEntityAttachRun,
			LifecycleHost:  todoEntityLifecycleHost,
			LiteralContent: func(ctx context.Context) bool { return entityServer(ctx) != nil },
			GitHub:         todoEntityGitHub,
		})
	})
	return todoEntityErr
}

// hostedEntityKey carries the hosted Server a todo entity operation acts for.
type hostedEntityKey struct{}

// withEntityServer makes ctx's todo entity operations act for s, a hosted organization's
// dashboard, rather than for this host.
func withEntityServer(ctx context.Context, s *Server) context.Context {
	return context.WithValue(ctx, hostedEntityKey{}, s)
}

// entityServer is the hosted Server ctx's todo entity operation acts for, or nil when it acts
// for this host.
func entityServer(ctx context.Context) *Server {
	s, _ := ctx.Value(hostedEntityKey{}).(*Server)
	return s
}

// errEntityHostFiles refuses, for a hosted organization, a todo operation that reads or writes
// this host's files.
var errEntityHostFiles = errors.New("importing and exporting todos reads this host's files, which a hosted organization can't")

// todoEntityAttachRun sends a hosted organization's run to a sandbox; a host's own runs are
// left as they are.
func todoEntityAttachRun(ctx context.Context, req *run.Request) error {
	s := entityServer(ctx)
	if s == nil {
		return nil
	}
	_, source, err := s.todoProviderContext(ctx, todoSource{Dir: req.Dir})
	if err != nil {
		return err
	}
	return s.attachRemote(ctx, source, req)
}

// todoEntityLifecycleHost loads a hosted organization's lifecycle from its repository's own
// configuration, and a host's from its own.
func todoEntityLifecycleHost(ctx context.Context, provider todos.Provider, dir string, kind lifecycle.HostKind) (*lifecycle.Host, error) {
	s := entityServer(ctx)
	if s == nil {
		return lifecycle.NewHost(provider, dir, kind)
	}
	_, source, err := s.todoProviderContext(ctx, todoSource{Dir: dir})
	if err != nil {
		return nil, err
	}
	return lifecycle.NewRepositoryHost(provider, source.Dir, kind)
}

// todoEntityGitHub is the account a TODO in dir is pushed to GitHub through: this host's
// credentials and dir's origin, or a hosted organization's GitHub App on the todo's own
// repository (see todoGitHub).
func todoEntityGitHub(ctx context.Context, todo *types.TODO, dir, repo string) (github.Options, error) {
	s := entityServer(ctx)
	if s == nil {
		return github.Options{WorkDir: dir, Repo: repo}, nil
	}
	_, source, err := s.todoProviderContext(ctx, todoSource{Dir: dir})
	if err != nil {
		return github.Options{}, err
	}
	account, err := s.todoGitHub(ctx, source, todo)
	if err != nil {
		return github.Options{}, err
	}
	if repo != "" && repo != account.Repo {
		return github.Options{}, fmt.Errorf("a hosted todo is pushed to its own repository %s, not %s", account.Repo, repo)
	}
	return account, nil
}

// todoEntityWorkspaces lists the projects an unscoped TODO list reads: a hosted organization's
// repositories, or the registered projects, read per call so one registered after startup is
// listed.
func todoEntityWorkspaces(ctx context.Context) ([]query.Workspace, error) {
	if s := entityServer(ctx); s != nil {
		projects, err := s.projectList(ctx)
		if err != nil {
			return nil, err
		}
		workspaces := make([]query.Workspace, len(projects))
		for i, project := range projects {
			workspaces[i] = query.Workspace{Name: project.Name, Dir: project.Dir}
		}
		return workspaces, nil
	}
	projects, err := LoadProjects(LoadProjectsOptions{})
	if err != nil {
		return nil, err
	}
	return TodoWorkspaces(projects), nil
}

// todoEntityDefaultDir is the workspace an action naming none acts on: this host's working
// directory. A hosted organization has none.
func todoEntityDefaultDir(ctx context.Context) (string, error) {
	if entityServer(ctx) != nil {
		return "", fmt.Errorf("%w: a hosted organization's todo operations name one of its repositories", todoentity.ErrNoDefaultWorkspace)
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve the working directory for the todos entity: %w", err)
	}
	return dir, nil
}

// todoEntityWorkspace resolves a directory to the registered project's
// workspace options, which is what the portable import/export address rows by.
func todoEntityWorkspace(ctx context.Context, dir string) (todoruntime.WorkspaceOptions, error) {
	if entityServer(ctx) != nil {
		return todoruntime.WorkspaceOptions{}, errEntityHostFiles
	}
	project, err := ProjectForDir(dir)
	if err != nil {
		return todoruntime.WorkspaceOptions{}, err
	}
	return project.WorkspaceOptions(), nil
}

// todoEntityProjectDir resolves a registered project name to its absolute
// directory, so `transfer` can name a destination the way the projects list
// displays it rather than by path. A hosted organization's project is one of its
// repositories, named repo:<id>.
func todoEntityProjectDir(ctx context.Context, name string) (string, error) {
	if s := entityServer(ctx); s != nil {
		projects, err := s.projectList(ctx)
		if err != nil {
			return "", err
		}
		for _, project := range projects {
			if project.Name == name {
				return project.Dir, nil
			}
		}
		return "", fmt.Errorf("%w: %q is not one of this organization's repositories", ErrProjectNotFound, name)
	}
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
