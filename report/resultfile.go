package report

import (
	"encoding/json"
	"strings"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// ResultFileMetadata is the subset of testui.SnapshotMetadata that readers of a
// result file care about. Snapshots written by `gavel test --format json=FILE`
// carry the exit code here rather than at the top level.
type ResultFileMetadata struct {
	ExitCode *int `json:"exit_code,omitempty"`
}

// ResultFile is the wire format of a gavel result JSON file. It is the single
// decoding contract shared by every reader of gavel output — `gavel summary`,
// the PR dashboard, and the todo check loop — so the shape cannot drift between
// producer and consumer.
//
// Three producers write it:
//   - `gavel test`:        a bare JSON array of parsers.Test
//   - `gavel test --lint`: a testui.Snapshot object with tests/lint/bench keys
//   - a crashed run:       the same object with Error set and no results, either
//     from gavel itself or from the composite action's fallback stub
type ResultFile struct {
	Tests []parsers.Test          `json:"tests"`
	Lint  []*linters.LinterResult `json:"lint,omitempty"`
	Bench *bench.BenchComparison  `json:"bench,omitempty"`

	// Error / ExitCode / LogTail describe a run that died before producing
	// results. ExitCode is top-level in the composite action's stub and under
	// Metadata in a gavel-written snapshot; use ExitCodeValue to read either.
	Error    string              `json:"error,omitempty"`
	ExitCode *int                `json:"exit_code,omitempty"`
	LogTail  string              `json:"log_tail,omitempty"`
	Metadata *ResultFileMetadata `json:"metadata,omitempty"`
}

// UnmarshalJSON accepts both shapes gavel emits: a plain JSON array of
// parsers.Test, or an object with tests/lint/bench keys.
func (r *ResultFile) UnmarshalJSON(data []byte) error {
	if strings.HasPrefix(strings.TrimSpace(string(data)), "[") {
		var tests []parsers.Test
		if err := json.Unmarshal(data, &tests); err != nil {
			return err
		}
		r.Tests = tests
		return nil
	}
	type alias ResultFile
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	*r = ResultFile(a)
	return nil
}

// ExitCodeValue returns the run's exit code, preferring the top-level field and
// falling back to metadata.exit_code. nil when neither producer recorded one.
func (r ResultFile) ExitCodeValue() *int {
	if r.ExitCode != nil {
		return r.ExitCode
	}
	if r.Metadata != nil {
		return r.Metadata.ExitCode
	}
	return nil
}

// IsCrash reports whether the file describes a run that died before producing
// any results. A run that failed *and* produced results is not a crash — its
// tests and violations are the useful output and Error is supplementary.
func (r ResultFile) IsCrash() bool {
	return len(r.Tests) == 0 && len(r.Lint) == 0 && r.Error != ""
}
