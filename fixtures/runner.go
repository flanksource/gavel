package fixtures

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3"
	"github.com/samber/lo"
)

// RunnerOptions configures the fixture runner
type RunnerOptions struct {
	Paths          []string // Fixture file paths/patterns
	Format         string   // Output format: tree, table, json, yaml, csv
	Filter         string   // Filter tests by name pattern (glob)
	NoColor        bool     // Disable colored output
	WorkDir        string   // Working directory
	MaxWorkers     int      // Maximum number of parallel workers
	Logger         logger.Logger
	ExecutablePath string              // Path to the current executable (for fixtures to use)
	OnResult       func(FixtureResult) // Called after each fixture completes
	OnParsed       func(*FixtureNode)  // Called after fixture files are parsed, before execution
}

// Runner manages fixture test execution using typed tasks
type Runner struct {
	options    RunnerOptions
	fixtures   []FixtureTest
	evaluator  *CELEvaluator
	tree       *FixtureNode // Hierarchical tree structure
	daemonCmd  *exec.Cmd
	daemonPort int
}

// NewRunner creates a new fixture runner
func NewRunner(opts RunnerOptions) (*Runner, error) {
	// Create CEL evaluator
	evaluator, err := NewCELEvaluator()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL evaluator: %w", err)
	}

	return &Runner{
		options:   opts,
		fixtures:  []FixtureTest{},
		evaluator: evaluator,
		tree: &FixtureNode{
			Name: "Fixtures",
			Type: SectionNode,
		},
	}, nil
}

// Tree returns the parsed fixture tree.
func (r *Runner) Tree() *FixtureNode {
	return r.tree
}

// SetOnResult sets the per-fixture completion callback (can be set after NewRunner).
func (r *Runner) SetOnResult(fn func(FixtureResult)) {
	r.options.OnResult = fn
}

// Run executes the fixture tests and returns the result tree.
// The caller is responsible for formatting/printing the output.
func (r *Runner) Run() (*FixtureNode, error) {
	if err := r.parseFixtureFiles(); err != nil {
		return nil, fmt.Errorf("failed to parse fixture files: %w", err)
	}

	if r.options.OnParsed != nil {
		r.options.OnParsed(r.tree)
	}

	if r.options.Filter != "" {
		r.filterTests()
	}

	if len(r.fixtures) == 0 {
		return nil, fmt.Errorf("no fixtures found")
	}

	results, err := r.executeFixtures()
	if err != nil {
		return nil, fmt.Errorf("failed to execute fixtures: %w", err)
	}

	clicky.WaitForGlobalCompletion()

	if results.Summary.Failed > 0 {
		return r.tree, fmt.Errorf("fixture tests failed")
	}

	return r.tree, nil
}

// parseFixtureFiles parses all fixture files from the provided paths and builds tree structure
func (r *Runner) parseFixtureFiles() error {
	var allFixtures []FixtureTest

	for _, pattern := range r.options.Paths {
		// Expand glob patterns
		matches, err := doublestar.FilepathGlob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern '%s': %w", pattern, err)
		}

		if len(matches) == 0 {
			logger.Warnf("No files matched pattern: %s", pattern)
			continue
		}

		for _, filepath := range matches {
			// Parse with tree structure
			fileTree, err := ParseMarkdownFixturesWithTree(filepath)
			if err != nil {
				return fmt.Errorf("failed to parse fixture file '%s': %w", filepath, err)
			}

			// Merge file tree into main tree
			if fileTree != nil {
				r.tree.AddChild(fileTree)
			}

			// Also maintain flat fixture list for backwards compatibility
			fixtures, err := ParseMarkdownFixtures(filepath)
			if err != nil {
				return fmt.Errorf("failed to parse fixture file '%s': %w", filepath, err)
			}

			logger.Debugf("Parsed %d fixtures from %s", len(fixtures), filepath)
			// Extract FixtureTest from each FixtureNode
			for _, node := range fixtures {
				if node.Test != nil {
					allFixtures = append(allFixtures, *node.Test)
				}
			}
		}
	}

	r.fixtures = allFixtures

	// Log the loaded fixtures
	fileCount := len(r.tree.Children)
	logger.Infof("Loaded %d total fixtures in %d files", len(allFixtures), fileCount)
	return nil
}

// filterTests applies name filtering to loaded tests
func (r *Runner) filterTests() {
	var filtered []FixtureTest

	for _, fixture := range r.fixtures {
		match, err := doublestar.Match(r.options.Filter, fixture.Name)
		if err != nil {
			logger.Warnf("Invalid filter pattern '%s': %v", r.options.Filter, err)
			continue
		}
		if match {
			filtered = append(filtered, fixture)
		}
	}

	logger.Infof("Filtered to %d fixtures matching '%s'", len(filtered), r.options.Filter)
	r.fixtures = filtered
}

