package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/baseline"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/verify"
	"github.com/spf13/cobra"
)

var (
	uiServer   *testui.Server
	uiListener net.Listener

	// hookTestsMu protects hookTests, which is the slice of pseudo-tests
	// rendered from pre/post hook state. The stream-forward goroutine
	// prepends it to every testrunner update so hooks stay visible once
	// real tests start streaming.
	hookTestsMu sync.Mutex
	hookTests   []parsers.Test
)

func runTests(opts testrunner.RunOptions) (any, error) {
	opts.AutoStop = testDurationFlags.AutoStop
	opts.IdleTimeout = testDurationFlags.IdleTimeout
	opts.Timeout = testDurationFlags.Timeout
	opts.LintTimeout = testDurationFlags.LintTimeout
	opts.TestTimeout = testDurationFlags.TestTimeout

	runCtx, cancelRun := newStopContext(nil, opts.Timeout)
	defer cancelRun()
	opts.Context = runCtx

	installTimeoutDiagnosticsHook(opts)

	// Route bare fmt.Print / os.Stderr writes through an in-memory buffer
	// while the task renderer is live, then flush everything in order at
	// the end of the command. Keeps the task progress display clean while
	// still surfacing every byte the run produced. Defer-driven so even
	// panics / early returns release the capture before Go's panic
	// printer tries to write to the now-redirected os.Stderr.
	clicky.StartCapturingOutput()
	defer clicky.StopCapturingOutput()

	clicky.ClearGlobalTasks()

	if opts.WorkDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		opts.WorkDir = wd
	}
	runStarted := time.Now().UTC()
	gitInfo := snapshotGitInfo(opts.WorkDir)

	// Dynamic default for --skip-hooks: when the user didn't pass the flag,
	// skip hooks unless $CI is set. This keeps local `gavel test` snappy and
	// auto-enables pre/post hooks in CI / SSH-push contexts.
	if testCmd != nil && !testCmd.Flags().Changed("skip-hooks") {
		opts.SkipHooks = os.Getenv("CI") == ""
	}

	// Load .gavel.yaml once at the top — previously the lint path loaded it
	// lazily, but the push-hook support needs Pre/Post from the same config
	// before tests run.
	gavelCfg, cfgErr := verify.LoadGavelConfig(opts.WorkDir)
	if cfgErr != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", cfgErr)
	}

	// Dry-run: print hooks and delegate to testrunner (which has its own
	// dry-run early-return), then exit without starting the UI, lint, or
	// push-hook execution.
	if opts.DryRun {
		if !opts.SkipHooks {
			printDryRunHooks(opts.WorkDir, gavelCfg.Pre, "pre-hook")
		}
		if _, err := testrunner.Run(opts); err != nil {
			return nil, err
		}
		if opts.Lint {
			for _, workDir := range lintWorkDirs(opts.WorkDir, opts.StartingPaths) {
				if err := displayLintDryRun(LintOptions{WorkDir: workDir, Timeout: "5m"}); err != nil {
					return nil, err
				}
			}
		}
		if !opts.SkipHooks {
			printDryRunHooks(opts.WorkDir, gavelCfg.Post, "post-hook")
		}
		return nil, nil
	}

	// Start the UI BEFORE running pre-hooks so their status/output renders
	// live instead of streaming past a pusher who only sees the UI URL at
	// the end. The stream adapter (below) forwards testrunner updates to
	// the UI while keeping hook pseudo-tests pinned at the top.
	var (
		attachUIUpdates func() chan []parsers.Test
	)
	if opts.UI {
		uiServer, uiListener = startTestUI(opts.Addr)
		if uiServer != nil {
			uiServer.SetVersion(version)
			uiServer.SetRunArgs(snapshotArgs(opts))
			uiServer.SetGitInfo(gitInfo)
			uiServer.EnableDiagnostics(os.Getpid())
			uiServer.SetStopFunc(cancelRun)
			uiServer.SetRunProcess(os.Getpid(), strings.Join(os.Args, " "))
			if len(opts.Frameworks) > 0 {
				uiServer.SetRunFrameworks(opts.Frameworks)
			}
		}
		attachUIUpdates = func() chan []parsers.Test {
			testrunnerUpdates := make(chan []parsers.Test, 16)
			uiUpdates := make(chan []parsers.Test, 16)
			uiServer.StreamFrom(uiUpdates)
			go func() {
				for batch := range testrunnerUpdates {
					uiUpdates <- mergeHooksWithTests(batch)
				}
				close(uiUpdates)
			}()
			return testrunnerUpdates
		}
		attachUIRerunUpdates := func() chan []parsers.Test {
			testrunnerUpdates := make(chan []parsers.Test, 16)
			uiUpdates := make(chan []parsers.Test, 16)
			uiServer.StreamRerunFrom(uiUpdates)
			go func() {
				for batch := range testrunnerUpdates {
					uiUpdates <- mergeHooksWithTests(batch)
				}
				close(uiUpdates)
			}()
			return testrunnerUpdates
		}
		uiServer.BeginRun("initial")
		opts.Updates = attachUIUpdates()
		uiServer.SetRerunFunc(func(req testui.RerunRequest, output *testui.RerunOutputBuffer) error {
			clicky.ClearGlobalTasks()
			rerunCtx, cancelRerun := newStopContext(nil, opts.Timeout)
			defer cancelRerun()
			uiServer.SetStopFunc(cancelRerun)
			uiServer.BeginRun("rerun")
			if req.Lint {
				results, err := executeLintRerun(LintOptions{
					WorkDir:   opts.WorkDir,
					Timeout:   "5m",
					Context:   rerunCtx,
					OutputTee: output.StdoutWriter(),
				}, req)
				if err != nil {
					return err
				}
				uiServer.SetLintResults(results)
				uiServer.MarkDone()
				return nil
			}
			rerunOpts := prepareRerunOptions(opts, req, attachUIRerunUpdates())
			rerunOpts.Context = rerunCtx
			rerunOpts.OutputTee = output.StdoutWriter()
			_, err := testrunner.Run(rerunOpts)
			return err
		})
	}

	if !opts.SkipHooks && len(gavelCfg.Pre) > 0 {
		if err := runPushHooksReportingUI(opts.WorkDir, gavelCfg.Pre, "pre"); err != nil {
			if opts.UI {
				// Flush final hook state to the UI before bailing.
				publishHookSnapshotToUI()
			}
			return nil, err
		}
	}

	// When --lint is set, run linters in parallel with tests
	var lintResults []*linters.LinterResult
	var lintErr error
	var wg sync.WaitGroup
	if opts.Lint {
		wg.Add(1)
		go func() {
			defer wg.Done()
			workDir := opts.WorkDir
			if workDir == "" {
				workDir, _ = os.Getwd()
			}
			lintTimeout := opts.LintTimeout
			if lintTimeout <= 0 {
				lintTimeout = 5 * time.Minute
			}
			for _, dir := range lintWorkDirs(workDir, opts.StartingPaths) {
				results, err := executeLinters(LintOptions{
					WorkDir: dir,
					Timeout: lintTimeout.String(),
					Context: opts.Context,
				})
				if err != nil {
					lintErr = err
					break
				}
				lintResults = append(lintResults, results...)
			}
			if lintErr == nil {
				linters.FilterIgnoredViolations(lintResults, gavelCfg.Lint.Ignore)
			}
			if uiServer != nil {
				uiServer.SetLintResults(lintResults)
			}
		}()
	}

	// fullSummary receives the unfiltered aggregate from the runner. The
	// returned `tests` slice is filtered down to failures-only for display,
	// so we can't derive real pass/skip counts from it.
	var fullSummary parsers.TestSummary
	opts.SummaryOut = &fullSummary

	result, err := testrunner.Run(opts)

	// Post-hooks run after tests, regardless of pass/fail. A failing post
	// hook does NOT mask the main exit code — it's logged as a warning.
	if !opts.SkipHooks && len(gavelCfg.Post) > 0 {
		if postErr := runPushHooksReportingUI(opts.WorkDir, gavelCfg.Post, "post"); postErr != nil {
			logger.Warnf("%v", postErr)
		}
	}

	// Wait for lint to finish
	wg.Wait()

	if lintErr != nil {
		logger.Warnf("Linting failed: %v", lintErr)
	}

	// Apply baseline filtering to lint results when --baseline is set.
	if opts.Baseline != "" && len(lintResults) > 0 {
		baselineSnap, baselineErr := baseline.LoadSnapshot(opts.Baseline)
		if baselineErr != nil {
			return nil, fmt.Errorf("--baseline: %w", baselineErr)
		}
		baseline.FilterNewViolations(lintResults, baseline.ExtractViolationKeys(baselineSnap.Lint))
	}

	// Count lint violations
	var lintViolations int
	for _, lr := range lintResults {
		if lr.Skipped {
			continue
		}
		lintViolations += len(lr.Violations)
	}
	if lintViolations > 0 {
		exitCode = 1
	}

	if err != nil {
		// Even on a runner-level failure (subprocess non-zero, timeout, …)
		// we want the details + summary visible so the user sees what did
		// complete. SummaryOut is populated by the runner before it returns
		// the error. Drain the clicky task pane first so its final ✓/✗/⚠
		// lines land before our details block and summary. When result
		// contains the partial tree we also dump failure / timeout / skip
		// sections; otherwise just the summary.
		clicky.StopCapturingOutput()
		clicky.WaitForGlobalCompletion()
		if tests, ok := result.([]parsers.Test); ok {
			printTestRunDetails(tests, opts.ShowStdout, opts.ShowStderr)
		}
		printTestRunSummary(fullSummary, lintResults)
		return result, err
	}
	if tests, ok := result.([]parsers.Test); ok {
		if opts.Baseline != "" {
			baselineSnap, baselineErr := baseline.LoadSnapshot(opts.Baseline)
			if baselineErr != nil {
				return nil, fmt.Errorf("--baseline: %w", baselineErr)
			}
			tests = baseline.FilterNewTestFailures(tests, baseline.ExtractFailedTestKeys(baselineSnap.Tests))
		}
		summary := parsers.Tests(tests).Sum()
		if summary.Failed > 0 {
			exitCode = 1
		}
		if uiServer != nil {
			if testDurationFlags.Detach {
				snapshot := buildTestSnapshot(opts, tests, lintResults, runStarted, time.Now().UTC(), captureFinalDiagnostics(opts.Diagnostics, os.Getpid()))
				if path, err := snapshots.Save(opts.WorkDir, &snapshot); err != nil {
					logger.Warnf("persist snapshot: %v", err)
				} else {
					logger.V(1).Infof("wrote snapshot to %s", path)
				}
				clicky.CancelAllGlobalTasks()
				// WaitForGlobalCompletion (not -Silent) so the final task
				// pane prints before our details + summary below.
				clicky.WaitForGlobalCompletion()
				if err := handoffDetachedUI(uiListener, snapshot, opts.AutoStop, opts.IdleTimeout); err != nil {
					logger.Warnf("Detached UI handoff failed: %v", err)
				}
				clicky.StopCapturingOutput()
				printTestRunDetails(tests, opts.ShowStdout, opts.ShowStderr)
				printTestRunSummary(fullSummary, lintResults)
				return nil, nil
			}
			// Release the stdout/stderr capture before we block on SIGINT:
			// failure details + any further logger output should print to
			// the real terminal immediately, not sit buffered until the
			// user Ctrl+Cs out of the UI minutes later. The deferred
			// StopCapturingOutput at the top is still safe because the
			// underlying Stop is a no-op when capture wasn't started.
			clicky.StopCapturingOutput()
			// Before blocking on SIGINT, mirror the non-UI CLI contract by
			// printing failure details to the terminal. The UI keeps the full
			// stdout/stderr for live inspection; the terminal render is
			// gated by --show-stdout / --show-stderr so CI logs don't get
			// flooded unless the user opts in.
			printTestRunDetails(tests, opts.ShowStdout, opts.ShowStderr)
			printTestRunSummary(fullSummary, lintResults)
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			return nil, nil
		}
		snapshot := buildTestSnapshot(opts, tests, lintResults, runStarted, time.Now().UTC(), captureFinalDiagnostics(opts.Diagnostics, os.Getpid()))
		if path, err := snapshots.Save(opts.WorkDir, &snapshot); err != nil {
			logger.Warnf("persist snapshot: %v", err)
		} else {
			logger.V(1).Infof("wrote snapshot to %s", path)
		}
		// Release the stdout/stderr capture started at the top of runTests
		// before printing so the details + summary show up on the terminal
		// instead of sitting in clicky's capture buffer until after we
		// return. The deferred StopCapturingOutput is still safe because
		// the underlying Stop is a no-op when capture was already released.
		// Drain the clicky task pane first so its final ✓/✗/⚠ lines land
		// above the details block — the user wants task state before test
		// details + summary, not after.
		clicky.StopCapturingOutput()
		clicky.WaitForGlobalCompletion()
		printTestRunDetails(tests, opts.ShowStdout, opts.ShowStderr)
		printTestRunSummary(fullSummary, lintResults)
		// Return nil rather than the snapshot so clicky doesn't also render
		// Snapshot.Pretty() — a single-line duplicate of the summary above.
		// The snapshot is already persisted to disk for PR/status consumers.
		return nil, nil
	}
	return result, nil
}

