package commit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/ai/history"
	"github.com/flanksource/clicky"
	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

var (
	ErrNothingStaged            = errors.New("nothing staged to commit")
	ErrSessionNoFiles           = errors.New("session edited no stageable files")
	ErrNothingToPush            = errors.New("nothing to commit and no local commits ahead of upstream")
	ErrLLMUnavailable           = errors.New("LLM agent unavailable")
	ErrCommitAllWithMessage     = errors.New("--commit-all does not support --message")
	ErrAIGroupWithMessage       = errors.New("--ai-group does not support --message")
	ErrAIGroupWithInteractive   = errors.New("--ai-group cannot be combined with --interactive")
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
	defaultMaxFiles   = 7
	defaultMaxLines   = 500
	defaultMaxCommits = 7
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
	// AIGroup asks the LLM to split the change set into logical commit groups
	// (plus a separate chore commit for lock/generated files) instead of
	// grouping by directory. Combine with CommitAll to stage all changes first;
	// otherwise it groups only the staged set.
	AIGroup     bool
	Interactive bool
	Summary     bool
	MaxFiles    int
	MaxLines    int
	// MaxCommits caps how many commits AI grouping (--ai-group) produces,
	// excluding the trailing chore commit for lock/generated files. It is fed to
	// the grouping prompt and re-enforced via a consolidation feedback prompt when
	// the LLM exceeds it. Defaults to defaultMaxCommits.
	MaxCommits int
	// GroupByScope makes AI grouping treat repomap scope as the primary commit
	// boundary. Default (false) groups by logical change with scope as a hint.
	GroupByScope bool
	DryRun       bool
	Force        bool
	NoCache      bool
	Push         bool
	// AutoMerge, with Push, enables GitHub auto-merge on a newly opened PR so
	// it merges once required checks pass. Only applies to PRs this run opens.
	AutoMerge bool
	// MergeType is the merge method used when AutoMerge is set: rebase|squash|merge.
	MergeType string
	// Model overrides the LLM for commit-message and PR-content generation
	// (CLI --model). GroupModel overrides the LLM for AI grouping (CLI
	// --group-model); both fall back to .gavel.yaml commit.{model,groupModel}.
	Model         string
	GroupModel    string
	Message       string
	PrecommitMode string
	CompatMode    string
	// AssumeYes auto-answers precommit triage prompts with their default
	// action: linked-dep violations auto-unstage. Set by `gavel commit -y`.
	AssumeYes bool
	// LintFlag and LintSecretsFlag are the raw string forms of --lint and
	// --lint-secrets. Empty = flag not provided; "true"/"false" override
	// .gavel.yaml commit.lint.{enabled,secrets}. Strings (not *bool) so the
	// clicky flag binding stays a plain string flag the user can set to
	// "true" or "false".
	LintFlag        string
	LintSecretsFlag string
	// TidyFlag is the raw string form of --tidy. Empty = flag not provided;
	// "true"/"false" override .gavel.yaml commit.tidy.enabled. String (not
	// *bool) so the clicky flag binding stays a plain string flag the user
	// can set to "true" or "false".
	TidyFlag string
	// Fixup, when non-empty, switches Run() to runFixup. The literal
	// FixupAuto value triggers per-file routing by last-touching commit on
	// base..HEAD; any other value is treated as an explicit target hash.
	Fixup string
	// Since, when non-empty, switches Run() to runIssueIdSquash: review
	// <Since>..HEAD and merge commits sharing a Gavel-Issue-Id trailer into a
	// single commit. History-only — staged files are ignored.
	Since string
	// Autosquash controls whether `git rebase -i --autosquash` runs after
	// fixup commits are created. Defaults to true at the CLI; tests / direct
	// callers must opt in explicitly.
	Autosquash bool
	// AddMetadata appends git trailers identifying the gavel todo issue and
	// agent session to each generated commit message. Defaults to true at the
	// CLI; direct callers must opt in. See applyCommitMetadata.
	AddMetadata bool
	// IssueID and SessionID are the gavel todo issue id and agent session id to
	// stamp when AddMetadata is set. Populated in-process by RunAfterAgent; when
	// empty applyCommitMetadata falls back to the GAVEL_ISSUE_ID /
	// GAVEL_SESSION_ID env vars.
	IssueID   string
	SessionID string
	Config    verify.CommitConfig

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
	if err := validateSinceOptions(opts); err != nil {
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
	case opts.Since != "":
		result, err = runIssueIdSquash(ctx, opts)
	case opts.Fixup != "":
		result, err = runFixup(ctx, opts)
	case opts.AIGroup:
		if opts.Message != "" {
			return nil, ErrAIGroupWithMessage
		}
		if opts.Interactive {
			return nil, ErrAIGroupWithInteractive
		}
		if opts.MaxFiles == 0 {
			opts.MaxFiles = defaultMaxFiles
		}
		if opts.MaxLines == 0 {
			opts.MaxLines = defaultMaxLines
		}
		if opts.MaxCommits == 0 {
			opts.MaxCommits = defaultMaxCommits
		}
		result, err = runCommitAIGroup(ctx, opts)
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
		if err := stageFiles(opts.WorkDir, opts.Stage, opts.Config); err != nil {
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

	source, err = applyPrecommitChecks(ctx, opts, source)
	if err != nil {
		return nil, err
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
		if isTokenLimitError(err) {
			logger.Infof("staged diff exceeds model context window, splitting commit by directory")
			return commitByDirectory(ctx, opts, source, result)
		}
		return result, fmt.Errorf("generate commit analysis: %w", err)
	}
	analysis.Message = applyCommitMetadata(opts, analysis.Message)
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
	restoreLocalReplaces(opts.WorkDir, source.PendingRestores)
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
	if err := stageCommitAllSource(opts.WorkDir, opts.Config); err != nil {
		return nil, err
	}

	source, result, err := prepareMultiCommit(ctx, opts)
	if err != nil {
		return result, err
	}

	return commitByDirectory(ctx, opts, source, result)
}

// runCommitAIGroup stages the change set (all changes when --commit-all is also
// set, otherwise just the staged set), asks the LLM to split it into logical
// commit groups plus a chore group for lock/generated files, and creates one
// commit per group.
func runCommitAIGroup(ctx context.Context, opts Options) (*Result, error) {
	if opts.CommitAll {
		if err := stageCommitAllSource(opts.WorkDir, opts.Config); err != nil {
			return nil, err
		}
	} else {
		if err := stageFiles(opts.WorkDir, opts.Stage, opts.Config); err != nil {
			return nil, fmt.Errorf("stage files (%s): %w", opts.Stage, err)
		}
	}

	source, result, err := prepareMultiCommit(ctx, opts)
	if err != nil {
		return result, err
	}

	groups, err := groupChangesByAIFunc(ctx, opts, source)
	if err != nil {
		return result, fmt.Errorf("ai grouping: %w", err)
	}

	return commitGroups(ctx, opts, source, result, groups)
}

// prepareMultiCommit runs the shared staging-completion pipeline for the
// multi-commit flows (runCommitAll, runCommitAIGroup): it reads the staged
// source, applies the precommit gates, runs hooks, re-reads the staged source,
// and applies the lint gate. Callers are responsible for staging beforehand
// (stageCommitAllSource for --commit-all, stageFiles for --ai-group). It
// returns the final staged source and a Result pre-populated with Staged/Hooks/
// Lint so error returns carry partial state.
func prepareMultiCommit(ctx context.Context, opts Options) (stagedSource, *Result, error) {
	source, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return source, nil, err
	}
	if len(source.Files) == 0 {
		return source, nil, ErrNothingStaged
	}

	source, err = applyPrecommitChecks(ctx, opts, source)
	if err != nil {
		return source, nil, err
	}

	result := &Result{Staged: source.Files, DryRun: opts.DryRun}

	if !opts.Force {
		hookResults, hookErr := RunHooks(opts.WorkDir, opts.Config.Hooks, source.Files)
		result.Hooks = hookResults
		if hookErr != nil {
			return source, result, hookErr
		}
	} else if len(opts.Config.Hooks) > 0 {
		logger.Infof("Skipping %d commit hook(s) due to --force", len(opts.Config.Hooks))
	}

	source, err = readStagedSource(opts.WorkDir)
	if err != nil {
		return source, result, err
	}
	if len(source.Files) == 0 {
		return source, result, ErrNothingStaged
	}
	result.Staged = source.Files

	lintRes, lintErr := applyLintGate(ctx, opts.WorkDir, source.Files, opts.lintGates)
	result.Lint = lintRes
	if lintErr != nil {
		return source, result, lintErr
	}

	return source, result, nil
}

// applyPrecommitChecks runs the gitignore, file-size, linked-deps and
// go-mod-tidy gates in order over the staged source, returning ErrNothingStaged
// as soon as any gate empties the staged set. Shared by runSingleCommit and
// prepareMultiCommit.
func applyPrecommitChecks(ctx context.Context, opts Options, source stagedSource) (stagedSource, error) {
	checks := []func(context.Context, Options, stagedSource) (stagedSource, error){
		applyGitIgnoreCheck,
		applyFileSizeCheck,
		applyLinkedDepsCheck,
		applyGoModTidy,
	}
	for _, check := range checks {
		var err error
		source, err = check(ctx, opts, source)
		if err != nil {
			return source, err
		}
		if len(source.Files) == 0 {
			return source, ErrNothingStaged
		}
	}
	return source, nil
}

// commitByDirectory groups the staged changes by directory and creates one
// commit per group. It backs `--commit-all` and is also the fallback when a
// single-commit AI analysis overflows the model context window.
func commitByDirectory(ctx context.Context, opts Options, source stagedSource, result *Result) (*Result, error) {
	if opts.MaxFiles == 0 {
		opts.MaxFiles = defaultMaxFiles
	}
	if opts.MaxLines == 0 {
		opts.MaxLines = defaultMaxLines
	}

	groups := groupChangesByDir(source.Changes, opts.MaxFiles, opts.MaxLines)
	return commitGroups(ctx, opts, source, result, groups)
}

// commitGroups generates a message for each commit group and creates one commit
// per group. A group carrying a preset Message (e.g. the --ai-group chore group
// for lock/generated files) skips the LLM and uses it verbatim. It backs both
// the directory grouper and the AI grouper.
func commitGroups(ctx context.Context, opts Options, source stagedSource, result *Result, groups []commitGroup) (*Result, error) {
	if len(groups) == 0 {
		return result, ErrNothingStaged
	}

	result.Commits = make([]CommitResult, 0, len(groups))
	for _, group := range groups {
		if group.Message != "" {
			result.Commits = append(result.Commits, CommitResult{
				Label:   group.Label,
				Message: applyCommitMetadata(opts, group.Message),
				Files:   group.Files(),
			})
			continue
		}
		analysis, msgErr := generateCommitAnalysis(ctx, opts, group.diff())
		if msgErr != nil {
			return result, fmt.Errorf("generate commit analysis for %s: %w", group.labelOrDefault(), msgErr)
		}
		result.Commits = append(result.Commits, CommitResult{
			Label:                group.Label,
			Message:              applyCommitMetadata(opts, analysis.Message),
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

	restoreLocalReplaces(opts.WorkDir, source.PendingRestores)

	return result, nil
}

// stageFiles stages changes for the given mode using git's own ignore-aware
// semantics, so an ignored path can never abort the whole `git add`:
//   - tracked modifications/deletions go in via `git add -u`, then
//     unstageGitIgnored removes any that match the repo's .gitignore (a
//     force-tracked bundle like testrunner/ui/dist/testui.js is left out of the
//     commit but stays tracked); files staged manually before the call are
//     preserved, and a !-negation in .gitignore re-includes a path;
//   - StageAll additionally adds untracked files that are matched by neither
//     .gitignore (handled by git) nor the repo's .gavel.yaml commit.gitignore.
//
// StageStaged is left untouched: an explicit `git add` is an intentional
// override of .gitignore for that commit.
//
// Any other mode value is a Claude session id: stageSessionFiles stages exactly
// the files that session's Edit/Write tools touched (see --stage=<session-id>).
func stageFiles(workDir, mode string, cfg verify.CommitConfig) error {
	switch mode {
	case StageStaged:
		return nil
	case StageUnstaged, StageAll:
		preStaged, err := stagedFiles(workDir)
		if err != nil {
			return fmt.Errorf("list pre-staged files: %w", err)
		}
		if err := gitAddUpdate(workDir); err != nil {
			return err
		}
		if mode == StageAll {
			if err := addUntracked(workDir, cfg); err != nil {
				return err
			}
		}
		return unstageGitIgnored(workDir, preStaged)
	default:
		return stageSessionFiles(workDir, mode, cfg)
	}
}

// stageSessionFiles stages exactly the files an agent wrote to during the Claude
// session identified by sessionID — the file_path of every Edit/Write/MultiEdit/
// NotebookEdit tool call — filtered by .gitignore and the repo's .gavel.yaml
// commit.gitignore. It scopes a commit to what the agent changed rather than the
// whole working tree, and backs both `gavel commit --stage=<session-id>` and the
// todo runner's auto-commit. Files edited outside workDir, no longer present, or
// matching an ignore rule are skipped (each logged).
func stageSessionFiles(workDir, sessionID string, cfg verify.CommitConfig) error {
	sessionFile, err := history.FindSessionFile(sessionID)
	if err != nil {
		return fmt.Errorf("stage session %q: no Claude session log found: %w", sessionID, err)
	}
	logger.Infof("commit: staging files from claude session %s (%s)", sessionID, sessionFile)

	modified, err := history.SessionModifiedFiles(sessionFile)
	if err != nil {
		return fmt.Errorf("stage session %q: read session log %s: %w", sessionID, sessionFile, err)
	}
	logger.Infof("commit: session %s edited %d file(s)", sessionID, len(modified))
	for _, p := range modified {
		logger.V(1).Infof("commit:   edited %s", p)
	}

	absWork, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve work dir %q: %w", workDir, err)
	}

	candidates := make([]string, 0, len(modified))
	for _, p := range modified {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(absWork, abs)
		}
		rel, err := filepath.Rel(absWork, abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			logger.Infof("commit: skipping %s (edited outside %s)", p, absWork)
			continue
		}
		if _, statErr := os.Stat(abs); statErr != nil {
			logger.Infof("commit: skipping %s (no longer present)", rel)
			continue
		}
		candidates = append(candidates, rel)
	}

	keep, err := filterIgnoredPaths(absWork, candidates, cfg)
	if err != nil {
		return err
	}
	logger.Infof("commit: staging %d of %d session file(s)", len(keep), len(modified))
	if len(keep) == 0 {
		return fmt.Errorf("stage session %q (%d edited, all skipped): %w", sessionID, len(modified), ErrSessionNoFiles)
	}
	return addFiles(workDir, keep)
}

// filterIgnoredPaths drops paths the repo's .gitignore (naming one to `git add`
// would error) or .gavel.yaml commit.gitignore excludes, logging each skip.
func filterIgnoredPaths(workDir string, candidates []string, cfg verify.CommitConfig) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	ignored := make(map[string]struct{})

	absToRel := make(map[string]string, len(candidates))
	abs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		p := filepath.Join(workDir, c)
		absToRel[p] = c
		abs = append(abs, p)
	}
	_, gitIgnored := utils.PartitionGitIgnored(abs, workDir)
	for _, p := range gitIgnored {
		rel := absToRel[p]
		logger.Infof("commit: skipping %s (matches .gitignore)", rel)
		ignored[rel] = struct{}{}
	}

	violations, err := EvaluateGitIgnoreMatches(candidates, cfg.GitIgnore, cfg.Allow)
	if err != nil {
		return nil, fmt.Errorf("evaluate .gavel.yaml commit.gitignore: %w", err)
	}
	for _, v := range violations {
		logger.Infof("commit: skipping %s (matches .gavel.yaml commit.gitignore %q)", v.File, v.Pattern)
		ignored[v.File] = struct{}{}
	}

	keep := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if _, skip := ignored[c]; !skip {
			keep = append(keep, c)
		}
	}
	return keep, nil
}

