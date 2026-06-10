package outline

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
)

// Build statically outlines tests under opts.WorkDir and enriches them with
// duplication, run history, and descriptions. Parse and tool failures are
// surfaced, not skipped.
func Build(opts Options) (*Report, error) {
	if opts.WorkDir == "" {
		return nil, fmt.Errorf("outline.Build: WorkDir is required")
	}
	workDir, err := filepath.Abs(opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workdir: %w", err)
	}
	opts.WorkDir = workDir
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	frameworks := opts.Frameworks
	if len(frameworks) == 0 {
		frameworks = []parsers.Framework{parsers.GoTest, parsers.Ginkgo, parsers.Vitest}
	}
	enabled := map[parsers.Framework]bool{}
	for _, fw := range frameworks {
		switch fw {
		case parsers.GoTest, parsers.Ginkgo, parsers.Vitest:
			enabled[fw] = true
		default:
			return nil, fmt.Errorf("framework %q is not supported by outline yet", fw)
		}
	}

	report := &Report{WorkDir: workDir}
	filters := normalizeFilters(opts.Paths)

	if enabled[parsers.GoTest] || enabled[parsers.Ginkgo] {
		entries, err := collectGoEntries(workDir, filters, enabled)
		if err != nil {
			return nil, err
		}
		report.Entries = entries
	}

	if enabled[parsers.Vitest] {
		vitestEntries, err := collectVitestTests(opts.Context, workDir, filters)
		if err != nil {
			return nil, err
		}
		for _, entry := range vitestEntries {
			if matchesFilters(entry.File, filters) {
				report.Entries = append(report.Entries, entry)
			}
		}
	}

	sortEntries(report.Entries)

	if opts.Duplication {
		if err := applyDuplication(opts.Context, report, workDir); err != nil {
			return nil, err
		}
	}
	if opts.History {
		runCount, err := joinHistory(report, opts)
		if err != nil {
			return nil, err
		}
		report.RunCount = runCount
	}
	applyDescriptions(report)
	if opts.AISummary {
		if err := applyAISummaries(opts.Context, report, workDir); err != nil {
			return nil, err
		}
	}
	return report, nil
}

func collectGoEntries(workDir string, filters []string, enabled map[parsers.Framework]bool) ([]*Entry, error) {
	var entries []*Entry
	err := utils.WalkGitIgnored(workDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relPath := relativeTo(path, workDir)
		if !matchesFilters(relPath, filters) {
			return nil
		}

		fset := token.NewFileSet()
		file, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", relPath, parseErr)
		}

		if importsGinkgo(file) {
			if !enabled[parsers.Ginkgo] {
				return nil
			}
			ginkgoEntries, ginkgoErr := extractGinkgoTests(fset, file, relPath)
			if ginkgoErr != nil {
				return ginkgoErr
			}
			entries = append(entries, ginkgoEntries...)
			// The bootstrap Test* func in a ginkgo file is runner plumbing,
			// not a test; plain Test* funcs coexisting with specs still count.
			if enabled[parsers.GoTest] {
				entries = append(entries, extractGoTests(fset, file, relPath)...)
			}
			return nil
		}
		if enabled[parsers.GoTest] {
			entries = append(entries, extractGoTests(fset, file, relPath)...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func importsGinkgo(file *ast.File) bool {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path == "github.com/onsi/ginkgo" || path == "github.com/onsi/ginkgo/v2" {
			return true
		}
	}
	return false
}

func normalizeFilters(paths []string) []string {
	var filters []string
	for _, path := range paths {
		cleaned := filepath.ToSlash(strings.TrimSpace(path))
		cleaned = strings.TrimPrefix(cleaned, "./")
		cleaned = strings.Trim(cleaned, "/")
		if cleaned != "" && cleaned != "." {
			filters = append(filters, cleaned)
		}
	}
	return filters
}

func matchesFilters(relPath string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if relPath == filter || strings.HasPrefix(relPath, filter+"/") {
			return true
		}
	}
	return false
}

func sortEntries(entries []*Entry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].File != entries[j].File {
			return entries[i].File < entries[j].File
		}
		return entries[i].Line < entries[j].Line
	})
}
