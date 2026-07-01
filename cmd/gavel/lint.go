package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/todosync"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

// LintOptions is the CLI-facing alias for the reusable lint.Options struct.
type LintOptions = lint.Options

func init() {
	lintCmd := clicky.AddNamedCommand("lint", rootCmd, LintOptions{}, runLint)
	registerLintLinterSubcommands(lintCmd)
	if f := lintCmd.Flags().Lookup("failed"); f != nil {
		f.NoOptDefVal = failedAutoSentinel
		f.Usage = "Path to previous results JSON; re-run only linters/files that had violations. Pass without a value to use .gavel/last.json."
	}
}

// resolveAIFix reports whether the AI fix loop should run: either --ai-fix was
// passed explicitly, or -y/--yes opted into auto-fixing lint violations.
func resolveAIFix(o LintOptions) bool { return o.AIFix || o.Yes }

func runLint(opts LintOptions) (any, error) {
	var err error
	opts, err = lint.NormalizeRootArg(opts)
	if err != nil {
		return nil, fmt.Errorf("normalize lint root: %w", err)
	}
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	runStarted := time.Now().UTC()
	if opts.Failed == failedAutoSentinel {
		resolved, err := snapshots.ResolveLast(opts.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("--failed: %w", err)
		}
		opts.Failed = resolved
	}
	clicky.ClearGlobalTasks()
	runCtx, cancelRun := newStopContext(opts.Context, 0)
	defer cancelRun()
	opts.Context = runCtx

	if opts.DryRun {
		if err := lint.DryRun(opts); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if opts.UI {
		uiServer, uiListener = startTestUI(opts.Addr)
		if uiServer != nil {
			uiServer.SetStopFunc(cancelRun)
			uiServer.BeginRun("initial")
			uiServer.SetRerunFunc(func(req testui.RerunRequest, output *testui.RerunOutputBuffer) error {
				clicky.ClearGlobalTasks()
				rerunCtx, cancelRerun := newStopContext(opts.Context, 0)
				defer cancelRerun()
				uiServer.SetStopFunc(cancelRerun)
				uiServer.BeginRun("rerun")
				rerunOpts := opts
				rerunOpts.Context = rerunCtx
				rerunOpts.OutputTee = output.StdoutWriter()
				results, err := executeLintRerun(rerunOpts, req)
				if err != nil {
					return err
				}
				uiServer.SetLintResults(results)
				uiServer.MarkDone()
				return nil
			})
		}
	}

	// Narrow --failed and establish the effective git root the same way
	// lint.Run does internally, so the post-run triage / sync-todos / snapshot
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
		return nil, err
	}

	if resolveAIFix(opts) {
		fixed, fixErr := runAIFix(opts, allResults)
		if fixErr != nil {
			return nil, fmt.Errorf("ai-fix: %w", fixErr)
		}
		allResults = fixed
	}

	if uiServer != nil {
		uiServer.SetLintResults(allResults)
		uiServer.MarkDone()
		var violations int
		for _, lr := range allResults {
			if lr.Skipped {
				continue
			}
			violations += len(lr.Violations)
		}
		if violations > 0 {
			exitCode = 1
		}
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		return nil, nil
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

	if opts.SyncTodos != "" {
		todosDir := filepath.Join(opts.SyncTodos, "lint")
		syncResult, err := todosync.SyncLintTodos(allResults, todosync.SyncOptions{
			TodosDir: todosDir,
			GroupBy:  opts.GroupBy,
			WorkDir:  opts.WorkDir,
		})
		if err != nil {
			return allResults, fmt.Errorf("failed to sync todos: %w", err)
		}
		logger.Infof("Synced TODOs: %d created, %d updated, %d completed",
			len(syncResult.Created), len(syncResult.Updated), len(syncResult.Completed))
	}

	snap := &testui.Snapshot{
		Metadata: &testui.SnapshotMetadata{
			Version: version,
			Started: runStarted,
			Ended:   time.Now().UTC(),
			Kind:    "lint",
		},
		Git: snapshotGitInfo(opts.WorkDir),
		Status: testui.SnapshotStatus{
			LintRun: true,
		},
		Lint: allResults,
	}
	if path, err := snapshots.Save(opts.WorkDir, snap); err != nil {
		logger.Warnf("persist snapshot: %v", err)
	} else {
		logger.V(1).Infof("wrote snapshot to %s", path)
	}
	// Per-run snapshot so lint-only runs appear in the .gavel run history
	// (the Tests dashboard scans run-*.json); Save() above only writes the
	// sha-keyed latest.
	if path, err := snapshots.SavePerRun(opts.WorkDir, snap, runStarted); err != nil {
		logger.Warnf("persist per-run snapshot: %v", err)
	} else {
		logger.V(1).Infof("wrote per-run snapshot to %s", path)
	}

	if opts.Summary {
		return newLintSummaryView(allResults, opts.SummaryLimit), nil
	}
	return allResults, nil
}

func executeLintRerun(base LintOptions, req testui.RerunRequest) ([]*linters.LinterResult, error) {
	workDir := base.WorkDir
	if req.WorkDir != "" {
		workDir = req.WorkDir
	}

	rerunOpts := LintOptions{
		WorkDir:   workDir,
		Timeout:   base.Timeout,
		Files:     append([]string(nil), req.LintFiles...),
		OutputTee: base.OutputTee,
	}
	if rerunOpts.Timeout == "" {
		rerunOpts.Timeout = "5m"
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