// unstageGitIgnored removes from the index any staged file that matches the
// repo's .gitignore, except files in preStaged (staged by the user before
// gavel ran). It uses `git reset --` so the working tree and tracking are
// untouched — the modification simply won't be in the commit. A !-negation in
// .gitignore keeps a path staged.
func unstageGitIgnored(workDir string, preStaged []string) error {
	staged, err := stagedFiles(workDir)
	if err != nil {
		return fmt.Errorf("list staged files: %w", err)
	}
	if len(staged) == 0 {
		return nil
	}

	absToRel := make(map[string]string, len(staged))
	abs := make([]string, 0, len(staged))
	for _, f := range staged {
		p := filepath.Join(workDir, f)
		absToRel[p] = f
		abs = append(abs, p)
	}

	preStagedSet := make(map[string]struct{}, len(preStaged))
	for _, f := range preStaged {
		preStagedSet[f] = struct{}{}
	}

	_, ignored := utils.PartitionGitIgnored(abs, workDir)
	toReset := make([]string, 0, len(ignored))
	for _, p := range ignored {
		rel := absToRel[p]
		if _, kept := preStagedSet[rel]; kept {
			continue
		}
		logger.Infof("commit: skipping %s (matches .gitignore)", rel)
		toReset = append(toReset, rel)
	}
	return resetFiles(workDir, toReset)
}

