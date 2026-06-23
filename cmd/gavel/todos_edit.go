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
	todoEditTitle    string
	todoEditBody     string
	todoEditBodyFile string

	todoCommentBody     string
	todoCommentBodyFile string

	todoReopenComment     string
	todoReopenCommentFile string
)

var todosEditCmd = &cobra.Command{
	Use:          "edit <id-or-file>",
	SilenceUsage: true,
	Short:        "Edit a TODO's title and/or body",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosEdit,
}

var todosCommentCmd = &cobra.Command{
	Use:          "comment <id-or-file> [message...]",
	SilenceUsage: true,
	Short:        "Add a comment to a TODO",
	Args:         cobra.MinimumNArgs(1),
	RunE:         runTodosComment,
}

var todosReopenCmd = &cobra.Command{
	Use:          "reopen <id-or-file>",
	SilenceUsage: true,
	Short:        "Reopen a completed TODO, optionally with a comment",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosReopen,
}

func init() {
	todosCmd.AddCommand(todosEditCmd)
	todosEditCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosEditCmd.Flags().StringVar(&todoEditTitle, "title", "", "New title")
	todosEditCmd.Flags().StringVar(&todoEditBody, "body", "", "New body")
	todosEditCmd.Flags().StringVar(&todoEditBodyFile, "body-file", "", "Read new body from file")

	todosCmd.AddCommand(todosCommentCmd)
	todosCommentCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosCommentCmd.Flags().StringVar(&todoCommentBody, "body", "", "Comment body")
	todosCommentCmd.Flags().StringVar(&todoCommentBodyFile, "body-file", "", "Read comment body from file")

	todosCmd.AddCommand(todosReopenCmd)
	todosReopenCmd.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
	todosReopenCmd.Flags().StringVar(&todoReopenComment, "comment", "", "Comment to add while reopening")
	todosReopenCmd.Flags().StringVar(&todoReopenCommentFile, "comment-file", "", "Read reopen comment from file")
}

func runTodosEdit(cmd *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	var edit todos.EditRequest
	if cmd.Flags().Changed("title") {
		title := strings.TrimSpace(todoEditTitle)
		if title == "" {
			return fmt.Errorf("--title cannot be empty")
		}
		edit.Title = &title
	}
	body, bodyProvided, err := readOptionalBody(todoEditBody, todoEditBodyFile, cmd.Flags().Changed("body"))
	if err != nil {
		return err
	}
	if bodyProvided {
		edit.Body = &body
	}
	if edit.IsEmpty() {
		return fmt.Errorf("nothing to edit: provide --title and/or --body")
	}

	if err := provider.Edit(ctx, todo, edit); err != nil {
		return err
	}
	return printTodo(ctx, provider, args[0], todo)
}

func runTodosComment(_ *cobra.Command, args []string) error {
	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	body := strings.TrimSpace(strings.Join(args[1:], " "))
	flagBody, flagProvided, err := readOptionalBody(todoCommentBody, todoCommentBodyFile, todoCommentBody != "")
	if err != nil {
		return err
	}
	if flagProvided {
		body = flagBody
	}
	if body == "" {
		return fmt.Errorf("comment body is required (pass a message, --body, or --body-file)")
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
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return err
	}
	ctx := context.Background()
	todo, err := provider.Get(ctx, args[0])
	if err != nil {
		return err
	}

	comment, hasComment, err := readOptionalBody(todoReopenComment, todoReopenCommentFile, todoReopenComment != "")
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

// readOptionalBody resolves body text from an inline flag or a file. Inline and
// file are mutually exclusive. provided reports whether either source was set.
func readOptionalBody(inline, file string, inlineSet bool) (text string, provided bool, err error) {
	file = strings.TrimSpace(file)
	if inlineSet && file != "" {
		return "", false, fmt.Errorf("--body and --body-file are mutually exclusive")
	}
	if file != "" {
		raw, rerr := os.ReadFile(file)
		if rerr != nil {
			return "", false, fmt.Errorf("read body file: %w", rerr)
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
