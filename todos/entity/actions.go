package entity

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/content"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/ops"
	"github.com/flanksource/gavel/todos/portable"
	"github.com/flanksource/gavel/todos/query"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/todosync"
	"gorm.io/gorm"
)

// itemActions are the operations that address one TODO, or none at all.
//
// They live apart from the bulk actions because they are genuinely not loops:
// creating a TODO has no selection to iterate, and linking two of them is one
// operation on a pair. Registering them here is what makes the entity the whole
// TODO vocabulary rather than most of it — before this an agent could set a
// status on forty TODOs but could not write one, because `create` existed only
// as a Cobra command in package main, which nothing else can import.
//
// Some are conditional on the dep they need being wired. That is deliberate: a
// registration that cannot resolve a project directory should not advertise a
// `transfer` tool that is certain to fail.
func (d Deps) itemActions() []clicky.EntityAction {
	destructive := true
	writes := entity.MCPToolHints{Icon: "edit", Group: "Content"}
	linkHints := entity.MCPToolHints{Icon: "link", Group: "Links"}
	reads := entity.MCPToolHints{Icon: "list", Group: "Read"}

	actions := []clicky.EntityAction{
		// No id: the TODO does not exist yet. WithOptionalID is what lets one
		// declaration be a collection operation over HTTP and a bare `create`
		// everywhere else.
		clicky.TypedActionWithContext("create", CreateFlags{}, d.create).
			WithShort("Create a TODO").
			WithToolHints(entity.MCPToolHints{Icon: "plus", Group: "Content"}).
			WithOptionalID(),

		clicky.TypedActionWithContext("edit", EditFlags{}, d.edit).
			WithShort("Edit a TODO's content, plan, status, and/or priority").
			WithToolHints(writes),

		clicky.TypedActionWithContext("link", LinkFlags{}, d.link).
			WithShort("Link a TODO to another as related, blocking, or duplicate").
			WithToolHints(linkHints),

		clicky.TypedActionWithContext("unlink", LinkFlags{}, d.unlink).
			WithShort("Remove a link between two TODOs").
			WithToolHints(entity.MCPToolHints{
				Icon: "unlink", Group: "Links", DestructiveHint: &destructive,
			}),

		clicky.TypedActionWithContext("links", NoFlags{}, d.links).
			WithShort("List a TODO's links").
			WithToolHints(reads),

		// Optional id: with one it answers "what can I do with this TODO now",
		// without one it describes the workspace's lifecycle itself.
		clicky.TypedActionWithContext("steps", StepsFlags{}, d.steps).
			WithShort("List the lifecycle's steps, or where one TODO stands in it").
			WithToolHints(reads).
			WithOptionalID(),

		clicky.TypedActionWithContext("sync", SyncFlags{}, d.sync).
			WithShort("Sync source TODO/FIXME comments into TODO issues").
			WithToolHints(entity.MCPToolHints{Icon: "refresh", Group: "Content"}).
			WithOptionalID(),
	}

	if d.ProjectDir != nil {
		actions = append(actions,
			clicky.TypedActionWithContext("transfer", TransferFlags{}, d.transfer).
				WithShort("Move a TODO to another registered project").
				WithToolHints(entity.MCPToolHints{
					Icon: "arrow-right", Group: "Content",
					DefaultPermission: entity.ToolPermissionAsk,
				}))
	}
	if d.portableReady() {
		actions = append(actions,
			clicky.TypedActionWithContext("import", ImportFlags{}, d.importPortable).
				WithShort("Import .todos Markdown into native PostgreSQL TODOs").
				WithToolHints(entity.MCPToolHints{
					Icon: "download", Group: "Portable",
					DefaultPermission: entity.ToolPermissionAsk,
				}).
				WithOptionalID())
	}
	return actions
}

// NoFlags is the empty parameter set, for an action whose only input is the id.
type NoFlags struct{}

func (NoFlags) ClickyActionFlags() {}

