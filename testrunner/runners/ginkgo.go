package runners

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var errGinkgoDetected = errors.New("ginkgo detected")

// Ginkgo discovers Ginkgo packages and runs them through go test.
type Ginkgo struct {
	workDir   string
	parser    parsers.ResultParser
	buildTags []string
}

// NewGinkgo creates a new Ginkgo runner.
func NewGinkgo(workDir string) *Ginkgo {
	return &Ginkgo{
		workDir: workDir,
		parser:  parsers.NewGinkgoGoJSON(workDir),
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

// FocusArgs maps a common focus pattern to Ginkgo's native focus filter.
func (r *Ginkgo) FocusArgs(pattern string) []string {
	return []string{"--focus", pattern}
}

func (r *Ginkgo) BuildTagsArgs(tags []string) []string {
	return []string{"--tags=" + strings.Join(tags, ",")}
}

func (r *Ginkgo) SetBuildTags(tags []string) {
	r.buildTags = append([]string(nil), tags...)
}

// Detect checks if Ginkgo is used (looks for ginkgo imports in test files).
// Like GoTest.Detect we do not gate on go.mod; we bail out early via a
// sentinel error on the first hit so we don't keep walking.
func (r *Ginkgo) Detect(workDir string) (bool, error) {
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
		if strings.HasSuffix(d.Name(), "_test.go") && hasGinkgoImports(path) {
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
		hasTests, err := r.dirHasGinkgoTests(workDir)
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
			matched, err := matchesBuildConstraints(filepath.Dir(path), d.Name(), r.buildTags)
			if err != nil {
				return err
			}
			if matched && hasGinkgoImports(path) {
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
	return r.dirHasGinkgoTests(dir)
}

// BuildCommand builds a go test command with an injected Ginkgo reporter.
func (r *Ginkgo) BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error) {
	overlay, err := BuildGinkgoOverlay(r.workDir, []string{packagePath}, r.buildTags)
	if err != nil {
		return nil, err
	}
	mapped, err := GinkgoGoTestArgs(extraArgs)
	if err != nil {
		return nil, err
	}
	args := []string{"test", "-json", "-v", "-overlay=" + overlay}
	var ginkgoArgs []string
	for _, arg := range mapped {
		if strings.HasPrefix(arg, "-ginkgo.") {
			ginkgoArgs = append(ginkgoArgs, arg)
		} else {
			args = append(args, arg)
		}
	}
	args = append(args, packagePath)
	if len(ginkgoArgs) > 0 {
		args = append(args, "-args")
		args = append(args, ginkgoArgs...)
	}

	process := exec.NewExec("go", args...).WithCwd(r.workDir).WithProcessGroup()
	process.SucceedOnNonZero = true

	return &TestRun{
		Framework: parsers.Ginkgo,
		Package:   Package(packagePath),
		Parser:    r.parser,
		Process:   process,
	}, nil
}

func (r *Ginkgo) dirHasGinkgoTests(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			matched, err := matchesBuildConstraints(dir, entry.Name(), r.buildTags)
			if err != nil {
				return false, err
			}
			if matched && hasGinkgoImports(filepath.Join(dir, entry.Name())) {
				return true, nil
			}
		}
	}
	return false, nil
}

// getRelativePath returns the relative path from workDir to the target directory.
func (r *Ginkgo) getRelativePath(dir string) string {
	if relPath, err := filepath.Rel(r.workDir, dir); err == nil {
		return "./" + filepath.ToSlash(relPath)
	}
	return dir
}
