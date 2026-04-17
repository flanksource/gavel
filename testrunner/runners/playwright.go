package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Playwright runs playwright tests via `playwright test --reporter=json`.
// The JSON output path is controlled by PLAYWRIGHT_JSON_OUTPUT_NAME, which
// Playwright requires to be an absolute path.
type Playwright struct {
	workDir string
	parser  parsers.ResultParser
}

func NewPlaywright(workDir string) *Playwright {
	return &Playwright{
		workDir: workDir,
		parser:  parsers.NewPlaywrightJSON(workDir),
	}
}

func (r *Playwright) Name() parsers.Framework      { return parsers.Playwright }
func (r *Playwright) Parser() parsers.ResultParser { return r.parser }

func detectPlaywright(dir string, pkg *pkgJSON) bool {
	if hasConfigFile(dir, []string{"playwright.config"}, nodeConfigExts) {
		return true
	}
	return hasNpmDep(pkg, "@playwright/test")
}

func (r *Playwright) Detect(workDir string) (bool, error) {
	pkg, _ := readPackageJSON(workDir)
	if detectPlaywright(workDir, pkg) {
		return true, nil
	}
	pkgs, err := walkNodePackages(workDir, detectPlaywright)
	if err != nil {
		return false, err
	}
	return len(pkgs) > 0, nil
}

func (r *Playwright) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		pkg, _ := readPackageJSON(workDir)
		if detectPlaywright(workDir, pkg) {
			return []string{"."}, nil
		}
		return nil, nil
	}
	return walkNodePackages(workDir, detectPlaywright)
}

func (r *Playwright) PackageHasTests(packagePath string) (bool, error) {
	dir := filepath.Join(r.workDir, packagePath)
	pkg, _ := readPackageJSON(dir)
	return detectPlaywright(dir, pkg), nil
}

func (r *Playwright) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	cwd := filepath.Join(r.workDir, packagePath)
	reportPath := filepath.Join(r.workDir, ".playwright", fmt.Sprintf("playwright-report-%s-%d.json", sanitizePkgPath(packagePath), time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return nil, fmt.Errorf("create playwright report dir: %w", err)
	}

	cmd, pre := detectPackageManager(cwd)
	args := append([]string{}, pre...)
	args = append(args, "playwright", "test", "--reporter=json")
	args = append(args, extraArgs...)

	process := exec.NewExec(cmd, args...).WithCwd(cwd).WithEnv(map[string]string{
		"PLAYWRIGHT_JSON_OUTPUT_NAME": reportPath,
	})
	process.SucceedOnNonZero = true

	return &TestRun{
		Framework:  parsers.Playwright,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

func (r *Playwright) NormalizeFilePath(filePath string) string {
	return normalizeNodeFilePath(r.workDir, filePath)
}
