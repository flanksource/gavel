package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// TestKey uniquely identifies a test across runs.
type TestKey struct {
	Framework   string
	PackagePath string
	FullName    string
}

// ViolationKey uniquely identifies a lint violation across runs.
// Line-insensitive since lines shift with edits.
type ViolationKey struct {
	Linter string
	File   string
	Rule   string
}

func LoadSnapshot(path string) (*testui.Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading baseline %s: %w", path, err)
	}
	var snapshot testui.Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("parsing baseline %s: %w", path, err)
	}
	return &snapshot, nil
}

func ExtractFailedTestKeys(tests []parsers.Test) map[TestKey]bool {
	keys := make(map[TestKey]bool)
	walkTests(tests, func(t parsers.Test) {
		if !t.Failed {
			return
		}
		keys[TestKey{
			Framework:   string(t.Framework),
			PackagePath: t.PackagePath,
			FullName:    t.FullName(),
		}] = true
	})
	return keys
}

func ExtractViolationKeys(results []*linters.LinterResult) map[ViolationKey]bool {
	keys := make(map[ViolationKey]bool)
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		for _, v := range r.Violations {
			file := v.File
			if r.WorkDir != "" && filepath.IsAbs(file) {
				if rel, err := filepath.Rel(r.WorkDir, file); err == nil {
					file = rel
				}
			}
			rule := ""
			if v.Rule != nil {
				rule = v.Rule.Method
			}
			keys[ViolationKey{
				Linter: r.Linter,
				File:   file,
				Rule:   rule,
			}] = true
		}
	}
	return keys
}

// FailedTestPackages returns framework → packagePath → failed test names.
func ExtractFailedTestPackages(tests []parsers.Test) map[parsers.Framework]map[string][]string {
	result := make(map[parsers.Framework]map[string][]string)
	walkTests(tests, func(t parsers.Test) {
		if !t.Failed || t.Framework == "" {
			return
		}
		if result[t.Framework] == nil {
			result[t.Framework] = make(map[string][]string)
		}
		result[t.Framework][t.PackagePath] = append(result[t.Framework][t.PackagePath], t.Name)
	})
	return result
}

// ExtractFailedLintTargets returns the linter names and files that had violations.
func ExtractFailedLintTargets(results []*linters.LinterResult) (linterNames []string, files []string) {
	seenLinters := make(map[string]bool)
	seenFiles := make(map[string]bool)
	for _, r := range results {
		if r == nil || r.Skipped || len(r.Violations) == 0 {
			continue
		}
		if !seenLinters[r.Linter] {
			seenLinters[r.Linter] = true
			linterNames = append(linterNames, r.Linter)
		}
		for _, v := range r.Violations {
			file := v.File
			if r.WorkDir != "" && filepath.IsAbs(file) {
				if rel, err := filepath.Rel(r.WorkDir, file); err == nil {
					file = rel
				}
			}
			if !seenFiles[file] {
				seenFiles[file] = true
				files = append(files, file)
			}
		}
	}
	return linterNames, files
}

// walkTests recursively visits leaf test nodes (nodes with a status set).
func walkTests(tests []parsers.Test, fn func(t parsers.Test)) {
	for _, t := range tests {
		if t.Failed || t.Passed || t.Skipped {
			fn(t)
		}
		if len(t.Children) > 0 {
			walkTests(t.Children, fn)
		}
	}
}
