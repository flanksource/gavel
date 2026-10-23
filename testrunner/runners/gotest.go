package runners

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var (
	goBenchFuncRe     = regexp.MustCompile(`(?m)^func\s+Benchmark[A-Z_]\w*\s*\(`)
	errGoTestDetected = errors.New("go test detected")
)

// GoTest implements the test runner for go test.
type GoTest struct {
	workDir   string
	parser    parsers.ResultParser
	buildTags []string
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

// FocusArgs maps a common focus pattern to go test's native name filter.
func (r *GoTest) FocusArgs(pattern string) []string {
	return []string{"-run", pattern}
}

func (r *GoTest) BuildTagsArgs(tags []string) []string {
	return []string{"-tags=" + strings.Join(tags, ",")}
}

func (r *GoTest) SetBuildTags(tags []string) {
	r.buildTags = append([]string(nil), tags...)
}

// Detect checks if go test is used (looks for *_test.go files). We
// deliberately do NOT gate on go.mod presence — a nested repo without its
// own module, or a workDir inside a module whose go.mod sits above
// the nearest .git boundary, still has runnable test files and should be
// surfaced. `go test` emits a useful error at run time if no module exists.
func (r *GoTest) Detect(workDir string) (bool, error) {
	root, _ := filepath.Abs(workDir)
	err := utils.WalkGitIgnoredBounded(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && isGoIgnoredDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), "_test.go") {
			return errGoTestDetected
		}
		return nil
	})
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errGoTestDetected) {
		return true, nil
	}
	return false, err
}

// packageHasNonGinkgoTests checks if a package has at least one test file
// that is not a Ginkgo test (using the shared AST-based import check).
func (r *GoTest) packageHasNonGinkgoTests(pkgDir string) (bool, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			matched, err := matchesBuildConstraints(pkgDir, entry.Name(), r.buildTags)
			if err != nil {
				return false, err
			}
			if !matched {
				continue
			}
			path := filepath.Join(pkgDir, entry.Name())
			if !hasGinkgoImports(path) {
				return true, nil
			}
		}
	}
	return false, nil
}

// PackageHasBenchmarks reports whether packagePath (relative to workDir or absolute)
// contains any Benchmark* funcs in non-Ginkgo test files.
func (r *GoTest) PackageHasBenchmarks(packagePath string) bool {
	dir := packagePath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(r.workDir, packagePath)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if hasGinkgoImports(path) {
			continue
		}
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if goBenchFuncRe.Match(content) {
			return true
		}
	}
	return false
}

// PackageHasRunnableTests reports whether packagePath contains a runnable Test* func,
// including the one hosting a Ginkgo suite's RunSpecs. TestMain alone does not
// produce runnable tests, so it does not count.
func (r *GoTest) PackageHasRunnableTests(packagePath string) bool {
	dir := packagePath
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(r.workDir, packagePath)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	testMainOnlyRe := regexp.MustCompile(`(?m)^func\s+(Test\w+)\s*\(`)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
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
// When recursive is false, only the given directory is checked. No go.mod gate
// here for the same reason as Detect — `go test` surfaces the error itself if
// the package is not inside a module.
func (r *GoTest) DiscoverPackages(workDir string, recursive bool) ([]string, error) {
	if !recursive {
		hasTests, err := r.packageHasNonGinkgoTests(workDir)
		if err != nil {
			return nil, err
		}
		if hasTests {
			return []string{r.getRelativePath(workDir)}, nil
		}
		return nil, nil
	}

	var packages []string
	seen := make(map[string]bool)
	root, _ := filepath.Abs(workDir)

	err := utils.WalkGitIgnoredBounded(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path != root && isGoIgnoredDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(d.Name(), "_test.go") {
			pkgDir := filepath.Dir(path)
			if !seen[pkgDir] {
				seen[pkgDir] = true
				hasTests, err := r.packageHasNonGinkgoTests(pkgDir)
				if err != nil {
					return err
				}
				if hasTests {
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
	return r.packageHasNonGinkgoTests(dir)
}

// BuildCommand builds the go test command for a package.
func (r *GoTest) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	// Build command args: standard flags, then extra args, then package path
	args := []string{"test", "-json"}
	args = append(args, extraArgs...)
	args = append(args, packagePath)

	// Build command using exec.Process (but don't execute)
	process := exec.NewExec("go", args...).WithCwd(r.workDir).WithProcessGroup()
	process.SucceedOnNonZero = true // go test returns non-zero on test failures

	return &TestRun{
		Framework: parsers.GoTest,
		Package:   Package(packagePath),
		Parser:    r.parser,
		Process:   process,
	}, nil
}

// getRelativePath returns the relative path from workDir to the target directory.
func (r *GoTest) getRelativePath(dir string) string {
	if relPath, err := filepath.Rel(r.workDir, dir); err == nil {
		return "./" + filepath.ToSlash(relPath)
	}
	return dir
}
