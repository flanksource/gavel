// Package entity declares TODOs as a clicky entity, so one declaration
// generates the CLI commands, the REST routes, the OpenAPI spec and the action
// catalog a front end renders from.
//
// Before this, a TODO action had to be written three times — once as a Cobra
// command, once as an HTTP handler, once in React — and the three drifted:
// the CLI could act on many TODOs, the API mostly could not, and the dashboard
// offered three of the dozen actions that existed. Registering the actions in
// one place is what makes "every registered action is executable" true by
// construction rather than by discipline.
package entity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/merge"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"gorm.io/gorm"
)

// Deps are the things the entity cannot resolve for itself.
//
// OpenProvider is injected because resolving a workspace directory to its
// provider goes through the projects registry, which lives with the dashboard's
// configuration rather than in the TODO domain. Registry is injected because a
// process must share one in-flight run map — the entity starting a run behind
// the dashboard's back would let the same TODO run twice.
type Deps struct {
	OpenProvider func(ctx context.Context, dir string) (todos.Provider, error)
	// OpenGlobal resolves a reference that names no workspace. A selection made
	// in the dashboard is grouped by severity or age rather than by repository,
	// so its refs regularly span workspaces and cannot be resolved against any
	// single one.
	OpenGlobal func(ctx context.Context) (todos.GlobalReferenceProvider, error)
	Registry   *run.Registry
	// Workspaces lists the registered projects. A list that names neither a
	// directory nor a project reads all of them, and a project named by a list
	// is looked up here.
	Workspaces func(ctx context.Context) ([]query.Workspace, error)
	// DefaultDir supplies the workspace when an action names none. The list does
	// not use it: an unscoped list reads every registered project.
	DefaultDir func(context.Context) (string, error)
	// ResolveRun turns a TODO plus the batch's overrides into run options.
	// Optional: bulk.DefaultRunResolver is used when unset. The dashboard
	// supplies its own because it applies a runtime catalog the CLI has no
	// equivalent of, and because it resolves as the approval-serving host.
	//
	// It is consulted only on an attended surface — see runtime. A registration
	// is process-global and shared by every surface, so a dep that is right for
	// one of them cannot be frozen in as the answer for all of them.
	ResolveRun bulk.RunResolver
	// Broker answers a batched run's tool-permission requests. Optional, and
	// nil is the unattended answer: a terminal batch has no one to ask, so a run
	// it starts must never be configured to. Like ResolveRun it is consulted
	// only on an attended surface.
	Broker func(ctx context.Context, dir string) todos.ApprovalBroker
	// PushBaseURL resolves the attachment origin for a pushed TODO's workspace.
	// Optional: without it only an explicit --base-url is honoured.
	PushBaseURL bulk.PushBaseURL

	// Workspace resolves a directory to the registered project's workspace
	// options. Required by import and export, which address rows by workspace
	// rather than by provider. Optional: those two actions are registered only
	// when it and DB are both supplied.
	Workspace func(ctx context.Context, dir string) (todoruntime.WorkspaceOptions, error)
	// DB opens the shared TODO database. Required by import and export, which
	// read and write whole rows rather than going through a provider.
	DB func(ctx context.Context) (*gorm.DB, error)
	// ProjectDir resolves a registered project name to its absolute directory.
	// Required by transfer, which names its destination the way the projects
	// list displays it. Optional: transfer is registered only when it is set.
	ProjectDir func(ctx context.Context, name string) (string, error)
}

func (d Deps) pushBaseURL() bulk.PushBaseURL {
	if d.PushBaseURL != nil {
		return d.PushBaseURL
	}
	return func(_, requested string) (string, error) { return githubpush.ResolveBaseURL(requested) }
}

func (d Deps) validate() error {
	if d.OpenProvider == nil {
		return fmt.Errorf("entity deps: OpenProvider is required")
	}
	if d.OpenGlobal == nil {
		return fmt.Errorf("entity deps: OpenGlobal is required")
	}
	if d.Registry == nil {
		return fmt.Errorf("entity deps: Registry is required")
	}
	if d.Workspaces == nil {
		return fmt.Errorf("entity deps: Workspaces is required")
	}
	return nil
}

