package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

type CommitOptions struct {
	Stage        string `flag:"stage" help:"Which changes to commit: staged|unstaged|all" default:"staged"`
	CommitAll    bool   `flag:"commit-all" short:"A" help:"Split the selected change set into commits grouped by directory"`
	AIGroup      bool   `flag:"ai-group" short:"G" help:"Ask the LLM to split the change set into logical commit groups (and a separate chore commit for lock/generated files) instead of grouping by directory. Combine with -A to first stage all changes."`
	Interactive  bool   `flag:"interactive" short:"i" help:"Open an interactive tree picker over all changed files (staged, unstaged, untracked); selecting confirms which files to commit"`
	Tree         bool   `flag:"tree" short:"t" help:"Alias for --interactive"`
	Summary      bool   `flag:"summary" short:"s" help:"With -i, print a gavel-status-style summary of the candidate files before the picker opens"`
	MaxFiles     int    `flag:"max-files" help:"Max files per commit group before splitting further by subdirectory" default:"7"`
	MaxLines     int    `flag:"max-lines" help:"Max changed lines (adds+dels, excluding new files) per commit group before splitting further by subdirectory" default:"500"`
	MaxCommits   int    `flag:"max-commits" help:"Max number of commits AI grouping (-G) should produce, excluding the chore commit for lock/generated files" default:"7"`
	GroupByScope bool   `flag:"group-by-scope" help:"With -G, group changes by repomap scope as the primary boundary; default groups by logical change with scope as a hint" default:"false"`
	Message      string `flag:"message" short:"m" help:"Explicit commit message (skips only the message-generation LLM call)"`
	Model        string `flag:"model" help:"Override LLM model for commit-message/PR generation from .gavel.yaml commit.model (fast/haiku-class)"`
	GroupModel   string `flag:"group-model" help:"Override LLM model for AI commit grouping (-G) from .gavel.yaml commit.groupModel (capable/sonnet-class); falls back to --model"`
	DryRun       bool   `flag:"dry-run" help:"Print the generated message without committing"`
	Force        bool   `flag:"force" help:"Skip pre-commit hooks"`
	NoCache      bool   `flag:"no-cache" help:"Bypass the LLM response cache at ~/.cache/clicky-ai.db"`
	Push         bool   `flag:"push" short:"p" help:"Push to a matching open PR or open a new PR. Skips the commit step when nothing is staged so existing local commits can be pushed."`
	AutoMerge    bool   `flag:"auto-merge" help:"With -p, when a new PR is opened, enable GitHub auto-merge so it merges once required checks pass."`
	MergeType    string `flag:"merge-type" help:"Merge method for --auto-merge: rebase|squash|merge" default:"rebase"`
	Fixup        string `flag:"fixup" help:"Squash staged files into existing commits. Pass a hash to target one commit, or use bare --fixup to auto-route each file by last-touched commit on origin/main..HEAD."`
	NoAutosquash bool   `flag:"no-autosquash" help:"With --fixup, skip the automatic 'git rebase -i --autosquash' that folds fixup commits into their targets."`
	Since        string `flag:"since" help:"Review <since>..HEAD and merge commits sharing a Gavel-Issue-Id trailer into one commit (history only; ignores staged files). Accepts a ref (origin/main), sha, or ~N / HEAD~N. Prompts before rewriting; -y to skip. Refuses to rewrite commits already on a remote."`
	Precommit    string `flag:"precommit" help:"Behavior for pre-commit gitignore and linked dependency checks: prompt|fail|skip|false"`
	Compat       string `flag:"compat" help:"Behavior for AI compatibility analysis and findings (default: skip): prompt|fail|skip|false"`
	Lint         string `flag:"lint" help:"Run all detected linters over staged files before committing: true|false (default: false; overrides .gavel.yaml commit.lint.enabled)"`
	LintSecrets  string `flag:"lint-secrets" help:"Run the betterleaks/secrets linter over staged files before committing: true|false (default: true; overrides .gavel.yaml commit.lint.secrets)"`
	Tidy         string `flag:"tidy" help:"Run 'go mod tidy' in every Go module before committing and stage any go.mod/go.sum updates: true|false (default: true; overrides .gavel.yaml commit.tidy.enabled). May stage previously-unstaged go.mod/go.sum edits."`
	WorkDir      string `flag:"work-dir" help:"Working directory"`
	Yes          bool   `flag:"yes" short:"y" help:"Assume yes: auto-unstage linked-dep replacements and auto-AI-fix lint findings instead of prompting."`
	AddMetadata  bool   `flag:"add-metadata" default:"true" help:"Append Gavel-Issue-Id / Claude-Session-Id trailers to commit messages, sourced from GAVEL_ISSUE_ID / GAVEL_SESSION_ID (set by 'gavel todos run')."`
}

