package main

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

type TodosLinkOptions struct {
	TodoTargetOptions
	Relation string `flag:"relation" default:"related-to" help:"Relation to create or remove"`
}

var todosLinkCmd *cobra.Command
var todosUnlinkCmd *cobra.Command
var todosLinksCmd *cobra.Command

func init() {
	todosLinkCmd = clicky.AddNamedCommand("link", todosCmd, TodosLinkOptions{}, func(opts TodosLinkOptions) (any, error) { return nil, runTodosLink(opts) })
	todosLinkCmd.Short = "Link two TODOs as related or dependent"
	todosUnlinkCmd = clicky.AddNamedCommand("unlink", todosCmd, TodosLinkOptions{}, func(opts TodosLinkOptions) (any, error) { return nil, runTodosUnlink(opts) })
	todosUnlinkCmd.Short = "Remove a link between two TODOs"
	todosLinksCmd = clicky.AddNamedCommand("links", todosCmd, TodosGetOptions{}, func(opts TodosGetOptions) (any, error) { return nil, runTodosLinks(opts) })
	todosLinksCmd.Short = "List a TODO's links"
}

func runTodosLink(opts TodosLinkOptions) error {
	args, err := opts.Many()
	if err != nil {
		return err
	}
	if len(args) != 2 {
		return fmt.Errorf("link requires a source and target TODO ID")
	}
	relation, err := types.ParseRelationKind(opts.Relation)
	if err != nil {
		return err
	}
	linker, todo, err := openTodoLinker(args[0])
	if err != nil {
		return err
	}
	link, err := linker.Link(context.Background(), todo, args[1], relation)
	if err != nil {
		return err
	}
	fmt.Printf("%s %s %s (%s)\n", todo.ShortID, link.Relation, link.TargetShortID, link.TargetTitle)
	return nil
}

func runTodosUnlink(opts TodosLinkOptions) error {
	args, err := opts.Many()
	if err != nil {
		return err
	}
	if len(args) != 2 {
		return fmt.Errorf("unlink requires a source and target TODO ID")
	}
	relation, err := types.ParseRelationKind(opts.Relation)
	if err != nil {
		return err
	}
	linker, todo, err := openTodoLinker(args[0])
	if err != nil {
		return err
	}
	if err := linker.Unlink(context.Background(), todo, args[1], relation); err != nil {
		return err
	}
	fmt.Printf("Removed %s %s %s\n", todo.ShortID, relation, args[1])
	return nil
}

func runTodosLinks(opts TodosGetOptions) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	linker, todo, err := openTodoLinker(ref)
	if err != nil {
		return err
	}
	links, err := linker.Links(context.Background(), todo)
	if err != nil {
		return err
	}
	if len(links) == 0 {
		fmt.Printf("%s has no links\n", todo.ShortID)
		return nil
	}
	rendered, err := clicky.Format(links)
	if err != nil {
		return err
	}
	fmt.Println(rendered)
	return nil
}

// openTodoLinker opens the workspace provider, asserts relationship support,
// and loads the source TODO.
func openTodoLinker(ref string) (todos.RelationshipProvider, *types.TODO, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	provider, err := newTodosProvider(workDir)
	if err != nil {
		return nil, nil, err
	}
	linker, ok := provider.(todos.RelationshipProvider)
	if !ok {
		return nil, nil, fmt.Errorf("TODO provider does not support links; native PostgreSQL storage is required")
	}
	todo, err := provider.Get(context.Background(), ref)
	if err != nil {
		return nil, nil, err
	}
	return linker, todo, nil
}
