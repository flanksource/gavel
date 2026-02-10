package types

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/flanksource/clicky"
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

	// Determine the base directory for working directory resolution
	// Prefer fixture.SourceDir (directory containing fixture file) over opts.WorkDir
	baseDir := opts.WorkDir
	if fixture.SourceDir != "" {
		baseDir = fixture.SourceDir
	}

	// Use the base directory as default working directory
	// If fixture.CWD is specified, resolve it relative to base directory
	workDir := baseDir
	if fixture.CWD != "" && fixture.CWD != "." {
		if filepath.IsAbs(fixture.CWD) {
			// If CWD is absolute, use it directly
			workDir = fixture.CWD
		} else {
			// If CWD is relative, resolve it from the base directory (fixture file location)
			workDir = filepath.Join(baseDir, fixture.CWD)
		}
	}
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

	p := clicky.Exec(exec.Exec, exec.Args...).WithCwd(workDir).Run().Result()

	result.Actual = p

	return fixture.Expected.Evaluate(result, *p)

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
