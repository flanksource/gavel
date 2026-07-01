package outline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/linters/jscpd"
)

type lineInterval struct{ start, end int }

// testFilePattern matches the test files the outline covers. jscpd 4.x only
// honors a single positional path, so duplication scans workDir with this
// pattern instead of an explicit file list.
const testFilePattern = "**/{*_test.go,*.test.*,*.spec.*}"

// applyDuplication runs jscpd over the workdir's test files and sets
// DuplicationPct on each leaf to the share of its body lines covered by a
// clone. The default jscpd linter excludes test files, so the outline invokes
// jscpd directly.
func applyDuplication(ctx context.Context, report *Report, workDir string) error {
	if len(report.Entries) == 0 {
		return nil
	}

	jscpdReport, err := runJscpd(ctx, workDir)
	if err != nil {
		return err
	}
	annotateDuplication(report, intervalsByFile(jscpdReport, workDir))
	return nil
}

func runJscpd(ctx context.Context, workDir string) (*jscpd.JscpdReport, error) {
	tempDir, err := os.MkdirTemp("", "gavel-outline-jscpd-*")
	if err != nil {
		return nil, fmt.Errorf("create jscpd temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	args := []string{
		"--reporters", "json",
		"--output", tempDir,
		"--pattern", testFilePattern,
		"--ignore", "**/node_modules/**,**/dist/**,**/.git/**",
		"--threshold", "100", // never fail on clone volume; the report is the product
		".",
	}
	cmd := exec.CommandContext(ctx, "jscpd", args...)
	cmd.Dir = workDir
	output, runErr := cmd.CombinedOutput()

	reportPath := filepath.Join(tempDir, "jscpd-report.json")
	data, readErr := os.ReadFile(reportPath)
	if runErr != nil {
		if errors.Is(runErr, exec.ErrNotFound) {
			return nil, fmt.Errorf("jscpd is not installed (needed for duplication metrics); install it or pass --duplication=false")
		}
		// jscpd exits non-zero when clones exceed its threshold; the report is
		// still authoritative when it was written.
		if readErr != nil {
			return nil, fmt.Errorf("jscpd failed: %w\nOutput:\n%s", runErr, string(output))
		}
	}
	if readErr != nil {
		return nil, fmt.Errorf("jscpd produced no report at %s: %w\nOutput:\n%s", reportPath, readErr, string(output))
	}

	var jscpdReport jscpd.JscpdReport
	if err := json.Unmarshal(data, &jscpdReport); err != nil {
		return nil, fmt.Errorf("parse jscpd report: %w", err)
	}
	return &jscpdReport, nil
}

// intervalsByFile collects cloned line ranges per workdir-relative file from
// both sides of every duplicate pair.
func intervalsByFile(report *jscpd.JscpdReport, workDir string) map[string][]lineInterval {
	intervals := map[string][]lineInterval{}
	add := func(ref jscpd.JscpdFileRef) {
		file := relativeTo(ref.Name, workDir)
		intervals[file] = append(intervals[file], lineInterval{start: ref.StartLoc.Line, end: ref.EndLoc.Line})
	}
	for _, dup := range report.Duplicates {
		add(dup.FirstFile)
		add(dup.SecondFile)
	}
	return intervals
}

func relativeTo(path, workDir string) string {
	cleaned := filepath.Clean(path)
	if filepath.IsAbs(cleaned) {
		if rel, err := filepath.Rel(workDir, cleaned); err == nil && !strings.HasPrefix(rel, "..") {
			cleaned = rel
		}
	}
	return filepath.ToSlash(cleaned)
}

func annotateDuplication(report *Report, intervals map[string][]lineInterval) {
	for _, leaf := range report.Leaves() {
		if leaf.SizeLines == 0 || leaf.Line == 0 {
			continue
		}
		covered := coveredLines(intervals[leaf.File], leaf.Line, leaf.EndLine)
		if covered > 0 {
			leaf.DuplicationPct = float64(covered) / float64(leaf.SizeLines) * 100
		}
	}
}

// coveredLines returns how many lines in [lo, hi] fall inside at least one
// interval, merging overlaps so double-counted clones don't inflate coverage.
func coveredLines(intervals []lineInterval, lo, hi int) int {
	var clipped []lineInterval
	for _, iv := range intervals {
		start, end := max(iv.start, lo), min(iv.end, hi)
		if start <= end {
			clipped = append(clipped, lineInterval{start, end})
		}
	}
	if len(clipped) == 0 {
		return 0
	}
	sort.Slice(clipped, func(i, j int) bool { return clipped[i].start < clipped[j].start })

	covered := 0
	current := clipped[0]
	for _, iv := range clipped[1:] {
		if iv.start <= current.end+1 {
			current.end = max(current.end, iv.end)
			continue
		}
		covered += current.end - current.start + 1
		current = iv
	}
	return covered + current.end - current.start + 1
}
