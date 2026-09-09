package commit

import (
	"context"
	"fmt"
	"sync"

	"github.com/flanksource/commons/logger"
)

// runCommitAll stages all changes, asks the LLM to split them into logical
// commit groups plus a chore group for lock/generated files, and creates one
// commit per group. It backs `-A` / `--commit-all` (also implied by
// --max-commits).
func runCommitAll(ctx context.Context, opts Options) (*Result, error) {
	if err := stageCommitAllSource(opts.WorkDir, opts.Config); err != nil {
		return nil, err
	}

	source, result, err := prepareMultiCommit(ctx, opts)
	if err != nil {
		return result, err
	}

	return commitByGrouping(ctx, opts, source, result)
}

// prepareMultiCommit runs the shared staging-completion pipeline for the
// grouped-commit flow (runCommitAll): it reads the staged source, applies the
// precommit gates, runs hooks, re-reads the staged source, and applies the lint
// gate. The caller is responsible for staging beforehand (stageCommitAllSource).
// It returns the final staged source and a Result pre-populated with
// Staged/Hooks/Lint so error returns carry partial state.
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

// commitByGrouping asks the LLM to split the staged changes into logical commit
// groups (plus a chore group for lock/generated files) and creates one commit
// per group. It backs `-A` / `--commit-all` and is also the fallback when a
// single-commit AI analysis overflows the model context window.
func commitByGrouping(ctx context.Context, opts Options, source stagedSource, result *Result) (*Result, error) {
	groups, err := groupChangesByAIFunc(ctx, opts, source)
	if err != nil {
		return result, fmt.Errorf("ai grouping: %w", err)
	}
	return commitGroups(ctx, opts, source, result, groups)
}

// commitSerializeMu serializes the git-index mutation (stage + commit) for a
// single group so only one commit is ever in flight, even when group message
// generation runs concurrently.
var commitSerializeMu sync.Mutex

// commitGroups creates one commit per group, committing each group as soon as
// its message is ready rather than batching every commit to the end — an
// interrupted run still lands the groups it finished. A group carrying a preset
// Message (e.g. the chore group for lock/generated files) skips the LLM and uses
// it verbatim.
func commitGroups(ctx context.Context, opts Options, source stagedSource, result *Result, groups []commitGroup) (*Result, error) {
	if len(groups) == 0 {
		return result, ErrNothingStaged
	}
	result.Commits = make([]CommitResult, 0, len(groups))

	// Dry run: generate and preview every message without touching the index.
	if opts.DryRun {
		for _, group := range groups {
			cr, err := analyzeGroup(ctx, opts, group)
			if err != nil {
				return result, err
			}
			result.Commits = append(result.Commits, cr)
		}
		printDryRunPreview(result)
		return result, nil
	}

	// Live: unstage everything once, then stage+commit each group the moment its
	// message is ready. commitSerializeMu keeps a single commit in flight.
	if err := resetFiles(opts.WorkDir, source.GitPaths()); err != nil {
		return result, fmt.Errorf("reset staged files: %w", err)
	}
	for _, group := range groups {
		cr, err := analyzeGroup(ctx, opts, group)
		if err != nil {
			return result, err
		}
		hash, err := stageAndCommitGroup(opts.WorkDir, group, cr.Message)
		if err != nil {
			return result, err
		}
		cr.Hash = hash
		result.Commits = append(result.Commits, cr)
		logger.Infof("Committed %s: %s", shortHash(hash), firstLine(cr.Message))
	}

	restoreLocalReplaces(opts.WorkDir, source.PendingRestores)

	return result, nil
}

// analyzeGroup builds the CommitResult for a group: a preset-Message group is
// used verbatim (no LLM); otherwise the message and compatibility findings come
// from an LLM analysis of the group's diff.
func analyzeGroup(ctx context.Context, opts Options, group commitGroup) (CommitResult, error) {
	if group.Message != "" {
		return CommitResult{
			Label:   group.Label,
			Message: applyCommitMetadata(opts, group.Message),
			Files:   group.Files(),
		}, nil
	}
	analysis, err := generateCommitAnalysis(ctx, opts, group.diff())
	if err != nil {
		return CommitResult{}, fmt.Errorf("generate commit analysis for %s: %w", group.labelOrDefault(), err)
	}
	return CommitResult{
		Label:   group.Label,
		Message: applyCommitMetadata(opts, analysis.Message),
		Files:   group.Files(),
	}, nil
}

// stageAndCommitGroup stages exactly the group's paths and creates one commit,
// holding commitSerializeMu so only one commit is ever in flight.
func stageAndCommitGroup(workDir string, group commitGroup, message string) (string, error) {
	commitSerializeMu.Lock()
	defer commitSerializeMu.Unlock()
	if err := addFiles(workDir, group.GitPaths()); err != nil {
		return "", fmt.Errorf("stage commit group %s: %w", group.labelOrDefault(), err)
	}
	hash, err := commitWithMessage(workDir, message)
	if err != nil {
		return "", fmt.Errorf("create commit for %s: %w", group.labelOrDefault(), err)
	}
	return hash, nil
}
