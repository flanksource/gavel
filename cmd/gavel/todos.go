package main

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	filterStatus     string
	checkTimeout     time.Duration
	checkConcurrency int
	maxBudget        float64
	maxTurns         int
	interactive      bool
	dirty            bool
	dryRun           bool
	commitAfter      bool
	// todosStep names the lifecycle step `todos run` dispatches. Empty lets the
	// lifecycle pick the step each todo needs next.
	todosStep     string
	todoModel     string
	todoEffort    string
	resumeSession bool
	// forceRun answers the "this todo is already running" question up front, so
	// an unattended run can dispatch alongside the live one without a prompt.
	forceRun bool
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
  gavel todos run                   # run the next step of every pending todo
  gavel todos run --step plan       # propose a reviewable plan first
  gavel todos steps <id>            # which steps apply to this todo now
  gavel todos check <id>            # run the issue's definition of done`,
}

var todosRunCmd = &cobra.Command{
	Use:          "run [todo-titles...]",
	SilenceUsage: true,
	Short:        "Run a lifecycle step for TODOs with a coding agent",
	Long: `Drive an AI coding agent (Claude or Codex, via cmux or headless) through the
project's todo lifecycle.

With no arguments it runs every pending TODO; pass titles, ids, or aliases to select
a subset, or -i to pick interactively. Each TODO runs ONE lifecycle step: the step
the lifecycle picks next for it — plan a todo that has no plan, verify one whose
implementation has not been checked, implement one that is ready — or the step
named with --step. 'gavel todos steps' lists the steps and, given a todo, which
apply to it now.

A step's outcome is what moves the todo: the lifecycle definition declares which
status each result lands, so the run itself never writes one. Implementation
steps commit their work through gavel commit (--commit, on by default). The
configured checks suite (.gavel.yaml checks.enabled, or a todo's own checks:
front matter) is part of the todo's definition of done: its failures feed back
to the agent until they pass. Use --dry-run to print the rendered prompt, the
layer stack that produced the run's spec, and the spec itself, without
dispatching.

Examples:
  gavel todos run                          # the next step of every pending todo
  gavel todos run "Fix flaky parser test"  # one todo by title
  gavel todos run -i                       # interactively select
  gavel todos run --step plan              # propose plans for review
  gavel todos run --step verify            # run the definition of done
  gavel todos run 3f2a1b --step triage     # a read-only triage pass
  gavel todos run --model cli:opus:high    # headless CLI on opus, high effort
  gavel todos run --dry-run                # preview the run, no changes`,
	RunE: runTodosRun,
}

type TodosListOptions struct {
	Status  string `json:"status" flag:"status" help:"Filter TODOs by status"`
	All     bool   `json:"all" flag:"all" help:"List PostgreSQL-backed TODOs from all registered projects"`
	Done    bool   `json:"done" flag:"done" help:"Include verified and completed TODOs"`
	Since   string `json:"since" flag:"since" help:"Show TODOs created or updated since (e.g. 7d, now-30d, 2024-01-01)"`
	GroupBy string `json:"group-by" flag:"group-by" help:"Group TODOs by: file, directory, repo, all, or none"`
}

func (opts TodosListOptions) GetName() string { return "list" }

var todosGetCmd = &cobra.Command{
	Use:          "get <id-or-alias>",
	SilenceUsage: true,
	Short:        "Display detailed information about a PostgreSQL-backed TODO",
	Long: `Show one TODO in full — metadata, body, acceptance criteria, and run history.

The argument matches a short id, a full id, the title, or an imported alias.

Examples:
  gavel todos get 3f2a1b

  gavel todos get "Fix flaky parser test"`,
	Args: cobra.ExactArgs(1),
	RunE: runTodosGet,
}

var todosCheckCmd = &cobra.Command{
	Use:          "check [ids-or-aliases...]",
	SilenceUsage: true,
	Short:        "Run TODOs' fixture-backed definitions of done",
	Long: `Run each TODO's complete definition of done and report pass/fail.

The check is the lifecycle's verify step, dispatched by name: configured
test/lint steps, the persisted Verification fixture, and the acceptance-criteria
AI checklist all run through the same gavel fixture/CEL pipeline as the
verification an implementation run performs, and the step's outcome is what
moves the todo to verified or unverified. Exits non-zero if any TODO fails.
With no arguments it checks every discovered TODO; pass ids or aliases to select some.

Examples:
  gavel todos check
  gavel todos check 3f2a1b
  gavel todos check --status in_progress`,
	RunE: runTodosCheck,
}
