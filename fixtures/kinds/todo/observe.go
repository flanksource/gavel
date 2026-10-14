package todo

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// todoState is one observation of a TODO: the item as GET /api/todos/item
// returns it (pr/ui todoSummary) and its prompt-run attempts as GET
// /api/todos/session/detail returns them (pr/ui todoAttemptDetail), newest
// first.
type todoState struct {
	ID       string
	Dir      string
	Item     map[string]any
	Attempts []any
}

// observeTodo reads the item, then its attempts through the item's owning
// workspace (`cwd`): session detail resolves a todo inside one workspace, and
// the server's default workspace is not necessarily the todo's.
func (c *apiClient) observeTodo(ctx context.Context, ref, dir string) (todoState, error) {
	query := url.Values{"ref": {ref}}
	if dir != "" {
		query.Set("dir", dir)
	}
	var item map[string]any
	if err := c.getJSON(ctx, "/api/todos/item", query, &item); err != nil {
		return todoState{}, err
	}
	id := stringField(item, "id")
	cwd := stringField(item, "cwd")
	if id == "" || cwd == "" {
		return todoState{}, fmt.Errorf("GET /api/todos/item?ref=%s: response has no id/cwd (id=%q, cwd=%q)", ref, id, cwd)
	}
	// todoSummary omits empty labels; CEL over `todo.labels` needs the list.
	if _, ok := item["labels"]; !ok {
		item["labels"] = []any{}
	}
	var detail struct {
		Attempts []any `json:"attempts"`
	}
	if err := c.getJSON(ctx, "/api/todos/session/detail", url.Values{"dir": {cwd}, "ref": {id}}, &detail); err != nil {
		return todoState{}, err
	}
	if detail.Attempts == nil {
		return todoState{}, fmt.Errorf("GET /api/todos/session/detail?ref=%s: response has no attempts list", id)
	}
	return todoState{ID: id, Dir: cwd, Item: item, Attempts: detail.Attempts}, nil
}

// outputs are the CEL roots every todo observation shares.
func (s todoState) outputs() map[string]any {
	return map[string]any{
		"id":       s.ID,
		"todo":     s.Item,
		"triage":   attemptSummary(s.latestAttempt(isTriageStep)),
		"attempts": s.Attempts,
	}
}

// latestAttempt is the newest attempt whose step matches.
func (s todoState) latestAttempt(match func(step string) bool) map[string]any {
	for _, raw := range s.Attempts {
		attempt, ok := raw.(map[string]any)
		if ok && match(stringField(attempt, "step")) {
			return attempt
		}
	}
	return nil
}

// attemptByRun is the attempt of one admitted prompt run.
func (s todoState) attemptByRun(promptRunID string) (map[string]any, error) {
	for _, raw := range s.Attempts {
		if attempt, ok := raw.(map[string]any); ok && stringField(attempt, "promptRunId") == promptRunID {
			return attempt, nil
		}
	}
	return nil, fmt.Errorf("admitted prompt run %s is not among TODO %s's %d attempts", promptRunID, s.ID, len(s.Attempts))
}

func isTriageStep(step string) bool { return step == "triage" || strings.HasPrefix(step, "triage.") }

// attemptSummary projects an attempt for CEL. `state` is captain's display
// status (running, planning, ask, zombie, completed, failed, cancelled) — the
// vocabulary the dashboard shows; `runState` is the prompt run's own state
// (pending, running, waiting, succeeded, failed, cancelled). A missing attempt
// is found=false with empty strings, so `triage.state == 'completed'` is false
// rather than a CEL error before the run is linked.
func attemptSummary(attempt map[string]any) map[string]any {
	return map[string]any{
		"found":       attempt != nil,
		"state":       stringField(attempt, "status"),
		"runState":    stringField(attempt, "state"),
		"error":       stringField(attempt, "error"),
		"step":        stringField(attempt, "step"),
		"promptRunId": stringField(attempt, "promptRunId"),
		"sessionId":   stringField(attempt, "sessionId"),
	}
}

// stringField reads a string a handler may omit (omitempty); absent is empty.
func stringField(m map[string]any, key string) string {
	value, _ := m[key].(string)
	return value
}

// mapField reads an object a handler may omit or send as null; absent is an
// empty map so CEL can test its keys.
func mapField(m map[string]any, key string) map[string]any {
	if value, ok := m[key].(map[string]any); ok {
		return value
	}
	return map[string]any{}
}
