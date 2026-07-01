package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/betterleaks"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/jscpd"
	"github.com/flanksource/gavel/linters/markdownlint"
	"github.com/flanksource/gavel/linters/oxlint"
	"github.com/flanksource/gavel/linters/pyright"
	"github.com/flanksource/gavel/linters/ruff"
	"github.com/flanksource/gavel/linters/tsc"
	"github.com/flanksource/gavel/linters/vale"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/todosync"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

type LintOptions struct {
	Linters      []string        `flag:"linters" help:"Only run the named linters (comma-separated or repeated). Empty = run every detected linter. Unknown names hard-fail."`
	Ignore       []string        `flag:"ignore" help:"Glob patterns to exclude from linting"`
	Triage       bool            `flag:"triage" help:"Interactive mode to select violation types to ignore"`
	Fix          bool            `flag:"fix" help:"Enable auto-fixing"`
	NoCache      bool            `flag:"no-lint-cache" help:"Disable linter result caching/debounce"`
	Timeout      string          `flag:"timeout" help:"Timeout per linter (e.g. 5m, 30s)" default:"5m"`
	SyncTodos    string          `flag:"sync-todos" help:"Sync violations to TODO files in directory (default: .todos/lint)"`
	GroupBy      string          `flag:"group-by" help:"Group synced TODOs by: file, package, message" default:"file"`
	WorkDir      string          `flag:"work-dir" help:"Working directory"`
	Changed      bool            `flag:"changed" help:"Only report new issues vs origin/main (or $GAVEL_CHANGED_BASE)"`
	Since        string          `flag:"since" help:"Only report new issues since <ref> (merge-base with HEAD)"`
	UI           bool            `flag:"ui" help:"Launch browser UI to view violations"`
	Addr         string          `flag:"addr" help:"Interface to bind --ui HTTP server. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine." default:"0.0.0.0"`
	DryRun       bool            `flag:"dry-run" help:"Print the linter commands that would run without executing them"`
	Baseline     string          `flag:"baseline" help:"Path to previous results JSON; only report NEW violations not in baseline"`
	Failed       string          `flag:"failed" help:"Path to previous results JSON; re-run only linters/files that had violations"`
	Summary      bool            `flag:"summary" help:"Collapse output: group by linter -> rule, show count and the first --summary-limit locations"`
	SummaryLimit int             `flag:"summary-limit" help:"Max example locations shown per rule in --summary mode" default:"5"`
	Files        []string        `args:"true"`
	OutputTee    io.Writer       `json:"-"`
	Context      context.Context `json:"-"`

	AIFix         bool `flag:"ai-fix" help:"Invoke the AI configured by 'captain configure' to fix violations and re-lint until clean (or bounded by --ai-fix-max-iterations / --budget)"`
	AIFixMaxIters int  `flag:"ai-fix-max-iterations" help:"Max AI→re-lint cycles" default:"3"`
	Yes           bool `flag:"yes" short:"y" help:"Assume yes: auto-AI-fix lint violations (implies --ai-fix). Does not enable --triage."`

	// Embedded: contributes --model, --backend, --api-key, --no-cache,
	// --budget, --debug, --max-tokens, --temperature, --permission-mode,
	// --edit, --allowed-tools, --disallowed-tools, --mcp, --hooks,
	// --skills, --skill-dir, --user, --project, --memory, --bare.
	// Defaults overlay from ~/.captain.yaml via captain configure.
	captaincli.AIRuntimeOptions
}

func (o LintOptions) Pretty() api.Text {
	t := clicky.Text("")
	if o.WorkDir != "" {
		t = t.Append("WorkDir: ", "text-muted").Append(o.WorkDir, "text-blue-500").Space()
	}
	if len(o.Linters) > 0 {
		t = t.Append("Linters: ", "text-muted").Append(strings.Join(o.Linters, ","), "text-blue-500").Space()
	}
	if o.Fix {
		t = t.Append("Fix: on", "text-green-500").Space()
	}
	if o.AIFix {
		label := "AIFix: on"
		if o.Model != "" {
			label = fmt.Sprintf("AIFix: %s", o.Model)
		}
		t = t.Append(label, "text-green-500").Space()
	}
	if o.Timeout != "" {
		t = t.Append("Timeout: ", "text-muted").Append(o.Timeout, "text-blue-500").Space()
	}
	if len(o.Files) > 0 {
		t = t.Append("Files: ", "text-muted").Append(clicky.CompactList(o.Files), "text-blue-500")
	}
	return t
}

