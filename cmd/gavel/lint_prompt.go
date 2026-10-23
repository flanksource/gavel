package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/flanksource/captain/pkg/claude"
	"github.com/flanksource/clicky"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	"github.com/spf13/cobra"
)

func registerLintPromptCommand(parent *cobra.Command) *cobra.Command {
	var cmd *cobra.Command
	cmd = clicky.AddNamedCommandWithContext("prompt", parent, lintprompt.Options{}, func(ctx context.Context, opts lintprompt.Options) (any, error) {
		opts.AIRuntimeOptions = opts.AIRuntimeOptions.WithChangedFlags(cmd.Flags())
		opts.PresetsSet = cmd.Flags().Changed("preset") || cmd.Flags().Changed("no-presets")
		opts.PromptRef = cmd.Flags().Lookup("prompt").Value.String()
		opts.File, opts.Paths = splitPromptArgs(cmd.Flags().Args(), cmd.Flags().Changed("prompt"))
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
			// --model and friends configure the reviewing prompt; the fixer runs
			// lint.fix from Gavel config.
			results, err = runAIFix(aiFixOptions{Context: ctx, WorkDir: dir, Linters: names, Runtime: defaultAIRuntimeOptions(), MaxIterations: opts.AIFixMaxIters, ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
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
	cmd.Use = "prompt [name|id|file] [path...]"
	cmd.Short = "Run a Captain prompt with the lint output preset"
	cmd.Long = "Run a Captain prompt with the lint output preset.\n\nThe first argument is the prompt; remaining arguments (all of them with -p) are files or directories passed to the prompt as the files variable. --changed, --since and --changed-files narrow that scope to changed files, as in gavel lint. --ai-fix repairs findings with lint.fix from Gavel config and re-runs the prompt until clean."
	cmd.Example = "  gavel lint prompt review.prompt\n  gavel lint prompt review.prompt ./pkg cmd/main.go\n  gavel lint prompt review.prompt --changed --ai-fix\n  git diff | gavel lint prompt review.prompt -V change=-\n  gavel lint prompt -p 'Review the current change' --since main ./pkg --model agent:sonnet"
	cmd.Args = cobra.ArbitraryArgs
	return cmd
}

// splitPromptArgs separates the prompt reference from the paths under review.
// An inline -p prompt leaves every positional argument as a path.
func splitPromptArgs(args []string, inline bool) (string, []string) {
	if inline || len(args) == 0 {
		return "", args
	}
	return args[0], args[1:]
}
