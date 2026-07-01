package main

import "github.com/flanksource/clicky"

// applyServeDefaults adapts PRListOptions for `gavel serve`: the command always
// serves the dashboard (UI mode, unless --menu-bar) and defaults to an org-wide
// fetch when the user hasn't narrowed it to specific repos. Passing repo args
// (e.g. `gavel serve owner/repo`) keeps the search scoped to those repos.
func applyServeDefaults(opts PRListOptions) PRListOptions {
	if !opts.MenuBar {
		opts.UI = true
	}
	if len(opts.Repos) == 0 {
		opts.All = true
	}
	return opts
}

// runServeDashboard backs `gavel serve` — the PR dashboard web UI. It reuses
// PRListOptions + runPRUI (same path as `pr list --ui`) and forces UI mode so
// the command serves the dashboard rather than printing a PR list.
func runServeDashboard(opts PRListOptions) (any, error) {
	return nil, runPRUI(applyServeDefaults(opts))
}

func init() {
	cmd := clicky.AddNamedCommand("serve", rootCmd, PRListOptions{}, runServeDashboard)
	cmd.Short = "Serve the PR dashboard web UI (live PR & check status)"
	cmd.Long = "Serve the PR dashboard web UI. Equivalent to `gavel pr list --ui`, " +
		"but defaults to an org-wide (--all) fetch unless you pass repo args to scope it. " +
		"Use --dev for Vite hot-module-reload and --menu-bar for the macOS menu-bar indicator."
	// `serve` always implies the web UI; hide the redundant --ui toggle.
	if f := cmd.Flags().Lookup("ui"); f != nil {
		_ = cmd.Flags().MarkHidden("ui")
	}
}