// printDryRunHooks renders each hook as a `sh -c` invocation using the
// shared dry-run format. Empty Run fields are skipped, matching
// runPushHooksReportingUI's behavior.
func printDryRunHooks(workDir string, hooks []verify.HookStep, label string) {
	for _, step := range hooks {
		if step.Run == "" {
			continue
		}
		name := step.Name
		if name == "" {
			name = label
		}
		testrunner.PrintDryRunCommand(label, name, "sh", []string{"-c", step.Run}, workDir)
	}
}

// lintWorkDirs resolves StartingPaths into WorkDir values for the lint
// codepath. Each directory path becomes its own WorkDir so linter config
// discovery happens relative to that directory (not the git root).
// When no paths are given, the fallback WorkDir is returned.
func lintWorkDirs(fallback string, startingPaths []string) []string {
	if len(startingPaths) == 0 {
		return []string{fallback}
	}
	var dirs []string
	for _, p := range startingPaths {
		abs := p
		if !filepath.IsAbs(p) {
			abs = filepath.Join(fallback, p)
		}
		abs, _ = filepath.Abs(abs)
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			dirs = append(dirs, abs)
		}
	}
	if len(dirs) == 0 {
		return []string{fallback}
	}
	return dirs
}

// runPushHooksReportingUI runs verify's hook list while also publishing
// each step to the testui (when active) as a pseudo-test: pending while
// running, passed or failed on completion. When the UI is not active,
// it falls back to the plain verify.RunPushHooks path so non-UI push
// output is unchanged.
//
// The pseudo-tests live in the package-level `hookTests` slice and are
// merged into every testrunner stream update by the adapter goroutine in
// runTests, so they stay visible once real tests start appearing.
func runPushHooksReportingUI(workDir string, hooks []verify.HookStep, phase string) error {
	if uiServer == nil {
		// Non-UI path: keep the old streaming-to-stderr behavior unchanged.
		return verify.RunPushHooks(workDir, hooks, phase)
	}

	for _, step := range hooks {
		if step.Run == "" {
			continue
		}
		name := step.Name
		if name == "" {
			name = phase
		}
		logger.Infof("===== %s: %s =====", phase, name)

		idx := appendRunningHookTest(phase, name, step.Run)
		publishHookSnapshotToUI()

		start := time.Now()
		// Capture stdout/stderr into a buffer for the UI AND tee to the
		// user's stderr so a non-UI observer (e.g. ssh push output) still
		// sees it streaming in real time.
		var buf bytes.Buffer
		cmd := exec.Command("sh", "-c", step.Run)
		cmd.Dir = workDir
		cmd.Stdout = io.MultiWriter(&buf, os.Stderr)
		cmd.Stderr = io.MultiWriter(&buf, os.Stderr)
		runErr := cmd.Run()
		finishHookTest(idx, time.Since(start), buf.String(), runErr)
		publishHookSnapshotToUI()

		if runErr != nil {
			return fmt.Errorf("%s hook %q failed: %w", phase, name, runErr)
		}
	}
	return nil
}

