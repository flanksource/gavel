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
	ErrNothingStaged            = errors.New("nothing staged to commit")
	ErrNothingToPush            = errors.New("nothing to commit and no local commits ahead of upstream")
	ErrLLMUnavailable           = errors.New("LLM agent unavailable")
	ErrInvalidStage             = errors.New("invalid --stage value")
	ErrCommitAllWithMessage     = errors.New("--commit-all does not support --message")
	ErrInteractiveWithCommitAll = errors.New("--interactive cannot be combined with --commit-all")
	ErrInteractiveWithMessage   = errors.New("--interactive cannot be combined with --message")
	ErrInteractiveNonTTY        = errors.New("--interactive requires an interactive terminal")
	ErrInteractiveCancelled     = errors.New("commit cancelled: interactive selection aborted")
	ErrInteractiveEmpty         = errors.New("commit cancelled: no files selected in interactive prompt")

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
	Interactive   bool
	Summary       bool
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
	// LintFlag and LintSecretsFlag are the raw string forms of --lint and
	// --lint-secrets. Empty = flag not provided; "true"/"false" override
	// .gavel.yaml commit.lint.{enabled,secrets}. Strings (not *bool) so the
	// clicky flag binding stays a plain string flag the user can set to
	// "true" or "false".
	LintFlag        string
	LintSecretsFlag string
	// Fixup, when non-empty, switches Run() to runFixup. The literal
	// FixupAuto value triggers per-file routing by last-touching commit on
	// base..HEAD; any other value is treated as an explicit target hash.
	Fixup string
	// Autosquash controls whether `git rebase -i --autosquash` runs after
	// fixup commits are created. Defaults to true at the CLI; tests / direct
	// callers must opt in explicitly.
	Autosquash bool
	Config     verify.CommitConfig

	// lintGates is the resolved on/off state. Populated by Run() before
	// dispatching into runSingleCommit / runCommitAll so the gate runs with
	// stable inputs even when flags are mis-typed. Unexported because
	// callers use LintFlag/LintSecretsFlag.
	lintGates LintGates
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
	Message string `json:"message"`
	Hash    string `json:"hash,omitempty"`
	DryRun  bool   `json:"dry_run,omitempty"`
	// PushOnly is set when --push was used with nothing staged — the
	// commits in this Result already exist in HEAD; we're only pushing
	// them. Pretty() uses this to switch the dry-run header from
	// "would create" to "would push existing".
	PushOnly bool           `json:"push_only,omitempty"`
	Staged   []string       `json:"staged,omitempty"`
	Hooks    []HookResult   `json:"hooks,omitempty"`
	Commits  []CommitResult `json:"commits,omitempty"`
	// Lint is set when the pre-commit lint gate ran. Non-nil whether it
	// passed (Violations==0) or blocked (Violations>0); the CLI uses it to
	// render findings and run the triage flow.
	Lint *LintGateResult `json:"-"`
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
	if err := validateInteractiveOptions(opts); err != nil {
		return nil, err
	}
	if err := validateFixupOptions(opts); err != nil {
		return nil, err
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

	gates, err := resolveLintGates(opts.LintFlag, opts.LintSecretsFlag, opts.Config.Lint)
	if err != nil {
		return nil, err
	}
	opts.lintGates = gates

	var (
		result *Result
	)
	switch {
	case opts.Fixup != "":
		result, err = runFixup(ctx, opts)
	case opts.CommitAll:
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
	case opts.Interactive && !opts.DryRun:
		result, err = runInteractiveLoop(ctx, opts)
	default:
		result, err = runSingleCommit(ctx, opts)
	}
	if err != nil {
		// With --push, "nothing staged" is not fatal: fall through and
		// push existing local commits / open a PR for what HEAD already
		// has ahead of upstream.
		if opts.Push && errors.Is(err, ErrNothingStaged) {
			result = &Result{DryRun: opts.DryRun, PushOnly: true}
			err = nil
		} else {
			return result, err
		}
	}
	if opts.Push {
		if perr := pushAfterCommit(ctx, opts, result); perr != nil {
			return result, perr
		}
	}
	return result, nil
}

