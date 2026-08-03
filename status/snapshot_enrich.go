package status

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	rpchttp "github.com/flanksource/clicky/rpc/http"
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

func enrichWithSnapshot(ctx context.Context, workDir string, result *Result) error {
	currentSHA, currentUncommitted, err := snapshotIDFunc(ctx, workDir)
	if err != nil {
		return err
	}
	result.CurrentSHA = currentSHA

	stopFile := rpchttp.Track(ctx, "file")
	pointer, err := loadPointerFunc(workDir, snapshots.PointerLast)
	stopFile()
	if err != nil || pointer == nil {
		return err
	}

	stopFile = rpchttp.Track(ctx, "file")
	snap, err := loadSnapshotFunc(workDir, pointer)
	stopFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Stale pointer: last.json references a snapshot whose dirty-state
			// hash no longer matches any file on disk. Enrichment is optional —
			// proceed without prior test/lint tags rather than aborting commit.
			return nil
		}
		return err
	}
	if snap == nil {
		return nil
	}

	result.ResultsSHA = pointer.SHA
	result.ResultsStale = pointer.SHA != currentSHA || pointer.Uncommitted != currentUncommitted

	testsByFile := map[string]TestStatus{}
	flattenTests(snap.Tests, workDir, testsByFile)

	lintByFile := map[string]LintStatus{}
	collectLintByFile(snap.Lint, workDir, lintByFile)

	problemsByFile := map[string][]Problem{}
	collectTestProblems(snap.Tests, workDir, problemsByFile)
	collectLintProblems(snap.Lint, workDir, problemsByFile)

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
		if p, ok := problemsByFile[f.Path]; ok {
			f.Problems = sortProblems(p)
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