// CreateFlags are the parameters of `create`.
//
// There is no "was it set" companion for the content fields, which the CLI gets
// from cobra's Changed(). Over HTTP and MCP no such signal exists — a JSON body
// either carries a field or does not — so an empty string means "not supplied"
// here, and an explicitly empty plan is caught by ops.ValidateCreatePlan rather
// than by flag bookkeeping.
type CreateFlags struct {
	Title        string   `flag:"title" help:"TODO title" required:"true"`
	Body         string   `flag:"body" help:"TODO body or @path"`
	Plan         string   `flag:"plan" help:"Reviewed implementation plan or @path"`
	Verification string   `flag:"verification" help:"Verification fixture markdown or @path"`
	Labels       []string `flag:"label" help:"Attach a label"`
	Priority     string   `flag:"priority" default:"medium" help:"TODO priority" enum:"high,medium,low"`
	Status       string   `flag:"status" default:"pending" help:"Initial status, or 'approved' to bless the supplied plan"`
	Dir          string   `flag:"dir" help:"Workspace directory; defaults to the current one"`
	GitHub       bool     `flag:"github" help:"Push the new TODO to a GitHub issue"`
	BaseURL      string   `flag:"base-url" help:"Absolute origin attachment links resolve against"`
	Repo         string   `flag:"repo" help:"Target owner/repo"`
}

func (CreateFlags) ClickyActionFlags() {}

func (d Deps) create(ctx context.Context, _ string, flags CreateFlags) (*types.TODO, error) {
	title := strings.TrimSpace(flags.Title)
	if title == "" {
		return nil, rejected(fmt.Errorf("title is required"))
	}
	dir, err := d.dir(ctx, query.ListOpts{Dir: flags.Dir})
	if err != nil {
		return nil, rejected(err)
	}
	resolved, err := content.ResolveCreate(dir, content.CreateOptions{
		Body: flags.Body, BodySet: flags.Body != "",
		Plan: flags.Plan, PlanSet: flags.Plan != "",
		Verification: flags.Verification, VerificationSet: flags.Verification != "",
	})
	if err != nil {
		return nil, rejected(err)
	}
	priority, err := ops.ParsePriority(flags.Priority)
	if err != nil {
		return nil, rejected(err)
	}
	start, err := ops.ParseLifecycle(flags.Status)
	if err != nil {
		return nil, rejected(err)
	}
	if err := ops.ValidateCreatePlan(resolved.Plan, start); err != nil {
		return nil, rejected(err)
	}
	provider, err := d.OpenProvider(ctx, dir)
	if err != nil {
		return nil, err
	}
	request := todos.CreateRequest{
		Title: title, Body: resolved.Body, Verification: resolved.Verification,
		Priority: priority, Status: start.Status, Labels: flags.Labels,
	}
	if resolved.Plan != "" {
		request.Plan = &todos.CreatePlanRequest{Markdown: resolved.Plan, Approved: start.PlanApproved}
	}
	todo, err := provider.Create(ctx, request)
	if err != nil {
		return nil, err
	}
	if !flags.GitHub {
		return todo, nil
	}
	baseURL, err := d.pushBaseURL()(dir, flags.BaseURL)
	if err != nil {
		return todo, err
	}
	if _, err := githubpush.Push(ctx, provider, todo.ID, githubpush.Options{
		GitHub:  github.Options{WorkDir: dir, Repo: flags.Repo},
		BaseURL: baseURL,
		Labels:  true,
	}); err != nil {
		// The TODO exists either way, so a push failure names both facts rather
		// than reading as "nothing happened".
		return todo, fmt.Errorf("created %s but could not push it to GitHub: %w", todo.DisplayID(), err)
	}
	return provider.Get(ctx, todo.ID)
}

// EditFlags are the parameters of `edit`. Like CreateFlags it reads emptiness as
// "not supplied"; --clear-labels is how an empty label set is asked for.
type EditFlags struct {
	Title        string   `flag:"title" help:"New title"`
	Body         string   `flag:"body" help:"New body, path, or @path"`
	Plan         string   `flag:"plan" help:"Set or replace the TODO plan"`
	Verification string   `flag:"verification" help:"New verification fixture, path, or @path"`
	Status       string   `flag:"status" help:"New status"`
	Priority     string   `flag:"priority" help:"New priority" enum:"high,medium,low"`
	Labels       []string `flag:"label" help:"Replace the TODO labels"`
	ClearLabels  bool     `flag:"clear-labels" help:"Remove every label from the TODO"`
}

func (EditFlags) ClickyActionFlags() {}

