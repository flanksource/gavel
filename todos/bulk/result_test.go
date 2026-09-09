package bulk

import (
	"context"
	"errors"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func todoRef(ref string) *types.TODO {
	t := &types.TODO{ShortID: ref}
	t.Title = ref
	return t
}

func targets(provider todos.Provider, refs ...string) []Target {
	out := make([]Target, 0, len(refs))
	for _, ref := range refs {
		out = append(out, Target{Ref: ref, Provider: provider, Todo: todoRef(ref)})
	}
	return out
}

// The reason this package exists: one failure in the middle must not cost the
// caller the items that worked.
func TestApplyReportsPerItemAndKeepsGoing(t *testing.T) {
	result := Apply(context.Background(), "status", targets(nil, "a", "b", "c"),
		func(_ context.Context, _ todos.Provider, todo *types.TODO) (ItemResult, error) {
			if todo.Title == "b" {
				return ItemResult{}, errors.New("archived in another tab")
			}
			return ItemResult{Status: "completed"}, nil
		})

	if result.Applied != 2 || result.Failed != 1 {
		t.Fatalf("applied=%d failed=%d, want 2/1", result.Applied, result.Failed)
	}
	if len(result.Results) != 3 {
		t.Fatalf("every item must be reported, got %d", len(result.Results))
	}
	if result.Results[1].Error != "archived in another tab" {
		t.Fatalf("the failure must name itself, got %q", result.Results[1].Error)
	}
	if result.Results[2].Status != "completed" {
		t.Fatal("items after a failure must still run")
	}
}

// A cancelled batch keeps what already landed and says why the rest did not,
// rather than returning a shorter list that looks like a smaller selection.
func TestApplyReportsRemainderWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	result := Apply(ctx, "run", targets(nil, "a", "b", "c"),
		func(_ context.Context, _ todos.Provider, todo *types.TODO) (ItemResult, error) {
			if todo.Title == "a" {
				cancel()
			}
			return ItemResult{}, nil
		})

	if len(result.Results) != 3 {
		t.Fatalf("cancelled items must still be reported, got %d", len(result.Results))
	}
	if result.Applied != 1 || result.Failed != 2 {
		t.Fatalf("applied=%d failed=%d, want 1/2", result.Applied, result.Failed)
	}
	if result.Results[1].Error == "" {
		t.Fatal("a skipped item must carry the cancellation as its error")
	}
}

func TestResolveRejectsTheWholeRequestOnMalformedRefs(t *testing.T) {
	// A lookup that always succeeds, so a rejection can only come from the
	// validation pass and never from a failed resolution.
	lookup := func(_ context.Context, ref string) (todos.Provider, *types.TODO, error) {
		return nil, todoRef(ref), nil
	}
	if _, _, err := Resolve(context.Background(), lookup, nil); err == nil {
		t.Fatal("an empty selection must be rejected")
	}
	if _, _, err := Resolve(context.Background(), lookup, []string{"a", "  "}); err == nil {
		t.Fatal("a blank ref must be rejected before any write lands")
	}
	if _, _, err := Resolve(context.Background(), lookup, []string{"a", "a"}); err == nil {
		t.Fatal("a duplicated ref must be rejected: it would apply twice")
	}
}

// The dashboard groups TODOs by severity and age, so a checked column crosses
// repositories. Each item must be written through the provider that owns it —
// routing the batch through one workspace would write to the wrong database.
func TestApplyWritesEachTodoThroughItsOwnProvider(t *testing.T) {
	ui, api := &labelProvider{}, &labelProvider{}
	byRef := map[string]todos.Provider{"a": ui, "b": api, "c": ui}
	lookup := func(_ context.Context, ref string) (todos.Provider, *types.TODO, error) {
		return byRef[ref], todoRef(ref), nil
	}

	resolved, failures, err := Resolve(context.Background(), lookup, []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("no ref should have failed, got %v", failures)
	}

	seen := map[todos.Provider][]string{}
	result := Apply(context.Background(), "labels", resolved,
		func(_ context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
			seen[provider] = append(seen[provider], todo.Title)
			return ItemResult{}, nil
		})

	if result.Applied != 3 {
		t.Fatalf("applied=%d, want 3", result.Applied)
	}
	if got := seen[ui]; len(got) != 2 || got[0] != "a" || got[1] != "c" {
		t.Fatalf("ui workspace saw %v, want [a c]", got)
	}
	if got := seen[api]; len(got) != 1 || got[0] != "b" {
		t.Fatalf("api workspace saw %v, want [b]", got)
	}
}

