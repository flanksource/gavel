package serve

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/commons/logger"
	"github.com/gliderlabs/ssh"
)

// HandleGitReceive handles a git-receive-pack session with a cached bare repo.
// Returns the exit code for the SSH session.
func HandleGitReceive(sess ssh.Session, repoPath, repoDir string, hookWriter func(string, string) error) int {
	bareRepo := filepath.Join(repoDir, repoPath+".git")
	logger.V(2).Infof("Bare repo path: %s", bareRepo)

	if err := ensureBareRepo(bareRepo); err != nil {
		fmt.Fprintf(sess.Stderr(), "failed to init repo: %v\n", err)
		return 1
	}

	gavelPath, err := os.Executable()
	if err != nil {
		gavelPath = "gavel"
	}
	logger.V(3).Infof("Writing hook with gavel path: %s", gavelPath)

	if err := hookWriter(bareRepo, gavelPath); err != nil {
		fmt.Fprintf(sess.Stderr(), "failed to write hook: %v\n", err)
		return 1
	}

	logger.V(1).Infof("exec: git receive-pack %s", bareRepo)
	cmd := exec.Command("git", "receive-pack", bareRepo)
	cmd.Stdin = sess
	cmd.Stdout = sess
	cmd.Stderr = sess.Stderr()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logger.V(1).Infof("git receive-pack exited with code %d", exitErr.ExitCode())
			return exitErr.ExitCode()
		}
		fmt.Fprintf(sess.Stderr(), "git receive-pack failed: %v\n", err)
		return 1
	}
	return 0
}

func ensureBareRepo(path string) error {
	if _, err := os.Stat(filepath.Join(path, "HEAD")); err == nil {
		return nil // already initialized
	}

	logger.Infof("Initializing bare repo at %s", path)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}

	cmd := exec.Command("git", "init", "--bare", path)
	return cmd.Run()
}
