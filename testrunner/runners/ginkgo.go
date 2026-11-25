package runners

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

var errGinkgoDetected = errors.New("ginkgo detected")

// Ginkgo discovers Ginkgo packages and runs them through go test.
type Ginkgo struct {
	workDir string
	parser  parsers.ResultParser
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

// GinkgoFlags are Ginkgo's own flags that map onto TestOptions. Ginkgo runs
// through go test, so go test's flags apply too.
var GinkgoFlags = append([]FlagSpec{
	{Names: []string{"label-filter", "ginkgo.label-filter"}, Kind: LabelFilterFlag},
	{Names: []string{"nodes"}, Kind: NodesFlag},
}, GoFlags...)

func (r *Ginkgo) Flags() []FlagSpec { return GinkgoFlags }

// Detect parses argv and reports Ginkgo when a test file under cwd or the
// starting paths imports it. Like go test we do not gate on go.mod.
func (r *Ginkgo) Detect(cwd string, argv []string) (*TestOptions, error) {
	return detect(r, cwd, argv, ginkgoFilesPresent, r.ownsFile)
}

// ownsFile: Ginkgo runs its spec files, and a Go source file of a Ginkgo package.
func (r *Ginkgo) ownsFile(file string) (string, bool, error) {
	if !strings.HasSuffix(file, ".go") {
		return "", false, nil
	}
	dir := filepath.Dir(file)
	if strings.HasSuffix(file, "_test.go") {
		return dir, hasGinkgoImports(file), nil
	}
	ok, err := dirHasGinkgoTests(dir, nil)
	return dir, ok, err
}

func ginkgoFilesPresent(workDir string) (bool, error) {
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

// DiscoverPackages returns the Ginkgo packages opts selects.
func (r *Ginkgo) DiscoverPackages(opts TestOptions) ([]string, error) {
	walk := func(dir string, recursive bool) ([]string, error) {
		return walkGoPackages(dir, recursive, func(pkgDir string) (bool, error) {
			return dirHasGinkgoTests(pkgDir, opts.Tags)
		})
	}
	return discoverPackages(opts, walk, r.ownsFile)
}

// BuildCommand builds a go test command with an injected Ginkgo reporter.
func (r *Ginkgo) BuildCommand(opts TestOptions) (*TestRun, error) {
	process, err := GoTestCommand(opts, opts.Packages)
	if err != nil {
		return nil, err
	}
	return &TestRun{
		Framework: parsers.Ginkgo,
		Package:   Package(strings.Join(opts.Packages, " ")),
		Parser:    r.parser,
		Process:   process,
	}, nil
}

func dirHasGinkgoTests(dir string, tags []string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
			matched, err := matchesBuildConstraints(dir, entry.Name(), tags)
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
