package changegraph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Pkg is a subset of the JSON emitted by `go list -json -deps ./...` — only
// the fields gavel's change graph needs. See `go help list` for the full shape.
type Pkg struct {
	ImportPath      string
	Dir             string
	Module          *Module
	Standard        bool
	Goroot          bool
	GoFiles         []string
	CgoFiles        []string
	CompiledGoFiles []string
	TestGoFiles     []string
	XTestGoFiles    []string
	EmbedFiles      []string
	TestEmbedFiles  []string
	XTestEmbedFiles []string
	SFiles          []string
	HFiles          []string
	CFiles          []string
	CXXFiles        []string
	IgnoredGoFiles  []string
	Imports         []string
	Deps            []string
	TestImports     []string
	XTestImports    []string
}

// Module is a subset of the module info inside a `go list -json` package.
type Module struct {
	Path      string
	Dir       string
	GoVersion string
}

// Graph is an in-memory package graph built from `go list -json -deps ./...`.
// It is keyed by import path and carries a reverse-dep index for fast
// affected-set BFS.
type Graph struct {
	// WorkDir is the absolute directory the graph was loaded from.
	WorkDir string
	// ModuleRoot is the absolute path to the module root (where go.mod lives).
	// Empty if no module context.
	ModuleRoot string

	// Packages is keyed by import path.
	Packages map[string]*Pkg

	// byDir maps absolute package directory → import path.
	byDir map[string]string

	// reverseDeps[ip] = set of import paths that import ip (directly or via
	// test imports). Populated once at Load time.
	reverseDeps map[string]map[string]struct{}
}

// Load runs `go list -json -deps ./...` in workDir and decodes the result.
func Load(workDir string) (*Graph, error) {
	absWork, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("abs workDir: %w", err)
	}

	cmd := exec.Command("go", "list", "-json", "-deps", "./...")
	cmd.Dir = absWork
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list -json -deps ./...: %w\n%s", err, stderr.String())
	}

	g := &Graph{
		WorkDir:     absWork,
		Packages:    map[string]*Pkg{},
		byDir:       map[string]string{},
		reverseDeps: map[string]map[string]struct{}{},
	}

	dec := json.NewDecoder(&stdout)
	for dec.More() {
		var p Pkg
		if err := dec.Decode(&p); err != nil {
			return nil, fmt.Errorf("decode go list output: %w", err)
		}
		// Stdlib/GOROOT packages are opaque leaves — we don't need their file
		// lists, just their existence for dep resolution.
		g.Packages[p.ImportPath] = &p
		if p.Dir != "" {
			g.byDir[p.Dir] = p.ImportPath
		}
		if p.Module != nil && p.Module.Dir != "" && g.ModuleRoot == "" {
			g.ModuleRoot = p.Module.Dir
		}
	}

	g.buildReverseIndex()
	return g, nil
}

// buildReverseIndex computes reverseDeps from Imports / TestImports /
// XTestImports on every package. Deps (the transitive closure) is not used
// here because we only want the direct edges — BFS will expand them.
func (g *Graph) buildReverseIndex() {
	add := func(importer, imported string) {
		if imported == importer {
			return
		}
		s, ok := g.reverseDeps[imported]
		if !ok {
			s = map[string]struct{}{}
			g.reverseDeps[imported] = s
		}
		s[importer] = struct{}{}
	}

	for ip, p := range g.Packages {
		for _, dep := range p.Imports {
			add(ip, dep)
		}
		for _, dep := range p.TestImports {
			add(ip, dep)
		}
		for _, dep := range p.XTestImports {
			add(ip, dep)
		}
	}
}

// PackageByDir looks up a package by its absolute directory.
func (g *Graph) PackageByDir(absDir string) (*Pkg, bool) {
	ip, ok := g.byDir[absDir]
	if !ok {
		return nil, false
	}
	return g.Packages[ip], true
}

// AffectedPackages returns the set of import paths affected by fs (a set of
// workdir-relative file paths). A package is affected if any of its declared
// files is in fs, or if it transitively imports such a package. Modifying a
// root-level file like go.mod or a file outside any package marks every
// non-stdlib package as affected.
func (g *Graph) AffectedPackages(fs FileSet) []string {
	if g == nil || len(fs) == 0 {
		return nil
	}

	dirty := map[string]struct{}{}
	rootBust := false

	for rel := range fs {
		absFile := filepath.Join(g.WorkDir, filepath.FromSlash(rel))

		// Root-level mod files mark everything dirty.
		base := filepath.Base(absFile)
		if base == "go.mod" || base == "go.sum" || base == "go.work" || base == "go.work.sum" {
			rootBust = true
			continue
		}

		// Resolve to the containing package directory.
		pkgDir := filepath.Dir(absFile)
		pkg, ok := g.PackageByDir(pkgDir)
		if !ok {
			// File belongs to no known package — could be a fixture under
			// testdata/, a doc file, or a subdir that `go list` skipped
			// (e.g. /internal/testdata). Walk upward looking for an ancestor
			// directory that IS a package; a file under pkg/foo/testdata/x
			// should mark pkg/foo dirty.
			if ip := g.resolveAncestorPackage(pkgDir); ip != "" {
				dirty[ip] = struct{}{}
				continue
			}
			// No containing package found — assume this is a cross-cutting
			// config file (Makefile, .golangci.yml, etc.). Bust root.
			rootBust = true
			continue
		}

		// File is directly owned by a package, but `go list` may report files
		// that don't affect compilation (e.g. IgnoredGoFiles with a build tag
		// gavel doesn't care about). We still mark the package dirty: the
		// package-level hash model is coarse on purpose.
		dirty[pkg.ImportPath] = struct{}{}
	}

	if rootBust {
		// Return every non-stdlib package.
		out := make([]string, 0, len(g.Packages))
		for ip, p := range g.Packages {
			if p.Standard || p.Goroot {
				continue
			}
			out = append(out, ip)
		}
		sort.Strings(out)
		return out
	}

	// BFS over reverse-dep edges.
	visited := map[string]struct{}{}
	queue := make([]string, 0, len(dirty))
	for ip := range dirty {
		visited[ip] = struct{}{}
		queue = append(queue, ip)
	}
	for len(queue) > 0 {
		ip := queue[0]
		queue = queue[1:]
		for importer := range g.reverseDeps[ip] {
			if _, seen := visited[importer]; seen {
				continue
			}
			visited[importer] = struct{}{}
			queue = append(queue, importer)
		}
	}

	out := make([]string, 0, len(visited))
	for ip := range visited {
		p, ok := g.Packages[ip]
		if !ok || p.Standard || p.Goroot {
			continue
		}
		out = append(out, ip)
	}
	sort.Strings(out)
	return out
}

// resolveAncestorPackage walks upward from dir looking for a directory that
// is a known package. Returns the import path, or "" if none found within
// the module root.
func (g *Graph) resolveAncestorPackage(dir string) string {
	stop := g.ModuleRoot
	if stop == "" {
		stop = g.WorkDir
	}
	cur := dir
	for {
		if ip, ok := g.byDir[cur]; ok {
			return ip
		}
		parent := filepath.Dir(cur)
		if parent == cur || !strings.HasPrefix(cur, stop) {
			return ""
		}
		cur = parent
	}
}
