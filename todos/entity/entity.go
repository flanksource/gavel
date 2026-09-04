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
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
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
	// DefaultDir supplies the workspace when a request names none.
	DefaultDir func() string
	// ResolveRun turns a TODO plus the batch's overrides into run options.
	// Optional: bulk.DefaultRunResolver is used when unset. The dashboard
	// supplies its own because it applies a runtime catalog the CLI has no
	// equivalent of, and because it resolves as the approval-serving host.
	ResolveRun bulk.RunResolver
	// Broker answers a batched run's tool-permission requests. Optional, and
	// nil is the CLI's answer: a terminal batch has no one to ask, so a run it
	// starts must never be configured to.
	Broker func(dir string) todos.ApprovalBroker
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
	return nil
}

// broker is the approval callback factory for the batch's workspace, or nil
// when the host answers no approvals.
func (d Deps) broker() todos.ApprovalBroker {
	if d.Broker == nil {
		return nil
	}
	return d.Broker(d.dir(query.ListOpts{}))
}

func (d Deps) dir(opts query.ListOpts) string {
	if dir := strings.TrimSpace(opts.Dir); dir != "" {
		return dir
	}
	if d.DefaultDir != nil {
		return d.DefaultDir()
	}
	return ""
}

// Register declares the todos entity. Call it once at startup, before
// entity.GenerateCLI or the RPC server reads the registry.
func Register(deps Deps) error {
	if err := deps.validate(); err != nil {
		return err
	}
	builder := clicky.NewEntity[*types.TODO, query.ListOpts, *types.TODO]("todo").
		Aliases("todos").
		ListWithContext(deps.list).
		GetWithContext(deps.get)

	for _, action := range deps.bulkActions() {
		builder = builder.WithBulkAction(action)
	}
	builder.Register()
	return nil
}

