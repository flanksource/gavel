// Package lint contains the reusable lint orchestration used by the gavel CLI
// and by other packages (e.g. fixtures/types) that need to run linters
// programmatically. It owns linter discovery, executable resolution, execution
// fan-out across project roots, and the post-run filter cascade. CLI-only
// concerns (cobra wiring, the --ui dashboard, interactive triage, TODO sync,
// snapshot persistence, and the AI-fix loop) stay in cmd/gavel and call into
// this package.
package lint

import (
	"context"
	"fmt"
	"io"
	"strings"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

type Options struct {
	Linters      []string        `flag:"linters" yaml:"linters,omitempty" help:"Only run the named linters (comma-separated or repeated). Empty = run every detected linter. Unknown names hard-fail."`
	Ignore       []string        `flag:"ignore" yaml:"ignore,omitempty" help:"Glob patterns to exclude from linting"`
	Triage       bool            `flag:"triage" yaml:"triage,omitempty" help:"Interactive mode to select violation types to ignore"`
	Fix          bool            `flag:"fix" yaml:"fix,omitempty" help:"Enable auto-fixing"`
	NoCache      bool            `flag:"no-lint-cache" yaml:"no-lint-cache,omitempty" help:"Disable linter result caching/debounce"`
	Timeout      string          `flag:"timeout" yaml:"timeout,omitempty" help:"Timeout per linter (e.g. 5m, 30s)" default:"5m"`
	SyncTodos    string          `flag:"sync-todos" yaml:"sync-todos,omitempty" help:"Sync violations to TODO files in directory (default: .todos/lint)"`
	GroupBy      string          `flag:"group-by" yaml:"group-by,omitempty" help:"Group synced TODOs by: file, package, message" default:"file"`
	WorkDir      string          `flag:"work-dir" yaml:"work-dir,omitempty" help:"Working directory"`
	Changed      bool            `flag:"changed" yaml:"changed,omitempty" help:"Only report new issues vs origin/main (or $GAVEL_CHANGED_BASE)"`
	Since        string          `flag:"since" yaml:"since,omitempty" help:"Only report new issues since <ref> (merge-base with HEAD)"`
	UI           bool            `flag:"ui" yaml:"ui,omitempty" help:"Launch browser UI to view violations"`
	Addr         string          `flag:"addr" yaml:"addr,omitempty" help:"Interface to bind --ui HTTP server. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine." default:"0.0.0.0"`
	DryRun       bool            `flag:"dry-run" yaml:"dry-run,omitempty" help:"Print the linter commands that would run without executing them"`
	Baseline     string          `flag:"baseline" yaml:"baseline,omitempty" help:"Path to previous results JSON; only report NEW violations not in baseline"`
	Failed       string          `flag:"failed" yaml:"failed,omitempty" help:"Path to previous results JSON; re-run only linters/files that had violations"`
	Summary      bool            `flag:"summary" yaml:"summary,omitempty" help:"Collapse output: group by linter -> rule, show count and the first --summary-limit locations"`
	SummaryLimit int             `flag:"summary-limit" yaml:"summary-limit,omitempty" help:"Max example locations shown per rule in --summary mode" default:"5"`
	Files        []string        `args:"true" yaml:"files,omitempty"`
	OutputTee    io.Writer       `json:"-" yaml:"-"`
	Context      context.Context `json:"-" yaml:"-"`

	AIFix         bool `flag:"ai-fix" yaml:"ai-fix,omitempty" help:"Invoke the AI configured by 'captain configure' to fix violations and re-lint until clean (or bounded by --ai-fix-max-iterations / --budget)"`
	AIFixMaxIters int  `flag:"ai-fix-max-iterations" yaml:"ai-fix-max-iterations,omitempty" help:"Max AI→re-lint cycles" default:"3"`
	Yes           bool `flag:"yes" short:"y" yaml:"yes,omitempty" help:"Assume yes: auto-AI-fix lint violations (implies --ai-fix). Does not enable --triage."`

	// Embedded: contributes --model, --backend, --api-key, --no-cache,
	// --budget, --debug, --max-tokens, --temperature, --permission-mode,
	// --edit, --allowed-tools, --disallowed-tools, --mcp, --hooks,
	// --skills, --skill-dir, --user, --project, --memory, --bare.
	// Defaults overlay from ~/.captain.yaml via captain configure.
	captaincli.AIRuntimeOptions
}

func (o Options) Pretty() api.Text {
	t := clicky.Text("")
	if o.WorkDir != "" {
		t = t.Append("WorkDir: ", "text-muted").Append(o.WorkDir, "text-blue-500").Space()
	}
	if len(o.Linters) > 0 {
		t = t.Append("Linters: ", "text-muted").Append(strings.Join(o.Linters, ","), "text-blue-500").Space()
	}
	if o.Fix {
		t = t.Append("Fix: on", "text-green-500").Space()
	}
	if o.AIFix {
		label := "AIFix: on"
		if o.Model != "" {
			label = fmt.Sprintf("AIFix: %s", o.Model)
		}
		t = t.Append(label, "text-green-500").Space()
	}
	if o.Timeout != "" {
		t = t.Append("Timeout: ", "text-muted").Append(o.Timeout, "text-blue-500").Space()
	}
	if len(o.Files) > 0 {
		t = t.Append("Files: ", "text-muted").Append(clicky.CompactList(o.Files), "text-blue-500")
	}
	return t
}

func (o Options) Help() string {
	return `Run linters on the project.

Automatically detects which linters are available and runs them.
Supports: golangci-lint, ruff, eslint, oxlint, react-doctor, pyright, tsc, markdownlint, vale, jscpd, betterleaks.

Examples:
  gavel lint
  gavel lint jscpd
  gavel lint jscpd eslint
  gavel lint react-doctor
  gavel lint secrets                 # alias for betterleaks
  gavel lint --linters=golangci-lint
  gavel lint --linters=golangci-lint,ruff
  gavel lint --fix
  gavel lint --triage
  gavel lint -y                      # auto-AI-fix violations (implies --ai-fix)
  gavel lint ./pkg/...`
}
