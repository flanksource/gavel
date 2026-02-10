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

// Ginkgo implements the test runner for Ginkgo with --json-report.
type Ginkgo struct {
	workDir string
	parser  parsers.ResultParser
}

// NewGinkgo creates a new Ginkgo runner.
func NewGinkgo(workDir string) *Ginkgo {
	return &Ginkgo{
		workDir: workDir,
		parser:  parsers.NewGinkgoJSON(),
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
func (r *Ginkgo) Detect(workDir string) (bool, error) {
	var found bool

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || found {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			if r.hasGinkgoImports(path) {
				found = true
			}
		}

		return nil
	})

	return found, err
}

// DiscoverPackages returns all packages with Ginkgo tests.
func (r *Ginkgo) DiscoverPackages(workDir string) ([]string, error) {
	var packages []string
	seen := make(map[string]bool)

	err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), "_test.go") {
			if r.hasGinkgoImports(path) {
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
			if r.hasGinkgoImports(path) {
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
	process := exec.NewExec("go", args...).WithCwd(r.workDir)
	process.SucceedOnNonZero = true // ginkgo returns non-zero on test failures

	return &TestRun{
		Framework:  parsers.Ginkgo,
		Package:    Package(packagePath),
		Parser:     r.parser,
		Process:    process,
		ReportPath: reportPath,
	}, nil
}

// NormalizeFilePath makes file paths relative to workDir (exposed for orchestrator use).
func (r *Ginkgo) NormalizeFilePath(filePath string) string {
	return r.normalizeFilePath(filePath)
}

// hasGinkgoImports checks if a file imports ginkgo.
func (r *Ginkgo) hasGinkgoImports(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	text := string(content)
	return strings.Contains(text, `"github.com/onsi/ginkgo`) ||
		strings.Contains(text, `"github.com/onsi/ginkgo/v2`) ||
		strings.Contains(text, ". \"github.com/onsi/ginkgo") ||
		strings.Contains(text, ". \"github.com/onsi/ginkgo/v2")
}

// getRelativePath returns the relative path from workDir to the target directory.
func (r *Ginkgo) getRelativePath(dir string) string {
	if relPath, err := filepath.Rel(r.workDir, dir); err == nil {
		return "./" + filepath.ToSlash(relPath)
	}
	return dir
}

// normalizeFilePath makes file paths relative to workDir
func (r *Ginkgo) normalizeFilePath(filePath string) string {
	// If path is already relative and not starting with .., return as-is
	if !filepath.IsAbs(filePath) && !strings.HasPrefix(filePath, "..") {
		return filePath
	}

	// Try to make it relative to workDir
	if relPath, err := filepath.Rel(r.workDir, filePath); err == nil {
		return relPath
	}

	// If that fails, return the original path
	return filePath
}
