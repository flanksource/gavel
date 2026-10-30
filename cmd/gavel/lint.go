package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
	"github.com/spf13/cobra"
)

// LintOptions is the CLI-facing alias for the reusable lint.Options struct.
type LintOptions = lint.Options

func init() {
	var lintCmd *cobra.Command
	lintCmd = clicky.AddNamedCommand("lint", rootCmd, LintOptions{}, func(opts LintOptions) (any, error) {
		opts, err := splitLintPassThroughArgs(lintCmd, opts)
		if err != nil {
			return nil, err
		}
		return runLint(lintCmd.Context(), opts)
	})
	bindLintSourceFlags(lintCmd)
	registerLintLinterSubcommands(lintCmd)
	registerLintShimCommands(lintCmd)
	registerLintPromptCommand(lintCmd)
	if f := lintCmd.Flags().Lookup("failed"); f != nil {
		f.NoOptDefVal = failedAutoSentinel
		f.Usage = "Path to previous results JSON; re-run only linters/files that had violations. Pass without a value to use .gavel/last.json."
	}
}

// splitLintPassThroughArgs moves the arguments after -- from Files to the
// selected linter's native command line.
func splitLintPassThroughArgs(cmd *cobra.Command, opts LintOptions) (LintOptions, error) {
	opts.OptionFields = configuredFlagFields(cmd)
	if cmd.Flags().Changed("generated") {
		generated, err := cmd.Flags().GetBool("generated")
		if err != nil {
			return opts, err
		}
		opts.Generated = &generated
	}
	files, through, err := splitAtDash(cmd, opts.Files)
	if err != nil {
		return opts, err
	}
	opts.Files, opts.PassThroughArgs = files, through
	return opts, nil
}

// resolveAIFix reports whether the AI fix loop should run: either --ai-fix was
// passed explicitly, or -y/--yes opted into auto-fixing lint violations.
func resolveAIFix(o LintOptions) bool { return o.AIFix || o.Yes }

