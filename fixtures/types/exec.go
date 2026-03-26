package types

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	osExec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/creack/pty"
	"github.com/flanksource/clicky"
	clickyExec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/repomap"
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
	// Compute root dirs from the base working directory (before CWD expansion)
	baseDir := ResolveWorkDir(fixture, opts)
	gitRoot := repomap.FindGitRoot(baseDir)
	goRoot := findGoModRoot(baseDir)
	rootDir := gitRoot
	if rootDir == "" {
		rootDir = goRoot
	}
	if rootDir == "" {
		rootDir = baseDir
	}

	// Inject auto-injected vars into TemplateVars so they're available
	// in both Template() expansion and CEL evaluation via AsMap()
	if fixture.TemplateVars == nil {
		fixture.TemplateVars = make(map[string]any)
	}
	fixture.TemplateVars["workDir"] = baseDir
	fixture.TemplateVars["executablePath"] = opts.ExecutablePath
	fixture.TemplateVars["GIT_ROOT_DIR"] = gitRoot
	fixture.TemplateVars["GO_ROOT_DIR"] = goRoot
	fixture.TemplateVars["ROOT_DIR"] = rootDir
	fixture.TemplateVars["GOOS"] = runtime.GOOS
	fixture.TemplateVars["GOARCH"] = runtime.GOARCH
	fixture.TemplateVars["GOPATH"] = os.Getenv("GOPATH")
	fixture.TemplateVars["CWD"] = baseDir

	result := fixtures.FixtureResult{
		Test:     fixture,
		Name:     fixture.Name,
		Type:     "exec",
		Metadata: make(map[string]interface{}),
	}

	templateData := fixture.AsMap()

	exec, err := fixture.ExecBase().Template(templateData)
	if err != nil {
		return result.Errorf(err, "failed to template exec base")
	}

	// Resolve final workDir: use expanded CWD if set, otherwise baseDir
	workDir := baseDir
	if exec.CWD != "" && exec.CWD != "." {
		if filepath.IsAbs(exec.CWD) {
			workDir = exec.CWD
		} else {
			workDir = filepath.Join(baseDir, exec.CWD)
		}
	}

	if exec.Env == nil {
		exec.Env = make(map[string]any)
	}
	for _, k := range []string{"GIT_ROOT_DIR", "GO_ROOT_DIR", "ROOT_DIR", "GOOS", "GOARCH", "GOPATH"} {
		if _, ok := exec.Env[k]; !ok {
			exec.Env[k] = templateData[k]
		}
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

func findGoModRoot(path string) string {
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if abs, err := filepath.Abs(dir); err == nil {
				return abs
			}
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func init() {
	// Register the exec fixture type
	_ = fixtures.Register(&ExecFixture{})
}
