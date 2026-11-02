package snapshots

import (
	"fmt"
	"slices"

	"github.com/flanksource/gavel/linters"
)

// LintDiff classifies a lint run's findings against the previous run by
// fingerprint. New, Persisting, Ignored and Dismissed are this run's; Fixed are
// the previous run's open findings that this run no longer reports at all.
type LintDiff struct {
	New        []string `json:"new,omitempty"`
	Fixed      []string `json:"fixed,omitempty"`
	Persisting []string `json:"persisting,omitempty"`
	Ignored    []string `json:"ignored,omitempty"`
	Dismissed  []string `json:"dismissed,omitempty"`
}

// DiffLint compares two runs' lint results. Every violation must carry a
// fingerprint, which lint.FilterConfiguredViolations assigns.
func DiffLint(prev, cur []*linters.LinterResult) (LintDiff, error) {
	before, err := openFingerprints(prev)
	if err != nil {
		return LintDiff{}, err
	}
	now, err := openFingerprints(cur)
	if err != nil {
		return LintDiff{}, err
	}
	var diff LintDiff
	seen := map[string]bool{}
	for _, result := range cur {
		if result == nil {
			continue
		}
		for _, ignored := range result.Ignored {
			diff.Ignored = append(diff.Ignored, ignored.Violation.Fingerprint)
			seen[ignored.Violation.Fingerprint] = true
		}
		for _, dismissed := range result.Dismissed {
			diff.Dismissed = append(diff.Dismissed, dismissed.Fingerprint)
			seen[dismissed.Fingerprint] = true
		}
	}
	for _, fp := range now {
		if slices.Contains(before, fp) {
			diff.Persisting = append(diff.Persisting, fp)
		} else {
			diff.New = append(diff.New, fp)
		}
		seen[fp] = true
	}
	for _, fp := range before {
		if !seen[fp] {
			diff.Fixed = append(diff.Fixed, fp)
		}
	}
	return diff, nil
}

// PreviousLintRun is the newest run before runID that ran linters, or nil when
// runID is the first.
func PreviousLintRun(workDir, runID string) (*RunInfo, error) {
	runs, err := ListRuns(workDir, ListRunsOptions{Before: runID})
	if err != nil {
		return nil, err
	}
	for _, run := range runs {
		if run.Kind == RunKindLint || run.Kind == RunKindTestLint {
			return &run, nil
		}
	}
	return nil, nil
}

func openFingerprints(results []*linters.LinterResult) ([]string, error) {
	var fingerprints []string
	for _, result := range results {
		if result == nil || result.Skipped {
			continue
		}
		for _, v := range result.Violations {
			if v.Fingerprint == "" {
				return nil, fmt.Errorf("%s violation at %s:%d has no fingerprint; lint results must pass through lint.FilterConfiguredViolations", result.Linter, v.File, v.Line)
			}
			fingerprints = append(fingerprints, v.Fingerprint)
		}
	}
	return fingerprints, nil
}
