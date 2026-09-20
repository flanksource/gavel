package bulk

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/types"
)

// PushFlags are the parameters of the `push` bulk action.
type PushFlags struct {
	// Update rewrites the issue an already-linked TODO points at. Without it a
	// linked TODO fails on its own rather than opening a duplicate issue.
	Update  bool   `flag:"update" help:"Rewrite the issue an already-linked TODO points at instead of failing it"`
	BaseURL string `flag:"base-url" help:"Absolute origin attachment links resolve against (default: each workspace's todos.baseUrl)"`
}

func (PushFlags) ClickyActionFlags() {}

// PushBaseURL resolves the absolute origin for one TODO's workspace. It is
// injected because the project config it reads lives with the host, and a
// batch can span workspaces with different configs.
type PushBaseURL func(dir, requested string) (string, error)

// Push opens — or with Update rewrites — a GitHub issue per TODO, targeting the
// repo of the workspace that owns it.
func Push(flags PushFlags, resolveBaseURL PushBaseURL) (ItemFunc, error) {
	if resolveBaseURL == nil {
		return nil, fmt.Errorf("push: a base URL resolver is required")
	}
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		baseURL, err := resolveBaseURL(todo.CWD, flags.BaseURL)
		if err != nil {
			return ItemResult{}, err
		}
		result, err := githubpush.Push(ctx, provider, todo.ID, githubpush.Options{
			GitHub:  github.Options{WorkDir: todo.CWD},
			BaseURL: baseURL,
			Update:  flags.Update,
			Labels:  true,
			Plan:    true,
		})
		if err != nil {
			return ItemResult{}, err
		}
		status := "opened"
		if result.Updated {
			status = "updated"
		}
		return ItemResult{URL: result.URL, Status: status}, nil
	}, nil
}
