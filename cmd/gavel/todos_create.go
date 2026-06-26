package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todoCreateTitle    string
	todoCreateBody     string
	todoCreateBodyFile string
	todoCreatePriority string
	todoCreateStatus   string
	todoCreateCriteria bool
	todoCriteriaModel  string
)

var todosCreateCmd = &cobra.Command{
	Use:          "create [title...]",
	Aliases:      []string{"new"},
	SilenceUsage: true,
	Short:        "Create a TODO",
	Args:         cobra.ArbitraryArgs,
	RunE:         runTodosCreate,
}

func init() {
	todosCmd.AddCommand(todosCreateCmd)
	todosCreateCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosCreateCmd.Flags().StringVar(&todoCreateTitle, "title", "", "TODO title")
	todosCreateCmd.Flags().StringVar(&todoCreateBody, "body", "", "TODO body")
	todosCreateCmd.Flags().StringVar(&todoCreateBodyFile, "body-file", "", "Read TODO body from file")
	todosCreateCmd.Flags().StringVar(&todoCreatePriority, "priority", string(types.PriorityMedium), "TODO priority: high, medium, or low")
	todosCreateCmd.Flags().StringVar(&todoCreateStatus, "status", string(types.StatusPending), "TODO status: draft, pending, in_progress, failed, verified, completed, or skipped")
	todosCreateCmd.Flags().BoolVar(&todoCreateCriteria, "criteria", true, "Draft AI acceptance criteria from the title/body when a model is available (use --criteria=false to skip)")
	todosCreateCmd.Flags().StringVar(&todoCriteriaModel, "model", "", "LLM model for acceptance-criteria generation")
}

func runTodosCreate(_ *cobra.Command, args []string) error {
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

	body, err := todoCreateBodyText()
	if err != nil {
		return err
	}
	priority, err := parseTodoCreatePriority(todoCreatePriority)
	if err != nil {
		return err
	}
	status, err := parseTodoCreateStatus(todoCreateStatus)
	if err != nil {
		return err
	}

	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}

	if todoCreateCriteria {
		body = withGeneratedCriteria(title, body)
	}

	todo, err := provider.Create(context.Background(), todos.CreateRequest{
		Title:    title,
		Body:     body,
		Priority: priority,
		Status:   status,
	})
	if err != nil {
		return err
	}

	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}

// withGeneratedCriteria drafts acceptance criteria for a new TODO and folds them
// into its body. Generation is best-effort: when no model/key is available it
// logs and returns the body unchanged rather than blocking creation.
func withGeneratedCriteria(title, body string) string {
	if strings.Contains(body, "## "+"Acceptance Criteria") {
		return body
	}
	agent, err := commitpkg.BuildAgent(commitpkg.Options{}, todoCriteriaModel)
	if err != nil {
		logger.Warnf("acceptance-criteria generation skipped: %v", err)
		return body
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	criteria, err := todos.Generate(ctx, agent, title, body)
	if err != nil {
		logger.Warnf("acceptance-criteria generation failed: %v", err)
		return body
	}
	if len(criteria) == 0 {
		return body
	}
	logger.Infof("Drafted %d acceptance criteria (edit with `gavel todos criteria`)", len(criteria))
	return todos.UpsertCriteriaSection(body, criteria)
}

func todoCreateBodyText() (string, error) {
	body := strings.TrimSpace(todoCreateBody)
	bodyFile := strings.TrimSpace(todoCreateBodyFile)
	if body != "" && bodyFile != "" {
		return "", fmt.Errorf("--body and --body-file are mutually exclusive")
	}
	if bodyFile == "" {
		return body, nil
	}
	raw, err := os.ReadFile(bodyFile)
	if err != nil {
		return "", fmt.Errorf("read --body-file: %w", err)
	}
	return strings.TrimSpace(string(raw)), nil
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

func parseTodoCreateStatus(raw string) (types.Status, error) {
	status := types.Status(strings.TrimSpace(raw))
	if status == "" {
		return types.StatusPending, nil
	}
	if !types.IsKnownStatus(status) {
		return "", fmt.Errorf("invalid --status %q: expected draft, pending, in_progress, failed, verified, completed, or skipped", raw)
	}
	return status, nil
}
