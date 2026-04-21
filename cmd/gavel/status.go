package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
)

type StatusOptions struct {
	WorkDir   string `flag:"work-dir" help:"Working directory"`
	NoRepomap bool   `flag:"no-repomap" help:"Skip repomap enrichment (faster)"`
}

func (o StatusOptions) Help() api.Textable {
	const (
		heading     = "font-bold text-purple-600"
		flagStyle   = "text-cyan-600 font-bold"
		muted       = "text-muted"
		staged      = "text-green-500 font-bold"
		modified    = "text-yellow-500 font-bold"
		untracked   = "text-purple-500 font-bold"
		conflicted  = "text-red-600 font-bold underline"
		deleted     = "text-red-500 font-bold"
		renamed     = "text-blue-500 font-bold"
	)

	legend := func(sym, style, label string) api.Text {
		return clicky.Text("  ").
			Append(sym, style).
			Append("  ", "").
			Append(label, muted)
	}

	t := clicky.Text("gavel status", heading).Space().
		Append("— list changed files with repomap context", muted).NewLine().NewLine().
		Append("Shows one row per file using ", "").
		Append("Starship git_status", "italic").
		Append(" symbols, plus line deltas, inferred scopes, and repomap findings (Kubernetes refs, architecture violations).", "").
		NewLine().NewLine().
		Append("SYMBOLS", heading).NewLine().
		Add(legend("+", staged, "staged")).NewLine().
		Add(legend("!", modified, "modified (unstaged)")).NewLine().
		Add(legend("?", untracked, "untracked")).NewLine().
		Add(legend("=", conflicted, "conflicted")).NewLine().
		Add(legend("✘", deleted, "deleted")).NewLine().
		Add(legend("»", renamed, "renamed / copied")).NewLine().NewLine().
		Append("FLAGS", heading).NewLine().
		Append("  --work-dir", flagStyle).Append("   path to a git repo (default: cwd)", muted).NewLine().
		Append("  --no-repomap", flagStyle).Append(" skip repomap enrichment (faster)", muted).NewLine().NewLine().
		Append("EXAMPLES", heading).NewLine().
		Append("  gavel status", flagStyle).Append("                 enriched view", muted).NewLine().
		Append("  gavel status --no-repomap", flagStyle).Append("    skip repomap lookup", muted).NewLine()

	return t
}

func init() {
	clicky.AddNamedCommand("status", rootCmd, StatusOptions{}, runStatus)
}

func runStatus(opts StatusOptions) (any, error) {
	workDir := opts.WorkDir
	if workDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		workDir = wd
	}
	if root := repomap.FindGitRoot(workDir); root != "" {
		workDir = root
	}

	return status.Gather(workDir, status.Options{NoRepomap: opts.NoRepomap})
}
