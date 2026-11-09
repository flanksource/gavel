package uirexport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Options is what an export needs beyond the run's tests.
type Options struct {
	// WorkDir is the run's working directory; the export covers the Go module that holds it.
	WorkDir               string
	StartedAt, FinishedAt time.Time
	// InputsDigest is the run's InputsDigest, taken before its tests ran so files they write do not
	// change it.
	InputsDigest     string
	Environment      map[string]string
	SelectionMode    string
	InventoryDigest  string
	Complete         bool
	IdentityVerified bool
	PlanID           string
	Inventory        []InventoryTest
	ContractVersion  int
	// Provenance is the code the tests ran on, also taken before they ran (CaptureProvenance).
	Provenance Provenance
	// CoverProfiles are the run's go test -coverprofile files, the run's aggregate coverage.
	CoverProfiles []string
	// CoverTest, when set, writes to profile the coverage of test run alone (testrunner.CoverTest).
	// Build calls it for every go test and ginkgo test that did not skip, with profiles in CoverDir.
	CoverTest func(test parsers.Test, profile string) error
	CoverDir  string
}

// Build turns a run's flat test results (testrunner.RunOptions.ResultsOut) into a one-run import.
// Only go test and ginkgo tests are exported, and not packages gavel's run cache skipped.
func Build(tests []parsers.Test, options Options) (Import, error) {
	m, err := findModule(options.WorkDir)
	if err != nil {
		return Import{}, err
	}
	run := Run{
		Producer: "gavel", GitCommit: options.Provenance.GitCommit, Dirty: options.Provenance.Dirty, GitTree: options.Provenance.GitTree,
		StartedAt: options.StartedAt.UTC(), FinishedAt: options.FinishedAt.UTC(), Tests: []Test{}, SelectionMode: options.SelectionMode,
		InventoryDigest: options.InventoryDigest, Complete: options.Complete, IdentityVerified: options.IdentityVerified, PlanID: options.PlanID}
	run.Inventory = options.Inventory
	seen := map[string]bool{}
	command := ""
	var walk func(tests []parsers.Test) error
	walk = func(tests []parsers.Test) error {
		for _, test := range tests {
			if err := walk(test.Children); err != nil {
				return err
			}
			if test.WorkDir == "" {
				test.WorkDir = options.WorkDir
			}
			// A cached test without attempts is gavel's run-cache standing in for a package it did not
			// run; go test's own cache replays are stamped like fresh results and are kept.
			if test.Framework != parsers.GoTest && test.Framework != parsers.Ginkgo || test.Cached && len(test.Attempts) == 0 ||
				placeholder(test) || !m.holds(test) {
				continue
			}
			exported, err := m.test(test, options.WorkDir)
			if err != nil {
				return err
			}
			key := strings.Join(append([]string{exported.Framework, exported.Package, exported.Name}, exported.Suite...), "\x00")
			if seen[key] {
				// uir rejects an import that reports a test twice; a second test of the same name
				// cannot be told apart from the first.
				logger.Warnf("uir export: %s %s %q is reported twice; exporting the first only", exported.Framework, exported.Package, strings.Join(append(append([]string(nil), exported.Suite...), exported.Name), " > "))
				continue
			}
			seen[key] = true
			if command == "" {
				command = test.Command
			}
			if options.CoverTest != nil && exported.Attempts[len(exported.Attempts)-1].Outcome != Skip {
				profile := filepath.Join(options.CoverDir, fmt.Sprintf("test-%d.out", len(run.Tests)))
				if err := options.CoverTest(test, profile); err != nil {
					logger.Warnf("uir export: no coverage of its own for %s: %v", exported.Name, err)
				} else if exported.Coverage, err = readProfiles([]string{profile}, m); err != nil {
					return err
				} else if !executed(exported.Coverage) {
					// go test writes a profile even when the rerun built nothing or matched no test, so a profile
					// that covers nothing is no evidence of what the test depends on.
					logger.Warnf("uir export: the coverage rerun of %s covered no statements; leaving it without coverage of its own", exported.Name)
					exported.Coverage = nil
				}
			}
			run.Tests = append(run.Tests, exported)
		}
		return nil
	}
	if err := walk(tests); err != nil {
		return Import{}, err
	}
	if options.Environment != nil {
		run.Environment = options.Environment
	} else if run.Environment, err = Environment(m.root, command); err != nil {
		return Import{}, err
	}
	run.InputsDigest = options.InputsDigest
	if run.Coverage, err = readProfiles(options.CoverProfiles, m); err != nil {
		return Import{}, err
	}
	version := options.ContractVersion
	if version == 0 {
		version = LegacyContractVersion
	}
	return Import{Version: version, Runs: []Run{run}}, nil
}