func (d Deps) edit(ctx context.Context, ref string, flags EditFlags) (*types.TODO, error) {
	provider, todo, err := d.lookup(ctx, ref)
	if err != nil {
		return nil, err
	}
	dir, err := d.todoDir(ctx, todo, "")
	if err != nil {
		return nil, rejected(err)
	}
	edits := ops.EditFlags{Status: flags.Status, Priority: flags.Priority, ClearLabels: flags.ClearLabels}
	if flags.Title != "" {
		edits.Title = &flags.Title
	}
	for _, field := range []struct {
		flag, value string
		into        **string
	}{
		{"--body", flags.Body, &edits.Body},
		{"--plan", flags.Plan, &edits.Plan},
		{"--verification", flags.Verification, &edits.Verification},
	} {
		if field.value == "" {
			continue
		}
		text, err := content.Resolve(content.Options{WorkDir: dir, Flag: field.flag, Value: field.value})
		if err != nil {
			return nil, rejected(err)
		}
		*field.into = &text
	}
	if flags.Labels != nil {
		edits.Labels = &flags.Labels
	}
	changes, err := ops.BuildEdit(edits)
	if err != nil {
		return nil, rejected(err)
	}
	if _, err := ops.ApplyEdit(ctx, provider, todo, changes); err != nil {
		return nil, err
	}
	return provider.Get(ctx, todo.ID)
}

// todoDir is the workspace a TODO's relative paths resolve against: the one that
// owns it, not the caller's. A dashboard or chat selection routinely spans
// repositories, so resolving `@plan.md` against the server's directory would
// read the wrong repository's file — or, worse, the right filename in it.
func (d Deps) todoDir(ctx context.Context, todo *types.TODO, fallback string) (string, error) {
	if dir := strings.TrimSpace(todo.CWD); dir != "" {
		return dir, nil
	}
	return d.dir(ctx, query.ListOpts{Dir: fallback})
}

// LinkFlags are the parameters of `link` and `unlink`.
type LinkFlags struct {
	To       string `flag:"to" help:"The other TODO's ID or alias" required:"true"`
	Relation string `flag:"relation" default:"related-to" help:"Relation to create or remove"`
}

func (LinkFlags) ClickyActionFlags() {}

// LinkResult is what all three link operations answer with: the TODO's links as
// they now stand. Returning the whole set rather than the one edge that changed
// is what lets a chat window render the outcome without a follow-up read.
type LinkResult struct {
	Todo  string       `json:"todo"`
	Links []todos.Link `json:"links"`
}

func (d Deps) link(ctx context.Context, ref string, flags LinkFlags) (LinkResult, error) {
	linker, todo, relation, err := d.linkTarget(ctx, ref, flags)
	if err != nil {
		return LinkResult{}, err
	}
	if _, err := linker.Link(ctx, todo, flags.To, relation); err != nil {
		return LinkResult{}, err
	}
	return linkResult(ctx, linker, todo)
}

func (d Deps) unlink(ctx context.Context, ref string, flags LinkFlags) (LinkResult, error) {
	linker, todo, relation, err := d.linkTarget(ctx, ref, flags)
	if err != nil {
		return LinkResult{}, err
	}
	if err := linker.Unlink(ctx, todo, flags.To, relation); err != nil {
		return LinkResult{}, err
	}
	return linkResult(ctx, linker, todo)
}

func (d Deps) links(ctx context.Context, ref string, _ NoFlags) (LinkResult, error) {
	linker, todo, err := d.linker(ctx, ref)
	if err != nil {
		return LinkResult{}, err
	}
	return linkResult(ctx, linker, todo)
}

func (d Deps) linkTarget(ctx context.Context, ref string, flags LinkFlags) (todos.RelationshipProvider, *types.TODO, types.RelationKind, error) {
	relation, err := types.ParseRelationKind(flags.Relation)
	if err != nil {
		return nil, nil, "", rejected(err)
	}
	if strings.TrimSpace(flags.To) == "" {
		return nil, nil, "", rejected(fmt.Errorf("--to is required"))
	}
	linker, todo, err := d.linker(ctx, ref)
	if err != nil {
		return nil, nil, "", err
	}
	return linker, todo, relation, nil
}

