package runners

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// packageWalker lists the package directories (absolute) under dir.
type packageWalker func(dir string, recursive bool) ([]string, error)

// fileOwner returns the package directory (absolute) that runs file, if this
// runner runs it at all.
type fileOwner func(file string) (string, bool, error)

// discoverPackages applies the options' starting paths, test files and skips
// to one runner's package walk and returns Cwd-relative "./x" packages. With
// neither paths nor files, the whole Cwd is discovered.
func discoverPackages(opts TestOptions, walk packageWalker, owner fileOwner) ([]string, error) {
	var dirs []string
	roots := slices.Clone(opts.Paths)
	if len(roots) == 0 && opts.Files.Empty() {
		roots = []string{"."}
	}
	for _, root := range roots {
		found, err := walk(filepath.Join(opts.Cwd, root), opts.Recursive)
		if err != nil {
			return nil, err
		}
		dirs = append(dirs, found...)
	}
	files, err := opts.selectedFiles()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		dir, ok, err := owner(filepath.Join(opts.Cwd, file))
		if err != nil {
			return nil, err
		}
		if ok {
			dirs = append(dirs, dir)
		}
	}
	var packages []string
	for _, dir := range dirs {
		pkg := opts.packagePath(dir)
		if !slices.Contains(packages, pkg) && !opts.packageSkipped(pkg) {
			packages = append(packages, pkg)
		}
	}
	return packages, nil
}

// selectedFiles expands Files to Cwd-relative paths; nil without a file selection.
func (o TestOptions) selectedFiles() ([]string, error) {
	if o.Files.Empty() {
		return nil, nil
	}
	return o.Files.Expand(o.Cwd, o.SkipFiles)
}

func (o TestOptions) packagePath(dir string) string {
	rel, err := filepath.Rel(o.Cwd, dir)
	if err != nil {
		return dir
	}
	return "./" + filepath.ToSlash(rel)
}

// packageSkipped reports whether SkipFiles names pkg or a directory above it.
func (o TestOptions) packageSkipped(pkg string) bool {
	pkg = strings.TrimPrefix(pkg, "./")
	for _, pattern := range o.SkipFiles.Glob {
		pattern = strings.TrimSuffix(strings.TrimPrefix(filepath.ToSlash(pattern), "./"), "/")
		for _, candidate := range []string{pattern, pattern + "/**"} {
			if ok, _ := doublestar.Match(candidate, pkg); ok {
				return true
			}
		}
	}
	for _, exact := range o.SkipFiles.Exact {
		exact = strings.TrimSuffix(strings.TrimPrefix(exact, "./"), "/")
		if pkg == exact || strings.HasPrefix(pkg, exact+"/") {
			return true
		}
	}
	return false
}

// PackageFiles returns the selected test files inside pkg, relative to pkg,
// or nil when the whole package runs: no file selection, a starting path that
// covers pkg, or no selected file inside it. nested includes files in
// subdirectories (node packages); Go packages own only their own directory.
func (o TestOptions) PackageFiles(pkg string, nested bool) ([]string, error) {
	if o.Files.Empty() || o.pathCoversPackage(pkg) {
		return nil, nil
	}
	clean := strings.TrimPrefix(pkg, "./")
	files, err := o.selectedFiles()
	if err != nil {
		return nil, err
	}
	var inside []string
	for _, file := range files {
		rel, err := filepath.Rel(clean, file)
		if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
			continue
		}
		if !nested && strings.Contains(filepath.ToSlash(rel), "/") {
			continue
		}
		inside = append(inside, filepath.ToSlash(rel))
	}
	return inside, nil
}

// pathCoversPackage reports whether a starting path selects all of pkg.
func (o TestOptions) pathCoversPackage(pkg string) bool {
	clean := strings.TrimPrefix(pkg, "./")
	for _, path := range o.Paths {
		path = strings.TrimPrefix(path, "./")
		if path == "." || clean == path || (o.Recursive && strings.HasPrefix(clean, path+"/")) {
			return true
		}
	}
	return false
}

// ForPackages returns a copy of o that runs only packages.
func (o TestOptions) ForPackages(packages ...string) TestOptions {
	o.Packages = slices.Clone(packages)
	return o
}