func runSingleCommit(ctx context.Context, opts Options) (*Result, error) {
	if opts.Interactive {
		if _, err := runInteractiveStaging(ctx, opts); err != nil {
			return nil, err
		}
	} else {
		if err := stageFiles(opts.WorkDir, opts.Stage); err != nil {
			return nil, fmt.Errorf("stage files (%s): %w", opts.Stage, err)
		}
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

	lintRes, lintErr := applyLintGate(ctx, opts.WorkDir, source.Files, opts.lintGates)
	result.Lint = lintRes
	if lintErr != nil {
		return result, lintErr
	}

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

// runInteractiveLoop runs runSingleCommit repeatedly so the user can keep
// picking subsets of changed files into separate commits without re-invoking
// `gavel commit -i`. The loop ends when:
//   - no candidate files remain (clean exit, returns the accumulated result),
//   - the picker is cancelled with esc/ctrl+c (clean exit),
//   - the picker is confirmed with no files selected (clean exit),
//   - any other error occurs (returned to the caller).
//
// The first iteration still surfaces ErrNothingStaged so `gavel commit -i`
// with no changed files behaves like the non-loop form. Subsequent
// iterations treat "nothing left" as success.
func runInteractiveLoop(ctx context.Context, opts Options) (*Result, error) {
	aggregate := &Result{}
	iteration := 0
	for {
		iteration++
		single, err := runSingleCommit(ctx, opts)
		if err != nil {
			if isInteractiveLoopExit(err) && iteration > 1 {
				return aggregate, nil
			}
			return mergeResults(aggregate, single), err
		}
		aggregate = mergeResults(aggregate, single)
		fmt.Fprintf(interactiveStdout, "\n— commit %d created; checking for more changes —\n\n", iteration)
	}
}

// isInteractiveLoopExit reports whether err is one of the "user is done"
// sentinels that should end the loop without surfacing as a failure.
func isInteractiveLoopExit(err error) bool {
	return errors.Is(err, ErrNothingStaged) ||
		errors.Is(err, ErrInteractiveCancelled) ||
		errors.Is(err, ErrInteractiveEmpty)
}

// mergeResults folds a per-iteration single-commit Result into the
// loop-wide aggregate. Hooks and Staged are tracked per-iteration only on
// the most recent single result; Commits and Hash carry the latest commit so
// pushAfterCommit can push HEAD as usual.
func mergeResults(agg, single *Result) *Result {
	if single == nil {
		return agg
	}
	if agg == nil {
		agg = &Result{}
	}
	agg.DryRun = single.DryRun
	agg.Message = single.Message
	agg.Hash = single.Hash
	agg.Staged = single.Staged
	agg.Hooks = single.Hooks
	agg.Commits = append(agg.Commits, single.Commits...)
	if single.Lint != nil {
		agg.Lint = single.Lint
	}
	return agg
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

	lintRes, lintErr := applyLintGate(ctx, opts.WorkDir, source.Files, opts.lintGates)
	result.Lint = lintRes
	if lintErr != nil {
		return result, lintErr
	}

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

	agent, err := BuildAgent(opts)
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

func BuildAgent(opts Options) (clickyai.Agent, error) {
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
//
// Live runs (non-dry-run) return empty text: per-commit "Committed <hash>"
// lines are already logged by runSingleCommit/runCommitAll, and `--push`
// prints PR title/body separately. The trailing block this used to emit
// only restated information the user just saw.
func (r *Result) Pretty() api.Text {
	if r == nil || len(r.Commits) == 0 || !r.DryRun {
		return clicky.Text("")
	}

	t := clicky.Text("")
	summary := fmt.Sprintf("would create %d commit(s)", len(r.Commits))
	if r.PushOnly {
		summary = fmt.Sprintf("would push %d existing commit(s)", len(r.Commits))
	}
	t = t.Append("DRY RUN", "font-bold text-yellow-600").
		Append(" ", "").
		Append(summary, "text-muted").
		NewLine()

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