// appendRunningHookTest adds a hook pseudo-test to the hookTests slice in
// the "running" (Pending) state and returns its index so the caller can
// flip it to passed/failed when the command completes.
func appendRunningHookTest(phase, name, run string) int {
	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	hookTests = append(hookTests, parsers.Test{
		Name:      name,
		Package:   "hook:" + phase,
		Framework: parsers.Framework("hook"),
		Command:   run,
		Pending:   true,
	})
	return len(hookTests) - 1
}

// finishHookTest flips the hook at idx from Pending to Passed or Failed,
// records its duration and captured output.
func finishHookTest(idx int, dur time.Duration, output string, runErr error) {
	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	if idx < 0 || idx >= len(hookTests) {
		return
	}
	t := &hookTests[idx]
	t.Pending = false
	t.Duration = dur
	t.Stdout = output
	if runErr != nil {
		t.Failed = true
		t.Message = runErr.Error()
	} else {
		t.Passed = true
	}
}

// printFailureDetails writes per-test failure details to stdout for failed
// tests in the tree, honouring --show-stdout / --show-stderr. Called from
// the --ui path right after tests complete so UI users still see the same
// CLI-side failure dump non-UI runs get via printTestRunDetails. Skips
// entirely when nothing failed so passing runs stay quiet.
func printFailureDetails(tests []parsers.Test, showStdout, showStderr testrunner.OutputMode) {
	failed := collectLeaves(tests, func(t parsers.Test) bool {
		return t.Failed && !t.TimedOut
	})
	if len(failed) == 0 {
		return
	}
	fmt.Println(clicky.MustFormat(clicky.Text("Test failures", "bold text-red-600")))
	for _, t := range failed {
		if !showStdout.ShouldShow(true) {
			t.Stdout = ""
		}
		if !showStderr.ShouldShow(true) {
			t.Stderr = ""
		}
		fmt.Println(clicky.MustFormat(t.Pretty()))
	}
}

