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

	newAgentFunc                                    = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) { return gavelai.NewAgent(cfg) }
	analyzeCommitMessageWithAIFunc                  = git.AnalyzeWithAI
	analyzeCompatibilityPromptsWithAIFunc           = git.AnalyzeCompatibilityPromptsWithAI
	dryRunOutput                          io.Writer = os.Stdout
)

const (
	defaultMaxFiles = 7
	defaultMaxLines = 500
)

const (
	StageStaged   = "staged"
	StageUnstaged = "unstaged"
	StageAll      = "all"

	testEnvVar  = "GAVEL_COMMIT_TEST"
	stubMessage = "chore: fixture stub"
)

type Options struct {
	WorkDir       string
	Stage         string
	CommitAll     bool
	MaxFiles      int
	MaxLines      int
	DryRun        bool
	Force         bool
	NoCache       bool
	Push          bool
	Model         string
	Message       string
	PrecommitMode string
	CompatMode    string
	Config        verify.CommitConfig
}

type CommitResult struct {
	Label                string   `json:"label,omitempty"`
	Message              string   `json:"message"`
	Hash                 string   `json:"hash,omitempty"`
	Files                []string `json:"files,omitempty"`
	FunctionalityRemoved []string `json:"functionality_removed,omitempty"`
	CompatibilityIssues  []string `json:"compatibility_issues,omitempty"`
}

type Result struct {
	Message string         `json:"message"`
	Hash    string         `json:"hash,omitempty"`
	DryRun  bool           `json:"dry_run,omitempty"`
	Staged  []string       `json:"staged,omitempty"`
	Hooks   []HookResult   `json:"hooks,omitempty"`
	Commits []CommitResult `json:"commits,omitempty"`
}