func runLint(ctx context.Context, opts LintOptions) (any, error) {
	if opts.OptionFields == nil {
		opts.OptionFields = runconfig.Present(opts)
	}
	dirs, err := commandDirs(ctx, opts.WorkDir)
	if err != nil {
		return nil, err
	}
	opts.WorkDir = dirs.Invocation
	if err := opts.Configure(); err != nil {
		return nil, err
	}
	opts, err = lint.NormalizeRootArg(opts)
	if err != nil {
		return nil, fmt.Errorf("normalize lint root: %w", err)
	}
	if opts.Dirs, err = commandDirs(ctx, opts.WorkDir); err != nil {
		return nil, err
	}
	opts.WorkDir = opts.Dirs.Invocation
	runStarted := time.Now().UTC()
	if opts.Failed == failedAutoSentinel {
		resolved, err := snapshots.ResolveLast(opts.Dirs.StateRoot)
		if err != nil {
			return nil, fmt.Errorf("--failed: %w", err)
		}
		opts.Failed = resolved
	}
	if opts.PR != "" {
		resolved, err := resolvePRFailed(opts.PR, opts.Dirs, opts.Failed, opts.Baseline, prNarrowLint)
		if err != nil {
			return nil, err
		}
		opts.Failed = resolved
		opts.PR = ""
	}
	clicky.ClearGlobalTasks()
	runCtx, cancelRun := newStopContext(opts.Context, 0)
	defer cancelRun()
	opts.Context = runCtx
	opts.Selection = opts.ChangeSelection()

	if opts.DryRun {
		if err := lint.DryRun(opts); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if opts.UI {
		var err error
		uiServer, err = startLintUI(lintUIOptions{Addr: opts.Addr, Context: opts.Context, Cancel: cancelRun,
			PersistRerun: func(snapshot *testui.Snapshot) error {
				_, err := lint.SaveSnapshot(opts.Dirs.StateRoot, snapshot)
				return err
			},
			Rerun: func(rctx context.Context, req testui.RerunRequest, output *testui.RerunOutputBuffer) ([]*linters.LinterResult, error) {
				rerunOpts := opts
				rerunOpts.Context = rctx
				rerunOpts.OutputTee = output.StdoutWriter()
				return executeLintRerun(rerunOpts, req)
			}})
		if err != nil {
			return nil, err
		}
	}

	// Narrow --failed and establish the effective git root the same way
	// lint.Run does internally, so the post-run triage and snapshot
	// blocks below operate on the git root of the (narrowed) first file.
	// runOpts keeps the full (un-collapsed) file set so lint.Run fans out
	// across every git root; opts collapses to the first group for the log line
	// and the post-run bookkeeping.
	if opts.Failed != "" {
		opts, err = lint.NarrowToFailed(opts)
		if err != nil {
			return nil, err
		}
	}
	runOpts := opts
	groups := lint.GroupFilesByGitRoot(opts)
	opts.WorkDir = groups[0].GitRoot
	opts.Files = groups[0].Files
	if uiServer != nil {
		uiServer.SetGitRoot(opts.WorkDir)
	}
	logger.Infof("Running linters %s", opts.Pretty().ANSI())

	allResults, err := lint.Run(runCtx, runOpts)
	if err != nil {
		// Same contract as a failed test run: serialized formats still get a
		// document, so `--format json=FILE` never silently produces nothing.
		return nil, testRunFailureValue(
			testrunner.RunOptions{WorkDir: opts.WorkDir, Lint: true},
			nil, nil, runStarted, err,
		)
	}

	if resolveAIFix(opts) {
		fixed, fixErr := runAIFix(aiFixOptions{
			Context:       opts.Context,
			WorkDir:       opts.WorkDir,
			Linters:       opts.Linters,
			Runtime:       opts.AIRuntimeOptions,
			MaxIterations: opts.AIFixMaxIters,
			ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
				rerunOpts := opts
				rerunOpts.Selection = opts.Selection.Fresh()
				rerunOpts.Context = rctx
				rerunOpts.AIFix = false
				return lint.Execute(rerunOpts)
			},
		}, allResults)
		if fixErr != nil {
			return nil, fmt.Errorf("ai-fix: %w", fixErr)
		}
		allResults = fixed
	}

	if uiServer != nil {
		return serveLintResults(lintUIResultOptions{Server: uiServer, Options: opts, Results: allResults, Started: runStarted})
	}

	if opts.Triage {
		newRules, err := runTriage(allResults, opts.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("triage failed: %w", err)
		}
		if len(newRules) > 0 {
			gitRoot := repomap.FindGitRoot(opts.WorkDir)
			if gitRoot == "" {
				gitRoot = opts.WorkDir
			}
			repoCfg, err := verify.LoadSingleGavelConfig(filepath.Join(gitRoot, ".gavel.yaml"))
			if err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read repo .gavel.yaml: %w", err)
			}
			repoCfg.Lint.Ignore = append(repoCfg.Lint.Ignore, newRules...)
			if err := verify.SaveGavelConfig(gitRoot, repoCfg); err != nil {
				return nil, fmt.Errorf("failed to save .gavel.yaml: %w", err)
			}
			logger.Infof("Saved %d new ignore rules to .gavel.yaml", len(newRules))
			linters.FilterIgnoredViolations(allResults, newRules)
		}
	}

	return finishLint(opts, allResults, runStarted)
}

type lintUIOptions struct {
	Addr         string
	Context      context.Context
	Cancel       context.CancelFunc
	Writer       io.Writer
	PersistRerun func(*testui.Snapshot) error
	// Rerun re-runs the lint for a rerun request from the UI.
	Rerun func(context.Context, testui.RerunRequest, *testui.RerunOutputBuffer) ([]*linters.LinterResult, error)
}