// printTestRunDetails writes a three-section breakdown of everything the
// user needs to see at the end of a run: failures, timeouts, and skips.
// Each section is elided when empty so passing runs stay quiet. Called from
// every non-UI exit path so the user sees what went wrong without having
// to scroll back through the streaming task pane.
func printTestRunDetails(tests []parsers.Test, showStdout, showStderr testrunner.OutputMode) {
	failed := collectLeaves(tests, func(t parsers.Test) bool {
		return t.Failed && !t.TimedOut
	})
	timedOut := collectLeaves(tests, func(t parsers.Test) bool { return t.TimedOut })
	skipped := collectLeaves(tests, func(t parsers.Test) bool { return t.Skipped })

	printSection := func(title, style string, items []parsers.Test, applyOutputMask bool) {
		if len(items) == 0 {
			return
		}
		fmt.Println(clicky.MustFormat(clicky.Text(fmt.Sprintf("%s (%d)", title, len(items)), "bold "+style)))
		for _, t := range items {
			if applyOutputMask {
				if !showStdout.ShouldShow(true) {
					t.Stdout = ""
				}
				if !showStderr.ShouldShow(true) {
					t.Stderr = ""
				}
			}
			fmt.Println(clicky.MustFormat(t.Pretty()))
		}
	}

	printSection("Test failures", "text-red-600", failed, true)
	printSection("Test timeouts", "text-amber-600", timedOut, true)
	printSection("Skipped tests", "text-yellow-500", skipped, false)
}

