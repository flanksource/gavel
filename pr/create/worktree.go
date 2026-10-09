package prcreate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	markerFile  = ".gavel-pr-create.json"
	tmpBranchNS = "gavel/pr-create-tmp/"
)

var scratchSub = filepath.FromSlash(".tmp/pr-create")

type worktree struct {
	path, tmpBranch, suffix string
}

// openWorktree adds a marked scratch worktree on a temporary branch off base.
func openWorktree(repoRoot, base, firstSHA string) (worktree, error) {
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	name := firstSHA[:8] + "-" + suffix
	ws := worktree{
		path:      filepath.Join(repoRoot, scratchSub, name),
		tmpBranch: tmpBranchNS + name,
		suffix:    suffix,
	}
	if err := os.MkdirAll(filepath.Dir(ws.path), 0o755); err != nil {
		return worktree{}, fmt.Errorf("create scratch dir: %w", err)
	}
	if err := runGitVerbose(repoRoot, "worktree", "add", "-b", ws.tmpBranch, "--end-of-options", ws.path, base); err != nil {
		return worktree{}, fmt.Errorf("git worktree add: %w", err)
	}
	if err := writeMarker(ws.path, firstSHA, repoRoot); err != nil {
		if cleanupErr := removeWorktree(repoRoot, ws.path); cleanupErr != nil {
			return worktree{}, fmt.Errorf("%w (worktree cleanup: %v)", err, cleanupErr)
		}
		return worktree{}, err
	}
	return ws, nil
}

func gitCherryPick(wtPath, fullSHA string, mainline int) error {
	args := []string{"cherry-pick"}
	if mainline > 0 {
		args = append(args, "-m", strconv.Itoa(mainline))
	}
	return runGitVerbose(wtPath, append(args, fullSHA)...)
}

// isNoOpCherryPick reports whether a failed cherry-pick left a clean tree,
// i.e. the commit is already on the base. The untracked marker file is
// ignored.
func isNoOpCherryPick(wtPath string) bool {
	out, err := captureGit(wtPath, "status", "--porcelain", "--untracked-files=no")
	return err == nil && out == ""
}

func gitPushTopic(wtPath, topic string) error {
	return runGitVerbose(wtPath, "push", "-u", "origin", "HEAD:refs/heads/"+topic)
}

func gitWorkingTreeDirty(repoRoot string) (bool, error) {
	out, err := captureGit(repoRoot, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// --- marker + safe cleanup --------------------------------------------------

type marker struct {
	CreatedAt  time.Time `json:"createdAt"`
	SourceSHA  string    `json:"sourceSHA"`
	ParentRepo string    `json:"parentRepo"`
	PID        int       `json:"pid"`
}

func writeMarker(wtPath, fullSHA, repoRoot string) error {
	data, err := json.MarshalIndent(marker{
		CreatedAt:  time.Now().UTC(),
		SourceSHA:  fullSHA,
		ParentRepo: repoRoot,
		PID:        os.Getpid(),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal marker: %w", err)
	}
	return os.WriteFile(filepath.Join(wtPath, markerFile), data, 0o644)
}

// removeWorktree deletes the worktree at path only after three independent
// safety checks pass. Any failure leaves the directory in place.
func removeWorktree(repoRoot, path string) error {
	if err := assertSafeWorktreePath(repoRoot, path); err != nil {
		return err
	}
	if err := runGitVerbose(repoRoot, "worktree", "remove", "--force", path); err != nil {
		return fmt.Errorf("git worktree remove %s: %w (leaving for manual cleanup)", path, err)
	}
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("os.RemoveAll %s: %w", path, err)
	}
	return nil
}

func assertSafeWorktreePath(repoRoot, path string) error {
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("abs repoRoot: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}
	scratchPrefix := filepath.Join(absRoot, scratchSub) + string(filepath.Separator)
	if !strings.HasPrefix(absPath, scratchPrefix) {
		return fmt.Errorf("refusing to remove %s: outside %s", absPath, scratchPrefix)
	}
	data, err := os.ReadFile(filepath.Join(absPath, markerFile))
	if err != nil {
		return fmt.Errorf("refusing to remove %s: marker missing (%w)", absPath, err)
	}
	var m marker
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("refusing to remove %s: marker unparseable (%w)", absPath, err)
	}
	if !worktreeRegistered(repoRoot, absPath) {
		return fmt.Errorf("refusing to remove %s: not in git worktree list", absPath)
	}
	return nil
}

func worktreeRegistered(repoRoot, absPath string) bool {
	out, err := captureGit(repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(out, "\n") {
		if path, ok := strings.CutPrefix(line, "worktree "); ok && path == absPath {
			return true
		}
	}
	return false
}

// --- git helpers ------------------------------------------------------------

// runGitVerbose streams git's output to the terminal, as the pr create command
// always has for worktree, cherry-pick and push steps.
func runGitVerbose(workDir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", workDir}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGitQuiet(workDir string, args ...string) error {
	return exec.Command("git", append([]string{"-C", workDir}, args...)...).Run()
}

// captureGit returns git's trimmed stdout, or an error carrying its stderr.
func captureGit(workDir string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", workDir}, args...)...).Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
