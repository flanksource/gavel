package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Jest runs jest tests via the native `--json --outputFile` reporter.
type Jest struct {
	workDir string
	parser  parsers.ResultParser
}

func NewJest(workDir string) *Jest {
	return &Jest{
		workDir: workDir,
		parser:  parsers.NewJestJSON(workDir, parsers.Jest),
	}
}

func (r *Jest) Name() parsers.Framework      { return parsers.Jest }
func (r *Jest) Parser() parsers.ResultParser { return r.parser }

// jestTestSuffixes lists the file-name suffixes Jest's default testMatch
// catches. A user who customizes testMatch will still detect via config.
var jestTestSuffixes = []string{".test.js", ".test.jsx", ".test.ts", ".test.tsx", ".test.mjs", ".test.cjs",
	".spec.js", ".spec.jsx", ".spec.ts", ".spec.tsx", ".spec.mjs", ".spec.cjs"}

// detectJest reports whether dir looks like a Jest project root. Two
// independent signals are enough on their own:
//
//  1. a dedicated jest config file (jest.config.*, .jestrc*),
//  2. a "jest" key in package.json.
//
// A dependency-only signal ("jest" in devDependencies) is paired with the
// presence of at least one *.test.*/*.spec.* file so a package that only
// declares Jest (no config, no tests yet) doesn't get launched into the
// runner just to fail with "No tests found".
func detectJest(dir string, pkg *pkgJSON) bool {
	if hasConfigFile(dir, []string{"jest.config", ".jestrc"}, nodeConfigExts) {
		return true
	}
	if pkg != nil && pkg.Jest != nil {
		return true
	}
	if hasNpmDep(pkg, "jest") && hasTestFile(dir, jestTestSuffixes, true) {
		return true
	}
	return false
}

func (r *Jest) Detect(workDir string) (bool, error) {
	return anyNodePackage(workDir, detectJest)
}

func (r *Jest) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		pkg, _ := readPackageJSON(workDir)
		if detectJest(workDir, pkg) {
			return []string{"."}, nil
		}
		return nil, nil
	}
	return walkNodePackages(workDir, detectJest)
}

func (r *Jest) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	cwd := filepath.Join(r.workDir, packagePath)
	reportPath := filepath.Join(r.workDir, ".jest", fmt.Sprintf("jest-report-%s-%d.json", sanitizePkgPath(packagePath), time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return nil, fmt.Errorf("create jest report dir: %w", err)
	}

	cmd, pre := detectPackageManager(cwd)
	args := append([]string{}, pre...)
	args = append(args, "jest", "--json", "--outputFile="+reportPath)
	args = append(args, extraArgs...)

	process := exec.NewExec(cmd, args...).WithCwd(cwd).WithProcessGroup()
	process.SucceedOnNonZero = true

	return &TestRun{
		Framework:  parsers.Jest,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

// sanitizePkgPath turns a gavel relative pkg path like "./apps/web" into a
// filename-safe slug for use in a report file name.
func sanitizePkgPath(p string) string {
	p = strings.TrimPrefix(p, "./")
	p = strings.ReplaceAll(p, "/", "-")
	if p == "" || p == "." {
		return "root"
	}
	return p
}
