package runners

import (
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Runner defines the interface for a test framework runner.
type Runner interface {
	// Name returns the framework name.
	Name() parsers.Framework

	// Parser returns the result parser for this runner.
	Parser() parsers.ResultParser

	// Flags lists the runner's own flag spellings, parsed after SharedFlags.
	Flags() []FlagSpec

	// Detect parses argv (starting paths, then -- and runner arguments)
	// against cwd and reports whether the framework is present under cwd or
	// the starting paths. A nil result means it is not.
	Detect(cwd string, argv []string) (*TestOptions, error)

	// DiscoverPackages returns the Cwd-relative packages opts selects.
	DiscoverPackages(opts TestOptions) ([]string, error)

	// BuildCommand builds the command that runs opts.Packages.
	BuildCommand(opts TestOptions) (*TestRun, error)
}

// detect is the shared Detect: parse argv, then run present under cwd or
// under the starting paths, or owns one of the test files.
func detect(r Runner, cwd string, argv []string, present func(dir string) (bool, error), owner fileOwner) (*TestOptions, error) {
	opts, err := ParseTestArgs(cwd, argv, r.Flags()...)
	if err != nil {
		return nil, err
	}
	files, err := opts.selectedFiles()
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if _, ok, err := owner(filepath.Join(cwd, file)); err != nil || ok {
			if err != nil {
				return nil, err
			}
			return opts, nil
		}
	}
	dirs := []string{cwd}
	if len(opts.Paths) > 0 || len(files) > 0 {
		dirs = nil
		for _, path := range opts.Paths {
			dirs = append(dirs, filepath.Join(cwd, path))
		}
	}
	for _, dir := range dirs {
		found, err := present(dir)
		if err != nil {
			return nil, err
		}
		if found {
			return opts, nil
		}
	}
	return nil, nil
}

type Package string
type TestRun struct {
	Slots      slots.Limits
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
