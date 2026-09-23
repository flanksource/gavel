package entity

import (
	"context"
	"testing"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	clickyentity "github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/query"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"gorm.io/gorm"
)

// declared registers the deps' actions under a throwaway entity name and reads
// the published catalog back. Going through registration rather than inspecting
// the specs directly is the point: what a front end, the OpenAPI document and an
// MCP client see is the registered form, not the builder's input.
func declared(t *testing.T, name string, deps Deps) clickyentity.EntityInfo {
	t.Helper()
	builder := clicky.NewEntity[Summary, query.ListOpts, *types.TODO](name).
		ListPagedWithContext(deps.list).
		GetWithContext(deps.get)
	for _, action := range deps.itemActions() {
		builder = builder.WithAction(action)
	}
	for _, action := range deps.bulkActions() {
		builder = builder.WithBulkAction(action)
	}
	builder.Register()

	for _, info := range clicky.GetEntities() {
		if info.Name == name {
			return info
		}
	}
	t.Fatalf("entity %q was not registered", name)
	return clickyentity.EntityInfo{}
}

func itemActionsByName(info clickyentity.EntityInfo) map[string]clickyentity.ActionInfo {
	byName := make(map[string]clickyentity.ActionInfo, len(info.Actions))
	for _, action := range info.Actions {
		byName[action.Name] = action
	}
	return byName
}

func portableDeps() Deps {
	deps := testDeps()
	deps.Workspace = func(context.Context, string) (todoruntime.WorkspaceOptions, error) {
		return todoruntime.WorkspaceOptions{}, nil
	}
	deps.DB = func(context.Context) (*gorm.DB, error) { return nil, nil }
	deps.ProjectDir = func(context.Context, string) (string, error) { return "/tmp/other", nil }
	return deps
}

// The gap this change exists to close: an agent could set a status on forty
// TODOs but could not write one, because create lived only in package main,
// which nothing else can import.
func TestItemActionsCoverTheWholeTodoVocabulary(t *testing.T) {
	actions := itemActionsByName(declared(t, "todo-vocabulary", portableDeps()))
	want := []string{"create", "edit", "link", "unlink", "links", "steps", "sync", "transfer", "import"}
	for _, name := range want {
		info, ok := actions[name]
		if !ok {
			t.Fatalf("item action %q was not declared", name)
		}
		if info.ContextDataFunc == nil && info.DataFunc == nil {
			t.Fatalf("%s: no handler", name)
		}
		// A front end and an MCP client both render from this; a bare name gives
		// them nothing to draw and nothing to describe.
		if info.Short == "" {
			t.Fatalf("%s: no description for the catalog", name)
		}
		if info.ToolHints.Icon == "" || info.ToolHints.Group == "" {
			t.Fatalf("%s: no icon/group", name)
		}
		if info.FlagsType == nil {
			t.Fatalf("%s: parameters were not published", name)
		}
	}
	if len(actions) != len(want) {
		t.Fatalf("declared %d item actions, want exactly %v", len(actions), want)
	}
}

// create, steps, sync and import address no existing TODO — create's does not
// exist yet and the other three take their target from flags. Without the
// optional id the generated command and route demand a positional nobody has.
func TestTargetlessActionsDoNotRequireAnID(t *testing.T) {
	actions := itemActionsByName(declared(t, "todo-optional-id", portableDeps()))
	for _, name := range []string{"create", "steps", "sync", "import"} {
		if !actions[name].OptionalID {
			t.Fatalf("%s must be invokable without an id", name)
		}
	}
	for _, name := range []string{"edit", "link", "unlink", "links", "transfer"} {
		if actions[name].OptionalID {
			t.Fatalf("%s acts on one existing TODO and must require its id", name)
		}
	}
}

// An action whose dep is missing would be a tool that is certain to fail, which
// is worse than an absent one: an agent reads the catalog as a list of things
// that work.
func TestOptionalActionsAreGatedOnTheirDependencies(t *testing.T) {
	bare := itemActionsByName(declared(t, "todo-bare-deps", testDeps()))
	for _, name := range []string{"transfer", "import"} {
		if _, ok := bare[name]; ok {
			t.Fatalf("%s must not be advertised without the dep it needs", name)
		}
	}

	// Half-wired portability stays absent: the portable import needs both a
	// workspace to address rows by and a database to write them to.
	half := testDeps()
	half.DB = func(context.Context) (*gorm.DB, error) { return nil, nil }
	if _, ok := itemActionsByName(declared(t, "todo-half-portable", half))["import"]; ok {
		t.Fatal("import must not be advertised with DB but no Workspace")
	}

	full := declared(t, "todo-full-deps", portableDeps())
	if _, ok := itemActionsByName(full)["import"]; !ok {
		t.Fatal("import must be declared once Workspace and DB are both wired")
	}
	exported := false
	for _, action := range full.BulkActions {
		if action.Name == "export" {
			exported = true
		}
	}
	if !exported {
		t.Fatal("export must be declared once Workspace and DB are both wired")
	}
}

// A run started from a terminal or a cron slot has nobody to answer its
// tool-permission prompt, so it must never be configured to ask — whatever the
// process-global registration was given. Registration cannot decide this,
// because one registration serves every surface.
func TestApprovalBrokerIsOnlyWiredOnAnAttendedSurface(t *testing.T) {
	deps := testDeps()
	deps.Broker = func(context.Context, string) todos.ApprovalBroker {
		return func(*todos.ExecutorContext) (captainapi.PermissionFunc, error) { return nil, nil }
	}

	for _, surface := range []string{"http", "mcp"} {
		ctx := clickyentity.ContextWithOperationSurface(context.Background(), surface)
		if _, broker := deps.runtime(ctx, "/tmp/workspace"); broker == nil {
			t.Fatalf("%s serves an approval endpoint and must get a broker", surface)
		}
	}
	// "" is a direct in-process call, which has no one watching either.
	for _, surface := range []string{"cli", "schedule", ""} {
		ctx := context.Background()
		if surface != "" {
			ctx = clickyentity.ContextWithOperationSurface(ctx, surface)
		}
		if _, broker := deps.runtime(ctx, "/tmp/workspace"); broker != nil {
			t.Fatalf("%q has nobody to answer a prompt and must get no broker", surface)
		}
	}
}

// The dashboard's run resolution applies a runtime catalog and resolves as the
// approval-serving host, neither of which is true of a terminal run.
func TestUnattendedCallersGetThePlainRunResolution(t *testing.T) {
	deps := testDeps()
	deps.ResolveRun = func(context.Context, bulk.RunRequest) (run.Options, error) {
		t.Fatal("the dashboard resolver must not be consulted from an unattended surface")
		return run.Options{}, nil
	}
	resolve, _ := deps.runtime(clickyentity.ContextWithOperationSurface(context.Background(), "cli"), "/tmp/workspace")
	if resolve == nil {
		t.Fatal("an unattended caller still needs a resolver")
	}
}

// The lifecycle host is the second surface-aware decision: HostDashboard's
// layers assume approval endpoints a terminal does not serve.
func TestHostKindFollowsTheSurface(t *testing.T) {
	deps := testDeps()
	for surface, want := range map[string]string{
		"http": "dashboard", "mcp": "dashboard", "cli": "cli", "schedule": "cli",
	} {
		ctx := clickyentity.ContextWithOperationSurface(context.Background(), surface)
		if got := string(deps.hostKind(ctx)); got != want {
			t.Fatalf("%s surface resolved host %q, want %q", surface, got, want)
		}
	}
}