// linker resolves a reference and asserts that its provider stores relations.
func (d Deps) linker(ctx context.Context, ref string) (todos.RelationshipProvider, *types.TODO, error) {
	provider, todo, err := d.lookup(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	linker, ok := provider.(todos.RelationshipProvider)
	if !ok {
		return nil, nil, fmt.Errorf("TODO provider does not support links; native PostgreSQL storage is required")
	}
	return linker, todo, nil
}

func linkResult(ctx context.Context, linker todos.RelationshipProvider, todo *types.TODO) (LinkResult, error) {
	links, err := linker.Links(ctx, todo)
	if err != nil {
		return LinkResult{}, err
	}
	return LinkResult{Todo: todo.DisplayID(), Links: links}, nil
}

// StepsFlags are the parameters of `steps`.
type StepsFlags struct {
	Dir string `flag:"dir" help:"Workspace directory; defaults to the current one"`
}

func (StepsFlags) ClickyActionFlags() {}

// StepsResult answers one of two questions depending on whether an id was
// given. They share a type because a typed action has exactly one result shape:
// Steps is populated for a TODO, Lifecycle for the workspace.
type StepsResult struct {
	Todo      string                `json:"todo,omitempty"`
	Lifecycle *lifecycle.Lifecycle  `json:"lifecycle,omitempty"`
	Steps     []lifecycle.StepState `json:"steps,omitempty"`
}

func (d Deps) steps(ctx context.Context, ref string, flags StepsFlags) (StepsResult, error) {
	if strings.TrimSpace(ref) == "" {
		dir, err := d.dir(ctx, query.ListOpts{Dir: flags.Dir})
		if err != nil {
			return StepsResult{}, rejected(err)
		}
		host, err := lifecycle.NewHost(nil, dir, d.hostKind(ctx))
		if err != nil {
			return StepsResult{}, err
		}
		def := host.Def.Definition()
		return StepsResult{Lifecycle: &def}, nil
	}
	provider, todo, err := d.lookup(ctx, ref)
	if err != nil {
		return StepsResult{}, err
	}
	dir, err := d.todoDir(ctx, todo, flags.Dir)
	if err != nil {
		return StepsResult{}, rejected(err)
	}
	host, err := lifecycle.NewHost(provider, dir, d.hostKind(ctx))
	if err != nil {
		return StepsResult{}, err
	}
	states, err := host.Steps(ctx, todo)
	if err != nil {
		return StepsResult{}, err
	}
	return StepsResult{Todo: todo.DisplayID(), Steps: states}, nil
}

// hostKind picks the lifecycle host for this invocation's surface, for the same
// reason runtime picks the broker: only an attended caller serves the approval
// endpoints the dashboard host's layers assume exist.
func (d Deps) hostKind(ctx context.Context) lifecycle.HostKind {
	if attended(ctx) {
		return lifecycle.HostDashboard
	}
	return lifecycle.HostCLI
}

// TransferFlags are the parameters of `transfer`.
type TransferFlags struct {
	To string `flag:"to" help:"Destination project name" required:"true"`
}

func (TransferFlags) ClickyActionFlags() {}

func (d Deps) transfer(ctx context.Context, ref string, flags TransferFlags) (*types.TODO, error) {
	if d.ProjectDir == nil {
		return nil, fmt.Errorf("entity deps: ProjectDir is required to transfer TODOs")
	}
	source, todo, err := d.lookup(ctx, ref)
	if err != nil {
		return nil, err
	}
	targetDir, err := d.ProjectDir(ctx, flags.To)
	if err != nil {
		return nil, rejected(err)
	}
	if strings.TrimSpace(todo.CWD) == targetDir {
		return nil, rejected(fmt.Errorf("%s already belongs to project %q", todo.DisplayID(), flags.To))
	}
	target, err := d.OpenProvider(ctx, targetDir)
	if err != nil {
		return nil, err
	}
	return todos.Transfer(ctx, source, target, todo.ID)
}

// SyncFlags are the parameters of `sync`.
type SyncFlags struct {
	Paths   []string `flag:"path" help:"Limit the scan to these paths"`
	Markers []string `flag:"markers" default:"TODO,FIXME" help:"Source comment markers to sync"`
	Ignore  []string `flag:"ignore" help:"Additional path glob to ignore during the source scan"`
	DryRun  bool     `flag:"dry-run" help:"Report the planned sync without updating TODOs"`
	Dir     string   `flag:"dir" help:"Workspace directory; defaults to the current one"`
}

func (SyncFlags) ClickyActionFlags() {}

func (d Deps) sync(ctx context.Context, _ string, flags SyncFlags) (*todosync.SourceCommentSyncResult, error) {
	dir, err := d.dir(ctx, query.ListOpts{Dir: flags.Dir})
	if err != nil {
		return nil, rejected(err)
	}
	provider, err := d.OpenProvider(ctx, dir)
	if err != nil {
		return nil, err
	}
	return todosync.SyncSourceComments(ctx, provider, todosync.SourceCommentSyncOptions{
		WorkDir: dir,
		Paths:   flags.Paths,
		Markers: flags.Markers,
		Ignore:  flags.Ignore,
		DryRun:  flags.DryRun,
	})
}

// ImportFlags are the parameters of `import`.
type ImportFlags struct {
	Files []string `flag:"file" help:"Markdown files to import; empty reads the whole directory"`
	Dir   string   `flag:"dir" default:".todos" help:"Directory to read when no files are supplied"`
	// Workspace is the repository whose TODOs are being written, which is not
	// the directory the Markdown is read from: the two differ whenever a caller
	// imports into a project other than the one it is standing in.
	Workspace string `flag:"workspace" help:"Workspace directory; defaults to the current one"`
}

func (ImportFlags) ClickyActionFlags() {}

func (d Deps) importPortable(ctx context.Context, _ string, flags ImportFlags) (*portable.ImportResult, error) {
	options, db, err := d.portableFor(ctx, flags.Workspace)
	if err != nil {
		return nil, err
	}
	return portable.Import(ctx, db, options, flags.Dir, flags.Files)
}

func (d Deps) portableReady() bool { return d.Workspace != nil && d.DB != nil }

// portableFor resolves the workspace and database the portable import/export
// address rows by. Both deps are re-checked even though registration gates on
// them, because Deps can be constructed directly.
func (d Deps) portableFor(ctx context.Context, dir string) (todoruntime.WorkspaceOptions, *gorm.DB, error) {
	if !d.portableReady() {
		return todoruntime.WorkspaceOptions{}, nil, fmt.Errorf("entity deps: Workspace and DB are required for portable import/export")
	}
	resolved, err := d.dir(ctx, query.ListOpts{Dir: dir})
	if err != nil {
		return todoruntime.WorkspaceOptions{}, nil, rejected(err)
	}
	options, err := d.Workspace(ctx, resolved)
	if err != nil {
		return todoruntime.WorkspaceOptions{}, nil, err
	}
	db, err := d.DB(ctx)
	if err != nil {
		return todoruntime.WorkspaceOptions{}, nil, err
	}
	return options, db, nil
}

// ExportFlags are the parameters of the `export` bulk action.
type ExportFlags struct {
	Dir   string `flag:"dir" default:".todos" help:"Directory the Markdown is written to"`
	Force bool   `flag:"force" help:"Overwrite files that already exist"`
}

func (ExportFlags) ClickyActionFlags() {}

// runExport writes the selection out as portable Markdown.
//
// It is an aggregate rather than a per-item loop because the export addresses
// rows by workspace, and a selection spans workspaces: exporting one TODO at a
// time would open the same database once per row and have each call contend for
// the same directory. So the targets are grouped by owning workspace and each
// group exported once, with the result reported per TODO so a partial success
// still names what landed.
func (d Deps) runExport(ctx context.Context, targets []bulk.Target, flags ExportFlags) (bulk.Result, error) {
	if !d.portableReady() {
		return bulk.Result{}, fmt.Errorf("entity deps: Workspace and DB are required for portable export")
	}
	byWorkspace := map[string][]bulk.Target{}
	for _, target := range targets {
		dir := strings.TrimSpace(target.Todo.CWD)
		if dir == "" {
			return bulk.Result{}, fmt.Errorf("%s has no owning workspace path to export from", target.Ref)
		}
		byWorkspace[dir] = append(byWorkspace[dir], target)
	}
	workspaces := make([]string, 0, len(byWorkspace))
	for dir := range byWorkspace {
		workspaces = append(workspaces, dir)
	}
	sort.Strings(workspaces)

	result := bulk.Result{Action: "export"}
	for _, dir := range workspaces {
		group := byWorkspace[dir]
		refs := make([]string, 0, len(group))
		for _, target := range group {
			refs = append(refs, target.Todo.ID)
		}
		// A failure is scoped to its workspace rather than fatal: one
		// unreadable repository must not discard what the others wrote.
		exported, err := d.exportWorkspace(ctx, dir, refs, flags)
		for _, target := range group {
			item := bulk.ItemResult{Ref: target.Ref, Dir: dir, Title: target.Todo.Title}
			if err != nil {
				item.Error = err.Error()
				result.Failed++
			} else {
				item.Status = exported.Directory
				result.Applied++
			}
			result.Results = append(result.Results, item)
		}
	}
	return result, nil
}

func (d Deps) exportWorkspace(ctx context.Context, dir string, refs []string, flags ExportFlags) (*portable.ExportResult, error) {
	options, db, err := d.portableFor(ctx, dir)
	if err != nil {
		return nil, err
	}
	return portable.Export(ctx, db, options, flags.Dir, refs, flags.Force)
}
