package outline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
)

// vitestListItem mirrors one element of `vitest list --json` output.
// vitest 2.x only emits location when includeTaskLocation is enabled in the
// project config; entries without it render without a line number.
type vitestListItem struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	ProjectName string `json:"projectName,omitempty"`
	Location    *struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"location,omitempty"`
}

// collectVitestTests uses vitest's native outline (`vitest list --json`) per
// discovered vitest package, skipping packages outside the path filters.
// vitest collects tests by importing test modules without executing the tests
// themselves. Size and complexity are not computed for JS tests.
func collectVitestTests(ctx context.Context, workDir string, filters []string) ([]*Entry, error) {
	pkgs, err := runners.NewVitest(workDir).DiscoverPackages(workDir, true)
	if err != nil {
		return nil, fmt.Errorf("discover vitest packages: %w", err)
	}

	var entries []*Entry
	for _, pkg := range pkgs {
		if !packageMatchesFilters(pkg, filters) {
			continue
		}
		items, err := vitestList(ctx, workDir, pkg)
		if err != nil {
			// A package that can't be collected (e.g. deps not installed) is
			// surfaced as an error row in the outline rather than aborting the
			// whole run, so go/ginkgo and other vitest packages still render.
			entries = append(entries, vitestErrorEntry(pkg, err))
			continue
		}
		for _, item := range items {
			entries = append(entries, vitestEntry(item, workDir))
		}
	}
	return entries, nil
}

// packageMatchesFilters reports whether a vitest package dir (gavel-style
// "./apps/web") could contain tests selected by the filters: either the
// package lies under a filter, or a filter points inside the package.
func packageMatchesFilters(pkg string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	rel := strings.Trim(strings.TrimPrefix(filepath.ToSlash(pkg), "./"), "/")
	if rel == "" || rel == "." {
		return true
	}
	for _, filter := range filters {
		if rel == filter || strings.HasPrefix(rel, filter+"/") || strings.HasPrefix(filter, rel+"/") {
			return true
		}
	}
	return false
}

// vitestErrorEntry anchors a collection failure on the package's package.json
// so it renders under the package's directory in the tree, summarizing why the
// package could not be listed (e.g. deps not installed). The noisy tool output
// is omitted from the row.
func vitestErrorEntry(pkg string, err error) *Entry {
	rel := strings.Trim(strings.TrimPrefix(filepath.ToSlash(pkg), "./"), "/")
	if rel == "" {
		rel = "."
	}
	return &Entry{
		Framework: parsers.Vitest,
		File:      path.Join(rel, "package.json"),
		Name:      "<vitest collection failed>",
		Error:     summarizeVitestErr(err),
	}
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// summarizeVitestErr pulls the most informative single line out of a (possibly
// ANSI-colored, multi-line) vitest failure: the module-resolution cause if
// present, else our own first-line prefix.
func summarizeVitestErr(err error) string {
	clean := ansiPattern.ReplaceAllString(err.Error(), "")
	for _, line := range strings.Split(clean, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Could not resolve") ||
			strings.Contains(line, "Cannot find") ||
			strings.Contains(line, "ERR_MODULE_NOT_FOUND") {
			return line
		}
	}
	first, _, _ := strings.Cut(clean, "\n")
	return strings.TrimSpace(first)
}

func vitestList(ctx context.Context, workDir, pkg string) ([]vitestListItem, error) {
	cwd := filepath.Join(workDir, pkg)
	outFile, err := os.CreateTemp("", "gavel-outline-vitest-*.json")
	if err != nil {
		return nil, fmt.Errorf("create vitest list output file: %w", err)
	}
	outPath := outFile.Name()
	_ = outFile.Close()
	defer os.Remove(outPath)

	command, pre := runners.DetectPackageManager(cwd)
	args := append(append([]string{}, pre...), "vitest", "list", "--json="+outPath)
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = cwd
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("vitest list in %s failed: %w\nOutput:\n%s", cwd, err, string(output))
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read vitest list output for %s: %w", cwd, err)
	}
	return parseVitestList(data, cwd)
}

func parseVitestList(data []byte, cwd string) ([]vitestListItem, error) {
	var items []vitestListItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parse vitest list output for %s: %w", cwd, err)
	}
	return items, nil
}

// vitestEntry converts a list item to a leaf entry. The name is the suite
// chain joined with " > ", matching the ancestorTitles join used by the run
// parser; titles containing " > " themselves are ambiguous and split anyway.
func vitestEntry(item vitestListItem, workDir string) *Entry {
	segments := strings.Split(item.Name, " > ")
	entry := &Entry{
		Framework: parsers.Vitest,
		File:      relativeTo(item.File, workDir),
		Name:      segments[len(segments)-1],
		Suite:     segments[:len(segments)-1],
	}
	if len(entry.Suite) == 0 {
		entry.Suite = nil
	}
	if item.Location != nil {
		entry.Line = item.Location.Line
	}
	return entry
}
