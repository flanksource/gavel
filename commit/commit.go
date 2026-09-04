package commit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/flanksource/captain/pkg/aiflags"
	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/verify"
)

var (
	ErrNothingStaged            = errors.New("nothing staged to commit")
	ErrSessionNoFiles           = errors.New("session edited no stageable files")
	ErrNothingToPush            = errors.New("nothing to commit and no local commits ahead of upstream")
	ErrLLMUnavailable           = errors.New("LLM agent unavailable")
	ErrCommitAllWithMessage     = errors.New("--commit-all does not support --message")
	ErrInteractiveWithCommitAll = errors.New("--interactive cannot be combined with --commit-all")
	ErrInteractiveWithMessage   = errors.New("--interactive cannot be combined with --message")
	ErrInteractiveNonTTY        = errors.New("--interactive requires an interactive terminal")
	ErrInteractiveCancelled     = errors.New("commit cancelled: interactive selection aborted")
	ErrInteractiveEmpty         = errors.New("commit cancelled: no files selected in interactive prompt")
	ErrFilesWithInteractive     = errors.New("file arguments cannot be combined with --interactive")
	ErrFilesWithSince           = errors.New("file arguments cannot be combined with --since")
	ErrFilesWithFixup           = errors.New("file arguments cannot be combined with --fixup")
	ErrFilesOutsideRepo         = errors.New("file argument is outside the repository")

	newAgentFunc                             = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) { return clickyai.NewAgent(cfg) }
	analyzeCommitMessageWithAIFunc           = git.AnalyzeWithAI
	dryRunOutput                   io.Writer = os.Stdout
)

const defaultMaxCommits = 7

const (
	StageStaged   = "staged"
	StageUnstaged = "unstaged"
	StageAll      = "all"
	// StageSession resolves the agent session id from the environment
	// (GAVEL_SESSION_ID, otherwise Captain's provider markers) and stages only
	// the files that session edited; with no session id in the environment it
	// falls back to committing the already-staged set (StageStaged).
	StageSession = "session"

	testEnvVar  = "GAVEL_COMMIT_TEST"
	stubMessage = "chore: fixture stub"
)

type Options struct {
	WorkDir string
	Stage   string
	// Files are explicit git-root-relative pathspecs from `gavel commit
	// <files>`. When non-empty, Run() resets the index and stages exactly these
	// paths (see stageExplicitFiles), so --stage is ignored. Rejected together
	// with --interactive / --since / --fixup (see validateFilesOptions).
	Files []string
	// CommitAll stages all changes and asks the LLM to split them into logical
	// commit groups (plus a separate chore commit for lock/generated files).
	// Set implicitly when MaxCommits is provided.
	CommitAll   bool
	Interactive bool
	// Batch makes each interactive selection an authoritative commit boundary.
	// AI message generation and commits start only after selection completes.
	Batch   bool
	Summary bool
	// MaxCommits caps how many logical commits grouping produces, excluding the
	// trailing chore commit for lock/generated files. It is rendered into the
	// grouping prompt's output schema as maxItems on the groups array and
	// enforced by captain's schemaStrictness=retry policy (bounded fix-up
	// re-asks, then a hard error). Defaults to defaultMaxCommits.
	MaxCommits int
	DryRun     bool
	Force      bool
	NoCache    bool
	Push       bool
	// AutoMerge, with Push, enables GitHub auto-merge on a newly opened PR so
	// it merges once required checks pass. Only applies to PRs this run opens.
	AutoMerge bool
	// MergeType is the merge method used when AutoMerge is set: rebase|squash|merge.
	MergeType string
	// Flags is captain's model-selection surface (--model/--mode/--backend/
	// --effort/--fallback/--temperature/--no-cache). It replaces a bare Model
	// string: a string cannot carry the backend or effort a compact selector
	// names, so those were dropped on the way to the provider.
	Flags aiflags.ModelFlags
	// GroupModel overrides the LLM for AI grouping alone (CLI --group-model),
	// on top of Flags.
	GroupModel    string
	Message       string
	PrecommitMode string
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
	// AI is the base spec (model/budget/effort defaults) every commit AI op
	// inherits. messageModel/groupModel fall back to AI.Model.Name when the
	// operation spec pins no model. Populated from GavelConfig.AI.
	AI captainapi.Spec
	// PR carries the pr: config (content spec, base branch, draft) for the
	// push/PR-open flow. Populated from GavelConfig.PR.
	PR verify.PRConfig

	// lintGates is the resolved on/off state. Populated by Run() before
	// dispatching into runSingleCommit / runCommitAll so the gate runs with
	// stable inputs even when flags are mis-typed. Unexported because
	// callers use LintFlag/LintSecretsFlag.
	lintGates LintGates
}

