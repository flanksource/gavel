package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/internal/runcache"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// selectorContext carries the per-invocation state for change-graph and
// run-cache selection. A single instance is lazily constructed on first use
// and shared across frameworks in the same gavel invocation.
type selectorContext struct {
	workDir string
	graph   *changegraph.Graph
	hasher  *runcache.Hasher
	store   *runcache.Store

	// absDirByPkgPath memoizes the gavel-relative pkg path → absolute
	// directory mapping, since gavel represents packages as "./pkg/foo"
	// and go list returns import paths.
	absDirByPkgPath map[string]string
}

// newSelectorContext initializes the graph and hasher. The cache store is
// opened lazily on first cache hit/miss.
func newSelectorContext(workDir string) (*selectorContext, error) {
	graph, err := changegraph.Load(workDir)
	if err != nil {
		return nil, fmt.Errorf("load package graph: %w", err)
	}
	return &selectorContext{
		workDir:         workDir,
		graph:           graph,
		hasher:          runcache.NewHasher(graph, nil),
		absDirByPkgPath: map[string]string{},
	}, nil
}

// openStore lazily opens the on-disk run cache.
func (s *selectorContext) openStore() (*runcache.Store, error) {
	if s.store != nil {
		return s.store, nil
	}
	store, err := runcache.Open("")
	if err != nil {
		return nil, err
	}
	s.store = store
	return store, nil
}

// importPathOf maps a gavel-style relative package path ("./pkg/foo") to the
// Go import path reported by `go list`. Returns "" if the package is not in
// the graph.
func (s *selectorContext) importPathOf(pkgPath string) string {
	absDir, ok := s.absDirByPkgPath[pkgPath]
	if !ok {
		absDir = filepath.Join(s.workDir, filepath.FromSlash(pkgPath))
		absDir, _ = filepath.Abs(absDir)
		s.absDirByPkgPath[pkgPath] = absDir
	}
	if pkg, ok := s.graph.PackageByDir(absDir); ok {
		return pkg.ImportPath
	}
	return ""
}

// filterByChangeGraph narrows pkgs to those affected by the given DiffOptions.
// A nil or all-zero DiffOptions bypasses filtering and returns pkgs unchanged.
func (s *selectorContext) filterByChangeGraph(pkgs []string, opts changegraph.DiffOptions) ([]string, error) {
	fs, err := changegraph.ComputeFileSet(s.workDir, opts)
	if err != nil {
		return nil, fmt.Errorf("compute change set: %w", err)
	}
	if fs.Len() == 0 {
		logger.V(3).Infof("change graph: no changes detected, nothing will run")
		return nil, nil
	}

	affected := s.graph.AffectedPackages(fs)
	if len(affected) == 0 {
		return nil, nil
	}
	affectedSet := make(map[string]struct{}, len(affected))
	for _, ip := range affected {
		affectedSet[ip] = struct{}{}
	}

	kept := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		ip := s.importPathOf(pkg)
		if ip == "" {
			// Unknown package (e.g. starting path outside the module).
			// Keep conservatively.
			kept = append(kept, pkg)
			continue
		}
		if _, ok := affectedSet[ip]; ok {
			kept = append(kept, pkg)
		}
	}
	return kept, nil
}

// cacheHit represents a package that was served from the run cache and does
// not need to be re-executed. The entry is a snapshot of the original run.
type cacheHit struct {
	PkgPath     string
	Framework   parsers.Framework
	Fingerprint string
	Entry       runcache.Entry
}

