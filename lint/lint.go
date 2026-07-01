package lint

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

// Run is the headless lint pipeline: it normalizes the root argument, applies
// --failed narrowing, runs every applicable linter across each git root, and
// applies --baseline filtering. It performs no UI, triage, TODO sync, snapshot,
// summary, or AI-fix work — those stay in cmd/gavel and wrap Run.
func Run(ctx context.Context, opts Options) ([]*linters.LinterResult, error) {
	var err error
	opts, err = NormalizeRootArg(opts)
	if err != nil {
		return nil, fmt.Errorf("normalize lint root: %w", err)
	}
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	if opts.Context == nil {
		opts.Context = ctx
	}
	if opts.Failed != "" {
		opts, err = NarrowToFailed(opts)
		if err != nil {
			return nil, err
		}
	}

	var allResults []*linters.LinterResult
	for _, g := range GroupFilesByGitRoot(opts) {
		groupOpts := opts
		groupOpts.WorkDir = g.GitRoot
		groupOpts.Files = g.Files
		results, err := Execute(groupOpts)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	if opts.Baseline != "" {
		baselineSnap, baselineErr := baseline.LoadSnapshot(opts.Baseline)
		if baselineErr != nil {
			return nil, fmt.Errorf("--baseline: %w", baselineErr)
		}
		baseline.FilterNewViolations(allResults, baseline.ExtractViolationKeys(baselineSnap.Lint))
	}
	return allResults, nil
}

// NarrowToFailed rewrites opts to re-run only the linters/files that had
// violations in the snapshot at opts.Failed. It clears opts.Failed so a
// subsequent Run does not narrow a second time. opts.Failed must already be a
// resolved path; the CLI resolves the `--failed` auto-sentinel before calling.
func NarrowToFailed(opts Options) (Options, error) {
	failedPath := opts.Failed
	snapshot, err := baseline.LoadSnapshot(failedPath)
	if err != nil {
		return opts, fmt.Errorf("--failed: %w", err)
	}
	linterNames, files := baseline.ExtractFailedLintTargets(snapshot.Lint)
	if len(linterNames) == 0 {
		return opts, fmt.Errorf("--failed: no lint violations found in %s", failedPath)
	}
	opts.Linters = linterNames
	if len(files) > 0 {
		opts.Files = files
	}
	opts.Failed = ""
	logger.Infof("--failed: narrowed to linters=%v files=%d from %s", linterNames, len(files), failedPath)
	return opts, nil
}

// Execute runs all applicable linters and returns their results. Reusable by
// both the lint command and the test --lint flag.
func Execute(opts Options) ([]*linters.LinterResult, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}

	timeout, err := time.ParseDuration(opts.Timeout)
	if err != nil {
		timeout = models.DefaultLinterTimeout
	}

	registry := buildLinterRegistry(opts.WorkDir)
	requestedLinters, explicit, err := resolveRequestedLinters(registry, opts.Linters)
	if err != nil {
		return nil, err
	}

	// Resolve the merge-base once per git-root so every per-module golangci
	// invocation shares the same --new-from-rev target.
	var golangciExtraArgs []string
	if ref := lintBaseRef(opts); ref != "" {
		if base, mbErr := resolveMergeBase(opts.WorkDir, ref); mbErr != nil {
			logger.Warnf("golangci-lint --new-from-rev: %v", mbErr)
		} else {
			golangciExtraArgs = []string{"--new-from-rev=" + base}
		}
	}

	group := clicky.StartGroup[*linters.LinterResult](testui.LintTaskGroupName, clickytask.WithConcurrency(1))
	var allResults []*linters.LinterResult
	var lintTasks []clickytask.TypedTask[*linters.LinterResult]
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	groupGitRoot := lintGitRoot(opts.WorkDir)
	for _, name := range requestedLinters {
		linter, ok := registry.Get(name)
		if !ok {
			// resolveRequestedLinters already validated every name; hitting
			// this path means the registry was mutated mid-flight.
			return nil, fmt.Errorf("internal: linter %q missing from registry", name)
		}

		invocations := resolveLinterInvocations(linter, opts)
		if len(invocations) == 0 {
			logger.V(2).Infof("Skipping %s: no project roots found", linter.Name())
			continue
		}

		anyScheduled := false
		skipReason := ""
		for _, inv := range invocations {
			projectCfg, _ := verify.LoadGavelConfig(inv.projectRoot)
			if ok, reason := shouldSelectLinter(inv.projectRoot, projectCfg, linter, explicit); !ok {
				logger.V(2).Infof("Skipping %s at %s: %s", linter.Name(), inv.projectRoot, reason)
				if linter.Name() == "betterleaks" && !anyScheduled {
					allResults = append(allResults, &linters.LinterResult{
						Linter:  linter.Name(),
						Skipped: true,
						Error:   reason,
					})
				}
				continue
			}
			hasDirectConfig := hasDirectMatchingFiles(inv.projectRoot, linterConfigPatterns(linter.Name()))
			executable, reason, err := resolveLinterExecutable(ctx, linter, groupGitRoot, hasDirectConfig, false)
			if err != nil {
				return nil, err
			}
			if executable == "" {
				logger.V(2).Infof("Skipping %s at %s: %s", linter.Name(), inv.projectRoot, reason)
				if skipReason == "" {
					skipReason = reason
				}
				continue
			}

			runOpts := linters.RunOptions{
				WorkDir:    inv.projectRoot,
				Executable: executable,
				Files:      inv.files,
				Ignores:    opts.Ignore,
				Fix:        opts.Fix,
				NoCache:    opts.NoCache,
				Timeout:    timeout,
				ForceJSON:  true,
				OutputTee:  opts.OutputTee,
			}
			if linter.Name() == "golangci-lint" && len(golangciExtraArgs) > 0 {
				runOpts.ExtraArgs = append(runOpts.ExtraArgs, golangciExtraArgs...)
			}

			invCopy := inv
			optsCopy := runOpts
			parentCtx := opts.Context
			lintTasks = append(lintTasks, group.Add(linter.Name(), func(ctx commonsContext.Context, t *clickytask.Task) (*linters.LinterResult, error) {
				runCtx := mergeContexts(ctx, parentCtx)
				result := linters.RunLinterWithTask(runCtx, t, invCopy.linter, optsCopy)
				result.WorkDir = invCopy.projectRoot
				return result, nil
			}))
			anyScheduled = true
		}
		if !anyScheduled && skipReason != "" {
			allResults = append(allResults, &linters.LinterResult{
				Linter:  linter.Name(),
				Skipped: true,
				Error:   skipReason,
			})
		}
	}

	// Wait for the entire group (handles dynamically queued tasks and the
	// concurrency semaphore correctly) before harvesting individual results.
	group.WaitFor()
	for _, task := range lintTasks {
		result, err := task.GetResult()
		if err != nil {
			return nil, err
		}
		if result != nil {
			allResults = append(allResults, result)
		}
	}

	applyPostLintFilters(allResults, opts.WorkDir, opts.Files)
	return allResults, nil
}

