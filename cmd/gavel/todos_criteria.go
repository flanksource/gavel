package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var todosCriteriaCmd = &cobra.Command{
	Use:          "criteria",
	SilenceUsage: true,
	Short:        "View and edit a TODO's acceptance criteria",
}

var todosCriteriaListCmd = &cobra.Command{
	Use:          "list <id-or-file>",
	SilenceUsage: true,
	Short:        "List a TODO's acceptance criteria",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosCriteriaList,
}

var todosCriteriaAddCmd = &cobra.Command{
	Use:          "add <id-or-file> <criterion...>",
	SilenceUsage: true,
	Short:        "Add a custom acceptance criterion",
	Args:         cobra.MinimumNArgs(2),
	RunE:         runTodosCriteriaAdd,
}

var todosCriteriaRemoveCmd = &cobra.Command{
	Use:          "remove <id-or-file> <number>",
	SilenceUsage: true,
	Short:        "Remove an acceptance criterion by its list number",
	Args:         cobra.ExactArgs(2),
	RunE:         runTodosCriteriaRemove,
}

var todosCriteriaEditCmd = &cobra.Command{
	Use:          "edit <id-or-file> <number> <criterion...>",
	SilenceUsage: true,
	Short:        "Replace an acceptance criterion's text by its list number",
	Args:         cobra.MinimumNArgs(3),
	RunE:         runTodosCriteriaEdit,
}

var todosCriteriaGenerateCmd = &cobra.Command{
	Use:          "generate <id-or-file>",
	SilenceUsage: true,
	Short:        "(Re)draft acceptance criteria for a TODO using an AI model",
	Args:         cobra.ExactArgs(1),
	RunE:         runTodosCriteriaGenerate,
}

func init() {
	todosCmd.AddCommand(todosCriteriaCmd)
	for _, c := range []*cobra.Command{
		todosCriteriaListCmd, todosCriteriaAddCmd, todosCriteriaRemoveCmd,
		todosCriteriaEditCmd, todosCriteriaGenerateCmd,
	} {
		c.Flags().StringVar(&todosDir, "dir", "", "TODOs directory (default: .todos)")
		todosCriteriaCmd.AddCommand(c)
	}
	todosCriteriaGenerateCmd.Flags().StringVar(&todoCriteriaModel, "model", "", "LLM model for acceptance-criteria generation")
}

func loadCriteriaTODO(ctx context.Context, ref string) (todos.Provider, *types.TODO, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir, todosDir)
	if err != nil {
		return nil, nil, err
	}
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	return provider, todo, nil
}

// saveCriteria rewrites the TODO body's acceptance-criteria section and prints
// the refreshed criteria list.
func saveCriteria(ctx context.Context, provider todos.Provider, todo *types.TODO, criteria []types.AcceptanceCriterion) error {
	body := todos.UpsertCriteriaSection(todo.MarkdownBody, criteria)
	if err := provider.Edit(ctx, todo, todos.EditRequest{Body: &body}); err != nil {
		return err
	}
	fmt.Println(prettyCriteria(criteria).ANSI())
	return nil
}

func runTodosCriteriaList(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	_, todo, err := loadCriteriaTODO(ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Println(prettyCriteria(todo.AcceptanceCriteria).ANSI())
	return nil
}

func runTodosCriteriaAdd(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	provider, todo, err := loadCriteriaTODO(ctx, args[0])
	if err != nil {
		return err
	}
	text := strings.TrimSpace(strings.Join(args[1:], " "))
	if text == "" {
		return fmt.Errorf("criterion text is required")
	}
	criteria := append(todo.AcceptanceCriteria, types.AcceptanceCriterion{Text: text})
	return saveCriteria(ctx, provider, todo, criteria)
}

func runTodosCriteriaRemove(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	provider, todo, err := loadCriteriaTODO(ctx, args[0])
	if err != nil {
		return err
	}
	idx, err := criterionIndex(args[1], len(todo.AcceptanceCriteria))
	if err != nil {
		return err
	}
	criteria := append(todo.AcceptanceCriteria[:idx], todo.AcceptanceCriteria[idx+1:]...)
	return saveCriteria(ctx, provider, todo, criteria)
}

func runTodosCriteriaEdit(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	provider, todo, err := loadCriteriaTODO(ctx, args[0])
	if err != nil {
		return err
	}
	idx, err := criterionIndex(args[1], len(todo.AcceptanceCriteria))
	if err != nil {
		return err
	}
	text := strings.TrimSpace(strings.Join(args[2:], " "))
	if text == "" {
		return fmt.Errorf("criterion text is required")
	}
	criteria := todo.AcceptanceCriteria
	criteria[idx].Text = text
	criteria[idx].CheckID = "" // editing the text makes it a custom criterion
	return saveCriteria(ctx, provider, todo, criteria)
}

func runTodosCriteriaGenerate(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	provider, todo, err := loadCriteriaTODO(ctx, args[0])
	if err != nil {
		return err
	}
	agent, err := commitpkg.BuildAgent(commitpkg.Options{}, todoCriteriaModel)
	if err != nil {
		return err
	}
	genCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	criteria, err := todos.Generate(genCtx, agent, todo.Title, todo.MarkdownBody)
	if err != nil {
		return err
	}
	if len(criteria) == 0 {
		return fmt.Errorf("model returned no acceptance criteria")
	}
	return saveCriteria(ctx, provider, todo, criteria)
}

func criterionIndex(arg string, count int) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || n < 1 || n > count {
		return 0, fmt.Errorf("invalid criterion number %q (have %d)", arg, count)
	}
	return n - 1, nil
}

func prettyCriteria(criteria []types.AcceptanceCriterion) api.Text {
	if len(criteria) == 0 {
		return clicky.Text("No acceptance criteria", "text-gray-500")
	}
	text := clicky.Text("Acceptance Criteria", "text-blue-600 font-bold")
	for i, c := range criteria {
		box := "[ ]"
		if c.Done {
			box = "[x]"
		}
		label := c.Text
		if c.CheckID != "" {
			label = c.CheckID + ": " + c.Text
		}
		text = text.NewLine().Append(fmt.Sprintf("  %d. %s %s", i+1, box, label), "")
	}
	return text
}
