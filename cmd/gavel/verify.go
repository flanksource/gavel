package main

import (
	"fmt"
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
)

type VerifyOptions struct {
	Model          string   `json:"model" flag:"model" help:"AI CLI to use: claude, gemini, codex (or model name like gemini-2.5-flash)" default:"claude"`
	CommitRange    string   `json:"range" flag:"range" help:"Commit range to review (e.g. main..HEAD)"`
	DisableChecks  []string `json:"disable-checks" flag:"disable-checks" help:"Check IDs to disable (comma-separated)"`
	Completeness   bool     `json:"completeness" flag:"completeness" help:"Enable completeness checks" default:"true"`
	CodeQuality    bool     `json:"code-quality" flag:"code-quality" help:"Enable code quality checks" default:"true"`
	Testing        bool     `json:"testing" flag:"testing" help:"Enable testing checks" default:"true"`
	Consistency    bool     `json:"consistency" flag:"consistency" help:"Enable consistency checks" default:"true"`
	Security       bool     `json:"security" flag:"security" help:"Enable security checks" default:"true"`
	Performance    bool     `json:"performance" flag:"performance" help:"Enable performance checks" default:"true"`
	AutoFix        bool     `json:"auto-fix" flag:"auto-fix" help:"Enable iterative AI-powered fix loop"`
	FixModel       string   `json:"fix-model" flag:"fix-model" help:"AI CLI to use for fixes (defaults to --model)"`
	MaxTurns       int      `json:"max-turns" flag:"max-turns" help:"Maximum verify-fix cycles" default:"3"`
	ScoreThreshold int      `json:"score-threshold" flag:"score-threshold" help:"Exit 0 if final score >= this value" default:"80"`
	PatchOnly      bool     `json:"patch-only" flag:"patch-only" help:"AI outputs patches instead of interactive tool-use"`
	SyncTodos      bool     `json:"sync-todos" flag:"sync-todos" help:"Create/update TODO files from verify findings"`
	Args           []string `json:"-" args:"true"`
}

func (o VerifyOptions) GetName() string { return "verify" }

func (o VerifyOptions) Help() api.Text {
	return clicky.Text(`AI-powered code review with prescribed checks and rated dimensions.

Reviews git diffs, commit ranges, PRs, branches, or specific files using AI
CLI tools (claude, gemini, codex) and returns boolean checks, rated dimensions,
and completeness assessment.

Arguments are auto-detected: commit SHAs, ranges (main..HEAD), branch names,
PR references (#123 or bare numbers), date ranges, files, and directories.

EXAMPLES:
  # Review uncommitted changes
  gavel verify

  # Review a commit range
  gavel verify main..HEAD

  # Review a specific commit
  gavel verify abc1234

  # Review branch diff against current
  gavel verify main

  # Review a GitHub PR
  gavel verify #123
  gavel verify 42

  # Review specific files
  gavel verify path/to/file.go

  # Review a directory
  gavel verify pkg/verify/

  # Review commits in a date range
  gavel verify 2024-01-01..2024-06-01

  # Auto-fix findings iteratively
  gavel verify --auto-fix

  # Sync findings to TODO files
  gavel verify --sync-todos`)
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
		for cat, enabled := range map[string]bool{
			"completeness": opts.Completeness,
			"code-quality": opts.CodeQuality,
			"testing":      opts.Testing,
			"consistency":  opts.Consistency,
			"security":     opts.Security,
			"performance":  opts.Performance,
		} {
			if !enabled {
				cfg.Checks.DisabledCategories = append(cfg.Checks.DisabledCategories, cat)
			}
		}

		runOpts := verify.RunOptions{
			Config:      cfg,
			RepoPath:    workDir,
			Args:        opts.Args,
			CommitRange: opts.CommitRange,
		}

		var verifyResult *verify.VerifyResult
		var response any

		if opts.AutoFix {
			fixResult, err := verify.RunAutoFix(runOpts, verify.AutoFixOptions{
				FixModel:       opts.FixModel,
				MaxTurns:       opts.MaxTurns,
				ScoreThreshold: opts.ScoreThreshold,
				PatchOnly:      opts.PatchOnly,
			})
			if err != nil {
				return nil, err
			}
			response = fixResult
			if len(fixResult.Turns) > 0 {
				verifyResult = fixResult.Turns[len(fixResult.Turns)-1].Result
			}
		} else {
			result, err := verify.RunVerify(runOpts)
			if err != nil {
				return nil, err
			}
			verifyResult = result
			response = result
		}

		if opts.SyncTodos && verifyResult != nil {
			syncResult, err := SyncTodos(verifyResult, SyncOptions{
				TodosDir:       filepath.Join(workDir, ".todos", "verify"),
				ScoreThreshold: opts.ScoreThreshold,
				RepoPath:       workDir,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to sync todos: %w", err)
			}
			logger.Infof("%s", clicky.MustFormat(syncResult.Pretty(), clicky.FormatOptions{Pretty: true}))
		}

		return response, nil
	})
}
