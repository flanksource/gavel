package commit

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
)

var (
	ErrFixupWithCommitAll   = errors.New("--fixup cannot be combined with --commit-all")
	ErrFixupWithInteractive = errors.New("--fixup cannot be combined with --interactive")
	ErrFixupWithMessage     = errors.New("--fixup cannot be combined with --message")
	ErrFixupInvalidTarget   = errors.New("--fixup target is not a valid commit")
	ErrFixupNoBase          = errors.New("--fixup auto-route needs a base ref (origin/main, origin/master, or @{upstream})")
)

// FixupAuto is the sentinel value cobra writes when the user passes a bare
// `--fixup` (no value). cmd/gavel/commit.go sets NoOptDefVal to this string.
const FixupAuto = "auto"

// fixupRoute is the per-target output of routeFilesByLastTouch: a target
// commit hash plus the staged files that should be folded into it.
type fixupRoute struct {
	Hash  string
	Files []string
}

// validateFixupOptions enforces flag combinations specific to --fixup. Mirrors
// validateInteractiveOptions in shape so the call site stays uniform.
func validateFixupOptions(opts Options) error {
	if opts.Fixup == "" {
		return nil
	}
	if opts.CommitAll {
		return ErrFixupWithCommitAll
	}
	if opts.Interactive {
		return ErrFixupWithInteractive
	}
	if strings.TrimSpace(opts.Message) != "" {
		return ErrFixupWithMessage
	}
	return nil
}

// runFixup is the entry point dispatched from Run() when opts.Fixup is set.
//
// Two modes:
//   - opts.Fixup == FixupAuto: route each staged file to the most recent
//     commit on base..HEAD that touched it. Files with no in-range match
//     fall through to a normal runSingleCommit() for the leftovers.
//   - opts.Fixup == <hash>: every staged file becomes one fixup commit
//     against that hash; no leftover handling.
//
// All pre-commit gates (gitignore / file-size / linked-deps / hooks / lint)
// run once over the union of staged files before any commit is created. The
// AI compatibility check is skipped — fixups are by definition a touch-up of
// an already-analyzed change.
//
// When opts.Autosquash is true (default), `git rebase -i --autosquash <base>`
// runs after all fixup + leftover commits exist, with a no-op sequence
// editor so the rebase is non-interactive.
func runFixup(ctx context.Context, opts Options) (*Result, error) {
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

	lintRes, lintErr := applyLintGate(ctx, opts.WorkDir, source.Files, opts.lintGates)
	result.Lint = lintRes
	if lintErr != nil {
		return result, lintErr
	}

	// Base ref is only needed for auto routing (per-file last-touch lookup)
	// and for the autosquash rebase. Explicit-hash mode without autosquash
	// doesn't need it; skip the lookup so an unconfigured remote isn't fatal.
	var base string
	if opts.Fixup == FixupAuto || opts.Autosquash {
		b, baseErr := resolveFixupBase(opts.WorkDir)
		if baseErr != nil {
			return result, baseErr
		}
		base = b
	}

	routes, leftovers, err := planFixup(opts.WorkDir, source.Files, base, opts.Fixup)
	if err != nil {
		return result, err
	}

	if opts.DryRun {
		printFixupDryRun(result, routes, leftovers, opts.Fixup)
		return result, nil
	}

	stagedSet := source.Files
	for _, route := range routes {
		if err := restageOnly(opts.WorkDir, stagedSet, route.Files); err != nil {
			return result, fmt.Errorf("restage for fixup of %s: %w", shortHash(route.Hash), err)
		}
		hash, commitErr := commitFixup(opts.WorkDir, route.Hash)
		if commitErr != nil {
			return result, fmt.Errorf("git commit --fixup=%s: %w", shortHash(route.Hash), commitErr)
		}
		msg, _ := commitMessage(opts.WorkDir, hash)
		result.Commits = append(result.Commits, CommitResult{
			Message: msg,
			Hash:    hash,
			Files:   append([]string(nil), route.Files...),
		})
		logger.Infof("Created fixup %s -> %s for %d file(s)", shortHash(hash), shortHash(route.Hash), len(route.Files))
	}

	if len(leftovers) > 0 {
		if err := restageOnly(opts.WorkDir, stagedSet, leftovers); err != nil {
			return result, fmt.Errorf("restage leftover files: %w", err)
		}
		leftoverOpts := opts
		leftoverOpts.Fixup = ""
		leftoverOpts.Stage = StageStaged
		leftoverResult, err := runSingleCommit(ctx, leftoverOpts)
		if err != nil {
			return mergeFixupResults(result, leftoverResult), err
		}
		result = mergeFixupResults(result, leftoverResult)
	}

	if opts.Autosquash && len(routes) > 0 {
		if err := runAutosquash(opts.WorkDir, base); err != nil {
			return result, err
		}
		// HEAD hash changes after rebase; refresh the most recent commit so
		// callers (e.g. --push) see the right ref.
		if newHead, herr := headHash(opts.WorkDir); herr == nil {
			result.Hash = newHead
		}
	}
	return result, nil
}

// planFixup computes the list of fixup routes plus leftovers. With an
// explicit hash all files become one route; with FixupAuto each file is
// routed by its last-touching commit on base..HEAD.
func planFixup(workDir string, files []string, base, fixupValue string) ([]fixupRoute, []string, error) {
	if fixupValue != FixupAuto {
		// `^{commit}` forces git to peel the ref to a commit object that
		// actually exists. Plain `--verify --quiet <40-hex>` accepts any
		// well-formed hash even if the object isn't in the database.
		if !validRef(workDir, fixupValue+"^{commit}") {
			return nil, nil, fmt.Errorf("%w: %q", ErrFixupInvalidTarget, fixupValue)
		}
		return []fixupRoute{{Hash: fixupValue, Files: append([]string(nil), files...)}}, nil, nil
	}
	return routeFilesByLastTouch(workDir, files, base)
}

