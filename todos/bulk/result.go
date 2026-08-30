// Package bulk applies one operation to many TODOs and reports what happened to
// each of them.
//
// The per-item reporting is the whole point. A batch of forty is a batch
// because the caller does not want to babysit forty requests, and one archived
// TODO in the middle must not lose the other thirty-nine. That constraint also
// dictates the error convention below, which is unusual enough to be worth
// reading before adding an action.
package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// ItemResult is one TODO's outcome. Exactly one of Error or the success fields
// is meaningful; which success field is populated depends on the action.
type ItemResult struct {
	Ref   string `json:"ref"`
	Dir   string `json:"dir,omitempty"`
	Title string `json:"title,omitempty"`
	// SessionID is set by the run-shaped actions (run/plan/triage).
	SessionID string `json:"sessionId,omitempty"`
	Status    string `json:"status,omitempty"`
	// URL is set by actions that publish somewhere, e.g. a pushed GitHub issue.
	URL   string `json:"url,omitempty"`
	Error string `json:"error,omitempty"`
}

// Result is the aggregate a bulk action returns.
//
// IMPORTANT: an action handler must return this with a **nil error** even when
// individual items failed. clicky's entity layer discards the result value
// whenever the error is non-nil (see dataOrError), so returning an error is how
// you throw away the thirty-nine items that succeeded. Reserve the error return
// for whole-request rejections — a malformed selector, an unknown status —
// which are decided before the first write lands.
type Result struct {
	Action  string       `json:"action"`
	Applied int          `json:"applied"`
	Failed  int          `json:"failed"`
	Results []ItemResult `json:"results"`
	// MatchedBy is the filter summary when the action resolved its own
	// selection, empty when it acted on explicit ids. The framework does not
	// tell a handler which mode it ran in, so the handler records it.
	MatchedBy string `json:"matchedBy,omitempty"`
}

func (r *Result) add(item ItemResult) {
	if item.Error != "" {
		r.Failed++
	} else {
		r.Applied++
	}
	r.Results = append(r.Results, item)
}

// Pretty renders the one-line summary the CLI prints.
func (r Result) String() string {
	summary := fmt.Sprintf("%s: %d applied", r.Action, r.Applied)
	if r.Failed > 0 {
		summary += fmt.Sprintf(", %d failed", r.Failed)
	}
	if r.MatchedBy != "" {
		summary += fmt.Sprintf(" (matched by %s)", r.MatchedBy)
	}
	return summary
}

// ItemFunc applies an action to one resolved TODO.
type ItemFunc func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error)

// Target is one resolved TODO together with the provider that owns it.
//
// The provider travels with the TODO rather than being fixed for the batch
// because a selection routinely spans workspaces: the dashboard groups TODOs by
// severity and age, so ticking a column of high-priority items can easily cross
// four repositories. Writing all of them through one workspace's provider would
// silently target the wrong database.
type Target struct {
	Ref      string
	Provider todos.Provider
	Todo     *types.TODO
}

// Lookup resolves one reference to its TODO and the provider that owns it.
type Lookup func(ctx context.Context, ref string) (todos.Provider, *types.TODO, error)

// TargetsFrom pairs TODOs that are already known to share a provider — the
// filter-mode case, where the selection came from listing one workspace.
func TargetsFrom(provider todos.Provider, todoList types.TODOS) []Target {
	targets := make([]Target, 0, len(todoList))
	for _, todo := range todoList {
		targets = append(targets, Target{Ref: todos.TODOReference(todo), Provider: provider, Todo: todo})
	}
	return targets
}

// Apply runs fn against every target in order and collects the outcomes.
//
// The loop is deliberately serial. The provider's optimistic version checks
// must not race each other, and a run-shaped action starts an agent session
// that has its own concurrency — forty at once would be forty agents.
func Apply(ctx context.Context, action string, targets []Target, fn ItemFunc) Result {
	result := Result{Action: action, Results: make([]ItemResult, 0, len(targets))}
	for _, target := range targets {
		item := ItemResult{Ref: target.Ref}
		if target.Todo != nil {
			item.Title = target.Todo.Title
			item.Dir = target.Todo.CWD
		}
		// A cancelled context stops the batch but keeps what already landed:
		// the remaining items are reported as failures rather than vanishing.
		if err := ctx.Err(); err != nil {
			item.Error = err.Error()
			result.add(item)
			continue
		}
		applied, err := fn(ctx, target.Provider, target.Todo)
		if err != nil {
			item.Error = err.Error()
			result.add(item)
			continue
		}
		applied.Ref = item.Ref
		if applied.Title == "" {
			applied.Title = item.Title
		}
		if applied.Dir == "" {
			applied.Dir = item.Dir
		}
		result.add(applied)
	}
	return result
}

// Resolve turns the caller's explicit ids into targets, failing the whole
// request on a blank or duplicated ref rather than half-applying. A ref that
// does not resolve is a per-item failure, not a rejection: one stale id in a
// selection of forty is the ordinary case when a dashboard tab has been open a
// while.
func Resolve(ctx context.Context, lookup Lookup, refs []string) ([]Target, []ItemResult, error) {
	if lookup == nil {
		return nil, nil, fmt.Errorf("bulk: a lookup is required to resolve references")
	}
	if len(refs) == 0 {
		return nil, nil, fmt.Errorf("at least one TODO reference is required")
	}
	// Validate the whole selection before the first lookup, let alone the first
	// write. A malformed request is the caller's mistake and rejecting it whole
	// is honest; discovering it halfway through would leave a partial batch
	// applied against a request that was never valid.
	seen := make(map[string]int, len(refs))
	cleaned := make([]string, 0, len(refs))
	for i, raw := range refs {
		ref := strings.TrimSpace(raw)
		if ref == "" {
			return nil, nil, fmt.Errorf("refs[%d] is blank", i)
		}
		if first, ok := seen[ref]; ok {
			return nil, nil, fmt.Errorf("refs[%d] duplicates refs[%d]: %q", i, first, ref)
		}
		seen[ref] = i
		cleaned = append(cleaned, ref)
	}

	resolved := make([]Target, 0, len(cleaned))
	var failures []ItemResult
	for _, ref := range cleaned {
		provider, todo, err := lookup(ctx, ref)
		if err != nil {
			failures = append(failures, ItemResult{Ref: ref, Error: err.Error()})
			continue
		}
		resolved = append(resolved, Target{Ref: ref, Provider: provider, Todo: todo})
	}
	return resolved, failures, nil
}
