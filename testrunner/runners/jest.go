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

// detectJest reports whether dir looks like a Jest project root: presence
// of a jest config file, or a "jest" key in package.json.
func detectJest(dir string, pkg *pkgJSON) bool {
	if hasConfigFile(dir, []string{"jest.config"}, nodeConfigExts) {
		return true
	}
	if pkg != nil && pkg.Jest != nil {
		return true
	}
	return false
}

func (r *Jest) Detect(workDir string) (bool, error) {
	pkg, _ := readPackageJSON(workDir)
	if detectJest(workDir, pkg) {
		return true, nil
	}
	pkgs, err := walkNodePackages(workDir, detectJest)
	if err != nil {
		return false, err
	}
	return len(pkgs) > 0, nil
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

func (r *Jest) PackageHasTests(packagePath string) (bool, error) {
	dir := filepath.Join(r.workDir, packagePath)
	pkg, _ := readPackageJSON(dir)
	return detectJest(dir, pkg), nil
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

	process := exec.NewExec(cmd, args...).WithCwd(cwd)
	process.SucceedOnNonZero = true

	return &TestRun{
		Framework:  parsers.Jest,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

func (r *Jest) NormalizeFilePath(filePath string) string {
	return normalizeNodeFilePath(r.workDir, filePath)
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
