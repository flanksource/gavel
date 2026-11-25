package runners

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var (
	goBenchFuncRe     = regexp.MustCompile(`(?m)^func\s+Benchmark[A-Z_]\w*\s*\(`)
	errGoTestDetected = errors.New("go test detected")
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

// GoFlags are go test's own flags that map onto TestOptions.
var GoFlags = []FlagSpec{
	{Names: []string{"tags"}, Kind: TagsFlag},
	{Names: []string{"bench"}, Kind: BenchFlag},
	{Names: []string{"p"}, Kind: ParallelFlag},
}

func (r *GoTest) Flags() []FlagSpec { return GoFlags }

// Detect parses argv and reports go test when a *_test.go file exists under
// cwd or the starting paths.
func (r *GoTest) Detect(cwd string, argv []string) (*TestOptions, error) {
	return detect(r, cwd, argv, goTestFilesPresent, r.ownsFile)
}

// ownsFile: go test runs a Go source file's package, and a test file that is
// not a Ginkgo spec file.
func (r *GoTest) ownsFile(file string) (string, bool, error) {
	if !strings.HasSuffix(file, ".go") {
		return "", false, nil
	}
	if strings.HasSuffix(file, "_test.go") && hasGinkgoImports(file) {
		return "", false, nil
	}
	return filepath.Dir(file), true, nil
}

// goTestFilesPresent looks for *_test.go files. We deliberately do NOT gate
// on go.mod presence — a nested repo without its own module, or a workDir
// inside a module whose go.mod sits above the nearest .git boundary, still
// has runnable test files and should be surfaced. `go test` emits a useful
// error at run time if no module exists.
func goTestFilesPresent(workDir string) (bool, error) {
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
func packageHasNonGinkgoTests(pkgDir string, tags []string) (bool, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			matched, err := matchesBuildConstraints(pkgDir, entry.Name(), tags)
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
func (r *GoTest) DiscoverPackages(opts TestOptions) ([]string, error) {
	walk := func(dir string, recursive bool) ([]string, error) {
		return walkGoPackages(dir, recursive, func(pkgDir string) (bool, error) {
			return packageHasNonGinkgoTests(pkgDir, opts.Tags)
		})
	}
	return discoverPackages(opts, walk, r.ownsFile)
}

// walkGoPackages returns the absolute package directories at or below dir for
// which has reports tests, skipping directories the go tool ignores.
func walkGoPackages(dir string, recursive bool, has func(pkgDir string) (bool, error)) ([]string, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if !recursive {
		ok, err := has(root)
		if err != nil || !ok {
			return nil, err
		}
		return []string{root}, nil
	}
	var packages []string
	seen := make(map[string]bool)
	err = utils.WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && isGoIgnoredDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		pkgDir := filepath.Dir(path)
		if seen[pkgDir] {
			return nil
		}
		seen[pkgDir] = true
		ok, err := has(pkgDir)
		if err != nil || !ok {
			return err
		}
		packages = append(packages, pkgDir)
		return nil
	})
	return packages, err
}

// BuildCommand builds the go test command for opts.Packages.
func (r *GoTest) BuildCommand(opts TestOptions) (*TestRun, error) {
	process, err := GoTestCommand(opts, nil)
	if err != nil {
		return nil, err
	}
	return &TestRun{
		Framework: parsers.GoTest,
		Package:   Package(strings.Join(opts.Packages, " ")),
		Parser:    r.parser,
		Process:   process,
	}, nil
}
