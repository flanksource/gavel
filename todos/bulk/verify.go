package bulk

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// CheckFlags are the parameters of the `check` bulk action.
type CheckFlags struct {
	Timeout string `flag:"timeout" help:"Per-TODO verification timeout, e.g. 10m"`
}

func (CheckFlags) ClickyActionFlags() {}

// Check runs each selected TODO's definition of done as the lifecycle's verify
// step — the same dispatch `gavel todos check` performs, so a verdict reached
// from the dashboard and one reached from a terminal are the same verdict.
//
// A TODO that fails verification is a per-item failure, not a rejection of the
// batch: checking forty is how you find the three that broke, and aborting on
// the first would report the opposite of what was asked.
func Check(flags CheckFlags, kind lifecycle.HostKind) (ItemFunc, error) {
	var request api.Spec
	if timeout := strings.TrimSpace(flags.Timeout); timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			return nil, fmt.Errorf("--timeout %q: %w", timeout, err)
		}
		request.Budget.Timeout = timeout
	}

	// A selection routinely spans workspaces and each one has its own lifecycle
	// definition, so the host is per-directory rather than per-batch. They are
	// cached because building one reads and resolves the workspace's prompts.
	var mu sync.Mutex
	hosts := map[string]*lifecycle.Host{}
	hostFor := func(provider todos.Provider, dir string) (*lifecycle.Host, error) {
		mu.Lock()
		defer mu.Unlock()
		if host, ok := hosts[dir]; ok {
			return host, nil
		}
		host, err := lifecycle.NewHost(provider, dir, kind)
		if err != nil {
			return nil, err
		}
		hosts[dir] = host
		return host, nil
	}

	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		dir := strings.TrimSpace(todo.CWD)
		if dir == "" {
			return ItemResult{}, fmt.Errorf("TODO has no owning workspace path to verify against")
		}
		host, err := hostFor(provider, dir)
		if err != nil {
			return ItemResult{}, err
		}
		result := todos.CheckTODO(ctx, todo, todos.CheckOptions{
			Runner:  host,
			Request: request,
			Logger:  logger.StandardLogger(),
		})
		if result.Error != nil {
			return ItemResult{}, result.Error
		}
		if !result.AllPassed {
			if result.ErrorText != "" {
				return ItemResult{}, fmt.Errorf("verification failed: %s", result.ErrorText)
			}
			return ItemResult{}, fmt.Errorf("verification failed")
		}
		return ItemResult{Status: string(types.StatusVerified)}, nil
	}, nil
}

// ReopenFlags are the parameters of the `reopen` bulk action.
type ReopenFlags struct {
	Comment string `flag:"comment" help:"Comment recorded on each TODO while reopening it"`
}

func (ReopenFlags) ClickyActionFlags() {}

// Reopen returns finished TODOs to pending. It writes the status rather than
// clearing run history: the history is what the next run reads to know the work
// was attempted before.
func Reopen(flags ReopenFlags) (ItemFunc, error) {
	comment := strings.TrimSpace(flags.Comment)
	pending := types.StatusPending
	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if err := provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending}); err != nil {
			return ItemResult{}, err
		}
		if comment != "" {
			if err := provider.Comment(ctx, todo, comment); err != nil {
				return ItemResult{}, err
			}
		}
		return ItemResult{Status: string(pending)}, nil
	}, nil
}
