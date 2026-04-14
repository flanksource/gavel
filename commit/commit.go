package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

var (
	ErrNothingStaged  = errors.New("nothing staged to commit")
	ErrLLMUnavailable = errors.New("LLM agent unavailable")
	ErrInvalidStage   = errors.New("invalid --stage value")
)

const (
	StageStaged   = "staged"
	StageUnstaged = "unstaged"
	StageAll      = "all"

	testEnvVar  = "GAVEL_COMMIT_TEST"
	stubMessage = "chore: fixture stub"
)

type Options struct {
	WorkDir string
	Stage   string
	DryRun  bool
	Force   bool
	NoCache bool
	Model   string
	Message string
	Config  verify.CommitConfig
}

type Result struct {
	Message string       `json:"message"`
	Hash    string       `json:"hash,omitempty"`
	DryRun  bool         `json:"dry_run,omitempty"`
	Staged  []string     `json:"staged,omitempty"`
	Hooks   []HookResult `json:"hooks,omitempty"`
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.Stage == "" {
		opts.Stage = StageStaged
	}
	if opts.WorkDir == "" {
		return nil, errors.New("commit.Run: WorkDir is required")
	}

	if err := stageFiles(opts.WorkDir, opts.Stage); err != nil {
		return nil, fmt.Errorf("stage files (%s): %w", opts.Stage, err)
	}

	files, err := stagedFiles(opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("list staged files: %w", err)
	}
	if len(files) == 0 {
		return nil, ErrNothingStaged
	}

	diff, err := stagedDiff(opts.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("read staged diff: %w", err)
	}
	if strings.TrimSpace(diff) == "" {
		return nil, fmt.Errorf("staged file list was non-empty but diff is empty: %v", files)
	}

	result := &Result{Staged: files, DryRun: opts.DryRun}

	if !opts.Force {
		hookResults, hookErr := RunHooks(opts.WorkDir, opts.Config.Hooks, files)
		result.Hooks = hookResults
		if hookErr != nil {
			return result, hookErr
		}
	} else if len(opts.Config.Hooks) > 0 {
		logger.Infof("Skipping %d commit hook(s) due to --force", len(opts.Config.Hooks))
	}

	msg, err := generateMessage(ctx, opts, diff)
	if err != nil {
		return result, fmt.Errorf("generate commit message: %w", err)
	}
	result.Message = msg

	if opts.DryRun {
		logger.Infof("[dry-run] would commit with message:\n%s", msg)
		return result, nil
	}

	hash, err := commitWithMessage(opts.WorkDir, msg)
	if err != nil {
		return result, fmt.Errorf("create commit: %w", err)
	}
	result.Hash = hash
	logger.Infof("Committed %s: %s", shortHash(hash), firstLine(msg))
	return result, nil
}

func stageFiles(workDir, mode string) error {
	switch mode {
	case StageStaged:
		return nil
	case StageUnstaged, StageAll:
		opts := changegraph.DiffOptions{IncludeUnstaged: true}
		if mode == StageAll {
			opts.IncludeUntracked = true
		}
		fs, err := changegraph.ComputeFileSet(workDir, opts)
		if err != nil {
			return fmt.Errorf("compute file set: %w", err)
		}
		return addFiles(workDir, fs.Sorted())
	default:
		return fmt.Errorf("%w: %q", ErrInvalidStage, mode)
	}
}

func generateMessage(ctx context.Context, opts Options, diff string) (string, error) {
	if opts.Message != "" {
		return opts.Message, nil
	}

	if os.Getenv(testEnvVar) == "1" {
		logger.V(1).Infof("%s=1, returning stub commit message", testEnvVar)
		return stubMessage, nil
	}

	cfg := ai.DefaultConfig()
	if opts.Model != "" {
		cfg.Model = opts.Model
	} else if opts.Config.Model != "" {
		cfg.Model = opts.Config.Model
	}
	if opts.NoCache {
		cfg.NoCache = true
	}

	agent, err := gavelai.NewAgent(cfg)
	if err != nil {
		return "", fmt.Errorf("%w: %w (set ANTHROPIC_API_KEY / CLAUDE_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY, or pass -m)", ErrLLMUnavailable, err)
	}

	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	analyzed, err := git.AnalyzeWithAI(ctx, analysis, agent, git.AnalyzeOptions{})
	if err != nil {
		return "", err
	}

	out := models.AIAnalysisOutput{
		Type:    analyzed.CommitType,
		Scope:   analyzed.Scope,
		Subject: analyzed.Subject,
		Body:    analyzed.Body,
	}
	return strings.TrimSpace(out.String()), nil
}

func shortHash(h string) string {
	if len(h) > 8 {
		return h[:8]
	}
	return h
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
