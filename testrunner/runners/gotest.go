package runners

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var (
	goTestFuncRe  = regexp.MustCompile(`(?m)^func\s+Test[A-Z_]\w*\s*\(`)
	goBenchFuncRe = regexp.MustCompile(`(?m)^func\s+Benchmark[A-Z_]\w*\s*\(`)
)

// GoTest implements the test runner for go test.
type GoTest struct {
	workDir string
	parser  parsers.ResultParser
}

// NewGoTest creates a new Go test runner.
func NewGoTest(workDir string) *GoTest {
	return &GoTest{
		workDir: workDir,
		parser:  parsers.NewGoTestJSON(workDir),
	}
}

// Name returns the framework name.
func (r *GoTest) Name() parsers.Framework {
	return parsers.GoTest
}

// Parser returns the result parser.
func (r *GoTest) Parser() parsers.ResultParser {
	return r.parser
}

// Detect checks if go test is used (looks for *_test.go files).
func (r *GoTest) Detect(workDir string) (bool, error) {
	matches, err := doublestar.Glob(os.DirFS(workDir), "**/*_test.go")
	if err != nil {
		return false, err
	}
	return len(matches) > 0, nil
}

// hasGinkgoImports checks if a test file imports Ginkgo
func (r *GoTest) hasGinkgoImports(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	text := string(content)
	return strings.Contains(text, `"github.com/onsi/ginkgo`) ||
		strings.Contains(text, `"github.com/onsi/ginkgo/v2`) ||
		strings.Contains(text, `. "github.com/onsi/ginkgo`) ||
		strings.Contains(text, `. "github.com/onsi/ginkgo/v2`)
}

// packageHasNonGinkgoTests checks if a package has at least one test file that is not a Ginkgo test
func (r *GoTest) packageHasNonGinkgoTests(pkgDir string) bool {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			path := filepath.Join(pkgDir, entry.Name())
			if !r.hasGinkgoImports(path) {
				return true
			}
		}
	}
	return false
}

// inspectPackage scans a package directory and reports whether it contains
// non-Ginkgo Test* functions and whether it contains Benchmark* functions.
// A file that imports Ginkgo is ignored entirely (those are Ginkgo's responsibility).
func (r *GoTest) inspectPackage(pkgDir string) (hasTests bool, hasBench bool) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false, false
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, entry.Name())
		if r.hasGinkgoImports(path) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if !hasTests && goTestFuncRe.Match(content) {
			hasTests = true
		}
		if !hasBench && goBenchFuncRe.Match(content) {
			hasBench = true
		}
		if hasTests && hasBench {
			return
		}
	}
	return
}

// PackageHasBenchmarks reports whether packagePath (relative to workDir or absolute)
// contains any Benchmark* funcs in non-Ginkgo test files.
func (r *GoTest) PackageHasBenchmarks(packagePath string) bool {
	dir := packagePath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(r.workDir, packagePath)
	}
	_, hasBench := r.inspectPackage(dir)
	return hasBench
}

// PackageHasGoTests reports whether packagePath contains any non-Ginkgo Test* funcs.
// Note: TestMain does not count (matches `Test[A-Z_]` — TestMain has an 'M' next which
// is uppercase and therefore matches; we explicitly exclude it here).
func (r *GoTest) PackageHasGoTests(packagePath string) bool {
	dir := packagePath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(r.workDir, packagePath)
	}
	hasTests, _ := r.inspectPackage(dir)
	if !hasTests {
		return false
	}
	return r.hasTestsBeyondTestMain(dir)
}

// hasTestsBeyondTestMain returns true if there's at least one Test* function
// other than TestMain. TestMain alone does not produce runnable tests.
func (r *GoTest) hasTestsBeyondTestMain(pkgDir string) bool {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false
	}
	testMainOnlyRe := regexp.MustCompile(`(?m)^func\s+(Test\w+)\s*\(`)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(pkgDir, entry.Name())
		if r.hasGinkgoImports(path) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, m := range testMainOnlyRe.FindAllSubmatch(content, -1) {
			if string(m[1]) != "TestMain" {
				return true
			}
		}
	}
	return false
}

// DiscoverPackages returns packages with go test files (excluding Ginkgo-only packages).
// When recursive is false, only the given directory is checked.
func (r *GoTest) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		if r.packageHasNonGinkgoTests(workDir) {
			return []string{r.getRelativePath(workDir)}, nil
		}
		return nil, nil
	}

	var packages []string
	seen := make(map[string]bool)

	err := utils.WalkGitIgnored(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			pkgDir := filepath.Dir(path)
			if !seen[pkgDir] {
				seen[pkgDir] = true
				if r.packageHasNonGinkgoTests(pkgDir) {
					relPath := r.getRelativePath(pkgDir)
					packages = append(packages, relPath)
				}
			}
		}

		return nil
	})

	return packages, err
}

// PackageHasTests checks if a package has non-Ginkgo go test files.
func (r *GoTest) PackageHasTests(packagePath string) (bool, error) {
	dir := filepath.Join(r.workDir, packagePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			path := filepath.Join(dir, entry.Name())
			// Return true if this file doesn't import Ginkgo
			if !r.hasGinkgoImports(path) {
				return true, nil
			}
		}
	}

	return false, nil
}

// BuildCommand builds the go test command for a package.
func (r *GoTest) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	// Build command args: standard flags, then extra args, then package path
	args := []string{"test", "-json"}
	args = append(args, extraArgs...)
	args = append(args, packagePath)

	// Build command using exec.Process (but don't execute)
	process := exec.NewExec("go", args...).WithCwd(r.workDir)
	process.SucceedOnNonZero = true // go test returns non-zero on test failures

	return &TestRun{
		Framework: parsers.GoTest,
		Package:   Package(packagePath),
		Parser:    r.parser,
		Process:   process,
	}, nil
}

// NormalizeFilePath makes file paths relative to workDir (exposed for orchestrator use).
func (r *GoTest) NormalizeFilePath(filePath string) string {
	return r.normalizeFilePath(filePath)
}

// getRelativePath returns the relative path from workDir to the target directory.
func (r *GoTest) getRelativePath(dir string) string {
	if relPath, err := filepath.Rel(r.workDir, dir); err == nil {
		return "./" + filepath.ToSlash(relPath)
	}
	return dir
}

// normalizeFilePath makes file paths relative to workDir
func (r *GoTest) normalizeFilePath(filePath string) string {
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
