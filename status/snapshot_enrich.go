package status

import (
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Injection points so tests can supply a synthetic snapshot instead of
// touching disk. Mirror the fetchFileMapFunc pattern in status.go.
var (
	loadPointerFunc  = snapshots.LoadPointer
	loadSnapshotFunc = snapshots.LoadByPointer
	snapshotIDFunc   = snapshots.SnapshotID
)

func enrichWithSnapshot(workDir string, result *Result) error {
	currentSHA, currentUncommitted, err := snapshotIDFunc(workDir)
	if err != nil {
		return err
	}
	result.CurrentSHA = currentSHA

	pointer, err := loadPointerFunc(workDir, snapshots.PointerLast)
	if err != nil || pointer == nil {
		return err
	}

	snap, err := loadSnapshotFunc(workDir, pointer)
	if err != nil || snap == nil {
		return err
	}

	result.ResultsSHA = pointer.SHA
	result.ResultsStale = pointer.SHA != currentSHA || pointer.Uncommitted != currentUncommitted

	testsByFile := map[string]TestStatus{}
	flattenTests(snap.Tests, workDir, testsByFile)

	lintByFile := map[string]LintStatus{}
	collectLintByFile(snap.Lint, workDir, lintByFile)

	for i := range result.Files {
		f := &result.Files[i]
		tagged := false
		if t, ok := testsByFile[f.Path]; ok {
			f.TestStatus = t
			tagged = true
		}
		if l, ok := lintByFile[f.Path]; ok {
			f.LintStatus = l
			tagged = true
		}
		if tagged && result.ResultsStale {
			f.ResultsStale = true
		}
	}
	return nil
}

func flattenTests(tests []parsers.Test, workDir string, out map[string]TestStatus) {
	for _, t := range tests {
		if path := normalisePath(t.File, workDir); path != "" {
			s := out[path]
			switch {
			case t.Failed:
				s.Failed++
			case t.Skipped:
				s.Skipped++
			case t.Passed:
				s.Passed++
			}
			out[path] = s
		}
		if len(t.Children) > 0 {
			flattenTests(t.Children, workDir, out)
		}
	}
}

func collectLintByFile(lint []*linters.LinterResult, workDir string, out map[string]LintStatus) {
	for _, lr := range lint {
		if lr == nil {
			continue
		}
		for _, v := range lr.Violations {
			path := normalisePath(v.File, workDir)
			if path == "" {
				continue
			}
			s := out[path]
			switch strings.ToLower(string(v.Severity)) {
			case "error", "critical", "high":
				s.Errors++
			case "warning", "medium":
				s.Warnings++
			default:
				s.Infos++
			}
			out[path] = s
		}
	}
}

func normalisePath(path, workDir string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(workDir, path)
		if err == nil {
			path = rel
		}
	}
	return filepath.ToSlash(path)
}