// addUntracked stages untracked files that git does not ignore, minus the
// repo's .gavel.yaml commit.gitignore rules (commit.allow re-includes) and minus
// embedded git repositories, which `git add` refuses. Every skip is logged so a
// dropped file is never silent.
func addUntracked(workDir string, cfg verify.CommitConfig) error {
	untracked, err := untrackedFiles(workDir)
	if err != nil {
		return fmt.Errorf("list untracked files: %w", err)
	}
	if len(untracked) == 0 {
		return nil
	}

	candidates := make([]string, 0, len(untracked))
	for _, f := range untracked {
		if strings.HasSuffix(f, "/") {
			logger.Infof("commit: skipping embedded repo or directory %s", f)
			continue
		}
		candidates = append(candidates, f)
	}

	violations, err := EvaluateGitIgnoreMatches(candidates, cfg.GitIgnore, cfg.Allow)
	if err != nil {
		return fmt.Errorf("evaluate .gavel.yaml commit.gitignore: %w", err)
	}
	ignored := make(map[string]struct{}, len(violations))
	for _, v := range violations {
		logger.Infof("commit: skipping %s (matches .gavel.yaml commit.gitignore %q)", v.File, v.Pattern)
		ignored[v.File] = struct{}{}
	}

	keep := make([]string, 0, len(candidates))
	for _, f := range candidates {
		if _, skip := ignored[f]; !skip {
			keep = append(keep, f)
		}
	}
	return addFiles(workDir, keep)
}

