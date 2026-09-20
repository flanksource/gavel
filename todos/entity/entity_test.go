package entity

import (
	"context"
	"os"
	"testing"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func testDeps() Deps {
	return Deps{
		OpenProvider: func(context.Context, string) (todos.Provider, error) { return nil, nil },
		OpenGlobal:   func(context.Context) (todos.GlobalReferenceProvider, error) { return nil, nil },
		Registry:     run.NewRegistry(),
		DefaultDir:   func(context.Context) (string, error) { return "/tmp/workspace", nil },
	}
}

// The entity registry is a process global with no exported reset, so the
// registration happens once for the whole binary and the tests read it back.
func TestMain(m *testing.M) {
	if err := Register(testDeps()); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func registered(t *testing.T) entity.EntityInfo {
	t.Helper()
	for _, info := range clicky.GetEntities() {
		if info.Name == "todo" {
			return info
		}
	}
	t.Fatal("the todo entity was not registered")
	return entity.EntityInfo{}
}

func TestRegisterRequiresItsDependencies(t *testing.T) {
	full := testDeps()
	missing := map[string]Deps{
		"OpenProvider": {OpenGlobal: full.OpenGlobal, Registry: full.Registry},
		// Without it a cross-workspace selection cannot be resolved at all.
		"OpenGlobal": {OpenProvider: full.OpenProvider, Registry: full.Registry},
		// Two entrypoints sharing a process must share one in-flight run map, or
		// the same TODO could be started twice.
		"Registry": {OpenProvider: full.OpenProvider, OpenGlobal: full.OpenGlobal},
	}
	for name, deps := range missing {
		if err := Register(deps); err == nil {
			t.Fatalf("a missing %s must fail loudly at registration", name)
		}
	}
}

// The point of the entity: registering an action makes it executable from the
// CLI, the API and the published catalog at once.
func TestEveryBulkActionIsDeclaredAndRenderable(t *testing.T) {
	want := map[string]bool{
		"status": false, "priority": false, "labels": false, "comment": false,
		"delete": false, "run": false, "plan": false, "triage": false, "push": false,
		"merge": false,
	}
	for _, info := range registered(t).BulkActions {
		if _, expected := want[info.Name]; !expected {
			t.Fatalf("unexpected bulk action %q", info.Name)
		}
		want[info.Name] = true

		if info.ContextDataFunc == nil {
			t.Fatalf("%s: no id-mode handler", info.Name)
		}
		if info.FilterFunc != nil || info.ContextFilterFunc != nil {
			t.Fatalf("%s: mutation must not accept filtered selections", info.Name)
		}
		// A front end renders from this; a name alone gives it nothing to draw.
		if info.Short == "" {
			t.Fatalf("%s: no description for the catalog", info.Name)
		}
		if info.ToolHints.Icon == "" || info.ToolHints.Group == "" {
			t.Fatalf("%s: no icon/group for a selection toolbar", info.Name)
		}
		// Parameters have to survive registration, or the generated CLI and
		// OpenAPI describe an action taking nothing while the handler expects a
		// value.
		if info.FlagsType == nil {
			t.Fatalf("%s: parameters were not published", info.Name)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("bulk action %q was not declared", name)
		}
	}
}

// delete is the one action that can destroy a filter-matched selection the
// caller never enumerated, so it has to announce itself as such.
func TestDeleteIsMarkedDestructiveAndGated(t *testing.T) {
	for _, info := range registered(t).BulkActions {
		if info.Name != "delete" {
			continue
		}
		if info.ToolHints.DestructiveHint == nil || !*info.ToolHints.DestructiveHint {
			t.Fatal("delete must carry a destructive hint")
		}
		if info.ToolHints.DefaultPermission != entity.ToolPermissionAsk {
			t.Fatalf("delete must default to asking, got %q", info.ToolHints.DefaultPermission)
		}
		if info.FlagsType.Name() != "DeleteFlags" {
			t.Fatalf("delete flags type = %q", info.FlagsType.Name())
		}
		return
	}
	t.Fatal("delete action was not declared")
}

func TestStatusActionPublishesItsTypedFlags(t *testing.T) {
	for _, info := range registered(t).BulkActions {
		if info.Name == "status" {
			if got := info.FlagsType.Name(); got != "StatusFlags" {
				t.Fatalf("status flags type = %q", got)
			}
			return
		}
	}
	t.Fatal("status action was not declared")
}

// GetID/GetName are what let a TODO be an entity at all; if they regress, the
// generated routes address the wrong value.
func TestTodoSatisfiesEntityItem(t *testing.T) {
	todo := &types.TODO{ID: "6f1b0c2e"}
	todo.Title = "Ship it"

	var item entity.EntityItem = todo
	if item.GetID() != "6f1b0c2e" {
		t.Fatalf("GetID = %q, want the canonical id", item.GetID())
	}
	if item.GetName() != "Ship it" {
		t.Fatalf("GetName = %q", item.GetName())
	}
}

// The whole promise: one declaration, and the CLI has the commands — with the
// action's own parameters as real flags rather than an undeclared key the
// handler hopes for.
func TestRegistrationGeneratesCLICommands(t *testing.T) {
	root := &cobra.Command{Use: "gavel"}
	clicky.GenerateCLI(root)

	for _, name := range []string{"status", "priority", "labels", "run", "plan", "triage", "delete", "push", "merge"} {
		cmd, _, err := root.Find([]string{"todo", name})
		if err != nil || cmd == nil || cmd.Name() != name {
			t.Fatalf("expected `gavel todo %s` to be generated, got err=%v", name, err)
		}
	}

	status, _, err := root.Find([]string{"todo", "status"})
	if err != nil {
		t.Fatalf("find status: %v", err)
	}
	if status.Flags().Lookup("to") == nil {
		t.Fatal("`gavel todo status` must expose --to")
	}
	for _, flag := range []string{"filter", "search", "priority", "status"} {
		if status.Flags().Lookup(flag) != nil {
			t.Fatalf("bulk status must not expose list selector --%s", flag)
		}
	}

	push, _, err := root.Find([]string{"todo", "push"})
	if err != nil {
		t.Fatalf("find push: %v", err)
	}
	for _, flag := range []string{"update", "base-url"} {
		if push.Flags().Lookup(flag) == nil {
			t.Fatalf("`gavel todo push` must expose --%s", flag)
		}
	}
}

// merge is the one aggregate action: it folds the selection into a single TODO
// and retires the rest, so it publishes its parameters, announces itself as
// destructive, and asks before running.
func TestMergeIsAnAggregateDestructiveAction(t *testing.T) {
	for _, info := range registered(t).BulkActions {
		if info.Name != "merge" {
			continue
		}
		if info.ToolHints.DestructiveHint == nil || !*info.ToolHints.DestructiveHint {
			t.Fatal("merge retires the TODOs it folds in and must carry a destructive hint")
		}
		if info.ToolHints.DefaultPermission != entity.ToolPermissionAsk {
			t.Fatalf("merge must default to asking, got %q", info.ToolHints.DefaultPermission)
		}
		if info.FlagsType == nil || info.FlagsType.Name() != "MergeFlags" {
			t.Fatalf("merge flags type = %v", info.FlagsType)
		}
		return
	}
	t.Fatal("merge action was not declared")
}

// Every merge parameter must be optional: the dashboard's toolbar dispatches an
// action with no parameters unless they are a closed set of values, so a
// required one would make merge unreachable from the UI it was built for.
func TestMergeParametersAreAllOptional(t *testing.T) {
	root := &cobra.Command{Use: "gavel"}
	clicky.GenerateCLI(root)

	cmd, _, err := root.Find([]string{"todo", "merge"})
	if err != nil {
		t.Fatalf("find merge: %v", err)
	}
	for _, flag := range []string{"into", "dry-run", "model", "effort"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Fatalf("`gavel todo merge` must expose --%s", flag)
		}
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if annotations := flag.Annotations[cobra.BashCompOneRequiredFlag]; len(annotations) > 0 && annotations[0] == "true" {
			t.Fatalf("--%s is required, but the dashboard cannot supply merge parameters", flag.Name)
		}
	})
}

// push publishes outside gavel, so an agent must ask before running it.
func TestPushAsksBeforePublishing(t *testing.T) {
	for _, info := range registered(t).BulkActions {
		if info.Name == "push" {
			if info.ToolHints.DefaultPermission != entity.ToolPermissionAsk {
				t.Fatalf("push must default to asking, got %q", info.ToolHints.DefaultPermission)
			}
			return
		}
	}
	t.Fatal("push action was not declared")
}
