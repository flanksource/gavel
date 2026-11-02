package snapshots

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/utils"
)

// GitInfo is the worktree a snapshot of workDir belongs to: its root, the
// root's base name as the repo, and HEAD. Outside a worktree the root is
// workDir itself and the SHA empty, which Save cannot key; nil for no workDir.
func GitInfo(workDir string) *testui.SnapshotGit {
	if workDir == "" {
		return nil
	}

	root := utils.FindGitRoot(workDir)
	inGit := root != ""
	if !inGit {
		abs, err := filepath.Abs(workDir)
		if err == nil {
			root = abs
		} else {
			root = workDir
		}
	}

	git := &testui.SnapshotGit{
		Repo: filepath.Base(root),
		Root: root,
	}
	if inGit {
		git.SHA = gitHeadSHA(root)
	}
	return git
}

func gitHeadSHA(workDir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		logger.V(2).Infof("git rev-parse HEAD in %s failed: %v", workDir, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}