// attended reports whether the caller can answer a tool-permission prompt.
//
// The entity is registered once per process and then reached from every
// surface, so "can this run ask for approval?" cannot be decided at
// registration. HTTP and MCP callers sit behind a dashboard or an agent loop
// that serves an approval endpoint; a CLI invocation and a scheduled operation
// do not, and a run of theirs that blocks on a prompt hangs a terminal or a
// cron slot with nobody watching. A direct in-process call (no surface set) is
// treated as unattended for the same reason.
func attended(ctx context.Context) bool {
	switch entity.OperationSurfaceFromContext(ctx) {
	case "http", "mcp":
		return true
	default:
		return false
	}
}

// runtime picks the run resolution and the approval broker for this
// invocation's surface. Unattended callers always get the plain resolution and
// no broker, whatever the deps were registered with.
func (d Deps) runtime(ctx context.Context, dir string) (bulk.RunResolver, todos.ApprovalBroker) {
	if !attended(ctx) {
		return bulk.DefaultRunResolver, nil
	}
	resolve := d.ResolveRun
	if resolve == nil {
		resolve = bulk.DefaultRunResolver
	}
	if d.Broker == nil {
		return resolve, nil
	}
	return resolve, d.Broker(ctx, dir)
}

func (d Deps) dir(ctx context.Context, opts query.ListOpts) (string, error) {
	if dir := strings.TrimSpace(opts.Dir); dir != "" {
		return dir, nil
	}
	if d.DefaultDir != nil {
		return d.DefaultDir(ctx)
	}
	return "", fmt.Errorf("entity deps: no workspace directory")
}

// Register declares the todos entity. Call it once at startup, before
// entity.GenerateCLI or the RPC server reads the registry.
func Register(deps Deps) error {
	if err := deps.validate(); err != nil {
		return err
	}
	builder := clicky.NewEntity[Summary, query.ListOpts, *types.TODO]("todo").
		Aliases("todos").
		ListPagedWithContext(deps.list).
		GetWithContext(deps.get)

	for _, action := range deps.bulkActions() {
		builder = builder.WithBulkAction(action)
	}
	for _, action := range deps.itemActions() {
		builder = builder.WithAction(action)
	}
	builder.Register()
	return nil
}

// list reads the workspaces the request scopes to and returns the requested
// window of the TODOs that match, as summaries, with the total that matched.
func (d Deps) list(ctx context.Context, opts query.ListOpts) (clicky.PagedResult[Summary], error) {
	workspaces, err := d.listWorkspaces(ctx, opts)
	if err != nil {
		return clicky.PagedResult[Summary]{}, err
	}
	matched, err := opts.Select(workspacesLister{ctx: ctx, workspaces: workspaces, open: d.OpenProvider}, time.Now())
	if err != nil {
		return clicky.PagedResult[Summary]{}, err
	}
	page, err := opts.Page(matched)
	if err != nil {
		return clicky.PagedResult[Summary]{}, rejected(err)
	}
	return clicky.NewPagedResult(summarizeAll(page), opts.Limit, opts.Offset, int64(len(matched))), nil
}

// listWorkspaces resolves the workspaces a list reads: the one named by dir or
// by project short name, or every registered project when the request names
// neither.
func (d Deps) listWorkspaces(ctx context.Context, opts query.ListOpts) ([]query.Workspace, error) {
	dir, project := strings.TrimSpace(opts.Dir), strings.TrimSpace(opts.Project)
	if dir != "" && project != "" {
		return nil, rejected(fmt.Errorf("name either dir or project, not both (got dir %q and project %q)", dir, project))
	}
	if dir != "" {
		return []query.Workspace{{Dir: dir}}, nil
	}
	registered, err := d.Workspaces(ctx)
	if err != nil {
		return nil, err
	}
	// Indexed even when no project is named: rows carry the short name, and two
	// projects sharing one would make those rows ambiguous.
	byShort, err := query.IndexByShortName(registered)
	if err != nil {
		return nil, err
	}
	if project == "" {
		return registered, nil
	}
	if ws, ok := byShort[project]; ok {
		return []query.Workspace{ws}, nil
	}
	shorts := make([]string, 0, len(registered))
	for _, ws := range registered {
		shorts = append(shorts, ws.Short())
	}
	return nil, rejected(fmt.Errorf("unknown project %q; registered projects: %s", project, strings.Join(shorts, ", ")))
}