type CommitResult struct {
	Label   string   `json:"label,omitempty"`
	Message string   `json:"message"`
	Hash    string   `json:"hash,omitempty"`
	Files   []string `json:"files,omitempty"`
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
	Message string
}

func Run(ctx context.Context, opts Options) (*Result, error) {
	if opts.Stage == "" {
		opts.Stage = StageSession
	}
	if opts.WorkDir == "" {
		return nil, errors.New("commit.Run: WorkDir is required")
	}
	// Normalize WorkDir to the repository top-level. Everything downstream —
	// `git diff --cached` (repo-root-relative paths), the `git reset`/`git add`
	// pathspecs, and where .gitignore / .gavel.yaml get written — assumes
	// WorkDir IS the git root. Running from a subdirectory otherwise silently
	// misfires (unstage no-ops, stray .gitignore written in the subdir).
	gitRoot, err := resolveGitRoot(opts.WorkDir)
	if err != nil {
		return nil, err
	}
	opts.WorkDir = gitRoot
	// --max-commits only has meaning for the grouping flow, so setting it implies
	// -A. Applied before the mutual-exclusion guards so a conflicting combination
	// (e.g. --max-commits with --fixup/--interactive) is rejected, not ignored.
	if opts.MaxCommits != 0 {
		opts.CommitAll = true
	}
	if err := validateFilesOptions(opts); err != nil {
		return nil, err
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

	gates, err := resolveLintGates(opts.LintFlag, opts.LintSecretsFlag, opts.Config.Lint)
	if err != nil {
		return nil, err
	}
	opts.lintGates = gates

	// Explicit file arguments (`gavel commit <files>`) define the commit set:
	// reset the index and stage exactly those paths, then treat the pre-staged
	// set as the source for every downstream flow. stageFiles(StageStaged) and
	// stageCommitAllSource both no-op on a non-empty index, so no per-flow
	// change is needed.
	if len(opts.Files) > 0 {
		if err := stageExplicitFiles(opts.WorkDir, opts.Files); err != nil {
			return nil, err
		}
		opts.Stage = StageStaged
	}

	var (
		result *Result
	)
	switch {
	case opts.Since != "":
		result, err = runIssueIdSquash(ctx, opts)
	case opts.Fixup != "":
		result, err = runFixup(ctx, opts)
	case opts.CommitAll:
		if opts.Message != "" {
			return nil, ErrCommitAllWithMessage
		}
		if opts.MaxCommits == 0 {
			opts.MaxCommits = defaultMaxCommits
		}
		result, err = runCommitAll(ctx, opts)
	case opts.Batch:
		result, err = runInteractiveBatch(ctx, opts)
	case opts.Interactive && !opts.DryRun:
		result, err = runInteractiveLoop(ctx, opts)
	default:
		result, err = runSingleCommit(ctx, opts)
	}
	if err != nil {
		// With --push, "nothing staged" is not fatal: fall through and
		// push existing local commits / open a PR for what HEAD already
		// has ahead of upstream.
		if opts.Push && errors.Is(err, ErrNothingStaged) && (result == nil || len(result.Commits) == 0) {
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
		if isTokenLimitError(err) && !opts.Batch {
			logger.Infof("staged diff exceeds model context window, splitting into logical commits")
			return commitByGrouping(ctx, opts, source, result)
		}
		return result, fmt.Errorf("generate commit analysis: %w", err)
	}
	analysis.Message = applyCommitMetadata(opts, analysis.Message)
	result.Message = analysis.Message
	result.Commits = []CommitResult{{
		Message: analysis.Message,
		Files:   source.Files,
	}}

	if opts.DryRun {
		printDryRunPreview(result)
		return result, nil
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
