package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/claude"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	lintprompt "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	testui "github.com/flanksource/gavel/testrunner/ui"
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
		prompt, promptArgs, err := captaincli.ParsePromptArgs(through)
		if err != nil {
			return nil, fmt.Errorf("prompt arguments after --: %w", err)
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
		// Each path is its own run, and an --ai-fix re-lint or a --ui rerun runs
		// them again, so piped stdin is read once and replayed to every run.
		var stdin []byte
		if claude.IsStdinPiped() {
			if stdin, err = io.ReadAll(os.Stdin); err != nil {
				return nil, fmt.Errorf("read stdin: %w", err)
			}
		}
		run := func(rctx context.Context, o lintprompt.Options) ([]*linters.LinterResult, error) {
			if stdin != nil {
				o.Stdin = bytes.NewReader(stdin)
			}
			renderer := captaincli.NewEventRenderer(os.Stderr)
			o.OnEvent = func(iteration int, event api.Event) {
				if event.Rejected || event.Kind == api.EventError || (event.Kind == api.EventSystem && strings.HasPrefix(event.Text, "schema validation failed")) {
					renderer.Handle(iteration, event)
				}
			}
			results, err := lintprompt.Run(rctx, o)
			return results, errors.Join(err, renderer.Flush())
		}
		var uiServer *testui.Server
		if opts.UI {
			runCtx, cancelRun := newStopContext(ctx, 0)
			defer cancelRun()
			ctx = runCtx
			// The UI follows each prompt run through Captain's run and session
			// routes; the host keeps them pruned.
			defer captaincli.StartSessionHost(ctx)()
			uiServer, err = startLintUI(lintUIOptions{Addr: opts.Addr, Context: ctx, Cancel: cancelRun,
				Rerun: func(rctx context.Context, req testui.RerunRequest, _ *testui.RerunOutputBuffer) ([]*linters.LinterResult, error) {
					rerun := opts
					rerun.Selection = opts.Selection.Fresh()
					// A file-scoped rerun reviews exactly those files, not the change selection.
					if len(req.LintFiles) > 0 {
						rerun.Paths, rerun.Selection, rerun.Changed, rerun.Since, rerun.ChangedFiles = req.LintFiles, nil, false, "", nil
					}
					return run(rctx, rerun)
				}})
			if err != nil {
				return nil, err
			}
			uiServer.SetGitRoot(dir)
			opts.Publisher = uiServer
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
		if uiServer != nil {
			serveLintResults(uiServer, results)
			return nil, nil
		}
		return finishLint(LintOptions{WorkDir: dir, Summary: opts.Summary, SummaryLimit: opts.SummaryLimit, Selection: opts.Selection}, results, started)
	})
	cmd.Use = "prompt [name|id|file] [path...] [-- prompt flags]"
	cmd.Short = "Run a Captain prompt with the lint output preset"
	cmd.Long = "Run a Captain prompt with the lint output preset.\n\nThe first argument is the prompt; remaining arguments (all of them with -p) are files or directories. Each one is reviewed by its own prompt run, with that path as the files variable, at most --concurrency at once. --changed, --since and --changed-files narrow that scope to changed files, as in gavel lint, one run per changed file.\n\nThe reviewing agent gets three tools: read_file reads a file under the work dir, lint runs the project's configured linters on paths, and report_violation records findings as soon as they are found, several per call, checking each quoted code snippet against its file and line. Reported findings are merged with the prompt's final violations. --ai-fix repairs findings with lint.fix from Gavel config and re-runs the prompt until clean. --ui shows the findings in the browser UI, where they can be re-run.\n\nFlags before -- configure gavel: scope, reporting and --ai-fix, with --model and the other runtime flags selecting the fixer. Arguments after -- go to the Captain prompt: --model, -p, -V, --preset and friends configure the reviewing prompt, and positional arguments there are the prompt and paths."
	cmd.Example = "  gavel lint prompt review.prompt\n  gavel lint prompt review.prompt ./pkg cmd/main.go\n  gavel lint prompt --changed --ai-fix --model agent:opus -- --model agent:sonnet review.prompt\n  gavel lint prompt --ui --changed review.prompt\n  git diff | gavel lint prompt review.prompt -- -V change=-\n  gavel lint prompt --since main ./pkg -- -p 'Review the current change' --model agent:sonnet"
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
