package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/verify"
)

var (
	uiServer   *testui.Server
	uiListener net.Listener
)

func runTests(opts testrunner.RunOptions) (any, error) {
	opts.AutoStop = testDurationFlags.AutoStop
	opts.IdleTimeout = testDurationFlags.IdleTimeout

	if opts.WorkDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		opts.WorkDir = wd
	}

	var updates chan []parsers.Test
	if opts.UI {
		uiServer, uiListener = startTestUI()
		updates = make(chan []parsers.Test, 16)
		opts.Updates = updates
		uiServer.StreamFrom(updates)
		uiServer.SetRerunFunc(func(req testui.RerunRequest) error {
			rerunOpts := opts
			rerunOpts.Lint = false
			rerunOpts.Updates = updates
			rerunOpts.StartingPaths = req.PackagePaths
			rerunOpts.ExtraArgs = buildRerunArgs(req)
			_, err := testrunner.Run(rerunOpts)
			return err
		})
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
			lintResults, lintErr = executeLinters(LintOptions{
				WorkDir: workDir,
				Linters: "*",
				Timeout: "5m",
			})
			if lintErr == nil {
				gavelCfg, err := verify.LoadGavelConfig(workDir)
				if err != nil {
					logger.Warnf("Failed to load .gavel.yaml: %v", err)
				}
				linters.FilterIgnoredViolations(lintResults, gavelCfg.Lint.Ignore)
			}
			if uiServer != nil {
				uiServer.SetLintResults(lintResults)
			}
		}()
	}

	result, err := testrunner.Run(opts)

	// Wait for lint to finish
	wg.Wait()

	if lintErr != nil {
		logger.Warnf("Linting failed: %v", lintErr)
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
		return result, err
	}
	if tests, ok := result.([]parsers.Test); ok {
		summary := parsers.Tests(tests).Sum()
		if summary.Failed > 0 {
			exitCode = 1
		}
		if uiServer != nil {
			if opts.AutoStop > 0 || opts.IdleTimeout > 0 {
				if err := handoffDetachedUI(uiListener, tests, lintResults, opts.AutoStop, opts.IdleTimeout); err != nil {
					logger.Warnf("Detached UI handoff failed: %v", err)
				}
				return nil, nil
			}
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			<-sig
			return nil, nil
		}
	}

	if opts.Lint {
		return struct {
			Tests any                     `json:"tests"`
			Lint  []*linters.LinterResult `json:"lint"`
		}{result, lintResults}, nil
	}
	return result, nil
}

// startTestUI binds an ephemeral TCP listener and starts serving the testui
// on it. Returns the server and the listener so the caller can later
// hand the listener off to a detached child process (fork path in
// handoffDetachedUI) instead of tearing the connection down and rebinding.
func startTestUI() (*testui.Server, net.Listener) {
	srv := testui.NewServer()
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Printf("Failed to start test UI server: %v\n", err)
		return nil, nil
	}
	port := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", port)

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

// testDurationFlags holds duration flags registered imperatively on `gavel
// test` because clicky cannot bind time.Duration fields via struct tags.
// runTests reads these back into the RunOptions it receives.
var testDurationFlags struct {
	AutoStop    time.Duration
	IdleTimeout time.Duration
}

func init() {
	testCmd := clicky.AddNamedCommand("test", rootCmd, testrunner.RunOptions{}, runTests)
	testCmd.Flags().SetInterspersed(false)
	testCmd.Flags().DurationVar(&testDurationFlags.AutoStop, "auto-stop", 0,
		"With --ui, fork a detached UI server that serves the completed run until this hard wall-clock deadline fires (0 = block on SIGINT).")
	testCmd.Flags().DurationVar(&testDurationFlags.IdleTimeout, "idle-timeout", 0,
		"With --ui --auto-stop, exit the detached UI server after this long with no HTTP requests.")
}
