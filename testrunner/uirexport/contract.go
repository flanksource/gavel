// Package uirexport writes a gavel test run as the JSON `uir tests import` reads: version 1 of
// uir's producer-neutral test run contract (github.com/flanksource/uir, testruns.Import). The types
// mirror that contract field for field; gavel does not import uir.
package uirexport

import "time"

// ContractVersion is the version of uir's test import contract these types follow.
const ContractVersion = 1

// Import is the document `uir tests import` reads.
type Import struct {
	Version int   `json:"version"`
	Runs    []Run `json:"runs"`
}

// Run is one gavel test run. Environment holds what makes two runs comparable: GOOS, GOARCH, go,
// race, tags and count. GitCommit, Dirty and, for a dirty worktree, GitTree name the code the tests
// ran on (Provenance).
type Run struct {
	Producer     string            `json:"producer"`
	GitCommit    string            `json:"git_commit"`
	Dirty        bool              `json:"dirty"`
	GitTree      string            `json:"git_tree,omitempty"`
	Environment  map[string]string `json:"environment"`
	StartedAt    time.Time         `json:"started_at"`
	FinishedAt   time.Time         `json:"finished_at"`
	TaskRunID    string            `json:"task_run_id,omitempty"`
	InputsDigest string            `json:"inputs_digest,omitempty"`
	Tests        []Test            `json:"tests"`
	Coverage     []Block           `json:"coverage,omitempty"`
}

// Test is one test the run reported. Package is a Go import path; File is module-relative with
// forward slashes.
type Test struct {
	Framework string    `json:"framework"`
	Package   string    `json:"package"`
	Suite     []string  `json:"suite,omitempty"`
	Name      string    `json:"name"`
	File      string    `json:"file,omitempty"`
	Line      int       `json:"line,omitempty"`
	Attempts  []Attempt `json:"attempts"`
	Coverage  []Block   `json:"coverage,omitempty"`
}

// Outcome is an attempt's result.
type Outcome string

const (
	Pass    Outcome = "pass"
	Fail    Outcome = "fail"
	Skip    Outcome = "skip"
	Timeout Outcome = "timeout"
	Aborted Outcome = "aborted"
)

// Attempt is one execution of a test within the run.
type Attempt struct {
	Outcome    Outcome `json:"outcome"`
	DurationMS int64   `json:"duration_ms"`
	Failure    string  `json:"failure,omitempty"`
}

// Block is a coverage block of a module-relative file: one-based lines and byte columns, the
// statements it holds and how many times it ran.
type Block struct {
	Path       string `json:"path"`
	StartLine  int    `json:"start_line"`
	StartCol   int    `json:"start_col"`
	EndLine    int    `json:"end_line"`
	EndCol     int    `json:"end_col"`
	Statements int    `json:"statements"`
	Count      int64  `json:"count"`
}
