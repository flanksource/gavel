package runners

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
)

// nodeConfigExts lists the extensions we accept for node tool config files.
// Order matters only for readability; hasConfigFile stat-checks each.
var nodeConfigExts = []string{"js", "ts", "mjs", "cjs", "json"}

// nodeSkipDirs names directories that should not be descended into when
// scanning for node packages. These commonly contain a `package.json` (for
// publishing metadata or vendored deps) but never a test root. node_modules
// is always skipped; the rest are build/cache outputs.
var nodeSkipDirs = map[string]bool{
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"coverage":     true,
	".next":        true,
	".nuxt":        true,
	".turbo":       true,
	".vercel":      true,
	".cache":       true,
	"out":          true,
}

// errNodePackageFound is a sentinel used by anyNodePackage to bail out of
// filepath.WalkDir on first hit.
var errNodePackageFound = errors.New("node package found")

// pkgJSON captures the subset of package.json we care about.
type pkgJSON struct {
	Jest            json.RawMessage   `json:"jest,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Workspaces      json.RawMessage   `json:"workspaces,omitempty"`
}

// readPackageJSON reads <dir>/package.json. Missing returns (nil, false).
// Malformed JSON returns (nil, false) AND logs at V(2) so a user with a
// broken package.json can diagnose why detection thinks the dir is empty.
func readPackageJSON(dir string) (*pkgJSON, bool) {
	path := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			logger.V(2).Infof("node detect: read %s: %v", path, err)
		}
		return nil, false
	}
	var pkg pkgJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		logger.V(2).Infof("node detect: parse %s: %v", path, err)
		return nil, false
	}
	return &pkg, true
}

// hasConfigFile reports whether any <basename>.<ext> exists in dir, or
// whether bare <basename> exists (useful for dot-rc style configs like
// ".jestrc").
func hasConfigFile(dir string, basenames, exts []string) bool {
	for _, base := range basenames {
		if _, err := os.Stat(filepath.Join(dir, base)); err == nil {
			return true
		}
		for _, ext := range exts {
			if _, err := os.Stat(filepath.Join(dir, base+"."+ext)); err == nil {
				return true
			}
		}
	}
	return false
}

// hasNpmDep reports whether pkg declares name in dependencies or devDependencies.
func hasNpmDep(pkg *pkgJSON, name string) bool {
	if pkg == nil {
		return false
	}
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	_, ok := pkg.DevDependencies[name]
	return ok
}

// DetectPackageManager returns the command + prefix args used to invoke a
// node tool from dir. It walks up looking for lockfiles and falls back to npx.
//   - pnpm-lock.yaml    → ("pnpm", ["exec"])
//   - yarn.lock         → ("yarn", nil)
//   - package-lock.json → ("npm", ["exec", "--"])
//   - bun.lockb         → ("bun", ["x"])
//   - none              → ("npx", nil)
func DetectPackageManager(dir string) (string, []string) {
	cur, _ := filepath.Abs(dir)
	for {
		if _, err := os.Stat(filepath.Join(cur, "pnpm-lock.yaml")); err == nil {
			return "pnpm", []string{"exec"}
		}
		if _, err := os.Stat(filepath.Join(cur, "yarn.lock")); err == nil {
			return "yarn", nil
		}
		if _, err := os.Stat(filepath.Join(cur, "bun.lockb")); err == nil {
			return "bun", []string{"x"}
		}
		if _, err := os.Stat(filepath.Join(cur, "package-lock.json")); err == nil {
			return "npm", []string{"exec", "--"}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "npx", nil
		}
		cur = parent
	}
}

// walkNodePackages finds every directory under root that (a) contains a
// package.json and (b) satisfies detect(dir, pkgJSON). Returns gavel-style
// package directories. Skips node_modules/ + build-output dirs
// (nodeSkipDirs) defensively, and honors .gitignore.
func walkNodePackages(root string, detect func(string, *pkgJSON) bool) ([]string, error) {
	var out []string
	seen := map[string]bool{}

	err := utils.WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && nodeSkipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() != "package.json" {
			return nil
		}
		pkgDir := filepath.Dir(path)
		if seen[pkgDir] {
			return nil
		}
		seen[pkgDir] = true
		pkg, _ := readPackageJSON(pkgDir)
		if detect(pkgDir, pkg) {
			out = append(out, pkgDir)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// anyNodePackage reports whether any package under root satisfies detect.
// Bails out on first match via a sentinel error so Detect() on Node runners
// doesn't walk the whole tree just to answer a boolean.
func anyNodePackage(root string, detect func(string, *pkgJSON) bool) (bool, error) {
	if pkg, ok := readPackageJSON(root); ok && detect(root, pkg) {
		return true, nil
	}
	err := utils.WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && nodeSkipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() != "package.json" {
			return nil
		}
		pkgDir := filepath.Dir(path)
		pkg, _ := readPackageJSON(pkgDir)
		if detect(pkgDir, pkg) {
			return errNodePackageFound
		}
		return nil
	})
	if err == nil {
		return false, nil
	}
	if errors.Is(err, errNodePackageFound) {
		return true, nil
	}
	return false, err
}

// hasTestFile reports whether dir contains at least one file whose name
// matches any of the given suffixes (case-insensitive). It does NOT recurse
// by default unless deep is true; node tools' `testMatch`/`testDir` usually
// points somewhere specific but a recursive walk catches `src/foo.test.ts`
// style layouts that 99% of real projects use.
func hasTestFile(dir string, suffixes []string, deep bool) bool {
	check := func(name string) bool {
		lower := strings.ToLower(name)
		for _, s := range suffixes {
			if strings.HasSuffix(lower, s) {
				return true
			}
		}
		return false
	}
	if !deep {
		entries, err := os.ReadDir(dir)
		if err != nil {
			logger.V(2).Infof("node detect: read dir %s: %v", dir, err)
			return false
		}
		for _, e := range entries {
			if !e.IsDir() && check(e.Name()) {
				return true
			}
		}
		return false
	}
	found := false
	// Bounded: a test file inside a nested checkout says nothing about whether
	// this project has node tests.
	_ = utils.WalkGitIgnoredBounded(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != dir && nodeSkipDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if check(d.Name()) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// discoverNodePackages applies opts to a node runner's package detection.
func discoverNodePackages(opts TestOptions, detect func(string, *pkgJSON) bool, suffixes []string) ([]string, error) {
	walk := func(dir string, recursive bool) ([]string, error) {
		var found []string
		if !recursive {
			if pkg, _ := readPackageJSON(dir); detect(dir, pkg) {
				found = []string{dir}
			}
		} else {
			var err error
			if found, err = walkNodePackages(dir, detect); err != nil {
				return nil, err
			}
		}
		if len(found) > 0 {
			return found, nil
		}
		pkg, ok, err := nodeEnclosingPackage(opts.Cwd, dir, detect, suffixes)
		if err != nil || !ok {
			return nil, err
		}
		return []string{pkg}, nil
	}
	return discoverPackages(opts, walk, nodeFileOwner(opts.Cwd, detect, suffixes))
}

// nodePresent reports whether a node runner has tests under dir: a detected
// package at or below it, or a detected package enclosing it.
func nodePresent(cwd string, detect func(string, *pkgJSON) bool, suffixes []string) func(string) (bool, error) {
	return func(dir string) (bool, error) {
		if found, err := anyNodePackage(dir, detect); err != nil || found {
			return found, err
		}
		_, ok, err := nodeEnclosingPackage(cwd, dir, detect, suffixes)
		return ok, err
	}
}

// nodeEnclosingPackage returns the detected package that encloses a starting
// directory holding the runner's test files, so `gavel test app/src/feature`
// runs that folder through app.
func nodeEnclosingPackage(cwd, dir string, detect func(string, *pkgJSON) bool, suffixes []string) (string, bool, error) {
	if !hasTestFile(dir, suffixes, true) {
		return "", false, nil
	}
	return nodePackageAbove(cwd, dir, detect)
}

// nodeFileOwner finds the nearest package directory at or above a test file,
// up to cwd, that the runner detects.
func nodeFileOwner(cwd string, detect func(string, *pkgJSON) bool, suffixes []string) fileOwner {
	return func(file string) (string, bool, error) {
		if !hasTestSuffix(file, suffixes) {
			return "", false, nil
		}
		return nodePackageAbove(cwd, filepath.Dir(file), detect)
	}
}

// nodePackageAbove finds the nearest package directory at or above dir, up to
// cwd, and reports whether the runner detects it.
func nodePackageAbove(cwd, dir string, detect func(string, *pkgJSON) bool) (string, bool, error) {
	root, err := filepath.Abs(cwd)
	if err != nil {
		return "", false, err
	}
	if dir, err = filepath.Abs(dir); err != nil {
		return "", false, err
	}
	for ; ; dir = filepath.Dir(dir) {
		if pkg, ok := readPackageJSON(dir); ok {
			return dir, detect(dir, pkg), nil
		}
		if dir == root || filepath.Dir(dir) == dir {
			return "", false, nil
		}
	}
}

// nodePathFilters returns the starting paths inside pkg, relative to it, that
// narrow its run; nil when a starting path covers the whole package.
func nodePathFilters(opts TestOptions, pkg string) ([]string, error) {
	if opts.pathCoversPackage(pkg) {
		return nil, nil
	}
	pkgDir, err := filepath.Abs(filepath.Join(opts.Cwd, pkg))
	if err != nil {
		return nil, err
	}
	var filters []string
	for _, path := range opts.Paths {
		dir, err := filepath.Abs(filepath.Join(opts.Cwd, path))
		if err != nil {
			return nil, err
		}
		owner, _, err := nodePackageAbove(opts.Cwd, dir, func(string, *pkgJSON) bool { return true })
		if err != nil {
			return nil, err
		}
		if owner != pkgDir {
			continue
		}
		rel, err := filepath.Rel(pkgDir, dir)
		if err != nil {
			return nil, err
		}
		filters = append(filters, filepath.ToSlash(rel))
	}
	return filters, nil
}

// nodeFlagSet names one node runner's spellings of the shared options.
type nodeFlagSet struct {
	framework string
	focus     string   // flag taking the name pattern
	skip      string   // flag excluding a name pattern; "" = unsupported
	parallel  string   // flag taking the worker count
	failFast  string   // flag (with value) stopping at the first failure
	suffixes  []string // test file suffixes the runner runs
}

// nodeRunnerArgs renders opts for one node package: name filters, workers,
// fail-fast, the package's selected test files, then the extra arguments.
func nodeRunnerArgs(opts TestOptions, pkg string, flags nodeFlagSet) ([]string, error) {
	if len(opts.Packages) != 1 {
		return nil, fmt.Errorf("%s runs one package at a time, got %v", flags.framework, opts.Packages)
	}
	if opts.LabelFilter != "" {
		return nil, UnsupportedOptionError{Framework: flags.framework, Option: "label filters"}
	}
	if opts.Bench != "" {
		return nil, UnsupportedOptionError{Framework: flags.framework, Option: "benchmarks"}
	}
	var args []string
	if !opts.Focus.Empty() {
		focus, err := opts.Focus.Pattern()
		if err != nil {
			return nil, err
		}
		args = append(args, flags.focus, focus)
	}
	if !opts.Skip.Empty() {
		if flags.skip == "" {
			return nil, UnsupportedOptionError{Framework: flags.framework, Option: "skipping tests by name"}
		}
		skip, err := opts.Skip.Pattern()
		if err != nil {
			return nil, err
		}
		args = append(args, flags.skip, skip)
	}
	if opts.Parallel > 0 {
		args = append(args, fmt.Sprintf("%s=%d", flags.parallel, opts.Parallel))
	}
	if opts.FailFast {
		args = append(args, flags.failFast)
	}
	files, err := opts.PackageFiles(pkg, true)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if hasTestSuffix(file, flags.suffixes) {
			args = append(args, file)
		}
	}
	filters, err := nodePathFilters(opts, pkg)
	if err != nil {
		return nil, err
	}
	args = append(args, filters...)
	return append(args, opts.ExtraArgs...), nil
}

func firstPackage(opts TestOptions) string {
	if len(opts.Packages) == 0 {
		return ""
	}
	return opts.Packages[0]
}

func hasTestSuffix(file string, suffixes []string) bool {
	lower := strings.ToLower(file)
	for _, suffix := range suffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}
