package bulk

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// familyStore is versioned storage with the single-level hierarchy: a write
// through a copy that is no longer current is refused, and so is deleting a
// TODO that still has open children. Every read hands out a copy, as a real
// provider does, so a batch holds exactly the staleness it would in production.
type familyStore struct {
	todos.Provider
	stored map[string]*types.TODO
}

func (s *familyStore) add(id, parentID string) *types.TODO {
	if s.stored == nil {
		s.stored = map[string]*types.TODO{}
	}
	todo := &types.TODO{ID: id, ShortID: id, ParentID: parentID, Version: 1}
	todo.Title, todo.Status = "TODO "+id, types.StatusPending
	s.stored[id] = todo
	return s.current(id)
}

func (s *familyStore) current(id string) *types.TODO {
	copied := *s.stored[id]
	return &copied
}

func (s *familyStore) targets(ids ...string) []Target {
	out := make([]Target, 0, len(ids))
	for _, id := range ids {
		out = append(out, Target{Ref: id, Provider: s, Todo: s.current(id)})
	}
	return out
}

func (s *familyStore) List(_ context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	var listed types.TODOS
	for _, id := range slices.Sorted(maps.Keys(s.stored)) {
		if filters.Matches(s.stored[id]) {
			listed = append(listed, s.current(id))
		}
	}
	return listed, nil
}

func (s *familyStore) Get(_ context.Context, ref string) (*types.TODO, error) {
	if _, ok := s.stored[ref]; !ok {
		return nil, fmt.Errorf("no TODO matched %q", ref)
	}
	return s.current(ref), nil
}

func (s *familyStore) write(todo *types.TODO, mutate func(*types.TODO)) error {
	stored := s.stored[todo.ID]
	if stored.Version != todo.Version {
		return errors.New("version conflict")
	}
	mutate(stored)
	stored.Version++
	*todo = *stored
	return nil
}

func (s *familyStore) Delete(ctx context.Context, todo *types.TODO) error {
	open, err := s.List(ctx, todos.DiscoveryFilters{ParentID: todo.ID, ExcludeStatuses: []types.Status{types.StatusCompleted}})
	if err != nil {
		return err
	}
	if len(open) > 0 {
		return fmt.Errorf("%d open children under %s; %s", len(open), todo.ID, todos.ChildrenChoice)
	}
	return s.write(todo, func(stored *types.TODO) { stored.Status = types.StatusCompleted })
}

func (s *familyStore) SetParent(_ context.Context, todo *types.TODO, parentRef string) error {
	return s.write(todo, func(stored *types.TODO) { stored.ParentID = parentRef })
}

// family is a parent with one open child and an unrelated TODO with none.
func family() *familyStore {
	store := &familyStore{}
	store.add("parent", "")
	store.add("child", "parent")
	store.add("plain", "")
	return store
}

func deleteWith(t *testing.T, store *familyStore, children string, ids ...string) Result {
	t.Helper()
	fn, err := Delete(DeleteFlags{Confirm: true, Children: children})
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	return Apply(context.Background(), "delete", store.targets(ids...), fn)
}

func statuses(result Result) map[string]string {
	out := map[string]string{}
	for _, item := range result.Results {
		out[item.Ref] = item.Status
		if item.Error != "" {
			out[item.Ref] = "error: " + item.Error
		}
	}
	return out
}

func TestDeleteRejectsAnUnknownChildrenChoiceBeforeAnyWrite(t *testing.T) {
	_, err := Delete(DeleteFlags{Confirm: true, Children: "purge"})
	if err == nil || !strings.Contains(err.Error(), `unknown children value "purge"`) {
		t.Fatalf("Delete = %v, want the unknown choice rejected", err)
	}
}

// The dashboard sends one choice for the whole batch, plain TODOs included.
func TestDeleteAppliesTheChildrenChoiceOnlyWhereThereAreOpenChildren(t *testing.T) {
	for _, tc := range []struct {
		children        string
		wantParent      string
		wantChildStatus types.Status
		wantChildParent string
	}{
		{"archive", "deleted; open children archived too: child", types.StatusCompleted, "parent"},
		{"detach", "deleted; open children made top-level TODOs: child", types.StatusPending, ""},
	} {
		t.Run(tc.children, func(t *testing.T) {
			store := family()

			result := deleteWith(t, store, tc.children, "parent", "plain")

			want := map[string]string{"parent": tc.wantParent, "plain": "deleted"}
			if got := statuses(result); !maps.Equal(got, want) {
				t.Errorf("results = %v, want %v", got, want)
			}
			child := store.stored["child"]
			if child.Status != tc.wantChildStatus || child.ParentID != tc.wantChildParent {
				t.Errorf("child is %s under %q, want %s under %q", child.Status, child.ParentID, tc.wantChildStatus, tc.wantChildParent)
			}
		})
	}
}

// Without a choice the parent is one failed item, carrying the provider's
// refusal, and the rest of the batch still lands.
func TestDeleteWithoutAChoiceFailsOnlyTheParentWithOpenChildren(t *testing.T) {
	store := family()

	result := deleteWith(t, store, "", "parent", "plain")

	want := map[string]string{
		"parent": "error: 1 open children under parent; " + todos.ChildrenChoice,
		"plain":  "deleted",
	}
	if got := statuses(result); !maps.Equal(got, want) {
		t.Errorf("results = %v, want %v", got, want)
	}
	if result.Applied != 1 || result.Failed != 1 {
		t.Errorf("applied=%d failed=%d, want 1/1", result.Applied, result.Failed)
	}
	if store.stored["parent"].Status != types.StatusPending || store.stored["child"].Status != types.StatusPending {
		t.Errorf("the refused parent and its child must stay open")
	}
}

// Select-all ticks a parent and its children together. Settling the children
// through the parent moves them on from the copies the batch resolved, so each
// child is read again before its own delete instead of failing on a stale one.
func TestDeleteOfAParentAndItsSelectedChildDeletesBoth(t *testing.T) {
	for _, children := range []string{"archive", "detach"} {
		t.Run(children, func(t *testing.T) {
			store := family()

			result := deleteWith(t, store, children, "parent", "child")

			if result.Failed != 0 {
				t.Fatalf("results = %v, want no failures", statuses(result))
			}
			for _, id := range []string{"parent", "child"} {
				if store.stored[id].Status != types.StatusCompleted {
					t.Errorf("%s is %s, want it deleted", id, store.stored[id].Status)
				}
			}
		})
	}
}