// printTestRunSummary writes the end-of-run summary block that mirrors
// `gavel lint`'s behavior: a "Test summary" header + pass/fail/skip/duration
// line, plus the lintSummaryView tree when --lint ran. Called from every exit
// path of runTests so the terminal always shows a visible roll-up after the
// streaming output. No-op when the summary is zero and no lint results ran.
func printTestRunSummary(summary parsers.TestSummary, lintResults []*linters.LinterResult) {
	if out := formatTestRunSummary(summary, lintResults); out != "" {
		fmt.Print(out)
	}
}

// formatTestRunSummary produces the multi-line string printTestRunSummary
// emits. Split from the print wrapper so tests can assert on the exact bytes
// without needing to reassign os.Stdout. Returns an empty string when the
// summary is zero and no lint results ran.
func formatTestRunSummary(summary parsers.TestSummary, lintResults []*linters.LinterResult) string {
	if summary.Total == 0 && len(lintResults) == 0 {
		return ""
	}
	var b strings.Builder
	if summary.Total > 0 {
		header := clicky.Text("Test summary:", "bold text-blue-500")
		b.WriteString(clicky.MustFormat(header.Add(clicky.Text(" ")).Add(summary.Pretty())))
		b.WriteByte('\n')
	}
	if len(lintResults) > 0 {
		b.WriteString(clicky.MustFormat(newLintSummaryView(lintResults, 0), clicky.FormatOptions{Pretty: true}))
		b.WriteByte('\n')
	}
	return b.String()
}

