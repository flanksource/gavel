package main

import (
	"fmt"
	"strings"

	"github.com/flanksource/commons/duration"
	"github.com/spf13/cobra"
)

var todosCmd = &cobra.Command{
	Use:          "todos",
	Aliases:      []string{"todo"},
	SilenceUsage: true,
	Args:         cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
	Short: "Automated TODO execution and fixture-backed verification",
	Long: `Run, check, and manage TODOs — units of work an AI coding agent implements
and gavel verifies with their persisted definition-of-done fixture.

A TODO is a PostgreSQL-backed issue with a title, body, status, priority,
acceptance criteria, verification fixtures, and execution history. Todos come
from source TODO/FIXME comments ('todos sync'), explicit portable imports,
or by hand ('todos create').

Every todo moves through the project's lifecycle: an ordered set of steps —
triage, plan, verify, run by default — each a captain prompt plus the CEL
predicates that decide when it applies and which status its result lands.
'todos steps' lists them; 'todos run --step' names one.

Subcommands:
  list      List TODOs (filter by --status, group with --group-by)
  get       Show one TODO in detail (accepts a short id, full id, title, or alias)
  create    Create a TODO
  run       Run the next lifecycle step for TODOs, or the one named with --step
  steps     List the lifecycle's steps, or where one TODO stands in it
  check     Run a TODO's fixture-backed definition of done (the verify step)
  push      Open a GitHub issue for a TODO and link the two
  edit / comment / reopen / criteria / sync / plan / transfer

Examples:
  gavel todos list
  gavel todos list --all            # list every registered project
  gavel todos list --all --done     # include verified/completed items
  gavel todos get <id>
  gavel todos run <id>              # run the next step for one TODO
  gavel todos run <id> --step plan  # propose a reviewable plan first
  gavel todos steps <id>            # which steps apply to this todo now
  gavel todos check <id>            # run the issue's definition of done`,
}

var todosRunCmd *cobra.Command

type TodosListOptions struct {
	Status  string `json:"status" flag:"status" help:"Filter TODOs by status"`
	All     bool   `json:"all" flag:"all" help:"List PostgreSQL-backed TODOs from all registered projects"`
	Done    bool   `json:"done" flag:"done" help:"Include verified and completed TODOs"`
	Since   string `json:"since" flag:"since" help:"Show TODOs created or updated since (e.g. 7d, now-30d, 2024-01-01)"`
	GroupBy string `json:"group-by" flag:"group-by" help:"Group TODOs by: file, directory, repo, all, or none"`
}

func (opts TodosListOptions) GetName() string { return "list" }

var todosGetCmd *cobra.Command
var todosCheckCmd *cobra.Command

type TodoTargetOptions struct {
	IDs []string `args:"true" required:"true"`
}

func (o TodoTargetOptions) One() (string, error) {
	if len(o.IDs) != 1 || strings.TrimSpace(o.IDs[0]) == "" {
		return "", fmt.Errorf("exactly one TODO ID or alias is required")
	}
	return strings.TrimSpace(o.IDs[0]), nil
}

func (o TodoTargetOptions) Many() ([]string, error) {
	if len(o.IDs) == 0 {
		return nil, fmt.Errorf("at least one TODO ID or alias is required")
	}
	ids := make([]string, len(o.IDs))
	seen := make(map[string]bool, len(o.IDs))
	for i, id := range o.IDs {
		ids[i] = strings.TrimSpace(id)
		if ids[i] == "" {
			return nil, fmt.Errorf("TODO ID %d is blank", i+1)
		}
		if seen[ids[i]] {
			return nil, fmt.Errorf("TODO ID %q was specified more than once", ids[i])
		}
		seen[ids[i]] = true
	}
	return ids, nil
}

type TodosGetOptions struct {
	TodoTargetOptions
}

type TodosCheckOptions struct {
	TodoTargetOptions
	Timeout     duration.Duration `flag:"timeout" help:"Test execution timeout"`
	Concurrency int               `flag:"concurrency" help:"How many TODOs to check at once (0 uses project configuration)"`
}
