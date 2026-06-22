package todos

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

// Transfer moves a todo from the source provider to the target provider. It
// reads the full todo from source, recreates it in target, then removes it from
// source so the todo ends up in exactly one place.
//
// Title, body, priority and status carry over. Provider-specific history (Grite
// comments/events, file attempt logs) and execution metadata (attempts, last
// run) do not — the target gets a fresh todo with the same content. Transfer
// works across backends, so a Grite issue can move into a .todos workspace and
// vice versa.
//
// If the target Create fails, nothing is moved and the source is left untouched.
// If Create succeeds but the source Delete fails, the created todo is returned
// alongside the error so callers can surface the resulting duplicate.
func Transfer(ctx context.Context, source, target Provider, ref string) (*types.TODO, error) {
	if source == nil || target == nil {
		return nil, fmt.Errorf("source and target providers are required")
	}
	todo, err := source.Get(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("load source todo %q: %w", ref, err)
	}
	created, err := target.Create(ctx, CreateRequest{
		Title:    todo.Title,
		Body:     transferBody(todo),
		Priority: todo.Priority,
		Status:   todo.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("create todo in target: %w", err)
	}
	if err := source.Delete(ctx, todo); err != nil {
		return created, fmt.Errorf("todo copied to target but removing it from source failed: %w", err)
	}
	return created, nil
}

// transferBody picks the fullest body representation a provider populates on
// Get: MarkdownBody holds the whole markdown body for both backends, with
// Implementation as the fallback (mirrors the dashboard's summarizeTodo).
func transferBody(todo *types.TODO) string {
	if body := strings.TrimSpace(todo.MarkdownBody); body != "" {
		return body
	}
	return strings.TrimSpace(todo.Implementation)
}