// collectLeaves returns every leaf node (no children) that satisfies pred,
// walking the tree depth-first. Container nodes are skipped; state surfaces
// via their descendants so callers see individual test identities, not
// suite-level rollups.
func collectLeaves(tests []parsers.Test, pred func(parsers.Test) bool) []parsers.Test {
	var out []parsers.Test
	var walk func(t parsers.Test)
	walk = func(t parsers.Test) {
		if len(t.Children) == 0 {
			if pred(t) {
				out = append(out, t)
			}
			return
		}
		for _, c := range t.Children {
			walk(c)
		}
	}
	for _, t := range tests {
		walk(t)
	}
	return out
}

// collectFailed is kept for the existing test_failure_print_test.go
// call site and back-compat with any external callers that relied on the
// leaf-only, failed-only predicate.
func collectFailed(tests []parsers.Test) []parsers.Test {
	return collectLeaves(tests, func(t parsers.Test) bool { return t.Failed })
}

// publishHookSnapshotToUI flushes the current hookTests slice to the UI
// directly via SetResults. This is used before testrunner.Run starts
// streaming (when no real tests have arrived yet) and as a last-ditch
// flush on hook failure.
func publishHookSnapshotToUI() {
	if uiServer == nil {
		return
	}
	hookTestsMu.Lock()
	snapshot := make([]parsers.Test, len(hookTests))
	copy(snapshot, hookTests)
	hookTestsMu.Unlock()
	uiServer.SetResults(snapshot)
}

// mergeHooksWithTests returns a new slice that puts the current hook
// snapshot before the given testrunner batch. Called by the adapter
// goroutine for every testrunner update so hooks stay visible after real
// tests start streaming.
func mergeHooksWithTests(batch []parsers.Test) []parsers.Test {
	hookTestsMu.Lock()
	defer hookTestsMu.Unlock()
	if len(hookTests) == 0 {
		return batch
	}
	merged := make([]parsers.Test, 0, len(hookTests)+len(batch))
	merged = append(merged, hookTests...)
	merged = append(merged, batch...)
	return merged
}

