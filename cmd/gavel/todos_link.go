package main

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/spf13/cobra"
)

var (
	todoLinkRelation   string
	todoUnlinkRelation string
)

var todosLinkCmd = &cobra.Command{
	Use:          "link <ref> <target-ref>",
	SilenceUsage: true,
	Short:        "Link two TODOs as related or dependent",
	Long: `Record a relationship between two TODOs in the current workspace.

  related-to   symmetric and non-blocking — duplicates, overlapping scope, or
               work that should be read together (the default)
  depends-on   <ref> is blocked until <target-ref> is verified or completed

Dependency cycles, cross-workspace links, and duplicates are rejected.`,
	Example: `  gavel todos link 3f2a1b 7c4d9e
  gavel todos link 3f2a1b 7c4d9e --relation depends-on`,
	Args: cobra.ExactArgs(2),
	RunE: runTodosLink,
}

var todosUnlinkCmd = &cobra.Command{
	Use:          "unlink <ref> <target-ref>",
	SilenceUsage: true,
	Short:        "Remove a link between two TODOs",
	Example: `  gavel todos unlink 3f2a1b 7c4d9e
  gavel todos unlink 3f2a1b 7c4d9e --relation depends-on`,
	Args: cobra.ExactArgs(2),
	RunE: runTodosUnlink,
}

var todosLinksCmd = &cobra.Command{
	Use:          "links <ref>",
	SilenceUsage: true,
	Short:        "List a TODO's links",
	Long: `List every relationship touching a TODO from its own perspective. Incoming
dependencies are reported as the derived read-only "blocks" relation.`,
	Example: `  gavel todos links 3f2a1b
  gavel todos links 3f2a1b --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosLinks,
}

func init() {
	todosCmd.AddCommand(todosLinkCmd)
	todosLinkCmd.Flags().StringVar(&todoLinkRelation, "relation", string(types.RelationRelatedTo),
		"Relation to create ("+joinStrings(types.LinkableRelations())+")")

	todosCmd.AddCommand(todosUnlinkCmd)
	todosUnlinkCmd.Flags().StringVar(&todoUnlinkRelation, "relation", string(types.RelationRelatedTo),
		"Relation to remove ("+joinStrings(types.LinkableRelations())+")")

	todosCmd.AddCommand(todosLinksCmd)
}

func runTodosLink(_ *cobra.Command, args []string) error {
	relation, err := types.ParseRelationKind(todoLinkRelation)
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

func runTodosUnlink(_ *cobra.Command, args []string) error {
	relation, err := types.ParseRelationKind(todoUnlinkRelation)
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

func runTodosLinks(_ *cobra.Command, args []string) error {
	linker, todo, err := openTodoLinker(args[0])
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
