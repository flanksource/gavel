package testui

import (
	"time"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/testrunner/bench"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type SnapshotMetadata struct {
	Version    string         `json:"version,omitempty"`
	Started    time.Time      `json:"started,omitempty"`
	Ended      time.Time      `json:"ended,omitempty"`
	Kind       string         `json:"kind,omitempty"`
	Sequence   int            `json:"sequence,omitempty"`
	Args       map[string]any `json:"args,omitempty"`
	PID        int            `json:"pid,omitempty"`        // gavel's own PID (the parent orchestrator)
	Command    string         `json:"command,omitempty"`    // gavel command line (argv[0] + args)
	Frameworks []string       `json:"frameworks,omitempty"` // test frameworks scheduled for the run (go test, ginkgo, …)
	ExitCode   *int           `json:"exit_code,omitempty"`  // final exit code once the run completes; nil while running
	TimedOut   bool           `json:"timed_out,omitempty"`  // true when the global --timeout fired during the run
}

type SnapshotGit struct {
	Repo string `json:"repo,omitempty"`
	Root string `json:"root,omitempty"`
	SHA  string `json:"sha,omitempty"`
}

type SnapshotStatus struct {
	Running              bool   `json:"running"`
	LintRun              bool   `json:"lint_run,omitempty"`
	DiagnosticsAvailable bool   `json:"diagnostics_available,omitempty"`
	StopSupported        bool   `json:"stop_supported,omitempty"`
	Stopped              bool   `json:"stopped,omitempty"`
	StopMessage          string `json:"stop_message,omitempty"`
}

type Snapshot struct {
	Metadata    *SnapshotMetadata       `json:"metadata,omitempty"`
	Git         *SnapshotGit            `json:"git,omitempty"`
	Status      SnapshotStatus          `json:"status"`
	Tests       []parsers.Test          `json:"tests"`
	Lint        []*linters.LinterResult `json:"lint,omitempty"`
	Bench       *bench.BenchComparison  `json:"bench,omitempty"`
	Diagnostics *DiagnosticsSnapshot    `json:"diagnostics,omitempty"`
}

func (s Snapshot) Pretty() api.Text {
	return parsers.Tests(s.Tests).Sum().Pretty()
}

func cloneSnapshotMetadata(meta *SnapshotMetadata) *SnapshotMetadata {
	if meta == nil {
		return nil
	}
	cloned := *meta
	if meta.Args != nil {
		cloned.Args = make(map[string]any, len(meta.Args))
		for k, v := range meta.Args {
			cloned.Args[k] = v
		}
	}
	if meta.Frameworks != nil {
		cloned.Frameworks = append([]string(nil), meta.Frameworks...)
	}
	if meta.ExitCode != nil {
		code := *meta.ExitCode
		cloned.ExitCode = &code
	}
	return &cloned
}

func cloneSnapshotGit(git *SnapshotGit) *SnapshotGit {
	if git == nil {
		return nil
	}
	cloned := *git
	return &cloned
}
