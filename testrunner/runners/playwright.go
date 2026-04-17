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

// playwrightTestSuffixes: Playwright's default tests live in *.spec.*, plus
// the common *.e2e.* naming for projects that split unit vs e2e.
var playwrightTestSuffixes = []string{".spec.js", ".spec.jsx", ".spec.ts", ".spec.tsx", ".spec.mjs", ".spec.cjs",
	".e2e.js", ".e2e.ts", ".e2e.tsx"}

// detectPlaywright matches on playwright.config.* or playwright-ct.config.*
// (component-testing variant), or on the @playwright/test dep paired with
// at least one spec file — otherwise `playwright test` just exits with
// "No tests found" and we'd rather gavel surface nothing than a noop.
func detectPlaywright(dir string, pkg *pkgJSON) bool {
	if hasConfigFile(dir, []string{"playwright.config", "playwright-ct.config"}, nodeConfigExts) {
		return true
	}
	if hasNpmDep(pkg, "@playwright/test") && hasTestFile(dir, playwrightTestSuffixes, true) {
		return true
	}
	return false
}

func (r *Playwright) Detect(workDir string) (bool, error) {
	return anyNodePackage(workDir, detectPlaywright)
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
