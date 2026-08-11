package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todoEditTitle        string
	todoEditBody         string
	todoEditPlan         string
	todoEditVerification string
	todoEditStatus       string
	todoEditPriority     string

	todoCommentBody string

	todoReopenComment     string
	todoReopenCommentFile string
)

var todosEditCmd = &cobra.Command{
	Use:          "edit <id-or-alias>",
	SilenceUsage: true,
	Short:        "Edit a TODO's content, plan, status, and/or priority",
	Example: `  gavel todos edit 3f2a1b --title "Fix parser panic"
  gavel todos edit 3f2a1b --body new-body.md
  gavel todos edit 3f2a1b --plan revised-plan.md --verification verification.md
  gavel todos edit 3f2a1b --priority low
  gavel todos edit 3f2a1b --status completed`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosEdit,
}

var todosCommentCmd = &cobra.Command{
	Use:          "comment <id-or-alias> [message...]",
	SilenceUsage: true,
	Short:        "Add a comment to a TODO",
	Example: `  gavel todos comment 3f2a1b "blocked on the upstream fix"
  gavel todos comment 3f2a1b --body @note.md`,
	Args: cobra.MinimumNArgs(1),
	RunE: runTodosComment,
}

var todosReopenCmd = &cobra.Command{
	Use:          "reopen <id-or-alias>",
	SilenceUsage: true,
	Short:        "Reopen a completed TODO, optionally with a comment",
	Example: `  gavel todos reopen 3f2a1b
  gavel todos reopen 3f2a1b --comment "regressed on main"`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosReopen,
}

func init() {
	todosCmd.AddCommand(todosEditCmd)
	todosEditCmd.Flags().StringVar(&todoEditTitle, "title", "", "New title")
	todosEditCmd.Flags().StringVar(&todoEditBody, "body", "", "New body, path, or @path")
	todosEditCmd.Flags().StringVar(&todoEditPlan, "plan", "", "Set or replace the todo's plan (markdown, path, or @path); the todo returns to review for approval")
	todosEditCmd.Flags().StringVar(&todoEditVerification, "verification", "", "New verification fixture, path, or @path")
	todosEditCmd.Flags().StringVar(&todoEditStatus, "status", "",
		"New status ("+joinStrings(types.AssignableStatuses())+")")
	todosEditCmd.Flags().StringVar(&todoEditPriority, "priority", "",
		"New priority ("+joinStrings(types.KnownPriorities())+")")

	todosCmd.AddCommand(todosCommentCmd)
	todosCommentCmd.Flags().StringVar(&todoCommentBody, "body", "", "Comment body or @path")

	todosCmd.AddCommand(todosReopenCmd)
	todosReopenCmd.Flags().StringVar(&todoReopenComment, "comment", "", "Comment to add while reopening")
	todosReopenCmd.Flags().StringVar(&todoReopenCommentFile, "comment-file", "", "Read reopen comment from file")
}

func runTodosEdit(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	flags := todoEditFlags{Status: todoEditStatus, Priority: todoEditPriority}
	if cmd.Flags().Changed("title") {
		flags.Title = &todoEditTitle
	}
	if cmd.Flags().Changed("body") {
		body, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: todoEditBody})
		if err != nil {
			return err
		}
		flags.Body = &body
	}
	if cmd.Flags().Changed("plan") {
		plan, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--plan", Value: todoEditPlan})
		if err != nil {
			return err
		}
		flags.Plan = &plan
	}
	if cmd.Flags().Changed("verification") {
		verification, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--verification", Value: todoEditVerification})
		if err != nil {
			return err
		}
		flags.Verification = &verification
	}
	changes, err := buildTodoEdit(flags)
	if err != nil {
		return err
	}
	var planRevisions todos.PlanRevisionProvider
	if changes.Plan != nil {
		var ok bool
		planRevisions, ok = provider.(todos.PlanRevisionProvider)
		if !ok {
			return fmt.Errorf("TODO provider does not support plan revisions")
		}
	}

	// Content first: Edit refreshes the TODO's optimistic-lock version, which
	// subsequent plan and state updates then reuse.
	if !changes.Content.IsEmpty() {
		if err := provider.Edit(ctx, todo, changes.Content); err != nil {
			return err
		}
	}
	if changes.Plan != nil {
		todo, err = planRevisions.SavePlanRevision(ctx, todo, *changes.Plan, "")
		if err != nil {
			return err
		}
		if todo == nil {
			return fmt.Errorf("plan revision provider returned no TODO")
		}
	}
	if changes.State.Status != nil || changes.State.Priority != nil {
		if err := provider.UpdateState(ctx, todo, changes.State); err != nil {
			return err
		}
	}
	return printTodo(ctx, provider, args[0], todo)
}

