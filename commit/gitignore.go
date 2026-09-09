package commit

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// resolveGitRoot returns the absolute top-level directory of the git working
// tree containing workDir, via `git rev-parse --show-toplevel`. It is
// worktree- and submodule-aware (unlike a hand-rolled `.git` walk). The commit
// flow normalizes WorkDir to this root so that staged paths (which git reports
// relative to the repository root) line up with the pathspecs handed to
// `git reset`/`git add` and with where `.gitignore` / `.gavel.yaml` are written.
func resolveGitRoot(workDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("resolve git root for %q: %w: %s", workDir, err, strings.TrimSpace(stderr.String()))
	}
	root := strings.TrimSpace(string(out))
	if root == "" {
		return "", fmt.Errorf("resolve git root for %q: empty toplevel", workDir)
	}
	return root, nil
}

// gitCheckIgnore returns the subset of relPaths that git would ignore, using
// `git check-ignore` as the single source of truth. This honors every exclude
// source git itself does — the repository's `.gitignore` files, `.git/info/
// exclude`, and the global `core.excludesFile` — and resolves correctly inside
// linked worktrees and submodules.
//
//   - `--no-index` makes the check purely pattern-based, so a *tracked* file that
//     matches an ignore rule (e.g. a committed-but-gitignored build bundle) is
//     still reported. Without it, tracked paths are never flagged and the whole
//     "strip gitignored files from the commit" behavior silently no-ops.
//   - `-z --stdin` feeds and reads NUL-separated paths, so filenames with spaces
//     or other awkward characters round-trip intact.
//
// relPaths are interpreted relative to dir (git's working directory for the
// invocation); the returned keys are the same strings that were passed in.
func gitCheckIgnore(dir string, relPaths []string) (map[string]struct{}, error) {
	ignored := make(map[string]struct{})
	if len(relPaths) == 0 {
		return ignored, nil
	}

	cmd := exec.Command("git", "check-ignore", "--no-index", "-z", "--stdin")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(strings.Join(relPaths, "\x00"))
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// `git check-ignore` exits 1 when no path matched — that is a normal
		// "nothing ignored" result, not a failure. Only a higher code (128,
		// fatal) is a real error to surface loudly.
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
			return ignored, nil
		}
		return nil, fmt.Errorf("git check-ignore in %q: %w: %s", dir, err, strings.TrimSpace(stderr.String()))
	}

	for _, p := range strings.Split(stdout.String(), "\x00") {
		if p != "" {
			ignored[p] = struct{}{}
		}
	}
	return ignored, nil
}
