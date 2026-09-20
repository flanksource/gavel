package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

type TodosEditOptions struct {
	TodoTargetOptions
	Title        string   `flag:"title" help:"New title"`
	Body         string   `flag:"body" help:"New body, path, or @path"`
	Plan         string   `flag:"plan" help:"Set or replace the TODO plan"`
	Verification string   `flag:"verification" help:"New verification fixture, path, or @path"`
	Status       string   `flag:"status" help:"New status"`
	Priority     string   `flag:"priority" help:"New priority"`
	Labels       []string `flag:"label" help:"Replace the TODO labels"`
	ClearLabels  bool     `flag:"clear-labels" help:"Remove every label from the TODO"`
}

type TodosCommentOptions struct {
	Parts []string `args:"true" required:"true"`
	Body  string   `flag:"body" help:"Comment body or @path"`
}

type TodosReopenOptions struct {
	TodoTargetOptions
	Comment     string `flag:"comment" help:"Comment to add while reopening"`
	CommentFile string `flag:"comment-file" help:"Read reopen comment from file"`
}

var todosEditCmd *cobra.Command

var todosCommentCmd *cobra.Command

var todosReopenCmd *cobra.Command

func init() {
	todosEditCmd = clicky.AddNamedCommand("edit", todosCmd, TodosEditOptions{}, func(opts TodosEditOptions) (any, error) { return nil, runTodosEdit(opts, changedTodoEditFlags()) })
	todosEditCmd.Short = "Edit a TODO's content, plan, status, and/or priority"

	todosCommentCmd = clicky.AddNamedCommand("comment", todosCmd, TodosCommentOptions{}, func(opts TodosCommentOptions) (any, error) {
		return nil, runTodosComment(opts, todosCommentCmd.Flags().Changed("body"))
	})
	todosCommentCmd.Short = "Add a comment to a TODO"

	todosReopenCmd = clicky.AddNamedCommand("reopen", todosCmd, TodosReopenOptions{}, func(opts TodosReopenOptions) (any, error) { return nil, runTodosReopen(opts) })
	todosReopenCmd.Short = "Reopen a completed TODO, optionally with a comment"
}

func changedTodoEditFlags() map[string]bool {
	changed := map[string]bool{}
	for _, name := range []string{"title", "body", "plan", "verification", "label"} {
		changed[name] = todosEditCmd.Flags().Changed(name)
	}
	return changed
}

func runTodosEdit(opts TodosEditOptions, changed map[string]bool) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return err
	}

	flags := todoEditFlags{Status: opts.Status, Priority: opts.Priority}
	if changed["title"] || opts.Title != "" {
		flags.Title = &opts.Title
	}
	if changed["body"] || opts.Body != "" {
		body, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: opts.Body})
		if err != nil {
			return err
		}
		flags.Body = &body
	}
	if changed["plan"] || opts.Plan != "" {
		plan, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--plan", Value: opts.Plan})
		if err != nil {
			return err
		}
		flags.Plan = &plan
	}
	if changed["verification"] || opts.Verification != "" {
		verification, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--verification", Value: opts.Verification})
		if err != nil {
			return err
		}
		flags.Verification = &verification
	}
	if changed["label"] || opts.Labels != nil {
		flags.Labels = &opts.Labels
	}
	flags.ClearLabels = opts.ClearLabels
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
	return printTodo(ctx, provider, ref, todo)
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
	// Labels is nil when --label was not passed. ClearLabels is --clear-labels;
	// the two are mutually exclusive because "replace with nothing" and "replace
	// with this list" cannot both be meant.
	Labels      *[]string
	ClearLabels bool
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
	if flags.ClearLabels && flags.Labels != nil {
		return changes, fmt.Errorf("--clear-labels cannot be combined with --label")
	}
	if flags.ClearLabels {
		changes.Content.Labels = &[]string{}
	} else if flags.Labels != nil {
		labelSet := make([]string, 0, len(*flags.Labels))
		for _, label := range *flags.Labels {
			if label = labels.Normalize(label); label != "" {
				labelSet = append(labelSet, label)
			}
		}
		changes.Content.Labels = &labelSet
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
		return changes, fmt.Errorf("nothing to edit: provide --title, --body, --plan, --verification, --status, --priority, --label, and/or --clear-labels")
	}
	return changes, nil
}

func runTodosComment(opts TodosCommentOptions, bodyChanged bool) error {
	if len(opts.Parts) == 0 {
		return fmt.Errorf("comment requires one TODO ID or alias")
	}
	ref, err := (TodoTargetOptions{IDs: opts.Parts[:1]}).One()
	if err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return err
	}

	body := strings.TrimSpace(strings.Join(opts.Parts[1:], " "))
	if bodyChanged || opts.Body != "" {
		flagBody, err := resolveTodoText(todoTextOptions{WorkDir: workDir, Flag: "--body", Value: opts.Body})
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
	return printTodo(ctx, provider, ref, todo)
}

func runTodosReopen(opts TodosReopenOptions) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return err
	}

	comment, hasComment, err := readOptionalComment(opts.Comment, opts.CommentFile, opts.Comment != "")
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
	return printTodo(ctx, provider, ref, todo)
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
