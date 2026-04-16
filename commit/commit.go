package commit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flanksource/clicky"
	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/internal/changegraph"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
)

var (
	ErrNothingStaged        = errors.New("nothing staged to commit")
	ErrLLMUnavailable       = errors.New("LLM agent unavailable")
	ErrInvalidStage         = errors.New("invalid --stage value")
	ErrCommitAllWithMessage = errors.New("--commit-all does not support --message")
	ErrInvalidCommitAllPlan = errors.New("invalid commit-all plan")

	newAgentFunc                      = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) { return gavelai.NewAgent(cfg) }
	analyzeCommitWithAIFunc           = git.AnalyzeWithAI
	planCommitGroupsFunc              = planCommitGroups
	dryRunOutput            io.Writer = os.Stdout
)

const (
	StageStaged   = "staged"
	StageUnstaged = "unstaged"
	StageAll      = "all"

	testEnvVar  = "GAVEL_COMMIT_TEST"
	stubMessage = "chore: fixture stub"
)

type Options struct {
	WorkDir   string
	Stage     string
	CommitAll bool
	Max       int
	DryRun    bool
	Force     bool
	NoCache   bool
	Model     string
	Message   string
	Config    verify.CommitConfig
}

type CommitResult struct {
	Label   string   `json:"label,omitempty"`
	Message string   `json:"message"`
	Hash    string   `json:"hash,omitempty"`
	Files   []string `json:"files,omitempty"`
}

type Result struct {
	Message string         `json:"message"`
	Hash    string         `json:"hash,omitempty"`
	DryRun  bool           `json:"dry_run,omitempty"`
	Staged  []string       `json:"staged,omitempty"`
	Hooks   []HookResult   `json:"hooks,omitempty"`
	Commits []CommitResult `json:"commits,omitempty"`
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.Stage == "" {
		opts.Stage = StageStaged
	}
	if opts.WorkDir == "" {
		return nil, errors.New("commit.Run: WorkDir is required")
	}
	if opts.CommitAll {
		if opts.Message != "" {
			return nil, ErrCommitAllWithMessage
		}
		return runCommitAll(ctx, opts)
	}
	return runSingleCommit(ctx, opts)
}