func (o CommitOptions) Help() string {
	return `Generate a conventional commit message via LLM and run pre-commit hooks.

Reads pre-commit hooks from .gavel.yaml under commit.hooks. Hooks run with
sh -c in the git root and abort the commit on non-zero exit. Pass --force
to skip hooks.

Before hooks run, staged files are checked against commit.gitignore patterns
(typically set in ~/.gavel.yaml). Matches trigger a per-file prompt to
(1) append the matched pattern to the repo .gitignore, (2) append the file's
folder, (3) append the exact file, (4) allow it via commit.allow in the repo's
.gavel.yaml, (5) continue this commit once without persisting any change, or
(6) cancel. --precommit=fail|skip|false overrides the prompt; non-TTY runs
auto-escalate prompt -> fail.

Staged go.mod / go.work / package.json files are also scanned for local
references that escape the git root (go.mod replace, go.work use,
package.json file:/link:/portal: or ../ paths). Newly introduced or changed
violations relative to HEAD prompt to (1) unstage the manifest so the bad edit
is dropped from the commit, (2) ignore and keep it in this commit, or
(3) cancel. --precommit controls this check too, and skip|false disables both
the gitignore and linked-deps checks entirely.

AI analysis can also check for removed functionality and compatibility issues.
By default this is skipped. Use --compat=prompt|fail to enable it; skip|false
disables the compatibility AI checks entirely, and non-TTY runs auto-escalate
prompt -> fail.

The -i / -t flags open an interactive tree picker over every changed file
(staged, unstaged, and untracked) — no need to git add first. Each row
shows the file's language and repomap scope (e.g. Go · architecture,
TypeScript · test) plus its line delta. Toggle individual files with
space, whole folders with 'a' (selecting a folder selects all its
descendants), every Go file with 'g', or every test-scoped file with
't'. Press '/' to filter the file tree by path, status, language, or
scope; enter keeps the current filter and esc clears it. Press 'i' to
add the highlighted file ('f'), its containing folder
('d'), or every file with its extension ('e') to .gitignore — already-
tracked matches are unstaged with 'git rm --cached' so the new ignore
takes effect immediately. Press enter to confirm; gavel resets the
index and stages exactly the chosen paths before running the normal
commit pipeline. After each commit, the picker reopens over the
remaining changed files so you can build several focused commits in
one session — exit any time with esc or ctrl+c. -i is mutually
exclusive with -A and -m. Pair with -s to print a status-style summary
of the candidate files before the picker opens. Combine with --dry-run
to preview a single commit without looping.

The -A flag groups staged files by their top-level directory and recursively
splits any group that exceeds --max-files or --max-lines. An LLM still writes
the conventional commit message for each group.

With --push (-p), if nothing is staged the commit step is skipped and the
existing local commits ahead of upstream are pushed instead. A new PR is
opened (or pushed to a matching open PR). When neither staged changes nor
ahead-of-upstream commits exist, gavel exits non-zero with "nothing to
commit and no local commits ahead of upstream".

Add --auto-merge to enable GitHub auto-merge on a PR that -p opens (so it
merges once required checks pass). --merge-type sets the method
(rebase|squash|merge; default rebase). Auto-merge only applies to PRs this
run opens; when -p instead pushes to an existing PR, --auto-merge is ignored
with a warning. If GitHub rejects enabling auto-merge (repo disallows it, the
chosen method isn't enabled, no branch protection), gavel exits non-zero.

Examples:
  gavel commit                          # LLM-generated message, staged changes
  gavel commit -i                       # tree picker over all changed files; no git add needed
  gavel commit -t                       # alias for the tree picker
  gavel commit -i -s                    # show a status summary before opening the picker
  gavel commit -i --dry-run             # preview message for the picked subset
  gavel commit -A                       # one commit per directory, split when large
  gavel commit -A --max-files=3         # tighter file cap; triggers deeper splits
  gavel commit -A --max-lines=50        # tighter line cap; triggers deeper splits
  gavel commit -m "chore: bump dep"     # explicit message, still run compatibility analysis
  gavel commit --stage all --dry-run    # stage everything, print message
  gavel commit --force                  # skip hooks
  gavel commit -y                       # auto-unstage linked-dep replacements, auto-AI-fix lint findings
  gavel commit --precommit=fail         # error on gitignore or linked-deps issues
  gavel commit --lint=true              # also run every detected linter on staged files
  gavel commit --lint-secrets=false     # skip the betterleaks secrets scan (default: on)
  gavel commit -p                       # commit (if anything staged) then push / open PR
  gavel commit -p                       # with nothing staged: skip commit, push HEAD, open PR
  gavel commit -p --auto-merge          # open PR and enable auto-merge (rebase) once checks pass
  gavel commit -p --auto-merge --merge-type=squash  # auto-merge with a squash merge
  gavel commit --fixup=<hash>           # squash all staged files into <hash>, then autosquash
  gavel commit --fixup                  # auto-route each file by last-touching commit; leftovers fall through to a normal commit
  gavel commit --fixup --no-autosquash  # leave fixup! commits in place; user runs rebase later
  gavel commit --since=origin/main      # merge commits sharing a Gavel-Issue-Id in origin/main..HEAD into one each
  gavel commit --since=~20 --dry-run    # preview which Gavel-Issue-Id groups in the last 20 commits would merge`
}

