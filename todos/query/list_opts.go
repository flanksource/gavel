// Package query owns the TODO selector: the one type that says which TODOs a
// caller means, whether that caller is a CLI invocation, an HTTP query string,
// or a bulk action asked to run against "everything matching these filters".
//
// It is deliberately one type and not three. The CLI had `DiscoveryFilters`,
// the dashboard had its own facet set, and a bulk endpoint had a list of
// explicit targets — so "the pending, high-severity TODOs" meant a different
// thing in each, and none of them could express what the other two could.
package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// ListOpts selects TODOs. It is the entity's list options, the CLI's flags and
// the bulk-action filter all at once — clicky binds the `flag:` tags as cobra
// flags and as OpenAPI query parameters, so the three surfaces cannot drift.
//
// Every facet is include/exclude rather than a single value, because that is
// what the dashboard's filter bar already offers and a selector that cannot say
// "everything except completed" cannot express the default view.
type ListOpts struct {
	// Dir is the workspace. Empty means the caller's working directory, resolved
	// by whoever opens the provider — the selector does not know about projects.
	Dir string `flag:"dir" help:"Workspace directory; defaults to the current one"`

	Status        []string `flag:"status" help:"Only TODOs in these statuses"`
	ExcludeStatus []string `flag:"exclude-status" help:"Skip TODOs in these statuses"`
	Priority      []string `flag:"priority" help:"Only TODOs at these severities" enum:"high,medium,low"`
	Label         []string `flag:"label" help:"Only TODOs carrying all of these labels"`
	ExcludeLabel  []string `flag:"exclude-label" help:"Skip TODOs carrying any of these labels"`

	// Since bounds by last activity, e.g. 24h or 7d. It reads the same field the
	// dashboard's activity range filters on.
	Since string `flag:"since" help:"Only TODOs active within this window (e.g. 24h, 7d)"`

	// Search is a case-insensitive substring match on the title.
	Search string `flag:"search" help:"Match TODO titles containing this text"`

	// Filter switches a bulk action from its explicit ids to this selector.
	// clicky triggers filter mode on this field being non-empty, so it carries a
	// human-readable summary of what is being matched rather than a bare "true"
	// — it ends up in the run's audit trail.
	Filter string `flag:"filter" help:"Run a bulk action against every matching TODO instead of named ids"`
}

// FilterMode reports whether a bulk action carrying these options was asked to
// resolve its own selection rather than act on named ids.
func (o ListOpts) FilterMode() bool { return strings.TrimSpace(o.Filter) != "" }

// Discovery projects the options onto the filter the provider can push down.
// Status and labels are the only facets the store understands; severity, recency
// and title matching are applied by Match afterwards. Keeping the projection
// explicit is what stops a caller assuming the returned rows are already fully
// filtered.
func (o ListOpts) Discovery() (todos.DiscoveryFilters, error) {
	filters := todos.DiscoveryFilters{IncludeLabels: trimAll(o.Label)}
	include, err := parseStatuses(o.Status)
	if err != nil {
		return filters, err
	}
	exclude, err := parseStatuses(o.ExcludeStatus)
	if err != nil {
		return filters, err
	}
	filters.IncludeStatuses = include
	filters.ExcludeStatuses = exclude
	return filters, nil
}

// Match applies the facets the store cannot. `now` is passed rather than read so
// a caller filtering a page of rows gets one consistent clock for all of them.
func (o ListOpts) Match(todo *types.TODO, now time.Time) (bool, error) {
	if todo == nil {
		return false, nil
	}
	if priorities := trimAll(o.Priority); len(priorities) > 0 {
		if !containsFold(priorities, string(todo.Priority)) {
			return false, nil
		}
	}
	for _, label := range trimAll(o.ExcludeLabel) {
		if containsFold(todo.Labels, label) {
			return false, nil
		}
	}
	if search := strings.TrimSpace(o.Search); search != "" {
		if !strings.Contains(strings.ToLower(todo.Title), strings.ToLower(search)) {
			return false, nil
		}
	}
	if window := strings.TrimSpace(o.Since); window != "" {
		cutoff, err := ParseWindow(window, now)
		if err != nil {
			return false, err
		}
		last := lastActivity(todo)
		// A TODO that has never run has no activity to be recent, so a recency
		// filter excludes it rather than silently keeping it.
		if last.IsZero() || last.Before(cutoff) {
			return false, nil
		}
	}
	return true, nil
}

// Select lists the workspace's TODOs and returns the ones these options match.
func (o ListOpts) Select(ctx contextLister, now time.Time) (types.TODOS, error) {
	filters, err := o.Discovery()
	if err != nil {
		return nil, err
	}
	listed, err := ctx.List(filters)
	if err != nil {
		return nil, err
	}
	matched := make(types.TODOS, 0, len(listed))
	for _, todo := range listed {
		ok, err := o.Match(todo, now)
		if err != nil {
			return nil, err
		}
		if ok {
			matched = append(matched, todo)
		}
	}
	return matched, nil
}

// contextLister is the sliver of a provider Select needs, so the selector can be
// unit-tested without a database and without importing a provider construction
// seam.
type contextLister interface {
	List(filters todos.DiscoveryFilters) (types.TODOS, error)
}

// ParseWindow turns a relative window ("24h", "7d", "2w") into the cutoff before
// which a TODO is considered stale. Days and weeks are spelled out because
// time.ParseDuration stops at hours and a backlog is reasoned about in days.
func ParseWindow(window string, now time.Time) (time.Time, error) {
	raw := strings.ToLower(strings.TrimSpace(window))
	if raw == "" {
		return time.Time{}, fmt.Errorf("window is required")
	}
	multiplier := time.Duration(0)
	switch {
	case strings.HasSuffix(raw, "d"):
		multiplier = 24 * time.Hour
		raw = strings.TrimSuffix(raw, "d")
	case strings.HasSuffix(raw, "w"):
		multiplier = 7 * 24 * time.Hour
		raw = strings.TrimSuffix(raw, "w")
	}
	if multiplier > 0 {
		var count float64
		if _, err := fmt.Sscanf(raw, "%g", &count); err != nil || count <= 0 {
			return time.Time{}, fmt.Errorf("invalid window %q", window)
		}
		return now.Add(-time.Duration(count * float64(multiplier))), nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil || parsed <= 0 {
		return time.Time{}, fmt.Errorf("invalid window %q", window)
	}
	return now.Add(-parsed), nil
}

func lastActivity(todo *types.TODO) time.Time {
	if todo.LastRun != nil {
		return *todo.LastRun
	}
	return time.Time{}
}

// parseStatuses accepts any *known* status, not just the assignable ones. A
// selector says which TODOs you mean, and "show me everything in progress" is a
// reasonable thing to mean even though `in progress` is projected from the last
// run and cannot be written.
func parseStatuses(values []string) ([]types.Status, error) {
	trimmed := trimAll(values)
	if len(trimmed) == 0 {
		return nil, nil
	}
	statuses := make([]types.Status, 0, len(trimmed))
	for _, value := range trimmed {
		status := types.Status(strings.ToLower(value))
		if !types.IsKnownStatus(status) {
			return nil, fmt.Errorf("unknown status %q", value)
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func trimAll(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func containsFold(haystack []string, needle string) bool {
	for _, value := range haystack {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(needle)) {
			return true
		}
	}
	return false
}