// todoEditFlags is the already-resolved (file references expanded) flag input to
// `todos edit`. Content pointers are nil when their flag was not set.
type todoEditFlags struct {
	Title        *string
	Body         *string
	Plan         *string
	Verification *string
	Status       string
	Priority     string
}

type todoEditChanges struct {
	Content todos.EditRequest
	Plan    *string
	State   todos.StateUpdate
}

// buildTodoEdit splits edit flags into content, plan, and state updates. It
// rejects statuses storage will not persist so callers see a failure rather
// than a silently declined write.
func buildTodoEdit(flags todoEditFlags) (todoEditChanges, error) {
	var changes todoEditChanges

	if flags.Title != nil {
		title := strings.TrimSpace(*flags.Title)
		if title == "" {
			return changes, fmt.Errorf("--title cannot be empty")
		}
		changes.Content.Title = &title
	}
	if flags.Body != nil {
		changes.Content.Body = flags.Body
	}
	if flags.Plan != nil {
		plan := strings.TrimSpace(*flags.Plan)
		if plan == "" {
			return changes, fmt.Errorf("--plan cannot be empty")
		}
		changes.Plan = &plan
	}
	if flags.Verification != nil {
		changes.Content.Verification = flags.Verification
	}
	if raw := strings.TrimSpace(flags.Status); raw != "" {
		status := types.Status(raw)
		if err := types.ValidateAssignableStatus(status); err != nil {
			return changes, err
		}
		changes.State.Status = &status
	}
	if raw := strings.TrimSpace(flags.Priority); raw != "" {
		priority := types.Priority(raw)
		if err := types.ValidatePriority(priority); err != nil {
			return changes, err
		}
		changes.State.Priority = &priority
	}

	if changes.Content.IsEmpty() && changes.Plan == nil && changes.State.Status == nil && changes.State.Priority == nil {
		return changes, fmt.Errorf("nothing to edit: provide --title, --body, --plan, --verification, --status, and/or --priority")
	}
	return changes, nil
}

func joinStrings[T ~string](values []T) string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, string(value))
	}
	return strings.Join(names, ", ")
}

func runTodosComment(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	body := strings.TrimSpace(strings.Join(args[1:], " "))
	if cmd.Flags().Changed("body") {
		flagBody, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: todoCommentBody})
		if err != nil {
			return err
		}
		body = flagBody
	}
	if body == "" {
		return fmt.Errorf("comment body is required (pass a message or --body)")
	}

	if err := provider.Comment(ctx, todo, body); err != nil {
		return err
	}
	return printTodo(ctx, provider, args[0], todo)
}

func runTodosReopen(_ *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	comment, hasComment, err := readOptionalComment(todoReopenComment, todoReopenCommentFile, todoReopenComment != "")
	if err != nil {
		return err
	}

	pending := types.StatusPending
	if err := provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending}); err != nil {
		return err
	}
	if hasComment {
		if err := provider.Comment(ctx, todo, comment); err != nil {
			return err
		}
	}
	return printTodo(ctx, provider, args[0], todo)
}

// readOptionalComment resolves comment text from an inline flag or a file. Inline and
// file are mutually exclusive. provided reports whether either source was set.
func readOptionalComment(inline, file string, inlineSet bool) (text string, provided bool, err error) {
	file = strings.TrimSpace(file)
	if inlineSet && file != "" {
		return "", false, fmt.Errorf("--comment and --comment-file are mutually exclusive")
	}
	if file != "" {
		raw, rerr := os.ReadFile(file)
		if rerr != nil {
			return "", false, fmt.Errorf("read comment file: %w", rerr)
		}
		return strings.TrimSpace(string(raw)), true, nil
	}
	if inlineSet {
		return inline, true, nil
	}
	return "", false, nil
}

// printTodo re-reads the TODO after a mutation so the printed detail reflects the
// provider's authoritative state (new body, comment event), falling back to the
// in-memory copy if the re-read fails.
func printTodo(ctx context.Context, provider todos.Provider, ref string, fallback *types.TODO) error {
	todo := fallback
	if refreshed, err := provider.Get(ctx, ref); err == nil {
		todo = refreshed
	}
	fmt.Println(todo.PrettyDetailed().ANSI())
	return nil
}
