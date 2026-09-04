package types

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/flanksource/clicky"
	clickyExec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/record"
	"github.com/flanksource/gomplate/v3"
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
	// Two directories, deliberately distinct. sourceDir is where the markdown
	// lives — golden files resolve against it. execBase is where commands run:
	// a `setup:` with a checkout moves it into the prepared worktree, and
	// without one the two are the same directory.
	sourceDir := ResolveSourceDir(fixture, opts)
	execBase := opts.Setup.Dir(sourceDir)

	// Compute root dirs from the exec base, not the markdown's directory: a
	// worktree fixture that kept GIT_ROOT_DIR pointing at the original repo
	// would silently exercise the tree it was trying to isolate itself from.
	// The CWD itself may reference these auto-injected vars, so they must be
	// resolved before the final working directory.
	gitRoot := repomap.FindGitRoot(execBase)
	goRoot := findGoModRoot(execBase)
	rootDir := gitRoot
	if rootDir == "" {
		rootDir = goRoot
	}
	if rootDir == "" {
		rootDir = execBase
	}

	if gitRoot != goRoot {
		logger.V(3).Infof("Directories: source=%s exec=%s git=%s go=%s root=%s", sourceDir, execBase, gitRoot, goRoot, rootDir)
	}

	// Inject auto-injected vars into TemplateVars so they're available
	// in both Template() expansion and CEL evaluation via AsMap()
	if fixture.TemplateVars == nil {
		fixture.TemplateVars = make(map[string]any)
	}
	fixture.TemplateVars["workDir"] = execBase
	fixture.TemplateVars["executablePath"] = opts.ExecutablePath
	fixture.TemplateVars["GIT_ROOT_DIR"] = gitRoot
	fixture.TemplateVars["GO_ROOT_DIR"] = goRoot
	fixture.TemplateVars["ROOT_DIR"] = rootDir
	fixture.TemplateVars["SETUP_DIR"] = execBase
	fixture.TemplateVars["GOOS"] = runtime.GOOS
	fixture.TemplateVars["GOARCH"] = runtime.GOARCH
	fixture.TemplateVars["GOPATH"] = os.Getenv("GOPATH")
	fixture.TemplateVars["CWD"] = execBase
	// `setup` carries what the prepared tree turned out to be — commit,
	// worktree, path, dirtyFiles. Absent rather than empty when no setup ran,
	// so `{{.setup.commit}}` in a fixture without one fails instead of
	// templating a blank.
	if opts.Setup != nil {
		fixture.TemplateVars["setup"] = opts.Setup.Extra
	}

	result := fixtures.FixtureResult{
		Test:     fixture,
		Name:     fixture.Name,
		Type:     "exec",
		Metadata: make(map[string]interface{}),
	}

	templateData := fixture.AsMap()
	templatedCWD, err := templateString(fixture.ExecBase().CWD, templateData)
	if err != nil {
		return result.Errorf(err, "failed to template cwd")
	}
	workDir := ResolveWorkDirFromCWD(templatedCWD, execBase, opts)

	fixture.TemplateVars["workDir"] = workDir
	fixture.TemplateVars["CWD"] = workDir
	result.Test = fixture
	templateData = fixture.AsMap()

	exec, err := fixture.ExecBase().Template(templateData)
	if err != nil {
		return result.Errorf(err, "failed to template exec base")
	}
	exec.CWD = templatedCWD

	result.CWD = workDir

	if exec.Env == nil {
		exec.Env = make(map[string]any)
	}
	// Precedence, highest first: the fixture's own `env:` > the setup's
	// environment > the auto-injected roots > whatever the process already
	// had. Both clicky and os/exec start the child from os.Environ(), so a
	// setup's environment is additive and overriding, never hermetic.
	if opts.Setup != nil {
		for k, v := range opts.Setup.Env {
			if _, ok := exec.Env[k]; !ok {
				exec.Env[k] = v
			}
		}
	}
	// The recorder's proxy sits below both declarations: a fixture or setup that
	// names HTTP_PROXY itself is pointing at something it needs, and silently
	// hijacking it would be worse than not recording. `requireEntries:` is how a
	// fixture turns "recorded nothing" into a failure.
	if opts.Recorder != nil {
		for k, v := range opts.Recorder.ProxyEnv {
			if _, ok := exec.Env[k]; !ok {
				exec.Env[k] = v
			}
		}
	}
	for _, k := range []string{"GIT_ROOT_DIR", "GO_ROOT_DIR", "ROOT_DIR", "SETUP_DIR", "GOOS", "GOARCH", "GOPATH"} {
		if _, ok := exec.Env[k]; !ok {
			exec.Env[k] = templateData[k]
		}
	}

	if exec.Exec == "" {
		return result.Errorf(fmt.Errorf("no command specified"), "no command specified")
	}

	// `record: ansi` implies a PTY: there is no ANSI to record from a pipe.
	var ansi *record.ANSIOptions
	if opts.Recorder != nil {
		ansi = opts.Recorder.ANSI
	}

	started := time.Now()
	var p *clickyExec.ExecResult
	var capture *fixtures.Capture
	if exec.Terminal == "pty" || ansi != nil {
		p, capture = runWithPTY(exec, workDir, ansi)
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

	// Harvested here rather than by the runner because the CEL roots have to
	// exist before Evaluate runs the fixture's expression. The window is the
	// child's own lifetime; under the default file scope every test in the file
	// shares one proxy, so overlapping fixtures see each other's traffic — that
	// is what `scope: test` is for.
	var harvest fixtures.Harvest
	evaluate := fixtures.EvaluateOptions{SourceDir: fixture.SourceDir, UpdateGolden: opts.UpdateGolden}
	if opts.Recorder != nil && opts.Recorder.Harvest != nil {
		harvest = opts.Recorder.Harvest(fixtures.HarvestRequest{
			Label:   fixture.Name,
			Start:   started,
			End:     time.Now(),
			Capture: capture,
		})
		result.Recordings = harvest.Recordings
		evaluate.CELVars = harvest.CELVars
	}
	// The change under review, as the CEL root `changed_files`, so a fixture that
	// only makes sense against particular files can say so; absent, not empty,
	// when the verification is not scoped to a change.
	if len(opts.Changed) > 0 {
		if evaluate.CELVars == nil {
			evaluate.CELVars = map[string]any{}
		}
		evaluate.CELVars["changed_files"] = opts.Changed
	}

	// Deliberately fixture.SourceDir, not execBase: `@golden` files belong next
	// to the markdown that asserts them. A worktree is disposable, so writing
	// an updated golden into one would discard it on cleanup. The command moves;
	// its expectations do not.
	evaluated := fixture.Expected.Evaluate(result, *p, evaluate)

	// A `requireEntries` shortfall only decides a fixture the assertions left
	// passing — the fixture's own expectations are the more specific answer.
	if harvest.Err != nil && evaluated.Error == "" {
		return evaluated.Failf("%s", harvest.Err.Error())
	}
	return evaluated
}