// routeFilesByLastTouch buckets each file by the most recent commit on
// base..HEAD that touched it. Files with no match are returned as leftovers.
// Routes are returned sorted by commit hash for deterministic output.
func routeFilesByLastTouch(workDir string, files []string, base string) ([]fixupRoute, []string, error) {
	groups := map[string][]string{}
	var leftovers []string
	for _, f := range files {
		hash, err := lastTouchingCommit(workDir, base, f)
		if err != nil {
			return nil, nil, err
		}
		if hash == "" {
			leftovers = append(leftovers, f)
			continue
		}
		groups[hash] = append(groups[hash], f)
	}
	hashes := make([]string, 0, len(groups))
	for h := range groups {
		hashes = append(hashes, h)
	}
	sort.Strings(hashes)
	routes := make([]fixupRoute, 0, len(hashes))
	for _, h := range hashes {
		sort.Strings(groups[h])
		routes = append(routes, fixupRoute{Hash: h, Files: groups[h]})
	}
	sort.Strings(leftovers)
	return routes, leftovers, nil
}

// resolveFixupBase picks the ref `base..HEAD` should walk. Priority:
// origin/main, origin/master, then the configured upstream. Returns
// ErrFixupNoBase when none of those refs exist locally.
func resolveFixupBase(workDir string) (string, error) {
	for _, candidate := range []string{"origin/main", "origin/master"} {
		if validRef(workDir, candidate) {
			return candidate, nil
		}
	}
	if validRef(workDir, "@{upstream}") {
		return "@{upstream}", nil
	}
	return "", ErrFixupNoBase
}

// restageOnly resets every file in the original staged set, then re-stages
// just `keep`. Used between fixup-route iterations so each commit captures
// exactly that route's files.
func restageOnly(workDir string, original, keep []string) error {
	if err := resetFiles(workDir, original); err != nil {
		return err
	}
	return addFiles(workDir, keep)
}

// runAutosquash runs `git rebase -i --autosquash <base>` non-interactively
// by setting GIT_SEQUENCE_EDITOR to a no-op so git accepts the auto-generated
// todo list as-is. On rebase conflict we abort and surface the failure with
// a clear message — callers can re-run with conflicts resolved manually.
func runAutosquash(workDir, base string) error {
	cmd := exec.Command("git", "-c", "sequence.editor=:", "rebase", "-i", "--autosquash", base)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GIT_SEQUENCE_EDITOR=:", "GIT_EDITOR=:")
	cmd.Stdout = os.Stdout
	var stderr strings.Builder
	cmd.Stderr = &stderr
	err := cmd.Run()
	stderrStr := stderr.String()
	if stderrStr != "" {
		os.Stderr.WriteString(stderrStr)
	}
	if err == nil {
		return nil
	}
	// On conflict, abort so the working tree is restored to pre-rebase state.
	if strings.Contains(stderrStr, "CONFLICT") || strings.Contains(stderrStr, "could not apply") {
		_ = runGitRebaseAbort(workDir)
		return fmt.Errorf("autosquash rebase onto %s conflicted; aborted. Resolve manually or rerun with --no-autosquash", base)
	}
	return fmt.Errorf("git rebase --autosquash %s: %w", base, err)
}

func headHash(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// mergeFixupResults folds a leftover-commit Result into the fixup-loop
// aggregate. Hooks/Lint/Staged stay as the most recent values; Commits
// accumulate so every fixup + leftover commit shows in the final output.
func mergeFixupResults(agg, leftover *Result) *Result {
	if leftover == nil {
		return agg
	}
	if agg == nil {
		agg = &Result{}
	}
	agg.Message = leftover.Message
	agg.Hash = leftover.Hash
	agg.Staged = leftover.Staged
	agg.Hooks = leftover.Hooks
	agg.DryRun = leftover.DryRun
	if leftover.Lint != nil {
		agg.Lint = leftover.Lint
	}
	agg.Commits = append(agg.Commits, leftover.Commits...)
	return agg
}

// printFixupDryRun renders the planned routing table without creating any
// commits. Mirrors the live path so users see exactly what would happen.
func printFixupDryRun(result *Result, routes []fixupRoute, leftovers []string, fixupValue string) {
	fmt.Fprintln(dryRunOutput, "DRY RUN: --fixup plan")
	if fixupValue != FixupAuto {
		fmt.Fprintf(dryRunOutput, "  target: %s (explicit)\n", shortHash(fixupValue))
	}
	for _, route := range routes {
		fmt.Fprintf(dryRunOutput, "  fixup -> %s (%d file(s))\n", shortHash(route.Hash), len(route.Files))
		for _, f := range route.Files {
			fmt.Fprintf(dryRunOutput, "      %s\n", f)
		}
		result.Commits = append(result.Commits, CommitResult{
			Message: "fixup! " + shortHash(route.Hash),
			Files:   append([]string(nil), route.Files...),
		})
	}
	if len(leftovers) > 0 {
		fmt.Fprintf(dryRunOutput, "  new commit (no in-range match): %d file(s)\n", len(leftovers))
		for _, f := range leftovers {
			fmt.Fprintf(dryRunOutput, "      %s\n", f)
		}
	}
}
