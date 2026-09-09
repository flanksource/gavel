package commit

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ResolveCommitFiles turns positional file arguments (from `gavel commit
// <files>`) into git-root-relative pathspecs.
//
// baseDir is the directory the user invoked gavel from; relative args resolve
// against it. gitRoot is the repository top-level. Callers derive gitRoot from
// baseDir via FindGitRoot so both share the same symlink representation and the
// relative result is stable (this sidesteps the macOS /var vs /private/var
// mismatch that would otherwise leak through filepath.Rel). Paths that escape
// the repository are rejected with ErrFilesOutsideRepo.
//
// The returned specs are consumed as git pathspecs, so "." and directories pass
// through unchanged (a directory stages everything beneath it).
func ResolveCommitFiles(gitRoot, baseDir string, args []string) ([]string, error) {
	specs := make([]string, 0, len(args))
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		abs := arg
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(baseDir, abs)
		}
		rel, err := filepath.Rel(gitRoot, abs)
		if err != nil {
			return nil, fmt.Errorf("resolve %q against %s: %w", arg, gitRoot, err)
		}
		rel = filepath.ToSlash(rel)
		if rel == ".." || strings.HasPrefix(rel, "../") {
			return nil, fmt.Errorf("%w: %s", ErrFilesOutsideRepo, arg)
		}
		specs = append(specs, rel)
	}
	return specs, nil
}

// validateFilesOptions rejects flag combinations whose staging or history
// semantics conflict with an explicit file list. --interactive drives its own
// picker selection, --since operates on history and ignores staged files, and
// --fixup already targets the staged set (use `git add <files>` first).
func validateFilesOptions(opts Options) error {
	if len(opts.Files) == 0 {
		return nil
	}
	if opts.Interactive {
		return ErrFilesWithInteractive
	}
	if opts.Since != "" {
		return ErrFilesWithSince
	}
	if opts.Fixup != "" {
		return ErrFilesWithFixup
	}
	return nil
}

// stageExplicitFiles resets the index then stages exactly files (git-root-
// relative pathspecs), so `gavel commit <files>` commits only those paths
// regardless of what was previously staged. Mirrors the interactive picker's
// reset-then-stage (runInteractiveStaging). Explicit selection intentionally
// overrides .gitignore — the user named the files — while the precommit
// gitignore gate (applyPrecommitChecks) still runs downstream.
func stageExplicitFiles(workDir string, files []string) error {
	if err := resetAllStaged(workDir); err != nil {
		return fmt.Errorf("reset index before staging file arguments: %w", err)
	}
	if err := addFiles(workDir, files); err != nil {
		return fmt.Errorf("stage file arguments: %w", err)
	}
	return nil
}
