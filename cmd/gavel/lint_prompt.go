package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"time"

	"github.com/flanksource/captain/pkg/claude"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/flags"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	"github.com/spf13/cobra"
)

func registerLintPromptCommand(parent *cobra.Command) *cobra.Command {
	var cmd *cobra.Command
	cmd = clicky.AddNamedCommandWithContext("prompt", parent, lintprompt.Options{}, func(ctx context.Context, opts lintprompt.Options) (any, error) {
		opts.AIRuntimeOptions = opts.WithChangedFlags(cmd.Flags())
		positional, through, err := splitAtDash(cmd, cmd.Flags().Args())
		if err != nil {
			return nil, err
		}
		prompt, promptArgs, err := parsePromptArgs(through)
		if err != nil {
			return nil, err
		}
		opts.Prompt = prompt
		opts.Prompt.File, opts.Paths = splitPromptArgs(append(positional, promptArgs...), prompt.Prompt != "")
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		opts.Selection = opts.LintScope(ctx, dir).ChangeSelection()
		if opts.DryRun {
			return lintprompt.Render(ctx, opts)
		}
		// An --ai-fix re-lint re-runs the prompt, so piped stdin is read once and
		// replayed to every run.
		var stdin []byte
		if opts.AIFix && claude.IsStdinPiped() {
			if stdin, err = io.ReadAll(os.Stdin); err != nil {
				return nil, fmt.Errorf("read stdin: %w", err)
			}
		}
		run := func(rctx context.Context, o lintprompt.Options) ([]*linters.LinterResult, error) {
			if stdin != nil {
				o.Stdin = bytes.NewReader(stdin)
			}
			return lintprompt.Run(rctx, o)
		}
		started := time.Now().UTC()
		results, err := run(ctx, opts)
		if err != nil {
			return nil, err
		}
		if opts.AIFix {
			var names []string
			for _, result := range results {
				if !slices.Contains(names, result.Linter) {
					names = append(names, result.Linter)
				}
			}
			results, err = runAIFix(aiFixOptions{Context: ctx, WorkDir: dir, Linters: names, Runtime: opts.AIRuntimeOptions, MaxIterations: opts.AIFixMaxIters, ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
				rerun := opts
				rerun.Selection = opts.Selection.Fresh()
				return run(rctx, rerun)
			}}, results)
			if err != nil {
				return nil, fmt.Errorf("ai-fix: %w", err)
			}
		}
		return finishLint(LintOptions{WorkDir: dir, Summary: opts.Summary, SummaryLimit: opts.SummaryLimit, Selection: opts.Selection}, results, started)
	})
	cmd.Use = "prompt [name|id|file] [path...] [-- prompt flags]"
	cmd.Short = "Run a Captain prompt with the lint output preset"
	cmd.Long = "Run a Captain prompt with the lint output preset.\n\nThe first argument is the prompt; remaining arguments (all of them with -p) are files or directories passed to the prompt as the files variable. --changed, --since and --changed-files narrow that scope to changed files, as in gavel lint. --ai-fix repairs findings with lint.fix from Gavel config and re-runs the prompt until clean.\n\nFlags before -- configure gavel: scope, reporting and --ai-fix, with --model and the other runtime flags selecting the fixer. Arguments after -- go to the Captain prompt: --model, -p, -V, --preset and friends configure the reviewing prompt, and positional arguments there are the prompt and paths."
	cmd.Example = "  gavel lint prompt review.prompt\n  gavel lint prompt review.prompt ./pkg cmd/main.go\n  gavel lint prompt --changed --ai-fix --model agent:opus -- --model agent:sonnet review.prompt\n  git diff | gavel lint prompt review.prompt -- -V change=-\n  gavel lint prompt --since main ./pkg -- -p 'Review the current change' --model agent:sonnet"
	cmd.Args = cobra.ArbitraryArgs
	return cmd
}

// parsePromptArgs parses the arguments after -- with the Captain prompt flag
// set, returning the prompt options and the remaining positional arguments.
func parsePromptArgs(args []string) (captaincli.AIPromptOptions, []string, error) {
	var opts captaincli.AIPromptOptions
	fields, err := flags.ParseStructFields(reflect.TypeOf(opts))
	if err != nil {
		return opts, nil, err
	}
	scratch := &cobra.Command{Use: "prompt"}
	var values []*flags.FlagValue
	for _, info := range fields {
		if fv := flags.BindFlag(scratch, info); fv != nil {
			values = append(values, fv)
		}
	}
	if err := scratch.Flags().Parse(args); err != nil {
		return opts, nil, fmt.Errorf("prompt arguments after --: %w", err)
	}
	value := reflect.ValueOf(&opts).Elem()
	for _, fv := range values {
		if err := flags.AssignFieldValue(value, fv, nil, false); err != nil {
			return opts, nil, fmt.Errorf("prompt arguments after --: %w", err)
		}
	}
	opts.AIRuntimeOptions = opts.WithChangedFlags(scratch.Flags())
	opts.PresetsSet = scratch.Flags().Changed("preset") || scratch.Flags().Changed("no-presets")
	opts.PromptRef = scratch.Flags().Lookup("prompt").Value.String()
	return opts, scratch.Flags().Args(), nil
}

// splitPromptArgs separates the prompt reference from the paths under review.
// An inline -p prompt leaves every positional argument as a path.
func splitPromptArgs(args []string, inline bool) (string, []string) {
	if inline || len(args) == 0 {
		return "", args
	}
	return args[0], args[1:]
}
