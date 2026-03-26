package types

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/flanksource/clicky"
	clickyExec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
)

// ExecFixture implements FixtureType for command execution tests
type ExecFixture struct{}

// ValidateFixture implements fixtures.FixtureType.
func (e *ExecFixture) ValidateFixture(fixture fixtures.FixtureTest) error {
	panic("unimplemented")
}

// ensure ExecFixture implements FixtureType
var _ fixtures.FixtureType = (*ExecFixture)(nil)

// Name returns the type identifier
func (e *ExecFixture) Name() string {
	return "exec"
}

// Run executes the command test with gomplate template support
func (e *ExecFixture) Run(ctx context.Context, fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	result := fixtures.FixtureResult{
		Test:     fixture,
		Name:     fixture.Name,
		Type:     "exec",
		Metadata: make(map[string]interface{}),
	}

	// Prepare template context
	templateData := fixture.AsMap()

	workDir := ResolveWorkDir(fixture, opts)
	templateData["workDir"] = workDir
	templateData["executablePath"] = opts.ExecutablePath

	exec, err := fixture.ExecBase().Template(templateData)
	if err != nil {
		return result.Errorf(err, "failed to template exec base")
	}

	bash := clicky.Exec("bash", "-c").AsWrapper()

	// Execute build command if specified (but skip it in task mode since build task handles it)
	if exec.Build != "" {
		logger.V(4).Infof("🔨 Build command: %s", exec.Build)

		p, err := bash(exec.Build)
		if err != nil {
			return result.Errorf(err, "build failed: %s", p.Pretty().ANSI())
		}

	}

	if exec.Exec == "" {
		return result.Errorf(fmt.Errorf("no command specified"), "no command specified")
	}

	var p *clickyExec.ExecResult
	if exec.Terminal == "pty" {
		p = runWithPTY(exec, workDir)
	} else {
		cmd := clicky.Exec(exec.Exec, exec.Args...).WithCwd(workDir)
		if len(exec.Env) > 0 {
			envMap := make(map[string]string, len(exec.Env))
			for k, v := range exec.Env {
				envMap[k] = fmt.Sprintf("%v", v)
			}
			cmd = cmd.WithEnv(envMap)
		}
		p = cmd.Run().Result()
	}

	result.Actual = p
	return fixture.Expected.Evaluate(result, *p)
}

func runWithPTY(execBase fixtures.ExecFixtureBase, workDir string) *clickyExec.ExecResult {
	cmdLine := execBase.Exec
	if len(execBase.Args) > 0 {
		cmdLine += " " + strings.Join(execBase.Args, " ")
	}

	cmd := osExec.Command("bash", "-c", cmdLine)
	cmd.Dir = workDir
	cmd.Env = os.Environ()
	for k, v := range execBase.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%v", k, v))
	}

	now := time.Now()
	var buf bytes.Buffer

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return &clickyExec.ExecResult{
			Stdout:  "",
			Stderr:  "",
			Error:   fmt.Errorf("failed to start PTY: %w", err),
			Started: &now,
		}
	}
	defer ptmx.Close()

	// PTY merges stdout+stderr into a single stream
	_, _ = io.Copy(&buf, ptmx)
	_ = cmd.Wait()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	return &clickyExec.ExecResult{
		Stdout:   buf.String(),
		Stderr:   buf.String(),
		ExitCode: exitCode,
		Started:  &now,
		Duration: time.Since(now),
	}
}

// ResolveWorkDir determines the working directory for fixture execution.
// Priority: test-level CWD > file-level frontmatter CWD > SourceDir > opts.WorkDir
// Relative CWD paths are resolved from SourceDir (fixture file location) or opts.WorkDir.
func ResolveWorkDir(fixture fixtures.FixtureTest, opts fixtures.RunOptions) string {
	baseDir := opts.WorkDir
	if fixture.SourceDir != "" {
		baseDir = fixture.SourceDir
	}

	// Get the merged CWD (file-level frontmatter + test-level override)
	cwd := fixture.ExecBase().CWD
	if cwd == "" || cwd == "." {
		return baseDir
	}
	if filepath.IsAbs(cwd) {
		return cwd
	}
	return filepath.Join(baseDir, cwd)
}

// GetRequiredFields returns required fields
func (e *ExecFixture) GetRequiredFields() []string {
	return []string{"CLI or CLIArgs"}
}

// GetOptionalFields returns optional fields
func (e *ExecFixture) GetOptionalFields() []string {
	return []string{"CWD", "CEL", "Expected.Output", "Expected.Error", "Expected.exitCode", "env"}
}

func init() {
	// Register the exec fixture type
	_ = fixtures.Register(&ExecFixture{})
}
