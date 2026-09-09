package commit

import (
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/verify"
)

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
// StageSession resolves the agent session id from the environment and stages
// only that session's edited files, falling back to StageStaged when no session
// id is set. Any other mode value is treated as an explicit session id (Claude
// or Codex): stageSessionFiles stages exactly the files that session touched
// (see --stage=session|<session-id>).
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
	case StageSession:
		sessionID := resolveEnvSessionID()
		if sessionID == "" {
			logger.V(1).Infof("commit: --stage=session but no agent session id in env; committing the staged set")
			return nil
		}
		return stageSessionFiles(workDir, sessionID, cfg)
	default:
		return stageSessionFiles(workDir, mode, cfg)
	}
}

// unstageGitIgnored removes from the index any staged file that matches the
// repo's .gitignore, except files in preStaged (staged by the user before
// gavel ran). It uses `git reset --` so the working tree and tracking are
// untouched — the modification simply won't be in the commit. Matching is done
// by `git check-ignore` (see gitCheckIgnore), so !-negations, .git/info/exclude,
// the global excludesFile, and worktrees are all honored.
func unstageGitIgnored(workDir string, preStaged []string) error {
	staged, err := stagedFiles(workDir)
	if err != nil {
		return fmt.Errorf("list staged files: %w", err)
	}
	if len(staged) == 0 {
		return nil
	}

	preStagedSet := make(map[string]struct{}, len(preStaged))
	for _, f := range preStaged {
		preStagedSet[f] = struct{}{}
	}

	ignored, err := gitCheckIgnore(workDir, staged)
	if err != nil {
		return err
	}
	toReset := make([]string, 0, len(ignored))
	for _, f := range staged {
		if _, isIgnored := ignored[f]; !isIgnored {
			continue
		}
		if _, kept := preStagedSet[f]; kept {
			continue
		}
		logger.Infof("commit: skipping %s (matches .gitignore)", f)
		toReset = append(toReset, f)
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
