package main

import "github.com/flanksource/clicky"

// runServeDashboard backs `gavel serve` — the PR dashboard web UI. It reuses
// PRListOptions + runPRUI (same path as `pr list --ui`) and forces UI mode so
// the command serves the dashboard rather than printing a PR list.
func runServeDashboard(opts PRListOptions) (any, error) {
	if !opts.MenuBar {
		opts.UI = true
	}
	return nil, runPRUI(opts)
}

func init() {
	cmd := clicky.AddNamedCommand("serve", rootCmd, PRListOptions{}, runServeDashboard)
	cmd.Short = "Serve the PR dashboard web UI (live PR & check status)"
	cmd.Long = "Serve the PR dashboard web UI. Equivalent to `gavel pr list --ui`, " +
		"with --dev for Vite hot-module-reload. Use --menu-bar for the macOS menu-bar indicator."
	// `serve` always implies the web UI; hide the redundant --ui toggle.
	if f := cmd.Flags().Lookup("ui"); f != nil {
		_ = cmd.Flags().MarkHidden("ui")
	}
}