// A ref that no longer resolves is one stale row in an open tab, not a reason
// to reject the other thirty-nine.
func TestResolveReportsUnresolvableRefsPerItem(t *testing.T) {
	lookup := func(_ context.Context, ref string) (todos.Provider, *types.TODO, error) {
		if ref == "gone" {
			return nil, nil, errors.New("no such TODO")
		}
		return nil, todoRef(ref), nil
	}

	resolved, failures, err := Resolve(context.Background(), lookup, []string{"a", "gone", "c"})
	if err != nil {
		t.Fatalf("a stale ref must not reject the request: %v", err)
	}
	if len(resolved) != 2 {
		t.Fatalf("resolved %d targets, want the 2 that still exist", len(resolved))
	}
	if len(failures) != 1 || failures[0].Ref != "gone" || failures[0].Error == "" {
		t.Fatalf("the stale ref must be reported by name, got %v", failures)
	}
}

func TestSetStatusRejectsAProjectedStatusUpFront(t *testing.T) {
	// in_progress is projected from the last run; writing it would claim
	// something the run history contradicts.
	if _, err := SetStatus(StatusFlags{To: "in_progress"}); err == nil {
		t.Fatal("a projected status must be rejected as unassignable")
	}
	if _, err := SetStatus(StatusFlags{To: "completed"}); err != nil {
		t.Fatalf("an assignable status must be accepted: %v", err)
	}
}

func TestEditLabelsRefusesContradictoryFlags(t *testing.T) {
	if _, err := EditLabels(LabelFlags{}); err == nil {
		t.Fatal("a labels action that changes nothing must be rejected")
	}
	if _, err := EditLabels(LabelFlags{Add: []string{"bug"}, Remove: []string{"BUG"}}); err == nil {
		t.Fatal("adding and removing the same label must be rejected")
	}
}

// The whole-set replace on the provider means each TODO's labels have to be
// computed against its own, not against the batch.
func TestEditLabelsComputesPerTodo(t *testing.T) {
	fn, err := EditLabels(LabelFlags{Add: []string{"area:ui"}, Remove: []string{"stale"}})
	if err != nil {
		t.Fatalf("EditLabels: %v", err)
	}
	provider := &labelProvider{}
	first := todoRef("a")
	first.Labels = []string{"stale", "bug"}
	second := todoRef("b")
	second.Labels = []string{"area:ui"}

	if _, err := fn(context.Background(), provider, first); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := fn(context.Background(), provider, second); err != nil {
		t.Fatalf("second: %v", err)
	}

	if got := provider.writes[0]; len(got) != 2 || got[0] != "bug" || got[1] != "area:ui" {
		t.Fatalf("first TODO labels = %v, want [bug area:ui]", got)
	}
	// Already carried the label; adding it again must not duplicate it.
	if got := provider.writes[1]; len(got) != 1 || got[0] != "area:ui" {
		t.Fatalf("second TODO labels = %v, want [area:ui]", got)
	}
}

func TestDeleteRefusesWithoutConfirmation(t *testing.T) {
	if _, err := Delete(DeleteFlags{}); err == nil {
		t.Fatal("delete must refuse without --confirm")
	}
	if _, err := Delete(DeleteFlags{Confirm: true}); err != nil {
		t.Fatalf("confirmed delete must be accepted: %v", err)
	}
}

// labelProvider records the label sets written, which is the only part of the
// provider these tests exercise.
type labelProvider struct {
	todos.Provider
	writes [][]string
}

func (p *labelProvider) Edit(_ context.Context, _ *types.TODO, edit todos.EditRequest) error {
	if edit.Labels != nil {
		p.writes = append(p.writes, *edit.Labels)
	}
	return nil
}
