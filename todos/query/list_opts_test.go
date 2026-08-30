package query

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

var now = time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC)

func todo(title string, mutate ...func(*types.TODO)) *types.TODO {
	t := &types.TODO{}
	t.Title = title
	t.Status = types.StatusPending
	t.Priority = types.PriorityMedium
	for _, fn := range mutate {
		fn(t)
	}
	return t
}

func ranAt(d time.Duration) func(*types.TODO) {
	return func(t *types.TODO) {
		at := now.Add(-d)
		t.LastRun = &at
	}
}

// The store understands statuses and labels; everything else is applied after
// the read. Getting that split wrong means a caller believes the rows are fully
// filtered when they are not.
func TestDiscoveryProjectsOnlyWhatTheStoreUnderstands(t *testing.T) {
	opts := ListOpts{
		Status:        []string{"pending", "draft"},
		ExcludeStatus: []string{"completed"},
		Label:         []string{"area:ui"},
		Priority:      []string{"high"},
		Search:        "ignored by the store",
	}
	filters, err := opts.Discovery()
	if err != nil {
		t.Fatalf("Discovery: %v", err)
	}
	if len(filters.IncludeStatuses) != 2 || filters.IncludeStatuses[0] != types.StatusPending {
		t.Fatalf("include statuses = %v", filters.IncludeStatuses)
	}
	if len(filters.ExcludeStatuses) != 1 || filters.ExcludeStatuses[0] != types.StatusCompleted {
		t.Fatalf("exclude statuses = %v", filters.ExcludeStatuses)
	}
	if len(filters.IncludeLabels) != 1 || filters.IncludeLabels[0] != "area:ui" {
		t.Fatalf("labels = %v", filters.IncludeLabels)
	}
}

// A projected status is not writable but is perfectly meaningful to filter on;
// rejecting it would make "show me what is running" unexpressible.
func TestDiscoveryAcceptsProjectedStatusesAndRejectsUnknownOnes(t *testing.T) {
	if _, err := (ListOpts{Status: []string{"in_progress"}}).Discovery(); err != nil {
		t.Fatalf("a projected status must be selectable: %v", err)
	}
	if _, err := (ListOpts{Status: []string{"not-a-status"}}).Discovery(); err == nil {
		t.Fatal("an unknown status must be rejected rather than matching nothing silently")
	}
}

func TestMatchAppliesFacetsTheStoreCannot(t *testing.T) {
	cases := []struct {
		name string
		opts ListOpts
		todo *types.TODO
		want bool
	}{
		{"priority hit", ListOpts{Priority: []string{"high"}},
			todo("x", func(t *types.TODO) { t.Priority = types.PriorityHigh }), true},
		{"priority miss", ListOpts{Priority: []string{"high"}}, todo("x"), false},
		{"priority is case-insensitive", ListOpts{Priority: []string{"HIGH"}},
			todo("x", func(t *types.TODO) { t.Priority = types.PriorityHigh }), true},
		{"excluded label", ListOpts{ExcludeLabel: []string{"wontfix"}},
			todo("x", func(t *types.TODO) { t.Labels = []string{"wontfix"} }), false},
		{"unexcluded label", ListOpts{ExcludeLabel: []string{"wontfix"}},
			todo("x", func(t *types.TODO) { t.Labels = []string{"bug"} }), true},
		{"search matches case-insensitively", ListOpts{Search: "REFACTOR"}, todo("Refactor the parser"), true},
		{"search miss", ListOpts{Search: "parser"}, todo("Ship the release"), false},
		{"recent enough", ListOpts{Since: "24h"}, todo("x", ranAt(2*time.Hour)), true},
		{"too old", ListOpts{Since: "24h"}, todo("x", ranAt(48*time.Hour)), false},
		{"days window", ListOpts{Since: "7d"}, todo("x", ranAt(3*24*time.Hour)), true},
		{"weeks window", ListOpts{Since: "2w"}, todo("x", ranAt(20*24*time.Hour)), false},
		// A TODO that never ran has no activity to be recent. Keeping it would
		// make "active this week" quietly include the whole untouched backlog.
		{"never run fails a recency filter", ListOpts{Since: "24h"}, todo("x"), false},
		{"no facets matches everything", ListOpts{}, todo("x"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.opts.Match(tc.todo, now)
			if err != nil {
				t.Fatalf("Match: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Match = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchRejectsAnUnparseableWindow(t *testing.T) {
	if _, err := (ListOpts{Since: "soon"}).Match(todo("x", ranAt(time.Hour)), now); err == nil {
		t.Fatal("an unparseable window must fail loudly, not match everything")
	}
}

func TestFilterModeIsDrivenByTheFilterField(t *testing.T) {
	if (ListOpts{}).FilterMode() {
		t.Fatal("no filter means explicit ids")
	}
	if (ListOpts{Filter: "   "}).FilterMode() {
		t.Fatal("whitespace is not a filter")
	}
	if !(ListOpts{Filter: "status == pending"}).FilterMode() {
		t.Fatal("a filter summary switches to filter mode")
	}
}

type stubLister struct {
	got  todos.DiscoveryFilters
	rows types.TODOS
}

func (s *stubLister) List(filters todos.DiscoveryFilters) (types.TODOS, error) {
	s.got = filters
	return s.rows, nil
}

func TestSelectPushesDownAndThenNarrows(t *testing.T) {
	lister := &stubLister{rows: types.TODOS{
		todo("keep", func(t *types.TODO) { t.Priority = types.PriorityHigh }),
		todo("drop", func(t *types.TODO) { t.Priority = types.PriorityLow }),
	}}
	opts := ListOpts{Status: []string{"pending"}, Priority: []string{"high"}}

	selected, err := opts.Select(lister, now)
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if len(lister.got.IncludeStatuses) != 1 {
		t.Fatalf("status should be pushed down to the store, got %+v", lister.got)
	}
	if len(selected) != 1 || selected[0].Title != "keep" {
		t.Fatalf("severity should narrow the listed rows, got %d rows", len(selected))
	}
}