// ptyWidth and ptyHeight are the terminal a fixture sees when it does not say.
// A size is mandatory: the previous implementation started a 0x0 PTY, which
// makes a CLI that queries its width fall back to a default that has nothing to
// do with what the fixture asserts.
const (
	ptyWidth  = 120
	ptyHeight = 40
)

// runWithPTY runs the command under a pseudo-terminal. ansi is non-nil when the
// run is being recorded, which adds settled-screen tracking on top; without it
// the capture is only the output stream, which costs no more than the plain
// io.Copy this replaced.
func runWithPTY(execBase fixtures.ExecFixtureBase, workDir string, ansi *record.ANSIOptions) (*clickyExec.ExecResult, *fixtures.Capture) {
	// Invoke the configured executable directly so shells like bash/sh don't
	// get double-wrapped (`bash -c "bash -c '<script>'"` mis-parses: the
	// outer shell treats the inner `bash` as the script and the rest as
	// positional args — the command never runs and we get the target
	// program's help banner instead).
	opts := fixtures.CaptureOptions{
		Command: append([]string{execBase.Exec}, execBase.Args...),
		Dir:     workDir,
		Width:   ptyWidth,
		Height:  ptyHeight,
	}
	for k, v := range execBase.Env {
		opts.Env = append(opts.Env, fmt.Sprintf("%s=%v", k, v))
	}
	if ansi != nil {
		opts.Snapshots = true
		opts.MaxBytes = int64(ansi.MaxBytes)
		opts.SnapshotInterval = ansi.Interval
		if ansi.Width > 0 {
			opts.Width = ansi.Width
		}
		if ansi.Height > 0 {
			opts.Height = ansi.Height
		}
	}

	now := time.Now()
	capture, err := fixtures.CaptureANSI(opts)
	if err != nil {
		return &clickyExec.ExecResult{
			Error:   fmt.Errorf("failed to start PTY: %w", err),
			Started: &now,
		}, nil
	}

	// PTY merges stdout+stderr into a single byte stream; there is no way
	// to separate them at the consumer. Assign the full capture to Stdout
	// only and leave Stderr empty so that CEL expressions (which build
	// `combined := stdout + stderr`) see the stream once, not twice. The
	// doubled form was flagging every non-empty line as a duplicate in
	// ansi.has_duplicates.
	return &clickyExec.ExecResult{
		Stdout:   capture.Raw(),
		ExitCode: capture.ExitCode,
		Started:  &now,
		Duration: time.Duration(capture.DurationMs) * time.Millisecond,
	}, capture
}

// ResolveWorkDir determines the working directory for fixture execution.
// Priority: test-level CWD > file-level frontmatter CWD > prepared setup cwd >
// SourceDir > opts.WorkDir. Relative CWD paths are resolved from the prepared
// setup's directory when the file declared one, otherwise from SourceDir (the
// fixture file's location) or opts.WorkDir.
func ResolveWorkDir(fixture fixtures.FixtureTest, opts fixtures.RunOptions) string {
	baseDir := opts.Setup.Dir(ResolveSourceDir(fixture, opts))

	// Get the merged CWD (file-level frontmatter + test-level override)
	cwd := fixture.ExecBase().CWD
	result := ResolveWorkDirFromCWD(cwd, baseDir, opts)
	logger.V(4).Infof("ResolveWorkDir: opts.WorkDir=%s sourceDir=%s cwd=%s → %s", opts.WorkDir, fixture.SourceDir, cwd, result)
	return result
}

func ResolveSourceDir(fixture fixtures.FixtureTest, opts fixtures.RunOptions) string {
	baseDir := opts.WorkDir
	if fixture.SourceDir != "" {
		baseDir = fixture.SourceDir
	}
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	return baseDir
}

func ResolveWorkDirFromCWD(cwd, baseDir string, opts fixtures.RunOptions) string {
	var result string
	if cwd == "" || cwd == "." {
		result = baseDir
	} else if filepath.IsAbs(cwd) {
		result = cwd
	} else {
		result = filepath.Join(baseDir, cwd)
	}
	logger.V(4).Infof("ResolveWorkDir: opts.WorkDir=%s baseDir=%s cwd=%s → %s", opts.WorkDir, baseDir, cwd, result)
	return result
}

func templateString(value string, data map[string]any) (string, error) {
	value = fixtures.ExpandVars(value, data)
	return gomplate.RunTemplate(data, gomplate.Template{Template: value})
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
