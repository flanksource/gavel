package main

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/verify"
)

type VerifyOptions struct {
	Model       string   `json:"model" flag:"model" help:"AI CLI to use: claude, gemini, codex (or model name like gemini-2.5-flash)" default:"claude"`
	CommitRange string   `json:"range" flag:"range" help:"Commit range to review (e.g. main..HEAD)"`
	Sections    []string `json:"sections" flag:"sections" help:"Sections to evaluate (comma-separated)"`
	Args        []string `json:"-" args:"true"`
}

func (o VerifyOptions) GetName() string { return "verify" }

func (o VerifyOptions) Help() api.Text {
	return clicky.Text(`AI-powered code review with structured scoring.

Reviews git diffs, commit ranges, or specific files using AI CLI tools
(claude, gemini, codex) and returns per-section scores.

EXAMPLES:
  # Review uncommitted changes
  gavel verify

  # Review a commit range
  gavel verify --range main..HEAD

  # Review specific files
  gavel verify path/to/file.go

  # Use a different AI model
  gavel verify --model gemini`)
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
		if len(opts.Sections) > 0 {
			cfg.Sections = opts.Sections
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
