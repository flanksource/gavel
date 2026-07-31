package main

import (
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// testRunFailure is the error a failed run returns for serialized output
// formats. Clicky's command runner drops the result value whenever the command
// returns an error, so a run that dies before finishing — a pre-build failure,
// a timeout, a crashed runner — would otherwise write no `--format json=FILE`
// document at all, leaving CI to guess. Embedding testui.Snapshot gives this
// error the Pretty/GetChildren interfaces clicky's renderableError looks for,
// so the same format pipeline that renders a successful run also writes the
// crash envelope, while the error is still returned and the process exits
// non-zero.
type testRunFailure struct {
	testui.Snapshot
	err error
}

func (f testRunFailure) Error() string { return f.err.Error() }

func (f testRunFailure) Unwrap() error { return f.err }

// newTestRunFailure builds the crash envelope from whatever the runner managed
// to produce. tests and lint may be empty — a pre-build failure yields neither,
// and the Error field is then the only useful output.
func newTestRunFailure(
	opts testrunner.RunOptions,
	tests []parsers.Test,
	lint []*linters.LinterResult,
	started time.Time,
	err error,
) testRunFailure {
	snapshot := buildTestSnapshot(opts, tests, lint, started, time.Now().UTC(), nil)
	snapshot.Error = err.Error()
	failed := 1
	snapshot.Metadata.ExitCode = &failed
	return testRunFailure{Snapshot: snapshot, err: err}
}

// testRunFailureValue picks the error a failed run should return. Pretty
// (terminal) output has already printed the breakdown and summary eagerly, so
// returning a renderable error there would render the snapshot tree a second
// time — mirroring testRunReturnValue, the crash envelope is reserved for
// serialized formats.
func testRunFailureValue(
	opts testrunner.RunOptions,
	tests []parsers.Test,
	lint []*linters.LinterResult,
	started time.Time,
	err error,
) error {
	if isPrettyFormat() {
		return err
	}
	return newTestRunFailure(opts, tests, lint, started, err)
}