// executeFixtures runs all fixtures using typed task groups
func (r *Runner) executeFixtures() (*FixtureGroup, error) {
	results := &FixtureGroup{
		Tests:   make([]FixtureNode, 0, len(r.fixtures)),
		Summary: Stats{},
	}

	ctx := flanksourceContext.NewContext(context.Background())

	// Run build command synchronously before any fixtures
	buildCmd := r.getBuildCommand()
	if buildCmd != "" {
		logger.V(2).Infof("Running build command: %s", buildCmd)
		if err := r.executeBuildCommand(ctx, buildCmd); err != nil {
			return nil, fmt.Errorf("build failed, skipping all fixtures: %w", err)
		}
		logger.V(2).Infof("Build completed successfully")
	}

	// Start daemon if configured
	daemonCmd := r.getDaemonCommand()
	if daemonCmd != "" {
		if err := r.startDaemon(ctx, daemonCmd); err != nil {
			return nil, fmt.Errorf("daemon failed to start: %w", err)
		}
		defer r.stopDaemon()
	}

	// Create typed task group for fixture execution
	fixtureGroup := task.StartGroup[FixtureResult]("Fixture Tests")

	taskToNodeMap := make(map[task.TypedTask[FixtureResult]]*FixtureNode)
	r.tree.Walk(func(node *FixtureNode) {
		if node.Test != nil {
			typedTask := fixtureGroup.Add(node.Test.String(), func(ctx flanksourceContext.Context, t *task.Task) (FixtureResult, error) {
				result, err := r.executeFixture(ctx, *node.Test)
				if r.options.OnResult != nil {
					r.options.OnResult(result)
				}
				return result, err
			}, clicky.WithTaskTimeout(2*time.Minute))
			taskToNodeMap[typedTask] = node
		}
	})

	// Wait for all fixture tasks to complete and collect results
	groupResult := fixtureGroup.WaitFor()
	if groupResult.Error != nil {
		logger.Warnf("Some fixture tests failed: %v", groupResult.Error)
	}

	// Process results
	fixtureResults, err := fixtureGroup.GetResults()
	if err != nil {
		return nil, fmt.Errorf("failed to get fixture results: %w", err)
	}

	for typedTask, result := range fixtureResults {
		// Create a FixtureNode for the result
		resultNode := FixtureNode{
			Name:    result.Name,
			Type:    TestNode,
			Results: &result,
		}
		results.Tests = append(results.Tests, resultNode)

		// Update the corresponding tree node with results
		if testNode, exists := taskToNodeMap[typedTask]; exists {
			testNode.Results = &result
		} else {
			logger.Warnf("No tree node found for task: %s", typedTask.Name())
		}
	}

	r.tree.Stats = lo.ToPtr(r.tree.GetStats())
	results.Summary = *r.tree.Stats

	// Prune empty sections from the tree
	r.tree.PruneEmptySections()

	return results, nil
}

// getBuildCommand extracts build command from first fixture that has one
func (r *Runner) getBuildCommand() string {
	for _, fixture := range r.fixtures {
		if fixture.Build != "" {
			return fixture.Build
		}
	}
	return ""
}

// executeBuildCommand runs the build command with context cancellation and gomplate templating
func (r *Runner) executeBuildCommand(ctx flanksourceContext.Context, buildCmd string) error {
	// Prepare template context for build command
	templateData := make(map[string]interface{})
	templateData["PWD"] = r.options.WorkDir
	templateData["WorkDir"] = r.options.WorkDir
	templateData["GOOS"] = runtime.GOOS
	templateData["GOARCH"] = runtime.GOARCH
	templateData["GOPATH"] = os.Getenv("GOPATH")

	// Template the build command (expand $VAR first, then gomplate)
	buildCmd = ExpandVars(buildCmd, templateData)
	templatedCmd, err := renderBuildTemplate(buildCmd, templateData)
	if err != nil {
		ctx.Errorf("Failed to template build command: %v", err)
		return fmt.Errorf("failed to template build command: %w", err)
	}

	ctx.Logger.V(4).Infof("🔨 Build command: %s", templatedCmd)

	cmd := exec.CommandContext(ctx, "sh", "-c", templatedCmd)
	cmd.Dir = r.options.WorkDir

	var buildOut bytes.Buffer
	cmd.Stdout = &buildOut
	cmd.Stderr = &buildOut

	if err := cmd.Run(); err != nil {
		ctx.Errorf("Build failed: %v\nOutput: %s", err, buildOut.String())
		return fmt.Errorf("build command failed: %v\nOutput: %s", err, buildOut.String())
	}

	if buildOut.Len() > 0 {
		ctx.Logger.V(5).Infof("Build output: %s", buildOut.String())
	}

	return nil
}