func stageCommitAllSource(workDir string, cfg verify.CommitConfig) error {
	files, err := stagedFiles(workDir)
	if err != nil {
		return fmt.Errorf("list staged files: %w", err)
	}
	if len(files) > 0 {
		return nil
	}
	if err := stageFiles(workDir, StageAll, cfg); err != nil {
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

	agent, err := BuildAgent(opts, opts.messageModel())
	if err != nil {
		return commitAIAnalysis{}, err
	}
	return generateCommitAnalysisWithAgent(ctx, diff, explicitMessage, opts.CompatMode, agent)
}

func generateCommitAnalysisWithAgent(ctx context.Context, diff, explicitMessage, compatMode string, agent clickyai.Agent) (commitAIAnalysis, error) {
	analysis := models.CommitAnalysis{Commit: models.Commit{Patch: diff}}
	message := explicitMessage
	if message == "" {
		maxBodyLines := maxBodyLinesForDiff(countDiffLines(diff))
		analyzed, err := analyzeCommitMessageWithAIFunc(ctx, analysis, agent, git.AnalyzeOptions{MaxBodyLines: maxBodyLines})
		if err != nil {
			return commitAIAnalysis{}, err
		}
		out := models.AIAnalysisOutput{
			Type:    analyzed.CommitType,
			Scope:   analyzed.Scope,
			Subject: analyzed.Subject,
			Body:    analyzed.Body,
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

// countDiffLines counts changed content lines in a unified diff: lines starting
// with '+' or '-', excluding the '+++'/'---' file headers.
func countDiffLines(diff string) int {
	n := 0
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			n++
		}
	}
	return n
}

// maxBodyLinesForDiff scales the commit-message body cap to the diff size:
// trivial diffs get a subject only (0), larger diffs allow a longer body.
func maxBodyLinesForDiff(changedLines int) int {
	switch {
	case changedLines <= 20:
		return 0
	case changedLines <= 100:
		return 3
	case changedLines <= 300:
		return 6
	case changedLines <= 800:
		return 10
	default:
		return 15
	}
}

// defaultGroupModel is the model used for AI commit grouping when neither
// --group-model nor commit.{groupModel,model} is set. Grouping reasons over the
// whole change set, so it defaults to a more capable tier than the haiku-class
// model used for message generation.
const defaultGroupModel = "claude-sonnet-4-5"

// messageModel resolves the LLM for commit-message and PR-content generation:
// the CLI --model override, then .gavel.yaml commit.model, else the clicky
// default (a fast haiku-class model).
func (opts Options) messageModel() string {
	if opts.Model != "" {
		return opts.Model
	}
	return opts.Config.Model
}

// groupModel resolves the LLM for AI commit grouping. Precedence: --group-model,
// then commit.groupModel, then the shared message model (--model / commit.model),
// else defaultGroupModel.
func (opts Options) groupModel() string {
	if opts.GroupModel != "" {
		return opts.GroupModel
	}
	if opts.Config.GroupModel != "" {
		return opts.Config.GroupModel
	}
	if m := opts.messageModel(); m != "" {
		return m
	}
	return defaultGroupModel
}

// BuildAgent constructs an LLM agent for opts using model. An empty model falls
// back to the clicky default. Callers resolve model per task via
// Options.messageModel / Options.groupModel.
func BuildAgent(opts Options, model string) (clickyai.Agent, error) {
	cfg := clickyai.DefaultConfig()
	if model != "" {
		cfg.Model = model
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
