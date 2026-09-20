package main

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

type TodosPushOptions struct {
	TodoTargetOptions
	BaseURL string `flag:"base-url" help:"Absolute origin attachment links resolve against"`
	Repo    string `flag:"repo" help:"Target owner/repo"`
	Force   bool   `flag:"force" help:"Open a second issue for an already linked TODO"`
	Update  bool   `flag:"update" help:"Rewrite the linked issue"`
	Issue   string `flag:"issue" help:"Rewrite this issue and link it"`
	Labels  bool   `flag:"labels" default:"true" help:"Copy TODO labels onto the issue"`
	Plan    bool   `flag:"plan" default:"true" help:"Include the TODO plan in the issue body"`
}

var todosPushCmd *cobra.Command

func init() {
	todosPushCmd = clicky.AddNamedCommand("push", todosCmd, TodosPushOptions{}, func(opts TodosPushOptions) (any, error) { return nil, runTodosPush(opts) })
	todosPushCmd.Short = "Push TODOs to GitHub issues"
}

func runTodosPush(opts TodosPushOptions) error {
	args, err := opts.Many()
	if err != nil {
		return err
	}
	ctx := context.Background()
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	todoList, err := resolveRequestedTODOs(ctx, provider, args, todos.DiscoveryFilters{})
	if err != nil {
		return err
	}
	if len(todoList) == 0 {
		return fmt.Errorf("no TODOs matched %v", args)
	}

	baseURL, err := resolveTodoPushBaseURL(workDir, opts.BaseURL)
	if err != nil {
		return err
	}

	if opts.Issue != "" && len(todoList) > 1 {
		return fmt.Errorf("--issue names a single issue but %d TODOs matched %v", len(todoList), args)
	}

	pushOpts := githubpush.Options{
		GitHub:  github.Options{WorkDir: workDir, Repo: opts.Repo},
		BaseURL: baseURL,
		Force:   opts.Force,
		Update:  opts.Update,
		Issue:   opts.Issue,
		Labels:  opts.Labels,
		Plan:    opts.Plan,
	}
	for _, todo := range todoList {
		result, err := githubpush.Push(ctx, provider, todo.ID, pushOpts)
		if err != nil {
			return err
		}
		verb := "opened"
		if result.Updated {
			verb = "updated"
		}
		fmt.Printf("%s → %s (%s)\n", todo.DisplayID(), result.URL, verb)
	}
	return nil
}

// resolveTodoPushBaseURL prefers the flag over the project config. Resolving it
// before any TODO is pushed keeps a malformed origin from opening some issues
// and then failing.
func resolveTodoPushBaseURL(workDir, requested string) (string, error) {
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml: %w", err)
	}
	baseURL, err := githubpush.ResolveBaseURL(requested, config.Todos.BaseURL)
	if err != nil {
		return "", err
	}
	if githubpush.IsLoopback(baseURL) {
		logger.Warnf("base URL %s is loopback: attachment images will only render for viewers on this machine", baseURL)
	}
	return baseURL, nil
}