// startLintUI starts the browser UI for a lint run, marking the initial run in
// progress and wiring the UI's stop and rerun controls.
func startLintUI(opts lintUIOptions) (*testui.Server, error) {
	srv, _, err := startTestUI(testUIOptions{Addr: opts.Addr, Writer: opts.Writer})
	if err != nil {
		return nil, err
	}
	srv.SetStopFunc(standaloneUIStop(opts.Cancel))
	srv.BeginRun("lint")
	srv.SetRerunFunc(func(req testui.RerunRequest, output *testui.RerunOutputBuffer) error {
		clicky.ClearGlobalTasks()
		rerunCtx, cancelRerun := newStopContext(opts.Context, 0)
		defer cancelRerun()
		srv.SetStopFunc(standaloneUIStop(cancelRerun))
		srv.BeginRun("lint")
		srv.SetLintResults(nil)
		results, err := opts.Rerun(rerunCtx, req, output)
		if err != nil {
			srv.SetRunError(err)
			srv.MarkDone()
			return err
		}
		srv.SetLintResults(results)
		snapshot := srv.Snapshot()
		snapshot.Tests = nil
		snapshot.Status.Running = false
		snapshot.Metadata.Kind = "lint"
		snapshot.Metadata.Ended = time.Now().UTC()
		snapshot.Git = snapshots.GitInfo(srv.GitRoot())
		if opts.PersistRerun != nil {
			if err := opts.PersistRerun(&snapshot); err != nil {
				srv.SetRunError(err)
				srv.MarkDone()
				return err
			}
		}
		srv.MarkDone()
		return nil
	})
	return srv, nil
}

type lintUIResultOptions struct {
	Server  *testui.Server
	Options LintOptions
	Results []*linters.LinterResult
	Started time.Time
}

// serveLintResults saves the completed run before waiting for shutdown, then
// returns the latest UI results for terminal rendering.
func serveLintResults(opts lintUIResultOptions) (any, error) {
	if _, err := finishLint(opts.Options, opts.Results, opts.Started); err != nil {
		return nil, err
	}
	// The run's context ends on the UI's Stop button; only a signal ends the UI.
	snapshot, err := lint.Serve(context.Background(), lint.ServeOptions{Server: opts.Server, Results: opts.Results})
	if err != nil {
		return nil, err
	}
	exitCode = 0
	if linters.HasFailures(snapshot.Lint) || snapshot.Error != "" {
		exitCode = 1
	}
	return lintRunReturnValue(&snapshot, opts.Options), nil
}

func finishLint(opts LintOptions, allResults []*linters.LinterResult, runStarted time.Time) (any, error) {
	snap := lint.Snapshot(lint.SnapshotOptions{
		WorkDir: opts.WorkDir, Version: version, Started: runStarted,
		Changes: opts.Selection.Snapshot(), Results: allResults,
	})
	if _, err := lint.SaveSnapshot(opts.Dirs.StateRoot, snap); err != nil {
		return nil, err
	}
	if linters.HasFailures(allResults) {
		exitCode = 1
	}
	return lintRunReturnValue(snap, opts), nil
}

func lintRunReturnValue(snap *testui.Snapshot, opts LintOptions) any {
	if !isPrettyFormat() {
		return snap
	}
	if opts.Summary {
		return linters.NewSummaryView(snap.Lint, opts.SummaryLimit)
	}
	return snap.Lint
}

func executeLintRerun(base LintOptions, req testui.RerunRequest) ([]*linters.LinterResult, error) {
	workDir := base.WorkDir
	if req.WorkDir != "" {
		workDir = req.WorkDir
	}

	rerunOpts := LintOptions{
		Changed:      base.Changed,
		Since:        base.Since,
		ChangedFiles: append([]string(nil), base.ChangedFiles...),
		Selection:    base.Selection.Fresh(),
		WorkDir:      workDir,
		Timeout:      base.Timeout,
		Exclude:      append([]string(nil), base.Exclude...),
		Generated:    base.Generated,
		Tags:         append([]string(nil), base.Tags...),
		Files:        append([]string(nil), req.LintFiles...),
		OutputTee:    base.OutputTee,
	}
	if len(req.LintLinters) > 0 {
		rerunOpts.Linters = append([]string(nil), req.LintLinters...)
	}

	results, err := lint.Execute(rerunOpts)
	if err != nil {
		return nil, err
	}
	return results, nil
}
