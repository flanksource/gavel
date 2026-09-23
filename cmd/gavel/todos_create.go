package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/content"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/ops"
	"github.com/spf13/cobra"
)

type TodosCreateOptions struct {
	Titles       []string `args:"true"`
	Title        string   `flag:"title" help:"TODO title"`
	Body         string   `flag:"body" help:"TODO body or @path"`
	Plan         string   `flag:"plan" help:"Reviewed implementation plan or @path"`
	Verification string   `flag:"verification" help:"Verification fixture markdown or @path"`
	Labels       []string `flag:"label" help:"Attach a label"`
	Priority     string   `flag:"priority" default:"medium" help:"TODO priority"`
	Status       string   `flag:"status" default:"pending" help:"TODO status"`
	GitHub       bool     `flag:"github" help:"Push the new TODO to a GitHub issue"`
	BaseURL      string   `flag:"base-url" help:"Absolute origin attachment links resolve against"`
	Repo         string   `flag:"repo" help:"Target owner/repo"`
}

var todosCreateCmd *cobra.Command

func init() {
	todosCreateCmd = clicky.AddNamedCommand("create", todosCmd, TodosCreateOptions{}, func(opts TodosCreateOptions) (any, error) { return nil, runTodosCreate(opts) })
	todosCreateCmd.Aliases = []string{"new"}
	todosCreateCmd.Short = "Create a TODO"
	todosCreateCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.ErrOrStderr(), todosCreateHelp(cmd).ANSI())
	})
}

func runTodosCreate(opts TodosCreateOptions) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	title := strings.TrimSpace(opts.Title)
	if title == "" {
		title = strings.TrimSpace(strings.Join(opts.Titles, " "))
	}
	if title == "" {
		return fmt.Errorf("title is required")
	}

	resolved, err := content.ResolveCreate(workDir, content.CreateOptions{
		Body: opts.Body, BodySet: todosCreateCmd.Flags().Changed("body") || opts.Body != "",
		Plan: opts.Plan, PlanSet: todosCreateCmd.Flags().Changed("plan") || opts.Plan != "",
		Verification:    opts.Verification,
		VerificationSet: todosCreateCmd.Flags().Changed("verification") || opts.Verification != "",
	})
	if err != nil {
		return err
	}
	priority, err := ops.ParsePriority(opts.Priority)
	if err != nil {
		return err
	}
	lifecycle, err := ops.ParseLifecycle(opts.Status)
	if err != nil {
		return err
	}
	if err := ops.ValidateCreatePlan(resolved.Plan, lifecycle); err != nil {
		return err
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}

	request := todos.CreateRequest{
		Title: title, Body: resolved.Body, Verification: resolved.Verification,
		Priority: priority, Status: lifecycle.Status, Labels: opts.Labels,
	}
	if resolved.Plan != "" {
		request.Plan = &todos.CreatePlanRequest{Markdown: resolved.Plan, Approved: lifecycle.PlanApproved}
	}
	todo, err := provider.Create(context.Background(), request)
	if err != nil {
		return err
	}

	// Print before pushing: the TODO exists either way, so a push failure must
	// not hide the record it was created against.
	fmt.Println(todo.PrettyDetailed().ANSI())
	if !opts.GitHub {
		return nil
	}
	baseURL, err := resolveTodoPushBaseURL(workDir, opts.BaseURL)
	if err != nil {
		return err
	}
	result, err := githubpush.Push(context.Background(), provider, todo.ID, githubpush.Options{
		GitHub:  github.Options{WorkDir: workDir, Repo: opts.Repo},
		BaseURL: baseURL,
		Labels:  true,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%s → %s\n", todo.DisplayID(), result.URL)
	return nil
}