// startTestUI binds an ephemeral TCP listener and starts serving the testui
// on it. Returns the server and the listener so the caller can later
// hand the listener off to a detached child process (fork path in
// handoffDetachedUI) instead of tearing the connection down and rebinding.
func startTestUI(addr string) (*testui.Server, net.Listener) {
	srv := testui.NewServer()
	if addr == "" {
		addr = "localhost"
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(addr, "0"))
	if err != nil {
		fmt.Printf("Failed to start test UI server: %v\n", err)
		return nil, nil
	}
	bound := listener.Addr().(*net.TCPAddr)
	url := fmt.Sprintf("http://%s", net.JoinHostPort(announceHost(addr), strconv.Itoa(bound.Port)))

	go http.Serve(listener, srv.Handler()) //nolint:errcheck

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Test UI at %s\n", url)
	openBrowser(url)
	return srv, listener
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}

// buildRerunArgs translates a UI rerun request into framework-specific flags.
// Empty TestName means rerun the whole package(s).
func buildRerunArgs(req testui.RerunRequest) []string {
	if req.TestName == "" {
		return nil
	}
	switch parsers.Framework(req.Framework) {
	case parsers.GoTest:
		return []string{"-run", "^" + req.TestName + "$"}
	case parsers.Ginkgo:
		focus := req.TestName
		if len(req.Suite) > 0 {
			focus = strings.Join(req.Suite, " ") + " " + req.TestName
		}
		return []string{"--focus", focus}
	}
	return nil
}

func prepareRerunOptions(base testrunner.RunOptions, req testui.RerunRequest, updates chan []parsers.Test) testrunner.RunOptions {
	rerunOpts := base
	rerunOpts.Lint = false
	rerunOpts.Updates = updates
	rerunOpts.RunKind = "rerun"
	if req.WorkDir != "" {
		rerunOpts.WorkDir = req.WorkDir
	}
	rerunOpts.StartingPaths = req.PackagePaths
	rerunOpts.ExtraArgs = buildRerunArgs(req)
	if len(req.PackagePaths) > 0 {
		rerunOpts.Recursive = false
	}
	return rerunOpts
}

// testDurationFlags holds duration flags registered imperatively on `gavel
// test` because clicky cannot bind time.Duration fields via struct tags.
// runTests reads these back into the RunOptions it receives.
var testDurationFlags struct {
	AutoStop    time.Duration
	IdleTimeout time.Duration
	Timeout     time.Duration
	LintTimeout time.Duration
	TestTimeout time.Duration
	Detach      bool
}

// testCmd is the cobra command for `gavel test`. Kept as a package var so
// runTests can query `testCmd.Flags().Changed("skip-hooks")` to distinguish
// an explicit --skip-hooks=... from the unset default.
var testCmd *cobra.Command

func init() {
	testCmd = clicky.AddNamedCommand("test", rootCmd, testrunner.RunOptions{}, runTests)
	testCmd.Flags().SetInterspersed(false)
	testCmd.Flags().BoolVar(&testDurationFlags.Detach, "detach", false,
		"With --ui, fork a detached UI server and exit. The child serves until --auto-stop (default 30m) or --idle-timeout (default 5m) fires.")
	testCmd.Flags().DurationVar(&testDurationFlags.AutoStop, "auto-stop", 0,
		"With --ui --detach, hard wall-clock deadline for the detached UI server (default 30m when --detach is set).")
	testCmd.Flags().DurationVar(&testDurationFlags.IdleTimeout, "idle-timeout", 0,
		"With --ui --detach, exit the detached UI server after this long with no HTTP requests (default 5m when --detach is set).")
	testCmd.Flags().DurationVar(&testDurationFlags.Timeout, "timeout", 10*time.Minute,
		"Global wall-clock deadline for the entire test+lint run. On timeout, diagnostics are captured and every subprocess is killed.")
	testCmd.Flags().DurationVar(&testDurationFlags.LintTimeout, "lint-timeout", 5*time.Minute,
		"Per-linter subprocess deadline when --lint is set. Applies to each linter invocation.")
	testCmd.Flags().DurationVar(&testDurationFlags.TestTimeout, "test-timeout", 5*time.Minute,
		"Per-test-package subprocess deadline. Applies to each go test / ginkgo / vitest invocation.")
}
