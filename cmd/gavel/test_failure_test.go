package main

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/flanksource/gavel/report"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// A run that dies before producing results must still serialize a document
// that report.ResultFile recognises as a crash, otherwise `--format json=FILE`
// writes nothing and downstream readers (gavel summary, the PR dashboard) have
// no way to explain the empty artifact.
func TestNewTestRunFailure_SerializesCrashEnvelope(t *testing.T) {
	const wantMsg = "pre-build: compiling Go test binaries failed (exit 1)"
	cause := errors.New(wantMsg)

	failure := newTestRunFailure(
		testrunner.RunOptions{WorkDir: t.TempDir()},
		nil, nil,
		time.Now().UTC().Add(-time.Minute),
		cause,
	)

	if failure.Error() != wantMsg {
		t.Errorf("Error() = %q, want %q", failure.Error(), wantMsg)
	}
	if !errors.Is(failure, cause) {
		t.Error("errors.Is could not reach the wrapped cause")
	}

	encoded, err := json.Marshal(failure)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded report.ResultFile
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("crash envelope is not readable as a result file: %v", err)
	}
	if !decoded.IsCrash() {
		t.Errorf("IsCrash() = false for %s", encoded)
	}
	if decoded.Error != wantMsg {
		t.Errorf("error = %q, want %q", decoded.Error, wantMsg)
	}
	if code := decoded.ExitCodeValue(); code == nil || *code != 1 {
		t.Errorf("exit code = %v, want 1", code)
	}
}

// Partial results survive into the envelope: whatever finished before the
// failure is still the most useful thing a reader can show.
func TestNewTestRunFailure_KeepsPartialResults(t *testing.T) {
	failure := newTestRunFailure(
		testrunner.RunOptions{WorkDir: t.TempDir()},
		[]parsers.Test{{Name: "TestServe", Failed: true}},
		nil,
		time.Now().UTC(),
		errors.New("runner exited 2"),
	)

	encoded, err := json.Marshal(failure)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded report.ResultFile
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Tests) != 1 || decoded.Tests[0].Name != "TestServe" {
		t.Fatalf("partial tests lost: %+v", decoded.Tests)
	}
	if decoded.IsCrash() {
		t.Error("a run with results should not be reported as a crash")
	}
}