func (o LintOptions) Help() string {
	return `Run linters on the project.

Automatically detects which linters are available and runs them.
Supports: golangci-lint, ruff, eslint, oxlint, pyright, markdownlint, vale, jscpd, betterleaks.

Examples:
  gavel lint
  gavel lint jscpd
  gavel lint jscpd eslint
  gavel lint secrets                 # alias for betterleaks
  gavel lint --linters=golangci-lint
  gavel lint --linters=golangci-lint,ruff
  gavel lint --fix
  gavel lint --triage
  gavel lint -y                      # auto-AI-fix violations (implies --ai-fix)
  gavel lint ./pkg/...`
}

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
	opts, err = normalizeLintRootArg(opts)
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
		groups := groupFilesByGitRoot(opts)
		for _, g := range groups {
			groupOpts := opts
			groupOpts.WorkDir = g.gitRoot
			groupOpts.Files = g.files
			if err := displayLintDryRun(groupOpts); err != nil {
				return nil, err
			}
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

	if opts.Failed != "" {
		snapshot, failedErr := baseline.LoadSnapshot(opts.Failed)
		if failedErr != nil {
			return nil, fmt.Errorf("--failed: %w", failedErr)
		}
		linterNames, files := baseline.ExtractFailedLintTargets(snapshot.Lint)
		if len(linterNames) == 0 {
			return nil, fmt.Errorf("--failed: no lint violations found in %s", opts.Failed)
		}
		opts.Linters = linterNames
		if len(files) > 0 {
			opts.Files = files
		}
		logger.Infof("--failed: narrowed to linters=%v files=%d from %s", linterNames, len(files), opts.Failed)
	}

	groups := groupFilesByGitRoot(opts)
	opts.WorkDir = groups[0].gitRoot
	opts.Files = groups[0].files
	if uiServer != nil {
		uiServer.SetGitRoot(opts.WorkDir)
	}
	logger.Infof("Running linters %s", opts.Pretty().ANSI())

	var allResults []*linters.LinterResult
	for _, g := range groups {
		groupOpts := opts
		groupOpts.WorkDir = g.gitRoot
		groupOpts.Files = g.files
		results, err := executeLinters(groupOpts)
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

	results, err := executeLinters(rerunOpts)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func shouldRunLinter(workDir string, cfg verify.GavelConfig, linterName string, cliExplicit bool, explicitEnabled bool, hasConfig bool) (bool, string) {
	if linterName == "golangci-lint" && utils.FindNearestGoModRoot(workDir) == "" {
		return false, "no go.mod found"
	}
	if linterName == "jscpd" && !cliExplicit && !cfg.Lint.IsLinterEnabled("jscpd", false) {
		return false, "disabled by default; set lint.linters.jscpd.enabled: true"
	}
	if linterName == "betterleaks" {
		if cfg.Secrets.Disabled {
			return false, "disabled via .gavel.yaml"
		}
	}
	if !cliExplicit && !explicitEnabled && linterRequiresDirectConfig(linterName) && !hasConfig {
		if linterName == "betterleaks" {
			return false, "no betterleaks/gitleaks config found"
		}
		return false, fmt.Sprintf("no %s config found in work dir", linterName)
	}
	return true, ""
}

// buildLinterRegistry registers every available linter rooted at workDir.
// Shared between the execute path and the dry-run path so both stay in sync.
func buildLinterRegistry(workDir string) *linters.Registry {
	registry := linters.NewRegistry()
	registry.Register(golangci.NewGolangciLint(workDir))
	registry.Register(ruff.NewRuff(workDir))
	registry.Register(eslint.NewESLint(workDir))
	registry.Register(oxlint.NewOxlint(workDir))
	registry.Register(pyright.NewPyright(workDir))
	registry.Register(tsc.NewTSC(workDir))
	registry.Register(markdownlint.NewMarkdownlint(workDir))
	registry.Register(vale.NewVale(workDir))
	registry.Register(jscpd.NewJSCPD(workDir))
	registry.Register(betterleaks.NewBetterleaks(workDir))
	return registry
}

// linterInvocation describes one scheduled linter run against a specific
// project root (go.mod / package.json / tsconfig.json / ...). A linter may
// produce multiple invocations when the input spans several project roots.
type linterInvocation struct {
	linter      linters.Linter
	projectRoot string   // absolute; WorkDir the linter runs from
	files       []string // relative to projectRoot (may be empty = whole root)
}

// resolveLinterInvocations splits one linter across the project roots it
// should run against. When the linter does not implement ProjectRooted, it
// runs once at opts.WorkDir (current behavior). When it does, roots are
// discovered via the input files (or by scanning opts.WorkDir when no files
// were passed) and files are bucketed + relativized per root.
func resolveLinterInvocations(linter linters.Linter, opts LintOptions) []linterInvocation {
	rooted, ok := linter.(linters.ProjectRooted)
	if !ok {
		return []linterInvocation{{linter: linter, projectRoot: opts.WorkDir, files: opts.Files}}
	}
	markers := rooted.ProjectRootMarkers()
	if len(markers) == 0 {
		return []linterInvocation{{linter: linter, projectRoot: opts.WorkDir, files: opts.Files}}
	}

	if len(opts.Files) == 0 {
		roots := utils.FindAllProjectRoots(opts.WorkDir, markers)
		if len(roots) == 0 {
			return nil
		}
		out := make([]linterInvocation, 0, len(roots))
		for _, root := range roots {
			out = append(out, linterInvocation{linter: linter, projectRoot: root})
		}
		return out
	}

	buckets := make(map[string][]string)
	var order []string
	for _, f := range opts.Files {
		abs := resolveLintPath(opts.WorkDir, f)
		dir := abs
		if info, err := os.Stat(abs); err != nil || !info.IsDir() {
			dir = filepath.Dir(abs)
		}
		root := utils.FindNearestProjectRoot(dir, markers)
		if root == "" {
			logger.V(2).Infof("Skipping %s for %s: no %v found", linter.Name(), f, markers)
			continue
		}
		if _, seen := buckets[root]; !seen {
			order = append(order, root)
		}
		rel, err := filepath.Rel(root, abs)
		if err != nil {
			rel = abs
		}
		buckets[root] = append(buckets[root], rel)
	}
	out := make([]linterInvocation, 0, len(order))
	for _, root := range order {
		out = append(out, linterInvocation{linter: linter, projectRoot: root, files: buckets[root]})
	}
	return out
}

// executeLinters runs all applicable linters and returns their results.
// Reusable by both the lint command and the test --lint flag.
func executeLinters(opts LintOptions) ([]*linters.LinterResult, error) {
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

// applyPostLintFilters runs the three post-filters every caller of
// executeLinters wants in the same order: user-supplied scope (drops
// violations outside the requested file set), gitignore (drops violations on
// gitignored paths), then .gavel.yaml lint.ignore rules. workDir is used to
// resolve relative scope entries and to locate the layered .gavel.yaml; each
// result's own WorkDir is used for gitignore discovery and absolute-path
// relativization inside FilterIgnoredViolations.
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

// lintBaseRef returns the git ref to use for --new-from-rev computation, or
// "" when neither --since nor --changed was set. --since wins if both are
// present.
func lintBaseRef(opts LintOptions) string {
	if opts.Since != "" {
		return opts.Since
	}
	if opts.Changed {
		if v := os.Getenv("GAVEL_CHANGED_BASE"); v != "" {
			return v
		}
		return "origin/main"
	}
	return ""
}

// resolveMergeBase returns the merge-base commit between HEAD and ref.
// Mirrors golangci-lint's own --new-from-rev semantics: issues present at
// merge-base are ignored, only regressions relative to it are reported.
func resolveMergeBase(workDir, ref string) (string, error) {
	cmd := exec.Command("git", "merge-base", "HEAD", ref)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git merge-base HEAD %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

type lintGroup struct {
	gitRoot string
	files   []string
}

func groupFilesByGitRoot(opts LintOptions) []lintGroup {
	if len(opts.Files) == 0 {
		return []lintGroup{{gitRoot: opts.WorkDir}}
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

	result := make([]lintGroup, 0, len(order))
	for _, root := range order {
		result = append(result, lintGroup{gitRoot: root, files: groups[root]})
	}
	return result
}

// displayLintDryRun mirrors the filter logic in executeLinters and prints
// the shell command each selected linter would run, without executing
// anything. Linters that are registered but filtered out (no matching files
// or not on PATH) are printed with a skipped reason so users see the full
// picture. Project-rooted linters print one line per discovered project root.
func displayLintDryRun(opts LintOptions) error {
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

// linterAliases maps CLI-friendly aliases onto registered linter names.
var linterAliases = map[string]string{
	"secrets":       "betterleaks",
	"golangci-lint": "golangci-lint",
	"golangci":      "golangci-lint",
}

// resolveRequestedLinters validates --linters names against the registry and
// returns the canonical list in registry order. Empty input returns every
// registered linter (explicit=false). Any unknown name returns an error with
// the known list so typos fail loudly instead of running the wrong subset.
func resolveRequestedLinters(registry *linters.Registry, requested []string) ([]string, bool, error) {
	if len(requested) == 0 {
		return registry.List(), false, nil
	}
	seen := make(map[string]bool, len(requested))
	out := make([]string, 0, len(requested))
	for _, raw := range requested {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if canonical, ok := linterAliases[name]; ok {
			name = canonical
		}
		if !registry.Has(name) {
			return nil, false, fmt.Errorf("unknown linter %q; known: %s", raw, strings.Join(registry.List(), ", "))
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out, true, nil
}

// hasMatchingFiles checks if any files in workDir match at least one of the glob patterns.
// Uses GlobWalk to bail early on first match instead of scanning the entire tree.
func hasMatchingFiles(workDir string, patterns []string) bool {
	fsys := os.DirFS(workDir)
	for _, pattern := range patterns {
		found := false
		_ = doublestar.GlobWalk(fsys, pattern, func(path string, d os.DirEntry) error {
			found = true
			return doublestar.SkipDir
		})
		if found {
			return true
		}
	}
	return false
}
