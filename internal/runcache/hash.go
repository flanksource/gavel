// Package runcache stores per-package test run outcomes keyed on a content
// fingerprint, so `gavel test --cache` can skip packages whose inputs have
// not changed. See /Users/moshe/.claude/plans/shimmying-squishing-otter.md
// for the design context.
package runcache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/internal/cache"
	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/utils"
)

// ModTimeCutoff mirrors Go's errFileTooNew guard. If any file hashed into a
// package fingerprint was modified within this window before the run, we
// refuse to *write* the cache entry, defending against mtime-granularity
// races. Reads are unaffected.
const ModTimeCutoff = 2 * time.Second

// Fingerprint carries the package-level hash plus metadata about freshness.
// Freshness.TooRecent == true means the caller should not persist a cache
// entry derived from this fingerprint.
type Fingerprint struct {
	ImportPath string
	Hex        string
	TooRecent  bool
}

// Hasher computes per-package fingerprints with memoization. One Hasher per
// gavel invocation — it is not safe for concurrent use by multiple goroutines
// without external synchronization, but is cheap to construct.
type Hasher struct {
	graph *changegraph.Graph

	// effective memoizes effectiveHash by import path. Populated lazily.
	effective map[string]Fingerprint

	// local memoizes localHash by absolute package directory.
	local map[string]localResult

	// tagsSalt captures GOOS/GOARCH/tags/CGO_ENABLED/toolchain — constant
	// across every package in a single run.
	tagsSalt string
}

type localResult struct {
	hex       string
	tooRecent bool
}

// NewHasher constructs a Hasher bound to a loaded graph. tags is the
// resolved set of build tags relevant for the current run (may be nil).
func NewHasher(graph *changegraph.Graph, tags []string) *Hasher {
	return &Hasher{
		graph:     graph,
		effective: map[string]Fingerprint{},
		local:     map[string]localResult{},
		tagsSalt:  computeTagsSalt(tags),
	}
}

// Effective returns the effective fingerprint for the package at importPath,
// recursively mixing in direct-dep fingerprints. Stdlib / GOROOT packages
// terminate the recursion and contribute only the toolchain version.
func (h *Hasher) Effective(importPath string) (Fingerprint, error) {
	return h.effectiveWith(importPath, map[string]struct{}{})
}

func (h *Hasher) effectiveWith(importPath string, stack map[string]struct{}) (Fingerprint, error) {
	if fp, ok := h.effective[importPath]; ok {
		return fp, nil
	}
	// Cycle guard. Go does not permit import cycles, but defensive behavior
	// here avoids infinite recursion if `go list` ever reports one.
	if _, cycling := stack[importPath]; cycling {
		return Fingerprint{ImportPath: importPath, Hex: "cycle"}, nil
	}
	stack[importPath] = struct{}{}
	defer delete(stack, importPath)

	pkg, ok := h.graph.Packages[importPath]
	if !ok {
		return Fingerprint{}, fmt.Errorf("unknown package: %s", importPath)
	}

	hasher := sha256.New()
	fmt.Fprintf(hasher, "gavel-pkg-v1\n")
	fmt.Fprintf(hasher, "import %s\n", importPath)
	fmt.Fprintf(hasher, "salt %s\n", h.tagsSalt)

	// Stdlib / GOROOT packages: opaque, just keyed on toolchain version
	// (already in tagsSalt) and import path.
	if pkg.Standard || pkg.Goroot {
		sum := hex.EncodeToString(hasher.Sum(nil))
		fp := Fingerprint{ImportPath: importPath, Hex: sum}
		h.effective[importPath] = fp
		return fp, nil
	}

	local, err := h.localHash(pkg)
	if err != nil {
		return Fingerprint{}, fmt.Errorf("localHash %s: %w", importPath, err)
	}
	fmt.Fprintf(hasher, "local %s\n", local.hex)

	tooRecent := local.tooRecent

	// Mix in direct deps in sorted order for determinism. We deliberately
	// use Imports (direct) and let recursion handle the transitive closure.
	// TestImports and XTestImports are also included because editing a test
	// dependency should bust the package's test cache.
	deps := mergeSorted(pkg.Imports, pkg.TestImports, pkg.XTestImports)
	for _, dep := range deps {
		if dep == importPath {
			continue
		}
		depFP, err := h.effectiveWith(dep, stack)
		if err != nil {
			// Missing dep in the graph (rare — can happen for conditional
			// build tags). Fold in the import path only, don't fail.
			fmt.Fprintf(hasher, "missing-dep %s\n", dep)
			continue
		}
		fmt.Fprintf(hasher, "dep %s %s\n", dep, depFP.Hex)
		if depFP.TooRecent {
			tooRecent = true
		}
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	fp := Fingerprint{ImportPath: importPath, Hex: sum, TooRecent: tooRecent}
	h.effective[importPath] = fp
	return fp, nil
}

// localHash hashes the package directory tree: every non-gitignored file
// under pkg.Dir, by (relative path, content sha256). This intentionally
// includes testdata/, config files, YAML fixtures, etc. — anything in the
// directory contributes to the fingerprint.
func (h *Hasher) localHash(pkg *changegraph.Pkg) (localResult, error) {
	if cached, ok := h.local[pkg.Dir]; ok {
		return cached, nil
	}

	type entry struct {
		rel  string
		hash string
	}
	var entries []entry
	var tooRecent bool
	now := time.Now()

	err := utils.WalkGitIgnored(pkg.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Don't descend into nested packages — each package hashes its
			// own directory. If pkg.Dir/sub has its own .go files, `go list`
			// reports it as a separate package, and effectiveHash(pkg) will
			// reference sub via the imports graph, not via the directory walk.
			if path == pkg.Dir {
				return nil
			}
			if _, ok := h.graph.PackageByDir(path); ok {
				return fs.SkipDir
			}
			return nil
		}
		// Skip hidden files (.DS_Store, editor swap files).
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if now.Sub(info.ModTime()) < ModTimeCutoff {
			tooRecent = true
		}
		fileHash, err := cache.GetFileHash(path)
		if err != nil {
			return fmt.Errorf("hash %s: %w", path, err)
		}
		rel, err := filepath.Rel(pkg.Dir, path)
		if err != nil {
			return err
		}
		entries = append(entries, entry{rel: filepath.ToSlash(rel), hash: fileHash})
		return nil
	})
	if err != nil {
		return localResult{}, err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })

	hasher := sha256.New()
	fmt.Fprintf(hasher, "gavel-local-v1\n")
	for _, e := range entries {
		fmt.Fprintf(hasher, "%s %s\n", e.rel, e.hash)
	}

	res := localResult{hex: hex.EncodeToString(hasher.Sum(nil)), tooRecent: tooRecent}
	h.local[pkg.Dir] = res
	return res, nil
}

// computeTagsSalt builds a stable string capturing everything that isn't
// per-package but still affects the build: GOOS/GOARCH, CGO_ENABLED, tags,
// and Go runtime version. This ensures that switching GOOS or upgrading Go
// invalidates every cache entry without us having to clear the cache.
func computeTagsSalt(tags []string) string {
	sorted := append([]string{}, tags...)
	sort.Strings(sorted)
	cgo := os.Getenv("CGO_ENABLED")
	return strings.Join([]string{
		"v1",
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		"cgo=" + cgo,
		"tags=" + strings.Join(sorted, ","),
	}, "|")
}

// mergeSorted returns a sorted, deduplicated union of string slices.
func mergeSorted(slices ...[]string) []string {
	seen := map[string]struct{}{}
	for _, s := range slices {
		for _, v := range s {
			seen[v] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
