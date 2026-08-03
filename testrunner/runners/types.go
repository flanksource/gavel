package runners

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Runner defines the interface for a test framework runner.
type Runner interface {
	// Name returns the framework name.
	Name() parsers.Framework

	// Parser returns the result parser for this runner.
	Parser() parsers.ResultParser

	// Detect checks if this framework is used in the given working directory.
	Detect(workDir string) (bool, error)

	// DiscoverPackages returns packages with tests for this framework.
	// When recursive is true, it walks subdirectories; otherwise only the given directory.
	DiscoverPackages(workDir string, recursive bool) ([]string, error)

	// BuildCommand builds the command to run tests for a package.
	// extraArgs are additional arguments passed to the test runner.
	BuildCommand(packagePath string, extraArgs ...string) (*TestRun, error)
}

// FocusMapper is an optional runner capability for translating a common test
// focus pattern into framework-native command-line arguments. Runners that do
// not implement it are excluded from top-level focused runs.
type FocusMapper interface {
	FocusArgs(pattern string) []string
}

// BuildTagsMapper is an optional runner capability for translating Go build
// tags into framework-native command-line arguments.
type BuildTagsMapper interface {
	BuildTagsArgs(tags []string) []string
}

type Package string
type TestRun struct {
	Framework  parsers.Framework
	Package    Package
	Parser     parsers.ResultParser
	Process    *exec.Process
	ReportPath string // Optional: path to JSON report file (used by Ginkgo)
}

func (tr TestRun) Pretty() api.Text {
	s := clicky.Text(string(tr.Package), "bold")
	s = s.Append(" (").Append(tr.Framework).Append(") ")
	s = s.Space().Append(tr.Process)
	return s
}
