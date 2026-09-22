package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// crossWorkspaceProvider is one workspace's provider: it holds only its own
// TODOs, and answers a global reference from the whole database the way the
// native runtime's GlobalGet does.
type crossWorkspaceProvider struct {
	todos.Provider
	dir string
	// own is keyed by every reference this workspace resolves directly.
	own map[string]*types.TODO
	// global is the database: every TODO, keyed by UUID, regardless of workspace.
	global    map[string]*types.TODO
	getRefs   []string
	listCalls int
}

func (p *crossWorkspaceProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	p.getRefs = append(p.getRefs, ref)
	if todo, ok := p.own[ref]; ok {
		return todo, nil
	}
	return nil, fmt.Errorf("native todo record not found: issue reference %q", ref)
}

func (p *crossWorkspaceProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	p.listCalls++
	out := make(types.TODOS, 0, len(p.own))
	for _, todo := range p.own {
		out = append(out, todo)
	}
	return out, nil
}

func (p *crossWorkspaceProvider) GetGlobal(_ context.Context, ref string) (*types.TODO, error) {
	if todo, ok := p.global[ref]; ok {
		return todo, nil
	}
	return nil, os.ErrNotExist
}

const (
	hereDir  = "/repo/here"
	otherDir = "/repo/other"
	otherID  = "b603acc4-105f-42e0-8862-ceee713a2d7b"
)

// crossWorkspaceSetup wires a caller workspace that holds nothing and an owning
// workspace that holds the TODO, reachable only by UUID.
func crossWorkspaceSetup(t *testing.T) (*crossWorkspaceProvider, *crossWorkspaceProvider, *types.TODO) {
	t.Helper()
	owned := &types.TODO{
		ID: otherID, ShortID: "b603acc4", Provider: todos.ProviderDB,
		TODOFrontmatter: types.TODOFrontmatter{Title: "Lives elsewhere", Status: types.StatusPending, CWD: otherDir},
	}
	database := map[string]*types.TODO{otherID: owned}

	here := &crossWorkspaceProvider{dir: hereDir, own: map[string]*types.TODO{}, global: database}
	other := &crossWorkspaceProvider{
		dir: otherDir, own: map[string]*types.TODO{otherID: owned, "b603acc4": owned}, global: database,
	}

	old := openRuntimeTodosProvider
	openRuntimeTodosProvider = func(_ context.Context, dir string) (todos.Provider, error) {
		if dir == otherDir {
			return other, nil
		}
		return nil, fmt.Errorf("unexpected workspace %q", dir)
	}
	t.Cleanup(func() { openRuntimeTodosProvider = old })
	return here, other, owned
}

// A UUID names one issue in the entire database, so there is nothing for a
// workspace filter to disambiguate — and filtering is what made `todos run <uuid>`
// report "native todo record not found" for an issue that existed the whole time.
func TestResolveRequestedTargetsResolvesAUUIDAcrossWorkspaces(t *testing.T) {
	here, other, owned := crossWorkspaceSetup(t)

	targets, err := resolveRequestedTargets(context.Background(), here, hereDir, []string{otherID}, todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("resolveRequestedTargets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].Todo != owned {
		t.Errorf("resolved %+v, want the TODO from the other workspace", targets[0].Todo)
	}
	// The run has to dispatch through the workspace that owns the issue, or it
	// would execute against this directory's tree and .gavel.yaml.
	if targets[0].WorkDir != otherDir {
		t.Errorf("workDir = %q, want %q", targets[0].WorkDir, otherDir)
	}
	if targets[0].Provider != todos.Provider(other) {
		t.Errorf("provider = %p, want the owning workspace's", targets[0].Provider)
	}
	// GlobalGet reconciles nothing, so the owning workspace must re-read its issue.
	if len(other.getRefs) != 1 || other.getRefs[0] != otherID {
		t.Errorf("owning workspace Get refs = %v, want a re-read of %s", other.getRefs, otherID)
	}
}

// A short id can name different TODOs in different projects, so it keeps the
// workspace filter that makes it unambiguous.
func TestResolveRequestedTargetsKeepsShortIDsWorkspaceScoped(t *testing.T) {
	here, _, _ := crossWorkspaceSetup(t)

	_, err := resolveRequestedTargets(context.Background(), here, hereDir, []string{"b603acc4"}, todos.DiscoveryFilters{})
	if err == nil {
		t.Fatal("a short id from another workspace must not resolve")
	}
	if !strings.Contains(err.Error(), `"b603acc4"`) {
		t.Errorf("error %q should quote the reference", err)
	}
}

// A UUID nothing owns must report the workspace-scoped failure, not a global miss
// that hides it.
func TestResolveRequestedTargetsReportsAnUnknownUUID(t *testing.T) {
	here, _, _ := crossWorkspaceSetup(t)
	const missing = "11111111-2222-3333-4444-555555555555"

	_, err := resolveRequestedTargets(context.Background(), here, hereDir, []string{missing}, todos.DiscoveryFilters{})
	if err == nil {
		t.Fatal("an unknown UUID must fail")
	}
	for _, want := range []string{missing, "not found"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

// The single-provider commands resolve the UUID too, but must refuse to act on it
// through the wrong workspace rather than writing against this directory.
func TestResolveRequestedTODOsRefusesAnotherWorkspace(t *testing.T) {
	here, _, _ := crossWorkspaceSetup(t)

	_, err := resolveRequestedTODOs(context.Background(), here, hereDir, []string{otherID}, todos.DiscoveryFilters{})
	if err == nil {
		t.Fatal("acting on another workspace's TODO through this provider must be refused")
	}
	for _, want := range []string{otherDir, "--cwd"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

// A UUID the caller's own workspace holds must not take the global path at all —
// the fast path stays a single Get.
func TestResolveRequestedTargetsPrefersTheLocalWorkspace(t *testing.T) {
	_, other, owned := crossWorkspaceSetup(t)

	targets, err := resolveRequestedTargets(context.Background(), other, otherDir, []string{otherID}, todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("resolveRequestedTargets: %v", err)
	}
	if len(targets) != 1 || targets[0].Todo != owned {
		t.Fatalf("targets = %+v, want the local TODO", targets)
	}
	if targets[0].WorkDir != otherDir || targets[0].Provider != todos.Provider(other) {
		t.Errorf("target = %+v, want the caller's own workspace", targets[0])
	}
	if len(other.getRefs) != 1 {
		t.Errorf("Get calls = %v, want exactly one for a locally held UUID", other.getRefs)
	}
	if other.listCalls != 0 {
		t.Errorf("List calls = %d, want 0 for a directly resolvable UUID", other.listCalls)
	}
}
