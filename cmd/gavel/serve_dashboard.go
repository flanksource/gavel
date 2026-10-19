package main

import (
	"os"

	"github.com/flanksource/clicky"
	"github.com/spf13/pflag"
)

// applyServeDefaults adapts PRListOptions for `gavel serve`: the command always
// serves the dashboard (UI mode, unless --menu-bar) and defaults to an org-wide
// fetch when the user hasn't narrowed it to specific repos. Passing repo args
// (e.g. `gavel serve owner/repo`) keeps the search scoped to those repos.
func applyServeDefaults(opts PRListOptions) PRListOptions {
	if !opts.MenuBar {
		opts.UI = true
	}
	opts.retention = true
	if len(opts.Repos) == 0 {
		opts.All = true
	}
	return opts
}

// runServeDashboard backs `gavel serve` — the PR dashboard web UI. It reuses
// PRListOptions + runPRUI (same path as `pr list --ui`) and forces UI mode so
// the command serves the dashboard rather than printing a PR list.
func runServeDashboard(opts DashboardOptions) (any, error) {
	opts = opts.withEnv(os.Getenv)
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	if err := loadServeProperties(opts.PropertiesFile); err != nil {
		return nil, err
	}
	if opts.Auth == authClerk {
		return nil, runHostedDashboard(opts)
	}
	return nil, runPRUI(applyServeDefaults(opts.PRListOptions), serveDatabaseWithMigrations)
}

func init() {
	cmd := clicky.AddNamedCommand("serve", rootCmd, DashboardOptions{}, runServeDashboard)
	cmd.Short = "Serve the PR dashboard web UI (live PR & check status)"
	cmd.Long = "Serve the PR dashboard web UI. Equivalent to `gavel pr list --ui`, " +
		"but defaults to an org-wide (--all) fetch unless you pass repo args to scope it. " +
		"Use --dev for Vite hot-module-reload and --menu-bar for the macOS menu-bar indicator."
	// `serve` always implies the web UI; hide the redundant --ui toggle.
	if f := cmd.Flags().Lookup("ui"); f != nil {
		_ = cmd.Flags().MarkHidden("ui")
	}
	// A credential's fields are separated by commas, which clicky's list flags split on.
	if f := cmd.Flags().Lookup("substrate-credential"); f != nil {
		f.Value = &wholeValues{Value: f.Value, SliceValue: f.Value.(pflag.SliceValue), flag: f}
	}
}

// wholeValues makes a repeated list flag keep each value whole instead of splitting it on commas.
type wholeValues struct {
	pflag.Value
	pflag.SliceValue
	flag *pflag.Flag
}

func (v *wholeValues) Set(value string) error {
	if !v.flag.Changed {
		return v.Replace([]string{value})
	}
	return v.Append(value)
}

func (v *wholeValues) Type() string { return "stringArray" }