func (d Deps) get(ctx context.Context, ref string) (*types.TODO, error) {
	_, todo, err := d.lookup(ctx, ref)
	return todo, err
}

// lookup resolves one reference the way a cross-workspace selection requires:
// globally first, then re-read through the provider for the workspace that
// actually owns it, so every subsequent mutation and run-lifecycle write lands
// in the right database with a matching optimistic version.
func (d Deps) lookup(ctx context.Context, ref string) (todos.Provider, *types.TODO, error) {
	global, err := d.OpenGlobal(ctx)
	if err != nil {
		return nil, nil, err
	}
	todo, err := global.GetGlobal(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	owner := strings.TrimSpace(todo.CWD)
	if owner == "" {
		return nil, nil, fmt.Errorf("resolved TODO %q has no owning workspace path", ref)
	}
	provider, err := d.OpenProvider(ctx, owner)
	if err != nil {
		return nil, nil, err
	}
	todo, err = provider.Get(ctx, todo.ID)
	if err != nil {
		return nil, nil, err
	}
	return provider, todo, nil
}

// workspacesLister adapts a set of workspaces to the sliver query.Select needs,
// so the selector stays unit-testable without a database.
type workspacesLister struct {
	ctx        context.Context
	workspaces []query.Workspace
	open       query.OpenProvider
}

func (w workspacesLister) List(filters todos.DiscoveryFilters) (types.TODOS, error) {
	return query.ListWorkspaces(w.ctx, w.workspaces, w.open, filters)
}

// bulkActions is the registry. Adding an action here is the only step needed to
// make it executable from the CLI, the API and any front end reading the
// catalog — which is the property the whole change exists to get.
func (d Deps) bulkActions() []clicky.EntityBulkAction {
	destructive, additive := true, false
	runHints := entity.MCPToolHints{Icon: "play", Group: "Run", DefaultPermission: entity.ToolPermissionAsk, DestructiveHint: &destructive}
	// plan investigates read-only and only adds a plan to the TODO. MCP reads an
	// unset hint on a writing tool as destructive, so it says so explicitly.
	planHints := runHints
	planHints.DestructiveHint = &additive

	actions := []clicky.EntityBulkAction{
		action(d, "status", "Set the status of many TODOs",
			entity.MCPToolHints{Icon: "check-circle", Group: "Status"},
			bulk.StatusFlags{}, func(_ context.Context, flags bulk.StatusFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.SetStatus(flags)
			}),

		action(d, "priority", "Set the severity of many TODOs",
			entity.MCPToolHints{Icon: "flag", Group: "Status"},
			bulk.PriorityFlags{}, func(_ context.Context, flags bulk.PriorityFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.SetPriority(flags)
			}),

		action(d, "labels", "Add or remove labels across many TODOs",
			entity.MCPToolHints{Icon: "tag", Group: "Labels"},
			bulk.LabelFlags{}, func(_ context.Context, flags bulk.LabelFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.EditLabels(flags)
			}),

		action(d, "comment", "Append a comment to many TODOs",
			entity.MCPToolHints{Icon: "message", Group: "Status"},
			bulk.CommentFlags{}, func(_ context.Context, flags bulk.CommentFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.AddComment(flags)
			}),

		action(d, "delete", "Delete many TODOs",
			entity.MCPToolHints{
				Icon: "trash", Group: "Danger",
				DestructiveHint:   &destructive,
				DefaultPermission: entity.ToolPermissionAsk,
			},
			bulk.DeleteFlags{}, func(_ context.Context, flags bulk.DeleteFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.Delete(flags)
			}),

		// N TODOs become one, so it is an aggregate rather than a per-item
		// operation, and it retires the ones it folds in — an agent asks first.
		aggregate(d, "merge", "Combine many TODOs into one with AI",
			entity.MCPToolHints{
				Icon: "merge", Group: "Danger",
				DestructiveHint:   &destructive,
				DefaultPermission: entity.ToolPermissionAsk,
			},
			bulk.MergeFlags{}, d.runMerge),

		// It publishes outside gavel, so an agent asks before running it.
		action(d, "push", "Push many TODOs to GitHub issues",
			entity.MCPToolHints{Icon: "github", Group: "GitHub", DefaultPermission: entity.ToolPermissionAsk},
			bulk.PushFlags{}, func(_ context.Context, flags bulk.PushFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.Push(flags, d.pushBaseURL())
			}),

		// The verify step, which is the only thing that decides whether a TODO
		// is done. A TODO that fails is a per-item failure, never a rejection of
		// the batch — see bulk.Check.
		action(d, "check", "Run many TODOs' definition of done",
			entity.MCPToolHints{Icon: "shield-check", Group: "Run", DefaultPermission: entity.ToolPermissionAsk},
			bulk.CheckFlags{}, func(ctx context.Context, flags bulk.CheckFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.Check(flags, d.hostKind(ctx))
			}),

		action(d, "reopen", "Return many finished TODOs to pending",
			entity.MCPToolHints{Icon: "rotate-ccw", Group: "Status"},
			bulk.ReopenFlags{}, func(_ context.Context, flags bulk.ReopenFlags, _ []string) (bulk.ItemFunc, error) {
				return bulk.Reopen(flags)
			}),
	}

	// Export addresses rows by workspace rather than through a provider, so it
	// is registered only when the deps that reach the database directly are
	// wired — the same gate the portable import is behind.
	if d.portableReady() {
		actions = append(actions,
			aggregate(d, "export", "Export many TODOs as portable .todos Markdown",
				entity.MCPToolHints{Icon: "upload", Group: "Portable"},
				ExportFlags{}, d.runExport))
	}

	// run, plan and triage are the same action with a different prompt name: a
	// prompt declares its own behaviour class, so nothing here distinguishes
	// them beyond the name the catalog resolves and whether it can destroy.
	for _, prompt := range []struct {
		name, short string
		hints       entity.MCPToolHints
	}{
		{"run", "Implement many TODOs", runHints},
		{"plan", "Plan many TODOs", planHints},
		{"triage", "Triage many TODOs", runHints},
	} {
		name := prompt.name
		actions = append(actions, action(d, name, prompt.short, prompt.hints, bulk.RunFlags{},
			func(ctx context.Context, flags bulk.RunFlags, batch []string) (bulk.ItemFunc, error) {
				dir, err := d.dir(ctx, query.ListOpts{})
				if err != nil {
					return nil, err
				}
				resolve, broker := d.runtime(ctx, dir)
				return bulk.StartRun(bulk.RunSpec{
					Step: name, Flags: flags, Batch: batch, Registry: d.Registry,
					Dir: dir, Resolve: resolve, Broker: broker,
				})
			}))
	}
	return actions
}

// aggregate binds an operation that acts on the selection as a whole.
//
// Every other bulk action is a loop: one function applied to each resolved
// TODO. Merge is not — N TODOs become one, and which one survives depends on
// the whole set — so it resolves the same references and then runs once. The
// error convention is identical: a rejection before the first write returns an
// error, and anything after it lives in the Result.
func aggregate[F entity.ActionFlags](
	d Deps,
	name, short string,
	hints entity.MCPToolHints,
	flags F,
	run func(ctx context.Context, targets []bulk.Target, flags F) (bulk.Result, error),
) clicky.EntityBulkAction {
	return clicky.BulkActionWithContext(name, func(ctx context.Context, ids []string, raw map[string]string) (bulk.Result, error) {
		decoded, err := clicky.BuildOpts[F](raw)
		if err != nil {
			return bulk.Result{}, rejected(err)
		}
		targets, unresolved, err := bulk.Resolve(ctx, d.lookup, ids)
		if err != nil {
			return bulk.Result{}, rejected(err)
		}
		// An aggregate cannot proceed around a ref that named nothing: the
		// selection it merges would not be the selection the caller chose.
		if len(unresolved) > 0 {
			refs := make([]string, 0, len(unresolved))
			for _, missing := range unresolved {
				refs = append(refs, fmt.Sprintf("%s (%s)", missing.Ref, missing.Error))
			}
			return bulk.Result{}, rejected(fmt.Errorf("%s: could not resolve %s", name, strings.Join(refs, ", ")))
		}
		result, err := run(ctx, targets, decoded)
		if err != nil {
			return bulk.Result{}, rejected(err)
		}
		return result, nil
	}).
		WithShort(short).
		WithFlags(flags).
		WithToolHints(hints)
}

// runMerge folds the selection into one TODO. The prompt configuration is the
// workspace's: every merged TODO shares one (merge.Targets refuses otherwise),
// so it is read from the survivor's directory.
func (d Deps) runMerge(ctx context.Context, targets []bulk.Target, flags bulk.MergeFlags) (bulk.Result, error) {
	if len(targets) == 0 {
		return bulk.Result{}, fmt.Errorf("merge needs at least two TODOs; got none")
	}
	dir := strings.TrimSpace(targets[0].Todo.CWD)
	if dir == "" {
		var err error
		if dir, err = d.dir(ctx, query.ListOpts{}); err != nil {
			return bulk.Result{}, err
		}
	}
	opts, err := merge.NewOptions(dir, flags)
	if err != nil {
		return bulk.Result{}, err
	}
	result, _, err := merge.Run(ctx, targets, opts)
	return result, err
}

// rejected marks an error as the caller's mistake rather than a server fault,
// so a malformed request answers 400 instead of 500. It is applied only where
// the failure is a bad request by construction — decoding flags, validating
// them, validating the selection — never to a provider or transport failure,
// which really is the server's problem and really should be retried.
func rejected(err error) error {
	if err == nil {
		return nil
	}
	var already *entity.StatusError
	if errors.As(err, &already) {
		return err
	}
	return entity.NewStatusError(http.StatusBadRequest, "invalid_request", err.Error())
}

// action binds a typed operation to explicit TODO references.
func action[F entity.ActionFlags](
	d Deps,
	name, short string,
	hints entity.MCPToolHints,
	flags F,
	build func(ctx context.Context, flags F, batch []string) (bulk.ItemFunc, error),
) clicky.EntityBulkAction {
	return clicky.BulkActionWithContext(name, func(ctx context.Context, ids []string, raw map[string]string) (bulk.Result, error) {
		return apply(ctx, d, name, ids, raw, build)
	}).
		WithShort(short).
		WithFlags(flags).
		WithToolHints(hints)
}

// apply decodes action flags, resolves explicit refs, and applies the operation.
//
// Note the error convention. Everything that can be decided before the first
// write — bad flags, a malformed selection — returns an error and the whole
// request is rejected. Once the loop starts, failures live inside the Result
// and the error stays nil, because clicky discards the result value whenever
// the error is non-nil and that would throw away every item that succeeded.
func apply[F entity.ActionFlags](
	ctx context.Context,
	d Deps,
	name string,
	ids []string,
	raw map[string]string,
	build func(ctx context.Context, flags F, batch []string) (bulk.ItemFunc, error),
) (bulk.Result, error) {
	flags, err := clicky.BuildOpts[F](raw)
	if err != nil {
		return bulk.Result{}, rejected(err)
	}
	// ids is the selection as the caller spelled it, which is exactly what a
	// run-shaped action passes on as the batch: every other action ignores it.
	fn, err := build(ctx, flags, ids)
	if err != nil {
		return bulk.Result{}, rejected(err)
	}
	targets, unresolved, err := bulk.Resolve(ctx, d.lookup, ids)
	if err != nil {
		return bulk.Result{}, rejected(err)
	}
	result := bulk.Apply(ctx, name, targets, fn)
	// A ref that named nothing is a per-item failure, not a rejection: one stale
	// id in a selection of forty is ordinary when a tab has been open a while.
	for _, missing := range unresolved {
		result.Failed++
		result.Results = append(result.Results, missing)
	}
	return result, nil
}
