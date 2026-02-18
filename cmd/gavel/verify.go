package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/verify"
)

type VerifyOptions struct {
	Model             string   `json:"model" flag:"model" help:"AI CLI to use: claude, gemini, codex (or model name like gemini-2.5-flash)" default:"claude"`
	CommitRange       string   `json:"range" flag:"range" help:"Commit range to review (e.g. main..HEAD)"`
	DisableChecks     []string `json:"disable-checks" flag:"disable-checks" help:"Check IDs to disable (comma-separated)"`
	DisableCategories []string `json:"disable-categories" flag:"disable-categories" help:"Check categories to disable (comma-separated)"`
	Args              []string `json:"-" args:"true"`
}

func (o VerifyOptions) GetName() string { return "verify" }

func (o VerifyOptions) Help() api.Text {
	return clicky.Text(`AI-powered code review with prescribed checks and rated dimensions.

Reviews git diffs, commit ranges, or specific files using AI CLI tools
(claude, gemini, codex) and returns boolean checks, rated dimensions, and
completeness assessment.

EXAMPLES:
  # Review uncommitted changes
  gavel verify

  # Review a commit range
  gavel verify --range main..HEAD

  # Review specific files
  gavel verify path/to/file.go

  # Use a different AI model
  gavel verify --model gemini

  # Disable specific checks
  gavel verify --disable-checks migration-included,config-changes-documented`)
}

func init() {
	clicky.AddCommand(rootCmd, VerifyOptions{}, func(opts VerifyOptions) (any, error) {
		workDir, err := getWorkingDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}

		cfg, err := verify.LoadConfig(workDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		if opts.Model != "" && opts.Model != "claude" {
			cfg.Model = opts.Model
		}
		if len(opts.DisableChecks) > 0 {
			cfg.Checks.Disabled = append(cfg.Checks.Disabled, opts.DisableChecks...)
		}
		if len(opts.DisableCategories) > 0 {
			cfg.Checks.DisabledCategories = append(cfg.Checks.DisabledCategories, opts.DisableCategories...)
		}

		result, err := verify.RunVerify(verify.RunOptions{
			Config:      cfg,
			RepoPath:    workDir,
			Args:        opts.Args,
			CommitRange: opts.CommitRange,
		})
		if err != nil {
			return nil, err
		}

		return result, nil
	})
}