// applyPostLintFilters runs the three post-filters every caller of Execute
// wants in the same order: user-supplied scope (drops violations outside the
// requested file set), gitignore (drops violations on gitignored paths), then
// .gavel.yaml lint.ignore rules. workDir is used to resolve relative scope
// entries and to locate the layered .gavel.yaml; each result's own WorkDir is
// used for gitignore discovery and absolute-path relativization inside
// FilterIgnoredViolations.
func applyPostLintFilters(results []*linters.LinterResult, workDir string, files []string) {
	if len(results) == 0 {
		return
	}
	if dropped := linters.FilterViolationsByUserScope(results, workDir, files); dropped > 0 {
		logger.Infof("Filtered %d violations outside requested paths in %s", dropped, workDir)
	}
	if filtered := linters.FilterViolationsByGitIgnoreInResults(results); filtered > 0 {
		logger.Infof("Filtered %d gitignored violations in %s", filtered, workDir)
	}
	if workDir == "" {
		return
	}
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml for %s: %v", workDir, err)
		return
	}
	if filtered := linters.FilterIgnoredViolations(results, cfg.Lint.Ignore); filtered > 0 {
		logger.Infof("Filtered %d ignored violations in %s", filtered, workDir)
	}
}

// Group is one git root and the files (relative to it) scoped to that root.
type Group struct {
	GitRoot string
	Files   []string
}

// GroupFilesByGitRoot buckets opts.Files by their enclosing git root so each
// bucket runs against a single repository. With no files, it returns a single
// group rooted at opts.WorkDir.
func GroupFilesByGitRoot(opts Options) []Group {
	if len(opts.Files) == 0 {
		return []Group{{GitRoot: opts.WorkDir}}
	}

	groups := make(map[string][]string)
	var order []string
	for _, f := range opts.Files {
		abs := resolveLintPath(opts.WorkDir, f)

		dir := abs
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			dir = filepath.Dir(abs)
		}

		gitRoot := utils.FindGitRoot(dir)
		if gitRoot == "" {
			gitRoot = dir
		}

		if _, ok := groups[gitRoot]; !ok {
			order = append(order, gitRoot)
		}
		// Preserve both files and directories as-passed. The linter fan-out
		// (resolveLinterInvocations) uses directory entries as project-root
		// seeds and individual files as bucketing keys.
		rel, err := filepath.Rel(gitRoot, abs)
		if err != nil {
			rel = abs
		}
		groups[gitRoot] = append(groups[gitRoot], rel)
	}

	result := make([]Group, 0, len(order))
	for _, root := range order {
		result = append(result, Group{GitRoot: root, Files: groups[root]})
	}
	return result
}