func runSingleCommit(ctx context.Context, opts Options) (*Result, error) {
	if err := stageFiles(opts.WorkDir, opts.Stage); err != nil {
		return nil, fmt.Errorf("stage files (%s): %w", opts.Stage, err)
	}

	source, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	result := &Result{Staged: source.Files, DryRun: opts.DryRun}

	if !opts.Force {
		hookResults, hookErr := RunHooks(opts.WorkDir, opts.Config.Hooks, source.Files)
		result.Hooks = hookResults
		if hookErr != nil {
			return result, hookErr
		}
	} else if len(opts.Config.Hooks) > 0 {
		logger.Infof("Skipping %d commit hook(s) due to --force", len(opts.Config.Hooks))
	}

	source, err = readStagedSource(opts.WorkDir)
	if err != nil {
		return result, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}
	result.Staged = source.Files

	msg, err := generateMessage(ctx, opts, source.Diff)
	if err != nil {
		return result, fmt.Errorf("generate commit message: %w", err)
	}
	result.Message = msg
	result.Commits = []CommitResult{{
		Message: msg,
		Files:   source.Files,
	}}

	if opts.DryRun {
		printDryRunPreview(result)
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

func runCommitAll(ctx context.Context, opts Options) (*Result, error) {
	if err := stageCommitAllSource(opts.WorkDir); err != nil {
		return nil, err
	}

	source, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	result := &Result{Staged: source.Files, DryRun: opts.DryRun}

	if !opts.Force {
		hookResults, hookErr := RunHooks(opts.WorkDir, opts.Config.Hooks, source.Files)
		result.Hooks = hookResults
		if hookErr != nil {
			return result, hookErr
		}
	} else if len(opts.Config.Hooks) > 0 {
		logger.Infof("Skipping %d commit hook(s) due to --force", len(opts.Config.Hooks))
	}

	source, err = readStagedSource(opts.WorkDir)
	if err != nil {
		return result, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}
	result.Staged = source.Files

	groupSpecs, err := planCommitGroupsFunc(ctx, opts, source.Changes)
	if err != nil {
		return result, fmt.Errorf("plan commit groups: %w", err)
	}

	groups, err := validateCommitPlan(groupSpecs, source.Changes)
	if err != nil {
		return result, err
	}

	result.Commits = make([]CommitResult, 0, len(groups))
	for _, group := range groups {
		msg, msgErr := generateMessage(ctx, opts, group.diff())
		if msgErr != nil {
			return result, fmt.Errorf("generate commit message for %s: %w", group.labelOrDefault(), msgErr)
		}
		result.Commits = append(result.Commits, CommitResult{
			Label:   group.Label,
			Message: msg,
			Files:   group.Files(),
		})
	}

	if opts.DryRun {
		printDryRunPreview(result)
		return result, nil
	}

	if err := resetFiles(opts.WorkDir, source.GitPaths()); err != nil {
		return result, fmt.Errorf("reset staged files: %w", err)
	}

	for i, group := range groups {
		if err := addFiles(opts.WorkDir, group.GitPaths()); err != nil {
			return result, fmt.Errorf("stage commit group %s: %w", group.labelOrDefault(), err)
		}
		hash, commitErr := commitWithMessage(opts.WorkDir, result.Commits[i].Message)
		if commitErr != nil {
			return result, fmt.Errorf("create commit for %s: %w", group.labelOrDefault(), commitErr)
		}
		result.Commits[i].Hash = hash
		logger.Infof("Committed %s: %s", shortHash(hash), firstLine(result.Commits[i].Message))
	}

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

func stageCommitAllSource(workDir string) error {
	files, err := stagedFiles(workDir)
	if err != nil {
		return fmt.Errorf("list staged files: %w", err)
	}
	if len(files) > 0 {
		return nil
	}
	if err := stageFiles(workDir, StageAll); err != nil {
		return fmt.Errorf("stage files (%s): %w", StageAll, err)
	}
	return nil
}

func generateMessage(ctx context.Context, opts Options, diff string) (string, error) {
	if opts.Message != "" {
		return opts.Message, nil
	}

	if os.Getenv(testEnvVar) == "1" {
		logger.V(1).Infof("%s=1, returning stub commit message", testEnvVar)
		return stubMessage, nil
	}

	agent, err := buildAgent(opts)
	if err != nil {
		return "", err
	}
	return generateMessageWithAgent(ctx, diff, agent)
}

func generateMessageWithAgent(ctx context.Context, diff string, agent clickyai.Agent) (string, error) {
	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	analyzed, err := analyzeCommitWithAIFunc(ctx, analysis, agent, git.AnalyzeOptions{})
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

func buildAgent(opts Options) (clickyai.Agent, error) {
	cfg := clickyai.DefaultConfig()
	if opts.Model != "" {
		cfg.Model = opts.Model
	} else if opts.Config.Model != "" {
		cfg.Model = opts.Config.Model
	}
	if opts.NoCache {
		cfg.NoCache = true
	}

	agent, err := newAgentFunc(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (set ANTHROPIC_API_KEY / CLAUDE_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY, or pass -m)", ErrLLMUnavailable, err)
	}
	return agent, nil
}

func commitLabel(commit CommitResult) string {
	if commit.Label != "" {
		return commit.Label
	}
	if len(commit.Files) == 1 {
		return commit.Files[0]
	}
	return fmt.Sprintf("%d files", len(commit.Files))
}

func printDryRunPreview(result *Result) {
	if result == nil || len(result.Commits) == 0 {
		return
	}
	fmt.Fprintln(dryRunOutput, renderDryRunPreview(result).ANSI())
}

func renderDryRunPreview(result *Result) api.Text {
	t := clicky.Text("DRY RUN", "font-bold text-yellow-600").
		Append(" ", "").
		Append(fmt.Sprintf("would create %d commit(s)", len(result.Commits)), "text-muted").
		NewLine()

	for i, commit := range result.Commits {
		if i > 0 {
			t = t.NewLine()
		}
		t = t.Add(renderDryRunCommit(i, len(result.Commits), commit))
	}
	return t
}

func renderDryRunCommit(index, total int, commit CommitResult) api.Text {
	parsed := git.NewCommit(commit.Message)
	ref := fmt.Sprintf("dry-run/%d", index+1)
	if total > 1 {
		ref = fmt.Sprintf("%s of %d", ref, total)
	}

	t := clicky.Text("").
		Append("commit ", "text-orange-500").
		Append(ref, "text-yellow-600").
		NewLine()

	if commit.Label != "" {
		t = t.Append("Label: ", "text-muted").Append(commit.Label, "font-bold").NewLine()
	}

	t = t.Append("    ", "").Add(parsed.PrettySubject()).NewLine()

	if parsed.Body != "" {
		for _, line := range strings.Split(parsed.Body, "\n") {
			t = t.Append("    ", "").Append(line).NewLine()
		}
	}

	if len(commit.Files) > 0 {
		t = t.Append("    ", "").Append("Files:", "text-muted").NewLine()
		for _, file := range commit.Files {
			t = t.Append("      ", "").Append(file, "font-mono text-green-600").NewLine()
		}
	}

	return t
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