// Write encodes imported to path.
func Write(path string, imported Import) error {
	data, err := json.Marshal(imported)
	if err != nil {
		return fmt.Errorf("encode the uir import: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write the uir import: %w", err)
	}
	return nil
}

// test is the contract's view of one gavel test: its package as an import path, its file relative to
// the module, and one attempt per execution.
func (m *module) test(test parsers.Test, runWorkDir string) (Test, error) {
	workDir := test.WorkDir
	if workDir == "" {
		workDir = runWorkDir
	}
	exported := Test{Framework: string(test.Framework), Package: test.Package, Suite: test.Suite, Name: test.Name}
	if test.PackagePath != "" {
		// A ginkgo test's Package is its directory's base name; derive the import path from its directory.
		if path, ok := m.importPath(filepath.Join(workDir, test.PackagePath)); ok {
			exported.Package = path
		}
	}
	if exported.File = m.relative(workDir, test.File); exported.File != "" {
		exported.Line = test.Line
	}
	for i, attempt := range test.Attempts {
		exported.Attempts = append(exported.Attempts, Attempt{
			Outcome:    outcome(attempt.Passed, attempt.Failed, attempt.Skipped, attempt.TimedOut, test.Aborted && i == len(test.Attempts)-1),
			DurationMS: attempt.Duration.Milliseconds(),
			Failure:    failure(attempt.Failed || attempt.TimedOut, attempt.Message, attempt.Stdout),
		})
	}
	if len(exported.Attempts) == 0 {
		exported.Attempts = []Attempt{{
			Outcome:    outcome(test.Passed, test.Failed, test.Skipped, test.TimedOut, test.Aborted),
			DurationMS: test.Duration.Milliseconds(),
			Failure:    failure(test.Failed || test.TimedOut, test.Message, test.Stdout),
		}}
	}
	if exported.Package == "" {
		return Test{}, fmt.Errorf("uir export: no package for %s %s", test.Framework, test.Name)
	}
	return exported, nil
}

// outcome maps gavel's state flags to the contract's outcome. A test with no final state, still
// pending or running when the run stopped, never finished: aborted.
func outcome(passed, failed, skipped, timedOut, aborted bool) Outcome {
	switch {
	case timedOut:
		return Timeout
	case failed:
		// Ginkgo's Abort() sets both failed and aborted; it is a failure.
		return Fail
	case aborted:
		return Aborted
	case skipped:
		return Skip
	case passed:
		return Pass
	}
	return Aborted
}

// placeholder reports the entries gavel adds for a package or suite rather than a test: build and
// package failures, an empty package, and Ginkgo suite-level setup, teardown and failure nodes.
func placeholder(test parsers.Test) bool {
	switch test.Name {
	case "No test files", "No tests to run", "Build Failed", "Package failed", "Package incomplete", "Go test execution":
		return test.Framework == parsers.GoTest
	}
	return test.Framework == parsers.Ginkgo && strings.HasPrefix(test.Name, "[") && strings.HasSuffix(test.Name, "]")
}

// executed reports whether any block ran.
func executed(blocks []Block) bool {
	for _, block := range blocks {
		if block.Count > 0 {
			return true
		}
	}
	return false
}

// failure is a failed attempt's message, or its output when it has none, as for a go test failure
// reported through t.Error.
func failure(failed bool, message, stdout string) string {
	if !failed {
		return ""
	}
	if message != "" {
		return message
	}
	return stdout
}