func init() {
	cmd := clicky.AddNamedCommand("commit", rootCmd, CommitOptions{}, runCommit)
	// Allow `gavel commit --fixup` (no value) to mean "auto-route per file";
	// `--fixup=<hash>` keeps explicit semantics. NoOptDefVal is the cobra
	// hook for this; clicky's struct-tag binding doesn't surface it.
	if f := cmd.Flags().Lookup("fixup"); f != nil {
		f.NoOptDefVal = commitpkg.FixupAuto
	}
}

func buildCommitOptions(opts CommitOptions, workDir string, cfg verify.GavelConfig) commitpkg.Options {
	return commitpkg.Options{
		WorkDir:         workDir,
		Stage:           opts.Stage,
		CommitAll:       opts.CommitAll,
		AIGroup:         opts.AIGroup,
		Interactive:     opts.Interactive || opts.Tree,
		Summary:         opts.Summary,
		MaxFiles:        opts.MaxFiles,
		MaxLines:        opts.MaxLines,
		MaxCommits:      opts.MaxCommits,
		GroupByScope:    opts.GroupByScope,
		DryRun:          opts.DryRun,
		Force:           opts.Force,
		NoCache:         opts.NoCache,
		Model:           opts.Model,
		GroupModel:      opts.GroupModel,
		Message:         opts.Message,
		Push:            opts.Push,
		AutoMerge:       opts.AutoMerge,
		MergeType:       opts.MergeType,
		Fixup:           opts.Fixup,
		Since:           opts.Since,
		Autosquash:      !opts.NoAutosquash,
		PrecommitMode:   opts.Precommit,
		CompatMode:      opts.Compat,
		LintFlag:        opts.Lint,
		LintSecretsFlag: opts.LintSecrets,
		TidyFlag:        opts.Tidy,
		AssumeYes:       opts.Yes,
		AddMetadata:     opts.AddMetadata,
		Config:          cfg.Commit,
	}
}

func runCommit(opts CommitOptions) (any, error) {
	workDir := opts.WorkDir
	if workDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		workDir = wd
	}
	if root := repomap.FindGitRoot(workDir); root != "" {
		workDir = root
	}

	if opts.Push && opts.AutoMerge {
		if _, err := github.MergeMethodFor(opts.MergeType); err != nil {
			return nil, err
		}
	}

	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	result, err := commitpkg.Run(context.Background(), buildCommitOptions(opts, workDir, cfg))

	if err != nil {
		if errors.Is(err, commitpkg.ErrNothingStaged) {
			fmt.Fprintln(os.Stderr, "nothing staged to commit")
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrGitIgnoreCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrLinkedDepsCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrCompatibilityCancelled) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrNothingToPush) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrInteractiveWithCommitAll) ||
			errors.Is(err, commitpkg.ErrInteractiveWithMessage) ||
			errors.Is(err, commitpkg.ErrInteractiveNonTTY) ||
			errors.Is(err, commitpkg.ErrInteractiveCancelled) ||
			errors.Is(err, commitpkg.ErrInteractiveEmpty) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrFixupWithCommitAll) ||
			errors.Is(err, commitpkg.ErrFixupWithInteractive) ||
			errors.Is(err, commitpkg.ErrFixupWithMessage) ||
			errors.Is(err, commitpkg.ErrFixupInvalidTarget) ||
			errors.Is(err, commitpkg.ErrFixupNoBase) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrSinceNoDuplicates) {
			fmt.Fprintln(os.Stderr, err.Error())
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrSinceInvalidRef) ||
			errors.Is(err, commitpkg.ErrSincePushed) ||
			errors.Is(err, commitpkg.ErrSinceNeedsConfirm) ||
			errors.Is(err, commitpkg.ErrSinceWithMessage) ||
			errors.Is(err, commitpkg.ErrSinceWithCommitAll) ||
			errors.Is(err, commitpkg.ErrSinceWithInteractive) {
			fmt.Fprintln(os.Stderr, err.Error())
			exitCode = 1
			return nil, nil
		}
		if errors.Is(err, commitpkg.ErrLintFindings) {
			outcome := handleCommitLintFindings(workDir, result, opts.Yes)
			switch outcome {
			case lintFindingsContinueOnce:
				retry := buildCommitOptions(opts, workDir, cfg)
				retry.LintFlag = "false"
				retry.LintSecretsFlag = "false"
				logger.Infof("lint: continuing this commit with lint gate disabled (one-time bypass)")
				retryResult, retryErr := commitpkg.Run(context.Background(), retry)
				if retryErr != nil {
					return retryResult, retryErr
				}
				return retryResult, nil
			case lintFindingsAIFixed:
				retry := buildCommitOptions(opts, workDir, cfg)
				logger.Infof("lint: ai-fix applied edits; re-running commit with lint gate enabled")
				retryResult, retryErr := commitpkg.Run(context.Background(), retry)
				if retryErr != nil {
					return retryResult, retryErr
				}
				return retryResult, nil
			default:
				exitCode = 1
				return nil, nil
			}
		}
		return result, err
	}
	return result, nil
}