type commitAIAnalysis struct {
	Message              string
	FunctionalityRemoved []string
	CompatibilityIssues  []string
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.Stage == "" {
		opts.Stage = StageStaged
	}
	if opts.WorkDir == "" {
		return nil, errors.New("commit.Run: WorkDir is required")
	}
	precommitMode, err := resolvePrecommitMode(opts.PrecommitMode, opts.Config)
	if err != nil {
		return nil, err
	}
	opts.PrecommitMode = precommitMode

	compatMode, err := resolveCompatMode(opts.CompatMode, opts.Config)
	if err != nil {
		return nil, err
	}
	opts.CompatMode = compatMode

	var (
		result *Result
	)
	if opts.CommitAll {
		if opts.Message != "" {
			return nil, ErrCommitAllWithMessage
		}
		if opts.MaxFiles == 0 {
			opts.MaxFiles = defaultMaxFiles
		}
		if opts.MaxLines == 0 {
			opts.MaxLines = defaultMaxLines
		}
		result, err = runCommitAll(ctx, opts)
	} else {
		result, err = runSingleCommit(ctx, opts)
	}
	if err != nil {
		return result, err
	}
	if opts.Push {
		if perr := pushAfterCommit(ctx, opts, result); perr != nil {
			return result, perr
		}
	}
	return result, nil
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

	source, err = applyGitIgnoreCheck(ctx, opts, source)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	source, err = applyFileSizeCheck(ctx, opts, source)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	source, err = applyLinkedDepsCheck(ctx, opts, source)
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

	analysis, err := generateCommitAnalysis(ctx, opts, source.Diff)
	if err != nil {
		return result, fmt.Errorf("generate commit analysis: %w", err)
	}
	result.Message = analysis.Message
	result.Commits = []CommitResult{{
		Message:              analysis.Message,
		Files:                source.Files,
		FunctionalityRemoved: analysis.FunctionalityRemoved,
		CompatibilityIssues:  analysis.CompatibilityIssues,
	}}

	if opts.DryRun {
		printDryRunPreview(result)
		return result, nil
	}

	if err := applyCompatibilityCheck(ctx, opts, result.Commits[0]); err != nil {
		return result, err
	}

	hash, err := commitWithMessage(opts.WorkDir, analysis.Message)
	if err != nil {
		return result, fmt.Errorf("create commit: %w", err)
	}
	result.Hash = hash
	logger.Infof("Committed %s: %s", shortHash(hash), firstLine(result.Message))
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

	source, err = applyGitIgnoreCheck(ctx, opts, source)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	source, err = applyFileSizeCheck(ctx, opts, source)
	if err != nil {
		return nil, err
	}
	if len(source.Files) == 0 {
		return nil, ErrNothingStaged
	}

	source, err = applyLinkedDepsCheck(ctx, opts, source)
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

	groups := groupChangesByDir(source.Changes, opts.MaxFiles, opts.MaxLines)
	if len(groups) == 0 {
		return result, ErrNothingStaged
	}

	result.Commits = make([]CommitResult, 0, len(groups))
	for _, group := range groups {
		analysis, msgErr := generateCommitAnalysis(ctx, opts, group.diff())
		if msgErr != nil {
			return result, fmt.Errorf("generate commit analysis for %s: %w", group.labelOrDefault(), msgErr)
		}
		result.Commits = append(result.Commits, CommitResult{
			Label:                group.Label,
			Message:              analysis.Message,
			Files:                group.Files(),
			FunctionalityRemoved: analysis.FunctionalityRemoved,
			CompatibilityIssues:  analysis.CompatibilityIssues,
		})
	}

	if opts.DryRun {
		printDryRunPreview(result)
		return result, nil
	}

	for _, commit := range result.Commits {
		if err := applyCompatibilityCheck(ctx, opts, commit); err != nil {
			return result, err
		}
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

func generateCommitAnalysis(ctx context.Context, opts Options, diff string) (commitAIAnalysis, error) {
	if os.Getenv(testEnvVar) == "1" {
		logger.V(1).Infof("%s=1, returning stub commit analysis", testEnvVar)
		msg := strings.TrimSpace(opts.Message)
		if msg == "" {
			msg = stubMessage
		}
		return commitAIAnalysis{Message: msg}, nil
	}
	explicitMessage := strings.TrimSpace(opts.Message)
	if explicitMessage != "" && !shouldRunCompatibilityAnalysis(opts.CompatMode) {
		return commitAIAnalysis{Message: explicitMessage}, nil
	}

	agent, err := buildAgent(opts)
	if err != nil {
		return commitAIAnalysis{}, err
	}
	return generateCommitAnalysisWithAgent(ctx, diff, explicitMessage, opts.CompatMode, agent)
}

func generateCommitAnalysisWithAgent(ctx context.Context, diff, explicitMessage, compatMode string, agent clickyai.Agent) (commitAIAnalysis, error) {
	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	message := explicitMessage
	if message == "" {
		analyzed, err := analyzeCommitMessageWithAIFunc(ctx, analysis, agent, git.AnalyzeOptions{})
		if err != nil {
			return commitAIAnalysis{}, err
		}
		out := models.AIAnalysisOutput{
			Type:    analyzed.Commit.CommitType,
			Scope:   analyzed.Commit.Scope,
			Subject: analyzed.Commit.Subject,
			Body:    analyzed.Commit.Body,
		}
		message = strings.TrimSpace(out.String())
		analysis = analyzed
	}

	result := commitAIAnalysis{
		Message: message,
	}
	if !shouldRunCompatibilityAnalysis(compatMode) {
		return result, nil
	}

	analyzed, err := analyzeCompatibilityPromptsWithAIFunc(ctx, analysis, agent, git.AnalyzeOptions{})
	if err != nil {
		result.CompatibilityIssues = []string{formatCompatibilityAnalysisFailure(err)}
		return result, nil
	}

	return commitAIAnalysis{
		Message:              result.Message,
		FunctionalityRemoved: analyzed.FunctionalityRemoved,
		CompatibilityIssues:  analyzed.CompatibilityIssues,
	}, nil
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
		return nil, fmt.Errorf("%w: %w (set ANTHROPIC_API_KEY / CLAUDE_API_KEY / OPENAI_API_KEY / GEMINI_API_KEY)", ErrLLMUnavailable, err)
	}
	return agent, nil
}

func printDryRunPreview(result *Result) {
	if result == nil || len(result.Commits) == 0 {
		return
	}
	fmt.Fprintln(dryRunOutput, result.Pretty().ANSI())
}

// Pretty renders the commit result in a `git log`-style colorized form:
// one header line per commit (short hash or dry-run ref + conventional
// subject) followed by indented body lines. The default reflection-based
// struct printer is intentionally bypassed.
func (r *Result) Pretty() api.Text {
	if r == nil || len(r.Commits) == 0 {
		return clicky.Text("")
	}

	t := clicky.Text("")
	if r.DryRun {
		t = t.Append("DRY RUN", "font-bold text-yellow-600").
			Append(" ", "").
			Append(fmt.Sprintf("would create %d commit(s)", len(r.Commits)), "text-muted").
			NewLine()
	}

	total := len(r.Commits)
	for i, commit := range r.Commits {
		if i > 0 {
			t = t.NewLine()
		}
		t = t.Add(commit.prettyAt(i, total, r.DryRun))
	}
	return t
}

func (c CommitResult) Pretty() api.Text {
	return c.prettyAt(0, 1, c.Hash == "")
}

func (c CommitResult) prettyAt(index, total int, dryRun bool) api.Text {
	parsed := git.NewCommit(c.Message)

	ref := shortHash(c.Hash)
	if ref == "" {
		ref = fmt.Sprintf("dry-run/%d", index+1)
		if dryRun && total > 1 {
			ref = fmt.Sprintf("%s of %d", ref, total)
		}
	}

	t := clicky.Text(ref, "text-yellow-600").Space().Add(parsed.PrettySubject()).NewLine()

	if parsed.Body != "" {
		for _, line := range strings.Split(parsed.Body, "\n") {
			t = t.Append("    ", "").Append(line).NewLine()
		}
	}
	t = appendCompatibilityPreview(t, c.FunctionalityRemoved, c.CompatibilityIssues)
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