// DryRun prints the shell command each selected linter would run, without
// executing anything. It loops over every git root discovered from opts.Files
// so multi-repo inputs each print their own plan.
func DryRun(opts Options) error {
	for _, g := range GroupFilesByGitRoot(opts) {
		groupOpts := opts
		groupOpts.WorkDir = g.GitRoot
		groupOpts.Files = g.Files
		if err := dryRunGroup(groupOpts); err != nil {
			return err
		}
	}
	return nil
}

// dryRunGroup mirrors the filter logic in Execute for a single git root.
// Linters that are registered but filtered out (no matching files or not on
// PATH) are printed with a skipped reason so users see the full picture.
// Project-rooted linters print one line per discovered project root.
func dryRunGroup(opts Options) error {
	logger.Infof("🔍 Dry-run mode: showing what would be executed")
	logger.Infof("")

	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	timeout, err := time.ParseDuration(opts.Timeout)
	if err != nil {
		timeout = models.DefaultLinterTimeout
	}

	registry := buildLinterRegistry(opts.WorkDir)
	requestedLinters, explicit, err := resolveRequestedLinters(registry, opts.Linters)
	if err != nil {
		return err
	}

	for _, name := range requestedLinters {
		linter, ok := registry.Get(name)
		if !ok {
			testrunner.PrintDryRunSkipped("lint", name, "unknown linter")
			continue
		}

		invocations := resolveLinterInvocations(linter, opts)
		if len(invocations) == 0 {
			testrunner.PrintDryRunSkipped("lint", linter.Name(), "no project roots found")
			continue
		}

		_, ok = linter.(linters.DryRunner)
		if !ok {
			testrunner.PrintDryRunSkipped("lint", linter.Name(), "no DryRunCommand support")
			continue
		}

		for _, inv := range invocations {
			projectCfg, _ := verify.LoadGavelConfig(inv.projectRoot)
			if ok, reason := shouldSelectLinter(inv.projectRoot, projectCfg, linter, explicit); !ok {
				testrunner.PrintDryRunSkipped("lint", linter.Name()+" @ "+inv.projectRoot, reason)
				continue
			}
			hasDirectConfig := hasDirectMatchingFiles(inv.projectRoot, linterConfigPatterns(linter.Name()))
			executable, reason, err := resolveLinterExecutable(opts.Context, linter, lintGitRoot(opts.WorkDir), hasDirectConfig, true)
			if err != nil {
				return err
			}
			if executable == "" {
				testrunner.PrintDryRunSkipped("lint", linter.Name()+" @ "+inv.projectRoot, reason)
				continue
			}

			runOpts := linters.RunOptions{
				WorkDir:    inv.projectRoot,
				Executable: executable,
				Files:      inv.files,
				Ignores:    opts.Ignore,
				Fix:        opts.Fix,
				NoCache:    opts.NoCache,
				Timeout:    timeout,
				ForceJSON:  true,
			}
			if linter.Name() == "golangci-lint" {
				if ref := lintBaseRef(opts); ref != "" {
					// Dry-run deliberately does not invoke `git merge-base` —
					// show the literal ref as a placeholder so users see the
					// intent without side effects.
					runOpts.ExtraArgs = append(runOpts.ExtraArgs, "--new-from-rev=<merge-base HEAD "+ref+">")
				}
			}
			cmdName, args := linters.PrepareCommand(linter, runOpts)
			testrunner.PrintDryRunCommand("lint", linter.Name(), cmdName, args, inv.projectRoot)
		}
	}
	return nil
}

// mergeContexts derives a child context that is cancelled when either parent
// fires. parent may be nil, in which case primary is returned unchanged.
func mergeContexts(primary commonsContext.Context, parent context.Context) commonsContext.Context {
	if parent == nil {
		return primary
	}
	child, cancel := context.WithCancel(primary)
	go func() {
		select {
		case <-parent.Done():
			cancel()
		case <-child.Done():
		}
	}()
	return commonsContext.NewContext(child)
}