func (d Deps) list(ctx context.Context, opts query.ListOpts) ([]*types.TODO, error) {
	provider, err := d.OpenProvider(ctx, d.dir(opts))
	if err != nil {
		return nil, err
	}
	return opts.Select(providerLister{ctx: ctx, provider: provider}, time.Now())
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

// providerLister adapts a Provider to the sliver query.Select needs, so the
// selector stays unit-testable without a database.
type providerLister struct {
	ctx      context.Context
	provider todos.Provider
}

func (p providerLister) List(filters todos.DiscoveryFilters) (types.TODOS, error) {
	return p.provider.List(p.ctx, filters)
}

// bulkActions is the registry. Adding an action here is the only step needed to
// make it executable from the CLI, the API and any front end reading the
// catalog — which is the property the whole change exists to get.
func (d Deps) bulkActions() []clicky.EntityBulkAction {
	destructive := true
	runHints := entity.MCPToolHints{Icon: "play", Group: "Run", DefaultPermission: entity.ToolPermissionAsk}

	actions := []clicky.EntityBulkAction{
		action(d, "status", "Set the status of many TODOs",
			entity.MCPToolHints{Icon: "check-circle", Group: "Status"},
			bulk.StatusFlags{}, bulk.SetStatus),

		action(d, "priority", "Set the severity of many TODOs",
			entity.MCPToolHints{Icon: "flag", Group: "Status"},
			bulk.PriorityFlags{}, bulk.SetPriority),

		action(d, "labels", "Add or remove labels across many TODOs",
			entity.MCPToolHints{Icon: "tag", Group: "Labels"},
			bulk.LabelFlags{}, bulk.EditLabels),

		action(d, "comment", "Append a comment to many TODOs",
			entity.MCPToolHints{Icon: "message", Group: "Status"},
			bulk.CommentFlags{}, bulk.AddComment),

		action(d, "delete", "Delete many TODOs",
			entity.MCPToolHints{
				Icon: "trash", Group: "Danger",
				DestructiveHint:   &destructive,
				DefaultPermission: entity.ToolPermissionAsk,
			},
			bulk.DeleteFlags{}, bulk.Delete),
	}

	// run, plan and triage are the same action with a different prompt name: a
	// prompt declares its own behaviour class, so nothing here distinguishes
	// them beyond the name the catalog resolves.
	for _, prompt := range []struct{ name, short string }{
		{"run", "Implement many TODOs"},
		{"plan", "Plan many TODOs"},
		{"triage", "Triage many TODOs"},
	} {
		name := prompt.name
		actions = append(actions, action(d, name, prompt.short, runHints, bulk.RunFlags{},
			func(flags bulk.RunFlags) (bulk.ItemFunc, error) {
				return bulk.StartRun(name, flags, d.Registry, d.dir(query.ListOpts{}), d.ResolveRun, d.broker())
			}))
	}
	return actions
}

// selector resolves which TODOs an invocation meant, in whichever mode it ran.
type selector func(ctx context.Context) ([]bulk.Target, []bulk.ItemResult, error)

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

// action wires one bulk action's two selector modes onto the same item
// function, so ids and filters cannot diverge in behaviour — only in how the
// TODOs were chosen.
//
// It is a free function rather than a method because Go has no generic methods
// and each action's flags are a different type.
func action[F entity.ActionFlags](
	d Deps,
	name, short string,
	hints entity.MCPToolHints,
	flags F,
	build func(F) (bulk.ItemFunc, error),
) clicky.EntityBulkAction {
	byIDs := func(ids []string, raw map[string]string) (bulk.Result, error) {
		return apply(d, name, query.ListOpts{}, raw, build,
			func(ctx context.Context) ([]bulk.Target, []bulk.ItemResult, error) {
				targets, unresolved, err := bulk.Resolve(ctx, d.lookup, ids)
				// Everything Resolve rejects is malformed input — a blank ref, a
				// duplicate that would apply twice. A ref that merely failed to
				// resolve comes back in unresolved, not as an error.
				return targets, unresolved, rejected(err)
			})
	}
	byFilter := func(opts query.ListOpts, raw map[string]string) (bulk.Result, error) {
		return apply(d, name, opts, raw, build,
			func(ctx context.Context) ([]bulk.Target, []bulk.ItemResult, error) {
				// A filter is answered by one workspace's provider, so every
				// TODO it matched is owned by that provider by construction.
				provider, err := d.OpenProvider(ctx, d.dir(opts))
				if err != nil {
					return nil, nil, err
				}
				selected, err := opts.Select(providerLister{ctx: ctx, provider: provider}, time.Now())
				if err != nil {
					return nil, nil, err
				}
				return bulk.TargetsFrom(provider, selected), nil, nil
			})
	}
	return clicky.BulkActionWithFilter(name, byIDs, byFilter).
		WithShort(short).
		WithFlags(flags).
		WithToolHints(hints)
}

// apply is the one path both selector modes run through: decode the action's
// own flags, build the per-item operation, resolve the selection, then apply.
//
// Note the error convention. Everything that can be decided before the first
// write — bad flags, a malformed selection — returns an error and the whole
// request is rejected. Once the loop starts, failures live inside the Result
// and the error stays nil, because clicky discards the result value whenever
// the error is non-nil and that would throw away every item that succeeded.
func apply[F entity.ActionFlags](
	d Deps,
	name string,
	opts query.ListOpts,
	raw map[string]string,
	build func(F) (bulk.ItemFunc, error),
	resolve selector,
) (bulk.Result, error) {
	flags, err := clicky.BuildOpts[F](raw)
	if err != nil {
		return bulk.Result{}, rejected(err)
	}
	fn, err := build(flags)
	if err != nil {
		return bulk.Result{}, rejected(err)
	}
	ctx := context.Background()
	targets, unresolved, err := resolve(ctx)
	if err != nil {
		return bulk.Result{}, err
	}
	result := bulk.Apply(ctx, name, targets, fn)
	result.MatchedBy = strings.TrimSpace(opts.Filter)
	// A ref that named nothing is a per-item failure, not a rejection: one stale
	// id in a selection of forty is ordinary when a tab has been open a while.
	for _, missing := range unresolved {
		result.Failed++
		result.Results = append(result.Results, missing)
	}
	return result, nil
}