// filterByRunCache removes packages whose fingerprint already has a cached
// successful entry for the given framework. Returns the packages that still
// need to run and the set of cache hits.
//
// A package whose fingerprint can't be computed (missing from the graph) is
// conservatively left in the "needs to run" list — gavel must never silently
// skip a test.
func (s *selectorContext) filterByRunCache(
	framework parsers.Framework,
	pkgs []string,
) (need []string, hits []cacheHit, err error) {
	store, err := s.openStore()
	if err != nil {
		return nil, nil, err
	}

	for _, pkg := range pkgs {
		ip := s.importPathOf(pkg)
		if ip == "" {
			need = append(need, pkg)
			continue
		}
		fp, err := s.hasher.Effective(ip)
		if err != nil {
			// Fingerprint failure = fall back to running.
			logger.V(3).Infof("run-cache: fingerprint %s: %v", ip, err)
			need = append(need, pkg)
			continue
		}
		key := framework.String() + ":" + fp.Hex
		if entry, ok := store.Lookup(key); ok {
			logger.V(3).Infof("run-cache HIT %s %s fp=%s", framework, pkg, shortHex(fp.Hex))
			hits = append(hits, cacheHit{
				PkgPath:     pkg,
				Framework:   framework,
				Fingerprint: key,
				Entry:       entry,
			})
			continue
		}
		logger.V(3).Infof("run-cache MISS %s %s fp=%s tooRecent=%v", framework, pkg, shortHex(fp.Hex), fp.TooRecent)
		need = append(need, pkg)
	}
	return need, hits, nil
}

// recordSuccess persists a passing test result to the run cache. Failures
// and too-recent edits are silently skipped inside Record.
func (s *selectorContext) recordSuccess(
	framework parsers.Framework,
	pkgPath string,
	suite parsers.TestSuiteResults,
	duration time.Duration,
) {
	if s == nil || s.store == nil {
		return
	}
	ip := s.importPathOf(pkgPath)
	if ip == "" {
		return
	}
	fp, err := s.hasher.Effective(ip)
	if err != nil {
		logger.V(3).Infof("run-cache: cannot fingerprint %s for record: %v", ip, err)
		return
	}
	sum := suite.All().Sum()
	if sum.Failed > 0 {
		return
	}
	key := framework.String() + ":" + fp.Hex
	entry := runcache.Entry{
		ImportPath:    ip,
		Framework:     framework.String(),
		ExitCode:      0,
		PassCount:     sum.Passed,
		FailCount:     sum.Failed,
		SkipCount:     sum.Skipped,
		DurationNanos: duration.Nanoseconds(),
		RecordedAt:    time.Now().UnixNano(),
	}
	if err := s.store.Record(key, fp.TooRecent, entry); err != nil {
		logger.V(3).Infof("run-cache: record %s: %v", key, err)
	}
}

// cachedSuiteResults builds a synthetic TestSuiteResults for a cache hit,
// so the rest of the runner pipeline treats it the same as a real run.
// Each cached package gets a single synthetic passing "Test" with Cached=true
// so the UI can show a "(cached)" marker.
func cachedSuiteResults(hit cacheHit) parsers.TestSuiteResults {
	return parsers.TestSuiteResults{{
		Framework: hit.Framework,
		ExitCode:  0,
		Tests: parsers.Tests{{
			Name:        hit.PkgPath,
			PackagePath: hit.PkgPath,
			Framework:   hit.Framework,
			Passed:      true,
			Cached:      true,
			Duration:    hit.Entry.Duration(),
		}},
	}}
}

// diffOptionsFromRunOptions translates user-facing flags into the git
// DiffOptions fed to ComputeFileSet. Changed or Since always imply
// staged+unstaged+untracked so the working tree is reflected.
func diffOptionsFromRunOptions(opts RunOptions) changegraph.DiffOptions {
	diff := changegraph.DiffOptions{
		IncludeStaged:    true,
		IncludeUnstaged:  true,
		IncludeUntracked: true,
	}
	if opts.Since != "" {
		diff.Since = opts.Since
		return diff
	}
	if opts.Changed {
		diff.Since = changedBaseRef()
	}
	return diff
}

// changedBaseRef is the base ref used by `--changed`. Defaults to
// origin/main but can be overridden via GAVEL_CHANGED_BASE.
func changedBaseRef() string {
	if v := os.Getenv("GAVEL_CHANGED_BASE"); v != "" {
		return v
	}
	return "origin/main"
}

// shortHex returns the first 12 characters of hex, safe for short hashes.
func shortHex(hex string) string {
	if len(hex) < 12 {
		return hex
	}
	return hex[:12]
}
