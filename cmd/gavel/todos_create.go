package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todoCreateTitle        string
	todoCreateBody         string
	todoCreatePlan         string
	todoCreateVerification string
	todoCreatePriority     string
	todoCreateStatus       string
)

var todosCreateCmd = &cobra.Command{
	Use:          "create [title...]",
	Aliases:      []string{"new"},
	SilenceUsage: true,
	Short:        "Create a TODO",
	Long:         `Create a TODO from a title and optional body, plan, and verification fixture.`,
	Args:         cobra.ArbitraryArgs,
	RunE:         runTodosCreate,
}

func init() {
	todosCmd.AddCommand(todosCreateCmd)
	todosCreateCmd.Flags().StringVar(&todoCreateTitle, "title", "", "TODO title")
	todosCreateCmd.Flags().StringVar(&todoCreateBody, "body", "", "TODO body or @path")
	todosCreateCmd.Flags().StringVar(&todoCreatePlan, "plan", "", "Reviewed implementation plan or @path")
	todosCreateCmd.Flags().StringVar(&todoCreateVerification, "verification", "", "Verification fixture markdown or @path")
	todosCreateCmd.Flags().StringVar(&todoCreatePriority, "priority", string(types.PriorityMedium), "TODO priority: high, medium, or low")
	todosCreateCmd.Flags().StringVar(&todoCreateStatus, "status", string(types.StatusPending), "TODO status, or approved for a reviewed plan ready to run")
	todosCreateCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.ErrOrStderr(), todosCreateHelp(cmd).ANSI())
	})
}

func runTodosCreate(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	title := strings.TrimSpace(todoCreateTitle)
	if title == "" {
		title = strings.TrimSpace(strings.Join(args, " "))
	}
	if title == "" {
		return fmt.Errorf("title is required")
	}

	content, err := resolveTodoCreateContent(workDir, todoCreateContentOptions{
		Body: todoCreateBody, BodySet: cmd.Flags().Changed("body") || todoCreateBody != "",
		Plan: todoCreatePlan, PlanSet: cmd.Flags().Changed("plan") || todoCreatePlan != "",
		Verification:    todoCreateVerification,
		VerificationSet: cmd.Flags().Changed("verification") || todoCreateVerification != "",
	})
	if err != nil {
		return err
	}
	priority, err := parseTodoCreatePriority(todoCreatePriority)
	if err != nil {
		return err
	}
	lifecycle, err := parseTodoCreateLifecycle(todoCreateStatus)
	if err != nil {
		return err
	}
	if err := validateTodoCreatePlan(content.Plan, lifecycle); err != nil {
		return err
	}

	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}

	request := todos.CreateRequest{
		Title: title, Body: content.Body, Verification: content.Verification,
		Priority: priority, Status: lifecycle.Status,
	}
	if content.Plan != "" {
		request.Plan = &todos.CreatePlanRequest{Markdown: content.Plan, Approved: lifecycle.PlanApproved}
	}
	todo, err := provider.Create(context.Background(), request)
	if err != nil {
		return err
	}

	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}

func parseTodoCreatePriority(raw string) (types.Priority, error) {
	priority := types.Priority(strings.TrimSpace(raw))
	if priority == "" {
		return types.PriorityMedium, nil
	}
	switch priority {
	case types.PriorityHigh, types.PriorityMedium, types.PriorityLow:
		return priority, nil
	default:
		return "", fmt.Errorf("invalid --priority %q: expected high, medium, or low", raw)
	}
}

type todoCreateLifecycle struct {
	Status       types.Status
	PlanApproved bool
}

func parseTodoCreateLifecycle(raw string) (todoCreateLifecycle, error) {
	if strings.EqualFold(strings.TrimSpace(raw), "approved") {
		return todoCreateLifecycle{Status: types.StatusPending, PlanApproved: true}, nil
	}
	status := types.Status(strings.TrimSpace(raw))
	if status == "" {
		return todoCreateLifecycle{Status: types.StatusPending}, nil
	}
	if !types.IsKnownStatus(status) {
		known := make([]string, 0, len(types.KnownStatuses())+1)
		for _, candidate := range types.KnownStatuses() {
			known = append(known, string(candidate))
		}
		known = append(known, "approved")
		return todoCreateLifecycle{}, fmt.Errorf("invalid --status %q: expected %s", raw, strings.Join(known, ", "))
	}
	return todoCreateLifecycle{Status: status}, nil
}

func validateTodoCreatePlan(plan string, lifecycle todoCreateLifecycle) error {
	if lifecycle.PlanApproved && strings.TrimSpace(plan) == "" {
		return fmt.Errorf("--status approved requires --plan")
	}
	return nil
}
