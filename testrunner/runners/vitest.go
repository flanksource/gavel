package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Vitest runs vitest tests via `vitest run --reporter=json --outputFile=...`.
// The JSON reporter output is Jest-compatible so we reuse the Jest parser.
type Vitest struct {
	workDir string
	parser  parsers.ResultParser
}

func NewVitest(workDir string) *Vitest {
	return &Vitest{
		workDir: workDir,
		parser:  parsers.NewJestJSON(workDir, parsers.Vitest),
	}
}

func (r *Vitest) Name() parsers.Framework      { return parsers.Vitest }
func (r *Vitest) Parser() parsers.ResultParser { return r.parser }

// detectVitest is true when the directory has a dedicated vitest config, or
// when vitest is a declared dependency. A bare vite.config.* without a vitest
// dep is treated as a build config, not a test config.
func detectVitest(dir string, pkg *pkgJSON) bool {
	if hasConfigFile(dir, []string{"vitest.config"}, nodeConfigExts) {
		return true
	}
	return hasNpmDep(pkg, "vitest")
}

func (r *Vitest) Detect(workDir string) (bool, error) {
	pkg, _ := readPackageJSON(workDir)
	if detectVitest(workDir, pkg) {
		return true, nil
	}
	pkgs, err := walkNodePackages(workDir, detectVitest)
	if err != nil {
		return false, err
	}
	return len(pkgs) > 0, nil
}

func (r *Vitest) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		pkg, _ := readPackageJSON(workDir)
		if detectVitest(workDir, pkg) {
			return []string{"."}, nil
		}
		return nil, nil
	}
	return walkNodePackages(workDir, detectVitest)
}

func (r *Vitest) PackageHasTests(packagePath string) (bool, error) {
	dir := filepath.Join(r.workDir, packagePath)
	pkg, _ := readPackageJSON(dir)
	return detectVitest(dir, pkg), nil
}

func (r *Vitest) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	cwd := filepath.Join(r.workDir, packagePath)
	reportPath := filepath.Join(r.workDir, ".vitest", fmt.Sprintf("vitest-report-%s-%d.json", sanitizePkgPath(packagePath), time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return nil, fmt.Errorf("create vitest report dir: %w", err)
	}

	cmd, pre := detectPackageManager(cwd)
	args := append([]string{}, pre...)
	args = append(args, "vitest", "run", "--reporter=json", "--outputFile="+reportPath)
	args = append(args, extraArgs...)

	process := exec.NewExec(cmd, args...).WithCwd(cwd)
	process.SucceedOnNonZero = true

	return &TestRun{
		Framework:  parsers.Vitest,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

func (r *Vitest) NormalizeFilePath(filePath string) string {
	return normalizeNodeFilePath(r.workDir, filePath)
}
