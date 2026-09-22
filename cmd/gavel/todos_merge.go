package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/merge"
	"github.com/spf13/cobra"
)

type TodosMergeOptions struct {
	TodoTargetOptions
	Into   string `flag:"into" help:"Short id of the TODO to merge into; defaults to the first id given"`
	DryRun bool   `flag:"dry-run" help:"Print the proposed merge without writing anything"`
	Model  string `flag:"model" help:"Override the model for this merge, as the compact mode:model:effort form"`
	Effort string `flag:"effort" help:"Reasoning effort" enum:"low,medium,high,xhigh"`
}

var todosMergeCmd *cobra.Command

func init() {
	todosMergeCmd = clicky.AddNamedCommand("merge", todosCmd, TodosMergeOptions{}, func(opts TodosMergeOptions) (any, error) {
		return nil, runTodosMerge(opts)
	})
	todosMergeCmd.Use = "merge <id> <id>..."
	todosMergeCmd.Short = "Combine several TODOs into one using AI"
	todosMergeCmd.Long = "Fold several TODOs into one: an AI pass writes the combined title, body, " +
		"verification fixture and plan onto the survivor, and the others are retired into it with a link " +
		"and the rationale recorded. Use --dry-run to see the proposal first."
}

func runTodosMerge(opts TodosMergeOptions) error {
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
	todoList, err := resolveRequestedTODOs(ctx, provider, workDir, args, todos.DiscoveryFilters{})
	if err != nil {
		return err
	}
	if len(todoList) < 2 {
		return fmt.Errorf("merge needs at least two TODOs; %v matched %d", args, len(todoList))
	}

	mergeOpts, err := merge.NewOptions(workDir, bulk.MergeFlags{
		Into: opts.Into, DryRun: opts.DryRun, Model: opts.Model, Effort: opts.Effort,
	})
	if err != nil {
		return err
	}
	result, proposal, err := merge.Run(ctx, bulk.TargetsFrom(provider, todoList), mergeOpts)
	if err != nil {
		return err
	}

	printMergeProposal(proposal, opts.DryRun)
	fmt.Println(result.String())
	for _, item := range result.Results {
		// The ref is whatever the caller addressed the TODO by — a full id from a
		// script — so the line is titled, which is what a person recognises.
		name := item.Title
		if name == "" {
			name = item.Ref
		}
		if item.Error != "" {
			fmt.Printf("  %s: %s\n", name, item.Error)
			continue
		}
		fmt.Printf("  %s: %s\n", name, item.Status)
	}
	if result.Failed > 0 {
		return fmt.Errorf("%d of %d TODOs failed to merge", result.Failed, len(result.Results))
	}
	return nil
}

// printMergeProposal shows what the merge produced. A dry run prints the whole
// artifact set — body, fixture and plan — because that is the only thing it
// produces; a live run prints the summary, since the TODO itself now holds the
// rest.
func printMergeProposal(proposal *merge.Proposal, dryRun bool) {
	if proposal == nil {
		return
	}
	fmt.Println(clicky.Text(proposal.Title, "text-bold").ANSI())
	if !dryRun {
		fmt.Printf("\n%s\n\n", strings.TrimSpace(proposal.Summary))
		return
	}
	fmt.Printf("\n%s\n", strings.TrimSpace(proposal.Body))
	if fixture := strings.TrimSpace(proposal.Verification); fixture != "" {
		fmt.Printf("\n## Verification\n\n%s\n", fixture)
	}
	if plan := strings.TrimSpace(proposal.Plan); plan != "" {
		fmt.Printf("\n## Plan\n\n%s\n", plan)
	}
	fmt.Printf("\n%s\n\n", strings.TrimSpace(proposal.Summary))
}
