package fixtures

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"time"

	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3"
)

// getBuildCommand extracts build command from first fixture that has one
func (r *Runner) getBuildCommand() (string, *PreparedSetup) {
	return r.firstDeclared("build", func(f FixtureTest) string { return f.Build })
}

// firstDeclared returns the first command any fixture declares, together with
// the setup prepared for the file that declared it. Tree order is the order
// parseFixtureFiles builds r.fixtures in, so first-wins means the same fixture
// either way.
//
// build: and daemon: are run-wide, but a setup is per file, so a run with more
// than one prepared tree can only put them in one of them. That is a real
// ambiguity rather than a bug to paper over — warn and name the file, and make
// them per-file only if the warning proves common.
func (r *Runner) firstDeclared(kind string, pick func(FixtureTest) string) (string, *PreparedSetup) {
	var command, owner string
	r.tree.Walk(func(node *FixtureNode) {
		if command != "" || node.Test == nil {
			return
		}
		if value := pick(*node.Test); value != "" {
			command = value
			if node.Origin != nil {
				owner = node.Origin.File
			}
		}
	})

	setup := r.setups[owner]
	switch {
	case command == "" || len(r.setups) == 0:
	case setup == nil:
		logger.Warnf("%s: runs in %s because %s declares no setup:, while other files in this run do",
			kind, r.options.WorkDir, owner)
	case len(r.setups) > 1:
		logger.Warnf("%s: runs in %s (declared by %s); the other %d prepared setup(s) in this run do not get their own",
			kind, setup.Cwd, owner, len(r.setups)-1)
	}
	return command, setup
}

// executeBuildCommand runs the build command with context cancellation and gomplate templating
func (r *Runner) executeBuildCommand(ctx flanksourceContext.Context, buildCmd string, setup *PreparedSetup) error {
	workDir := setup.Dir(r.options.WorkDir)

	// Prepare template context for build command
	templateData := make(map[string]interface{})
	templateData["PWD"] = workDir
	templateData["WorkDir"] = workDir
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
	cmd.Dir = workDir
	cmd.Env = setup.Environ()

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

// getDaemonCommand extracts daemon command from first fixture that has one
func (r *Runner) getDaemonCommand() (string, *PreparedSetup) {
	return r.firstDeclared("daemon", func(f FixtureTest) string { return f.FrontMatter.Daemon })
}

// startDaemon picks a free port, templates the command, starts the process, and waits for the port to be ready.
func (r *Runner) startDaemon(ctx flanksourceContext.Context, daemonCmd string, setup *PreparedSetup) error {
	port, err := freePort()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	r.daemonPort = port

	workDir := setup.Dir(r.options.WorkDir)
	templateData := map[string]interface{}{
		"port":    strconv.Itoa(port),
		"PWD":     workDir,
		"WorkDir": workDir,
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
	cmd.Dir = workDir
	cmd.Env = setup.Environ()
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
