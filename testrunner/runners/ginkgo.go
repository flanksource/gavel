package runners

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var errGinkgoDetected = errors.New("ginkgo detected")

// Ginkgo implements the test runner for Ginkgo with --json-report.
type Ginkgo struct {
	workDir string
	parser  parsers.ResultParser
}

// NewGinkgo creates a new Ginkgo runner.
func NewGinkgo(workDir string) *Ginkgo {
	return &Ginkgo{
		workDir: workDir,
		parser:  parsers.NewGinkgoJSON(workDir),
	}
}

// Name returns the framework name.
func (r *Ginkgo) Name() parsers.Framework {
	return parsers.Ginkgo
}

// Parser returns the result parser.
func (r *Ginkgo) Parser() parsers.ResultParser {
	return r.parser
}

// Detect checks if Ginkgo is used (looks for ginkgo imports in test files).
// Like GoTest.Detect we do not gate on go.mod; we bail out early via a
// sentinel error on the first hit so we don't keep walking.
func (r *Ginkgo) Detect(workDir string) (bool, error) {
	err := utils.WalkGitIgnoredBounded(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") && matchesBuildContext(path) && hasGinkgoImports(path) {
			return errGinkgoDetected
		}
		return nil
	})
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errGinkgoDetected) {
		return true, nil
	}
	return false, err
}

// DiscoverPackages returns packages with Ginkgo tests.
// When recursive is false, only the given directory is checked. No go.mod
// gate (same reasoning as Detect).
func (r *Ginkgo) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		if r.dirHasGinkgoTests(workDir) {
			return []string{r.getRelativePath(workDir)}, nil
		}
		return nil, nil
	}

	var packages []string
	seen := make(map[string]bool)

	err := utils.WalkGitIgnoredBounded(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") && matchesBuildContext(path) {
			if hasGinkgoImports(path) {
				pkgDir := filepath.Dir(path)
				if !seen[pkgDir] {
					seen[pkgDir] = true
					relPath := r.getRelativePath(pkgDir)
					packages = append(packages, relPath)
				}
			}
		}

		return nil
	})

	return packages, err
}

// PackageHasTests checks if a package has Ginkgo tests.
func (r *Ginkgo) PackageHasTests(packagePath string) (bool, error) {
	dir := filepath.Join(r.workDir, packagePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			path := filepath.Join(dir, entry.Name())
			if hasGinkgoImports(path) {
				return true, nil
			}
		}
	}

	return false, nil
}

// BuildCommand builds the ginkgo command for a package.
func (r *Ginkgo) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	reportPath := filepath.Join(".ginkgo", fmt.Sprintf("ginkgo-report-%s-%d.json", strings.ReplaceAll(packagePath, "/", "-"), time.Now().UnixNano()))
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create report directory: %w", err)
	}

	// Build command args: standard flags, then extra args, then package path
	args := []string{
		"run",
		"github.com/onsi/ginkgo/v2/ginkgo",
		fmt.Sprintf("--json-report=%s", reportPath),
		"--show-node-events=false",
	}
	args = append(args, extraArgs...)
	args = append(args, packagePath)

	// Build command using exec.Process (but don't execute)
	process := exec.NewExec("go", args...).WithCwd(r.workDir).WithProcessGroup()
	process.SucceedOnNonZero = true // ginkgo returns non-zero on test failures

	return &TestRun{
		Framework:  parsers.Ginkgo,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

func (r *Ginkgo) dirHasGinkgoTests(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			if hasGinkgoImports(filepath.Join(dir, entry.Name())) {
				return true
			}
		}
	}
	return false
}

// getRelativePath returns the relative path from workDir to the target directory.
func (r *Ginkgo) getRelativePath(dir string) string {
	if relPath, err := filepath.Rel(r.workDir, dir); err == nil {
		return "./" + filepath.ToSlash(relPath)
	}
	return dir
}