type lintFindingsOutcome int

const (
	// lintFindingsBlocked is returned when the user triages or cancels.
	// Caller should exit non-zero. Triage rules (if any) have already been
	// written to .gavel.yaml.
	lintFindingsBlocked lintFindingsOutcome = iota
	// lintFindingsContinueOnce is returned when the user opts to bypass the
	// lint gate for this commit only. Caller should re-run commit with lint
	// flags forced off.
	lintFindingsContinueOnce
	// lintFindingsAIFixed is returned when Claude was invoked, edits were
	// applied, and the post-fix lint pass came back clean. Caller should
	// re-run commit with the lint gate STILL ON (no bypass).
	lintFindingsAIFixed
)

// handleCommitLintFindings prints the per-violation report and asks the user
// whether to triage (persist ignore rules), continue this commit anyway
// (one-time bypass, no .gavel.yaml change), or cancel. Returns
// lintFindingsContinueOnce when the caller should retry the commit with the
// lint gate disabled; otherwise returns lintFindingsBlocked.
func handleCommitLintFindings(workDir string, result *commitpkg.Result, assumeYes bool) lintFindingsOutcome {
	if result == nil || result.Lint == nil {
		fmt.Fprintln(os.Stderr, "commit blocked: lint reported violations")
		return lintFindingsBlocked
	}
	for _, lr := range result.Lint.Results {
		if lr == nil || lr.Skipped {
			continue
		}
		for _, v := range lr.Violations {
			fmt.Fprintln(os.Stderr, formatCommitLintViolation(lr.Linter, v))
		}
	}
	fmt.Fprintf(os.Stderr, "\ncommit blocked: %d lint violation(s)\n", result.Lint.Violations)

	if assumeYes {
		return runCommitAIFix(workDir, result, true)
	}

	switch promptLintFindingsAction() {
	case lintActionAIFix:
		return runCommitAIFix(workDir, result, false)
	case lintActionContinueOnce:
		return lintFindingsContinueOnce
	case lintActionCancel:
		return lintFindingsBlocked
	}

	newRules, triageErr := runTriage(result.Lint.Results, workDir)
	if triageErr != nil {
		fmt.Fprintf(os.Stderr, "triage failed: %v\n", triageErr)
		return lintFindingsBlocked
	}
	if len(newRules) == 0 {
		return lintFindingsBlocked
	}
	cfgPath := filepath.Join(workDir, ".gavel.yaml")
	repoCfg, err := verify.LoadSingleGavelConfig(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", cfgPath, err)
		return lintFindingsBlocked
	}
	repoCfg.Lint.Ignore = append(repoCfg.Lint.Ignore, newRules...)
	if err := verify.SaveGavelConfig(workDir, repoCfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", cfgPath, err)
		return lintFindingsBlocked
	}
	fmt.Fprintf(os.Stderr, "Saved %d new ignore rule(s) to %s. Re-run `gavel commit` to retry.\n", len(newRules), cfgPath)
	return lintFindingsBlocked
}

func formatCommitLintViolation(linter string, v models.Violation) string {
	rule := ""
	if v.Rule != nil {
		rule = v.Rule.Method
	}
	msg := ""
	if v.Message != nil {
		msg = *v.Message
	}
	loc := v.File
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, v.Line)
	}
	if rule != "" {
		return fmt.Sprintf("  %s [%s/%s] %s", loc, linter, rule, msg)
	}
	return fmt.Sprintf("  %s [%s] %s", loc, linter, msg)
}
