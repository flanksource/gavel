package services

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/fixtures"
)

// seedGit writes the entry's files into a fresh repository under the step's
// TmpDir and, with commit, records them as one commit. Hooks, signing and the
// caller's identity are disabled so the seed is hermetic.
func seedGit(ctx context.Context, step *fixtures.StepContext, entry Entry) (map[string]any, error) {
	dir := filepath.Join(step.TmpDir, "git", entry.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entry.Git.Files))
	for path := range entry.Git.Files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeSeedFile(dir, path, entry.Git.Files[path]); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry.Name, err)
		}
	}
	commands := [][]string{{"init", "-q", "-b", "main"}}
	if entry.Git.Commit {
		commands = append(commands, []string{"add", "-A"},
			[]string{"-c", "user.name=gavel fixtures", "-c", "user.email=fixtures@gavel.invalid", "-c", "commit.gpgsign=false",
				"-c", "core.hooksPath=/dev/null", "commit", "-q", "-m", "seed " + entry.Name})
	}
	for _, args := range commands {
		if _, err := runGit(ctx, step, dir, args...); err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry.Name, err)
		}
	}
	outputs := map[string]any{"type": "git", "dir": dir, "ready": true}
	if entry.Git.Commit {
		sha, err := runGit(ctx, step, dir, "rev-parse", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("entry %q: %w", entry.Name, err)
		}
		outputs["commit"] = sha
	}
	return outputs, nil
}

func writeSeedFile(dir, path, content string) error {
	clean := filepath.Clean(path)
	if filepath.IsAbs(clean) || clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return fmt.Errorf("git file %q must be a relative path inside the repository", path)
	}
	target := filepath.Join(dir, clean)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(content), 0o644)
}

func runGit(ctx context.Context, step *fixtures.StepContext, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(step.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