// executeFixture runs a single fixture test
func (r *Runner) executeFixture(ctx flanksourceContext.Context, fixture FixtureTest) (FixtureResult, error) {
	if reason := fixture.ShouldSkip(); reason != "" {
		return FixtureResult{
			Name:   fixture.Name,
			Status: task.StatusSKIP,
			Test:   fixture,
			Error:  reason,
		}, nil
	}

	// Get the appropriate fixture type from registry
	fixtureType, err := DefaultRegistry.GetForFixture(fixture)
	if err != nil {
		return FixtureResult{
			Name:   fixture.Name,
			Status: task.StatusERR,
			Test:   fixture,
			Error:  err.Error(),
		}, nil
	}

	if r.daemonPort > 0 {
		if fixture.TemplateVars == nil {
			fixture.TemplateVars = make(map[string]any)
		}
		fixture.TemplateVars["port"] = strconv.Itoa(r.daemonPort)
	}

	if r.options.WorkDir == "" {
		r.options.WorkDir, _ = os.Getwd()
	}
	ctx.Logger.V(5).Infof("Using CWD: %s", r.options.WorkDir)

	// Prepare run options with flanksource context
	opts := RunOptions{
		WorkDir:        r.options.WorkDir,
		Verbose:        ctx.Logger.IsLevelEnabled(logger.Debug),
		NoCache:        false,
		Evaluator:      r.evaluator,
		ExecutablePath: r.options.ExecutablePath,
		ExtraArgs: map[string]interface{}{
			"flanksource_context": ctx,
		},
	}

	start := time.Now()
	// Run the fixture test
	result := fixtureType.Run(ctx, fixture, opts)
	result.Duration = time.Since(start)

	return result, nil
}

// getDaemonCommand extracts daemon command from first fixture that has one
func (r *Runner) getDaemonCommand() string {
	for _, fixture := range r.fixtures {
		if fixture.FrontMatter.Daemon != "" {
			return fixture.FrontMatter.Daemon
		}
	}
	return ""
}

// startDaemon picks a free port, templates the command, starts the process, and waits for the port to be ready.
func (r *Runner) startDaemon(ctx flanksourceContext.Context, daemonCmd string) error {
	port, err := freePort()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	r.daemonPort = port

	templateData := map[string]interface{}{
		"port":    strconv.Itoa(port),
		"PWD":     r.options.WorkDir,
		"WorkDir": r.options.WorkDir,
		"GOOS":    runtime.GOOS,
		"GOARCH":  runtime.GOARCH,
	}

	daemonCmd = ExpandVars(daemonCmd, templateData)
	templated, err := renderBuildTemplate(daemonCmd, templateData)
	if err != nil {
		return fmt.Errorf("failed to template daemon command: %w", err)
	}

	logger.Infof("Starting daemon on port %d: %s", port, templated)

	cmd := exec.CommandContext(ctx, "sh", "-c", templated)
	cmd.Dir = r.options.WorkDir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}
	r.daemonCmd = cmd

	// Wait for port to be ready
	addr := net.JoinHostPort("localhost", strconv.Itoa(port))
	for i := 0; i < 60; i++ {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			logger.Infof("Daemon ready on port %d", port)
			return nil
		}
		// Check if process died
		if cmd.ProcessState != nil {
			return fmt.Errorf("daemon exited prematurely with code %d", cmd.ProcessState.ExitCode())
		}
		time.Sleep(500 * time.Millisecond)
	}
	r.stopDaemon()
	return fmt.Errorf("daemon did not start listening on port %d within 30s", port)
}

// stopDaemon sends SIGTERM, waits up to 5s, then SIGKILL.
func (r *Runner) stopDaemon() {
	if r.daemonCmd == nil || r.daemonCmd.Process == nil {
		return
	}

	logger.Infof("Stopping daemon (PID %d)", r.daemonCmd.Process.Pid)

	// Kill the process group to include child processes
	pgid := -r.daemonCmd.Process.Pid
	_ = syscall.Kill(pgid, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		_ = r.daemonCmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		logger.Warnf("Daemon did not exit after SIGTERM, sending SIGKILL")
		_ = syscall.Kill(pgid, syscall.SIGKILL)
		<-done
	}

	r.daemonCmd = nil
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}

// renderBuildTemplate renders a gomplate template for build commands
func renderBuildTemplate(template string, data map[string]interface{}) (string, error) {
	return gomplate.RunTemplate(data, gomplate.Template{
		Template: template,
	})
}
