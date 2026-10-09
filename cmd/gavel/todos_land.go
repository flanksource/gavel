package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	prcreate "github.com/flanksource/gavel/pr/create"
	"github.com/flanksource/gavel/todos/land"
	"github.com/flanksource/gavel/todos/native"
	"github.com/spf13/cobra"
)

type TodosLandOptions struct {
	TodoTargetOptions
	Merge bool   `flag:"merge" help:"Cherry-pick the run's commits onto the current branch"`
	PR    bool   `flag:"pr" help:"Open a pull request carrying the run's commits"`
	Draft bool   `flag:"draft" help:"Open the pull request as a draft (--pr only)"`
	Base  string `flag:"base" help:"Base ref the pull request branches from (--pr only; default: .gavel.yaml pr.base, else origin/main)"`
}

var todosLandCmd *cobra.Command

func init() {
	todosLandCmd = clicky.AddNamedCommand("land", todosCmd, TodosLandOptions{}, func(opts TodosLandOptions) (any, error) {
		return nil, runTodosLand(opts)
	})
	todosLandCmd.Use = "land <id> --merge|--pr"
	todosLandCmd.Short = "Land a TODO run's worktree commits by merge or pull request"
	todosLandCmd.Long = "Land the commits the TODO's newest run step made in its worktree. --merge cherry-picks them " +
		"onto the current branch (aborting cleanly on a conflict); --pr opens a pull request carrying them. The run's " +
		"setup snapshot of your in-progress changes is never landed. Once every commit is confirmed landed the run's " +
		"worktree and shell branch are removed and the landing is recorded on the TODO."
}

func (o TodosLandOptions) landOptions() (land.Options, error) {
	if o.Merge == o.PR {
		return land.Options{}, errors.New("exactly one of --merge or --pr is required")
	}
	if o.Merge {
		if o.Draft || o.Base != "" {
			return land.Options{}, errors.New("--draft and --base only apply to --pr")
		}
		return land.Options{Via: native.LandingMerge}, nil
	}
	return land.Options{Via: native.LandingPR, Draft: o.Draft, Base: o.Base, Deps: prcreate.DefaultDeps()}, nil
}

func runTodosLand(opts TodosLandOptions) error {
	ref, err := opts.One()
	if err != nil {
		return err
	}
	landOpts, err := opts.landOptions()
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
	backed, ok := provider.(land.Provider)
	if !ok {
		return land.ErrNotNative
	}
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return err
	}
	landing, err := land.Land(ctx, backed, todo, landOpts)
	var conflict *prcreate.ConflictError
	if errors.As(err, &conflict) {
		printConflictHint(conflict.Worktree, conflict.Branch)
	}
	if landing != nil {
		fmt.Println(landingText(todo.DisplayID(), landing).ANSI())
	}
	return err
}

func landingText(todoID string, landing *native.RunLanding) api.Text {
	t := clicky.Text("Landed "+todoID, "font-bold text-green-600")
	switch landing.Via {
	case native.LandingPR:
		t = t.Append(fmt.Sprintf(" as PR #%d", *landing.PRNumber), "font-bold").
			Append(" against "+landing.TargetBranch, "").NewLine().
			Append(landing.PRURL, "text-muted")
	default:
		t = t.Append(" on "+landing.TargetBranch, "font-bold").Append(" at "+shortSHA(landing.LandedSHA), "text-muted")
	}
	t = t.NewLine()
	if landing.BranchDeletedAt != nil {
		return t.Append("run worktree and branch removed", "text-muted")
	}
	return t.Append("run branch kept", "text-yellow-600")
}
